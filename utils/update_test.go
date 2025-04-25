package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

// Mocking the SafeJsonPayloads
type MockSafeJsonPayloads struct {
	mock.Mock
}

func (m *MockSafeJsonPayloads) Set(key string, value interface{}) {
	m.Called(key, value)
}

func (m *MockSafeJsonPayloads) GetData() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

/**func TestProcessMQTTData(t *testing.T) {
	// Setup mock data
	messages := []model.Message{
		{
			Address: "device_1",
			Value:   "value_1",
		},
		{
			Address: "device_2",
			Value:   "value_2",
		},
	}

	// Convert the messages to a JSON string
	messagesJSON, err := json.Marshal(messages)
	assert.NoError(t, err)

	// Create the channel to simulate receiving messages
	receivedMessagesJSONChan := make(chan string, 1)
	receivedMessagesJSONChan <- string(messagesJSON)

	// Create a mock SafeJsonPayloads
	mockJsonPayloads := new(MockSafeJsonPayloads)
	mockJsonPayloads.On("Set", "device_1", "value_1").Return()
	mockJsonPayloads.On("Set", "device_2", "value_2").Return()

	// Set up a proper trigger value
	trigger := "device_1_value" // Ensure this matches the expected format

	// Call the ProcessMQTTData function in a goroutine
	stopProcessing := make(chan struct{})
	go func() {
		ProcessMQTTData(
			"apiUrl",
			"serviceRoleKey",
			receivedMessagesJSONChan,
			"function",
			trigger,
			2.0,
			"filter",
		)
	}()

	// Wait a short time to allow the function to process
	time.Sleep(2 * time.Second)

	// Assert that Set was called with the correct arguments
	mockJsonPayloads.AssertExpectations(t)

	// Test stop processing channel
	close(stopProcessing)
}**/

func TestPrettyPrintJSONWithTime(t *testing.T) {
	// Sample data for testing prettyPrintJSONWithTime
	data := map[string]interface{}{
		"device": "device_1",
		"value":  "value_1",
	}

	// Test the function with a map[string]interface{} type
	startTime := time.Now()
	prettyPrintJSONWithTime(data, time.Since(startTime))

	// Test with SafeJsonPayloads (mock)
	mockJsonPayloads := new(MockSafeJsonPayloads)
	mockJsonPayloads.On("GetData").Return(data)

	// Test the function with SafeJsonPayloads type
	prettyPrintJSONWithTime(mockJsonPayloads, time.Since(startTime))
}
