package main

import (
	"context"
	"net"
	"testing"
	"time"

	pb "github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func startTestServer(t *testing.T, cfg auth.Config) (pb.FiatRampsServiceClient, func()) {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	opts, err := cfg.ServerOptions()
	if err != nil {
		t.Fatalf("server options: %v", err)
	}
	s := grpc.NewServer(opts...)
	pb.RegisterFiatRampsServiceServer(s, &pb.UnimplementedFiatRampsServiceServer{})
	go func() { _ = s.Serve(lis) }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		cancel()
		t.Fatalf("dial: %v", err)
	}
	return pb.NewFiatRampsServiceClient(conn), func() {
		cancel()
		_ = conn.Close()
		s.Stop()
		_ = lis.Close()
	}
}

func TestEveryRPCUnauthenticatedWithoutCertOrToken(t *testing.T) {
	client, cleanup := startTestServer(t, auth.Config{ServiceToken: "not-provided-to-client"})
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	calls := map[string]func() error{
		"GetSupportedCountries": func() error {
			_, err := client.GetSupportedCountries(ctx, &pb.GetSupportedCountriesRequest{})
			return err
		},
		"GetQuotes": func() error {
			_, err := client.GetQuotes(ctx, &pb.GetQuotesRequest{})
			return err
		},
		"GetRates": func() error {
			_, err := client.GetRates(ctx, &pb.GetRatesRequest{})
			return err
		},
		"GetTransactionFee": func() error {
			_, err := client.GetTransactionFee(ctx, &pb.GetTransactionFeeRequest{})
			return err
		},
		"GetBeneficiaries": func() error {
			_, err := client.GetBeneficiaries(ctx, &pb.GetBeneficiariesRequest{})
			return err
		},
		"GetBeneficiary": func() error {
			_, err := client.GetBeneficiary(ctx, &pb.GetBeneficiaryRequest{})
			return err
		},
		"GetPaymentMethods": func() error {
			_, err := client.GetPaymentMethods(ctx, &pb.GetPaymentMethodsRequest{})
			return err
		},
		"AddBeneficiary": func() error {
			_, err := client.AddBeneficiary(ctx, &pb.AddBeneficiaryRequest{})
			return err
		},
		"GetPaymentNetworks": func() error {
			_, err := client.GetPaymentNetworks(ctx, &pb.GetPaymentNetworksRequest{})
			return err
		},
		"GetSupportedCountryByCode": func() error {
			_, err := client.GetSupportedCountryByCode(ctx, &pb.GetSupportedCountryRequest{})
			return err
		},
		"CreateBeneficiaryPaymentRequest": func() error {
			_, err := client.CreateBeneficiaryPaymentRequest(ctx, &pb.CreateBeneficiaryPaymentRequestInput{})
			return err
		},
		"GetPaymentRequestTransaction": func() error {
			_, err := client.GetPaymentRequestTransaction(ctx, &pb.GetPaymentRequestTransactionRequest{})
			return err
		},
		"GetHotWalletAddress": func() error {
			_, err := client.GetHotWalletAddress(ctx, &pb.GetHotWalletAddressRequest{})
			return err
		},
		"CreatePlatformCollectionRequest": func() error {
			_, err := client.CreatePlatformCollectionRequest(ctx, &pb.CreatePlatformCollectionRequestInput{})
			return err
		},
		"GetCollectionRequestTransaction": func() error {
			_, err := client.GetCollectionRequestTransaction(ctx, &pb.GetCollectionRequestTransactionRequest{})
			return err
		},
		"UpdateCollectionTransactionStatus": func() error {
			_, err := client.UpdateCollectionTransactionStatus(ctx, &pb.UpdateCollectionTransactionStatusRequest{})
			return err
		},
		"ResolveBankAccount": func() error {
			_, err := client.ResolveBankAccount(ctx, &pb.ResolveBankAccountRequest{})
			return err
		},
		"UpdatePaymentTransactionStatus": func() error {
			_, err := client.UpdatePaymentTransactionStatus(ctx, &pb.UpdatePaymentTransactionStatusRequest{})
			return err
		},
	}

	for _, method := range pb.FiatRampsService_ServiceDesc.Methods {
		fn, ok := calls[method.MethodName]
		if !ok {
			t.Errorf("missing unauthenticated coverage for %s", method.MethodName)
			continue
		}
		err := fn()
		if status.Code(err) != codes.Unauthenticated {
			t.Errorf("%s: got %v, want Unauthenticated", method.MethodName, err)
		}
	}
}

func TestServerStartupFailsWithoutTLSOrAuth(t *testing.T) {
	if _, err := (auth.Config{}).ServerOptions(); err == nil {
		t.Fatal("expected startup to fail when TLS/auth is not configured")
	}
	if _, err := (auth.Config{LocalDev: true}).ServerOptions(); err != nil {
		t.Fatalf("local-dev should start: %v", err)
	}
}
