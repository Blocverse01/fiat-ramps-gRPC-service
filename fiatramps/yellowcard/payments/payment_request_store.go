package payments

import (
	"encoding/json"
	"fmt"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/database"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/beneficiaries"
	"io"
	"log"
	"net/http"
)

type PaymentRequestRecord struct {
	Beneficiary           string  `json:"beneficiary"`
	LocalAmount           float64 `json:"localAmount"`
	OwnerId               string  `json:"ownerId"`
	TokenSymbol           string  `json:"tokenSymbol"`
	BlockchainNetwork     string  `json:"blockchainNetwork"`
	TokenAddress          string  `json:"tokenAddress"`
	HotWalletAddress      string  `json:"hotWalletAddress"`
	FromWalletAddress     string  `json:"fromWalletAddress"`
	TransactionHash       string  `json:"transactionHash"`
	PlatformFeePercentage float64 `json:"platformFeePercentage"`
	ConversionRate        float64 `json:"conversionRate"`
	SequenceId            string  `json:"sequenceId"`
	UsdAmount             float64 `json:"usdAmount"` // Total USD amount including transaction fees.
	YellowCardStatus      string  `json:"yellowCardStatus"`
}

type StoreBeneficiaryPaymentRequestResponse struct {
	Id string `json:"id"`
}

var TableName = "yellowcard_offramp_transaction"

func StoreBeneficiaryPaymentRequest(input PaymentRequestRecord) (string, error) {
	endpoint := fmt.Sprintf("%s/data", TableName)

	bodyData, err := fiatramps.BufferBodyData(input)

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

	var response StoreBeneficiaryPaymentRequestResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return "", err
	}

	return response.Id, nil
}

type GetPaymentRequestResponse struct {
	Records []StoredPaymentRequestRecord `json:"records"`
}

type StoredPaymentRequestRecord struct {
	Beneficiary           beneficiaries.AddBeneficiaryRequestBody `json:"beneficiary"`
	LocalAmount           float64                                 `json:"localAmount"`
	OwnerId               string                                  `json:"ownerId"`
	TokenSymbol           string                                  `json:"tokenSymbol"`
	BlockchainNetwork     string                                  `json:"blockchainNetwork"`
	TokenAddress          string                                  `json:"tokenAddress"`
	HotWalletAddress      string                                  `json:"hotWalletAddress"`
	FromWalletAddress     string                                  `json:"fromWalletAddress"`
	TransactionHash       string                                  `json:"transactionHash"`
	PlatformFeePercentage float64                                 `json:"platformFeePercentage"`
	ConversionRate        float64                                 `json:"conversionRate"`
	SequenceId            string                                  `json:"sequenceId"`
	UsdAmount             float64                                 `json:"usdAmount"` // Total USD amount including transaction fees.
	YellowCardStatus      string                                  `json:"yellowCardStatus"`
}

func GetPaymentRequestTransaction(sequenceId string) (StoredPaymentRequestRecord, error) {
	endpoint := fmt.Sprintf("%s/query", TableName)

	var filter map[string]string

	filter = map[string]string{
		"sequenceId": sequenceId,
	}

	reqBody := database.TableDataQueryRequestBody{
		Columns: []string{"*", "beneficiary"},
		Filter:  filter,
	}

	bodyData, err := fiatramps.BufferBodyData(reqBody)

	if err != nil {
		return StoredPaymentRequestRecord{}, err
	}

	resp, err := database.MakeTablesRequest("POST", endpoint, bodyData)

	if err != nil {
		log.Printf("Error making database request: %v", err)
		return StoredPaymentRequestRecord{}, err
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
		return StoredPaymentRequestRecord{}, fmt.Errorf("failed response with status: %s", resp.Status)
	}

	var response GetPaymentRequestResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return StoredPaymentRequestRecord{}, err
	}

	if len(response.Records) == 0 {
		return StoredPaymentRequestRecord{}, fmt.Errorf("no payment request with sequenceId: %v found", sequenceId)
	}

	return response.Records[0], nil
}
