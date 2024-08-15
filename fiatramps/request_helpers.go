package fiatramps

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func HandleFailedResponse(resp *http.Response) {
	if resp.StatusCode != http.StatusOK {
		// Read the response body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Failed to read response body: %v", err)
		}

		// Parse the response body as JSON
		var responseBody map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &responseBody); err != nil {
			log.Fatalf("Failed to parse JSON response: %v", err)
		}

		// Log the parsed response
		log.Fatalf("Request failed with status %d and response: %v", resp.StatusCode, responseBody)
	}
}

func BufferBodyData(body any) *bytes.Buffer {
	// Marshal the request body into JSON
	reqBodyJson, err := json.Marshal(body)

	if err != nil {
		log.Fatalf("Failed to marshal request body: %v", err)
	}

	bodyData := bytes.NewBuffer(reqBodyJson)

	return bodyData
}
