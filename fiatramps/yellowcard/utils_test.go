package yellowcard

import "testing"

func TestUtils(t *testing.T) {
	t.Run("Test GeneratePaymentSequenceId", func(t *testing.T) {
		prefix := "azza-p"

		first, err := GeneratePaymentSequenceId(PaymentSequenceIdInput{
			TransactionHash: "0xpublic",
			Prefix:          prefix,
		})
		if err != nil {
			t.Fatal(err)
		}
		second, err := GeneratePaymentSequenceId(PaymentSequenceIdInput{
			TransactionHash: "0xpublic",
			Prefix:          prefix,
		})
		if err != nil {
			t.Fatal(err)
		}
		if first == second {
			t.Fatal("offramp sequence ids must be random, not derived from txHash")
		}
		if len(first) < len(prefix)+5 {
			t.Fatalf("unexpected sequence id %s", first)
		}
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
