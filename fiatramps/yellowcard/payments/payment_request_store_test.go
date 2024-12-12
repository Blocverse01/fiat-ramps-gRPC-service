package payments

import (
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"testing"
)

func TestGetPaymentRequestTransaction(t *testing.T) {
	fiatramps.LoadEnv("../../../.env")

	t.Run("Test Get Payment Request Transaction", func(t *testing.T) {
		paymentRequestTransaction, err := GetPaymentRequestTransaction("azza-p-793e5b225f64c25bae44ad48")

		if err != nil {
			t.Errorf("Error occurred while getting payment request transaction: %v", err)
		}

		t.Logf("Payment Request Transaction: %v", paymentRequestTransaction.Beneficiary)
	})
}
