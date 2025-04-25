package utils

import (
	"fmt"
	"gopatch/model"
	"log"
	"os"
	"strings"
)

// Helper function to process the trigger for each channel;
// for CASE 4 and CASE 7
func processChannelTrigger(triggerEnvVar, prefix string, jsonPayloads *SafeJsonPayloads,
	messages []model.Message, loop float64) {

	TRIGGER, ok := jsonPayloads.Get(os.Getenv(triggerEnvVar))
	if !ok {
		log.Printf("Trigger key %s not found", os.Getenv(triggerEnvVar))
		return
	}
	switch v := TRIGGER.(type) {
	case string:
		if v == "1" {
			processAndPrint(prefix, jsonPayloads, messages, loop)
		}
	case float64:
		if v == 1 {
			processAndPrint(prefix, jsonPayloads, messages, loop)
		}
	}
}

// Helper function for assigning the common logic
// to a function and then call that function inside each case
// Handle the common logic for case if not nil;
// for CASE 4 & CASE 7.
func processAndPrintforVacuum(key string, jsonPayloads *SafeJsonPayloads, messages []model.Message, loop float64) {
	processedPayloadsMap[key] = ProcessTriggerGeneric(jsonPayloads, messages, loop,
		func(payload *SafeJsonPayloads) map[string]interface{} {

			return _hold_changeName_generic(payload, "CASE_4_VACUUM_")
		})
	//fmt.Println(processedPayloadsMap[key])
}

// Helper function to compares and updates values in a nested map based on the provided keys.
// It updates the map if the new value is larger than the existing one; for CASE 7 only
func CompareAndUpdateNestedMap(parentMap map[string]map[string]interface{}, parentKey string,
	updateData map[string]interface{}, keysToCheck []string) {

	// Access the nested map
	nestedMap := parentMap[parentKey]
	if nestedMap == nil {
		return // Do nothing if no nested map exists for the parentKey
	}

	// Iterate over the keys to compare and update
	for _, checkKey := range keysToCheck {
		// Extract the new value from the update data
		newValue, okNew := updateData[checkKey].(float64)
		if !okNew {
			continue // Skip if the value is not a float64
		}

		// Check if the key exists in the nested map
		existingValue, exists := nestedMap[checkKey]
		if exists {
			// Safely type assert the existing value
			existingFloat, okExisting := existingValue.(float64)
			if okExisting && newValue > existingFloat {
				nestedMap[checkKey] = newValue // Update if the new value is larger
			}
		} else {
			nestedMap[checkKey] = newValue // Add key if it doesn't exist
		}
	}
}

// Helper Function takes an input string and returns a new string with its characters reversed.
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Helper Function, a generic function to replace device names in the JSON payload
// with readable keys for a specific case.
func _hold_changeName_generic(jsonPayloads *SafeJsonPayloads, key string) map[string]interface{} {
	// Define a mapping of key transformations
	holdkeyTransformations := GetKeyTransformationsFromEnv(key)
	result := make(map[string]interface{})

	// Iterate through key transformations and apply them, deleting old keys during transformation
	for newKey, oldKey := range holdkeyTransformations {

		// Replace old key with new key if the old key exists, delete old key
		if value, oldKeyExists := jsonPayloads.Get(oldKey); oldKeyExists {
			//if numericValue, isNumeric := value.(float64); isNumeric && numericValue != 0 {
			result[newKey] = value
			// delete(jsonPayloads, oldKey) - consider whether to delete old keys
			//}
		}
	}

	// Apply the specific transformation function
	return result
}

// Helper Function retrieves key transformations from environment variables based on a given prefix.
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

// Helper Function to merges non-empty maps and returns a new map.
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

// Helper Function splits a string of trigger keys and case numbers,
// returning a slice of TriggerKey structs.
func parseTriggerKey(triggerKey string) []TriggerKey {
	triggerKeySlice := strings.Split(triggerKey, ",")
	var triggerkeys []TriggerKey

	// Check if the number of items in the triggerKeySlice is even
	if len(triggerKeySlice)%2 != 0 {
		fmt.Println("Warning: Malformed triggerKey input. Ensure it contains pairs of trigger and case numbers.")
		return triggerkeys // Return empty slice if the input is malformed
	}

	for i := 0; i < len(triggerKeySlice); i += 2 {
		caseNumber := triggerKeySlice[i+1]

		triggerkeys = append(triggerkeys, TriggerKey{
			triggerKey: triggerKeySlice[i],
			caseKey:    fmt.Sprint(caseNumber),
		})
	}

	return triggerkeys
}

// Helper Function replaces device names in the JSON payload with readable keys.
func changeName(jsonPayloads *SafeJsonPayloads) {
	// Define a mapping of key transformations
	keyTransformations := GetKeyTransformationsFromEnv("KEY_TRANSFORMATION_")

	// Repeat channel 1's sequence count (PLC's device name) for channel 2 and channel 3.
	if d760, exists := jsonPayloads.Get("d760"); exists {
		jsonPayloads.Set("ch1_sequence", d760)
		jsonPayloads.Set("ch2_sequence", d760)
		jsonPayloads.Set("ch3_sequence", d760)
	}
	// Remove Channel 1, 2, 3 keys after processing
	jsonPayloads.Delete("d160")
	jsonPayloads.Delete("d460")
	jsonPayloads.Delete("d760")

	// Iterate through key transformations and apply them, deleting old keys during transformation
	for newKey, oldKey := range keyTransformations {
		// Replace old key with new key if the old key exists, delete old key
		if value, exists := jsonPayloads.Get(oldKey); exists {
			jsonPayloads.Set(newKey, value)
			jsonPayloads.Delete(oldKey)
		}
	}
}

// Helper Function to computes and stores an 'ink_lot' value based on specific keys in the JSON payload.
func calculateAndStoreInklot(jsonPayloads *SafeJsonPayloads) {
	d171Value, d171Exists := jsonPayloads.GetString("d171")
	d172Value, d172Exists := jsonPayloads.GetString("d172")
	d173Value, d173Exists := jsonPayloads.GetString("d173")

	var inklotValue string
	if d171Exists && d172Exists && d173Exists {
		// Concatenate reversed strings of d171, d172, and d173
		inklotValue = reverseString(d171Value) + reverseString(d172Value) + reverseString(d173Value)
	}
	jsonPayloads.Set("ink_lot", inklotValue)

	// Remove "d171", "d172", and "d173" keys from the map
	jsonPayloads.Delete("d171")
	jsonPayloads.Delete("d172")
	jsonPayloads.Delete("d173")
}
