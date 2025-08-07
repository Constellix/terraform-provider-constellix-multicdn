package preferenceclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/constellix/terraform-provider-multicdn/clients/httpclient"
	"github.com/constellix/terraform-provider-multicdn/clients/httpclient/response"
)

// Client represents the CDN Preference API client
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

// GetPreferencesPage retrieves CDN configurations for the authenticated account with pagination
func (c *Client) GetPreferencesPage(ctx context.Context, pageNumber, pageSize int) ([]PreferencePage, error) {
	path := fmt.Sprintf("/preference?page=%d&size=%d", pageNumber, pageSize)
	resp, err := c.MakeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var preferences []PreferencePage
	if err := response.Parse(resp, &preferences); err != nil {
		return nil, err
	}

	return preferences, nil
}

// CreatePreference creates a new configuration preference
func (c *Client) CreatePreference(ctx context.Context, preference *Preference) error {
	resp, err := c.MakeRequest(ctx, http.MethodPost, "/preference", preference)
	if err != nil {
		return err
	}

	return response.Parse(resp, nil)
}

// GetPreference gets a specific CDN configuration by resourceId
func (c *Client) GetPreference(ctx context.Context, resourceID int64) (*Preference, error) {
	path := fmt.Sprintf("/preference/%d", resourceID)
	resp, err := c.MakeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var preference Preference
	if err := response.Parse(resp, &preference); err != nil {
		return nil, err
	}

	return &preference, nil
}

// UpdatePreference updates an existing configuration by resourceId
func (c *Client) UpdatePreference(ctx context.Context, resourceID int64, preference *Preference) error {
	path := fmt.Sprintf("/preference/%d", resourceID)
	resp, err := c.MakeRequest(ctx, http.MethodPut, path, preference)
	if err != nil {
		return err
	}

	return response.Parse(resp, nil)
}

// DeletePreference deletes a configuration by resourceId
func (c *Client) DeletePreference(ctx context.Context, resourceID int64) error {
	path := fmt.Sprintf("/preference/%d", resourceID)
	resp, err := c.MakeRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return response.Parse(resp, nil)
}

// GetAvailabilityThresholds gets availability thresholds for a specific resourceId
func (c *Client) GetAvailabilityThresholds(ctx context.Context, resourceID int64) (*AvailabilityThresholds, error) {
	path := fmt.Sprintf("/preference/%d/availabilityThresholds", resourceID)
	resp, err := c.MakeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var thresholds AvailabilityThresholds
	if err := response.Parse(resp, &thresholds); err != nil {
		return nil, err
	}

	return &thresholds, nil
}

// GetPerformanceFiltering gets performance filtering config for a specific resourceId
func (c *Client) GetPerformanceFiltering(ctx context.Context, resourceID int64) (*PerformanceFiltering, error) {
	path := fmt.Sprintf("/preference/%d/performanceFiltering", resourceID)
	resp, err := c.MakeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var filtering PerformanceFiltering
	if err := response.Parse(resp, &filtering); err != nil {
		return nil, err
	}

	return &filtering, nil
}

// GetEnabledSubdivisionCountries gets enabled subdivisions countries for a specific resourceId
func (c *Client) GetEnabledSubdivisionCountries(ctx context.Context, resourceID int64) (*EnabledSubdivisionCountries, error) {
	path := fmt.Sprintf("/preference/%d/enabledSubdivisionCountries", resourceID)
	resp, err := c.MakeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var countries EnabledSubdivisionCountries
	if err := response.Parse(resp, &countries); err != nil {
		return nil, err
	}

	return &countries, nil
}
