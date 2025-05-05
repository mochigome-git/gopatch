package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Application-wide configuration variables
var (
	APIUrl         string  // URL for the API endpoint
	ServiceRoleKey string  // Key to identify the service role
	Function       string  // Name of the function to invoke for processing
	Trigger        string  // Trigger identifier for the device or operation
	LoopStr        string  // Looping parameter in string format
	Loop           float64 // Looping parameter converted to float64
	Filter         string  // Filter for processing MQTT messages

	Broker        string // MQTT broker hostname
	Port          string // MQTT broker port
	Topic         string // MQTT topic to subscribe to
	MQTTSStr      string // Indicates if MQTT Secure (MQTTS) is enabled ("true"/"false")
	ECScaCert     string // ESC version direct read from params store
	ECSclientCert string // ESC version direct read from params store
	ECSclientKey  string // ESC version direct read from params store
)

type MqttConfig struct {
	Broker        string
	Port          string
	Topic         string
	MQTTSStr      string
	ECScaCert     string
	ECSclientCert string
	ECSclientKey  string
}

func GetMqttConfig() MqttConfig {
	return MqttConfig{
		Broker:        Broker,
		Port:          Port,
		Topic:         Topic,
		MQTTSStr:      MQTTSStr,
		ECScaCert:     ECScaCert,
		ECSclientCert: ECSclientCert,
		ECSclientKey:  ECSclientKey,
	}
}

// Load initializes all configuration variables from environment variables
func Load(files ...string) {
	// Try to load from the specified file first
	if len(files) > 0 {
		for _, file := range files {
			err := godotenv.Load(file)
			if err != nil {
				log.Printf("Info: %s not found or failed to load local.env, falling back to system environment", file)
			}
		}
	}

	APIUrl = os.Getenv("API_URL")
	ServiceRoleKey = getEnv("SERVICE_ROLE_KEY", "")
	Function = getEnv("BASH_API", "")
	Trigger = getEnv("TRIGGER_DEVICE", "")
	Filter = getEnv("FILTER", "d174")

	LoopStr = getEnv("LOOPING", "1")
	Loop, _ = strconv.ParseFloat(LoopStr, 64)

	Broker = os.Getenv("MQTT_HOST")
	Port = getEnv("MQTT_PORT", "8883")
	Topic = os.Getenv("MQTT_TOPIC")
	MQTTSStr = getEnv("MQTTS_ON", "true")
	ECScaCert = os.Getenv("ECS_MQTT_CA_CERTIFICATE")
	ECSclientCert = os.Getenv("ECS_MQTT_CLIENT_CERTIFICATE")
	ECSclientKey = os.Getenv("ECS_MQTT_PRIVATE_KEY")
}

// Helper to get environment variable with fallback
// AWS ECS only allow os.Getenv
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
