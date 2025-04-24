package utils

import (
	"encoding/json"
	"fmt"
	"gopatch/model"
	"os"
	"strconv"
	"strings"
	"time"
)

func handleTrigger(
	jsonPayloads *SafeJsonPayloads,
	messages []model.Message,
	triggerKey string,
	loop float64,
	filter string,
	apiUrl string,
	serviceRoleKey string,
	function string,
) {
	triggerKeys := parseTriggerKey(triggerKey)

	for _, tk := range triggerKeys {
		switch {
		case strings.Contains(tk.caseKey, "time.duration"):
			handleTimeDurationCase(tk, jsonPayloads, messages, loop, filter, apiUrl, serviceRoleKey, function)

		case tk.caseKey == "standard":
			handleStandardCase(tk, jsonPayloads, messages, loop, filter, apiUrl, serviceRoleKey, function)

		case tk.caseKey == "trigger":
			handleTriggerCase(tk, jsonPayloads, messages, loop, filter, apiUrl, serviceRoleKey, function)

		case tk.caseKey == "hold":
			if accum_rate, exists := jsonPayloads.GetFloat64(os.Getenv("CASE_4_AVOID_0")); exists && accum_rate == 0 {
				// Skip further processing if accum_rate is 0
				return
			}
			handleHoldCase(jsonPayloads, messages, loop, apiUrl, serviceRoleKey, function)

		case tk.caseKey == "special":
			handleSpecialCase(tk, jsonPayloads, messages, loop, apiUrl, serviceRoleKey, function)

		case tk.caseKey == "holdfilling":
			handleHoldFillingCase(jsonPayloads, messages, loop, apiUrl, serviceRoleKey, function)

		case tk.caseKey == "holdfillingweight":
			handleHoldFillingWeightCase(jsonPayloads, messages, loop, apiUrl, serviceRoleKey, function)

		case tk.caseKey == "weight":
			if accum_rate, exists := jsonPayloads.GetFloat64(os.Getenv("CASE_4_AVOID_0")); exists && accum_rate == 0 {
				// Skip further processing if accum_rate is 0
				chance = true
			}
			handleWeight(jsonPayloads, messages, loop, apiUrl, serviceRoleKey, function, chance)

		}
	}
}

// CASE 1, time.Duration; handling the process of time taken from 0 to 1, and record the total time duration
func handleTimeDurationCase(tk TriggerKey, jsonPayloads *SafeJsonPayloads, messages []model.Message,
	loop float64, filter string, apiUrl string, serviceRoleKey string, function string) {

	processKey := generateProcessKey(tk.triggerKey)

	if tk.triggerKey != processPrevTriggerKeyMap[processKey] {
		processPrevTriggerKeyMap[processKey] = tk.triggerKey

		if trigger, ok := jsonPayloads.GetFloat64(tk.triggerKey); ok && trigger != 0 {
			handleTimeDurationTrigger(tk, jsonPayloads, messages, loop, filter, apiUrl, serviceRoleKey, function)
		}
	}
}

// CASE 2, Standard; handling a devices value and patch it, when the trigger is different with previous key
func handleStandardCase(tk TriggerKey, jsonPayloads *SafeJsonPayloads, messages []model.Message,
	loop float64, filter string, apiUrl string, serviceRoleKey string, function string) {

	processKey := generateProcessKey(tk.triggerKey)

	if tk.triggerKey != processPrevTriggerKeyMap[processKey] {
		processPrevTriggerKeyMap[processKey] = tk.triggerKey

		if trigger, ok := jsonPayloads.GetFloat64(tk.triggerKey); ok && trigger != 0 {
			var startTime time.Time
			processMessagesLoop(jsonPayloads, messages, startTime, loop)

			calculateAndStoreInklot(jsonPayloads)
			changeName(jsonPayloads)

			if trigger, ok := jsonPayloads.GetFloat64(tk.triggerKey); ok && trigger == 0 {
				fmt.Println("Case 1")
				fmt.Println(jsonPayloads)

				jsonData, err := json.Marshal(jsonPayloads)
				if err != nil {
					fmt.Println("Error marshaling JSON:", err)
					return
				}

				_, err = sendPatchRequest(apiUrl, serviceRoleKey, jsonData, function)
				if err != nil {
					panic(err)
				}

				elapsedTime := time.Since(startTime)
				prettyPrintJSONWithTime(jsonPayloads, elapsedTime)
			}
		}
	}
}

// CASE 3, Trigger; handling the device when triggered and hold for 4second to collect data to patch.
func handleTriggerCase(tk TriggerKey, jsonPayloads *SafeJsonPayloads, messages []model.Message,
	loop float64, filter string, apiUrl string, serviceRoleKey string, function string) {

	if value, ok := jsonPayloads.GetFloat64(tk.triggerKey); ok && value == 1 {

		startTime := time.Now()
		processMessagesLoop(jsonPayloads, messages, startTime, loop)

		if _filter, ok := jsonPayloads.GetFloat64(filter); ok && _filter != 0 {
			calculateAndStoreInklot(jsonPayloads)
			changeName(jsonPayloads)

			jsonData, err := json.Marshal(jsonPayloads)
			if err != nil {
				fmt.Println("Error marshaling JSON:", err)
				return
			}

			_, err = sendPatchRequest(apiUrl, serviceRoleKey, jsonData, function)
			if err != nil {
				panic(err)
			}

			elapsedTime := time.Since(startTime)
			prettyPrintJSONWithTime(jsonPayloads, elapsedTime)
		}
	}
}

// CASE 4, Hold; hold the data and wait until patch trigger
func handleHoldCase(jsonPayloads *SafeJsonPayloads, messages []model.Message, loop float64,
	apiUrl string, serviceRoleKey string, function string) {

	// handle the different types (string and float64) of CH1_TRIGGER.
	// And Store the Filling parameter of CH1 when the trigger is true.
	processChannelTrigger("CASE_4_TRIGGER_CH1", "ch1_", jsonPayloads, messages, loop)
	processChannelTrigger("CASE_4_TRIGGER_CH2", "ch2_", jsonPayloads, messages, loop)
	processChannelTrigger("CASE_4_TRIGGER_CH3", "ch3_", jsonPayloads, messages, loop)

	VACUUM_TRIGGER, _ := jsonPayloads.Get(os.Getenv("CASE_4_VACUUM_reach_20pa"))
	if VACUUM_TRIGGER != nil {
		processAndPrintforVacuum("vacuum", jsonPayloads, messages, loop)
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
			prevSealing = sealing
		} else if sealing == 0 && prevSealing == 1 {
			// Use the function to merge payloads
			data := MergeNonEmptyMaps(
				processedPayloadsMap["ch1_"],
				processedPayloadsMap["ch2_"],
				processedPayloadsMap["ch3_"],
				processedPayloadsMap["vacuum"],
			)

			startTime := time.Now()
			jsonData, err := json.Marshal(data)
			if err != nil {
				fmt.Println("Error marshaling JSON:", err)
				return
			}

			_, err = sendPatchRequest(apiUrl, serviceRoleKey, jsonData, function)
			if err != nil {
				panic(err)
			}

			elapsedTime := time.Since(startTime)
			prettyPrintJSONWithTime(data, elapsedTime)
			// Update the previous state of sealing
			prevSealing = sealing
		}
	}
}

// CASE 5, Special; handling a device's highest value and average value and patch it, when the trigger is 1
func handleSpecialCase(tk TriggerKey, jsonPayloads *SafeJsonPayloads, messages []model.Message,
	loop float64, apiUrl string, serviceRoleKey string, function string) {
	// Assuming these variables need to be declared and initialized
	var startTime time.Time

	if trigger, ok := jsonPayloads.GetFloat64(tk.triggerKey); ok {

		if trigger == 1 {
			isProcessing = true
			// Assuming processedPayloadsMap is a map[string]map[string]interface{}
			if processedPayloadsMap["degas"]["pica1"] == nil {
				processedPayloadsMap["degas"]["pica1"] = make([]float64, 0)
			}

			result := ProcessTriggerGenericSpecial(jsonPayloads, messages, trigger, func(payload *SafeJsonPayloads) map[string]interface{} {
				return _hold_changeName_generic(payload, "CASE_5_DEGAS_")
			})

			// Assuming pica1 is a float64 value in the result map
			if pica1, ok := result["pica1"].(float64); ok {
				processedPayloadsMap["degas"]["pica1"] = append(processedPayloadsMap["degas"]["pica1"].([]float64), pica1)
			}

			//fmt.Println(processedPayloadsMap["degas"]["pica1"])
		}

		if trigger == 0 && isProcessing {
			isProcessing = false

			pica1Values, ok := processedPayloadsMap["degas"]["pica1"].([]float64)

			if ok && len(pica1Values) > 0 { // Check if there are values in the slice

				// Calculate max
				max := pica1Values[0]
				for _, value := range pica1Values {
					if value > max {
						max = value
					}
				}
				processedPayloadsMap["degas"]["pica1_max"] = max

				// Calculate average
				var sum float64
				for _, value := range pica1Values {
					sum += value
				}
				average := sum / float64(len(pica1Values))
				processedPayloadsMap["degas"]["pica1_average"] = average
			} else {
				// Handle the case where there are no values in the pica1Values slice
				fmt.Println("No values found for pica1.")
			}

			// Clear degas values
			delete(processedPayloadsMap["degas"], "pica1")

			// Convert processedPayloadsMap["degas"] to JSON, patch to API, print, etc.
			jsonData, err := json.Marshal(processedPayloadsMap["degas"])
			if err != nil {
				fmt.Println("Error marshaling JSON:", err)
				return
			}

			_, err = sendPatchRequest(apiUrl, serviceRoleKey, jsonData, function)
			if err != nil {
				panic(err)
			}

			elapsedTime := time.Since(startTime)
			prettyPrintJSONWithTime(processedPayloadsMap["degas"], elapsedTime)
			processedPayloadsMap["degas"] = make(map[string]interface{})
		}
	}
}

// CASE 6, HoldFilling; handling the device when triggered and hold for 4second to collect data to patch.
func handleHoldFillingCase(jsonPayloads *SafeJsonPayloads, messages []model.Message,
	loop float64, apiUrl string, serviceRoleKey string, function string) {

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
			processedPayloadsMap[channel][channel+"_fill"] = 1
			isProcessing = true
		}
	}

	ch1Success, ok1 := jsonPayloads.GetFloat64(os.Getenv("CASE_6_TRIGGER_ch1"))
	ch2Success, ok2 := jsonPayloads.GetFloat64(os.Getenv("CASE_6_TRIGGER_ch2"))
	ch3Success, ok3 := jsonPayloads.GetFloat64(os.Getenv("CASE_6_TRIGGER_ch3"))

	if ok1 && ok2 && ok3 && isProcessing && ch1Success == 0 && ch2Success == 0 && ch3Success == 0 {
		prevDo := false

		processedPayloadsMap["do"] = ProcessTriggerGeneric(jsonPayloads, messages, loop, func(payload *SafeJsonPayloads) map[string]interface{} {
			prevDo = true
			return _hold_changeName_generic(payload, "CASE_6_DO_")
		})

		if prevDo {
			data := MergeNonEmptyMaps(
				processedPayloadsMap["ch1"],
				processedPayloadsMap["ch2"],
				processedPayloadsMap["ch3"],
				processedPayloadsMap["do"],
			)

			startTime := time.Now()
			jsonData, err := json.Marshal(data)
			if err != nil {
				fmt.Println("Error marshaling JSON:", err)
				return
			}

			_, err = sendPatchRequest(apiUrl, serviceRoleKey, jsonData, function)
			if err != nil {
				panic(err)
			}

			elapsedTime := time.Since(startTime)
			prettyPrintJSONWithTime(data, elapsedTime)
			// Update the previous state of sealing
			isProcessing = false
			prevDo = false
		}
	}

}

// CASE 7, Weight; hold the data and wait until weighing scale trigger to collect data to patch.
func handleWeight(jsonPayloads *SafeJsonPayloads, messages []model.Message, loop float64,
	apiUrl string, serviceRoleKey string, function string, chance bool) {

	// Process to handling counter when ch1 started
	processChannelTrigger("CASE_4_TRIGGER_CH1", "counterch_", jsonPayloads, messages, loop)

	// Handle different types (string and float64) of CH1_TRIGGER, CH2_TRIGGER, CH3_TRIGGER.
	// Process triggers for each channel
	processChannelTrigger("CASE_4_TRIGGER_CH1", "ch1_", jsonPayloads, messages, loop)
	processChannelTrigger("CASE_4_TRIGGER_CH2", "ch2_", jsonPayloads, messages, loop)
	processChannelTrigger("CASE_4_TRIGGER_CH3", "ch3_", jsonPayloads, messages, loop)

	// Process Vacuum Trigger
	VACUUM_TRIGGER, _ := jsonPayloads.Get(os.Getenv("CASE_4_VACUUM_reach_20pa"))
	if VACUUM_TRIGGER != nil {
		processAndPrintforVacuum("vacuum", jsonPayloads, messages, loop)
	}

	// Process CH1, CH2, CH3 Weight Triggers
	ProcessWeightTriggers(jsonPayloads, messages, loop)

	// Check if all weight triggers (CH1, CH2, CH3) are inactive, but were previously active
	if !weightTriggerCh1 && !weightTriggerCh2 && !weightTriggerCh3 &&
		prevWeightTriggerCh1 && prevWeightTriggerCh2 && prevWeightTriggerCh3 && chance {

		fmt.Println("All weight triggers are now inactive. Processing the patch.")

		// Merge data from different channels
		data := MergeNonEmptyMaps(
			processedPayloadsMap["ch1_"],
			processedPayloadsMap["ch2_"],
			processedPayloadsMap["ch3_"],
			processedPayloadsMap["vacuum"],
			processedPayloadsMap["weightch1_"],
			processedPayloadsMap["weightch2_"],
			processedPayloadsMap["weightch3_"],
			processedPayloadsMap["counterch_"],
		)

		// Send the data to the database
		startTime := time.Now()
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return
		}

		_, err = sendPatchRequest(apiUrl, serviceRoleKey, jsonData, function)
		if err != nil {
			panic(err)
		}

		// Measure elapsed time and log
		elapsedTime := time.Since(startTime)
		prettyPrintJSONWithTime(data, elapsedTime)

		// Clear the processedPayloadsMap after the patch
		for key := range processedPayloadsMap {
			delete(processedPayloadsMap, key)
		}

		// Reset previous triggers to avoid reprocessing
		prevWeightTriggerCh1 = false
		prevWeightTriggerCh2 = false
		prevWeightTriggerCh3 = false
		chance = false
	}
}

// CASE 8, HoldFillingWeight; hold the data and wait until weighing scale trigger to collect data to patch.
func handleHoldFillingWeightCase(jsonPayloads *SafeJsonPayloads, messages []model.Message,
	loop float64, apiUrl string, serviceRoleKey string, function string) {

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
			processedPayloadsMap[channel][channel+"_fill"] = 1
			isProcessing = true
		}
	}

	ch1Success, ok1 := jsonPayloads.GetFloat64(os.Getenv("CASE_6_TRIGGER_ch1"))
	ch2Success, ok2 := jsonPayloads.GetFloat64(os.Getenv("CASE_6_TRIGGER_ch2"))
	ch3Success, ok3 := jsonPayloads.GetFloat64(os.Getenv("CASE_6_TRIGGER_ch3"))

	if ok1 && ok2 && ok3 && isProcessing && ch1Success == 0 && ch2Success == 0 && ch3Success == 0 {
		prevDo := false

		processedPayloadsMap["do"] = ProcessTriggerGeneric(jsonPayloads, messages, loop, func(payload *SafeJsonPayloads) map[string]interface{} {
			prevDo = true
			return _hold_changeName_generic(payload, "CASE_6_DO_")
		})

		// Process CH1, CH2, CH3 Weight Triggers
		ProcessWeightTriggers(jsonPayloads, messages, loop)

		// Check if all weight triggers (CH1, CH2, CH3) are inactive, but were previously active
		if !weightTriggerCh1 && !weightTriggerCh2 && !weightTriggerCh3 &&
			prevWeightTriggerCh1 && prevWeightTriggerCh2 && prevWeightTriggerCh3 {

			fmt.Println("All weight triggers are now inactive. Processing the patch.")

			if prevDo {
				data := MergeNonEmptyMaps(
					processedPayloadsMap["ch1"],
					processedPayloadsMap["ch2"],
					processedPayloadsMap["ch3"],
					processedPayloadsMap["do"],
					processedPayloadsMap["weightch1_"],
					processedPayloadsMap["weightch2_"],
					processedPayloadsMap["weightch3_"],
				)

				startTime := time.Now()
				jsonData, err := json.Marshal(data)
				if err != nil {
					fmt.Println("Error marshaling JSON:", err)
					return
				}

				_, err = sendPatchRequest(apiUrl, serviceRoleKey, jsonData, function)
				if err != nil {
					panic(err)
				}

				elapsedTime := time.Since(startTime)
				prettyPrintJSONWithTime(data, elapsedTime)
				// Update the previous state of sealing
				isProcessing = false
				prevWeightTriggerCh1 = false
				prevWeightTriggerCh2 = false
				prevWeightTriggerCh3 = false
				prevDo = false
			}

		}

	}

}
