package collection

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/ycstatus"
)

type ProviderStatusFetcher func(sequenceId string) (string, error)

var FetchCollectionProviderStatus ProviderStatusFetcher = FetchYellowCardCollectionStatus

type yellowCardCollectionStatusResponse struct {
	Status      string `json:"status"`
	SequenceId  string `json:"sequenceId"`
	Collections []struct {
		Status     string `json:"status"`
		SequenceId string `json:"sequenceId"`
	} `json:"collections"`
}

func FetchYellowCardCollectionStatus(sequenceId string) (string, error) {
	resp, err := yellowcard.MakeYellowCardRequest("GET", collectionsEndpoint, nil, map[string]string{"sequenceId": sequenceId})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fiatramps.HandleFailedResponse(resp)
		return "", fmt.Errorf("failed to fetch collection status: %s", resp.Status)
	}
	var body yellowCardCollectionStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	if body.Status != "" {
		return body.Status, nil
	}
	for _, item := range body.Collections {
		if item.SequenceId == sequenceId || item.SequenceId == "" {
			return item.Status, nil
		}
	}
	return "", fmt.Errorf("provider status not found for sequenceId %s", sequenceId)
}

func UpdateCollectionRequestStatus(sequenceId string, status string) error {
	return applyCollectionStatusUpdate(sequenceId, status, FetchCollectionProviderStatus, GetCollectionRequestTransaction, patchCollectionRequestStatus)
}

func applyCollectionStatusUpdate(
	sequenceId string,
	status string,
	fetch ProviderStatusFetcher,
	load func(string) (RequestRecord, error),
	patch func(string, string) error,
) error {
	current, err := load(sequenceId)
	if err != nil {
		return err
	}
	if fetch == nil {
		return fmt.Errorf("provider status fetch is not configured")
	}
	providerStatus, err := fetch(sequenceId)
	if err != nil {
		return fmt.Errorf("failed to re-fetch provider status: %w", err)
	}
	if err := ycstatus.ValidateStatusWrite(current.YellowCardStatus, status, providerStatus); err != nil {
		return err
	}
	if ycstatus.Normalize(current.YellowCardStatus) == ycstatus.Normalize(status) {
		return nil
	}
	return patch(sequenceId, ycstatus.Normalize(status))
}
