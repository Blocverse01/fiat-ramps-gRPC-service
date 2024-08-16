package beneficiaries

import (
	"encoding/json"
	"fmt"
	"github.com/blocverse01/globachain-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/globachain-ramps-grpc-service/fiatramps/database"
	"io"
	"log"
	"net/http"
)

type GetBeneficiariesRequestBody struct {
	Columns []string          `json:"columns"`
	Filter  map[string]string `json:"filter"`
}

type BeneficiaryCountry struct {
	Code           string `json:"code"`
	Country        string `json:"country"`
	CurrencySymbol string `json:"currencySymbol"`
	CurrencyName   string `json:"currencyName"`
	Id             string `json:"id"`
	Flag           string `json:"flag"`
}

type YellowCardBeneficiary struct {
	FirstName          string             `json:"firstName"`
	LastName           string             `json:"lastName"`
	AccountNumber      string             `json:"accountNumber"`
	BankName           string             `json:"bankName"`
	AccountName        string             `json:"accountName"`
	AccountType        string             `json:"accountType"`
	NetworkId          string             `json:"networkId"`
	ChannelId          string             `json:"channelId"`
	Country            BeneficiaryCountry `json:"country"`
	MobileProvider     string             `json:"mobileProvider"`
	MobileProviderCode string             `json:"mobileProviderCode"`
	MobileNumber       string             `json:"mobileNumber"`
	Id                 string             `json:"id"`
}

type SdkBeneficiary struct {
	OwnerId               string                `json:"ownerId"`
	YellowCardBeneficiary YellowCardBeneficiary `json:"yellowCardBeneficiary"`
}

type GetBeneficiariesResponse struct {
	Records []SdkBeneficiary `json:"records"`
}

var beneficiariesTableName = "sdk_user_beneficiary"

func GetBeneficiaries(ownerId string, countryCode string, accountType string) []YellowCardBeneficiary {
	endpoint := fmt.Sprintf("%s/query", beneficiariesTableName)

	var filter map[string]string

	if accountType != "" {
		filter = map[string]string{
			"ownerId":                            ownerId,
			"yellowCardBeneficiary.country.code": countryCode,
			"yellowCardBeneficiary.accountType":  accountType,
		}
	} else {
		filter = map[string]string{
			"ownerId":                            ownerId,
			"yellowCardBeneficiary.country.code": countryCode,
		}
	}

	reqBody := GetBeneficiariesRequestBody{
		Columns: []string{"yellowCardBeneficiary.*", "yellowCardBeneficiary.country.*"},
		Filter:  filter,
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

	var response GetBeneficiariesResponse

	// Use json.NewDecoder to decode the response body into the response struct
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatalf("Failed to decode response body: %v", err)
	}

	var beneficiaries []YellowCardBeneficiary

	for _, beneficiary := range response.Records {
		beneficiaries = append(beneficiaries, beneficiary.YellowCardBeneficiary)
	}

	return beneficiaries
}
