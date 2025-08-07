package httpclient

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

// Client represents the agnostic http client
type Client struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
}

// ClientOption allows for customization of the client
type ClientOption func(*Client)

// New creates a new agnostic HTTP client with the provided base URL, API key, and API secret.
// It accepts optional ClientOption functions to customize the client further.
// The base URL should be the root endpoint of the API, e.g., "https://api.example.com/v1".
// The API key and secret are used for authentication in requests.
// The client uses the default HTTP client from the net/http package, which can be customized with options.
// The client is designed to be used for making authenticated requests to an API that requires HMAC authentication.
func New(baseURL, apiKey, apiSecret string, options ...ClientOption) *Client {
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

// MakeRequest is the core function to make HTTP requests.
func (c *Client) MakeRequest(ctx context.Context, method, path string, body any) (*http.Response, error) {
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
