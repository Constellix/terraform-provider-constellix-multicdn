package httpclient

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewHTTPClient(t *testing.T) {
	baseURL := "https://api.example.com"
	apiKey := "test-key"
	apiSecret := "test-secret"

	// Test default client creation
	client := New(baseURL, apiKey, apiSecret)
	if client.baseURL != baseURL {
		t.Errorf("Expected baseURL %s, got %s", baseURL, client.baseURL)
	}
	if client.apiKey != apiKey {
		t.Errorf("Expected apiKey %s, got %s", apiKey, client.apiKey)
	}
	if client.apiSecret != apiSecret {
		t.Errorf("Expected apiSecret %s, got %s", apiSecret, client.apiSecret)
	}

	client = New(
		baseURL,
		apiKey,
		apiSecret,
	)
}

func TestGenerateAuthToken(t *testing.T) {
	client := New("https://api.example.com", "test-key", "test-secret")

	token, err := client.generateAuthToken()
	if err != nil {
		t.Fatalf("Error generating auth token: %v", err)
	}

	// Token should be in the format "apiKey:hmacHash:timestamp"
	parts := strings.Split(token, ":")
	if len(parts) != 3 {
		t.Fatalf("Token format is incorrect. Expected 3 parts separated by ':', got %d parts: %s", len(parts), token)
	}

	// First part should be the API key
	if parts[0] != "test-key" {
		t.Errorf("Expected first part of token to be 'test-key', got '%s'", parts[0])
	}

	// Third part should be a timestamp (numeric)
	timestamp := parts[2]
	_, err = time.Parse(time.RFC3339, timestamp)
	if err == nil {
		// This is not a Unix timestamp, which is expected
		t.Errorf("Expected third part of token to be a Unix timestamp, got a time in RFC3339 format: %s", timestamp)
	}

	// // Check that we can parse it as a number
	// _, err = time.UnixMilli(int64(len(timestamp)))
	// if err != nil && !strings.ContainsAny(timestamp, "0123456789") {
	// 	t.Errorf("Expected third part of token to contain numeric characters: %s", timestamp)
	// }
}

func TestMakeRequest(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name         string
		method       string
		path         string
		body         any
		responseCode int
		responseBody string
	}{
		{
			name:         "successful_get",
			method:       http.MethodGet,
			path:         "/preference/123",
			body:         nil,
			responseCode: http.StatusOK,
			responseBody: `{"resourceId": 123}`,
		},
		{
			name:         "successful_post",
			method:       http.MethodPost,
			path:         "/preference",
			body:         map[string]interface{}{"resourceId": 123},
			responseCode: http.StatusCreated,
			responseBody: ``,
		},
		{
			name:         "client_error",
			method:       http.MethodGet,
			path:         "/preference/999",
			body:         nil,
			responseCode: http.StatusNotFound,
			responseBody: `{"error": "Resource not found"}`,
		},
		{
			name:         "server_error",
			method:       http.MethodGet,
			path:         "/preference/500",
			body:         nil,
			responseCode: http.StatusInternalServerError,
			responseBody: `{"error": "Internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != tc.method {
					t.Errorf("Expected method %s, got %s", tc.method, r.Method)
				}

				// Verify request path
				if r.URL.Path != tc.path {
					t.Errorf("Expected path %s, got %s", tc.path, r.URL.Path)
				}

				// Verify auth header exists and value is correct
				authHeader := r.Header.Get("x-cns-security-token")
				if authHeader == "" {
					t.Error("Auth token header missing")
				} else {
					// Validate token format and value
					parts := strings.Split(authHeader, ":")
					if len(parts) != 3 {
						t.Errorf("Auth token format invalid: %s", authHeader)
					} else {
						expectedApiKey := "test-key"
						expectedApiSecret := "test-secret"
						timestamp := parts[2]
						expectedHmac := computeHMACTest(expectedApiSecret, timestamp)
						expectedToken := expectedApiKey + ":" + expectedHmac + ":" + timestamp
						if authHeader != expectedToken {
							t.Errorf("Auth token value incorrect.\nExpected: %s\nGot:      %s", expectedToken, authHeader)
						}
					}
				}

				// Verify content type for requests with body
				if tc.body != nil {
					contentType := r.Header.Get("Content-Type")
					if contentType != "application/json" {
						t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
					}

					// Check that the body was sent correctly
					var requestBody map[string]interface{}
					err := json.NewDecoder(r.Body).Decode(&requestBody)
					if err != nil {
						t.Errorf("Error decoding request body: %v", err)
					}
					if requestBody["resourceId"] != float64(123) { // JSON numbers are float64
						t.Errorf("Expected resourceId 123, got %v", requestBody["resourceId"])
					}
				}

				// Set response headers and status code
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.responseCode)
				_, err := w.Write([]byte(tc.responseBody))
				if err != nil {
					t.Errorf("Error writing response body: %v", err)
				}
			}))
			defer server.Close()

			// Create client pointing to test server
			client := New(server.URL, "test-key", "test-secret")

			// Make request
			ctx := context.Background()
			resp, err := client.MakeRequest(ctx, tc.method, tc.path, tc.body)
			if err != nil {
				t.Errorf("Error making request: %v", err)
			}
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					t.Errorf("Error closing response body: %v", err)
				}
			}(resp.Body)

			if resp.StatusCode != tc.responseCode {
				t.Errorf("Expected status code %d, got %d", tc.responseCode, resp.StatusCode)
			}

			// Read and verify response body
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("Error reading response body: %v", err)
			}

			if string(bodyBytes) != tc.responseBody {
				t.Errorf("Expected response body '%s', got '%s'", tc.responseBody, string(bodyBytes))
			}
		})
	}
}

func TestContextCancellation(t *testing.T) {
	// Create a test server with a delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep to simulate a long-running request
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"resourceId": 123}`))
		if err != nil {
			t.Errorf("Error writing response body: %v", err)
		}
	}))
	defer server.Close()

	// Create client
	client := New(server.URL, "test-key", "test-secret")

	// Create a context with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Make request
	resp, err := client.MakeRequest(ctx, http.MethodGet, "/preference/123", nil)

	// Request should be canceled by context timeout
	if err == nil {
		t.Error("Expected error due to context cancellation but got none")
	}
	if resp != nil {
		t.Error("Expected nil response after context cancellation")
	}
}

// computeHMACTest is copied from client.go for test verification
func computeHMACTest(secretKey, timestamp string) string {
	h := hmac.New(sha1.New, []byte(secretKey))
	h.Write([]byte(timestamp))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
