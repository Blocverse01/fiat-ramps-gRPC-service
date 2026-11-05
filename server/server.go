package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	pb "github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/auth"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/beneficiaries"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/collection"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/countries"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/payment_methods"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/payment_networks"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/payments"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/quotes"
	"google.golang.org/grpc"
)

var (
	port         = flag.Int("port", 50051, "The server port")
	tls          = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile     = flag.String("cert_file", "", "The TLS cert file")
	keyFile      = flag.String("key_file", "", "The TLS key file")
	clientCAFile = flag.String("client_ca_file", "", "The CA used to verify client certificates (mTLS)")
	localDev     = flag.Bool("local_dev", false, "Allow insecure startup for local development only")
	envFile      = flag.String("env_path", ".env", "The path to the .env file")
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

	var conditionalFeePercentConfigs []*pb.ConditionalFeePercentConfig

	for _, fee := range fiatQuotes.Fees.ConditionalFeeConfig() {
		conditionalFeePercentConfigs = append(conditionalFeePercentConfigs, &pb.ConditionalFeePercentConfig{
			Rule:          fee.Rule,
			TriggerAmount: float64(fee.TriggerAmount),
			FeePercentage: fee.FeePercentage,
			RampType:      fee.RampType,
		})
	}

	return &pb.GetQuotesResponse{
		Fees: &pb.GetTransactionFeeResponse{
			OnrampFeePercentage:          fiatQuotes.Fees.OnrampFeePercentage,
			OfframpFeePercentage:         fiatQuotes.Fees.OfframpFeePercentage,
			ConditionalFeePercentConfigs: conditionalFeePercentConfigs,
			OfframpMinFeeUsd:             fiatQuotes.Fees.OfframpMinFeeUsd,
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

	var conditionalFeePercentConfigs []*pb.ConditionalFeePercentConfig

	for _, fee := range fees.ConditionalFeeConfig() {
		conditionalFeePercentConfigs = append(conditionalFeePercentConfigs, &pb.ConditionalFeePercentConfig{
			Rule:          fee.Rule,
			TriggerAmount: float64(fee.TriggerAmount),
			FeePercentage: fee.FeePercentage,
			RampType:      fee.RampType,
		})
	}

	return &pb.GetTransactionFeeResponse{
		OnrampFeePercentage:          fees.OnrampFeePercentage,
		OfframpFeePercentage:         fees.OfframpFeePercentage,
		ConditionalFeePercentConfigs: conditionalFeePercentConfigs,
		OfframpMinFeeUsd:             fees.OfframpMinFeeUsd,
	}, nil
}

func (server *fiatRampsServer) GetBeneficiaries(ctx context.Context, req *pb.GetBeneficiariesRequest) (*pb.GetBeneficiariesResponse, error) {
	ownerId, err := auth.ResolveOwnerID(ctx, req.OwnerId)
	if err != nil {
		return nil, err
	}
	ownerBeneficiaries, err := beneficiaries.GetBeneficiaries(ownerId, req.CountryCode, req.AccountType)
	if err != nil {
		return nil, err
	}

	// Map to store country details (avoids redundant calls)
	countryCache := make(map[string]*pb.SupportedCountry)

	var beneficiariesResponse []*pb.Beneficiary
	for _, beneficiary := range ownerBeneficiaries {
		// Check if country details are already fetched
		if _, exists := countryCache[beneficiary.Country]; !exists {
			country, err := countries.GetCountry(map[string]string{
				"code": beneficiary.Country,
			})
			if err != nil {
				return nil, err
			}
			// Cache the country details
			countryCache[beneficiary.Country] = &pb.SupportedCountry{
				Code:           country.Code,
				Name:           country.Name,
				CurrencySymbol: country.CurrencySymbol,
				CurrencyName:   country.CurrencyName,
				Id:             country.Id,
				Flag:           country.Flag,
			}
		}

		beneficiariesResponse = append(beneficiariesResponse, &pb.Beneficiary{
			Id:                   beneficiary.Id,
			Country:              countryCache[beneficiary.Country],
			AccountType:          beneficiary.AccountType,
			AccountName:          beneficiary.AccountName,
			AccountNumber:        beneficiary.AccountNumber,
			ProviderName:         beneficiary.NetworkName,
			NetworkId:            beneficiary.NetworkId,
			ChannelId:            beneficiary.ChannelId,
			NameEnquiryReference: beneficiary.NameEnquiryReference,
			SeerbitId:            beneficiary.SeerBitId,
		})
	}

	return &pb.GetBeneficiariesResponse{
		Beneficiaries: beneficiariesResponse,
	}, nil
}

func (server *fiatRampsServer) GetBeneficiary(ctx context.Context, req *pb.GetBeneficiaryRequest) (*pb.Beneficiary, error) {
	ownerId, err := auth.ResolveOwnerID(ctx, req.GetOwnerId())
	if err != nil {
		return nil, err
	}
	beneficiary, err := beneficiaries.GetBeneficiaryForOwner(req.BeneficiaryId, ownerId)

	if err != nil {
		return nil, err
	}

	country, err := countries.GetCountry(map[string]string{
		"code": beneficiary.Country,
	})

	if err != nil {
		return nil, err
	}

	return &pb.Beneficiary{
		Id: req.BeneficiaryId,
		Country: &pb.SupportedCountry{
			Code:           country.Code,
			Name:           country.Name,
			CurrencySymbol: country.CurrencySymbol,
			CurrencyName:   country.CurrencyName,
			Id:             country.Id,
			Flag:           country.Flag,
		},
		AccountType:          beneficiary.AccountType,
		AccountName:          beneficiary.AccountName,
		AccountNumber:        beneficiary.AccountNumber,
		ProviderName:         beneficiary.NetworkName,
		NetworkId:            beneficiary.NetworkId,
		ChannelId:            beneficiary.ChannelId,
		NameEnquiryReference: beneficiary.NameEnquiryReference,
		SeerbitId:            beneficiary.SeerBitId,
	}, nil
}

func (server *fiatRampsServer) GetSupportedCountryByCode(_ context.Context, req *pb.GetSupportedCountryRequest) (*pb.SupportedCountry, error) {
	country, err := countries.GetCountry(map[string]string{
		"code": req.CountryCode,
	})

	if err != nil {
		return nil, err
	}

	return &pb.SupportedCountry{
		Id:             country.Id,
		Code:           country.Code,
		Name:           country.Name,
		CurrencySymbol: country.CurrencySymbol,
		CurrencyName:   country.CurrencyName,
		Flag:           country.Flag,
	}, nil
}

func (server *fiatRampsServer) AddBeneficiary(ctx context.Context, req *pb.AddBeneficiaryRequest) (*pb.AddBeneficiaryResponse, error) {
	ownerId, err := auth.ResolveOwnerID(ctx, req.OwnerId)
	if err != nil {
		return nil, err
	}
	newBeneficiaryId, err := beneficiaries.AddBeneficiary(beneficiaries.AddBeneficiaryRequestBody{
		AccountNumber:        req.AccountNumber,
		AccountType:          req.AccountType,
		AccountName:          req.AccountName,
		ChannelId:            req.ChannelId,
		NetworkId:            req.NetworkId,
		Country:              req.Country,
		OwnerId:              ownerId,
		NetworkName:          req.NetworkName,
		NameEnquiryReference: req.NameEnquiryReference,
		SeerBitId:            req.SeerbitId,
	})

	if err != nil {
		return nil, err
	}

	return &pb.AddBeneficiaryResponse{
		NewBeneficiaryId: newBeneficiaryId,
	}, nil
}

func (server *fiatRampsServer) GetPaymentMethods(_ context.Context, req *pb.GetPaymentMethodsRequest) (*pb.GetPaymentMethodsResponse, error) {
	paymentMethods, err := payment_methods.GetPaymentMethods(req.CountryCode, req.Route)

	if err != nil {
		return nil, err
	}

	var paymentChannels []*pb.PaymentChannel

	for _, channel := range paymentMethods.PaymentMethods {
		paymentChannels = append(paymentChannels, &pb.PaymentChannel{
			ChannelName: channel.ChannelName,
			ChannelId:   channel.ChannelId,
		})
	}

	return &pb.GetPaymentMethodsResponse{
		Channels: paymentChannels,
		Limits: &pb.CountryFiatLimits{
			Min: float32(paymentMethods.Limits.Min),
			Max: float32(paymentMethods.Limits.Max),
		},
	}, nil
}

func (server *fiatRampsServer) GetPaymentNetworks(_ context.Context, req *pb.GetPaymentNetworksRequest) (*pb.GetPaymentNetworksResponse, error) {
	paymentNetworks, err := payment_networks.GetPaymentNetworks(req.CountryCode, req.ChannelId)

	if err != nil {
		return nil, err
	}

	var activePaymentNetworks []*pb.ActivePaymentNetwork

	for _, network := range paymentNetworks {
		activePaymentNetworks = append(activePaymentNetworks, &pb.ActivePaymentNetwork{
			NetworkId: network.NetworkId,
			Name:      network.Name,
			ChannelId: network.ChannelId,
			SeerbitId: network.SeerBitId,
		})
	}

	return &pb.GetPaymentNetworksResponse{
		Networks: activePaymentNetworks,
	}, nil
}

func (server *fiatRampsServer) CreateBeneficiaryPaymentRequest(ctx context.Context, req *pb.CreateBeneficiaryPaymentRequestInput) (*pb.CreatePaymentRequestResponse, error) {
	ownerId, err := auth.ResolveOwnerID(ctx, req.OwnerId)
	if err != nil {
		return nil, err
	}
	paymentRequestResponse, err := payments.CreateBeneficiaryPaymentRequest(payments.BeneficiaryPaymentRequestInput{
		BeneficiaryId: req.BeneficiaryId,
		LocalAmount:   req.LocalAmount,
		OwnerId:       ownerId,
		Sender: yellowcard.RetailCustomer{
			Name:               req.Sender.Name,
			Phone:              req.Sender.Phone,
			Country:            req.Sender.Country,
			Address:            req.Sender.Address,
			Dob:                req.Sender.Dob,
			IdType:             req.Sender.IdType,
			IdNumber:           req.Sender.IdNumber,
			AdditionalIdNumber: req.Sender.AdditionalIdNumber,
			AdditionalIdType:   req.Sender.AdditionalIdType,
		},
		TokenSymbol:           req.TokenSymbol,
		TokenAddress:          req.TokenAddress,
		TransactionHash:       req.TransactionHash,
		BlockchainNetwork:     req.BlockchainNetwork,
		HotWalletAddress:      req.HotWalletAddress,
		FromWalletAddress:     req.FromWalletAddress,
		PlatformFeePercentage: req.PlatformFeePercentage,
		ConversionRate:        req.ConversionRate,
		UsdAmount:             req.UsdAmount,
		TransactionFeeUsd:     req.TransactionFeeUsd,
	})

	if err != nil {
		return nil, err
	}

	return &pb.CreatePaymentRequestResponse{
		Id:              paymentRequestResponse.Id,
		ChannelId:       paymentRequestResponse.ChannelId,
		SequenceId:      paymentRequestResponse.SequenceId,
		Currency:        paymentRequestResponse.Currency,
		Country:         paymentRequestResponse.Country,
		Amount:          paymentRequestResponse.Amount,
		Reason:          paymentRequestResponse.Reason,
		ConvertedAmount: paymentRequestResponse.ConvertedAmount,
		Status:          paymentRequestResponse.Status,
		Rate:            paymentRequestResponse.Rate,
		Sender: &pb.YellowCardRetailCustomer{
			Name:               paymentRequestResponse.Sender.Name,
			Phone:              paymentRequestResponse.Sender.Phone,
			Country:            paymentRequestResponse.Sender.Country,
			Address:            paymentRequestResponse.Sender.Address,
			Dob:                paymentRequestResponse.Sender.Dob,
			IdNumber:           paymentRequestResponse.Sender.IdNumber,
			IdType:             paymentRequestResponse.Sender.IdType,
			AdditionalIdNumber: paymentRequestResponse.Sender.AdditionalIdNumber,
			AdditionalIdType:   paymentRequestResponse.Sender.AdditionalIdType,
		},
		Destination: &pb.PaymentRequestDestination{
			AccountNumber: paymentRequestResponse.Destination.AccountNumber,
			AccountType:   paymentRequestResponse.Destination.AccountType,
			NetworkId:     paymentRequestResponse.Destination.NetworkId,
			NetworkName:   paymentRequestResponse.Destination.NetworkName,
			Country:       paymentRequestResponse.Destination.Country,
			AccountName:   paymentRequestResponse.Destination.AccountName,
		},
	}, nil
}

func (server *fiatRampsServer) CreatePlatformCollectionRequest(ctx context.Context, req *pb.CreatePlatformCollectionRequestInput) (*pb.CreateCollectionResponse, error) {
	ownerId, err := auth.ResolveOwnerID(ctx, req.OwnerId)
	if err != nil {
		return nil, err
	}
	collectionRequestResponse, err := collection.CreatePlatformCollectionRequest(collection.PlatformCollectionRequestPayload{
		LocalAmount: req.LocalAmount,
		OwnerId:     ownerId,
		Recipient: yellowcard.RetailCustomer{
			Name:               req.Recipient.Name,
			Phone:              req.Recipient.Phone,
			Country:            req.Recipient.Country,
			Address:            req.Recipient.Address,
			Dob:                req.Recipient.Dob,
			IdType:             req.Recipient.IdType,
			IdNumber:           req.Recipient.IdNumber,
			AdditionalIdNumber: req.Recipient.AdditionalIdNumber,
			AdditionalIdType:   req.Recipient.AdditionalIdType,
		},
		Source: collection.Source{
			AccountNumber: req.Source.AccountNumber,
			AccountType:   req.Source.AccountType,
			NetworkId:     req.Source.NetworkId,
		},
		TokenSymbol:              req.TokenSymbol,
		BlockchainNetwork:        req.BlockchainNetwork,
		DestinationWalletAddress: req.DestinationWalletAddress,
		PlatformFeePercentage:    req.PlatformFeePercentage,
		ConversionRate:           req.ConversionRate,
		UsdAmount:                req.UsdAmount,
		ChannelId:                req.ChannelId,
		TransactionFeeUsd:        req.TransactionFeeUsd,
		CollectionCountryCode:    req.CollectionCountryCode,
		IsAgentTransaction:       req.IsAgentTransaction,
	})

	if err != nil {
		return nil, err
	}

	return &pb.CreateCollectionResponse{
		Id:              collectionRequestResponse.Id,
		ChannelId:       collectionRequestResponse.ChannelId,
		SequenceId:      collectionRequestResponse.SequenceId,
		Currency:        collectionRequestResponse.Currency,
		Country:         collectionRequestResponse.Country,
		Amount:          collectionRequestResponse.Amount,
		ConvertedAmount: collectionRequestResponse.ConvertedAmount,
		Status:          collectionRequestResponse.Status,
		Rate:            collectionRequestResponse.Rate,
		Recipient: &pb.YellowCardRetailCustomer{
			Name:               collectionRequestResponse.Recipient.Name,
			Phone:              collectionRequestResponse.Recipient.Phone,
			Country:            collectionRequestResponse.Recipient.Country,
			Address:            collectionRequestResponse.Recipient.Address,
			Dob:                collectionRequestResponse.Recipient.Dob,
			IdNumber:           collectionRequestResponse.Recipient.IdNumber,
			IdType:             collectionRequestResponse.Recipient.IdType,
			AdditionalIdNumber: collectionRequestResponse.Recipient.AdditionalIdNumber,
			AdditionalIdType:   collectionRequestResponse.Recipient.AdditionalIdType,
		},
		Source: &pb.CollectionSource{
			AccountNumber: collectionRequestResponse.Source.AccountNumber,
			AccountType:   collectionRequestResponse.Source.AccountType,
			NetworkId:     collectionRequestResponse.Source.NetworkId,
		},
		BankInfo: &pb.BankInfo{
			AccountNumber: collectionRequestResponse.BankInfo.AccountNumber,
			AccountName:   collectionRequestResponse.BankInfo.AccountName,
			Name:          collectionRequestResponse.BankInfo.Name,
		},
	}, nil
}

func (server *fiatRampsServer) GetPaymentRequestTransaction(ctx context.Context, req *pb.GetPaymentRequestTransactionRequest) (*pb.GetPaymentRequestTransactionResponse, error) {
	ownerId, err := auth.ResolveOwnerID(ctx, req.GetOwnerId())
	if err != nil {
		return nil, err
	}
	paymentRequest, err := payments.GetPaymentRequestTransactionForOwner(req.SequenceId, ownerId)

	if err != nil {
		return nil, err
	}

	country, err := countries.GetCountry(map[string]string{
		"code": paymentRequest.Beneficiary.Country,
	})

	if err != nil {
		return nil, err
	}

	return &pb.GetPaymentRequestTransactionResponse{
		SequenceId:          paymentRequest.SequenceId,
		Status:              paymentRequest.YellowCardStatus,
		Rate:                paymentRequest.ConversionRate,
		UsdAmount:           paymentRequest.UsdAmount,
		LocalAmount:         paymentRequest.LocalAmount,
		OwnerId:             paymentRequest.OwnerId,
		TokenSymbol:         paymentRequest.TokenSymbol,
		BlockchainNetwork:   paymentRequest.BlockchainNetwork,
		LocalCurrencySymbol: country.CurrencySymbol,
		TransactionHash:     paymentRequest.TransactionHash,
		TransactionFeeUsd:   paymentRequest.TransactionFeeUsd,
		Destination: &pb.PaymentRequestDestination{
			AccountNumber: paymentRequest.Beneficiary.AccountNumber,
			AccountType:   paymentRequest.Beneficiary.AccountType,
			NetworkId:     paymentRequest.Beneficiary.NetworkId,
			NetworkName:   paymentRequest.Beneficiary.NetworkName,
			Country:       paymentRequest.Beneficiary.Country,
			AccountName:   paymentRequest.Beneficiary.AccountName,
		},
		CreatedAt:             paymentRequest.Xata.CreatedAt,
		HotWalletAddress:      paymentRequest.HotWalletAddress,
		FromWalletAddress:     paymentRequest.FromWalletAddress,
		PlatformFeePercentage: paymentRequest.PlatformFeePercentage,
	}, nil
}

func (server *fiatRampsServer) GetCollectionRequestTransaction(ctx context.Context, req *pb.GetCollectionRequestTransactionRequest) (*pb.GetCollectionRequestResponse, error) {
	ownerId, err := auth.ResolveOwnerID(ctx, req.GetOwnerId())
	if err != nil {
		return nil, err
	}
	collectionRequest, err := collection.GetCollectionRequestTransactionForOwner(req.SequenceId, ownerId)

	if err != nil {
		return nil, err
	}

	country, err := countries.GetCountry(map[string]string{
		"code": collectionRequest.CollectionCountry,
	})

	if err != nil {
		return nil, err
	}

	return &pb.GetCollectionRequestResponse{
		SequenceId:               collectionRequest.SequenceId,
		Status:                   collectionRequest.YellowCardStatus,
		ConversionRate:           collectionRequest.ConversionRate,
		UsdAmount:                collectionRequest.UsdAmount,
		LocalAmount:              collectionRequest.LocalAmount,
		OwnerId:                  collectionRequest.OwnerId,
		TokenSymbol:              collectionRequest.TokenSymbol,
		BlockchainNetwork:        collectionRequest.BlockchainNetwork,
		LocalCurrencySymbol:      country.CurrencySymbol,
		DestinationWalletAddress: collectionRequest.DestinationWalletAddress,
		CollectionCountry:        collectionRequest.CollectionCountry,
		TransactionFeeUsd:        collectionRequest.TransactionFeeUsd,
		IsAgentTransaction:       collectionRequest.IsAgentTransaction,
		CreatedAt:                collectionRequest.Xata.CreatedAt,
	}, nil
}

func (server *fiatRampsServer) GetHotWalletAddress(_ context.Context, req *pb.GetHotWalletAddressRequest) (*pb.GetHotWalletAddressResponse, error) {
	hotWalletAddress, err := payments.GetHotWalletAddress(req.NetworkType)

	if err != nil {
		return nil, err
	}

	return &pb.GetHotWalletAddressResponse{
		Address: hotWalletAddress,
	}, nil
}

func (server *fiatRampsServer) UpdateCollectionTransactionStatus(_ context.Context, req *pb.UpdateCollectionTransactionStatusRequest) (*pb.UpdateCollectionTransactionStatusResponse, error) {
	err := collection.UpdateCollectionRequestStatus(req.SequenceId, req.Status)

	if err != nil {
		return nil, err
	}

	return &pb.UpdateCollectionTransactionStatusResponse{
		Status:     req.Status,
		Success:    true,
		SequenceId: req.SequenceId,
	}, nil
}

func (server *fiatRampsServer) UpdatePaymentTransactionStatus(_ context.Context, req *pb.UpdatePaymentTransactionStatusRequest) (*pb.UpdatePaymentTransactionStatusResponse, error) {
	err := payments.UpdatePaymentRequestStatus(req.SequenceId, req.Status)

	if err != nil {
		return nil, err
	}

	return &pb.UpdatePaymentTransactionStatusResponse{
		Status:     req.Status,
		Success:    true,
		SequenceId: req.SequenceId,
	}, nil
}

func (server *fiatRampsServer) ResolveBankAccount(_ context.Context, req *pb.ResolveBankAccountRequest) (*pb.ResolveBankAccountResponse, error) {
	response, err := payments.ResolveBankAccount(payments.ResolveBankAccountInput{
		AccountNumber: req.AccountNumber,
		NetworkId:     req.NetworkId,
	})

	if err != nil {
		return nil, err
	}

	return &pb.ResolveBankAccountResponse{
		AccountName:   response.AccountName,
		AccountNumber: response.AccountNumber,
		AccountBank:   response.AccountBank,
	}, nil
}

func listenAddr() string {
	listenPort := *port
	if p := os.Getenv("PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			listenPort = n
		}
	}
	return fmt.Sprintf("0.0.0.0:%d", listenPort)
}

func main() {
	flag.Parse()

	val, ok := os.LookupEnv("DEPLOYED_ON_RENDER")
	if !ok || val != "true" {
		pb.LoadEnv(*envFile)
	}

	cfg := auth.LoadConfig(*localDev, *tls, *certFile, *keyFile, *clientCAFile)
	opts, err := cfg.ServerOptions()
	if err != nil {
		log.Fatalf("server startup failed: %v", err)
	}

	lis, err := net.Listen("tcp", listenAddr())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(opts...)
	server := &fiatRampsServer{}
	pb.RegisterFiatRampsServiceServer(grpcServer, server)

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
