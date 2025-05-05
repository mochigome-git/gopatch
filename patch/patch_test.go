package patch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestSendPatchRequest(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody interface{}
		expectBody   interface{}
		expectError  bool
	}{
		{
			name:         "Success - 200 OK",
			statusCode:   http.StatusOK,
			responseBody: map[string]string{"message": "updated"},
			expectBody:   map[string]string{"message": "updated"},
			expectError:  false,
		},
		{
			name:         "Success - 204 No Content",
			statusCode:   http.StatusNoContent,
			responseBody: nil,
			expectBody:   nil,
			expectError:  false,
		},
		{
			name:         "Failure - 400 Bad Request",
			statusCode:   http.StatusBadRequest,
			responseBody: map[string]string{"error": "bad request"},
			expectBody:   map[string]string{"error": "bad request"},
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.responseBody != nil {
					json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			jsonPayload := []byte(`{"key":"value"}`)
			body, err := SendPatchRequest(server.URL, "dummy-key", jsonPayload, "PATCH")

			if (err != nil) != tt.expectError {
				t.Fatalf("Expected error: %v, got: %v", tt.expectError, err)
			}

			// Compare body if expected body is not nil
			if tt.expectBody != nil {
				var actual map[string]interface{}
				var expected map[string]interface{}

				if err := json.Unmarshal(body, &actual); err != nil {
					t.Fatalf("Failed to unmarshal actual body: %v", err)
				}
				if b, err := json.Marshal(tt.expectBody); err == nil {
					_ = json.Unmarshal(b, &expected)
				}

				if !reflect.DeepEqual(actual, expected) {
					t.Errorf("Expected body: %+v, got: %+v", expected, actual)
				}
			} else if body != nil {
				t.Errorf("Expected nil body, got: %s", string(body))
			}
		})
	}
}
