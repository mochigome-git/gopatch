package utils

import (
	"gopatch/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Mock the model.Message structure for testing
type MockMessage struct {
	Address string
	Value   interface{}
}

// You can also directly convert MockMessage into model.Message
func (m *MockMessage) ToModelMessage() model.Message {
	return model.Message{
		Address: m.Address,
		Value:   m.Value,
	}
}
func TestProcessMessagesLoop(t *testing.T) {
	// Create a new SafeJsonPayloads object
	jsonPayloads := NewSafeJsonPayloads()

	// Create mock messages of type MockMessage
	mockMessages := []MockMessage{
		{Address: "Sensor1", Value: 42},
		{Address: "Sensor2", Value: "active"},
	}

	// Convert MockMessage to model.Message
	var messages []model.Message
	for _, msg := range mockMessages {
		messages = append(messages, msg.ToModelMessage()) // Convert to model.Message
	}

	// Run the processMessagesLoop function with a 1-second duration
	startTime := time.Now()
	processMessagesLoop(jsonPayloads, messages, startTime, 1)

	// Check if the data has been set correctly
	val1, _ := jsonPayloads.Get("sensor1")
	val2, _ := jsonPayloads.Get("sensor2")

	assert.Equal(t, 42, val1)
	assert.Equal(t, "active", val2)
}

func TestProcessTriggerGenericSpecial(t *testing.T) {
	// Create a new SafeJsonPayloads object
	jsonPayloads := NewSafeJsonPayloads()

	// Create mock messages of type MockMessage
	mockMessages := []MockMessage{
		{Address: "Sensor1", Value: 42},
		{Address: "Sensor2", Value: "active"},
	}

	// Convert MockMessage to model.Message
	var messages []model.Message
	for _, msg := range mockMessages {
		messages = append(messages, msg.ToModelMessage()) // Convert to model.Message
	}

	// Test ProcessTriggerGenericSpecial
	processedPayload := ProcessTriggerGenericSpecial(jsonPayloads, messages, 1, func(payload *SafeJsonPayloads) map[string]interface{} {
		// Your custom logic here
		return payload.GetData() // Example: just return the payload data
	})

	// Check the processed data (you can customize this check)
	assert.NotNil(t, processedPayload)
}

/**func TestHandleTimeDurationTrigger(t *testing.T) {
	// Create a new SafeJsonPayloads object
	jsonPayloads := NewSafeJsonPayloads()

	// Create mock messages of type MockMessage
	mockMessages := []MockMessage{
		{Address: "Sensor1", Value: 42},
		{Address: "Sensor2", Value: "active"},
	}

	// Convert MockMessage to model.Message
	var messages []model.Message
	for _, msg := range mockMessages {
		messages = append(messages, msg.ToModelMessage()) // Convert to model.Message
	}

	// Test handleTimeDurationTrigger (make sure TriggerKey is set up)
	tk := TriggerKey{triggerKey: "Sensor1"}
	handleTimeDurationTrigger(tk, jsonPayloads, messages, 1)

	// Check if the correct data has been processed (you can customize this check)
	val, _ := jsonPayloads.Get("sensor1")
	assert.Equal(t, 42, val)
} **/

// Mock version of Getenv
/**var getenvMock = os.Getenv

// Set mock environment variables for testing
func SetMockEnv(key, value string) {
	os.Setenv(key, value)
	getenvMock = os.Getenv // Set to the mock getenv
}

// Restore original getenv function
func RestoreEnv() {
	getenvMock = os.Getenv
}

// We will create a mock of the global processedPayloadsMap for testing purposes
var mockProcessedPayloadsMap = map[string]map[string]interface{}{
	"weightch1_": make(map[string]interface{}),
	"weightch2_": make(map[string]interface{}),
	"weightch3_": make(map[string]interface{}),
}

// Your unit test function
func TestProcessWeightTriggers(t *testing.T) {
	// Set mock environment variables
	SetMockEnv("CASE_7_TRIGGER_WEIGHING_CH1", "m3330")
	SetMockEnv("CASE_7_TRIGGER_WEIGHING_CH2", "m3400")
	SetMockEnv("CASE_7_TRIGGER_WEIGHING_CH3", "m3500")
	defer RestoreEnv() // Restore environment after the test

	// Create mock payloads and messages
	jsonPayloads := &SafeJsonPayloads{}
	mockMessages := []MockMessage{
		{Address: "Sensor1", Value: 42},
		{Address: "Sensor2", Value: "active"},
	}

	var messages []model.Message
	for _, msg := range mockMessages {
		messages = append(messages, msg.ToModelMessage())
	}

	// Add WaitGroup to wait for goroutines to finish
	var wg sync.WaitGroup

	// Add 3 goroutines to the WaitGroup
	wg.Add(3)

	// Mock function to process weight triggers (simplified)
	go func() {
		defer wg.Done()
		fmt.Println("Processing weight trigger for CH1...")
		ProcessWeightTriggers(jsonPayloads, messages, 1) // Call your function with test inputs
	}()

	// Wait for goroutines to finish
	wg.Wait()

	// Perform assertions after waiting for all goroutines to complete
	// Check the values in mockProcessedPayloadsMap
	sensor1Value, exists := mockProcessedPayloadsMap["weightch1_"]["sensor1"]
	assert.True(t, exists, "sensor1 value should exist")
	assert.Equal(t, 42, sensor1Value, "sensor1 value should be 42")

	sensor2Value, exists := mockProcessedPayloadsMap["weightch2_"]["sensor2"]
	assert.True(t, exists, "sensor2 value should exist")
	assert.Equal(t, "active", sensor2Value, "sensor2 value should be 'active'")
}**/
