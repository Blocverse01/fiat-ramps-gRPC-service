package payments

import (
	"fmt"

	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/ycstatus"
)

type xataPaymentStore struct{}

func (xataPaymentStore) FindByTransactionHash(txHash string) (*StoredPaymentRequestRecord, error) {
	return FindPaymentRequestByFilter(map[string]string{"transactionHash": txHash})
}

func (xataPaymentStore) FindBySequenceId(sequenceId string) (*StoredPaymentRequestRecord, error) {
	return FindPaymentRequestByFilter(map[string]string{"sequenceId": sequenceId})
}

func (xataPaymentStore) Insert(record PaymentRequestRecord) (string, error) {
	return StoreBeneficiaryPaymentRequest(record)
}

func (xataPaymentStore) UpdateStatus(sequenceId, status string) error {
	return patchPaymentRequestStatus(sequenceId, status)
}

type ProviderStatusFetcher func(sequenceId string) (string, error)

var FetchPaymentProviderStatus ProviderStatusFetcher = FetchYellowCardPaymentStatus

func UpdatePaymentRequestStatus(sequenceId string, status string) error {
	return applyPaymentStatusUpdate(sequenceId, status, FetchPaymentProviderStatus, GetPaymentRequestTransaction, patchPaymentRequestStatus)
}

func applyPaymentStatusUpdate(
	sequenceId string,
	status string,
	fetch ProviderStatusFetcher,
	load func(string) (StoredPaymentRequestRecord, error),
	patch func(string, string) error,
) error {
	current, err := load(sequenceId)
	if err != nil {
		return err
	}
	if fetch == nil {
		return fmt.Errorf("provider status fetch is not configured")
	}
	providerStatus, err := fetch(sequenceId)
	if err != nil {
		return fmt.Errorf("failed to re-fetch provider status: %w", err)
	}
	if err := ycstatus.ValidateStatusWrite(current.YellowCardStatus, status, providerStatus); err != nil {
		return err
	}
	if ycstatus.Normalize(current.YellowCardStatus) == ycstatus.Normalize(status) {
		return nil
	}
	return patch(sequenceId, ycstatus.Normalize(status))
}
