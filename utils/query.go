package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func sendPatchRequest(apiUrl, serviceRoleKey string, jsonPayload []byte, function string) ([]byte, error) {
	// Create a PATCH request
	req, err := http.NewRequest(function, apiUrl, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set request headers
	req.Header.Set("apikey", serviceRoleKey)
	req.Header.Set("Authorization", "Bearer "+serviceRoleKey)
	req.Header.Set("Content-Type", "application/json")

	// Reuse an HTTP client
	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Check the HTTP status code
	if resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	return body, nil
}
