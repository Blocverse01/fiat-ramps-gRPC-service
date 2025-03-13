package collection

import (
	"fmt"

	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
)

type PlatformCollectionRequestPayload struct {
	OwnerId                  string                    `json:"ownerId"`
	Source                   Source                    `json:"source"`
	Recipient                yellowcard.RetailCustomer `json:"recipient"`
	TokenSymbol              string                    `json:"tokenSymbol"`
	BlockchainNetwork        string                    `json:"blockchainNetwork"`
	DestinationWalletAddress string                    `json:"fromWalletAddress"`
	PlatformFeePercentage    float64                   `json:"platformFeePercentage"`
	ConversionRate           float64                   `json:"conversionRate"`
	UsdAmount                float64                   `json:"usdAmount"` // Total USD amount including transaction fees.
	LocalAmount              float64                   `json:"localAmount"`
	ChannelId                string                    `json:"channelId"`
	TransactionFeeUsd        float64                   `json:"transactionFeeUsd"`
	CollectionCountryCode    string                    `json:"collectionCountryCode"`
}

func (p PlatformCollectionRequestPayload) Validate() error {
	if p.OwnerId == "" {
		return fmt.Errorf("OwnerId is required")
	}
	if p.TokenSymbol == "" {
		return fmt.Errorf("TokenSymbol is required")
	}
	if p.BlockchainNetwork == "" {
		return fmt.Errorf("BlockchainNetwork is required")
	}
	if p.DestinationWalletAddress == "" {
		return fmt.Errorf("DestinationWalletAddress is required")
	}
	if p.PlatformFeePercentage < 0 {
		return fmt.Errorf("PlatformFeePercentage must be greater than zero")
	}
	if p.ConversionRate <= 0 {
		return fmt.Errorf("ConversionRate must be greater than zero")
	}
	if p.UsdAmount <= 0 {
		return fmt.Errorf("UsdAmount must be greater than zero")
	}
	if p.LocalAmount <= 0 {
		return fmt.Errorf("LocalAmount must be greater than zero")
	}
	if p.ChannelId == "" {
		return fmt.Errorf("ChannelId is required")
	}
	if p.CollectionCountryCode == "" {
		return fmt.Errorf("CollectionCountryCode is required")
	}

	// Todo Test this

	//localAmount := decimal.NewFromFloat(p.LocalAmount).Round(2)
	//inputUsdAmount := decimal.NewFromFloat(p.UsdAmount).Round(2)
	//
	//conversionRate := decimal.NewFromFloat(p.ConversionRate)
	//convertedUsdAmount := localAmount.Div(conversionRate)
	//platformFeePercentage := decimal.NewFromFloat(p.PlatformFeePercentage).Div(yellowcard.Percent)
	//
	//expectedUsdAmount := convertedUsdAmount.Sub(convertedUsdAmount.Mul(platformFeePercentage)).Round(2)
	//
	//
	//if !expectedUsdAmount.Equal(inputUsdAmount) {
	//	return fmt.Errorf("UsdAmount %s should be equal to expectedUsdAmount %s", inputUsdAmount.String(), expectedUsdAmount.String())
	//}

	return nil
}

const sequenceIdPrefix = "azza-c"

func CreatePlatformCollectionRequest(input PlatformCollectionRequestPayload) (CreateCollectionResponse, error) {
	if err := input.Validate(); err != nil {
		return CreateCollectionResponse{}, err
	}

	sequenceId, err := yellowcard.GenerateCollectionSequenceId(sequenceIdPrefix)

	if err != nil {
		return CreateCollectionResponse{}, err
	}

	storedRequestId, err := StoreCollectionRequest(StoreCollectionRequestInput{
		OwnerId:                  input.OwnerId,
		CollectionAccountType:    input.Source.AccountType,
		CollectionCountry:        input.CollectionCountryCode,
		ChannelId:                input.ChannelId,
		SequenceId:               sequenceId,
		BlockchainNetwork:        input.BlockchainNetwork,
		TokenSymbol:              input.TokenSymbol,
		ConversionRate:           input.ConversionRate,
		UsdAmount:                input.UsdAmount,
		LocalAmount:              input.LocalAmount,
		PlatformFeePercentage:    input.PlatformFeePercentage,
		DestinationWalletAddress: input.DestinationWalletAddress,
		NetworkId:                input.Source.NetworkId,
		YellowCardStatus:         "recorded",
		TransactionFeeUsd:        input.TransactionFeeUsd,
	})

	if err != nil {
		return CreateCollectionResponse{}, fmt.Errorf("failed to store collection request: %w", err)
	}

	fmt.Printf("Stored Collection Request ID: %s", storedRequestId)

	collectionRequest, err := CreateCollectionRequest(RequestPayload{
		Source:       input.Source,
		Recipient:    input.Recipient,
		LocalAmount:  input.LocalAmount,
		SequenceId:   sequenceId,
		ChannelId:    input.ChannelId,
		ForceAccept:  true,
		CustomerType: "retail",
	})

	if err != nil {
		return CreateCollectionResponse{}, err
	}

	UpdateCollectionRequestAccountDetails(sequenceId, CollectAccountDetails{
		CollectionAccountNumber: collectionRequest.BankInfo.AccountNumber,
		CollectionBankName:      collectionRequest.BankInfo.Name,
		CollectionAccountName:   collectionRequest.BankInfo.AccountName,
	})

	return collectionRequest, nil
}
