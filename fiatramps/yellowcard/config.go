package yellowcard

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"net/http"
	"net/url"
	"time"
	"unicode"
)

type AuthenticationCredentials struct {
	Date          string
	Authorization string
}

func generateAuthenticationCredentials(path string, method string, body *bytes.Buffer) AuthenticationCredentials {
	var publicKey = fiatramps.GetEnvConfig("YELLOW_CARD_PUBLIC_KEY")
	var secretKey = fiatramps.GetEnvConfig("YELLOW_CARD_SECRET_KEY")

	var date = time.Now().UTC().Format("2006-01-02T15:04:05Z07:00")
	var message = fmt.Sprintf("%s%s%s", date, path, method)

	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(secretKey))

	// If there is a body, hash it and add to HMAC
	if len(body.String()) > 0 {
		bodyHash := sha256.Sum256(body.Bytes())
		bodyHmac := base64.StdEncoding.EncodeToString(bodyHash[:])
		message += bodyHmac
	}

	h.Write([]byte(message))

	// Get the final HMAC as a base64 string
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	authorization := fmt.Sprintf("YcHmacV1 %s:%s", publicKey, signature)

	return AuthenticationCredentials{
		Date:          date,
		Authorization: authorization,
	}
}

type RequestOptions struct {
	Method  string
	Headers map[string]string
}

type RetailCustomer struct {
	Name               string `json:"name"`
	Phone              string `json:"phone"`
	Country            string `json:"country"`
	Address            string `json:"address"`
	Dob                string `json:"dob"`
	IdNumber           string `json:"idNumber"`
	IdType             string `json:"idType"`
	AdditionalIdType   string `json:"additionalIdType"`
	AdditionalIdNumber string `json:"additionalIdNumber"`
}

func (c RetailCustomer) Validate() (RetailCustomer, error) {
	if c.Name == "" {
		return RetailCustomer{}, fmt.Errorf("Sender.Name is required")
	}
	if c.Country == "" {
		return RetailCustomer{}, fmt.Errorf("Sender.Country is required")
	}
	transformedDate, err := TransformDateString(c.Dob)
	if err != nil {
		return RetailCustomer{}, fmt.Errorf("Sender.Dob should be a valid date: %v", err)
	}
	c.Dob = transformedDate
	if string(c.Phone[0]) != "+" {
		return RetailCustomer{}, fmt.Errorf("Sender.Phone should be in international format and start with a '+'")
	}
	if c.Phone == "" || len(c.Phone) < 8 || len(c.Phone) > 15 {
		return RetailCustomer{}, fmt.Errorf("Sender.Phone should be a valid phone number with 8-15 digits")
	}
	for _, r := range c.Phone[1:] {
		if !unicode.IsDigit(r) {
			return RetailCustomer{}, fmt.Errorf("Sender.Phone should contain only digits")
		}
	}

	// Todo: Validate acceptable ID Types and their corresponding lengths
	if c.IdType == "" {
		return RetailCustomer{}, fmt.Errorf("Sender.IdType is required")
	}
	if c.IdNumber == "" {
		return RetailCustomer{}, fmt.Errorf("Sender.IdNumber is required")
	}

	return c, nil
}

func CreateRequestOptions(path string, method string, body *bytes.Buffer) RequestOptions {
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
func MakeYellowCardRequest(method, endpoint string, body *bytes.Buffer, queryParams map[string]string) (*http.Response, error) {
	apiBaseUrl := fiatramps.GetEnvConfig("YELLOWCARD_BASE_URL")
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

	if body == nil {
		body = bytes.NewBuffer([]byte{}) // This ensures an empty body if needed
	}

	requestOptions := CreateRequestOptions(endpoint, method, body)

	// Create a new HTTP request
	req, err := http.NewRequest(requestOptions.Method, requestUrl, body)
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
