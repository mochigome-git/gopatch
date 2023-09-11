package main

import (
	"os"
	"sync"
	"time"

	"patch/utils"
)

// apiUrl stores the URL for the API service.
var apiUrl string

// serviceRoleKey stores the key for authenticating with the service.
var serviceRoleKey string

// broker stores the MQTT broker's hostname.
var broker string

// port stores the MQTT broker's port number.
var port string

// topic of the MQTT broker
var topic string

func main() {
	clientDone := make(chan struct{})
	stopProcessing := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		utils.Client(broker, port, topic)
	}()

	go func() {
		defer close(clientDone)

		for {
			select {
			case <-stopProcessing:
				return
			default:
				utils.ProcessMQTTData(apiUrl, serviceRoleKey)
			}
		}
	}()

	time.Sleep(3 * time.Second)

	close(stopProcessing)

	wg.Wait()
}

func init() {
	apiUrl = os.Getenv("API_URL")
	serviceRoleKey = os.Getenv("SERVICE_ROLE_KEY")
	broker = os.Getenv("MQTT_HOST")
	port = os.Getenv("MQTT_PORT")
	topic = os.Getenv("MQTT_TOPIC")
}
