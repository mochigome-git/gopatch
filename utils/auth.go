package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func Login() {
	// Define the API endpoint and request body data.
	url := "http://192.168.0.6:8000/auth/v1/token?grant_type=password"
	serviceRoleKey := os.Getenv("SERVICE_ROLE_KEY")
	apiKey := serviceRoleKey
	requestData := map[string]string{
		"email":    "it@general-i.com.my",
		"password": "P@ssw0rd",
	}

	// Marshal the request data to JSON.
	requestDataJSON, err := json.Marshal(requestData)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	// Create a new HTTP request.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestDataJSON))
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return
	}

	// Set the necessary headers.
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Create an HTTP client and send the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending HTTP request:", err)
		return
	}
	defer resp.Body.Close()

	// Check the response status code.
	if resp.StatusCode != http.StatusOK {
		fmt.Println("HTTP request failed with status code:", resp.StatusCode)
		return
	}

	// Read and print the response body.
	var response map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&response); err != nil {
		fmt.Println("Error decoding JSON response:", err)
		return
	}

	fmt.Println("Response:", response)
}
