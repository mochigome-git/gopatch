package utils

import "time"

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

// Create a map to store processed payloads (ch1, ch2, ch3; _xx_jsonPayloads)
var processedPayloadsMap = map[string]map[string]interface{}{
	"ch1":    make(map[string]interface{}),
	"ch2":    make(map[string]interface{}),
	"ch3":    make(map[string]interface{}),
	"vacuum": make(map[string]interface{}),
	"degas":  make(map[string]interface{}),
	"do":     make(map[string]interface{}),
}

// To store the trigger of condition judgement in case 4
var prevSealing float64

// Flag to track if the process is active
var isProcessing bool

// Case 5
var triggerChanCase5 = make(chan int)
