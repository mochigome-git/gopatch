package utils

import (
	"fmt"
	"os"
	"patch/model"
	"strings"
	"time"
)

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
	keyTransformations := getKeyTransformationsFromEnv()

	// Repeat channel 1's sequence count (PLC's device name) for channel 2 and channel 3.
	jsonPayloads["ch1_sequence"] = jsonPayloads["d760"]
	jsonPayloads["ch2_sequence"] = jsonPayloads["d760"]
	jsonPayloads["ch3_sequence"] = jsonPayloads["d760"]
	// Remove Channel 1,2,3 key after process
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

// Function to get key transformations from environment variables
func getKeyTransformationsFromEnv() map[string]string {
	keyTransformations := make(map[string]string)

	// Read environment variables and populate key transformations
	envPrefix := "KEY_TRANSFORMATION_"
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, envPrefix) {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimPrefix(parts[0], envPrefix)
				value := parts[1]
				keyTransformations[key] = value
			}
		}
	}

	return keyTransformations
}

// Input and returns a new string with its characters reversed
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
