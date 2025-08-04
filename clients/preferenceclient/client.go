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

const (
	defaultRetryCount   = 3
	defaultRetryBackoff = 500 * time.Millisecond
)

// Client represents the CDN Preference API client
type Client struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
	retryCount int
	retryDelay time.Duration
}

// ClientOption allows for customization of the client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithRetryPolicy customizes the retry policy
func WithRetryPolicy(count int, delay time.Duration) ClientOption {
	return func(c *Client) {
		c.retryCount = count
		c.retryDelay = delay
	}
}

// NewClient creates a new CDN Preference API client
func NewClient(baseURL, apiKey, apiSecret string, options ...ClientOption) *Client {
	client := &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		httpClient: http.DefaultClient,
		retryCount: defaultRetryCount,
		retryDelay: defaultRetryBackoff,
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

// makeRequest is the core function to make HTTP requests with retry logic
func (c *Client) makeRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
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
	req.Header.Set("Authorization", "Bearer "+token)

	// Implement retry logic
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= c.retryCount; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.retryDelay * time.Duration(attempt)):
				// Exponential backoff
			}
		}

		resp, lastErr = c.httpClient.Do(req)
		if lastErr != nil {
			continue
		}

		// Retry on rate limiting (429) errors
		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		// Close response body on retry
		err := resp.Body.Close()
		if err != nil {
			return nil, err
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d attempts: %w", c.retryCount+1, lastErr)
	}

	return resp, fmt.Errorf("request failed after %d attempts with status code: %d", c.retryCount+1, resp.StatusCode)
}

// parseResponse handles deserializing the response body
func parseResponse(resp *http.Response, v interface{}) error {
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
