package main

import (
	"context"
	"flag"
	pb "github.com/blocverse01/globachain-ramps-grpc-service/fiatramps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverAddr         = flag.String("addr", "localhost:50051", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.example.com", "The server name used to verify the hostname returned by the TLS handshake")
)

func printSupportedCountries(client pb.FiatRampsServiceClient) {
	log.Printf("Getting supported countries")
	req := &pb.GetSupportedCountriesRequest{}
	ctx := context.Background()
	startReqTime := time.Now()
	res, err := client.GetSupportedCountries(ctx, req)
	endReqTime := time.Now()

	timeTaken := endReqTime.Sub(startReqTime)

	log.Printf("Time taken to get supported countries: %v", timeTaken)

	if err != nil {
		log.Fatalf("Failed to get supported countries: %v", err)
	}

	log.Printf("Supported countries: %v", res.Countries)
}

func printTransactionFee(client pb.FiatRampsServiceClient) {
	log.Printf("Getting transaction fee")
	req := &pb.GetTransactionFeeRequest{
		CountryCode: "NG",
	}
	ctx := context.Background()
	startReqTime := time.Now()
	res, err := client.GetTransactionFee(ctx, req)
	endReqTime := time.Now()

	timeTaken := endReqTime.Sub(startReqTime)

	log.Printf("Time taken to get transaction fee: %v", timeTaken)

	if err != nil {
		log.Fatalf("Failed to get transaction fee: %v", err)
	}

	log.Printf("Transaction Fees:\nOnramp:%f, Offramp:%f", res.OnrampFeePercentage, res.OfframpFeePercentage)
}

func printRates(client pb.FiatRampsServiceClient) {
	log.Printf("Getting rates")
	req := &pb.GetRatesRequest{
		TargetCurrencies: []string{"NGN"},
	}
	ctx := context.Background()
	startReqTime := time.Now()
	res, err := client.GetRates(ctx, req)
	endReqTime := time.Now()

	timeTaken := endReqTime.Sub(startReqTime)

	log.Printf("Time taken to get rates: %v", timeTaken)

	if err != nil {
		log.Fatalf("Failed to get rates: %v", err)
	}

	log.Printf("Rates: %v", res.Rates)
}

func printQuotes(client pb.FiatRampsServiceClient) {
	log.Printf("Getting quotes")
	req := &pb.GetQuotesRequest{
		CountryCode:    "NG",
		CurrencySymbol: "NGN",
	}
	ctx := context.Background()
	startReqTime := time.Now()
	res, err := client.GetQuotes(ctx, req)
	endReqTime := time.Now()

	timeTaken := endReqTime.Sub(startReqTime)

	log.Printf("Time taken to get quotes: %v", timeTaken)

	if err != nil {
		log.Fatalf("Failed to get quotes: %v", err)
	}

	log.Printf("Quotes: %v", res)
}

func printBeneficiaries(client pb.FiatRampsServiceClient) {
	log.Printf("Getting beneficiaries")
	req := &pb.GetBeneficiariesRequest{
		CountryCode: "NG",
		OwnerId:     "2348143100808",
	}
	ctx := context.Background()
	res, err := client.GetBeneficiaries(ctx, req)

	if err != nil {
		log.Fatalf("Failed to get beneficiaries: %v", err)
	}

	log.Printf("Beneficiaries: %v", res)
}

func main() {
	flag.Parse()

	var opts []grpc.DialOption

	if *tls {
		if *caFile == "" {
			log.Fatalf("cert_file and key_file must be set when tls is enabled")
		}
		creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostOverride)
		if err != nil {
			log.Fatalf("Failed to create TLS credentials: %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("Failed to close connection: %v", err)
		}
	}(conn)

	client := pb.NewFiatRampsServiceClient(conn)

	printSupportedCountries(client)
	printTransactionFee(client)

	if *serverAddr != "localhost:50051" {
		printRates(client)
		printQuotes(client)
	}

	printBeneficiaries(client)
}
