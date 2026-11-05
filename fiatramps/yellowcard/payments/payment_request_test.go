package payments

import (
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/quotes"
	"github.com/shopspring/decimal"
	"testing"
)

func TestCreateBeneficiaryPaymentRequest(t *testing.T) {
	fiatramps.SkipWithoutLiveEnv(t, "../../../.env")
	fiatramps.LoadEnv("../../../.env")

	const phoneNumberForTesting = "+2348143100808"
	//"2348123456789"
	const sampleAddress = "202 Gingseng street, LLama city, Lagos state, Nigeria."

	t.Run("should return error if payment request creation fails", func(t *testing.T) {
		amountToSell := decimal.NewFromFloat(2)
		const beneficiaryId = "rec_ctcbgla6ntp0lapg2d60"
		const ownerId, senderPhone = phoneNumberForTesting, phoneNumberForTesting

		quote := quotes.GetQuotes(quotes.FiatConfig{
			CountryCode:    "NG",
			CurrencySymbol: "NGN",
		})

		platformFeePercentage := decimal.NewFromFloat(float64(quote.Fees.OfframpFeePercentage))
		localAmount := amountToSell.Mul(decimal.NewFromFloat(float64(quote.Rate.Sell)))
		transactionFee := amountToSell.Mul(platformFeePercentage.Div(yellowcard.Percent))
		usdAmount := amountToSell.Add(transactionFee)

		paymentRequest, err := CreateBeneficiaryPaymentRequest(BeneficiaryPaymentRequestInput{
			BeneficiaryId: beneficiaryId,
			LocalAmount:   localAmount.InexactFloat64(),
			OwnerId:       ownerId,
			Sender: yellowcard.RetailCustomer{
				Name:     "Joshua Avoaja",
				Phone:    senderPhone,
				Country:  "NG",
				Address:  sampleAddress,
				Dob:      "1999/12/01",
				IdNumber: "80812345678",
				IdType:   "NIN",
			},
			TokenSymbol:           "USDC",
			TokenAddress:          "0x00000000000000000000000000000",
			TransactionHash:       "0x1516f1b307f63567703daa5434354d909c5055e016db66bdbac958cb0a9a8942",
			BlockchainNetwork:     "Base",
			HotWalletAddress:      "0x62c021e584702C40D0c14923ba6934791a8DaaD3",
			FromWalletAddress:     "0x4200000000000000000000000000000000000042",
			PlatformFeePercentage: platformFeePercentage.InexactFloat64(),
			ConversionRate:        float64(quote.Rate.Sell),
			UsdAmount:             usdAmount.InexactFloat64(),
		})

		if err != nil {
			t.Fatalf("error creating payment request: %v", err)
		}

		t.Log("payment request created: ", paymentRequest)
	})
}
