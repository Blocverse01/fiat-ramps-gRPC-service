package quotes

import (
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"testing"
)

func TestGetQuotes(t *testing.T) {
	fiatramps.SkipWithoutLiveEnv(t, "../../../.env")
	fiatramps.LoadEnv("../../../.env")
	t.Run("Test GetTransactionFee", func(t *testing.T) {
		// Test the GetTransactionFee function
		expectedOnrampFeePercentage := float32(1)
		expectedOfframpFeePercentage := float32(1)
		transactionFee, err := GetTransactionFee("NG")

		if err != nil {
			t.Errorf("Error occurred while getting transaction fee: %v", err)
		}

		if transactionFee.OnrampFeePercentage != expectedOnrampFeePercentage {
			t.Errorf("Expected onramp fee percentage to be %v, got %v", expectedOnrampFeePercentage, transactionFee.OnrampFeePercentage)
		}

		if transactionFee.OfframpFeePercentage != expectedOfframpFeePercentage {
			t.Errorf("Expected offramp fee percentage to be %v, got %v", expectedOfframpFeePercentage, transactionFee.OfframpFeePercentage)
		}
	})
}
