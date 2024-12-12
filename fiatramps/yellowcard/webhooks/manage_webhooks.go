package webhooks

import (
	"encoding/json"
	"fmt"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
	"log"
	"net/http"
)

type CreateWebhookRequest struct {
	Url    string `json:"url"`
	State  string `json:"state"`
	Active bool   `json:"active"`
}

type CreateWebhookResponse struct {
	PartnerId string `json:"partnerId"`
	Url       string `json:"url"`
	State     string `json:"state"`
	Active    bool   `json:"active"`
	Id        string `json:"id"`
}

const webhooksEndpoint = "/business/webhooks"

func CreateWebhook(input CreateWebhookRequest) (CreateWebhookResponse, error) {
	bodyData, err := fiatramps.BufferBodyData(input)

	if err != nil {
		return CreateWebhookResponse{}, err
	}

	resp, err := yellowcard.MakeYellowCardRequest("POST", webhooksEndpoint, bodyData, map[string]string{})
	if err != nil {
		log.Printf("Failed to send request: %v", err)
		return CreateWebhookResponse{}, err
	}

	log.Printf("Response status for request [%s]: %s", webhooksEndpoint, resp.Status)
	if resp.StatusCode != http.StatusOK {
		fiatramps.HandleFailedResponse(resp)
		return CreateWebhookResponse{}, fmt.Errorf("failed response with status: %s", resp.Status)
	}

	var responseData CreateWebhookResponse

	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return CreateWebhookResponse{}, err
	}

	return responseData, nil
}
