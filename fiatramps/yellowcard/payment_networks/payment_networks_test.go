package payment_networks

import (
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"testing"
)

func TestGetPaymentNetworks(t *testing.T) {
	fiatramps.SkipWithoutLiveEnv(t, "../../../.env")
	fiatramps.LoadEnv("../../../.env")

	t.Run("Test Get Payment Networks", func(t *testing.T) {
		const channelId = NgnBankChannelId
		paymentNetworks, err := GetPaymentNetworks("NG", channelId)

		if err != nil {
			t.Errorf("Error occurred while getting payment networks: %v", err)
		}

		t.Logf("Payment networks: %v", paymentNetworks)
	})
}
