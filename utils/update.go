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

			if value, ok := jsonPayloads[trigger]; ok {
				if trigger, ok := value.(float64); ok {
					if trigger == 1 {
						startTime := time.Now()
						for {
							for _, message := range messages {
								fieldNameLower := strings.ToLower(message.Address)
								fieldValue := message.Value
								jsonPayloads[fieldNameLower] = fieldValue
							}
							time.Sleep(time.Second)

							if time.Since(startTime).Seconds() >= loop {
								break
							}
						}

						if _filter, ok := jsonPayloads[filter].(float64); ok && _filter != 0 {
							// The following code will be executed after the loop is broken
							// Call the function to calculate "inklot" and remove "d171", "d172", and "d173"
							calculateAndStoreInklot(jsonPayloads)
							changeName(jsonPayloads)

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

						}
					}
				}
			}

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
	fmt.Println(output)
}

// Define a function to clear the cache and data.
func clearCacheAndData(collectedData JsonPayloads) JsonPayloads {
	// Create a new empty map to replace the existing one.
	return make(JsonPayloads)
}

func calculateAndStoreInklot(jsonPayloads JsonPayloads) {
	d171Value, d171Exists := jsonPayloads["d171"].(string)
	d172Value, d172Exists := jsonPayloads["d172"].(string)
	d173Value, d173Exists := jsonPayloads["d173"].(string)

	var inklotValue string
	if d171Exists && d172Exists && d173Exists {
		inklotValue = d171Value + d172Value + d173Value
	}
	jsonPayloads["ink_lot"] = inklotValue

	// Delete the keys "d171", "d172", and "d173" from the map
	delete(jsonPayloads, "d171")
	delete(jsonPayloads, "d172")
	delete(jsonPayloads, "d173")
}

func changeName(jsonPayloads JsonPayloads) {
	// Define a mapping of key transformations
	keyTransformations := map[string]string{
		"ch1_crtridge_weight_g":   "d162",
		"ch1_filling_weight_g":    "d164",
		"ch1_helium_pressure_kpa": "d166",
		"ch1_head_suction_kpa":    "d167",
		"ch1_flow_rate_ml":        "d168",
		"ch1_cycle_time_sec":      "d170",
		"ch1_error_code":          "d175",
		"ch2_crtridge_weight_g":   "d462",
		"ch2_filling_weight_g":    "d464",
		"ch2_helium_pressure_kpa": "d466",
		"ch2_head_suction_kpa":    "d467",
		"ch2_flow_rate_ml":        "d468",
		"ch2_cycle_time_sec":      "d470",
		"ch2_error_code":          "d475",
		"ch3_crtridge_weight_g":   "d762",
		"ch3_filling_weight_g":    "d764",
		"ch3_helium_pressure_kpa": "d766",
		"ch3_head_suction_kpa":    "d767",
		"ch3_flow_rate_ml":        "d768",
		"ch3_cycle_time_sec":      "d770",
		"ch3_error_code":          "d775",
		"model":                   "d174",
	}

	jsonPayloads["ch1_sequence"] = jsonPayloads["d760"]
	jsonPayloads["ch2_sequence"] = jsonPayloads["d760"]
	jsonPayloads["ch3_sequence"] = jsonPayloads["d760"]

	keysToDelete := []string{
		"d160", "d460", "d760",
	}

	for _, key := range keysToDelete {
		delete(jsonPayloads, key)
	}

	for newKey, oldKey := range keyTransformations {
		jsonPayloads[newKey] = jsonPayloads[oldKey]
		delete(jsonPayloads, oldKey)
	}
}
