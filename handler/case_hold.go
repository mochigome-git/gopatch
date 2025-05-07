package handler

import (
	"encoding/json"
	"fmt"
	"gopatch/config"
	"gopatch/model"
	"gopatch/patch"
	"gopatch/utils"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CASE 3, Trigger; handling the device when triggered and hold for 4second to collect data to patch.
func handleTriggerCase(tk utils.TriggerKey, jsonPayloads *utils.SafeJsonPayloads, messages []model.Message,
	cfg config.AppConfig) {

	if value, ok := jsonPayloads.GetFloat64(tk.TriggerKey); ok && value == 1 {

		startTime := time.Now()
		processMessagesLoop(jsonPayloads, messages, startTime, cfg.Loop)

		if _filter, ok := jsonPayloads.GetFloat64(cfg.Filter); ok && _filter != 0 {
			utils.CalculateAndStoreInklot(jsonPayloads)
			utils.ChangeName(jsonPayloads)

			jsonData, err := json.Marshal(jsonPayloads)
			if err != nil {
				fmt.Println("Error marshaling JSON:", err)
				return
			}

			_, err = patch.SendPatchRequest(cfg.APIUrl, cfg.ServiceRoleKey, jsonData, cfg.Function)
			if err != nil {
				panic(err)
			}

			elapsedTime := time.Since(startTime)
			prettyPrintJSONWithTime(jsonPayloads, elapsedTime)
		}
	}
}

// CASE 4, Hold; hold the data and wait until patch trigger
func handleHoldCase(session *Session, jsonPayloads *utils.SafeJsonPayloads, messages []model.Message,
	cfg config.AppConfig, checkAccumulateRate AccumCheckFunc) {

	if checkAccumulateRate() {
		return
	}

	// handle the different types (string and float64) of CH1_TRIGGER.
	// And Store the Filling parameter of CH1 when the trigger is true.
	processChannelTrigger("CASE_4_TRIGGER_CH1", "ch1_", jsonPayloads, messages, session)
	processChannelTrigger("CASE_4_TRIGGER_CH2", "ch2_", jsonPayloads, messages, session)
	processChannelTrigger("CASE_4_TRIGGER_CH3", "ch3_", jsonPayloads, messages, session)

	VACUUM_TRIGGER, _ := jsonPayloads.Get(os.Getenv("CASE_4_VACUUM_reach_20pa"))
	if VACUUM_TRIGGER != nil {
		processAndPrintforVacuum("vacuum", jsonPayloads, messages, session)
	}

	if sealing, ok := jsonPayloads.GetFloat64(os.Getenv("CASE_4_SEALING")); ok {
		if sealing == 1 {
			// Use the function with the condition
			//processAndPrintforVacuum("vacuum", jsonPayloads, messages, loop)
			value, exists := jsonPayloads.Get("vacuum")
			if exists {
				fmt.Println(value)
			} else {
				fmt.Println("Key not found")
			}

			// After the goroutine has finished, set prevSealing = sealing
			session.PrevSealing = sealing
		} else if sealing == 0 && session.PrevSealing == 1 {
			// Use the function to merge payloads
			data := mergeNonEmptyMaps(
				session.ProcessedPayloadsMap["ch1_"],
				session.ProcessedPayloadsMap["ch2_"],
				session.ProcessedPayloadsMap["ch3_"],
				session.ProcessedPayloadsMap["vacuum"],
			)

			startTime := time.Now()
			jsonData, err := json.Marshal(data)
			if err != nil {
				fmt.Println("Error marshaling JSON:", err)
				return
			}

			_, err = patch.SendPatchRequest(cfg.APIUrl, cfg.ServiceRoleKey, jsonData, cfg.Function)
			if err != nil {
				panic(err)
			}

			elapsedTime := time.Since(startTime)
			prettyPrintJSONWithTime(data, elapsedTime)
			// Update the previous state of sealing
			session.PrevSealing = sealing
		}
	}
}

// CASE 6, HoldFilling; handling the device when triggered and hold for 4second to collect data to patch.
func handleHoldFillingCase(session *Session, jsonPayloads *utils.SafeJsonPayloads, messages []model.Message,
	cfg config.AppConfig, rMsgJSONChan <-chan string) {

	triggerChannels := []string{"ch1", "ch2", "ch3"}

	for _, channel := range triggerChannels {
		// Retrieve NUMBERofSTATE from environment variable and convert to float64
		NUMBERofSTATEStr := os.Getenv("CASE_6_TRIGGER_NUMBERofSTATE")
		NUMBERofSTATE, err := strconv.ParseFloat(NUMBERofSTATEStr, 64)
		if err != nil {
			fmt.Println("Error parsing NUMBERofSTATE:", err)
			continue
		}

		// Retrieve trigger value from JSON payload
		triggerValue, ok := jsonPayloads.GetFloat64(os.Getenv("CASE_6_TRIGGER_" + channel))
		if ok && triggerValue == NUMBERofSTATE {
			session.Mutex.Lock()
			defer session.Mutex.Unlock()

			session.ProcessedPayloadsMap[channel][channel+"_fill"] = 1
			session.IsProcessing = true
		}
	}

	// Check if all channels are successful and processing is active
	// Use a flag to track if all channels have success = 0
	ch1, ok1 := jsonPayloads.GetFloat64(os.Getenv("CASE_6_TRIGGER_ch1"))
	ch2, ok2 := jsonPayloads.GetFloat64(os.Getenv("CASE_6_TRIGGER_ch2"))
	ch3, ok3 := jsonPayloads.GetFloat64(os.Getenv("CASE_6_TRIGGER_ch3"))
	session.AllSuccessZero = ok1 && ok2 && ok3 && ch1 == 0 && ch2 == 0 && ch3 == 0

	if session.AllSuccessZero && session.IsProcessing {
		prevDo := false

		session.ProcessedPayloadsMap["do"] = processTriggerGeneric(jsonPayloads, messages, func(payload *utils.SafeJsonPayloads) map[string]interface{} {
			prevDo = true
			return _hold_changeName_generic(payload, "CASE_6_DO_")
		})

		processWeightTriggers(session, jsonPayloads, messages)
		if shouldPatch("case8", prevDo, session) {
			keys := []string{
				"ch1", "ch2", "ch3", "do",
			}
			processPatch(session, keys, cfg, func() { prevDo = false }, rMsgJSONChan)
		}

	}

}

// CASE 7, Weight; hold the data and wait until weighing scale trigger to collect data to patch.
func handleWeight(session *Session, jsonPayloads *utils.SafeJsonPayloads, messages []model.Message,
	cfg config.AppConfig, chance bool, checkAccumulateRate AccumCheckFunc, rMsgJSONChan <-chan string) {

	if checkAccumulateRate() {
		chance = true
	}

	// Process to handling counter when ch1 started
	processChannelTrigger("CASE_4_TRIGGER_CH1", "counterch_", jsonPayloads, messages, session)

	// Process triggers for each channel
	// Handle different types (string and float64) of CH1_TRIGGER, CH2_TRIGGER, CH3_TRIGGER.
	for _, channel := range []string{"ch1_", "ch2_", "ch3_"} {
		processChannelTrigger("CASE_4_TRIGGER_"+strings.ToUpper(channel[:3]), channel, jsonPayloads, messages, session)
	}

	// Process Vacuum Trigger
	vacuumTrigger, _ := jsonPayloads.Get(os.Getenv("CASE_4_VACUUM_reach_20pa"))
	if vacuumTrigger != nil {
		processAndPrintforVacuum("vacuum", jsonPayloads, messages, session)
	}

	// Process CH1, CH2, CH3 Weight Triggers
	// Check if all weight triggers (CH1, CH2, CH3) are inactive, but were previously active
	processWeightTriggers(session, jsonPayloads, messages)
	if shouldPatch("case7", chance, session) {
		keys := []string{
			"ch1_", "ch2_", "ch3_", "vacuum", "weightch1_", "weightch2_", "weightch3_", "counterch_",
		}
		processPatch(session, keys, cfg, func() { session.IsProcessing = false }, rMsgJSONChan)
	}

}

// CASE 8, HoldFillingWeight; hold the data and wait until weighing scale trigger to collect data to patch.
func handleHoldFillingWeightCase(session *Session, jsonPayloads *utils.SafeJsonPayloads, messages []model.Message,
	cfg config.AppConfig, rMsgJSONChan <-chan string) {

	for _, channel := range []string{"ch1", "ch2", "ch3"} {
		// Retrieve NUMBERofSTATE from environment variable and convert to float64
		NUMBERofSTATEStr := os.Getenv("CASE_6_TRIGGER_NUMBERofSTATE")
		NUMBERofSTATE, err := strconv.ParseFloat(NUMBERofSTATEStr, 64)
		if err != nil {
			fmt.Println("Error parsing NUMBERofSTATE:", err)
			continue
		}

		triggerValue, ok := jsonPayloads.GetFloat64(os.Getenv("CASE_6_TRIGGER_" + channel))
		if ok && triggerValue == NUMBERofSTATE {
			session.Mutex.Lock()

			if session.ProcessedPayloadsMap[channel] == nil {
				session.ProcessedPayloadsMap[channel] = make(map[string]interface{})
			}
			session.ProcessedPayloadsMap[channel][channel+"_fill"] = 1
			session.Mutex.Unlock()
			session.IsProcessing = true
		}
	}

	// Check if all channels are successful and processing is active
	// Use a flag to track if all channels have success = 0
	ch1, ok1 := jsonPayloads.GetFloat64(os.Getenv("CASE_6_TRIGGER_ch1"))
	ch2, ok2 := jsonPayloads.GetFloat64(os.Getenv("CASE_6_TRIGGER_ch2"))
	ch3, ok3 := jsonPayloads.GetFloat64(os.Getenv("CASE_6_TRIGGER_ch3"))
	session.AllSuccessZero = ok1 && ok2 && ok3 && ch1 == 0 && ch2 == 0 && ch3 == 0

	if session.AllSuccessZero && session.IsProcessing {
		prevDo := false

		session.ProcessedPayloadsMap["do"] = processTriggerGeneric(jsonPayloads, messages, func(payload *utils.SafeJsonPayloads) map[string]interface{} {
			prevDo = true
			return _hold_changeName_generic(payload, "CASE_6_DO_")
		})

		processWeightTriggers(session, jsonPayloads, messages)
		if shouldPatch("case8", prevDo, session) {
			keys := []string{
				"ch1", "ch2", "ch3", "do", "weightch1_", "weightch2_", "weightch3_",
			}
			processPatch(session, keys, cfg, func() { prevDo = false }, rMsgJSONChan)
		}

	}

}

func shouldPatch(caseID string, ready bool, session *Session) bool {
	if caseID == "case7" || caseID == "case8" {
		// Case 7 & Case 8: Wait for all channels to deactivate after being active
		return !session.WeightTriggerCh1 && !session.WeightTriggerCh2 && !session.WeightTriggerCh3 &&
			session.PrevWeightTriggerCh1 && session.PrevWeightTriggerCh2 && session.PrevWeightTriggerCh3 && ready
	}
	// Default: don't patch
	return false
}

// Reset previous triggers to avoid reprocessing
func resetWeightTriggers(session *Session) {
	session.AllSuccessZero = false
	session.PrevWeightTriggerCh1 = false
	session.PrevWeightTriggerCh2 = false
	session.PrevWeightTriggerCh3 = false
	*session.PrevWeightValueCh1 = 0
	*session.PrevWeightValueCh2 = 0
	*session.PrevWeightValueCh3 = 0
}

func processPatch(session *Session, keys []string, cfg config.AppConfig, after func(), rMsgJSONChan <-chan string) {
	fmt.Println("All weight triggers are now inactive. Processing the patch.")

	parts := []map[string]interface{}{}
	for _, key := range keys {
		parts = append(parts, session.ProcessedPayloadsMap[key])
	}
	data := mergeNonEmptyMaps(parts...)

	startTime := time.Now()
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	if _, err := patch.SendPatchRequest(cfg.APIUrl, cfg.ServiceRoleKey, jsonData, cfg.Function); err != nil {
		log.Fatal("Error sending patch request:", err)
	}

	prettyPrintJSONWithTime(data, time.Since(startTime))

	for key := range session.ProcessedPayloadsMap {
		delete(session.ProcessedPayloadsMap, key)
	}

	// Always reset weight triggers
	resetWeightTriggers(session)

	// Call the extra cleanup if provided
	if after != nil {
		after()
	}

	drainChannel(rMsgJSONChan)
}

// Helper Function to merges non-empty maps and returns a new map.
func mergeNonEmptyMaps(maps ...map[string]interface{}) map[string]interface{} {
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

// Helper function to compares and updates values in a nested map based on the provided keys.
// It updates the map if the new value is larger than the existing one; for CASE 7 only
func compareAndUpdateNestedMap(parentMap map[string]map[string]interface{}, parentKey string,
	updateData map[string]interface{}, keysToCheck []string, prevWeightValue *float64) {

	nestedMap := parentMap[parentKey]
	if nestedMap == nil {
		return
	}

	for _, checkKey := range keysToCheck {
		// Retrieve the existing value from the nested map and check if it's a float64
		// If the existing value is greater than the previous weight value, update it
		existingFloat, okExist := nestedMap[checkKey].(float64)

		// Retrieve the new value from the updateData and validate it (must be a non-zero float64)
		newValue, okNew := updateData[checkKey].(float64)
		if !okNew || newValue == 0 {
			continue
		}

		// fmt.Println("Comparing:", checkKey, newValue, existingFloat, *prevWeightValue)

		if !okExist {
			nestedMap[checkKey] = newValue
			*prevWeightValue = newValue
			continue
		}

		// If the new value is greater than the existing one and greater than or equal to the previous weight
		if newValue > existingFloat && newValue >= *prevWeightValue {
			// fmt.Println("Updating value:", checkKey, existingFloat, "->", newValue, "prevWeight:", *prevWeightValue)
			nestedMap[checkKey] = newValue
			*prevWeightValue = newValue
		} else if newValue >= *prevWeightValue {
			// update prevWeightValue to avoid being stuck at 0
			*prevWeightValue = newValue
		}
	}
}

// Procees to assigning the common logic to a function and then call that function inside each case
// Handle the common logic for case string and float64;
func processAndPrint(session *Session, key string, jsonPayloads *utils.SafeJsonPayloads, messages []model.Message, prevWeightValue *float64) {
	session.Mutex.Lock()
	defer session.Mutex.Unlock()

	processed := processTriggerGeneric(jsonPayloads, messages, func(payload *utils.SafeJsonPayloads) map[string]interface{} {
		updatedMap := _hold_changeName_generic(payload, "HOLD_KEY_TRANSOFRMATION_"+key)

		keysToCheck := []string{"ch3_weighing", "ch1_weighing", "ch2_weighing"}

		// // Check if any of the keys has a value of 0 — if so, drop the payload
		// for _, k := range keysToCheck {
		// 	if val, ok := updatedMap[k]; ok {
		// 		if num, ok := val.(float64); ok && num == 0 {
		// 			// Drop the payload by returning nil
		// 			return nil
		// 		}
		// 	}
		// }

		compareAndUpdateNestedMap(session.ProcessedPayloadsMap, key, updatedMap, keysToCheck, prevWeightValue)

		return updatedMap
	})

	// fmt.Println(session.ProcessedPayloadsMap)
	if processed != nil {
		session.ProcessedPayloadsMap[key] = processed
	}
}

// Helper Function, a generic function to replace device names in the JSON payload
// with readable keys for a specific case.
func _hold_changeName_generic(jsonPayloads *utils.SafeJsonPayloads, key string) map[string]interface{} {
	// Define a mapping of key transformations
	holdkeyTransformations := utils.GetKeyTransformationsFromEnv(key)
	result := make(map[string]interface{})

	// Iterate through key transformations and apply them, deleting old keys during transformation
	for newKey, oldKey := range holdkeyTransformations {

		// Replace old key with new key if the old key exists, delete old key
		if value, oldKeyExists := jsonPayloads.Get(oldKey); oldKeyExists {
			if numericValue, isNumeric := value.(float64); isNumeric && numericValue != 0 {
				result[newKey] = value
				// delete(jsonPayloads, oldKey) - consider whether to delete old keys
			}
		}
	}

	// Apply the specific transformation function
	return result
}

// ProcessTriggerGeneric is a generic function to process trigger key
// and return the corresponding processed payload
func processTriggerGeneric(jsonPayloads *utils.SafeJsonPayloads, messages []model.Message,
	changeNameFunc func(*utils.SafeJsonPayloads) map[string]interface{}) map[string]interface{} {

	//startTime := time.Now()
	//processMessagesLoop(jsonPayloads, messages, startTime, 1)

	processMessagesOnce(jsonPayloads, messages)
	utils.CalculateAndStoreInklot(jsonPayloads)
	processedPayload := changeNameFunc(jsonPayloads)

	return processedPayload
}

// Helper function to process the trigger for each channel;
// for CASE 4 and CASE 7
func processChannelTrigger(triggerEnvVar, prefix string, jsonPayloads *utils.SafeJsonPayloads,
	messages []model.Message, session *Session) {

	TRIGGER, ok := jsonPayloads.Get(os.Getenv(triggerEnvVar))
	if !ok {
		fmt.Printf("Trigger key %s not found", os.Getenv(triggerEnvVar))
		return
	}
	switch v := TRIGGER.(type) {
	case string:
		if v == "1" {
			processAndPrint(session, prefix, jsonPayloads, messages, nil)
		}
	case float64:
		if v == 1 {
			processAndPrint(session, prefix, jsonPayloads, messages, nil)
		}
	}
}

// Helper function for assigning the common logic
// to a function and then call that function inside each case
// Handle the common logic for case if not nil;
// for CASE 4 & CASE 7.
func processAndPrintforVacuum(key string, jsonPayloads *utils.SafeJsonPayloads, messages []model.Message, session *Session) {
	session.ProcessedPayloadsMap[key] = processTriggerGeneric(jsonPayloads, messages,
		func(payload *utils.SafeJsonPayloads) map[string]interface{} {

			return _hold_changeName_generic(payload, "CASE_4_VACUUM_")
		})
}

// Process for weight triggers (CH1, CH2, CH3); for CASE 7 & CASE 8
func processWeightTriggers(session *Session, jsonPayloads *utils.SafeJsonPayloads, messages []model.Message) {
	var wg sync.WaitGroup

	// A helper function to process each weight trigger concurrently
	processWeightTrigger := func(channel string, triggerKey string, weightTrigger *bool,
		prevWeightTrigger *bool, prevWeightValue *float64) {

		defer wg.Done()

		triggerValue, ok := jsonPayloads.GetDC(os.Getenv(triggerKey))
		if !ok {
			fmt.Printf("Trigger key %s not found\n", os.Getenv(triggerKey))
			return
		}

		isTriggered := false
		switch v := triggerValue.(type) {
		case string:
			isTriggered = (v == "1")
		case float64:
			isTriggered = (v == 1)
		default:
			fmt.Printf("Unexpected type for trigger value: %T\n", v)
			return
		}

		if isTriggered {
			processAndPrint(session, channel, jsonPayloads, messages, prevWeightValue)
			*weightTrigger = true
			*prevWeightTrigger = true
		} else {
			*weightTrigger = false
		}
	}

	// Add three goroutines to the WaitGroup
	wg.Add(3)

	// Run each trigger processing in its own goroutine
	go processWeightTrigger("weightch1_", "CASE_7_TRIGGER_WEIGHING_CH1", &session.WeightTriggerCh1, &session.PrevWeightTriggerCh1, session.PrevWeightValueCh1)
	go processWeightTrigger("weightch2_", "CASE_7_TRIGGER_WEIGHING_CH2", &session.WeightTriggerCh2, &session.PrevWeightTriggerCh2, session.PrevWeightValueCh2)
	go processWeightTrigger("weightch3_", "CASE_7_TRIGGER_WEIGHING_CH3", &session.WeightTriggerCh3, &session.PrevWeightTriggerCh3, session.PrevWeightValueCh3)

	// Wait for all goroutines to finish
	wg.Wait()
}
