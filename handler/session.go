package handler

import (
	"sync"
)

var (
	sessionStore = make(map[string]*Session)
	sessionMutex sync.Mutex // prevent race conditions if accessed concurrently
	prev         map[string]interface{}
)

type Session struct {
	Mutex                sync.Mutex // Sync to protect session.ProcessedPayloadsMap
	PrevSealing          float64    // To store the trigger of condition judgement in case 4
	IsProcessing         bool       // Flag to track if the process is active
	AllSuccessZero       bool
	WeightTriggerCh1     bool
	WeightTriggerCh2     bool
	WeightTriggerCh3     bool
	PrevWeightTriggerCh1 bool
	PrevWeightTriggerCh2 bool
	PrevWeightTriggerCh3 bool
	PrevWeightValueCh1   *float64
	PrevWeightValueCh2   *float64
	PrevWeightValueCh3   *float64
	ProcessedPayloadsMap map[string]map[string]interface{}
}

func NewSession() *Session {
	return &Session{
		// Create a map to store processed payloads (ch1, ch2, ch3; _xx_jsonPayloads) for holdCase
		ProcessedPayloadsMap: map[string]map[string]interface{}{
			"ch1_":       make(map[string]interface{}),
			"ch2_":       make(map[string]interface{}),
			"ch3_":       make(map[string]interface{}),
			"ch1":        make(map[string]interface{}),
			"ch2":        make(map[string]interface{}),
			"ch3":        make(map[string]interface{}),
			"vacuum":     make(map[string]interface{}),
			"degas":      make(map[string]interface{}),
			"do":         make(map[string]interface{}),
			"weightch1_": make(map[string]interface{}),
			"weightch2_": make(map[string]interface{}),
			"weightch3_": make(map[string]interface{}),
			"counter":    make(map[string]interface{}),
		},
		PrevWeightValueCh1: new(float64),
		PrevWeightValueCh2: new(float64),
		PrevWeightValueCh3: new(float64),
		AllSuccessZero:     false,
	}
}

// GetOrCreateSession ensures a session exists for a caseKey
func GetOrCreateSession(caseKey string) *Session {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	if s, ok := sessionStore[caseKey]; ok {
		return s
	}

	newSession := NewSession()
	sessionStore[caseKey] = newSession
	return newSession
}

func ClearSession(caseKey string) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	delete(sessionStore, caseKey)
}
