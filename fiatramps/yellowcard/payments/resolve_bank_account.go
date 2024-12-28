package payments

import (
	"encoding/json"
	"fmt"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
	"log"
	"net/http"
)

type ResolveBankAccountInput struct {
	AccountNumber string `json:"accountNumber"`
	NetworkId     string `json:"networkId"`
}

type ResolveBankAccountResponse struct {
	AccountName   string `json:"accountName"`
	AccountNumber string `json:"accountNumber"`
	AccountBank   string `json:"accountBank"`
}

const resolveBankAccountEndpoint = "/business/details/bank"

func ResolveBankAccount(input ResolveBankAccountInput) (ResolveBankAccountResponse, error) {
	bodyData, err := fiatramps.BufferBodyData(input)

	if err != nil {
		return ResolveBankAccountResponse{}, err
	}

	resp, err := yellowcard.MakeYellowCardRequest("POST", resolveBankAccountEndpoint, bodyData, map[string]string{})
	if err != nil {
		log.Printf("Failed to send request: %v", err)
		return ResolveBankAccountResponse{}, err
	}

	log.Printf("Response status for request [%s]: %s", resolveBankAccountEndpoint, resp.Status)
	if resp.StatusCode != http.StatusOK {
		fiatramps.HandleFailedResponse(resp)
		return ResolveBankAccountResponse{}, fmt.Errorf("failed response with status: %s", resp.Status)
	}

	var responseData ResolveBankAccountResponse

	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return ResolveBankAccountResponse{}, err
	}

	return responseData, nil
}
