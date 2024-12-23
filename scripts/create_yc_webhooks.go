package main

import (
	"fmt"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/yellowcard/webhooks"
)

var rampTypes = [2]string{"payment", "collection"}
var states = [4]string{"complete", "failed", "pending", "processing"}

func createWebhook(rampType string, state string) {
	const webhookBaseUrl = "https://azza-whatsapp-bot.onrender.com/api/fiat-ramp/webhook"
	url := fmt.Sprintf("%s/%s/%s", webhookBaseUrl, rampType, state)

	fmt.Println(url)

	rampTypeState := fmt.Sprintf("%s.%s", rampType, state)

	response, err := webhooks.CreateWebhook(webhooks.CreateWebhookRequest{
		Url:    url,
		State:  rampTypeState,
		Active: true,
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(response)
}

func main() {
	for _, rampType := range rampTypes {
		for _, state := range states {
			createWebhook(rampType, state)
		}
	}
}
