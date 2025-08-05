// Package cdnclient provides a client for the CDN Configuration API
package cdnclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/constellix/terraform-provider-multicdn/clients/httpclient"
	"github.com/constellix/terraform-provider-multicdn/clients/httpclient/response"
)

// Client represents the CDN Configuration API client
type Client struct {
	*httpclient.Client // HTTP client for making requests
}

// New creates a new CDN Preference API client
func New(httpclient *httpclient.Client) *Client {
	client := &Client{
		Client: httpclient,
	}

	return client
}

// GetCdnConfigsPage retrieves CDN configurations for the authenticated account with pagination
func (c *Client) GetCdnConfigsPage(ctx context.Context, pageNumber, pageSize int) (*CdnConfigurationPage, error) {
	path := fmt.Sprintf("/cdn-configs?page=%d&size=%d", pageNumber, pageSize)
	resp, err := c.MakeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var configPage CdnConfigurationPage
	if err := response.Parse(resp, &configPage); err != nil {
		return nil, err
	}

	return &configPage, nil
}

// CreateCdnConfig creates a new CDN configuration
func (c *Client) CreateCdnConfig(ctx context.Context, config *CdnConfiguration) (*CdnConfigurationResponse, error) {
	resp, err := c.MakeRequest(ctx, http.MethodPost, "/cdn-configs", config)
	if err != nil {
		return nil, err
	}

	var createdConfig CdnConfigurationResponse
	if err := response.Parse(resp, &createdConfig); err != nil {
		return nil, err
	}

	return &createdConfig, nil
}

// GetCdnConfig gets a specific CDN configuration by resourceId
func (c *Client) GetCdnConfig(ctx context.Context, resourceID int) (*CdnConfigurationResponse, error) {
	path := fmt.Sprintf("/cdn-configs/%d", resourceID)
	resp, err := c.MakeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var config CdnConfigurationResponse
	if err := response.Parse(resp, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// UpdateCdnConfig updates an existing CDN configuration by resourceId
func (c *Client) UpdateCdnConfig(ctx context.Context, resourceID int, config *CdnConfiguration) (*CdnConfigurationResponse, error) {
	path := fmt.Sprintf("/cdn-configs/%d", resourceID)
	resp, err := c.MakeRequest(ctx, http.MethodPut, path, config)
	if err != nil {
		return nil, err
	}

	var updatedConfig CdnConfigurationResponse
	if err := response.Parse(resp, &updatedConfig); err != nil {
		return nil, err
	}

	return &updatedConfig, nil
}

// DeleteCdnConfig deletes a CDN configuration by resourceId
func (c *Client) DeleteCdnConfig(ctx context.Context, resourceID int) error {
	path := fmt.Sprintf("/cdn-configs/%d", resourceID)
	resp, err := c.MakeRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return response.Parse(resp, nil)
}

// GetCdnEntries gets the CDN provider registry for a specific resourceId
func (c *Client) GetCdnEntries(ctx context.Context, resourceID int) ([]CdnEntry, error) {
	path := fmt.Sprintf("/cdn-configs/%d/cdns", resourceID)
	resp, err := c.MakeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var entries []CdnEntry
	if err := response.Parse(resp, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

// GetCdnEnablementMap gets the CDN enablement map for a specific resourceId
func (c *Client) GetCdnEnablementMap(ctx context.Context, resourceID int) (*CdnEnablementMap, error) {
	path := fmt.Sprintf("/cdn-configs/%d/enablement", resourceID)
	resp, err := c.MakeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var enablementMap CdnEnablementMap
	if err := response.Parse(resp, &enablementMap); err != nil {
		return nil, err
	}

	return &enablementMap, nil
}

// GetTrafficDistribution gets the traffic distribution rules for a specific resourceId
func (c *Client) GetTrafficDistribution(ctx context.Context, resourceID int) (*TrafficDistribution, error) {
	path := fmt.Sprintf("/cdn-configs/%d/trafficDistribution", resourceID)
	resp, err := c.MakeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var distribution TrafficDistribution
	if err := response.Parse(resp, &distribution); err != nil {
		return nil, err
	}

	return &distribution, nil
}
