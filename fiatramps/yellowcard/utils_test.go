package yellowcard

import "testing"

func TestUtils(t *testing.T) {
	t.Run("Test GeneratePaymentSequenceId", func(t *testing.T) {
		transactionHash := "0x1516f1b307f63567703daa5434354d909c5055e016db66bcbac958ab0a7a66c8"
		prefix := "azza-p"

		sequenceId, err := GeneratePaymentSequenceId(PaymentSequenceIdInput{
			TransactionHash: transactionHash,
			Prefix:          prefix,
		})

		if err != nil {
			t.Error(err)
		}

		t.Log("sequenceId: ", sequenceId)
	})

	t.Run("Test date formatter", func(t *testing.T) {
		var sampleDates = []string{
			"2000-02-07",
			"2000/02/07",
			"07-02-2000",
			"07/02/2000",
			"02-07-2000",
			"02/07/2000",
		}

		for _, date := range sampleDates {
			t.Log("date: ", date)
			transformedDate, err := TransformDateString(date)
			if err != nil {
				t.Error(err)
			}
			t.Log("transformedDate: ", transformedDate)
		}
	})
}
