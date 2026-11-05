package auth

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestValidateFailsClosedWithoutAuth(t *testing.T) {
	if err := (Config{}).Validate(); err == nil {
		t.Fatal("expected startup failure when TLS/auth is not configured")
	}
}

func TestValidateAllowsLocalDev(t *testing.T) {
	if err := (Config{LocalDev: true}).Validate(); err != nil {
		t.Fatalf("local-dev should allow startup: %v", err)
	}
}

func TestValidateAllowsServiceToken(t *testing.T) {
	if err := (Config{ServiceToken: "test-token"}).Validate(); err != nil {
		t.Fatalf("service token should allow startup: %v", err)
	}
}

func TestValidateAllowsMTLS(t *testing.T) {
	cfg := Config{TLS: true, CertFile: "cert.pem", KeyFile: "key.pem", ClientCAFile: "ca.pem"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("mTLS should allow startup: %v", err)
	}
}

func TestValidateRejectsTLSWithoutCertsOrToken(t *testing.T) {
	if err := (Config{TLS: true}).Validate(); err == nil {
		t.Fatal("expected failure when tls is set without certs")
	}
	if err := (Config{TLS: true, CertFile: "c", KeyFile: "k"}).Validate(); err == nil {
		t.Fatal("expected failure when server TLS is set without client CA or token")
	}
}

func TestUnaryInterceptorRejectsMissingToken(t *testing.T) {
	cfg := Config{ServiceToken: "expected-token"}
	interceptor := UnaryInterceptor(cfg)
	_, err := interceptor(context.Background(), nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler must not run")
		return nil, nil
	})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("got %v, want Unauthenticated", err)
	}
}

func TestUnaryInterceptorAcceptsConstantTimeToken(t *testing.T) {
	cfg := Config{ServiceToken: "expected-token"}
	interceptor := UnaryInterceptor(cfg)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(authorizationHeader, bearerPrefix+"expected-token", OwnerMetadataKey, "owner-1"))
	called := false
	_, err := interceptor(ctx, nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		if OwnerIDFromContext(ctx) != "owner-1" {
			t.Fatalf("owner not propagated")
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("expected success: %v", err)
	}
	if !called {
		t.Fatal("handler was not called")
	}
}

func TestUnaryInterceptorRejectsWrongToken(t *testing.T) {
	cfg := Config{ServiceToken: "expected-token"}
	interceptor := UnaryInterceptor(cfg)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(authorizationHeader, "wrong"))
	_, err := interceptor(ctx, nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler must not run")
		return nil, nil
	})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("got %v, want Unauthenticated", err)
	}
}

func TestResolveOwnerID(t *testing.T) {
	ctx := ContextWithOwnerID(context.Background(), "owner-a")
	got, err := ResolveOwnerID(ctx, "owner-a")
	if err != nil || got != "owner-a" {
		t.Fatalf("got %q %v", got, err)
	}
	if _, err := ResolveOwnerID(ctx, "owner-b"); err == nil {
		t.Fatal("expected mismatch error")
	}
	got, err = ResolveOwnerID(context.Background(), "owner-c")
	if err != nil || got != "owner-c" {
		t.Fatalf("request owner fallback: %q %v", got, err)
	}
	if _, err := ResolveOwnerID(context.Background(), ""); err == nil {
		t.Fatal("expected missing owner error")
	}
}
