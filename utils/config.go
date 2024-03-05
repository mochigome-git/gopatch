package utils

import (
	"log"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from multiple .env.local files in a specific order
func LoadEnv(files ...string) {
	for _, file := range files {
		err := godotenv.Load(file)
		if err != nil {
			log.Fatalf("Error loading %s file: %v", file, err)
		}
	}
}
