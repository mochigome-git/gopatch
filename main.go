package main

import (
	//"log"
	//"net/http"
	//_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"patch/utils"
)

var (
	apiUrl         string // Database url
	serviceRoleKey string
	function       string
	trigger        string
	loopStr        string
	loop           float64
	filter         string

	broker        string // broker stores the MQTT broker's hostname
	port          string // mqttport stores the MQTT broker's port number
	topic         string // topic stores the topic of the MQTT broker
	mqttsStr      string // mqtts true or false
	ECScaCert     string // ESC verion direct read from params store
	ECSclientCert string // ESC verion direct read from params store
	ECSclientKey  string // ESC verion direct read from params store
)

func main() {
	// Register the profiling handlers with the default HTTP server mux.
	// This will serve the profiling endpoints at /debug/pprof.
	/**
	Memory profile: http://localhost:6060/debug/pprof/heap
	Goroutine profile: http://localhost:6060/debug/pprof/goroutine
	CPU profile: http://localhost:6060/debug/pprof/profile
	**/
	// Start profiling server
	//go func() {
	//	if err := http.ListenAndServe("192.168.0.126:6060", nil); err != nil {
	//		log.Fatalf("Error starting profiling server: %v", err)
	//	}
	//}()

	// Channels for communication and termination
	stopProcessing := make(chan struct{})
	clientDone := make(chan struct{})
	receivedMessagesJSONChan := make(chan string) // Create a channel for received JSON data

	// Start MQTT client
	go utils.Client(broker, port, topic, mqttsStr, ECScaCert, ECSclientCert, ECSclientKey, receivedMessagesJSONChan, clientDone)

	// Process MQTT data
	go func() {
		defer close(clientDone)
		for {
			select {
			case <-stopProcessing:
				return
			default:
				utils.ProcessMQTTData(apiUrl, serviceRoleKey, receivedMessagesJSONChan, function, trigger, loop, filter)
			}
		}
	}()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	<-sigCh

	// Signal to stop processing
	close(stopProcessing)

	// Wait for client to finish
	<-clientDone
}

func init() {
	// local test
	utils.LoadEnv(".env.local")
	apiUrl = os.Getenv("API_URL")
	serviceRoleKey = os.Getenv("SERVICE_ROLE_KEY")
	function = os.Getenv("BASH_API")
	trigger = os.Getenv("TRIGGER_DEVICE")
	loopStr = os.Getenv("LOOPING")
	loop, _ = strconv.ParseFloat(loopStr, 64)
	filter = os.Getenv("FILTER")

	broker = os.Getenv("MQTT_HOST")
	port = os.Getenv("MQTT_PORT")
	topic = os.Getenv("MQTT_TOPIC")
	mqttsStr = os.Getenv("MQTTS_ON")
	ECScaCert = os.Getenv("ECS_MQTT_CA_CERTIFICATE")
	ECSclientCert = os.Getenv("ECS_MQTT_CLIENT_CERTIFICATE")
	ECSclientKey = os.Getenv("ECS_MQTT_PRIVATE_KEY")
}
