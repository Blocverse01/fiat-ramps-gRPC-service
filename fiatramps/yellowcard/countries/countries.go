package countries

import (
	"context"
	"log"

	pb "github.com/blocverse01/globachain-ramps-grpc-service/fiatramps"
	"github.com/xataio/xata-go/xata"
)

var tableName = "YellowCardSupportedCurrency"

func GetSupportedCountries() []*pb.SupportedCountry {
	// Initialize Xata search client
	searchClient, err := xata.NewSearchAndFilterClient()
	if err != nil {
		log.Fatalf("Failed to create Xata client: %v", err)
	}

	// Perform the query
	data, err := searchClient.Query(context.TODO(), xata.QueryTableRequest{
		TableName: tableName,
		Payload: xata.QueryTableRequestPayload{
			Columns: []string{"country", "id", "flag", "currencyName", "currencySymbol", "code"},
			Sort: xata.NewSortExpressionFromStringSortOrderMapList([]map[string]xata.SortOrder{
				{"country": xata.SortOrderAsc},
			}),
		},
	})
	if err != nil {
		log.Fatalf("Failed to query Xata: %v", err)
	}

	// Check if data.Records is nil
	if data.Records == nil {
		log.Fatalf("No records found in the Xata response")
	}

	// Initialize an empty slice to hold the supported countries
	var supportedCountries []*pb.SupportedCountry

	// Iterate over the data and map it to SupportedCountry
	for _, rec := range data.Records {
		// Dereference the pointer and access the underlying map
		record := *rec

		// Safely extract fields and handle missing keys
		countryName, _ := record["country"].(string)
		id, _ := record["id"].(string)
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
