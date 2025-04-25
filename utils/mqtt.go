package utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
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
	receivedMessagesMutex sync.Mutex
	mqttData              MqttData
)

// Define the size of the fixed size queue
const MaxQueueSize = 200

func getClientOptions(broker, port string) *mqtt.ClientOptions {
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

func ECSgetClientOptionsTLS(broker, port, ECScaCert, ECSclientCert, ECSclientKey string) (*mqtt.ClientOptions, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("mqtts://%s:%s", broker, port))
	clientID := "go_mqtt_subscriber_" + uuid.New().String()

	// Load client certificate and key
	cert, err := tls.X509KeyPair([]byte(ECSclientCert), []byte(ECSclientKey))
	if err != nil {
		return nil, fmt.Errorf("error loading client certificate/key: %s", err)
	}

	// Create a certificate pool and add CA certificate
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM([]byte(ECScaCert)) {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	}

	opts.SetClientID(clientID)
	opts.SetUsername("emqx")
	opts.SetPassword("public")
	opts.SetTLSConfig(tlsConfig)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	return opts, nil
}

func Client(broker, port, topic, mqttsStr, ECScaCert, ECSclientCert, ECSclientKey string, receivedMessagesJSONChan chan<- string, clientDone chan<- struct{}) {
	// Parse the string value into a boolean, defaulting to false if parsing fails
	mqtts, _ := strconv.ParseBool(mqttsStr)
	var opts *mqtt.ClientOptions

	if mqtts {
		var err error
		// Standard verion
		//opts, err = getClientOptionsTLS(broker, port, caCertFile, clientCertFile, clientKeyFile)
		// AWS ECS version
		opts, err = ECSgetClientOptionsTLS(broker, port, ECScaCert, ECSclientCert, ECSclientKey)
		if err != nil {
			log.Fatalf("Error requesting MQTT TLS configuration: %v", err.Error())
			return
		}
	} else {
		opts = getClientOptions(broker, port)
	}
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error connecting to MQTT broker: %v", token.Error())
		return
	}

	if token := client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		messageReceived(msg, receivedMessagesJSONChan)
	}); token.Wait() && token.Error() != nil {
		log.Fatalf("Error subscribing to topic: %v", token.Error())
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

// messageReceived handles the received MQTT message
func messageReceived(msg mqtt.Message, receivedMessagesJSONChan chan<- string) {
	// Unmarshal the received MQTT message payload into mqttData
	if err := json.Unmarshal(msg.Payload(), &mqttData); err != nil {
		log.Printf("Error parsing JSON: %v\n", err)
		return
	}

	// Lock the mutex to synchronize access to the message queue
	receivedMessagesMutex.Lock()
	defer receivedMessagesMutex.Unlock()

	// Append the newly received message to the message queue
	receivedMessages = append(receivedMessages, mqttData)

	// If the queue reaches MaxQueueSize, reset and send the messages
	if len(receivedMessages) >= MaxQueueSize {
		//log.Println("Queue is full, resetting and sending messages")
		resetAndSendMessages(receivedMessagesJSONChan)
	}

}

// resetAndSendMessages marshals the received messages, sends them to the processing channel,
// and resets the receivedMessages queue
func resetAndSendMessages(receivedMessagesJSONChan chan<- string) {
	// Marshal the received messages into JSON
	jsonData, err := json.Marshal(receivedMessages)
	if err != nil {
		log.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	select {
	case receivedMessagesJSONChan <- string(jsonData):
		// Successfully sent JSON data to the processing channel
	default:
		// Processing channel is full, drop the message or handle accordingly
		//log.Println("Received data dropped, channel full")
	}

	// Reset the receivedMessages queue
	receivedMessages = nil
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected to MQTT broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Fatalf("Connection lost: %v\n", err)
}

func ResetReceivedMessages() {
	// Reset the receivedMessages slice to contain only mqttData
	receivedMessages = []MqttData{}
}
