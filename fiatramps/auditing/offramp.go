package auditing

import (
	"encoding/json"
	"fmt"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/database"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/payments"
	"io"
	"log"
	"net/http"
	"time"
)

func ExportOfframpAuditData(fromTime time.Time, toTime time.Time) []AuditExport {
	transactions := getOfframpTransactions(fromTime, toTime)

	var auditData []AuditExport

	for _, transaction := range transactions {
		auditData = append(auditData, AuditExport{
			UsdAmount:             transaction.UsdAmount,
			CreatedAt:             transaction.Xata.CreatedAt,
			TransactionFeeUsd:     transaction.TransactionFeeUsd,
			PlatformFeePercentage: transaction.PlatformFeePercentage,
			SequenceId:            transaction.SequenceId,
			TransactionHash:       transaction.TransactionHash,
			BlockchainNetwork:     transaction.BlockchainNetwork,
			TransactionType:       "Offramp",
		})
	}

	return auditData
}

const offrampTxsTableName = "yellowcard_offramp_transaction"

func getOfframpTransactions(fromTime time.Time, toTime time.Time) []payments.StoredPaymentRequestRecord {
	endpoint := fmt.Sprintf("%s/query", offrampTxsTableName)

	var filter map[string]map[string]string

	filter = map[string]map[string]string{
		"xata_createdat": {
			"$ge": fromTime.Format("2006-01-02T15:04:05Z"),
			"$le": toTime.Format("2006-01-02T15:04:05Z"),
		},
		"yellowCardStatus": {
			"$is": "complete",
		},
	}

	reqBody := TableDataQueryRequestBody{
		Columns: []string{"*"},
		Filter:  filter,
	}

	bodyData, err := fiatramps.BufferBodyData(reqBody)

	if err != nil {
		panic(err)
	}

	resp, err := database.MakeTablesRequest("POST", endpoint, bodyData, AuditDbBranch)

	if err != nil {
		log.Printf("Error making database request")
		panic(err)
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
		panic(fmt.Errorf("failed response with status: %s", resp.Status))
	}

	var response payments.GetPaymentRequestResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		panic(fmt.Errorf("failed to decode response body: %v", err))
	}

	return response.Records
}
