package utils

import (
	"fmt"
	"log"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// TestMQTTClient tests the MQTT client subscribing to a topic
func TestMQTTClient(t *testing.T) {
	// Test variables, replace with your actual test parameters
	broker := "localhost" // Replace with actual broker
	port := "1883"        // Use 1883 for non-TLS or 8883 for TLS
	topic := "test/topic" // Replace with the topic you want to subscribe to
	mqttsStr := "false"   // Set to "true" for MQTT over TLS
	ECScaCert := ""       // CA Cert for TLS (empty if not using TLS)
	ECSclientCert := ""   // Client Cert for TLS (empty if not using TLS)
	ECSclientKey := ""    // Client Key for TLS (empty if not using TLS)

	// Channel to receive processed messages
	receivedMessagesJSONChan := make(chan string, 1)
	clientDone := make(chan struct{})

	// Start the MQTT client
	go Client(broker, port, topic, mqttsStr, ECScaCert, ECSclientCert, ECSclientKey, receivedMessagesJSONChan, clientDone)

	// Wait for some messages to be received
	select {
	case msg := <-receivedMessagesJSONChan:
		log.Printf("Received message: %s", msg)
	case <-time.After(10 * time.Second): // Timeout after 10 seconds if no message received
		t.Fatal("Test timed out. No message received.")
	}

	// Close the client when done
	close(clientDone)
}

// TestMQTTPublisher tests the MQTT publisher publishing a message to a topic
func TestMQTTPublisher(t *testing.T) {
	// Test variables, same as in the subscriber test
	broker := "localhost" // Replace with actual broker
	port := "1883"        // Use 1883 for non-TLS or 8883 for TLS
	topic := "test/topic" // Topic you want to publish to

	// Create MQTT client options
	opts := getClientOptions(broker, port)
	client := mqtt.NewClient(opts)

	// Connect the client
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		t.Fatalf("Error connecting to MQTT broker: %v", token.Error())
	}

	// Publish a test message
	payload := fmt.Sprintf(`{"address": "localhost", "value": "test"}`)
	token := client.Publish(topic, 0, false, payload)
	token.Wait()

	log.Println("Test message published to topic:", topic)

	// Disconnect the client
	client.Disconnect(250)
}

func Test(t *testing.T) {
	// Run the tests
	fmt.Println("Running MQTT Subscriber Test...")
	TestMQTTClient(t)

	// Allow time for test completion
	time.Sleep(2 * time.Second)

	fmt.Println("Running MQTT Publisher Test...")
	TestMQTTPublisher(t)

	// Wait for test completion
	time.Sleep(2 * time.Second)
}
