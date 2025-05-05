package patch

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func SendPatchRequest(apiUrl, serviceRoleKey string, jsonPayload []byte, function string) ([]byte, error) {
	// Create a PATCH request
	req, err := http.NewRequest(function, apiUrl, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set request headers
	req.Header.Set("apikey", serviceRoleKey)
	req.Header.Set("Authorization", "Bearer "+serviceRoleKey)
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Prefer", "return=minimal")

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
	switch resp.StatusCode {
	case http.StatusOK:
		return body, nil
	case http.StatusNoContent:
		return nil, nil
	case http.StatusCreated:
		return nil, nil
	default:
		return body, fmt.Errorf("request failed with status code: %d - Response: %s", resp.StatusCode, string(body))
	}
}
