package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

type MqttData struct {
	Address string      `json:"address"`
	Value   interface{} `json:"value"`
}

var (
	receivedMessages      []MqttData
	receivedMessagesJSON  string
	receivedMessagesMutex sync.Mutex
)

var mqttData MqttData

func getClientOptions(broker, port, topic string) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%s", broker, port))
	clientID := "go_mqtt_subscriber_" + uuid.New().String()
	opts.SetClientID(clientID)
	opts.SetUsername("emqx")
	opts.SetPassword("public")
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	return opts
}

func Client(broker, port, topic string, receivedMessagesJSONChan chan<- string, clientDone chan<- struct{}) {
	opts := getClientOptions(broker, port, topic)
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("Error connecting to MQTT broker: %v", token.Error())
		return
	}

	if token := client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		messageReceived(client, msg, receivedMessagesJSONChan)
	}); token.Wait() && token.Error() != nil {
		log.Printf("Error subscribing to topic: %v", token.Error())
		return
	}

	log.Printf("Subscribed to topic: %s\n", topic)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	client.Unsubscribe(topic)
	client.Disconnect(250)

	close(clientDone)
}

func messageReceived(client mqtt.Client, msg mqtt.Message, receivedMessagesJSONChan chan<- string) {
	if err := json.Unmarshal(msg.Payload(), &mqttData); err != nil {
		log.Printf("Error parsing JSON: %v\n", err)
		return
	}

	receivedMessagesMutex.Lock()
	defer receivedMessagesMutex.Unlock()
	receivedMessages = append(receivedMessages, mqttData)
	jsonData, err := json.Marshal(receivedMessages)
	if err != nil {
		log.Printf("Error marshaling JSON: %v\n", err)
	} else {
		receivedMessagesJSON = string(jsonData)
	}

	// Send the received JSON data to the processing channel
	select {
	case receivedMessagesJSONChan <- receivedMessagesJSON:
		//log.Printf("Received and sent JSON data: %s\n", receivedMessagesJSON)
		ResetReceivedMessages()
	default:
		//log.Println("Received data dropped, channel full")
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected to MQTT broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connection lost: %v\n", err)
}

func ResetReceivedMessages() {
	// Reset the receivedMessages slice to contain only mqttData
	receivedMessages = []MqttData{mqttData}
}
