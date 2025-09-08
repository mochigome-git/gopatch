package patch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gopatch/config"
	"gopatch/internal/app"
	"io/ioutil"
	"net/http"
	"strings"
)

type VacuumData struct {
	ID              string   `json:"id"`
	CreatedAt       string   `json:"created_at"`
	VacuumStart     int      `json:"vacuum_start"`
	VacuumLeave1min float64  `json:"vacuum_leave_1min"`
	VacuumLeave2min float64  `json:"vacuum_leave_2min"`
	VacuumLeave3min float64  `json:"vacuum_leave_3min"`
	XStatus         string   `json:"x_status"`
	YStatus         string   `json:"y_status"`
	VacuumStatus    bool     `json:"vacuum_status"`
	X               *float64 `json:"x"`
	Y               *float64 `json:"y"`
}

func SendUpsertRequest(apiUrl, serviceRoleKey string, jsonPayload []byte, cfg config.AppConfig, plcApp *app.Application) ([]byte, error) {
	// Create a PATCH request
	req, err := http.NewRequest(cfg.Function, apiUrl, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set request headers
	req.Header.Set("apikey", serviceRoleKey)
	req.Header.Set("Authorization", "Bearer "+serviceRoleKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")
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

	var result []VacuumData

	// Try to unmarshal as array first
	if err := json.Unmarshal(body, &result); err != nil {
		// If it fails, try single object
		var single VacuumData
		if err2 := json.Unmarshal(body, &single); err2 != nil {
			return nil, fmt.Errorf("failed to parse response JSON: %v", err)
		}
		result = append(result, single)
	}

	if len(result) > 0 && plcApp != nil {
		devicesStr := strings.Split(cfg.Plc.PlcDeviceUpsert, ",")
		if len(devicesStr)%4 != 0 {
			return nil, fmt.Errorf("invalid device config string")
		}

		dataList := []any{result[0].YStatus, result[0].XStatus, result[0].VacuumStatus}
		deviceCount := len(devicesStr) / 4
		if deviceCount != len(dataList) {
			return nil, fmt.Errorf("mismatch between device count and data count")
		}

		for i := 0; i < deviceCount; i++ {
			// Compose full device string: "Type,Number,ProcessNumber,Registers"
			deviceStr := strings.Join(devicesStr[i*4:i*4+4], ",")

			if err := plcApp.WritePLC(context.Background(), deviceStr, dataList[i]); err != nil {
				fmt.Printf("PLC write failed for device %s: %v\n", deviceStr, err)
				return nil, err
			}
		}
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
