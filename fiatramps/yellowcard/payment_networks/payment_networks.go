package payment_networks

import (
	"encoding/json"
	"fmt"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
	"io"
	"log"
	"net/http"
)

type NetworkStatus string

const (
	Active NetworkStatus = "active"
)

type PaymentNetwork struct {
	Name              string        `json:"name"`
	Id                string        `json:"id"`
	Status            NetworkStatus `json:"status"`
	Country           string        `json:"country"`
	AccountNumberType string        `json:"accountNumberType"`
	ChannelIds        []string      `json:"channelIds"`
}

type GetNetworksResponse struct {
	Networks []PaymentNetwork `json:"networks"`
}

type ActiveNetwork struct {
	Name      string `json:"name"`
	NetworkId string `json:"networkId"`
	ChannelId string `json:"channelId"`
}

const networksEndpoint = "/business/networks"

func GetPaymentNetworks(countryCode string, channelId string) ([]ActiveNetwork, error) {
	resp, err := yellowcard.MakeYellowCardRequest("GET", networksEndpoint, nil, map[string]string{
		"country": countryCode,
	})
	if err != nil {
		log.Printf("Failed to send request: %v", err)
		return []ActiveNetwork{}, err
	}

	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			log.Printf("Failed to close body stream: %v", err)
		}
	}(resp.Body)

	log.Printf("Response status for request [%s]: %s", networksEndpoint, resp.Status)
	if resp.StatusCode != http.StatusOK {
		fiatramps.HandleFailedResponse(resp)
		return []ActiveNetwork{}, fmt.Errorf("failed response with status: %s", resp.Status)
	}

	var response GetNetworksResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return []ActiveNetwork{}, err
	}

	// Filter for active status
	var activeNetworks []ActiveNetwork

	// Filter active networks that have `channelId` in the list of `network.channelIds
	for _, network := range response.Networks {
		if network.Status == Active {
			for _, id := range network.ChannelIds {
				if id == channelId {
					activeNetworks = append(activeNetworks, ActiveNetwork{
						Name:      network.Name,
						NetworkId: network.Id,
						ChannelId: id,
					})
					break
				}
			}
		}
	}

	return activeNetworks, nil
}
