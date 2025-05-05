package mqtts

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"gopatch/config"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

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
	droppedMessagesCount  int64
)

const (
	MinFlushSize  = 100             // Only flush if at least 100 messages
	MaxQueueSize  = 200             // Optional: maximum buffer size
	FlushInterval = 1 * time.Second // Force flush every second
)

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

func Client(cfg config.MqttConfig, receivedMessagesJSONChan chan<- string, clientDone chan<- struct{}) {
	// Parse the string value into a boolean, defaulting to false if parsing fails
	mqtts, _ := strconv.ParseBool(cfg.MQTTSStr)
	var opts *mqtt.ClientOptions
	if mqtts {
		var err error
		// Standard verion
		//opts, err = getClientOptionsTLS(broker, port, caCertFile, clientCertFile, clientKeyFile)

		// AWS ECS version
		opts, err = ECSgetClientOptionsTLS(cfg.Broker, cfg.Port, cfg.ECScaCert, cfg.ECSclientCert, cfg.ECSclientKey)
		if err != nil {
			log.Fatalf("Error requesting MQTT TLS configuration: %v", err.Error())
			return
		}
	} else {
		opts = getClientOptions(cfg.Broker, cfg.Port)
	}
	client := mqtt.NewClient(opts)

	maxAttempts := 5
	for i := 1; i <= maxAttempts; i++ {
		if token := client.Connect(); token.Wait() && token.Error() == nil {
			break
		} else {
			log.Printf("MQTT connect failed (attempt %d/%d): %v", i, maxAttempts, token.Error())
			time.Sleep(2 * time.Second)
			if i == maxAttempts {
				log.Fatalf("MQTT connect failed after %d attempts", maxAttempts)
			}
		}
	}

	if token := client.Subscribe(cfg.Topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		messageReceived(msg)
	}); token.Wait() && token.Error() != nil {
		log.Fatalf("Error subscribing to topic: %v", token.Error())
		return
	}

	log.Printf("Subscribed to topic: %s\n", cfg.Topic)

	// Start background batch flusher
	stopFlusher := make(chan struct{})
	go startBatchFlusher(receivedMessagesJSONChan, stopFlusher)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	close(stopFlusher)
	client.Unsubscribe(cfg.Topic)
	client.Disconnect(250)
	close(clientDone)
	log.Println("MQTT client shut down gracefully.")
}

// messageReceived handles the received MQTT message
func messageReceived(msg mqtt.Message) {
	var mqttData MqttData
	if err := json.Unmarshal(msg.Payload(), &mqttData); err != nil {
		log.Printf("Error parsing JSON: %v\n", err)
		return
	}

	receivedMessagesMutex.Lock()
	receivedMessages = append(receivedMessages, mqttData)
	receivedMessagesMutex.Unlock()
}

func startBatchFlusher(receivedMessagesJSONChan chan<- string, stopFlusher <-chan struct{}) {
	ticker := time.NewTicker(FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			flushMessages(receivedMessagesJSONChan, true) // Forced flush by timer
		case <-stopFlusher:
			flushMessages(receivedMessagesJSONChan, true) // Final flush
			return
		default:
			time.Sleep(50 * time.Millisecond)
			flushMessages(receivedMessagesJSONChan, false) // Soft flush
		}
	}
}

func flushMessages(receivedMessagesJSONChan chan<- string, force bool) {
	receivedMessagesMutex.Lock()
	defer receivedMessagesMutex.Unlock()

	queueLen := len(receivedMessages)
	if queueLen == 0 {
		return
	}

	// Only flush if queue is big enough OR if forced flush
	if queueLen >= MinFlushSize || force {
		messagesToSend := receivedMessages
		receivedMessages = nil

		jsonData, err := json.Marshal(messagesToSend)
		if err != nil {
			log.Printf("Error marshaling JSON: %v\n", err)
			return
		}

		select {
		case receivedMessagesJSONChan <- string(jsonData):
		default:
			atomic.AddInt64(&droppedMessagesCount, 1)
			log.Println("Received data dropped, channel full")
		}
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected to MQTT broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Fatalf("Connection lost: %v\n", err)
}

func ResetReceivedMessages() {
	receivedMessagesMutex.Lock()
	receivedMessages = []MqttData{}
	receivedMessagesMutex.Unlock()
}
