package payments

import (
	"fmt"
	"testing"

	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/beneficiaries"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/ycstatus"
)

type stubDeposit struct {
	amount float64
	err    error
}

func (s stubDeposit) Verify(input DepositVerifyInput) (VerifiedDeposit, error) {
	if s.err != nil {
		return VerifiedDeposit{}, s.err
	}
	return VerifiedDeposit{
		TransactionHash: input.TransactionHash,
		ToAddress:       input.HotWalletAddress,
		FromAddress:     input.FromWalletAddress,
		TokenAddress:    input.TokenAddress,
		TokenAmount:     s.amount,
		Confirmed:       true,
	}, nil
}

type memStore struct {
	byHash  map[string]*StoredPaymentRequestRecord
	bySeq   map[string]*StoredPaymentRequestRecord
	inserts int
}

func newMemStore() *memStore {
	return &memStore{
		byHash: map[string]*StoredPaymentRequestRecord{},
		bySeq:  map[string]*StoredPaymentRequestRecord{},
	}
}

func (m *memStore) FindByTransactionHash(txHash string) (*StoredPaymentRequestRecord, error) {
	return m.byHash[txHash], nil
}

func (m *memStore) FindBySequenceId(sequenceId string) (*StoredPaymentRequestRecord, error) {
	return m.bySeq[sequenceId], nil
}

func (m *memStore) Insert(record PaymentRequestRecord) (string, error) {
	if _, ok := m.byHash[record.TransactionHash]; ok {
		return "", fmt.Errorf("duplicate payment request for sequenceId or transactionHash")
	}
	rec := &StoredPaymentRequestRecord{
		OwnerId:          record.OwnerId,
		LocalAmount:      record.LocalAmount,
		TransactionHash:  record.TransactionHash,
		SequenceId:       record.SequenceId,
		YellowCardStatus: record.YellowCardStatus,
		ConversionRate:   record.ConversionRate,
		UsdAmount:        record.UsdAmount,
	}
	m.byHash[record.TransactionHash] = rec
	m.bySeq[record.SequenceId] = rec
	m.inserts++
	return "rec_1", nil
}

func (m *memStore) UpdateStatus(sequenceId, status string) error {
	if rec, ok := m.bySeq[sequenceId]; ok {
		rec.YellowCardStatus = status
	}
	return nil
}

func validPayoutInput(ownerId string, localAmount, rate, usdAmount float64) BeneficiaryPaymentRequestInput {
	return BeneficiaryPaymentRequestInput{
		BeneficiaryId: "ben-1",
		LocalAmount:   localAmount,
		OwnerId:       ownerId,
		Sender: yellowcard.RetailCustomer{
			Name:     "Ada Lovelace",
			Phone:    "+2348143100808",
			Country:  "NG",
			Address:  "1 Test Street",
			Dob:      "1990-01-02",
			IdType:   "NIN",
			IdNumber: "12345678901",
		},
		TokenSymbol:           "USDC",
		BlockchainNetwork:     "Base",
		TokenAddress:          "0xToken",
		HotWalletAddress:      "0xHot",
		FromWalletAddress:     "0xFrom",
		TransactionHash:       "0xabc123",
		PlatformFeePercentage: 0,
		ConversionRate:        rate,
		UsdAmount:             usdAmount,
	}
}

func ownerBeneficiary(ownerId string) beneficiaries.YellowCardBeneficiary {
	return beneficiaries.YellowCardBeneficiary{
		Id:            "ben-1",
		OwnerId:       ownerId,
		AccountType:   "bank",
		AccountName:   "Ada",
		AccountNumber: "1234567890",
		Country:       "NG",
		ChannelId:     "channel",
		NetworkId:     "net",
		NetworkName:   "bank",
	}
}

func TestCreateBeneficiaryPaymentRequestRejectsMismatchedOwnerAndAmount(t *testing.T) {
	store := newMemStore()
	ycCalls := 0
	svc := &PayoutService{
		GetBeneficiary: func(beneficiaryId, ownerId string) (beneficiaries.YellowCardBeneficiary, error) {
			if ownerId != "owner-1" {
				return beneficiaries.YellowCardBeneficiary{}, fmt.Errorf("no beneficiary found")
			}
			return ownerBeneficiary("owner-1"), nil
		},
		Deposits: stubDeposit{amount: 10},
		Rates:    func(countryCode string) (float64, error) { return 1500, nil },
		Store:    store,
		Payout: func(input PaymentRequestInput) (CreatePaymentRequestResponse, error) {
			ycCalls++
			return CreatePaymentRequestResponse{SequenceId: input.SequenceId, Status: "sent"}, nil
		},
	}

	_, err := svc.CreateBeneficiaryPaymentRequest(validPayoutInput("owner-2", 15000, 1500, 10))
	if err == nil {
		t.Fatal("expected mismatched ownerId to be rejected")
	}

	_, err = svc.CreateBeneficiaryPaymentRequest(validPayoutInput("owner-1", 1, 1500, 10))
	if err == nil {
		t.Fatal("expected amount that does not match verified deposit × rate to be rejected")
	}
	if ycCalls != 0 {
		t.Fatalf("YC should not be called on rejected requests, got %d", ycCalls)
	}

	got, err := svc.CreateBeneficiaryPaymentRequest(validPayoutInput("owner-1", 15000, 1500, 10))
	if err != nil {
		t.Fatalf("valid payout failed: %v", err)
	}
	if got.SequenceId == "" || ycCalls != 1 {
		t.Fatalf("expected one successful YC call, got %+v calls=%d", got, ycCalls)
	}
}

func TestDuplicateTxHashDoesNotCallYellowCardAgain(t *testing.T) {
	store := newMemStore()
	store.byHash["0xabc123"] = &StoredPaymentRequestRecord{
		TransactionHash:  "0xabc123",
		SequenceId:       "azza-p-existing",
		YellowCardStatus: ycstatus.Sent,
		OwnerId:          "owner-1",
	}
	ycCalls := 0
	svc := &PayoutService{
		GetBeneficiary: func(beneficiaryId, ownerId string) (beneficiaries.YellowCardBeneficiary, error) {
			return ownerBeneficiary(ownerId), nil
		},
		Deposits: stubDeposit{amount: 10},
		Rates:    func(countryCode string) (float64, error) { return 1500, nil },
		Store:    store,
		Payout: func(input PaymentRequestInput) (CreatePaymentRequestResponse, error) {
			ycCalls++
			return CreatePaymentRequestResponse{SequenceId: input.SequenceId}, nil
		},
	}

	got, err := svc.CreateBeneficiaryPaymentRequest(validPayoutInput("owner-1", 15000, 1500, 10))
	if err != nil {
		t.Fatalf("duplicate replay should return existing row: %v", err)
	}
	if got.SequenceId != "azza-p-existing" {
		t.Fatalf("got sequence %s", got.SequenceId)
	}
	if ycCalls != 0 {
		t.Fatalf("duplicate txHash produced %d YC calls", ycCalls)
	}
	if store.inserts != 0 {
		t.Fatalf("duplicate insert occurred")
	}
}

func TestDuplicateSequenceIdDoesNotCallYellowCardAgain(t *testing.T) {
	store := newMemStore()
	ycCalls := 0
	svc := &PayoutService{
		GetBeneficiary: func(beneficiaryId, ownerId string) (beneficiaries.YellowCardBeneficiary, error) {
			return ownerBeneficiary("owner-1"), nil
		},
		Deposits: stubDeposit{amount: 10},
		Rates:    func(countryCode string) (float64, error) { return 1500, nil },
		Store:    store,
		Payout: func(input PaymentRequestInput) (CreatePaymentRequestResponse, error) {
			ycCalls++
			return CreatePaymentRequestResponse{SequenceId: input.SequenceId}, nil
		},
	}

	first, err := svc.CreateBeneficiaryPaymentRequest(validPayoutInput("owner-1", 15000, 1500, 10))
	if err != nil {
		t.Fatalf("first payout: %v", err)
	}
	if ycCalls != 1 {
		t.Fatalf("first payout YC calls=%d", ycCalls)
	}

	store.byHash["0xabc123"] = &StoredPaymentRequestRecord{
		TransactionHash:  "0xabc123",
		SequenceId:       first.SequenceId,
		YellowCardStatus: ycstatus.Complete,
	}
	_, err = svc.CreateBeneficiaryPaymentRequest(validPayoutInput("owner-1", 15000, 1500, 10))
	if err != nil {
		t.Fatalf("duplicate sequence replay: %v", err)
	}
	if ycCalls != 1 {
		t.Fatalf("duplicate sequenceId produced a second YC call (%d)", ycCalls)
	}
}
