package provider

import (
	"github.com/constellix/terraform-provider-constellix-multicdn/clients/cdnclient"
	"github.com/constellix/terraform-provider-constellix-multicdn/clients/httpclient"
	"github.com/constellix/terraform-provider-constellix-multicdn/clients/preferenceclient"
)

// APIClient wraps the preference client
type APIClient struct {
	preference *preferenceclient.Client
	cdn        *cdnclient.Client
}

// NewAPIClient creates a new API client for the provider
func NewAPIClient(baseURL, apiKey, apiSecret string) *APIClient {
	httpClient := httpclient.New(baseURL, apiKey, apiSecret)
	return &APIClient{
		preference: preferenceclient.New(httpClient),
		cdn:        cdnclient.New(httpClient),
	}
}
