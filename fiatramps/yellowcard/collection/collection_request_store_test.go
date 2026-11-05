package collection

import (
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"testing"
)

func TestUpdateCollectionRequestStatus(t *testing.T) {
	fiatramps.SkipWithoutLiveEnv(t, "../../../.env")
	fiatramps.LoadEnv("../../../.env")

	t.Run("Test Update Collection Request Status", func(t *testing.T) {
		sequenceId := "azza-c-1vzBmu8KWRI8N0aUJuCDPKuR"
		status := "processing"

		err := UpdateCollectionRequestStatus(sequenceId, status)

		if err != nil {
			t.Fatalf("error updating collection request status: %v", err)
		}
	})
}
