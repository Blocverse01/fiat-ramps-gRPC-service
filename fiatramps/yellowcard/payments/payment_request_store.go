package payments

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/database"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/beneficiaries"
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
	UsdAmount             float64 `json:"usdAmount"` // Total USD amount excluding transaction fees.
	YellowCardStatus      string  `json:"yellowCardStatus"`
	TransactionFeeUsd     float64 `json:"transactionFeeUsd"`
}

type StoreBeneficiaryPaymentRequestResponse struct {
	Id string `json:"id"`
}

var tableName = "yellowcard_offramp_transaction"

func StoreBeneficiaryPaymentRequest(input PaymentRequestRecord) (string, error) {
	endpoint := fmt.Sprintf("%s/data", tableName)

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
		_, err := io.Copy(io.Discard, Body) // Discard remaining data
		if err != nil {
			log.Printf("Error discarding response body: %v", err)
		}
		err = Body.Close()
		if err != nil {
			log.Printf("Error closing body: %v", err)
		}
	}(resp.Body)

	// Handle the response
	log.Printf("Response status for request [%s]: %s", endpoint, resp.Status)

	if database.IsUniqueViolation(resp.StatusCode) {
		return "", fmt.Errorf("duplicate payment request for sequenceId or transactionHash")
	}

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
	UsdAmount             float64                                 `json:"usdAmount"`
	YellowCardStatus      string                                  `json:"yellowCardStatus"`
	Id                    string                                  `json:"id"`
	TransactionFeeUsd     float64                                 `json:"transactionFeeUsd"`
	Xata                  database.XataInfo                       `json:"xata"`
}

func GetPaymentRequestTransaction(sequenceId string) (StoredPaymentRequestRecord, error) {
	rec, err := FindPaymentRequestByFilter(map[string]string{"sequenceId": sequenceId})
	if err != nil {
		return StoredPaymentRequestRecord{}, err
	}
	if rec == nil {
		return StoredPaymentRequestRecord{}, fmt.Errorf("no payment request with sequenceId: %v found", sequenceId)
	}
	return *rec, nil
}

func GetPaymentRequestTransactionForOwner(sequenceId, ownerId string) (StoredPaymentRequestRecord, error) {
	if ownerId == "" {
		return StoredPaymentRequestRecord{}, fmt.Errorf("ownerId is required")
	}
	rec, err := FindPaymentRequestByFilter(map[string]string{
		"sequenceId": sequenceId,
		"ownerId":    ownerId,
	})
	if err != nil {
		return StoredPaymentRequestRecord{}, err
	}
	if rec == nil {
		return StoredPaymentRequestRecord{}, fmt.Errorf("no payment request with sequenceId: %v found", sequenceId)
	}
	return *rec, nil
}

func FindPaymentRequestByFilter(filter map[string]string) (*StoredPaymentRequestRecord, error) {
	endpoint := fmt.Sprintf("%s/query", tableName)

	reqBody := database.TableDataQueryRequestBody{
		Columns: []string{"*", "beneficiary"},
		Filter:  filter,
	}

	bodyData, err := fiatramps.BufferBodyData(reqBody)

	if err != nil {
		return nil, err
	}

	resp, err := database.MakeTablesRequest("POST", endpoint, bodyData)

	if err != nil {
		log.Printf("Error making database request: %v", err)
		return nil, err
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
		return nil, fmt.Errorf("failed response with status: %s", resp.Status)
	}

	var response GetPaymentRequestResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return nil, err
	}

	if len(response.Records) == 0 {
		return nil, nil
	}

	return &response.Records[0], nil
}

func patchPaymentRequestStatus(sequenceId string, status string) error {
	paymentRequest, err := GetPaymentRequestTransaction(sequenceId)

	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/data/%s", tableName, paymentRequest.Id)

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
