package main

import (
	//"log"
	//"net/http"
	//_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"gopatch/utils"
)

// Application-wide configuration variables
var (
	apiUrl         string  // URL for the API endpoint
	serviceRoleKey string  // Key to identify the service role
	function       string  // Name of the function to invoke for processing
	trigger        string  // Trigger identifier for the device or operation
	loopStr        string  // Looping parameter in string format
	loop           float64 // Looping parameter converted to float64
	filter         string  // Filter for processing MQTT messages

	broker        string // MQTT broker hostname
	port          string // MQTT broker port
	topic         string // MQTT topic to subscribe to
	mqttsStr      string // Indicates if MQTT Secure (MQTTS) is enabled ("true"/"false")
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
	// Channel for receiving MQTT messages as JSON strings
	receivedMessagesJSONChan := make(chan string)

	// Start the MQTT client in a separate goroutine
	go utils.Client(
		broker, port, topic, mqttsStr,
		ECScaCert, ECSclientCert, ECSclientKey,
		receivedMessagesJSONChan, clientDone,
	)

	// Process MQTT data
	go func() {
		defer close(clientDone)
		for {
			select {
			case <-stopProcessing:
				return
			default:
				utils.ProcessMQTTData(
					apiUrl, serviceRoleKey, receivedMessagesJSONChan,
					function, trigger, loop, filter,
				)
			}
		}
	}()

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	<-sigCh

	// Initiate graceful shutdown
	close(stopProcessing)

	// Wait for client to finish
	<-clientDone
}

// init initializes application configuration by loading environment variables.
func init() {
	// Load configuration (for local testing)
	utils.LoadEnv(".env.local")

	// Initialize application variable
	apiUrl = os.Getenv("API_URL")
	serviceRoleKey = os.Getenv("SERVICE_ROLE_KEY")
	function = os.Getenv("BASH_API")
	trigger = os.Getenv("TRIGGER_DEVICE")
	loopStr = os.Getenv("LOOPING")
	loop, _ = strconv.ParseFloat(loopStr, 64)
	filter = os.Getenv("FILTER")

	// MQTT configuration
	broker = os.Getenv("MQTT_HOST")
	port = os.Getenv("MQTT_PORT")
	topic = os.Getenv("MQTT_TOPIC")
	mqttsStr = os.Getenv("MQTTS_ON")

	// MQTT security certificates
	ECScaCert = os.Getenv("ECS_MQTT_CA_CERTIFICATE")
	ECSclientCert = os.Getenv("ECS_MQTT_CLIENT_CERTIFICATE")
	ECSclientKey = os.Getenv("ECS_MQTT_PRIVATE_KEY")
}
