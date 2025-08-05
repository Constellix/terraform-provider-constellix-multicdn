package preferenceclient

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents the CDN Preference API client
type Client struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
}

// ClientOption allows for customization of the client
type ClientOption func(*Client)

// NewClient creates a new CDN Preference API client
func NewClient(baseURL, apiKey, apiSecret string, options ...ClientOption) *Client {
	client := &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		httpClient: http.DefaultClient,
	}

	// Apply options
	for _, option := range options {
		option(client)
	}

	return client
}

// generateAuthToken generates an HMAC token using the API key and secret
func (c *Client) generateAuthToken() (string, error) {
	timestamp := fmt.Sprintf("%d", time.Now().UTC().UnixMilli())
	hmacHash := computeHMAC(c.apiSecret, timestamp)
	return fmt.Sprintf("%s:%s:%s", c.apiKey, hmacHash, timestamp), nil
}

// computeHMAC generates base64-encoded HMAC-SHA1 digest
func computeHMAC(secretKey, timestamp string) string {
	h := hmac.New(sha1.New, []byte(secretKey))
	h.Write([]byte(timestamp))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// makeRequest is the core function to make HTTP requests.
func (c *Client) makeRequest(ctx context.Context, method, path string, body any) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Generate and add authentication token
	token, err := c.generateAuthToken()
	if err != nil {
		return nil, fmt.Errorf("error generating auth token: %w", err)
	}
	req.Header.Set("x-cns-security-token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	return resp, nil
}

// parseResponse handles deserializing the response body
func parseResponse(resp *http.Response, v any) error {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("error closing response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode >= 400 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading error response body: %w", err)
		}
		return fmt.Errorf("API error: status code %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	if v == nil {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("error decoding response body: %w", err)
	}

	return nil
}
