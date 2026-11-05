package payments

import (
	"fmt"

	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
	"github.com/shopspring/decimal"
)

func RecomputeLocalAmount(verifiedTokenAmount, conversionRate float64) float64 {
	return decimal.NewFromFloat(verifiedTokenAmount).
		Mul(decimal.NewFromFloat(conversionRate)).
		Round(2).
		InexactFloat64()
}

func LocalAmountMatchesVerifiedDeposit(localAmount, verifiedTokenAmount, conversionRate float64) bool {
	expected := decimal.NewFromFloat(RecomputeLocalAmount(verifiedTokenAmount, conversionRate)).Round(2)
	got := decimal.NewFromFloat(localAmount).Round(2)
	return got.Equal(expected)
}

func expectedUsdAmount(verifiedTokenAmount, platformFeePercentage float64) decimal.Decimal {
	token := decimal.NewFromFloat(verifiedTokenAmount)
	feePct := decimal.NewFromFloat(platformFeePercentage).Div(yellowcard.Percent)
	return token.Add(token.Mul(feePct)).Round(2)
}

func validateUsdAgainstDeposit(usdAmount, verifiedTokenAmount, platformFeePercentage float64) error {
	expected := expectedUsdAmount(verifiedTokenAmount, platformFeePercentage)
	got := decimal.NewFromFloat(usdAmount).Round(2)
	if !got.Equal(expected) {
		return fmt.Errorf("UsdAmount %s should be equal to expectedUsdAmount %s", got.String(), expected.String())
	}
	return nil
}
