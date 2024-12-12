package payments

import (
	"encoding/json"
	"fmt"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/database"
	"io"
	"log"
	"net/http"
)

type HotWalletAddress struct {
	Id          string `json:"id"`
	NetworkType string `json:"networkType"`
	Address     string `json:"address"`
}

type GetHotWalletAddressResponse struct {
	Records []HotWalletAddress `json:"records"`
}

const hotWalletAddressTable = "hot_wallet_addresses"

func GetHotWalletAddress(networkType string) (string, error) {
	if networkType != "evm" && networkType != "solana" && networkType != "tron" {
		return "", fmt.Errorf("invalid network type")
	}

	reqBody := database.TableDataQueryRequestBody{
		Columns: []string{"*"},
		Filter: map[string]string{
			"networkType": networkType,
		},
	}

	bodyData, err := fiatramps.BufferBodyData(reqBody)

	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("%s/query", hotWalletAddressTable)
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
	log.Printf("Response status: %s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		fiatramps.HandleFailedResponse(resp)
		return "", fmt.Errorf("failed response with status: %s", resp.Status)
	}

	var response GetHotWalletAddressResponse

	// Use json.NewDecoder to decode the response body into the response struct
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return "", err
	}

	if len(response.Records) == 0 {
		return "", fmt.Errorf("no hot wallet address found for network type: %v", networkType)
	}

	return response.Records[0].Address, nil
}
