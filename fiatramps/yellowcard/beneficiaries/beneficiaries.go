package beneficiaries

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/database"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/countries"
)

type GetBeneficiariesRequestBody struct {
	Columns []string          `json:"columns"`
	Filter  map[string]string `json:"filter"`
	Page    map[string]any    `json:"page"`
}

type YellowCardBeneficiary struct {
	AccountNumber        string `json:"accountNumber"`
	AccountName          string `json:"accountName"`
	AccountType          string `json:"accountType"`
	NetworkId            string `json:"networkId"`
	NetworkName          string `json:"networkName"`
	ChannelId            string `json:"channelId"`
	Country              string `json:"country"`
	Id                   string `json:"id"`
	OwnerId              string `json:"ownerId"`
	NameEnquiryReference string `json:"nameEnquiryReference"`
	SeerBitId            string `json:"seerbitId"`
}

type GetBeneficiariesResponse struct {
	Records []YellowCardBeneficiary `json:"records"`
}

var beneficiariesTableName = "yellowcard_beneficiary"

func GetBeneficiaries(ownerId string, countryCode string, accountType string) ([]YellowCardBeneficiary, error) {
	endpoint := fmt.Sprintf("%s/query", beneficiariesTableName)

	var filter map[string]string

	if accountType != "" {
		filter = map[string]string{
			"ownerId":     ownerId,
			"country":     countryCode,
			"accountType": accountType,
		}
	} else {
		filter = map[string]string{
			"ownerId": ownerId,
			"country": countryCode,
		}
	}

	reqBody := GetBeneficiariesRequestBody{
		Columns: []string{"*"},
		Filter:  filter,
		Page: map[string]any{
			"size": 200,
		},
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

	var response GetBeneficiariesResponse

	// Use json.NewDecoder to decode the response body into the response struct
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return nil, err
	}

	return response.Records, nil
}

func GetBeneficiary(beneficiaryId string) (YellowCardBeneficiary, error) {
	endpoint := fmt.Sprintf("%s/query", beneficiariesTableName)

	reqBody := GetBeneficiariesRequestBody{
		Columns: []string{"*", "country"},
		Filter: map[string]string{
			"id": beneficiaryId,
		},
	}

	bodyData, err := fiatramps.BufferBodyData(reqBody)

	if err != nil {
		return YellowCardBeneficiary{}, err
	}

	resp, err := database.MakeTablesRequest("POST", endpoint, bodyData)

	if err != nil {
		log.Printf("Error making database request: %v", err)
		return YellowCardBeneficiary{}, err
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
		fiatramps.HandleFailedResponse(resp)
		return YellowCardBeneficiary{}, fmt.Errorf("failed response with status: %s", resp.Status)
	}

	var response GetBeneficiariesResponse

	// Use json.NewDecoder to decode the response body into the response struct
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return YellowCardBeneficiary{}, err
	}

	if len(response.Records) == 0 {
		return YellowCardBeneficiary{}, fmt.Errorf("no beneficiary with ID: %v found", beneficiaryId)
	}

	return response.Records[0], nil
}

type AddBeneficiaryRequestBody struct {
	AccountNumber        string `json:"accountNumber"`
	AccountName          string `json:"accountName"`
	AccountType          string `json:"accountType"`
	NetworkId            string `json:"networkId"`
	NetworkName          string `json:"networkName"`
	ChannelId            string `json:"channelId"`
	Country              string `json:"country"`
	OwnerId              string `json:"ownerId"`
	NameEnquiryReference string `json:"nameEnquiryReference"`
	SeerBitId            string `json:"seerbitId"`
	
}

type AddBeneficiaryResponse struct {
	Id string `json:"id"`
}

func AddBeneficiary(beneficiary AddBeneficiaryRequestBody) (string, error) {

	fmt.Print("got here")

	_, err := countries.GetCountry(map[string]string{
		"code": beneficiary.Country,
	})

	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("%s/data", beneficiariesTableName)

	bodyData, err := fiatramps.BufferBodyData(beneficiary)

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

	var response AddBeneficiaryResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return "", err
	}

	return response.Id, nil
}
