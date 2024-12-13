package countries

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/database"
	"io"
	"log"
	"net/http"

	pb "github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/xataio/xata-go/xata"
)

var tableName = "yellowcard_supported_country"

func GetSupportedCountries() []*pb.SupportedCountry {
	// Initialize Xata search client
	searchClient, err := xata.NewSearchAndFilterClient()
	if err != nil {
		log.Printf("Failed to create Xata client: %v", err)
		return nil
	}

	// Perform the query
	data, err := searchClient.Query(context.TODO(), xata.QueryTableRequest{
		TableName: tableName,
		Payload: xata.QueryTableRequestPayload{
			Columns: []string{"name", "xata_id", "flag", "currencyName", "currencySymbol", "code"},
			Sort: xata.NewSortExpressionFromStringSortOrderMapList([]map[string]xata.SortOrder{
				{"name": xata.SortOrderAsc},
			}),
		},
	})
	if err != nil {
		log.Printf("Failed to query Xata: %v", err)
		return nil
	}

	// Check if data.Records is nil
	if data.Records == nil {
		log.Printf("No records found in the Xata response")
		return nil
	}

	// Initialize an empty slice to hold the supported countries
	var supportedCountries []*pb.SupportedCountry

	// Iterate over the data and map it to SupportedCountry
	for _, rec := range data.Records {
		// Dereference the pointer and access the underlying map
		record := *rec

		// Safely extract fields and handle missing keys
		countryName, _ := record["name"].(string)
		id, _ := record["xata_id"].(string)
		flag, _ := record["flag"].(string)
		currencyName, _ := record["currencyName"].(string)
		currencySymbol, _ := record["currencySymbol"].(string)
		countryCode, _ := record["code"].(string)

		// Create the SupportedCountry object
		country := &pb.SupportedCountry{
			Id:             id,
			Name:           countryName,
			Code:           countryCode,
			CurrencySymbol: currencySymbol,
			CurrencyName:   currencyName,
			Flag:           flag,
		}

		// Append the pointer to the slice
		supportedCountries = append(supportedCountries, country)
	}

	// Return the list of supported countries
	return supportedCountries
}

type GetCountryRequestBody struct {
	Columns []string          `json:"columns"`
	Filter  map[string]string `json:"filter"`
}
type CountryRecord struct {
	Name           string `json:"name"`
	Code           string `json:"code"`
	Flag           string `json:"flag"`
	CurrencyName   string `json:"currencyName"`
	CurrencySymbol string `json:"currencySymbol"`
	Id             string `json:"id"`
}
type GetCountryResponse struct {
	Records []CountryRecord `json:"records"`
}

func GetCountry(filter map[string]string) (CountryRecord, error) {
	endpoint := fmt.Sprintf("%s/query", tableName)

	reqBody := GetCountryRequestBody{
		Columns: []string{"*"},
		Filter:  filter,
	}

	bodyData, err := pb.BufferBodyData(reqBody)

	if err != nil {
		return CountryRecord{}, err
	}

	resp, err := database.MakeTablesRequest("POST", endpoint, bodyData)

	if err != nil {
		log.Printf("Error making database request: %v", err)
		return CountryRecord{}, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing body %v", err)
		}
	}(resp.Body)

	// Handle the response
	log.Printf("Response status for request [%s]: %s", endpoint, resp.Status)

	if resp.StatusCode != http.StatusOK {
		pb.HandleFailedResponse(resp)
		return CountryRecord{}, fmt.Errorf("failed response with status: %s", resp.Status)
	}

	var response GetCountryResponse

	// Use json.NewDecoder to decode the response body into the response struct
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return CountryRecord{}, err
	}

	if len(response.Records) == 0 {
		return CountryRecord{}, fmt.Errorf("no country with filter: %v found", filter)
	}

	return response.Records[0], nil
}
