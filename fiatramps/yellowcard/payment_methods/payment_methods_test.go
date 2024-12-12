package payment_methods

import (
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"testing"
)

func TestGetPaymentMethods(t *testing.T) {
	fiatramps.LoadEnv("../../../.env")

	t.Run("Test Get Payment Methods", func(t *testing.T) {
		const countryCode = "NG"
		const route = "withdraw"
		paymentMethods, err := GetPaymentMethods(countryCode, route)

		if err != nil {
			t.Errorf("Error occurred while getting payment methods: %v", err)
		}

		t.Logf("Payment methods: %v", paymentMethods)
	})
}
