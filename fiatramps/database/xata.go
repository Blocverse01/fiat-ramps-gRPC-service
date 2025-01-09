package database

import (
	"bytes"
	"fmt"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"net/http"
)

func MakeTablesRequest(method string, endpoint string, bodyData *bytes.Buffer, branchAndDbUrl ...string) (*http.Response, error) {
	var apiKey = fiatramps.GetEnvConfig("XATA_API_KEY")

	var branchValue string
	var dbUrl string

	if len(branchAndDbUrl) > 0 {
		branchValue = branchAndDbUrl[0]

		if len(branchAndDbUrl) > 1 {
			dbUrl = branchAndDbUrl[1]
		}
	}

	if branchValue == "" {
		branchValue = fiatramps.GetEnvConfig("XATA_BRANCH")
	}
	if dbUrl == "" {
		dbUrl = fiatramps.GetEnvConfig("XATA_DATABASE_URL")
	}

	fullURL := fmt.Sprintf("%s:%s/tables/%s", dbUrl, branchValue, endpoint)

	req, err := http.NewRequest(method, fullURL, bodyData)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// Use an HTTP client to send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type TableDataQueryRequestBody struct {
	Columns []string          `json:"columns"`
	Filter  map[string]string `json:"filter"`
}

type XataInfo struct {
	CreatedAt string `json:"createdAt"`
}
