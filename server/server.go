package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/beneficiaries"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/collection"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/countries"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/payment_methods"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/payment_networks"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/payments"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/quotes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

func (server *fiatRampsServer) GetBeneficiaries(_ context.Context, req *pb.GetBeneficiariesRequest) (*pb.GetBeneficiariesResponse, error) {
	ownerBeneficiaries, err := beneficiaries.GetBeneficiaries(req.OwnerId, req.CountryCode, req.AccountType)

	if err != nil {
		return nil, err
	}

	var beneficiariesResponse []*pb.Beneficiary
	for _, beneficiary := range ownerBeneficiaries {
		country, err := countries.GetCountry(map[string]string{
			"code": beneficiary.Country,
		})
		if err != nil {
			return nil, err
		}
		beneficiariesResponse = append(beneficiariesResponse, &pb.Beneficiary{
			Id: beneficiary.Id,
			Country: &pb.SupportedCountry{
				Code:           country.Code,
				Name:           country.Name,
				CurrencySymbol: country.CurrencySymbol,
				CurrencyName:   country.CurrencyName,
				Id:             country.Id,
				Flag:           country.Flag,
			},
			AccountType:   beneficiary.AccountType,
			AccountName:   beneficiary.AccountName,
			AccountNumber: beneficiary.AccountNumber,
			ProviderName:  beneficiary.NetworkName,
			NetworkId:     beneficiary.NetworkId,
			ChannelId:     beneficiary.ChannelId,
		})
	}

	return &pb.GetBeneficiariesResponse{
		Beneficiaries: beneficiariesResponse,
	}, nil
}

func (server *fiatRampsServer) GetBeneficiary(_ context.Context, req *pb.GetBeneficiaryRequest) (*pb.Beneficiary, error) {
	beneficiary, err := beneficiaries.GetBeneficiary(req.BeneficiaryId)

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
		AccountType:   beneficiary.AccountType,
		AccountName:   beneficiary.AccountName,
		AccountNumber: beneficiary.AccountNumber,
		ProviderName:  beneficiary.NetworkName,
		NetworkId:     beneficiary.NetworkId,
		ChannelId:     beneficiary.ChannelId,
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

func (server *fiatRampsServer) AddBeneficiary(_ context.Context, req *pb.AddBeneficiaryRequest) (*pb.AddBeneficiaryResponse, error) {
	newBeneficiaryId, err := beneficiaries.AddBeneficiary(beneficiaries.AddBeneficiaryRequestBody{
		AccountNumber: req.AccountNumber,
		AccountType:   req.AccountType,
		AccountName:   req.AccountName,
		ChannelId:     req.ChannelId,
		NetworkId:     req.NetworkId,
		Country:       req.Country,
		OwnerId:       req.OwnerId,
		NetworkName:   req.NetworkName,
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
		})
	}

	return &pb.GetPaymentNetworksResponse{
		Networks: activePaymentNetworks,
	}, nil
}

func (server *fiatRampsServer) CreateBeneficiaryPaymentRequest(_ context.Context, req *pb.CreateBeneficiaryPaymentRequestInput) (*pb.CreatePaymentRequestResponse, error) {
	paymentRequestResponse, err := payments.CreateBeneficiaryPaymentRequest(payments.BeneficiaryPaymentRequestInput{
		BeneficiaryId: req.BeneficiaryId,
		LocalAmount:   req.LocalAmount,
		OwnerId:       req.OwnerId,
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

func (server *fiatRampsServer) CreatePlatformCollectionRequest(_ context.Context, req *pb.CreatePlatformCollectionRequestInput) (*pb.CreateCollectionResponse, error) {
	collectionRequestResponse, err := collection.CreatePlatformCollectionRequest(collection.PlatformCollectionRequestPayload{
		LocalAmount: req.LocalAmount,
		OwnerId:     req.OwnerId,
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

func (server *fiatRampsServer) GetPaymentRequestTransaction(_ context.Context, req *pb.GetPaymentRequestTransactionRequest) (*pb.GetPaymentRequestTransactionResponse, error) {
	paymentRequest, err := payments.GetPaymentRequestTransaction(req.SequenceId)

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

func (server *fiatRampsServer) GetCollectionRequestTransaction(_ context.Context, req *pb.GetCollectionRequestTransactionRequest) (*pb.GetCollectionRequestResponse, error) {
	collectionRequest, err := collection.GetCollectionRequestTransaction(req.SequenceId)

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
