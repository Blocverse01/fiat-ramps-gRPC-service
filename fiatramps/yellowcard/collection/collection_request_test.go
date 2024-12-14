package collection

import (
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/payment_methods"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/quotes"
	"github.com/shopspring/decimal"
	"testing"
)

func TestCreateCollectionRequest(t *testing.T) {
	fiatramps.LoadEnv("../../../.env")

	const phoneNumberForTesting = "+2348143100808"
	const sequenceIdPrefix = "azza-c"
	t.Run("Test Create Collection Request", func(t *testing.T) {
		amountToBuy := decimal.NewFromFloat(15)

		const currency = "NGN"
		const countryCode = "NG"

		quote := quotes.GetQuotes(quotes.FiatConfig{
			CountryCode:    countryCode,
			CurrencySymbol: currency,
		})

		paymentChannels, err := payment_methods.GetPaymentMethods(countryCode, "deposit")

		if err != nil {
			t.Fatalf("error getting payment channels: %v", err)
		}

		var paymentMethod payment_methods.PaymentMethod
		for _, channel := range paymentChannels.PaymentMethods {
			if channel.ChannelName == "bank" {
				paymentMethod = channel
				break
			}
		}

		localAmount := amountToBuy.Mul(decimal.NewFromFloat(float64(quote.Rate.Buy)))
		sequenceId, err := yellowcard.GenerateCollectionSequenceId(sequenceIdPrefix)

		if err != nil {
			t.Fatalf("error generating sequence id: %v", err)
		}

		collectionRequest, err := CreateCollectionRequest(RequestPayload{
			LocalAmount:  localAmount.InexactFloat64(),
			SequenceId:   sequenceId,
			ChannelId:    paymentMethod.ChannelId,
			ForceAccept:  true,
			CustomerType: "retail",
			Source: Source{
				AccountType: "bank",
			},
			Recipient: yellowcard.RetailCustomer{
				Name:     "Joshua Avoaja",
				Phone:    phoneNumberForTesting,
				Country:  "NG",
				Dob:      "1999/12/01",
				IdNumber: "80812345678",
				IdType:   "NIN",
			},
		})

		if err != nil {
			t.Fatalf("error creating collection request: %v", err)
		}

		t.Logf("Collection Request: %v", collectionRequest)
	})
}
