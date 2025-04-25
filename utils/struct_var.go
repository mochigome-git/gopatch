package utils

import (
	"sync"
	"time"
)

// Define the device struct with the address field
type TriggerKey struct {
	triggerKey string
	caseKey    string
}

// JsonPayloads is a type representing the payload structurr for All case.
// Case 1,2,3,4 involving all the device
type JsonPayloads map[string]interface{}

// Time process to handling data
var stopProcessing = make(chan struct{})

// Stopwatch to count the device duration in Case 1.
var deviceStartTimeMap = make(map[string]time.Time)

// To store process-specific previous trigger keys
var processPrevTriggerKeyMap = make(map[string]string)

// Sync to protect processedPayloadsMap
var processedPayloadsMu sync.Mutex

// Create a map to store processed payloads (ch1, ch2, ch3; _xx_jsonPayloads)
var processedPayloadsMap = map[string]map[string]interface{}{
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
}

// To store the trigger of condition judgement in case 4
var prevSealing float64

// Flag to track if the process is active
var isProcessing bool

// Case 5
var triggerChanCase5 = make(chan int)

// Global variables to Track previous trigger state for Case 7
var prevWeightTriggerCh1 bool
var prevWeightTriggerCh2 bool
var prevWeightTriggerCh3 bool
var weightTriggerCh1 bool
var weightTriggerCh2 bool
var weightTriggerCh3 bool
var chance bool

var prevWeightValueCh1 = new(float64)
var prevWeightValueCh2 = new(float64)
var prevWeightValueCh3 = new(float64)
