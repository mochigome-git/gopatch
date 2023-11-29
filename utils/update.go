package utils

import (
	"encoding/json"
	"fmt"
	"patch/model"
	"strings"
	"time"
)

var (
	stopProcessing     = make(chan struct{})
	deviceStartTimeMap = make(map[string]time.Time)
	// To store process-specific previous trigger keys
	processPrevTriggerKeyMap = make(map[string]string)
)

// Define the device struct with the address field
type TriggerKey struct {
	triggerKey string
	caseKey    string
}

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

			//fmt.Println(jsonPayloads)

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
		// それぞれの文字列を逆転して連結します
		inklotValue = reverseString(d171Value) + reverseString(d172Value) + reverseString(d173Value)
	}
	jsonPayloads["ink_lot"] = inklotValue

	// "d171"、"d172"、および"d173"のキーをマップから削除
	delete(jsonPayloads, "d171")
	delete(jsonPayloads, "d172")
	delete(jsonPayloads, "d173")
}

// Function to replace device's name to readable key
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

	// Repeat channel 1's sequence count (PLC's device name) for channel 2 and channel 3.
	jsonPayloads["ch1_sequence"] = jsonPayloads["d760"]
	jsonPayloads["ch2_sequence"] = jsonPayloads["d760"]
	jsonPayloads["ch3_sequence"] = jsonPayloads["d760"]
	// Remove Channel 1,2,3 key after process
	keysToDelete := []string{"d160", "d460", "d760"}
	for _, key := range keysToDelete {
		delete(jsonPayloads, key)
	}

	for newKey, oldKey := range keyTransformations {
		// Check if the old key exists before replacing it
		if value, oldKeyExists := jsonPayloads[oldKey]; oldKeyExists {
			jsonPayloads[newKey] = value
			delete(jsonPayloads, oldKey)
		}
	}
}

// Input and returns a new string with its characters reversed
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// 指定時間内でキーを受信し、マップに保持して出力する。同等なキーが繰り返されて受信する場合、上書きされていく
func processMessagesLoop(jsonPayloads JsonPayloads, messages []model.Message, startTime time.Time, loop float64) {
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
}

// トリガーキーの羅列からキーとケースナンバーを分割する
func parseTriggerKey(triggerKey string) []TriggerKey {
	triggerKeySlice := strings.Split(triggerKey, ",")
	var triggerkeys []TriggerKey

	for i := 0; i < len(triggerKeySlice); i += 2 {
		caseNumber := triggerKeySlice[i+1]

		triggerkeys = append(triggerkeys, TriggerKey{
			triggerKey: triggerKeySlice[i],
			caseKey:    fmt.Sprint(caseNumber),
		})
	}

	return triggerkeys
}

// Function to generate a unique key for each process based on relevant parameters
func generateProcessKey(triggerKey string) string {
	// You can concatenate relevant parameters to create a unique key
	return triggerKey /* + other parameters as needed */
}

func handleTrigger(
	jsonPayloads JsonPayloads,
	messages []model.Message,
	triggerKey string,
	loop float64,
	filter string,
	apiUrl string,
	serviceRoleKey string,
	function string,
) {

	// Parse Data to trigger device, case option
	// Splitting the triggerKey into a slice of strings
	triggerkeys := parseTriggerKey(triggerKey)

	// Loop the triggerkey and option to case
	for _, tk := range triggerkeys {
		switch {
		case strings.Contains(tk.caseKey, "time.duration"):
			processKey := generateProcessKey(tk.triggerKey)

			// Check if triggerKey is different from the previous one
			if tk.triggerKey != processPrevTriggerKeyMap[processKey] {
				processPrevTriggerKeyMap[processKey] = tk.triggerKey

				fmt.Println(processPrevTriggerKeyMap)

				if trigger, ok := jsonPayloads[tk.triggerKey].(float64); ok && trigger != 0 {
					fmt.Printf("Device name: %s, Payload: %v\n", tk.triggerKey, jsonPayloads[tk.triggerKey])

					// Check if device start time is not set
					if _startTime, exists := deviceStartTimeMap[tk.triggerKey]; !exists {
						// Set the start time for the device
						deviceStartTimeMap[tk.triggerKey] = time.Now()
					} else {

						// Calculate the duration from 1 to 0
						duration := time.Since(_startTime).Seconds()
						fmt.Println("Duration for device", tk.triggerKey, ":", duration)

						// Check if the value changed from 1 to 0
						if trigger, ok := jsonPayloads[tk.triggerKey].(float64); ok && trigger == 0 {
							// The following code will be executed after the loop is broken
							// Call the function to calculate "inklot" and remove "d171", "d172", and "d173"
							calculateAndStoreInklot(jsonPayloads)
							changeName(jsonPayloads)

							// Call the function to process messages in a loop
							processMessagesLoop(jsonPayloads, messages, deviceStartTimeMap[tk.triggerKey], loop)

							//jsonData, err := json.Marshal(jsonPayloads)
							//if err != nil {
							//    fmt.Println("Error marshaling JSON:", err)
							//    return
							//}
							//
							//// Send the PATCH request using the sendPatchRequest function
							//_, err = sendPatchRequest(apiUrl, serviceRoleKey, jsonData, function)
							//if err != nil {
							//    panic(err)
							//}
							//
							//// Print the formatted JSON payload with the time duration
							//elapsedTime := time.Since(startTime)
							//prettyPrintJSONWithTime(jsonPayloads, elapsedTime)
						}

						// Reset the start time for the device
						deviceStartTimeMap[tk.triggerKey] = time.Now()
					}
				}
			}

		case tk.caseKey == "standard":
			// Check if triggerKey is different from the previous one
			processKey := generateProcessKey(tk.triggerKey)

			// Check if triggerKey is different from the previous one
			if tk.triggerKey != processPrevTriggerKeyMap[processKey] {
				processPrevTriggerKeyMap[processKey] = tk.triggerKey

				if trigger, ok := jsonPayloads[tk.triggerKey].(float64); ok && trigger != 0 {
					// Store the time when the trigger transitions from 1 to 0
					var startTime time.Time

					// Call the function to process messages in a loop
					processMessagesLoop(jsonPayloads, messages, startTime, loop)

					// The following code will be executed after the loop is broken
					// Call the function to calculate "inklot" and remove "d171", "d172", and "d173"
					calculateAndStoreInklot(jsonPayloads)
					changeName(jsonPayloads)

					// Check if the value changed from 1 to 0
					if trigger, ok := jsonPayloads[tk.triggerKey].(float64); ok && trigger == 0 {
						fmt.Println("Case 1")
						fmt.Println(jsonPayloads)

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

		case tk.caseKey == "trigger3":
			// Handle specific logic for trigger3
			// This case is for the third condition
			// Add code to patch data based on the third condition
		}
	}
}
