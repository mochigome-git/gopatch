package utils

import (
	"encoding/json"
	"fmt"
	"patch/model"
	"strings"
	"time"
)

var stopProcessing = make(chan struct{})

type JsonPayloads map[string]interface{}

func ProcessMQTTData(apiUrl string, serviceRoleKey string, receivedMessagesJSONChan <-chan string, function string) {
	// Create a map to store all JSON payloads
	jsonPayloads := make(JsonPayloads)
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
				time.Sleep(time.Second)
				continue
			}

			// Prepare JSON payloads for each message
			for _, message := range messages {
				fieldNameLower := strings.ToLower(message.Address)
				fieldValue := message.Value
				jsonPayloads[fieldNameLower] = fieldValue
			}

			startTime := time.Now()

			for {
				for _, message := range messages {
					fieldNameLower := strings.ToLower(message.Address)
					fieldValue := message.Value
					jsonPayloads[fieldNameLower] = fieldValue
				}
				time.Sleep(time.Second)

				if time.Since(startTime).Seconds() >= 10 {
					break
				}
			}

			// Marshal the entire JSON data
			jsonData, err := json.Marshal(jsonPayloads)
			if err != nil {
				fmt.Println("Error marshaling JSON:", err)
				return
			}

			// Send the PATCH request using the sendPatchRequest function
			_, err = sendPatchRequest(apiUrl, serviceRoleKey, jsonData, function)
			if err != nil {
				panic(err)
			}
			// Print the formatted JSON payload with the time duration
			elapsedTime := time.Since(startTime)
			prettyPrintJSONWithTime(jsonPayloads, elapsedTime)
			clearCacheAndData(jsonPayloads)

		case <-stopProcessing:
			return
		}
	}
}

// To stop the goroutine, you can close the stopProcessing channel:
func StopProcessing() {
	close(stopProcessing)
}

func prettyPrintJSONWithTime(data map[string]interface{}, duration time.Duration) {
	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Error formatting JSON:", err)
		return
	}

	// Define ANSI escape codes for colors
	greenColor := "\x1b[32m" // Green color
	pinkColor := "\x1b[35m"  // Pink color
	resetColor := "\x1b[0m"  // Reset color to default

	// Convert the time duration to milliseconds
	elapsedTime := fmt.Sprintf("%s%.2f s%s", greenColor, float64(duration.Seconds()), resetColor)

	// Format the JSON data in pink color
	jsonFormatted := fmt.Sprintf("%s%s%s", pinkColor, string(formatted), resetColor)

	// Concatenate the time and JSON data into a single string
	output := fmt.Sprintf(">= %s %s", elapsedTime, jsonFormatted)

	// Print the combined output
	fmt.Println(output)
}

// Define a function to clear the cache and data.
func clearCacheAndData(collectedData JsonPayloads) JsonPayloads {
	// Create a new empty map to replace the existing one.
	return make(JsonPayloads)
}
