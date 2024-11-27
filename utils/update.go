package utils

import (
	"encoding/json"
	"fmt"
	"gopatch/model"
	"log"
	"strings"
	"time"
)

func ProcessMQTTData(
	apiUrl string,
	serviceRoleKey string,
	receivedMessagesJSONChan <-chan string,
	function string,
	trigger string,
	loop float64,
	filter string,
) {
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

			// Start to collect data when trigger specify device
			// collect the data for few seconds, process for further handling method.
			// Change Payloads title or delete the extra devices and etc..
			handleTrigger(jsonPayloads, messages, trigger, loop, filter, apiUrl, serviceRoleKey, function)

			clearCacheAndData(jsonPayloads)
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
	log.Println(output)
}
