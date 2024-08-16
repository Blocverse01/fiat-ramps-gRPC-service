package beneficiaries

import (
	"github.com/blocverse01/globachain-ramps-grpc-service/fiatramps"
	"testing"
)

func TestGetBeneficiaries(t *testing.T) {
	fiatramps.LoadEnv("../../../.env")

	t.Run("Test Get Beneficiaries", func(t *testing.T) {
		ownerId := "2348143100808"
		countryCode := "NG"
		beneficiaries := GetBeneficiaries(ownerId, countryCode, "")

		t.Logf("Beneficiaries: %v", beneficiaries)
	})
}
