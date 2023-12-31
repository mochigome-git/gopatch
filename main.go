package main

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"patch/utils"
)

var (
	apiUrl         string
	serviceRoleKey string
	broker         string
	port           string
	topic          string
	function       string
	trigger        string
	loopStr        string
	loop           float64
	filter         string
)

func main() {
	stopProcessing := make(chan struct{})
	clientDone := make(chan struct{})
	receivedMessagesJSONChan := make(chan string) // Create a channel for received JSON data

	go utils.Client(broker, port, topic, receivedMessagesJSONChan, clientDone)

	go func() {
		defer close(clientDone)

		for {
			select {
			case <-stopProcessing:
				return
			default:
				utils.ProcessMQTTData(
					apiUrl,
					serviceRoleKey,
					receivedMessagesJSONChan,
					function,
					trigger,
					loop,
					filter,
				)
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
	//config.LoadEnv(".env.local")
	apiUrl = os.Getenv("API_URL")
	serviceRoleKey = os.Getenv("SERVICE_ROLE_KEY")
	broker = os.Getenv("MQTT_HOST")
	port = os.Getenv("MQTT_PORT")
	topic = os.Getenv("MQTT_TOPIC")
	function = os.Getenv("BASH_API")
	trigger = os.Getenv("TRIGGER_DEVICE")
	loopStr = os.Getenv("LOOPING")
	loop, _ = strconv.ParseFloat(loopStr, 64)
	filter = os.Getenv("FILTER")
}
