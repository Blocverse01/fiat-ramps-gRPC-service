package collection

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/database"
)

type StoreCollectionRequestInput struct {
	OwnerId                  string  `json:"ownerId"`
	CollectionCountry        string  `json:"collectionCountry"`
	CollectionAccountNumber  string  `json:"collectionAccountNumber"`
	CollectionBankName       string  `json:"collectionBankName"`
	CollectionAccountType    string  `json:"collectionAccountType"`
	CollectionAccountName    string  `json:"collectionAccountName"`
	ChannelId                string  `json:"channelId"`
	SequenceId               string  `json:"sequenceId"`
	BlockchainNetwork        string  `json:"blockchainNetwork"`
	TokenSymbol              string  `json:"tokenSymbol"`
	ConversionRate           float64 `json:"conversionRate"`
	UsdAmount                float64 `json:"usdAmount"`
	LocalAmount              float64 `json:"localAmount"`
	PlatformFeePercentage    float64 `json:"platformFeePercentage"`
	DestinationWalletAddress string  `json:"destinationWalletAddress"`
	NetworkId                string  `json:"networkId"`
	YellowCardStatus         string  `json:"yellowCardStatus"`
	TransactionFeeUsd        float64 `json:"transactionFeeUsd"`
	IsAgentTransaction       bool    `json:"isAgentTransaction"`
}

type RequestRecord struct {
	OwnerId                  string            `json:"ownerId"`
	CollectionCountry        string            `json:"collectionCountry"`
	CollectionAccountNumber  string            `json:"collectionAccountNumber"`
	CollectionBankName       string            `json:"collectionBankName"`
	CollectionAccountType    string            `json:"collectionAccountType"`
	CollectionAccountName    string            `json:"collectionAccountName"`
	ChannelId                string            `json:"channelId"`
	SequenceId               string            `json:"sequenceId"`
	BlockchainNetwork        string            `json:"blockchainNetwork"`
	TokenSymbol              string            `json:"tokenSymbol"`
	ConversionRate           float64           `json:"conversionRate"`
	UsdAmount                float64           `json:"usdAmount"`
	LocalAmount              float64           `json:"localAmount"`
	PlatformFeePercentage    float64           `json:"platformFeePercentage"`
	DestinationWalletAddress string            `json:"destinationWalletAddress"`
	NetworkId                string            `json:"networkId"`
	YellowCardStatus         string            `json:"yellowCardStatus"`
	Id                       string            `json:"id"`
	TransactionFeeUsd        float64           `json:"transactionFeeUsd"`
	IsAgentTransaction       bool    `json:"isAgentTransaction"`
	Xata                     database.XataInfo `json:"xata"`
}

type StoreCollectionRequestResponse struct {
	Id string `json:"id"`
}

var tableName = "yellowcard_onramp_transaction"

func StoreCollectionRequest(input StoreCollectionRequestInput) (string, error) {
	endpoint := fmt.Sprintf("%s/data", tableName)

	bodyData, err := fiatramps.BufferBodyData(input)

	fmt.Println("bodyData:", bodyData)

	if err != nil {
		return "", err
	}

	resp, err := database.MakeTablesRequest("POST", endpoint, bodyData)

	if err != nil {
		log.Printf("Error making database request: %v", err)
		return "", err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing body %v", err)
		}
	}(resp.Body)

	// Handle the response
	log.Printf("Response status for request [%s]: %s", endpoint, resp.Status)

	if resp.StatusCode != http.StatusCreated {
		fiatramps.HandleFailedResponse(resp)
		return "", fmt.Errorf("failed response with status: %s", resp.Status)
	}

	var response StoreCollectionRequestResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return "", err
	}

	return response.Id, nil
}

type GetCollectionRequestResponse struct {
	Records []RequestRecord `json:"records"`
}

func GetCollectionRequestTransaction(sequenceId string) (RequestRecord, error) {
	endpoint := fmt.Sprintf("%s/query", tableName)

	var filter map[string]string

	filter = map[string]string{
		"sequenceId": sequenceId,
	}

	reqBody := database.TableDataQueryRequestBody{
		Columns: []string{"*"},
		Filter:  filter,
	}

	bodyData, err := fiatramps.BufferBodyData(reqBody)

	if err != nil {
		return RequestRecord{}, err
	}

	resp, err := database.MakeTablesRequest("POST", endpoint, bodyData)

	if err != nil {
		log.Printf("Error making database request: %v", err)
		return RequestRecord{}, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing body %v", err)
		}
	}(resp.Body)

	// Handle the response
	log.Printf("Response status: %s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		fiatramps.HandleFailedResponse(resp)
		return RequestRecord{}, fmt.Errorf("failed response with status: %s", resp.Status)
	}

	var response GetCollectionRequestResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return RequestRecord{}, err
	}

	if len(response.Records) == 0 {
		return RequestRecord{}, fmt.Errorf("no collection request with sequenceId: %v found", sequenceId)
	}

	return response.Records[0], nil
}

func UpdateCollectionRequestStatus(sequenceId string, status string) error {
	collectionRequest, err := GetCollectionRequestTransaction(sequenceId)

	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/data/%s", tableName, collectionRequest.Id)

	reqBody := map[string]string{
		"yellowCardStatus": status,
	}

	bodyData, err := fiatramps.BufferBodyData(reqBody)

	if err != nil {
		return err
	}

	resp, err := database.MakeTablesRequest("PATCH", endpoint, bodyData)

	if err != nil {
		log.Printf("Error making database request: %v", err)
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing body %v", err)
		}
	}(resp.Body)

	// Handle the response
	log.Printf("Response status: %s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		fiatramps.HandleFailedResponse(resp)
		return fmt.Errorf("failed response with status: %s", resp.Status)
	}

	return nil
}

type CollectAccountDetails struct {
	CollectionAccountNumber string `json:"collectionAccountNumber"`
	CollectionBankName      string `json:"collectionBankName"`
	CollectionAccountName   string `json:"collectionAccountName"`
}

func UpdateCollectionRequestAccountDetails(sequenceId string, collectionAccount CollectAccountDetails) error {
	collectionRequest, err := GetCollectionRequestTransaction(sequenceId)

	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/data/%s", tableName, collectionRequest.Id)

	reqBody := map[string]string{
		"collectionAccountNumber": collectionAccount.CollectionAccountNumber,
		"collectionBankName":      collectionAccount.CollectionBankName,
		"collectionAccountName":   collectionAccount.CollectionAccountName,
	}

	bodyData, err := fiatramps.BufferBodyData(reqBody)

	if err != nil {
		return err
	}

	resp, err := database.MakeTablesRequest("PATCH", endpoint, bodyData)

	if err != nil {
		log.Printf("Error making database request: %v", err)
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing body %v", err)
		}
	}(resp.Body)

	// Handle the response
	log.Printf("Response status: %s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		fiatramps.HandleFailedResponse(resp)
		return fmt.Errorf("failed response with status: %s", resp.Status)
	}

	return nil
}
