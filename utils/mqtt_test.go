package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock MQTT Client
type MockClient struct {
	mock.Mock
	mqtt.Client
}

func (m *MockClient) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	args := m.Called(topic, qos, retained, payload)
	return args.Get(0).(mqtt.Token)
}

func (m *MockClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	args := m.Called(topic, qos, callback)
	if args.Get(0) != nil {
		callback(nil, args.Get(1).(mqtt.Message))
	}
	return args.Get(0).(mqtt.Token)
}

func (m *MockClient) Connect() mqtt.Token {
	args := m.Called()
	return args.Get(0).(mqtt.Token)
}

func TestMessageReceived(t *testing.T) {
	receivedMessagesJSONChan := make(chan string, 1)

	ResetReceivedMessages()

	// Fill up the queue to trigger flush
	for i := 0; i < MaxQueueSize; i++ {
		payload := fmt.Sprintf(`{"address":"address_%d","value":"value_%d"}`, i, i)
		testMessage := &mockMessage{payload: []byte(payload)}
		messageReceived(testMessage, receivedMessagesJSONChan)
	}

	// Wait for and validate JSON output
	select {
	case jsonOutput := <-receivedMessagesJSONChan:
		var messages []MqttData
		err := json.Unmarshal([]byte(jsonOutput), &messages)
		assert.NoError(t, err, "JSON should unmarshal correctly")
		assert.Len(t, messages, MaxQueueSize, fmt.Sprintf("JSON should contain %d messages", MaxQueueSize))
		assert.Equal(t, "address_0", messages[0].Address)
		assert.Equal(t, "value_0", messages[0].Value)
		assert.Equal(t, fmt.Sprintf("address_%d", MaxQueueSize-1), messages[MaxQueueSize-1].Address)
	case <-time.After(2 * time.Second):
		t.Error("Expected JSON output but received none")
	}
}

func TestQueueReset(t *testing.T) {
	// Ensure a clean state
	receivedMessages = nil

	// Fill the queue to MaxQueueSize
	for i := 0; i < MaxQueueSize; i++ {
		message := MqttData{
			Address: "address_" + strconv.Itoa(i),
			Value:   i,
		}
		receivedMessages = append(receivedMessages, message)
	}

	t.Logf("Queue length before reset: %d", len(receivedMessages)) // Debugging

	// Create a channel for JSON data
	jsonChan := make(chan string, 1)

	// Reset and send messages
	resetAndSendMessages(jsonChan)

	// Verify queue reset
	assert.Empty(t, receivedMessages, "Message queue should be empty after reset")

	// Verify JSON output
	select {
	case jsonData := <-jsonChan:
		var messages []MqttData
		err := json.Unmarshal([]byte(jsonData), &messages)
		assert.NoError(t, err, "JSON should unmarshal correctly")
		assert.Len(t, messages, MaxQueueSize, "JSON should contain only MaxQueueSize messages")
	default:
		t.Error("Expected JSON output but received none")
	}
}

// Mock MQTT Message
type mockMessage struct {
	payload []byte
}

func (m *mockMessage) Duplicate() bool   { return false }
func (m *mockMessage) Qos() byte         { return 0 }
func (m *mockMessage) Retained() bool    { return false }
func (m *mockMessage) Topic() string     { return "test_topic" }
func (m *mockMessage) MessageID() uint16 { return 0 }
func (m *mockMessage) Payload() []byte   { return m.payload }
func (m *mockMessage) Ack()              {}
