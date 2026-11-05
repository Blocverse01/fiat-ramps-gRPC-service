package auth

import (
	"context"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

const (
	ServiceTokenEnv     = "GRPC_SERVICE_TOKEN"
	ServiceTokenAltEnv  = "FIAT_RAMPS_SERVICE_TOKEN"
	LocalDevEnv         = "LOCAL_DEV"
	OwnerMetadataKey    = "x-owner-id"
	authorizationHeader = "authorization"
	serviceTokenHeader  = "x-service-token"
	bearerPrefix        = "Bearer "
)

type ownerCtxKey struct{}

// Config controls gRPC authentication. The server fails closed unless
// mTLS (RequireAndVerifyClientCert) or a service token is configured,
// or LocalDev is explicitly set.
type Config struct {
	LocalDev     bool
	TLS          bool
	CertFile     string
	KeyFile      string
	ClientCAFile string
	ServiceToken string
}

func LoadConfig(localDev, useTLS bool, certFile, keyFile, clientCAFile string) Config {
	if os.Getenv(LocalDevEnv) == "true" {
		localDev = true
	}
	token := os.Getenv(ServiceTokenEnv)
	if token == "" {
		token = os.Getenv(ServiceTokenAltEnv)
	}
	return Config{
		LocalDev:     localDev,
		TLS:          useTLS,
		CertFile:     certFile,
		KeyFile:      keyFile,
		ClientCAFile: clientCAFile,
		ServiceToken: token,
	}
}

func (c Config) HasMTLS() bool {
	return c.TLS && c.CertFile != "" && c.KeyFile != "" && c.ClientCAFile != ""
}

func (c Config) HasServiceToken() bool {
	return c.ServiceToken != ""
}

// Validate fails closed when TLS/auth is not configured, unless LocalDev is set.
func (c Config) Validate() error {
	if c.LocalDev {
		return nil
	}
	if c.TLS && (c.CertFile == "" || c.KeyFile == "") {
		return fmt.Errorf("cert_file and key_file must be set when tls is enabled")
	}
	if !c.HasMTLS() && !c.HasServiceToken() {
		return fmt.Errorf("server startup failed: TLS/auth is not configured; set mTLS (tls + cert_file + key_file + client_ca_file) or %s/%s, or pass -local_dev / %s=true for local development only", ServiceTokenEnv, ServiceTokenAltEnv, LocalDevEnv)
	}
	return nil
}

func (c Config) ServerTLSConfig() (*tls.Config, error) {
	if !c.TLS {
		return nil, nil
	}
	cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load server TLS cert: %w", err)
	}
	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}
	if c.ClientCAFile != "" {
		pem, err := os.ReadFile(c.ClientCAFile)
		if err != nil {
			return nil, fmt.Errorf("read client CA: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("client CA file does not contain any certificates")
		}
		tlsCfg.ClientCAs = pool
		tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return tlsCfg, nil
}

func (c Config) ServerOptions() ([]grpc.ServerOption, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	opts := []grpc.ServerOption{grpc.UnaryInterceptor(UnaryInterceptor(c))}
	if c.TLS {
		tlsCfg, err := c.ServerTLSConfig()
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsCfg)))
	}
	return opts, nil
}

func UnaryInterceptor(cfg Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if cfg.LocalDev && !cfg.HasMTLS() && !cfg.HasServiceToken() {
			return handler(withOwnerFromMetadata(ctx), req)
		}

		authenticated := false
		if cfg.HasMTLS() && hasVerifiedClientCert(ctx) {
			authenticated = true
		}
		if !authenticated && cfg.HasServiceToken() {
			ok, err := tokenMatches(ctx, cfg.ServiceToken)
			if err != nil {
				return nil, err
			}
			authenticated = ok
		}

		if !authenticated {
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		}
		return handler(withOwnerFromMetadata(ctx), req)
	}
}

func hasVerifiedClientCert(ctx context.Context) bool {
	p, ok := peer.FromContext(ctx)
	if !ok || p.AuthInfo == nil {
		return false
	}
	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return false
	}
	return len(tlsInfo.State.VerifiedChains) > 0
}

func tokenMatches(ctx context.Context, expected string) (bool, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	provided := firstMetadata(md, authorizationHeader)
	if provided == "" {
		provided = firstMetadata(md, serviceTokenHeader)
	}
	if provided == "" {
		return false, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	provided = strings.TrimPrefix(provided, bearerPrefix)
	if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
		return false, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	return true, nil
}

func firstMetadata(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func withOwnerFromMetadata(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	owner := firstMetadata(md, OwnerMetadataKey)
	if owner == "" {
		return ctx
	}
	return ContextWithOwnerID(ctx, owner)
}

func ContextWithOwnerID(ctx context.Context, ownerID string) context.Context {
	return context.WithValue(ctx, ownerCtxKey{}, ownerID)
}

func OwnerIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ownerCtxKey{}).(string); ok {
		return v
	}
	return ""
}

// ResolveOwnerID returns the owner used for authorization. A metadata owner
// wins; a request owner is accepted from an authenticated caller; mismatch fails.
func ResolveOwnerID(ctx context.Context, requestOwner string) (string, error) {
	ctxOwner := OwnerIDFromContext(ctx)
	if ctxOwner != "" && requestOwner != "" && ctxOwner != requestOwner {
		return "", status.Error(codes.PermissionDenied, "ownerId mismatch")
	}
	if ctxOwner != "" {
		return ctxOwner, nil
	}
	if requestOwner != "" {
		return requestOwner, nil
	}
	return "", status.Error(codes.InvalidArgument, "ownerId is required")
}
