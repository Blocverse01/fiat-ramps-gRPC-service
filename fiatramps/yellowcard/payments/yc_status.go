package payments

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
)

type yellowCardPaymentStatusResponse struct {
	Status     string `json:"status"`
	SequenceId string `json:"sequenceId"`
	Payments   []struct {
		Status     string `json:"status"`
		SequenceId string `json:"sequenceId"`
	} `json:"payments"`
}

func FetchYellowCardPaymentStatus(sequenceId string) (string, error) {
	resp, err := yellowcard.MakeYellowCardRequest("GET", paymentRequestEndpoint, nil, map[string]string{"sequenceId": sequenceId})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fiatramps.HandleFailedResponse(resp)
		return "", fmt.Errorf("failed to fetch payment status: %s", resp.Status)
	}
	var body yellowCardPaymentStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	if body.Status != "" {
		return body.Status, nil
	}
	for _, payment := range body.Payments {
		if payment.SequenceId == sequenceId || payment.SequenceId == "" {
			return payment.Status, nil
		}
	}
	return "", fmt.Errorf("provider status not found for sequenceId %s", sequenceId)
}
