package payments

import (
	"encoding/json"
	"fmt"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
	"log"
	"net/http"
)

type PaymentRequestDestination struct {
	AccountNumber string `json:"accountNumber"`
	AccountType   string `json:"accountType"`
	NetworkId     string `json:"networkId"`
	NetworkName   string `json:"networkName"`
	Country       string `json:"country"`
	AccountName   string `json:"accountName"`
}

type PaymentRequestInput struct {
	LocalAmount  float64                   `json:"localAmount"`
	SequenceId   string                    `json:"sequenceId"`
	ChannelId    string                    `json:"channelId"`
	Reason       string                    `json:"reason"`
	Sender       yellowcard.RetailCustomer `json:"sender"`
	Destination  PaymentRequestDestination `json:"destination"`
	ForceAccept  bool                      `json:"forceAccept"`
	CustomerType string                    `json:"customerType"`
}

type CreatePaymentRequestResponse struct {
	Id              string                    `json:"id"`
	ChannelId       string                    `json:"channelId"`
	SequenceId      string                    `json:"sequenceId"`
	Currency        string                    `json:"currency"`
	Country         string                    `json:"country"`
	Amount          float64                   `json:"amount"` // USD amount
	Reason          string                    `json:"reason"`
	ConvertedAmount float64                   `json:"convertedAmount"`
	Status          string                    `json:"status"`
	Rate            float64                   `json:"rate"`
	Sender          yellowcard.RetailCustomer `json:"sender"`
	Destination     PaymentRequestDestination `json:"destination"`
}

const paymentRequestEndpoint = "/business/payments"

func CreatePaymentRequest(input PaymentRequestInput) (CreatePaymentRequestResponse, error) {
	bodyData, err := fiatramps.BufferBodyData(input)

	if err != nil {
		return CreatePaymentRequestResponse{}, err
	}

	resp, err := yellowcard.MakeYellowCardRequest("POST", paymentRequestEndpoint, bodyData, map[string]string{})
	if err != nil {
		log.Printf("Failed to send request: %v", err)
		return CreatePaymentRequestResponse{}, err
	}

	log.Printf("Response status for request [%s]: %s", paymentRequestEndpoint, resp.Status)
	if resp.StatusCode != http.StatusCreated {
		fiatramps.HandleFailedResponse(resp)
		return CreatePaymentRequestResponse{}, fmt.Errorf("failed response with status: %s", resp.Status)
	}

	var responseData CreatePaymentRequestResponse

	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return CreatePaymentRequestResponse{}, err
	}

	return responseData, nil
}
