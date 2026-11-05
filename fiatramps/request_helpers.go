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
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Printf("Error closing body %v", err)
			}
		}(resp.Body)

		// Read the response body
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			log.Printf("Failed to read response body: %v", err)
		}

		log.Printf("Request failed with status [%d] (response body redacted)", resp.StatusCode)
		_ = body
	}
}

func BufferBodyData(body any) (*bytes.Buffer, error) {
	// Marshal the request body into JSON
	reqBodyJson, err := json.Marshal(body)

	if err != nil {
		log.Printf("Failed to marshal request body: %v", err)
		return nil, err
	}

	bodyData := bytes.NewBuffer(reqBodyJson)

	return bodyData, nil
}
