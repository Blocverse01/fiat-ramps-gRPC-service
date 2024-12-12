package quotes

import (
	"encoding/json"
	"fmt"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/database"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
	"io"
	"log"
	"net/http"
)

var transactionFeeTableName = "fiat_ramps_transaction_fee"

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

const ratesEndpoint = "/business/rates"

func GetRates() []Rate {
	resp, err := yellowcard.MakeYellowCardRequest("GET", ratesEndpoint, nil, map[string]string{})

	if err != nil {
		log.Printf("failed to send request: %v", err)
		return nil
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Failed to close body stream: %v", err)
		}
	}(resp.Body)

	log.Printf("Response status for request [%s]: %s", ratesEndpoint, resp.Status)
	if resp.StatusCode != http.StatusOK {
		fiatramps.HandleFailedResponse(resp)
		return nil
	}

	// Initialize a variable to hold the decoded response
	var response RatesResponse

	// Use json.NewDecoder to decode the response body into the response struct
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return nil
	}

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
			"country": countryCode,
		},
	}

	bodyData, err := fiatramps.BufferBodyData(reqBody)

	if err != nil {
		return TransactionFee{}, err
	}

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
	log.Printf("Response status for request %s: %s", endpoint, resp.Status)

	if resp.StatusCode != http.StatusOK {
		fiatramps.HandleFailedResponse(resp)
		return TransactionFee{}, fmt.Errorf("failed response with status: %s", resp.Status)
	}

	// Initialize a variable to hold the decoded response
	var response TransactionFeeResponse

	// Use json.NewDecoder to decode the response body into the response struct
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return TransactionFee{}, err
	}

	if len(response.Records) == 0 {
		return TransactionFee{}, fmt.Errorf("transaction fee not found for country code: %v", countryCode)
	}

	return response.Records[0], nil
}
