package utils

import (
	"encoding/json"
	"fmt"
	"patch/model"
	"strings"
	"sync"
	"time"
)

var mu sync.RWMutex
var stopProcessing = make(chan struct{})

func ProcessMQTTData(apiUrl string, serviceRoleKey string, function string) {
	startTime := time.Now()
	for {
		mu.RLock()
		jsonString := ExportedReceivedMessagesJSON
		mu.RUnlock()

		if jsonString == "" {
			fmt.Println("JSON string is empty")
			time.Sleep(time.Second)
		}

		var messages []model.Message

		if err := json.Unmarshal([]byte(jsonString), &messages); err != nil {
			fmt.Printf("Error unmarshaling JSON: %v\n", err)
			time.Sleep(time.Second)
			continue
		}

		// Create a map to store all JSON payloads
		jsonPayloads := make(map[string]interface{})

		// Prepare JSON payloads for each message
		for _, message := range messages {
			fieldNameLower := strings.ToLower(message.Address)
			fieldValue := message.Value
			jsonPayloads[fieldNameLower] = fieldValue
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

		select {
		case <-stopProcessing:
			return
		default:
			continue
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
