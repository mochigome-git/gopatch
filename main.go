package main

import (
	//"log"
	//"net/http"
	//_ "net/http/pprof"

	"os"
	"os/signal"
	"syscall"

	"gopatch/config"
	"gopatch/handler"
	"gopatch/mqtts"
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

	// Load configuration
	config.Load(".env.local")

	// Channels for communication and termination
	stopProcessing := make(chan struct{})
	clientDone := make(chan struct{})
	// Channel for receiving MQTT messages as JSON strings
	receivedMessagesJSONChan := make(chan string, 1000)

	// Start the MQTT client in a separate goroutine
	go mqtts.Client(
		config.GetMqttConfig(),
		receivedMessagesJSONChan,
		clientDone,
	)

	// Process MQTT data
	go func() {
		defer close(clientDone)
		for {
			select {
			case <-stopProcessing:
				return
			default:
				handler.ProcessMQTTData(
					config.GetAppConfig(), receivedMessagesJSONChan)
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
