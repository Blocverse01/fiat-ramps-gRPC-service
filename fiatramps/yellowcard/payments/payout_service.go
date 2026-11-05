package payments

import (
	"fmt"

	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/beneficiaries"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/countries"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/quotes"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/ycstatus"
)

type BeneficiaryLookup func(beneficiaryId, ownerId string) (beneficiaries.YellowCardBeneficiary, error)
type RateLookup func(countryCode string) (float64, error)
type YellowCardPayout func(input PaymentRequestInput) (CreatePaymentRequestResponse, error)

type PaymentStore interface {
	FindByTransactionHash(txHash string) (*StoredPaymentRequestRecord, error)
	FindBySequenceId(sequenceId string) (*StoredPaymentRequestRecord, error)
	Insert(record PaymentRequestRecord) (string, error)
	UpdateStatus(sequenceId, status string) error
}

type PayoutService struct {
	GetBeneficiary BeneficiaryLookup
	Deposits       DepositVerifier
	Rates          RateLookup
	Store          PaymentStore
	Payout         YellowCardPayout
}

func NewDefaultPayoutService() *PayoutService {
	return &PayoutService{
		GetBeneficiary: beneficiaries.GetBeneficiaryForOwner,
		Deposits:       CompositeDepositVerifier{},
		Rates:          lookupServerSellRate,
		Store:          xataPaymentStore{},
		Payout:         CreatePaymentRequest,
	}
}

var defaultPayoutService = NewDefaultPayoutService()

func CreateBeneficiaryPaymentRequest(input BeneficiaryPaymentRequestInput) (CreatePaymentRequestResponse, error) {
	return defaultPayoutService.CreateBeneficiaryPaymentRequest(input)
}

func (s *PayoutService) CreateBeneficiaryPaymentRequest(input BeneficiaryPaymentRequestInput) (CreatePaymentRequestResponse, error) {
	validInput, err := input.Validate()
	if err != nil {
		return CreatePaymentRequestResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	existing, err := s.Store.FindByTransactionHash(validInput.TransactionHash)
	if err != nil {
		return CreatePaymentRequestResponse{}, fmt.Errorf("failed to check existing payment: %w", err)
	}
	if existing != nil {
		if ycstatus.IsSentOrComplete(existing.YellowCardStatus) {
			return storedPaymentToResponse(*existing), nil
		}
		return CreatePaymentRequestResponse{}, fmt.Errorf("payment request already exists for transactionHash %s", validInput.TransactionHash)
	}

	beneficiary, err := s.GetBeneficiary(validInput.BeneficiaryId, validInput.OwnerId)
	if err != nil {
		return CreatePaymentRequestResponse{}, fmt.Errorf("failed to get beneficiary: %w", err)
	}
	if beneficiary.OwnerId != validInput.OwnerId {
		return CreatePaymentRequestResponse{}, fmt.Errorf("beneficiary ownerId does not match request ownerId")
	}

	deposit, err := s.Deposits.Verify(DepositVerifyInput{
		TransactionHash:   validInput.TransactionHash,
		HotWalletAddress:  validInput.HotWalletAddress,
		FromWalletAddress: validInput.FromWalletAddress,
		TokenAddress:      validInput.TokenAddress,
		TokenSymbol:       validInput.TokenSymbol,
		BlockchainNetwork: validInput.BlockchainNetwork,
	})
	if err != nil {
		return CreatePaymentRequestResponse{}, fmt.Errorf("deposit verification failed: %w", err)
	}

	rate := validInput.ConversionRate
	if s.Rates != nil {
		serverRate, rateErr := s.Rates(beneficiary.Country)
		if rateErr != nil {
			return CreatePaymentRequestResponse{}, fmt.Errorf("failed to load server-side rate: %w", rateErr)
		}
		rate = serverRate
	}

	localAmount := RecomputeLocalAmount(deposit.TokenAmount, rate)
	if !LocalAmountMatchesVerifiedDeposit(validInput.LocalAmount, deposit.TokenAmount, rate) {
		return CreatePaymentRequestResponse{}, fmt.Errorf("LocalAmount %v does not match verified deposit %v × rate %v (expected %v)", validInput.LocalAmount, deposit.TokenAmount, rate, localAmount)
	}
	if err := validateUsdAgainstDeposit(validInput.UsdAmount, deposit.TokenAmount, validInput.PlatformFeePercentage); err != nil {
		return CreatePaymentRequestResponse{}, err
	}

	sequenceId, err := yellowcard.GeneratePaymentSequenceId(yellowcard.PaymentSequenceIdInput{
		Prefix: sequenceIdPrefix,
	})
	if err != nil {
		return CreatePaymentRequestResponse{}, fmt.Errorf("failed to generate sequence id: %w", err)
	}
	existingBySeq, err := s.Store.FindBySequenceId(sequenceId)
	if err != nil {
		return CreatePaymentRequestResponse{}, fmt.Errorf("failed to check sequenceId: %w", err)
	}
	if existingBySeq != nil {
		if ycstatus.IsSentOrComplete(existingBySeq.YellowCardStatus) {
			return storedPaymentToResponse(*existingBySeq), nil
		}
		return CreatePaymentRequestResponse{}, fmt.Errorf("payment request already exists for sequenceId %s", sequenceId)
	}

	normalizedAccountType := beneficiary.AccountType
	if beneficiary.AccountType == "phone" {
		normalizedAccountType = "momo"
	}
	normalizedBeneficiaryName := stripDigits(beneficiary.AccountName)

	_, err = s.Store.Insert(PaymentRequestRecord{
		Beneficiary:           validInput.BeneficiaryId,
		TokenAddress:          validInput.TokenAddress,
		FromWalletAddress:     validInput.FromWalletAddress,
		HotWalletAddress:      validInput.HotWalletAddress,
		OwnerId:               validInput.OwnerId,
		LocalAmount:           localAmount,
		TokenSymbol:           validInput.TokenSymbol,
		BlockchainNetwork:     validInput.BlockchainNetwork,
		TransactionHash:       validInput.TransactionHash,
		PlatformFeePercentage: validInput.PlatformFeePercentage,
		ConversionRate:        rate,
		UsdAmount:             validInput.UsdAmount,
		SequenceId:            sequenceId,
		YellowCardStatus:      defaultPaymentStatus,
		TransactionFeeUsd:     validInput.TransactionFeeUsd,
	})
	if err != nil {
		return CreatePaymentRequestResponse{}, fmt.Errorf("failed to record payment request: %w", err)
	}

	paymentRequest, err := s.Payout(PaymentRequestInput{
		Sender:      validInput.Sender,
		ChannelId:   beneficiary.ChannelId,
		LocalAmount: localAmount,
		SequenceId:  sequenceId,
		Reason:      defaultPaymentReason,
		Destination: PaymentRequestDestination{
			AccountType:   normalizedAccountType,
			AccountNumber: beneficiary.AccountNumber,
			AccountName:   normalizedBeneficiaryName,
			Country:       beneficiary.Country,
			NetworkName:   beneficiary.NetworkName,
			NetworkId:     beneficiary.NetworkId,
		},
		CustomerType: retailCustomerType,
		ForceAccept:  true,
	})
	if err != nil {
		return CreatePaymentRequestResponse{}, fmt.Errorf("failed to send payment request to YC: %w", err)
	}

	if updateErr := s.Store.UpdateStatus(sequenceId, sentPaymentRequestStatus); updateErr != nil {
		return paymentRequest, fmt.Errorf("payment sent but failed to update status: %w", updateErr)
	}

	return paymentRequest, nil
}

func lookupServerSellRate(countryCode string) (float64, error) {
	country, err := countries.GetCountry(map[string]string{"code": countryCode})
	if err != nil {
		return 0, err
	}
	rates := quotes.GetRates()
	if len(rates) == 0 {
		return 0, fmt.Errorf("no rates available")
	}
	rate, err := quotes.FilterRatesByCurrency(rates, country.CurrencySymbol)
	if err != nil {
		return 0, err
	}
	if rate.Sell <= 0 {
		return 0, fmt.Errorf("invalid server sell rate")
	}
	return float64(rate.Sell), nil
}

func storedPaymentToResponse(rec StoredPaymentRequestRecord) CreatePaymentRequestResponse {
	return CreatePaymentRequestResponse{
		SequenceId:      rec.SequenceId,
		Amount:          rec.UsdAmount,
		ConvertedAmount: rec.LocalAmount,
		Status:          rec.YellowCardStatus,
		Rate:            rec.ConversionRate,
		Country:         rec.Beneficiary.Country,
		Destination: PaymentRequestDestination{
			AccountNumber: rec.Beneficiary.AccountNumber,
			AccountType:   rec.Beneficiary.AccountType,
			NetworkId:     rec.Beneficiary.NetworkId,
			NetworkName:   rec.Beneficiary.NetworkName,
			Country:       rec.Beneficiary.Country,
			AccountName:   rec.Beneficiary.AccountName,
		},
	}
}

func stripDigits(name string) string {
	var out string
	for _, char := range name {
		if char >= '0' && char <= '9' {
			continue
		}
		out += string(char)
	}
	return out
}
