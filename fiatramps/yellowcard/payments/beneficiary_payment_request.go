package payments

import (
	"fmt"

	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
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
	TransactionFeeUsd     float64                   `json:"transactionFeeUsd"`
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

	return b, nil
}

const sequenceIdPrefix = "azza-p"
const defaultPaymentReason = "other"
const retailCustomerType = "retail"
const defaultPaymentStatus = "recorded"
const sentPaymentRequestStatus = "sent"
