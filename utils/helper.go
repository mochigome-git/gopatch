package utils

import (
	"fmt"
	"os"
	"strings"
)

// Helper function to compares and updates values in a nested map based on the provided keys.
// It updates the map if the new value is larger than the existing one; for CASE 7 only
func CompareAndUpdateNestedMap(parentMap map[string]map[string]interface{}, parentKey string,
	updateData map[string]interface{}, keysToCheck []string, prevWeightValue *float64) {

	nestedMap := parentMap[parentKey]
	if nestedMap == nil {
		return
	}

	for _, checkKey := range keysToCheck {
		// Retrieve the existing value from the nested map and check if it's a float64
		// If the existing value is greater than the previous weight value, update it
		existingFloat, okExist := nestedMap[checkKey].(float64)
		if okExist && existingFloat > *prevWeightValue {
			*prevWeightValue = existingFloat
		}

		// Retrieve the new value from the updateData and validate it (must be a non-zero float64)
		newValue, okNew := updateData[checkKey].(float64)
		if !okNew || newValue == 0 {
			continue
		}

		fmt.Println("Comparing:", checkKey, newValue, existingFloat, *prevWeightValue)

		if !okExist {
			continue
		}

		// If the new value is greater than the existing one and greater than or equal to the previous weight
		if newValue > existingFloat && newValue >= *prevWeightValue {
			fmt.Println("Updating value:", checkKey, existingFloat, "->", newValue, "prevWeight:", *prevWeightValue)
			nestedMap[checkKey] = newValue
			*prevWeightValue = newValue
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

type TriggerKey struct {
	TriggerKey string
	CaseKey    string
}

// Helper Function splits a string of trigger keys and case numbers,
// returning a slice of TriggerKey structs.
func ParseTriggerKey(triggerKey string) []TriggerKey {
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
			TriggerKey: triggerKeySlice[i],
			CaseKey:    fmt.Sprint(caseNumber),
		})
	}

	return triggerkeys
}

// Helper Function replaces device names in the JSON payload with readable keys.
func ChangeName(jsonPayloads *SafeJsonPayloads) {
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
func CalculateAndStoreInklot(jsonPayloads *SafeJsonPayloads) {
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
