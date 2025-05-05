package handler

import (
	"gopatch/model"
	"gopatch/utils"
	"os"
)

type AccumCheckFunc func() bool // Check Accumalate Rate if 0 skip process

func Trigger(
	session *Session,
	jsonPayloads *utils.SafeJsonPayloads,
	messages []model.Message,
	triggerKey string,
	loop float64,
	filter string,
	apiUrl string,
	serviceRoleKey string,
	function string,
) {
	// Parse trigger keys once
	triggerKeys := utils.ParseTriggerKey(triggerKey)

	// Avoiding repeated logic for accum_rate checks
	isAccRate := func() bool {
		accum_rate, exists := jsonPayloads.GetFloat64(os.Getenv("CASE_4_AVOID_0"))
		return exists && accum_rate == 0
	}

	// Iterate over trigger keys
	for _, tk := range triggerKeys {
		// Map of case keys to handler functions
		caseHandlers := map[string]func(){
			"time.duration": func() { handleTimeDurationCase(tk, jsonPayloads, messages, loop) },
			"standard":      func() { handleStandardCase(tk, jsonPayloads, messages, loop, apiUrl, serviceRoleKey, function) },
			"trigger":       func() { handleTriggerCase(tk, jsonPayloads, messages, loop, filter, apiUrl, serviceRoleKey, function) },
			"hold":          func() { handleHoldCase(session, jsonPayloads, messages, apiUrl, serviceRoleKey, function, isAccRate) },
			"special":       func() { handleSpecialCase(session, tk, jsonPayloads, messages, apiUrl, serviceRoleKey, function) },
			"holdfilling":   func() { handleHoldFillingCase(session, jsonPayloads, messages, apiUrl, serviceRoleKey, function) },
			"weight": func() {
				handleWeight(session, jsonPayloads, messages, apiUrl, serviceRoleKey, function, false, isAccRate)
			},
			"holdfillingweight": func() { handleHoldFillingWeightCase(session, jsonPayloads, messages, apiUrl, serviceRoleKey, function) },
		}
		// Check if the current caseKey is in the map, and handle accordingly
		if handler, exists := caseHandlers[tk.CaseKey]; exists {
			handler()
		}
	}
}
