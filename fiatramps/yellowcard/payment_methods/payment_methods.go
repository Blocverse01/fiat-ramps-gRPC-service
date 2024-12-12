package payment_methods

import (
	"encoding/json"
	"fmt"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard"
	"io"
	"log"
	"net/http"
)

type ChannelStatus string

const (
	Active   ChannelStatus = "active"
	Inactive ChannelStatus = "inactive"
)

type ChannelType string

const (
	Bank  ChannelType = "bank"
	Momo  ChannelType = "momo"
	Eft   ChannelType = "eft"
	Spenn ChannelType = "spenn"
	P2P   ChannelType = "p2p"
)

type PaymentChannel struct {
	Id          string        `json:"id"`
	ChannelType ChannelType   `json:"channelType"`
	RampType    string        `json:"rampType"`
	Status      ChannelStatus `json:"status"`
	Max         float64       `json:"max"`
	Min         float64       `json:"min"`
}

type GetChannelsResponse struct {
	Channels []PaymentChannel `json:"channels"`
}

type PaymentMethod struct {
	ChannelName string `json:"channelName"`
	ChannelId   string `json:"channelId"`
}

type FiatLimits struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type GetPaymentMethodsResponse struct {
	PaymentMethods []PaymentMethod `json:"paymentMethods"`
	Limits         FiatLimits      `json:"limits"`
}

const channelsEndpoint = "/business/channels"

// GetPaymentMethods retrieves the payment methods for a given country and route.
// It returns the selected payment methods or an error if something goes wrong.
func GetPaymentMethods(countryCode string, route string) (GetPaymentMethodsResponse, error) {
	resp, err := yellowcard.MakeYellowCardRequest("GET", channelsEndpoint, nil, map[string]string{
		"country": countryCode,
	})
	if err != nil {
		log.Printf("Failed to send request: %v", err)
		return GetPaymentMethodsResponse{}, err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			log.Printf("Failed to close body stream: %v", err)
		}
	}(resp.Body)

	log.Printf("Response status for request [%s]: %s", channelsEndpoint, resp.Status)
	if resp.StatusCode != http.StatusOK {
		fiatramps.HandleFailedResponse(resp)
		return GetPaymentMethodsResponse{}, fmt.Errorf("failed response with status: %s", resp.Status)
	}

	var response GetChannelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return GetPaymentMethodsResponse{}, err
	}

	recommendedChannels := GetRecommendedChannels(countryCode, route)
	fetchedChannelsForRoute := filterActiveChannelsByRoute(response.Channels, route)
	recommendedChannelIds := make([]string, len(recommendedChannels))

	for i, channel := range recommendedChannels {
		recommendedChannelIds[i] = channel.ID
	}

	recommendedChannelsForRoute := filterRecommendedChannels(fetchedChannelsForRoute, recommendedChannelIds)
	channelsToUse := recommendedChannelsForRoute

	if len(channelsToUse) == 0 {
		channelsToUse = fetchedChannelsForRoute
	}

	if len(channelsToUse) == 0 {
		return GetPaymentMethodsResponse{
			PaymentMethods: []PaymentMethod{},
			Limits:         FiatLimits{},
		}, nil // Return an empty array if no channels are available
	}

	channelTypeMap := map[ChannelType]string{
		Bank:  "bank",
		Eft:   "bank",
		P2P:   "bank",
		Momo:  "phone",
		Spenn: "phone",
	}

	var bankPaymentMethods []PaymentMethod
	var phonePaymentMethods []PaymentMethod

	for _, channel := range channelsToUse {
		if channelName, ok := channelTypeMap[channel.ChannelType]; ok {
			paymentMethod := PaymentMethod{
				ChannelName: channelName,
				ChannelId:   channel.Id,
			}
			if channelName == "bank" {
				bankPaymentMethods = append(bankPaymentMethods, paymentMethod)
			} else if channelName == "phone" {
				phonePaymentMethods = append(phonePaymentMethods, paymentMethod)
			}
		}
	}

	log.Printf("Bank payment methods: %v", bankPaymentMethods)
	log.Printf("Phone payment methods: %v", phonePaymentMethods)

	var selectedPaymentMethods []PaymentMethod

	if len(bankPaymentMethods) > 0 {
		selectedPaymentMethods = append(selectedPaymentMethods, bankPaymentMethods[0])
	} else {
		log.Printf("No bank payment methods available for route: %s and country: %s", route, countryCode)
	}

	if len(phonePaymentMethods) > 0 {
		selectedPaymentMethods = append(selectedPaymentMethods, phonePaymentMethods[0])
	} else {
		log.Printf("No phone payment methods available for route: %s and country: %s", route, countryCode)
	}

	if len(selectedPaymentMethods) == 0 {
		return GetPaymentMethodsResponse{
			PaymentMethods: []PaymentMethod{},
			Limits:         FiatLimits{},
		}, fmt.Errorf("no payment methods found for route: %s and country: %s", route, countryCode)
	}

	return GetPaymentMethodsResponse{
		PaymentMethods: selectedPaymentMethods,
		Limits: FiatLimits{
			Min: channelsToUse[0].Min,
			Max: channelsToUse[0].Max,
		},
	}, nil
}

func filterActiveChannelsByRoute(channels []PaymentChannel, route string) []PaymentChannel {
	var filteredChannels []PaymentChannel
	for _, channel := range channels {
		if channel.RampType == route && channel.Status == Active {
			filteredChannels = append(filteredChannels, channel)
		}
	}
	return filteredChannels
}

func filterRecommendedChannels(channels []PaymentChannel, recommendedChannels []string) []PaymentChannel {
	recommendedSet := make(map[string]struct{}, len(recommendedChannels))
	for _, id := range recommendedChannels {
		recommendedSet[id] = struct{}{}
	}

	var filteredChannels []PaymentChannel
	for _, channel := range channels {
		if _, ok := recommendedSet[channel.Id]; ok {
			filteredChannels = append(filteredChannels, channel)
		}
	}
	return filteredChannels
}
