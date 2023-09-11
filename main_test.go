package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestPatch(*testing.T) {
	apiUrl := "http://192.168.0.6:8000/rest/v1/posts?id=eq.1"
	serviceRoleKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.ewogICAgInJvbGUiOiAic2VydmljZV9yb2xlIiwKICAgICJpc3MiOiAic3VwYWJhc2UiLAogICAgImlhdCI6IDE2NjYwMjI0MDAsCiAgICAiZXhwIjogMTgyMzc4ODgwMAp9.sbuBA2BnmzMP1CIMIyPWPEnAkGSnBUhFsOwcXEng5qg"

	// Create a JSON payload for the PATCH request
	jsonPayload := []byte(`{ "m24": "true", "d650": "24.67" }`)

	// Create a PATCH request
	req, err := http.NewRequest("PATCH", apiUrl, bytes.NewBuffer(jsonPayload))
	if err != nil {
		panic(err)
	}

	// Set request headers
	req.Header.Set("apikey", serviceRoleKey)
	req.Header.Set("Authorization", "Bearer "+serviceRoleKey)
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Prefer", "return=minimal")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Print the response body
	fmt.Println("Response Body:", string(body))

	if resp.StatusCode == http.StatusOK {
		// Request was successful
		// You can handle the response body here if needed
	} else {
		// Request failed
		// You can handle errors or response body here
		fmt.Println("Code:", resp.Status)
	}
}

func TestPost(*testing.T) {
	url := "http://192.168.0.6:8000/rest/v1/posts"
	serviceRoleKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.ewogICAgInJvbGUiOiAic2VydmljZV9yb2xlIiwKICAgICJpc3MiOiAic3VwYWJhc2UiLAogICAgImlhdCI6IDE2NjYwMjI0MDAsCiAgICAiZXhwIjogMTgyMzc4ODgwMAp9.sbuBA2BnmzMP1CIMIyPWPEnAkGSnBUhFsOwcXEng5qg"

	// JSON payload
	payload := []byte(`{"d650": "100"}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set headers
	req.Header.Set("apikey", serviceRoleKey)
	req.Header.Set("Authorization", "Bearer "+serviceRoleKey)
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Prefer", "return=minimal")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Print the response body
	fmt.Println("Response Body:", string(body))

	if resp.StatusCode == http.StatusOK {
		// Request was successful
		// You can handle the response body here if needed
	} else {
		// Request failed
		// You can handle errors or response body here
		fmt.Println("Failed:", resp.Status)
	}
}
