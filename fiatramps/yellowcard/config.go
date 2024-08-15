package yellowcard

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/blocverse01/globachain-ramps-grpc-service/fiatramps"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// apiBaseUrl YellowCardApiBaseUrl YellowCard config
const apiBaseUrl = "https://api.yellowcard.io"

type AuthenticationCredentials struct {
	Date          string
	Authorization string
}

func generateAuthenticationCredentials(path string, method string, body string) AuthenticationCredentials {
	var apiKey = fiatramps.GetEnvConfig("YELLOW_CARD_API_KEY")
	var secretKey = fiatramps.GetEnvConfig("YELLOW_CARD_SECRET_KEY")

	date := time.Now().UTC().Format(time.RFC1123)

	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(secretKey))

	// Write the data to the HMAC object
	h.Write([]byte(date))
	h.Write([]byte(path))
	h.Write([]byte(method))

	// If there is a body, hash it and add to HMAC
	if len(body) > 0 {
		bodyBytes, _ := json.Marshal(body)
		bodyHash := sha256.Sum256(bodyBytes)
		bodyHmac := base64.StdEncoding.EncodeToString(bodyHash[:])
		h.Write([]byte(bodyHmac))
	}

	// Get the final HMAC as a base64 string
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	authorization := fmt.Sprintf("YcHmacV1 %s:%s", apiKey, signature)

	return AuthenticationCredentials{
		Date:          date,
		Authorization: authorization,
	}
}

type RequestOptions struct {
	Method  string
	Headers map[string]string
}

func CreateRequestOptions(path string, method string, body string) RequestOptions {
	credentials := generateAuthenticationCredentials(path, method, body)

	headers := map[string]string{
		"accept":         "application/json",
		"content-type":   "application/json",
		"X-YC-Timestamp": credentials.Date,
		"Authorization":  credentials.Authorization,
	}

	return RequestOptions{
		Method:  method,
		Headers: headers,
	}
}

// MakeYellowCardRequest is a reusable wrapper for making YellowCard API requests.
func MakeYellowCardRequest(method, endpoint, body string, queryParams map[string]string) (*http.Response, error) {
	requestUrl := apiBaseUrl + endpoint

	// Add query parameters to the URL
	if len(queryParams) > 0 {
		u, err := url.Parse(requestUrl)
		if err != nil {
			return nil, err
		}
		q := u.Query()
		for key, value := range queryParams {
			q.Add(key, value)
		}
		u.RawQuery = q.Encode()
		requestUrl = u.String()
	}

	requestOptions := CreateRequestOptions(endpoint, method, body)

	// Convert the body string to an io.Reader
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	// Create a new HTTP request
	req, err := http.NewRequest(requestOptions.Method, requestUrl, bodyReader)
	if err != nil {
		return nil, err
	}

	// Add headers to the request
	for key, value := range requestOptions.Headers {
		req.Header.Add(key, value)
	}

	// Use an HTTP client to send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
