package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/robfig/cron"
)

// In-memory storing of known timestamp
var latestKnownTimestamp int64 = 0

func main() {
	// starts a new cron session
	c := cron.New()
	// Adds the function to run every 10 minutes
	c.AddFunc("* 10 * * *", getTimestamp)
	// starts the cron session
	c.Start()
	for {
		// For using http.Post function
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Fatal(err)
		}
	}

	//getTimestamp()
}

func getTimestamp() {
	// Hardcoded webhook Url, explained in README, PLEASE DO NOT MISSUSE
	webhookURL := "https://discordapp.com/api/webhooks/506125670343245838/KGFKq19syBBPdygeypfnepVJv7wpVraj64f5EebhKUabYVYQ_rqUC1rE9S-e7WJNSl4j"
	response, err := http.Get("https://pure-stream-73485.herokuapp.com/paragliding/api/ticker/latest")
	if err != nil {
		fmt.Println("Error getting latest")
		return
	}
	// Gets the latest known timestamp of the database
	timeCheck, err := ioutil.ReadAll(response.Body)
	tc, _ := strconv.Atoi(string(timeCheck))
	actualTimestamp := int64(tc)

	// Checks if there is any new timestamps
	if actualTimestamp > latestKnownTimestamp {
		fmt.Println("Webhook invoked")

		payload := fmt.Sprintf("New timestamp:  %d", actualTimestamp)

		//puts the string into a struct to encode
		type returnContent struct {
			Payload string `json:"content"`
		}
		payloadStruct := returnContent{
			Payload: payload,
		}
		// Encodes the payload and posts it to a Discord webhook
		temp, err := json.Marshal(payloadStruct)
		if err != nil {
			fmt.Println("fail in marshal")
		}
		//Posts the webhook to discord
		_, err = http.Post(webhookURL, "application/json", bytes.NewBuffer(temp))
		if err != nil {
			fmt.Println("Post failed")
		}
	}
	latestKnownTimestamp = actualTimestamp
}
