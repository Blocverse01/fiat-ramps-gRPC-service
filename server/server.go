package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/blocverse01/globachain-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/globachain-ramps-grpc-service/fiatramps/yellowcard/beneficiaries"
	"github.com/blocverse01/globachain-ramps-grpc-service/fiatramps/yellowcard/countries"
	"github.com/blocverse01/globachain-ramps-grpc-service/fiatramps/yellowcard/quotes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"os"
)

var (
	port     = flag.Int("port", 50051, "The server port")
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert_file", "", "The TLS cert file")
	keyFile  = flag.String("key_file", "", "The TLS key file")
	envFile  = flag.String("env_path", ".env", "The path to the .env file")
)

type fiatRampsServer struct {
	pb.UnimplementedFiatRampsServiceServer
}

func (server *fiatRampsServer) GetSupportedCountries(_ context.Context, _ *pb.GetSupportedCountriesRequest) (*pb.GetSupportedCountriesResponse, error) {
	supportedCountries := countries.GetSupportedCountries()

	return &pb.GetSupportedCountriesResponse{Countries: supportedCountries}, nil
}

func (server *fiatRampsServer) GetQuotes(_ context.Context, req *pb.GetQuotesRequest) (*pb.GetQuotesResponse, error) {
	// Get the quotes
	fiatQuotes := quotes.GetQuotes(quotes.FiatConfig{
		CountryCode:    req.CountryCode,
		CurrencySymbol: req.CurrencySymbol,
	})

	return &pb.GetQuotesResponse{
		Fees: &pb.GetTransactionFeeResponse{
			OnrampFeePercentage:  fiatQuotes.Fees.OnrampFeePercentage,
			OfframpFeePercentage: fiatQuotes.Fees.OfframpFeePercentage,
		},
		Rate: &pb.Rate{
			Buy:    fiatQuotes.Rate.Buy,
			Sell:   fiatQuotes.Rate.Sell,
			Locale: fiatQuotes.Rate.Locale,
			Code:   fiatQuotes.Rate.Code,
		},
	}, nil
}

func (server *fiatRampsServer) GetRates(_ context.Context, req *pb.GetRatesRequest) (*pb.GetRatesResponse, error) {
	// Get all rates
	allRates := quotes.GetRates()
	var filteredRates []quotes.Rate

	if len(req.TargetCurrencies) > 0 {
		filteredRates = quotes.FilterRatesByCurrencies(allRates, req.TargetCurrencies)
	} else {
		filteredRates = allRates
	}

	// Convert filtered rates to the format expected by the response
	var ratesResponse []*pb.Rate
	for _, rate := range filteredRates {
		ratesResponse = append(ratesResponse, &pb.Rate{
			Buy:    rate.Buy,
			Sell:   rate.Sell,
			Locale: rate.Locale,
			Code:   rate.Code,
		})
	}

	return &pb.GetRatesResponse{
		Rates: ratesResponse,
	}, nil
}

func (server *fiatRampsServer) GetTransactionFee(_ context.Context, req *pb.GetTransactionFeeRequest) (*pb.GetTransactionFeeResponse, error) {
	fees, err := quotes.GetTransactionFee(req.CountryCode)

	if err != nil {
		return nil, err
	}

	return &pb.GetTransactionFeeResponse{
		OnrampFeePercentage:  fees.OnrampFeePercentage,
		OfframpFeePercentage: fees.OfframpFeePercentage,
	}, nil
}

func (server *fiatRampsServer) GetBeneficiaries(_ context.Context, req *pb.GetBeneficiariesRequest) (*pb.GetBeneficiariesResponse, error) {
	ownerBeneficiaries := beneficiaries.GetBeneficiaries(req.OwnerId, req.CountryCode, req.AccountType)

	var beneficiariesResponse []*pb.Beneficiary
	for _, beneficiary := range ownerBeneficiaries {
		var accountName string
		var accountNumber string
		var providerName string

		if beneficiary.AccountType == "bank" {
			accountName = beneficiary.AccountName
			accountNumber = beneficiary.AccountNumber
			providerName = beneficiary.BankName
		} else {
			accountName = fmt.Sprintf("%s %s", beneficiary.FirstName, beneficiary.LastName)
			accountNumber = beneficiary.MobileNumber
			providerName = beneficiary.MobileProvider
		}
		beneficiariesResponse = append(beneficiariesResponse, &pb.Beneficiary{
			Id: beneficiary.Id,
			Country: &pb.SupportedCountry{
				Code:           beneficiary.Country.Code,
				Country:        beneficiary.Country.Country,
				CurrencySymbol: beneficiary.Country.CurrencySymbol,
				CurrencyName:   beneficiary.Country.CurrencyName,
				Id:             beneficiary.Country.Id,
				Flag:           beneficiary.Country.Flag,
			},
			AccountType:         beneficiary.AccountType,
			AccountName:         accountName,
			AccountNumber:       accountNumber,
			ProviderName:        providerName,
			NetworkId:           beneficiary.NetworkId,
			ChannelId:           beneficiary.ChannelId,
			ProviderDescription: beneficiary.MobileProviderCode,
		})
	}

	return &pb.GetBeneficiariesResponse{
		Beneficiaries: beneficiariesResponse,
	}, nil
}

func main() {
	// Parse command-line flags
	flag.Parse()

	val, ok := os.LookupEnv("DEPLOYED_ON_RENDER")
	if !ok || val != "true" {
		pb.LoadEnv(*envFile)
	}

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	if *tls {
		if *certFile == "" || *keyFile == "" {
			log.Fatalf("cert_file and key_file must be set when tls is enabled")
		}
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("failed to load TLS credentials: %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	grpcServer := grpc.NewServer(opts...)
	server := &fiatRampsServer{}
	pb.RegisterFiatRampsServiceServer(grpcServer, server)

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
