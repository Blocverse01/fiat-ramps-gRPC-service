package payments

import (
	"fmt"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/beneficiaries"
	"github.com/shopspring/decimal"
	"unicode"
)

type BeneficiaryPaymentRequestInput struct {
	BeneficiaryId         string                    `json:"beneficiaryId"`
	LocalAmount           float64                   `json:"localAmount"`
	OwnerId               string                    `json:"ownerId"`
	Sender                yellowcard.RetailCustomer `json:"sender"`
	TokenSymbol           string                    `json:"tokenSymbol"`
	BlockchainNetwork     string                    `json:"blockchainNetwork"`
	TokenAddress          string                    `json:"tokenAddress"`
	HotWalletAddress      string                    `json:"hotWalletAddress"`
	FromWalletAddress     string                    `json:"fromWalletAddress"`
	TransactionHash       string                    `json:"transactionHash"`
	PlatformFeePercentage float64                   `json:"platformFeePercentage"`
	ConversionRate        float64                   `json:"conversionRate"`
	UsdAmount             float64                   `json:"usdAmount"` // Total USD amount including transaction fees.
}

func (b BeneficiaryPaymentRequestInput) Validate() (BeneficiaryPaymentRequestInput, error) {
	if b.BeneficiaryId == "" {
		return BeneficiaryPaymentRequestInput{}, fmt.Errorf("BeneficiaryId is required")
	}
	if b.LocalAmount <= 0 {
		return BeneficiaryPaymentRequestInput{}, fmt.Errorf("LocalAmount must be greater than zero")
	}
	if b.OwnerId == "" {
		return BeneficiaryPaymentRequestInput{}, fmt.Errorf("OwnerId is required")
	}

	validSender, err := b.Sender.Validate()
	if err != nil {
		return BeneficiaryPaymentRequestInput{}, err
	}
	b.Sender = validSender

	if b.TokenSymbol == "" {
		return BeneficiaryPaymentRequestInput{}, fmt.Errorf("TokenSymbol is required")
	}
	if b.BlockchainNetwork == "" {
		return BeneficiaryPaymentRequestInput{}, fmt.Errorf("BlockchainNetwork is required")
	}
	if b.TokenAddress == "" {
		return BeneficiaryPaymentRequestInput{}, fmt.Errorf("TokenAddress is required")
	}
	if b.HotWalletAddress == "" {
		return BeneficiaryPaymentRequestInput{}, fmt.Errorf("HotWalletAddress is required")
	}
	if b.FromWalletAddress == "" {
		return BeneficiaryPaymentRequestInput{}, fmt.Errorf("FromWalletAddress is required")
	}
	if b.TransactionHash == "" {
		return BeneficiaryPaymentRequestInput{}, fmt.Errorf("TransactionHash is required")
	}
	if b.PlatformFeePercentage < 0 {
		return BeneficiaryPaymentRequestInput{}, fmt.Errorf("PlatformFeePercentage must be greater than zero")
	}
	if b.ConversionRate <= 0 {
		return BeneficiaryPaymentRequestInput{}, fmt.Errorf("ConversionRate must be greater than zero")
	}
	if b.UsdAmount <= 0 {
		return BeneficiaryPaymentRequestInput{}, fmt.Errorf("UsdAmount must be greater than zero")
	}

	localAmount := decimal.NewFromFloat(b.LocalAmount)
	conversionRate := decimal.NewFromFloat(b.ConversionRate)
	convertedUsdAmount := localAmount.Div(conversionRate)
	platformFeePercentage := decimal.NewFromFloat(b.PlatformFeePercentage).Div(yellowcard.Percent)

	expectedUsdAmount := convertedUsdAmount.Add(convertedUsdAmount.Mul(platformFeePercentage)).Round(2)
	inputUsdAmount := decimal.NewFromFloat(b.UsdAmount).Round(2)

	if !expectedUsdAmount.Equal(inputUsdAmount) {
		return BeneficiaryPaymentRequestInput{}, fmt.Errorf("UsdAmount %s should be equal to expectedUsdAmount %s", inputUsdAmount.String(), expectedUsdAmount.String())
	}

	return b, nil
}

const sequenceIdPrefix = "azza-p"
const defaultPaymentReason = "other"
const retailCustomerType = "retail"

func CreateBeneficiaryPaymentRequest(input BeneficiaryPaymentRequestInput) (CreatePaymentRequestResponse, error) {
	// Validate the input
	validInput, err := input.Validate()
	if err != nil {
		return CreatePaymentRequestResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	beneficiary, err := beneficiaries.GetBeneficiary(validInput.BeneficiaryId)

	if err != nil {
		return CreatePaymentRequestResponse{}, fmt.Errorf("failed to get beneficiary: %w", err)
	}

	sequenceId, err := yellowcard.GeneratePaymentSequenceId(yellowcard.PaymentSequenceIdInput{
		TransactionHash: validInput.TransactionHash,
		Prefix:          sequenceIdPrefix,
	})

	if err != nil {
		return CreatePaymentRequestResponse{}, fmt.Errorf("failed to generate sequence id: %w", err)
	}

	var normalizedAccountType = beneficiary.AccountType

	if beneficiary.AccountType == "phone" {
		normalizedAccountType = "momo"
	}

	var beneficiaryAccountNameWithoutDigits string

	for _, char := range beneficiary.AccountName {
		if unicode.IsDigit(char) {
			continue
		}
		beneficiaryAccountNameWithoutDigits += string(char)
	}

	paymentRequest, err := CreatePaymentRequest(PaymentRequestInput{
		Sender:      validInput.Sender,
		ChannelId:   beneficiary.ChannelId,
		LocalAmount: validInput.LocalAmount,
		SequenceId:  sequenceId,
		Reason:      defaultPaymentReason,
		Destination: PaymentRequestDestination{
			AccountType:   normalizedAccountType,
			AccountNumber: beneficiary.AccountNumber,
			AccountName:   beneficiaryAccountNameWithoutDigits,
			Country:       beneficiary.Country,
			NetworkName:   beneficiary.NetworkName,
			NetworkId:     beneficiary.NetworkId,
		},
		CustomerType: retailCustomerType,
		ForceAccept:  true,
	})

	if err != nil {
		return CreatePaymentRequestResponse{}, fmt.Errorf("failed to create payment request: %w", err)
	}

	storedPaymentRequestId, err := StoreBeneficiaryPaymentRequest(PaymentRequestRecord{
		Beneficiary:           input.BeneficiaryId,
		TokenAddress:          input.TokenAddress,
		FromWalletAddress:     input.FromWalletAddress,
		HotWalletAddress:      input.HotWalletAddress,
		OwnerId:               input.OwnerId,
		LocalAmount:           input.LocalAmount,
		TokenSymbol:           input.TokenSymbol,
		BlockchainNetwork:     input.BlockchainNetwork,
		TransactionHash:       input.TransactionHash,
		PlatformFeePercentage: input.PlatformFeePercentage,
		ConversionRate:        paymentRequest.Rate,
		UsdAmount:             input.UsdAmount,
		SequenceId:            sequenceId,
		YellowCardStatus:      paymentRequest.Status,
	})

	if err != nil {
		return CreatePaymentRequestResponse{}, fmt.Errorf("failed to store payment request: %w", err)
	}

	fmt.Printf("Stored Payment Request ID: %s", storedPaymentRequestId)

	return paymentRequest, nil
}
