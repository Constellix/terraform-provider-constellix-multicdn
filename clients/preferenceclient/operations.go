package preferenceclient

import (
	"context"
	"fmt"
	"net/http"
)

// GetPreferencesPage retrieves all CDN configurations for the authenticated account
func (c *Client) GetPreferencesPage(ctx context.Context) ([]PreferencePage, error) {
	resp, err := c.makeRequest(ctx, http.MethodGet, "/preference", nil)
	if err != nil {
		return nil, err
	}

	var preferences []PreferencePage
	if err := parseResponse(resp, &preferences); err != nil {
		return nil, err
	}

	return preferences, nil
}

// CreatePreference creates a new configuration preference
func (c *Client) CreatePreference(ctx context.Context, preference *Preference) error {
	resp, err := c.makeRequest(ctx, http.MethodPost, "/preference", preference)
	if err != nil {
		return err
	}

	return parseResponse(resp, nil)
}

// GetPreference gets a specific CDN configuration by resourceId
func (c *Client) GetPreference(ctx context.Context, resourceID int) (*Preference, error) {
	path := fmt.Sprintf("/preference/%d", resourceID)
	resp, err := c.makeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var preference Preference
	if err := parseResponse(resp, &preference); err != nil {
		return nil, err
	}

	return &preference, nil
}

// UpdatePreference updates an existing configuration by resourceId
func (c *Client) UpdatePreference(ctx context.Context, resourceID int, preference *Preference) error {
	path := fmt.Sprintf("/preference/%d", resourceID)
	resp, err := c.makeRequest(ctx, http.MethodPut, path, preference)
	if err != nil {
		return err
	}

	return parseResponse(resp, nil)
}

// DeletePreference deletes a configuration by resourceId
func (c *Client) DeletePreference(ctx context.Context, resourceID int) error {
	path := fmt.Sprintf("/preference/%d", resourceID)
	resp, err := c.makeRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return parseResponse(resp, nil)
}

// GetAvailabilityThresholds gets availability thresholds for a specific resourceId
func (c *Client) GetAvailabilityThresholds(ctx context.Context, resourceID int) (*AvailabilityThresholds, error) {
	path := fmt.Sprintf("/preference/%d/availabilityThresholds", resourceID)
	resp, err := c.makeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var thresholds AvailabilityThresholds
	if err := parseResponse(resp, &thresholds); err != nil {
		return nil, err
	}

	return &thresholds, nil
}

// GetPerformanceFiltering gets performance filtering config for a specific resourceId
func (c *Client) GetPerformanceFiltering(ctx context.Context, resourceID int) (*PerformanceFiltering, error) {
	path := fmt.Sprintf("/preference/%d/performanceFiltering", resourceID)
	resp, err := c.makeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var filtering PerformanceFiltering
	if err := parseResponse(resp, &filtering); err != nil {
		return nil, err
	}

	return &filtering, nil
}

// GetEnabledSubdivisionCountries gets enabled subdivisions countries for a specific resourceId
func (c *Client) GetEnabledSubdivisionCountries(ctx context.Context, resourceID int) (*EnabledSubdivisionCountries, error) {
	path := fmt.Sprintf("/preference/%d/enabledSubdivisionCountries", resourceID)
	resp, err := c.makeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var countries EnabledSubdivisionCountries
	if err := parseResponse(resp, &countries); err != nil {
		return nil, err
	}

	return &countries, nil
}
