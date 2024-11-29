package utils

import (
	"gopatch/model"
	"os"
	"strings"
)

// Helper function to process the trigger for each channel;
// for CASE 4 and CASE 7
func processChannelTrigger(triggerEnvVar, prefix string, jsonPayloads JsonPayloads, messages []model.Message, loop float64) {
	TRIGGER := jsonPayloads[os.Getenv(triggerEnvVar)]
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
func processAndPrintforVacuum(key string, jsonPayloads JsonPayloads, messages []model.Message, loop float64) {
	processedPayloadsMap[key] = ProcessTriggerGeneric(jsonPayloads, messages, loop, func(payload JsonPayloads) map[string]interface{} {
		return _hold_changeName_generic(payload, "CASE_4_VACUUM_")
	})
	//fmt.Println(processedPayloadsMap[key])
}

// Helper function to compares and updates values in a nested map based on the provided keys.
// It updates the map if the new value is larger than the existing one; for CASE 7 only
func CompareAndUpdateNestedMap(parentMap map[string]map[string]interface{}, parentKey string, updateData map[string]interface{}, keysToCheck []string) {
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

// Helper Function, a generic function to replace device names in the JSON payload with readable keys for a specific case.
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
