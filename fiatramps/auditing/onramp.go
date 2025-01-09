package auditing

import (
	"encoding/json"
	"fmt"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/database"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/collection"
	"io"
	"log"
	"net/http"
	"time"
)

func ExportOnrampAuditData(fromTime time.Time, toTime time.Time) []AuditExport {
	transactions := getOnrampTransactions(fromTime, toTime)

	authorityTokens := make([]string, 0)
	sequencedIdAuthorityTokenMap := make(map[string]string)

	for _, transaction := range transactions {
		encodedAuthorityToken, err := encodePayoutAuthorityToken(DecodedPayoutAuthorityToken{
			Reason:               "ONRAMP",
			TriggerTransactionId: transaction.SequenceId,
		})
		if err != nil {
			panic(err)
		}
		authorityTokens = append(authorityTokens, encodedAuthorityToken)
		sequencedIdAuthorityTokenMap[transaction.SequenceId] = encodedAuthorityToken
	}

	payoutTransactions := getCryptoPayoutTransactions(authorityTokens)

	var auditData []AuditExport

	for _, transaction := range transactions {
		transactionHash := searchForPayoutTransaction(sequencedIdAuthorityTokenMap[transaction.SequenceId], payoutTransactions)

		auditData = append(auditData, AuditExport{
			UsdAmount:             transaction.UsdAmount,
			CreatedAt:             transaction.Xata.CreatedAt,
			TransactionFeeUsd:     transaction.TransactionFeeUsd,
			PlatformFeePercentage: transaction.PlatformFeePercentage,
			SequenceId:            transaction.SequenceId,
			TransactionHash:       transactionHash,
			BlockchainNetwork:     transaction.BlockchainNetwork,
			TransactionType:       "Onramp",
		})
	}

	return auditData
}

const onrampTxsTableName = "yellowcard_onramp_transaction"

type TableDataQueryRequestBody struct {
	Columns []string                     `json:"columns"`
	Filter  map[string]map[string]string `json:"filter"`
}

func getOnrampTransactions(fromTime time.Time, toTime time.Time) []collection.RequestRecord {
	endpoint := fmt.Sprintf("%s/query", onrampTxsTableName)

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

	var response collection.GetCollectionRequestResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		panic(fmt.Errorf("failed to decode response body: %v", err))
	}

	return response.Records
}

const cryptoPayoutTxsTableName = "CryptoPayoutOrder"

type PayoutTableDataQueryRequestBody struct {
	Columns []string                       `json:"columns"`
	Filter  map[string]map[string][]string `json:"filter"`
}
type CryptoPayoutTransaction struct {
	PayoutTransactionHash string `json:"payoutTransactionHash"`
	OrderAuthorityToken   string `json:"orderAuthorityToken"`
}
type CryptoPayoutTransactionResponse struct {
	Records []CryptoPayoutTransaction `json:"records"`
}

func getCryptoPayoutTransactions(authorityTokens []string) []CryptoPayoutTransaction {
	endpoint := fmt.Sprintf("%s/query", cryptoPayoutTxsTableName)
	dbUrl := fiatramps.GetEnvConfig("AUDIT_CRYPTO_PAYOUT_XATA_DATABASE_URL")

	if len(authorityTokens) == 0 {
		return []CryptoPayoutTransaction{}
	}

	var filter map[string]map[string][]string

	filter = map[string]map[string][]string{
		"orderAuthorityToken": {
			"$any": authorityTokens,
		},
	}

	reqBody := PayoutTableDataQueryRequestBody{
		Columns: []string{"*"},
		Filter:  filter,
	}

	bodyData, err := fiatramps.BufferBodyData(reqBody)

	if err != nil {
		panic(err)
	}

	resp, err := database.MakeTablesRequest("POST", endpoint, bodyData, AuditDbBranch, dbUrl)

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

	var response CryptoPayoutTransactionResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		panic(fmt.Errorf("failed to decode response body: %v", err))
	}

	return response.Records
}

type DecodedPayoutAuthorityToken struct {
	Reason               string `json:"reason"`
	TriggerTransactionId string `json:"triggerTransactionId"`
}

func encodePayoutAuthorityToken(token DecodedPayoutAuthorityToken) (string, error) {
	authorityTokenKey := fiatramps.GetEnvConfig("CRYPTO_PAYOUT_AUTHORITY_TOKEN_KEY")
	authorityTokenIv := fiatramps.GetEnvConfig("CRYPTO_PAYOUT_AUTHORITY_TOKEN_IV")

	tokenJson, err := json.Marshal(token)

	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return "", err
	}

	encryptedToken, err := EncryptWithAES(string(tokenJson), authorityTokenKey, authorityTokenIv)

	if err != nil {
		fmt.Println("Error encrypting token:", err)
		return "", err
	}

	return encryptedToken, nil
}

func searchForPayoutTransaction(authorityToken string, payoutTransactions []CryptoPayoutTransaction) string {
	for _, payoutTransaction := range payoutTransactions {
		if payoutTransaction.OrderAuthorityToken == authorityToken {
			return payoutTransaction.PayoutTransactionHash
		}
	}

	return ""
}
