package response_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/constellix/terraform-provider-constellix-multicdn/clients/httpclient/response"
	"github.com/constellix/terraform-provider-constellix-multicdn/clients/preferenceclient"
)

func TestParseResponse(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		targetStruct interface{}
		expectError  bool
	}{
		{
			name:         "successful_parse_preference",
			statusCode:   http.StatusOK,
			responseBody: `{"resourceId": 123, "contentType": "application/json", "description": "Test"}`,
			targetStruct: &preferenceclient.Preference{},
			expectError:  false,
		},
		{
			name:         "successful_parse_thresholds",
			statusCode:   http.StatusOK,
			responseBody: `{"world": 95, "continents": {"NA": {"default": 98}}}`,
			targetStruct: &preferenceclient.AvailabilityThresholds{},
			expectError:  false,
		},
		{
			name:         "client_error",
			statusCode:   http.StatusBadRequest,
			responseBody: `{"error": "Bad request"}`,
			targetStruct: &preferenceclient.Preference{},
			expectError:  true,
		},
		{
			name:         "invalid_json",
			statusCode:   http.StatusOK,
			responseBody: `{invalid json`,
			targetStruct: &preferenceclient.Preference{},
			expectError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test response
			resp := &http.Response{
				StatusCode: tc.statusCode,
				Body:       io.NopCloser(strings.NewReader(tc.responseBody)),
				Header:     make(http.Header),
			}
			resp.Header.Set("Content-Type", "application/json")

			// Parse the response
			err := response.Parse(resp, tc.targetStruct)

			// Check error
			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// For successful parses, check struct contents
			if !tc.expectError {
				switch v := tc.targetStruct.(type) {
				case *preferenceclient.Preference:
					if v.ResourceID != 123 {
						t.Errorf("Expected ResourceID 123, got %d", v.ResourceID)
					}
					if v.ContentType != "application/json" {
						t.Errorf("Expected ContentType 'application/json', got %s", v.ContentType)
					}
					if v.Description != "Test" {
						t.Errorf("Expected Description 'Test', got %s", v.Description)
					}
				case *preferenceclient.AvailabilityThresholds:
					if v.World != 95 {
						t.Errorf("Expected World 95, got %d", v.World)
					}
					if continent, ok := v.Continents["NA"]; ok {
						if continent.Default != 98 {
							t.Errorf("Expected NA default 98, got %d", continent.Default)
						}
					} else {
						t.Error("Expected continent 'NA' but not found")
					}
				}
			}
		})
	}
}
