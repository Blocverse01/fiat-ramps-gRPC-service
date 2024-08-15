package quotes

import (
	"encoding/json"
	"fmt"
	"github.com/blocverse01/globachain-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/globachain-ramps-grpc-service/fiatramps/database"
	"github.com/blocverse01/globachain-ramps-grpc-service/fiatramps/yellowcard"
	"io"
	"log"
	"net/http"
)

var transactionFeeTableName = "CountryTransactionFees"

type FiatConfig struct {
	CountryCode    string `json:"countryCode"`
	CurrencySymbol string `json:"currencySymbol"`
}

type Rate struct {
	Buy    float32 `json:"buy"`
	Sell   float32 `json:"sell"`
	Locale string  `json:"locale"`
	Code   string  `json:"code"`
}

// RatesResponse Define a struct to match the overall structure of the JSON response
type RatesResponse struct {
	Rates []Rate `json:"rates"`
}

type Quote struct {
	Rate Rate           `json:"rate"`
	Fees TransactionFee `json:"fees"`
}

func GetQuotes(config FiatConfig) Quote {
	rates := GetRates()
	rateForCurrency, err := FilterRatesByCurrency(rates, config.CurrencySymbol)

	if err != nil {
		log.Fatalf("Error getting rates for currency: %v", err)
	}

	transactionFee, txFeeError := GetTransactionFee(config.CountryCode)

	if txFeeError != nil {
		log.Fatalf("Error getting transaction fee for country: %v", config.CountryCode)
	}

	return Quote{
		Rate: rateForCurrency,
		Fees: transactionFee,
	}
}

func GetRates() []Rate {
	const endpoint = "/business/rates"

	resp, err := yellowcard.MakeYellowCardRequest("GET", endpoint, "", map[string]string{})

	if err != nil {
		log.Fatalf("failed to send request: %v", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("Failed to close body stream: %v", err)
		}
	}(resp.Body)

	// Initialize a variable to hold the decoded response
	var response RatesResponse

	// Use json.NewDecoder to decode the response body into the response struct
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatalf("Failed to decode response body: %v", err)
	}

	// Handle the response
	log.Printf("Response status: %s", resp.Status)

	if resp.StatusCode != 200 {
		log.Fatalf("Request failed with response: %v", resp)
	}

	// Return a dummy Rate for now
	return response.Rates
}

func FilterRatesByCurrency(rates []Rate, currencySymbol string) (Rate, error) {
	// Iterate over the rates to find the one that matches the currencySymbol
	for _, rate := range rates {
		if rate.Code == currencySymbol {
			return rate, nil
		}
	}

	// Return an error if no matching rate is found
	return Rate{}, fmt.Errorf("rate for currency symbol %s not found", currencySymbol)
}

func FilterRatesByCurrencies(rates []Rate, currencySymbols []string) []Rate {
	// Create a map to store the desired currency symbols for quick lookup
	symbolSet := make(map[string]bool)
	for _, symbol := range currencySymbols {
		symbolSet[symbol] = true
	}

	// Create a slice to store the filtered rates
	var filteredRates []Rate

	// Iterate over the rates and add the matching ones to the filteredRates slice
	for _, rate := range rates {
		if symbolSet[rate.Code] {
			filteredRates = append(filteredRates, rate)
		}
	}

	// Return the slice of filtered rates
	return filteredRates
}

type GetTransactionFeeRequestBody struct {
	Columns []string          `json:"columns"`
	Filter  map[string]string `json:"filter"`
}

type TransactionFee struct {
	OnrampFeePercentage  float32 `json:"onrampFeePercentage"`
	OfframpFeePercentage float32 `json:"offrampFeePercentage"`
}

type TransactionFeeResponse struct {
	Records []TransactionFee `json:"records"`
}

func GetTransactionFee(countryCode string) (TransactionFee, error) {
	endpoint := fmt.Sprintf("%s/query", transactionFeeTableName)

	reqBody := GetTransactionFeeRequestBody{
		Columns: []string{"onrampFeePercentage", "offrampFeePercentage"},
		Filter: map[string]string{
			"countryCode": countryCode,
		},
	}

	bodyData := fiatramps.BufferBodyData(reqBody)

	resp, err := database.MakeTablesRequest("POST", endpoint, bodyData)

	if err != nil {
		log.Fatalf("Error making database request: %v", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("Error closing body %v", err)
		}
	}(resp.Body)

	// Handle the response
	log.Printf("Response status: %s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		fiatramps.HandleFailedResponse(resp)
	}

	// Initialize a variable to hold the decoded response
	var response TransactionFeeResponse

	// Use json.NewDecoder to decode the response body into the response struct
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatalf("Failed to decode response body: %v", err)
	}

	if len(response.Records) == 0 {
		return TransactionFee{}, fmt.Errorf("transaction fee not found for country code: %v", countryCode)
	}

	return response.Records[0], nil
}
