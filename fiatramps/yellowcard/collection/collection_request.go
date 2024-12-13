package collection

import (
	"encoding/json"
	"fmt"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
	"log"
	"net/http"
)

type Source struct {
	AccountType   string `json:"accountType"`
	NetworkId     string `json:"networkId"`
	AccountNumber string `json:"accountNumber"`
}
type RequestPayload struct {
	LocalAmount  float64                   `json:"localAmount"`
	SequenceId   string                    `json:"sequenceId"`
	ChannelId    string                    `json:"channelId"`
	ForceAccept  bool                      `json:"forceAccept"`
	CustomerType string                    `json:"customerType"`
	Source       Source                    `json:"source"`
	Recipient    yellowcard.RetailCustomer `json:"recipient"`
}

type BankInfo struct {
	Name          string `json:"name"`
	AccountNumber string `json:"accountNumber"`
	AccountName   string `json:"accountName"`
}
type CreateCollectionResponse struct {
	Source          Source                    `json:"source"`
	Recipient       yellowcard.RetailCustomer `json:"recipient"`
	ChannelId       string                    `json:"channelId"`
	SequenceId      string                    `json:"sequenceId"`
	Amount          float64                   `json:"amount"`
	ConvertedAmount float64                   `json:"convertedAmount"`
	Rate            float64                   `json:"rate"`
	Id              string                    `json:"id"`
	Status          string                    `json:"status"`
	Currency        string                    `json:"currency"`
	Country         string                    `json:"country"`
	BankInfo        BankInfo                  `json:"bankInfo"`
}

func (s Source) Validate() error {
	if (s.AccountType != "bank") && (s.AccountType != "momo") {
		return fmt.Errorf("invalid Source AccountType")
	}
	if s.NetworkId == "" && s.AccountType == "momo" {
		return fmt.Errorf("Source.NetworkId is required when using mobile money")
	}
	if s.AccountNumber == "" && s.AccountType == "momo" {
		return fmt.Errorf("Source.AccountNumber is required when using mobile money")
	}

	return nil
}

func (r RequestPayload) Validate() error {
	if r.LocalAmount <= 0 {
		return fmt.Errorf("LocalAmount must be greater than zero")
	}
	if r.SequenceId == "" {
		return fmt.Errorf("SequenceId is required")
	}
	if r.ChannelId == "" {
		return fmt.Errorf("ChannelId is required")
	}
	if r.CustomerType == "" {
		return fmt.Errorf("CustomerType is required")
	}
	if r.CustomerType != "retail" {
		return fmt.Errorf("CustomerType must be retail")
	}

	if err := r.Source.Validate(); err != nil {
		return err
	}
	if err := r.Recipient.Validate(); err != nil {
		return err
	}

	return nil
}

const collectionsEndpoint = "/business/collections"

func CreateCollectionRequest(input RequestPayload) (CreateCollectionResponse, error) {
	if err := input.Validate(); err != nil {
		return CreateCollectionResponse{}, err
	}

	println("Phone in request body: ", input.Recipient.Phone)

	bodyData, err := fiatramps.BufferBodyData(input)
	if err != nil {
		return CreateCollectionResponse{}, err
	}

	resp, err := yellowcard.MakeYellowCardRequest("POST", collectionsEndpoint, bodyData, map[string]string{})
	if err != nil {
		log.Printf("Failed to send request: %v", err)
		return CreateCollectionResponse{}, err
	}

	log.Printf("Response status for request [%s]: %s", collectionsEndpoint, resp.Status)
	if resp.StatusCode != http.StatusCreated {
		fiatramps.HandleFailedResponse(resp)
		return CreateCollectionResponse{}, fmt.Errorf("failed response with status: %s", resp.Status)
	}

	var response CreateCollectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return CreateCollectionResponse{}, err
	}

	return response, nil
}
