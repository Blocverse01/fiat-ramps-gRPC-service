package payments

import (
	"fmt"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/beneficiaries"
	"github.com/shopspring/decimal"
	"time"
	"unicode"
)

type BeneficiaryPaymentRequestInput struct {
	BeneficiaryId         string               `json:"beneficiaryId"`
	LocalAmount           float64              `json:"localAmount"`
	OwnerId               string               `json:"ownerId"`
	Sender                PaymentRequestSender `json:"sender"`
	TokenSymbol           string               `json:"tokenSymbol"`
	BlockchainNetwork     string               `json:"blockchainNetwork"`
	TokenAddress          string               `json:"tokenAddress"`
	HotWalletAddress      string               `json:"hotWalletAddress"`
	FromWalletAddress     string               `json:"fromWalletAddress"`
	TransactionHash       string               `json:"transactionHash"`
	PlatformFeePercentage float64              `json:"platformFeePercentage"`
	ConversionRate        float64              `json:"conversionRate"`
	UsdAmount             float64              `json:"usdAmount"` // Total USD amount including transaction fees.
}

func (b BeneficiaryPaymentRequestInput) Validate() error {
	if b.BeneficiaryId == "" {
		return fmt.Errorf("BeneficiaryId is required")
	}
	if b.LocalAmount <= 0 {
		return fmt.Errorf("LocalAmount must be greater than zero")
	}
	if b.OwnerId == "" {
		return fmt.Errorf("OwnerId is required")
	}
	if b.Sender.Name == "" {
		return fmt.Errorf("Sender.Name is required")
	}
	if b.Sender.Country == "" {
		return fmt.Errorf("Sender.Country is required")
	}
	if _, err := time.Parse("2006/01/02", b.Sender.Dob); err != nil {
		return fmt.Errorf("Sender.Dob should be a valid date: %v", err)
	}
	if b.Sender.Phone == "" || len(b.Sender.Phone) < 8 || len(b.Sender.Phone) > 15 {
		return fmt.Errorf("Sender.Phone should be a valid phone number with 8-15 digits")
	}
	for _, r := range b.Sender.Phone {
		if !unicode.IsDigit(r) {
			return fmt.Errorf("Sender.Phone should contain only digits")
		}
	}
	// Todo: Validate acceptable ID Types and their corresponding lengths
	if b.Sender.IdType == "" {
		return fmt.Errorf("Sender.IdType is required")
	}
	if b.Sender.IdNumber == "" {
		return fmt.Errorf("Sender.IdNumber is required")
	}
	if b.TokenSymbol == "" {
		return fmt.Errorf("TokenSymbol is required")
	}
	if b.BlockchainNetwork == "" {
		return fmt.Errorf("BlockchainNetwork is required")
	}
	if b.TokenAddress == "" {
		return fmt.Errorf("TokenAddress is required")
	}
	if b.HotWalletAddress == "" {
		return fmt.Errorf("HotWalletAddress is required")
	}
	if b.FromWalletAddress == "" {
		return fmt.Errorf("FromWalletAddress is required")
	}
	if b.TransactionHash == "" {
		return fmt.Errorf("TransactionHash is required")
	}
	if b.PlatformFeePercentage < 0 {
		return fmt.Errorf("PlatformFeePercentage must be greater than zero")
	}
	if b.ConversionRate <= 0 {
		return fmt.Errorf("ConversionRate must be greater than zero")
	}
	if b.UsdAmount <= 0 {
		return fmt.Errorf("UsdAmount must be greater than zero")
	}

	localAmount := decimal.NewFromFloat(b.LocalAmount)
	conversionRate := decimal.NewFromFloat(b.ConversionRate)
	convertedUsdAmount := localAmount.Div(conversionRate)
	platformFeePercentage := decimal.NewFromFloat(b.PlatformFeePercentage).Div(Percent)

	expectedUsdAmount := convertedUsdAmount.Add(convertedUsdAmount.Mul(platformFeePercentage)).Round(2)
	inputUsdAmount := decimal.NewFromFloat(b.UsdAmount).Round(2)

	if !expectedUsdAmount.Equal(inputUsdAmount) {
		return fmt.Errorf("UsdAmount %s should be equal to expectedUsdAmount %s", inputUsdAmount.String(), expectedUsdAmount.String())
	}

	return nil
}

const sequenceIdPrefix = "azza-p"
const defaultPaymentReason = "other"
const retailCustomerType = "retail"

func CreateBeneficiaryPaymentRequest(input BeneficiaryPaymentRequestInput) (CreatePaymentRequestResponse, error) {
	// Validate the input
	if err := input.Validate(); err != nil {
		return CreatePaymentRequestResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	beneficiary, err := beneficiaries.GetBeneficiary(input.BeneficiaryId)

	if err != nil {
		return CreatePaymentRequestResponse{}, fmt.Errorf("failed to get beneficiary: %w", err)
	}

	sequenceId, err := GeneratePaymentSequenceId(PaymentSequenceIdInput{
		TransactionHash: input.TransactionHash,
		Prefix:          sequenceIdPrefix,
	})

	if err != nil {
		return CreatePaymentRequestResponse{}, fmt.Errorf("failed to generate sequence id: %w", err)
	}

	var normalizedAccountType = beneficiary.AccountType

	if beneficiary.AccountType == "phone" {
		normalizedAccountType = "momo"
	}

	paymentRequest, err := CreatePaymentRequest(PaymentRequestInput{
		Sender:      input.Sender,
		ChannelId:   beneficiary.ChannelId,
		LocalAmount: input.LocalAmount,
		SequenceId:  sequenceId,
		Reason:      defaultPaymentReason,
		Destination: PaymentRequestDestination{
			AccountType:   normalizedAccountType,
			AccountNumber: beneficiary.AccountNumber,
			AccountName:   beneficiary.AccountName,
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
		ConversionRate:        input.ConversionRate,
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
