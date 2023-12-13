package utils

import (
	"encoding/json"
	"fmt"
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
	if value, ok := jsonPayloads[tk.triggerKey]; ok {
		if trigger, ok := value.(float64); ok {
			if trigger == 1 {
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
	}
}
