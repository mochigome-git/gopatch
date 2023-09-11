package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttData struct {
	Address string      `json:"address"`
	Value   interface{} `json:"value"`
}

var ExportedReceivedMessages []MqttData
var ExportedReceivedMessagesJSON string

func Client(broker string, port string, topic string) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%s", broker, port))
	opts.SetClientID("go_mqtt_subscriber")
	opts.SetUsername("emqx")
	opts.SetPassword("public")
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error connecting to MQTT broker: %v", token.Error())
	}

	// Subscribe to a topic
	if token := client.Subscribe(topic, 0, messageReceived); token.Wait() && token.Error() != nil {
		log.Fatalf("Error subscribing to topic: %v", token.Error())
	}

	fmt.Printf("Subscribed to topic: %s\n", topic)

	// Wait for signals to gracefully shut down the subscriber
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Unsubscribe and disconnect from the MQTT broker
	client.Unsubscribe(topic)
	client.Disconnect(250)
}

func messageReceived(client mqtt.Client, msg mqtt.Message) {
	var mqttData MqttData
	if err := json.Unmarshal(msg.Payload(), &mqttData); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}

	ExportedReceivedMessages = append(ExportedReceivedMessages, mqttData)
	mu.Lock() // Lock access to the shared resource

	jsonData, err := json.Marshal(ExportedReceivedMessages)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		mu.Unlock()
		return
	}
	ExportedReceivedMessagesJSON = string(jsonData)
	mu.Unlock()
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected to MQTT broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection lost: %v\n", err)
}
