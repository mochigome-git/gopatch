package utils

import (
	"fmt"
	"gopatch/model"
	"os"
	"strings"
	"time"
)

// processMessagesLoop receives messages within a specified time and updates a JSON payload map.
// If a key is repeated, it overwrites the existing value.
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

// processMessagesLoop receives messages within a specified time,
// handle average value and highest value of message and updates a JSON payload map.
// If a key is repeated, it overwrites the existing value.
// Using for Case Special
// ProcessTriggerGenericSpecial is a generic function to process trigger key and return the corresponding processed payload
func ProcessTriggerGenericSpecial(jsonPayloads JsonPayloads, messages []model.Message, loop float64, changeNameFunc func(JsonPayloads) map[string]interface{}) map[string]interface{} {
	startTime := time.Now()
	processMessagesLoop(jsonPayloads, messages, startTime, 1)
	processedPayload := changeNameFunc(jsonPayloads)

	return processedPayload
}

// parseTriggerKey splits a string of trigger keys and case numbers, returning a slice of TriggerKey structs.
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

// generateProcessKey creates a unique key for each process based on relevant parameters.
func generateProcessKey(triggerKey string) string {
	// You can concatenate relevant parameters to create a unique key
	return triggerKey /* + other parameters as needed */
}

// clearCacheAndData replaces the existing data map with a new empty one.
func clearCacheAndData(collectedData JsonPayloads) JsonPayloads {
	// Create a new empty map to replace the existing one.
	return make(JsonPayloads)
}

// calculateAndStoreInklot computes and stores an 'ink_lot' value based on specific keys in the JSON payload.
func calculateAndStoreInklot(jsonPayloads JsonPayloads) {
	d171Value, d171Exists := jsonPayloads["d171"].(string)
	d172Value, d172Exists := jsonPayloads["d172"].(string)
	d173Value, d173Exists := jsonPayloads["d173"].(string)

	var inklotValue string
	if d171Exists && d172Exists && d173Exists {
		// Concatenate reversed strings of d171, d172, and d173
		inklotValue = reverseString(d171Value) + reverseString(d172Value) + reverseString(d173Value)
	}
	jsonPayloads["ink_lot"] = inklotValue

	// Remove "d171", "d172", and "d173" keys from the map
	delete(jsonPayloads, "d171")
	delete(jsonPayloads, "d172")
	delete(jsonPayloads, "d173")
}

// changeName replaces device names in the JSON payload with readable keys.
func changeName(jsonPayloads JsonPayloads) {
	// Define a mapping of key transformations
	keyTransformations := GetKeyTransformationsFromEnv("KEY_TRANSFORMATION_")

	// Repeat channel 1's sequence count (PLC's device name) for channel 2 and channel 3.
	jsonPayloads["ch1_sequence"] = jsonPayloads["d760"]
	jsonPayloads["ch2_sequence"] = jsonPayloads["d760"]
	jsonPayloads["ch3_sequence"] = jsonPayloads["d760"]
	// Remove Channel 1, 2, 3 keys after processing
	keysToDelete := []string{"d160", "d460", "d760"}
	for _, key := range keysToDelete {
		delete(jsonPayloads, key)
	}

	// Iterate through key transformations and apply them, deleting old keys during transformation
	for newKey, oldKey := range keyTransformations {
		// Replace old key with new key if the old key exists, delete old key
		if value, oldKeyExists := jsonPayloads[oldKey]; oldKeyExists {
			jsonPayloads[newKey] = value
			delete(jsonPayloads, oldKey)
		}
	}
}

// ProcessTriggerGeneric is a generic function to process trigger key and return the corresponding processed payload
func ProcessTriggerGeneric(jsonPayloads JsonPayloads, messages []model.Message, loop float64, changeNameFunc func(JsonPayloads) map[string]interface{}) map[string]interface{} {
	startTime := time.Now()
	processMessagesLoop(jsonPayloads, messages, startTime, 1)
	calculateAndStoreInklot(jsonPayloads)
	processedPayload := changeNameFunc(jsonPayloads)

	return processedPayload
}

// Procees to assigning the common logic to a function and then call that function inside each case
// Handle the common logic for case string and float64; for CASE 4
func processAndPrint(key string, jsonPayloads JsonPayloads, messages []model.Message, loop float64) {
	processedPayloadsMap[key] = ProcessTriggerGeneric(jsonPayloads, messages, loop, func(payload JsonPayloads) map[string]interface{} {
		updatedMap := _hold_changeName_generic(payload, "HOLD_KEY_TRANSOFRMATION_"+key)

		// Define the keys to check
		keysToCheck := []string{"ch3_weighing", "ch1_weighing", "ch2_weighing"}

		// Use the helper function to compare and update the nested map
		CompareAndUpdateNestedMap(processedPayloadsMap, key, updatedMap, keysToCheck)

		return updatedMap
	})
	// fmt.Println(processedPayloadsMap[key])
}

// Process for CASE 1, check the time taken from 0 to 1.
func handleTimeDurationTrigger(tk TriggerKey, jsonPayloads JsonPayloads, messages []model.Message, loop float64, filter string, apiUrl string, serviceRoleKey string, function string) {
	fmt.Printf("Device name: %s, Payload: %v\n", tk.triggerKey, jsonPayloads[tk.triggerKey])

	if startTime, exists := deviceStartTimeMap[tk.triggerKey]; !exists {
		deviceStartTimeMap[tk.triggerKey] = time.Now()
	} else {
		//duration := time.Since(startTime).Seconds()

		if trigger, ok := jsonPayloads[tk.triggerKey].(float64); ok && trigger == 0 {
			calculateAndStoreInklot(jsonPayloads)
			changeName(jsonPayloads)
			processMessagesLoop(jsonPayloads, messages, startTime, loop)
		}

		deviceStartTimeMap[tk.triggerKey] = time.Now()
	}
}

// Process for weight triggers (CH1, CH2, CH3); for CASE 7
func ProcessWeightTriggers(jsonPayloads JsonPayloads, messages []model.Message, loop float64) {
	// A helper function to process each weight trigger
	processWeightTrigger := func(channel string, triggerKey string, weightTrigger *bool, prevWeightTrigger *bool) {
		triggerValue := jsonPayloads[os.Getenv(triggerKey)]
		switch v := triggerValue.(type) {
		case string:
			if v == "1" {
				processAndPrint(channel, jsonPayloads, messages, loop)
				*weightTrigger = true
				*prevWeightTrigger = true
			} else {
				*weightTrigger = false
			}
		case float64:
			if v == 1 {
				processAndPrint(channel, jsonPayloads, messages, loop)
				*weightTrigger = true
				*prevWeightTrigger = true
			} else {
				*weightTrigger = false
			}
		}
	}

	// Process weight triggers for CH1, CH2, and CH3
	processWeightTrigger("weightch1_", "CASE_7_TRIGGER_WEIGHING_CH1", &weightTriggerCh1, &prevWeightTriggerCh1)
	processWeightTrigger("weightch2_", "CASE_7_TRIGGER_WEIGHING_CH2", &weightTriggerCh2, &prevWeightTriggerCh2)
	processWeightTrigger("weightch3_", "CASE_7_TRIGGER_WEIGHING_CH3", &weightTriggerCh3, &prevWeightTriggerCh3)
}
