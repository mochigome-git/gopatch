package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"patch/model"
	"strings"
	"time"
)

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
			if accum_rate, exists := jsonPayloads[os.Getenv("CASE_4_AVOID_0")].(float64); exists && accum_rate == 0 {
				// Skip further processing if accum_rate is 0
				return
			}
			handleHoldCase(tk, jsonPayloads, messages, loop, filter, apiUrl, serviceRoleKey, function)

		case tk.caseKey == "special":
			handleSpecialCase(tk, jsonPayloads, messages, loop, apiUrl, serviceRoleKey, function)
		}
	}
}

// CASE 1, time.Duration; handling the process of time taken from 0 to 1, and record the total time duration
func handleTimeDurationCase(tk TriggerKey, jsonPayloads JsonPayloads, messages []model.Message, loop float64, filter string, apiUrl string, serviceRoleKey string, function string) {
	processKey := generateProcessKey(tk.triggerKey)

	if tk.triggerKey != processPrevTriggerKeyMap[processKey] {
		processPrevTriggerKeyMap[processKey] = tk.triggerKey

		if trigger, ok := jsonPayloads[tk.triggerKey].(float64); ok && trigger != 0 {
			handleTimeDurationTrigger(tk, jsonPayloads, messages, loop, filter, apiUrl, serviceRoleKey, function)
		}
	}
}

// process for case1, check the time taken from 0 to 1.
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

// CASE 2, Standard; handling a devices value and patch it, when the trigger is different with previous key
func handleStandardCase(tk TriggerKey, jsonPayloads JsonPayloads, messages []model.Message, loop float64, filter string, apiUrl string, serviceRoleKey string, function string) {
	processKey := generateProcessKey(tk.triggerKey)

	if tk.triggerKey != processPrevTriggerKeyMap[processKey] {
		processPrevTriggerKeyMap[processKey] = tk.triggerKey

		if trigger, ok := jsonPayloads[tk.triggerKey].(float64); ok && trigger != 0 {
			var startTime time.Time
			processMessagesLoop(jsonPayloads, messages, startTime, loop)

			calculateAndStoreInklot(jsonPayloads)
			changeName(jsonPayloads)

			if trigger, ok := jsonPayloads[tk.triggerKey].(float64); ok && trigger == 0 {
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
func handleTriggerCase(tk TriggerKey, jsonPayloads JsonPayloads, messages []model.Message, loop float64, filter string, apiUrl string, serviceRoleKey string, function string) {
	if value, ok := jsonPayloads[tk.triggerKey].(float64); ok && value == 1 {

		startTime := time.Now()
		processMessagesLoop(jsonPayloads, messages, startTime, loop)

		if _filter, ok := jsonPayloads[filter].(float64); ok && _filter != 0 {
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
func handleHoldCase(tk TriggerKey, jsonPayloads JsonPayloads, messages []model.Message, loop float64, filter string, apiUrl string, serviceRoleKey string, function string) {

	if CH1_TRIGGER, ok := jsonPayloads[os.Getenv("CASE_4_TRIGGER_CH1")].(float64); ok && CH1_TRIGGER == 1 {
		// Use ProcessTriggerGeneric for ch2
		processedPayloadsMap["ch1"] = ProcessTriggerGeneric(jsonPayloads, messages, loop, func(payload JsonPayloads) map[string]interface{} {
			return _hold_changeName_generic(payload, "HOLD_KEY_TRANSOFRMATION_ch1_")
		})
	}
	if CH2_TRIGGER, ok := jsonPayloads[os.Getenv("CASE_4_TRIGGER_CH2")].(float64); ok && CH2_TRIGGER == 1 {
		// Use ProcessTriggerGeneric for ch2
		processedPayloadsMap["ch2"] = ProcessTriggerGeneric(jsonPayloads, messages, loop, func(payload JsonPayloads) map[string]interface{} {
			return _hold_changeName_generic(payload, "HOLD_KEY_TRANSOFRMATION_ch2_")
		})
	}
	if CH3_TRIGGER, ok := jsonPayloads[os.Getenv("CASE_4_TRIGGER_CH3")].(float64); ok && CH3_TRIGGER == 1 {
		// Use ProcessTriggerGeneric for ch3
		processedPayloadsMap["ch3"] = ProcessTriggerGeneric(jsonPayloads, messages, loop, func(payload JsonPayloads) map[string]interface{} {
			return _hold_changeName_generic(payload, "HOLD_KEY_TRANSOFRMATION_ch3_")
		})
	}

	if sealing, ok := jsonPayloads[os.Getenv("CASE_4_SEALING")].(float64); ok {
		//fmt.Printf("Sealing: %d; prevSealing: %d\n", int(sealing), int(prevSealing))

		if sealing == 1 {
			processedPayloadsMap["vacuum"] = ProcessTriggerGeneric(jsonPayloads, messages, loop, func(payload JsonPayloads) map[string]interface{} {
				return _hold_changeName_generic(payload, "CASE_4_VACUUM_")
			})
			prevSealing = sealing
		}

		if sealing == 0 && prevSealing == 1 {
			data := MergeNonEmptyMaps(
				processedPayloadsMap["ch1"],
				processedPayloadsMap["ch2"],
				processedPayloadsMap["ch3"],
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

// CASE 5, Special;
// handling a device's highest value and average value and patch it,
// when the trigger is 1
func handleSpecialCase(tk TriggerKey, jsonPayloads JsonPayloads, messages []model.Message, loop float64, apiUrl string, serviceRoleKey string, function string) {
	// Assuming these variables need to be declared and initialized
	var startTime time.Time

	if trigger, ok := jsonPayloads[tk.triggerKey].(float64); ok {

		if trigger == 1 {
			isProcessing = true
			// Assuming processedPayloadsMap is a map[string]map[string]interface{}
			if processedPayloadsMap["degas"]["pica1"] == nil {
				processedPayloadsMap["degas"]["pica1"] = make([]float64, 0)
			}

			result := ProcessTriggerGenericSpecial(jsonPayloads, messages, trigger, func(payload JsonPayloads) map[string]interface{} {
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
