package database

import (
	"bytes"
	"fmt"
	"github.com/blocverse01/globachain-ramps-grpc-service/fiatramps"
	"net/http"
)

func MakeTablesRequest(method string, endpoint string, bodyData *bytes.Buffer) (*http.Response, error) {
	var apiKey = fiatramps.GetEnvConfig("XATA_API_KEY")
	var branch = fiatramps.GetEnvConfig("XATA_BRANCH")
	var baseURL = fiatramps.GetEnvConfig("XATA_DATABASE_URL")

	fullURL := fmt.Sprintf("%s:%s/tables/%s", baseURL, branch, endpoint)

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
