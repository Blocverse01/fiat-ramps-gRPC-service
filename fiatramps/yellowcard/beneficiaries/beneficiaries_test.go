package beneficiaries

import (
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/payment_networks"
	"testing"
)

func TestBeneficiaries(t *testing.T) {
	fiatramps.LoadEnv("../../../.env")

	t.Run("Test Get Beneficiaries", func(t *testing.T) {
		ownerId := "2348143100808"
		countryCode := "NG"
		beneficiaries, err := GetBeneficiaries(ownerId, countryCode, "bank")

		if err != nil {
			t.Errorf("Error occurred while getting beneficiaries: %v", err)
		}

		t.Logf("Beneficiaries: %v", beneficiaries)
	})

	t.Run("Test Get Beneficiary", func(t *testing.T) {
		beneficiaryId := "rec_ctcbgla6ntp0lapg2d60"
		beneficiary, err := GetBeneficiary(beneficiaryId)

		if err != nil {
			t.Errorf("Error occurred while getting beneficiary: %v", err)
		}

		t.Logf("Beneficiary: %v", beneficiary)
	})

	t.Run("Test add beneficiary", func(t *testing.T) {
		ownerId := "2348143100808"

		var countryCode = "NG"

		networks, err := payment_networks.GetPaymentNetworks(countryCode, payment_networks.NgnBankChannelId)

		if err != nil {
			t.Errorf("Error occurred while getting payment networks: %v", err)
		}

		targetNetwork := networks[0]

		newBeneficiaryId, err := AddBeneficiary(AddBeneficiaryRequestBody{
			AccountName:   "Test Account",
			AccountNumber: "1111111111",
			NetworkName:   targetNetwork.Name,
			NetworkId:     targetNetwork.NetworkId,
			ChannelId:     targetNetwork.ChannelId,
			OwnerId:       ownerId,
			Country:       countryCode,
			AccountType:   "bank",
		})

		t.Logf("New beneficiary ID: %v", newBeneficiaryId)
	})
}
