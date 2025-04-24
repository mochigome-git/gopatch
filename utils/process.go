package utils

import (
	"fmt"
	"gopatch/model"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// processMessagesLoop receives messages within a specified time and updates a JSON payload map.
// If a key is repeated, it overwrites the existing value.
func processMessagesLoop(jsonPayloads *SafeJsonPayloads, messages []model.Message,
	startTime time.Time, loop float64) {

	for {
		for _, message := range messages {
			fieldNameLower := strings.ToLower(message.Address)
			fieldValue := message.Value
			jsonPayloads.Set(fieldNameLower, fieldValue)
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
// ProcessTriggerGenericSpecial is a generic function to process trigger key
// and return the corresponding processed payload
// Using for Case Special
func ProcessTriggerGenericSpecial(jsonPayloads *SafeJsonPayloads, messages []model.Message,
	loop float64, changeNameFunc func(*SafeJsonPayloads) map[string]interface{}) map[string]interface{} {

	startTime := time.Now()
	processMessagesLoop(jsonPayloads, messages, startTime, 1)
	processedPayload := changeNameFunc(jsonPayloads)

	return processedPayload
}

// generateProcessKey creates a unique key for each process based on relevant parameters.
func generateProcessKey(triggerKey string) string {
	// You can concatenate relevant parameters to create a unique key
	return triggerKey /* + other parameters as needed */
}

// clearCacheAndData replaces the existing data map with a new empty one.
func clearCacheAndData(collectedData *SafeJsonPayloads) JsonPayloads {
	// Create a new empty map to replace the existing one.
	return make(JsonPayloads)
}

// ProcessTriggerGeneric is a generic function to process trigger key
// and return the corresponding processed payload
func ProcessTriggerGeneric(jsonPayloads *SafeJsonPayloads, messages []model.Message,
	loop float64, changeNameFunc func(*SafeJsonPayloads) map[string]interface{}) map[string]interface{} {

	startTime := time.Now()
	processMessagesLoop(jsonPayloads, messages, startTime, 1)
	calculateAndStoreInklot(jsonPayloads)
	processedPayload := changeNameFunc(jsonPayloads)

	return processedPayload
}

// Procees to assigning the common logic to a function and then call that function inside each case
// Handle the common logic for case string and float64; for CASE 4
func processAndPrint(key string, jsonPayloads *SafeJsonPayloads, messages []model.Message, loop float64) {
	processedPayloadsMap[key] = ProcessTriggerGeneric(jsonPayloads, messages,
		loop, func(payload *SafeJsonPayloads) map[string]interface{} {

			updatedMap := _hold_changeName_generic(payload, "HOLD_KEY_TRANSOFRMATION_"+key)

			// Define the keys to check
			keysToCheck := []string{"ch3_weighing", "ch1_weighing", "ch2_weighing"}

			// Use the helper function to compare and update the nested map
			CompareAndUpdateNestedMap(processedPayloadsMap, key, updatedMap, keysToCheck)

			return updatedMap
		})
	fmt.Println(processedPayloadsMap[key])
}

// Process to check the time taken from 0 to 1; or CASE 1
func handleTimeDurationTrigger(
	tk TriggerKey,
	jsonPayloads *SafeJsonPayloads,
	messages []model.Message,
	loop float64,
	filter string,
	apiUrl string,
	serviceRoleKey string,
	function string,
) {
	if val, ok := jsonPayloads.Get(tk.triggerKey); ok {
		fmt.Printf("Device name: %s, Payload: %v\n", tk.triggerKey, val)
	} else {
		fmt.Printf("Device name: %s, Payload: <no data>\n", tk.triggerKey)
	}

	if startTime, exists := deviceStartTimeMap[tk.triggerKey]; !exists {
		deviceStartTimeMap[tk.triggerKey] = time.Now()
	} else {
		if trigger, ok := jsonPayloads.GetFloat64(tk.triggerKey); ok && trigger == 0 {
			calculateAndStoreInklot(jsonPayloads)
			changeName(jsonPayloads)
			processMessagesLoop(jsonPayloads, messages, startTime, loop)
		}

		deviceStartTimeMap[tk.triggerKey] = time.Now()
	}
}

// Process for weight triggers (CH1, CH2, CH3); for CASE 7 & CASE 8
func ProcessWeightTriggers(jsonPayloads *SafeJsonPayloads, messages []model.Message, loop float64) {
	var wg sync.WaitGroup

	// A helper function to process each weight trigger concurrently
	processWeightTrigger := func(channel string, triggerKey string, weightTrigger *bool,
		prevWeightTrigger *bool) {

		defer wg.Done()

		triggerValue, ok := jsonPayloads.Get(os.Getenv(triggerKey))
		if !ok {
			log.Printf("Trigger key %s not found", os.Getenv(triggerKey))
			return
		}
		log.Printf("Trigger value for %s: %v", os.Getenv(triggerKey), triggerValue) // Debugging line
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

	// Add three goroutines to the WaitGroup
	wg.Add(3)

	// Run each trigger processing in its own goroutine
	go processWeightTrigger("weightch1_", "CASE_7_TRIGGER_WEIGHING_CH1", &weightTriggerCh1, &prevWeightTriggerCh1)
	go processWeightTrigger("weightch2_", "CASE_7_TRIGGER_WEIGHING_CH2", &weightTriggerCh2, &prevWeightTriggerCh2)
	go processWeightTrigger("weightch3_", "CASE_7_TRIGGER_WEIGHING_CH3", &weightTriggerCh3, &prevWeightTriggerCh3)

	// Wait for all goroutines to finish
	wg.Wait()
}
