package handler

import (
	"encoding/json"
	"fmt"

	"gopatch/config"
	"gopatch/model"
	"gopatch/utils"
	"log"
	"strings"
	"time"
)

// Time process to handling data
var stopProcessing = make(chan struct{})

func ProcessMQTTData(
	cfg config.AppConfig,
	receivedMessagesJSONChan <-chan string,
) {
	// Create a persistent session once
	// Use unique key per logical case
	caseKey := cfg.Function + "_" + cfg.Trigger
	session := GetOrCreateSession(caseKey)

	// Create a map to store all JSON payloads
	jsonPayloads := utils.NewSafeJsonPayloads()
	for {
		select {
		case jsonString := <-receivedMessagesJSONChan:
			if jsonString == "" {
				fmt.Println("JSON string is empty")
				continue
			}

			var messages []model.Message

			if err := json.Unmarshal([]byte(jsonString), &messages); err != nil {
				fmt.Printf("Error unmarshaling JSON: %v\n", err)
				// time.Sleep(time.Second)
				continue
			}

			// Prepare JSON payloads for each message
			for _, message := range messages {
				fieldNameLower := strings.ToLower(message.Address)
				fieldValue := message.Value
				jsonPayloads.Set(fieldNameLower, fieldValue)
			}

			// Start to collect data when trigger specify device
			// collect the data for few seconds, process for further handling method.
			// Change Payloads title or delete the extra devices and etc..
			Trigger(session, jsonPayloads, messages, cfg, receivedMessagesJSONChan)
			jsonPayloads.Clear()

			return

		case <-stopProcessing:
			return
		}
	}
}

// To stop the goroutine, you can close the stopProcessing channel:
func StopProcessing() {
	close(stopProcessing)
}

func drainChannel(ch <-chan string) {
	for {
		select {
		case <-ch:
			// Discard the value
		default:
			// Exit when there's nothing left
			return
		}
	}
}

// prettyPrintJSONWithTime handles both map[string]interface{} and *SafeJsonPayloads types
func prettyPrintJSONWithTime(data interface{}, duration time.Duration) {
	// Handle nil data case
	if data == nil {
		log.Println("Error: Provided data is nil.")
		return
	}

	// Determine if data is a map[string]interface{} or *SafeJsonPayloads
	var formatted []byte
	var err error
	switch v := data.(type) {
	case map[string]interface{}:
		// Handle normal map[string]interface{} directly
		formatted, err = json.MarshalIndent(v, "", "  ")
	case *utils.SafeJsonPayloads:
		// Handle *SafeJsonPayloads, extracting the map from it
		formatted, err = json.MarshalIndent(v.GetData(), "", "  ")
	default:
		log.Println("Error: Unsupported data type.")
		return
	}

	if err != nil {
		fmt.Println("Error formatting JSON:", err)
		return
	}

	// Define ANSI escape codes for colors
	greenColor := "\x1b[32m" // Green color for time
	pinkColor := "\x1b[35m"  // Pink color for JSON
	resetColor := "\x1b[0m"  // Reset color to default

	// Convert the time duration to seconds and format it
	elapsedTime := fmt.Sprintf("%s%.2f s%s", greenColor, float64(duration.Seconds()), resetColor)

	// Format the JSON data with pink color
	jsonFormatted := fmt.Sprintf("%s%s%s", pinkColor, string(formatted), resetColor)

	// Concatenate the time and JSON data into a single output string
	output := fmt.Sprintf(">= %s %s", elapsedTime, jsonFormatted)

	// Print the combined output
	log.Println(output)
}
