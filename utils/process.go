package utils

import (
	"fmt"
	"os"
	"patch/model"
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

// _hold_changeName_generic is a generic function to replace device names in the JSON payload with readable keys for a specific case.
func _hold_changeName_generic(jsonPayloads JsonPayloads, key string) map[string]interface{} {
	// Define a mapping of key transformations
	holdkeyTransformations := GetKeyTransformationsFromEnv(key)
	result := make(map[string]interface{})

	// Iterate through key transformations and apply them, deleting old keys during transformation
	for newKey, oldKey := range holdkeyTransformations {

		// Replace old key with new key if the old key exists, delete old key
		if value, oldKeyExists := jsonPayloads[oldKey]; oldKeyExists {
			//if numericValue, isNumeric := value.(float64); isNumeric && numericValue != 0 {
			result[newKey] = value
			// delete(jsonPayloads, oldKey) - consider whether to delete old keys
			//}
		}
	}

	// Apply the specific transformation function
	return result
}

// GetKeyTransformationsFromEnv retrieves key transformations from environment variables based on a given prefix.
func GetKeyTransformationsFromEnv(prefix string) map[string]string {
	keyTransformations := make(map[string]string)

	// Iterate over all environment variables
	for _, env := range os.Environ() {
		// Split the environment variable into key and value
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]

			// Check if the key starts with the specified prefix
			if strings.HasPrefix(key, prefix) {
				// Trim the prefix and add the key-value pair to the map
				key = strings.TrimPrefix(key, prefix)
				keyTransformations[key] = value
			}
		}
	}

	return keyTransformations
}

// ProcessTriggerGeneric is a generic function to process trigger key and return the corresponding processed payload
func ProcessTriggerGeneric(jsonPayloads JsonPayloads, messages []model.Message, loop float64, changeNameFunc func(JsonPayloads) map[string]interface{}) map[string]interface{} {
	startTime := time.Now()
	processMessagesLoop(jsonPayloads, messages, startTime, 1)
	calculateAndStoreInklot(jsonPayloads)
	processedPayload := changeNameFunc(jsonPayloads)

	return processedPayload
}

// MergeNonEmptyMaps merges non-empty maps and returns a new map.
func MergeNonEmptyMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for _, m := range maps {
		if len(m) > 0 {
			for key, value := range m {
				result[key] = value
			}
		}
	}

	return result
}

// reverseString takes an input string and returns a new string with its characters reversed.
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
