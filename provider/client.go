package provider

import (
	"github.com/constellix/terraform-provider-multicdn/clients/preferenceclient"
)

// APIClient wraps the preference client
type APIClient struct {
	preference *preferenceclient.Client
}

// NewAPIClient creates a new API client for the provider
func NewAPIClient(baseURL, apiKey, apiSecret string) *APIClient {
	return &APIClient{
		preference: preferenceclient.NewClient(baseURL, apiKey, apiSecret),
	}
}
