package preferenceclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/constellix/terraform-provider-constellix-multicdn/clients/httpclient"
)

func TestGetPreferencesPage(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name       string
		pageNumber int
		pageSize   int
		response   string
		statusCode int
		expectErr  bool
	}{
		{
			name:       "successful_response",
			pageNumber: 0,
			pageSize:   10,
			statusCode: http.StatusOK,
			response: `[{
				"preferenceConfigs": [
					{
						"resourceId": 123,
						"contentType": "application/json",
						"description": "Test preference",
						"version": "1.0",
						"lastUpdated": "2023-01-01T00:00:00Z",
						"availabilityThresholds": {
							"world": 95,
							"continents": {
								"NA": {
									"default": 98,
									"countries": {
										"US": 99
									}
								}
							}
						},
						"performanceFiltering": {
							"world": {
								"mode": "relative",
								"relativeThreshold": 0.2
							}
						},
						"enabledSubdivisionCountries": {
							"continents": {
								"NA": {
									"countries": ["US", "CA"]
								}
							}
						}
					}
				],
				"totalElements": 1,
				"totalPages": 1,
				"pageNumber": 0,
				"pageSize": 10,
				"numberOfElements": 1,
				"first": true,
				"last": true,
				"empty": false
			}]`,
			expectErr: false,
		},
		{
			name:       "error_response",
			pageNumber: 0,
			pageSize:   10,
			statusCode: http.StatusInternalServerError,
			response:   `{"error": "Internal server error"}`,
			expectErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check request parameters
				if r.URL.Path != "/preference" {
					t.Errorf("Expected path /preference, got %s", r.URL.Path)
				}

				// Check query parameters
				q := r.URL.Query()
				if q.Get("page") != strconv.Itoa(tc.pageNumber) {
					t.Errorf("Expected page %d, got %s", tc.pageNumber, q.Get("page"))
				}
				if q.Get("size") != strconv.Itoa(tc.pageSize) {
					t.Errorf("Expected size %d, got %s", tc.pageSize, q.Get("size"))
				}

				// Return response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.response))
			}))
			defer server.Close()

			// Create client pointing to test server
			client := givenPreferenceClient(server.URL)

			// Call the method
			ctx := context.Background()
			preferences, err := client.GetPreferencesPage(ctx, tc.pageNumber, tc.pageSize)

			// Validate results
			if tc.expectErr && err == nil {
				t.Error("Expected error but got none")
			} else if !tc.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tc.expectErr && err == nil {
				// Validate response data
				if len(preferences) == 0 {
					t.Error("Expected preferences but got none")
				} else {
					pref := preferences[0]
					if pref.TotalElements != 1 {
						t.Errorf("Expected 1 total element, got %d", pref.TotalElements)
					}
					if len(pref.PreferenceConfigs) != 1 {
						t.Errorf("Expected 1 preference config, got %d", len(pref.PreferenceConfigs))
					}
					if pref.PreferenceConfigs[0].ResourceID != 123 {
						t.Errorf("Expected resource ID 123, got %d", pref.PreferenceConfigs[0].ResourceID)
					}
				}
			}
		})
	}
}

func TestCreatePreference(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name       string
		preference Preference
		statusCode int
		expectErr  bool
	}{
		{
			name: "successful_create",
			preference: Preference{
				ResourceID:  123,
				ContentType: "application/json",
				Description: "Test preference",
				AvailabilityThresholds: AvailabilityThresholds{
					World: 95,
					Continents: map[string]ContinentThreshold{
						"NA": {
							Default: 98,
							Countries: map[string]int64{
								"US": 99,
							},
						},
					},
				},
				PerformanceFiltering: PerformanceFiltering{
					World: PerformanceConfig{
						Mode:              "relative",
						RelativeThreshold: toFloat64Pointer(0.2),
					},
				},
				EnabledSubdivisionCountries: EnabledSubdivisionCountries{
					Continents: map[string]ContinentSubdivisions{
						"NA": {
							Countries: []string{"US", "CA"},
						},
					},
				},
			},
			statusCode: http.StatusCreated,
			expectErr:  false,
		},
		{
			name: "error_create",
			preference: Preference{
				ResourceID: 123,
			},
			statusCode: http.StatusBadRequest,
			expectErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check request method and path
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST method, got %s", r.Method)
				}
				if r.URL.Path != "/preference" {
					t.Errorf("Expected path /preference, got %s", r.URL.Path)
				}

				// Decode request body
				var req Preference
				decoder := json.NewDecoder(r.Body)
				if err := decoder.Decode(&req); err != nil {
					t.Errorf("Failed to decode request body: %v", err)
				}

				// Validate request body
				if req.ResourceID != tc.preference.ResourceID {
					t.Errorf("Expected resource ID %d, got %d", tc.preference.ResourceID, req.ResourceID)
				}

				// Return response
				w.WriteHeader(tc.statusCode)
			}))
			defer server.Close()

			// Create client pointing to test server
			client := givenPreferenceClient(server.URL)

			// Call the method
			ctx := context.Background()
			err := client.CreatePreference(ctx, &tc.preference)

			// Validate results
			if tc.expectErr && err == nil {
				t.Error("Expected error but got none")
			} else if !tc.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func toFloat64Pointer(f float64) *float64 {
	return &f
}

func givenPreferenceClient(serverURL string) *Client {
	return New(httpclient.New(serverURL, "test-key", "test-secret"))
}

func TestGetPreference(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name       string
		resourceID int64
		response   string
		statusCode int
		expectErr  bool
	}{
		{
			name:       "successful_get",
			resourceID: 123,
			statusCode: http.StatusOK,
			response: `{
				"resourceId": 123,
				"contentType": "application/json",
				"description": "Test preference",
				"version": "1.0",
				"lastUpdated": "2023-01-01T00:00:00Z",
				"availabilityThresholds": {
					"world": 95,
					"continents": {
						"NA": {
							"default": 98,
							"countries": {
								"US": 99
							}
						}
					}
				},
				"performanceFiltering": {
					"world": {
						"mode": "relative",
						"relativeThreshold": 0.2
					}
				},
				"enabledSubdivisionCountries": {
					"continents": {
						"NA": {
							"countries": ["US", "CA"]
						}
					}
				}
			}`,
			expectErr: false,
		},
		{
			name:       "not_found",
			resourceID: 999,
			statusCode: http.StatusNotFound,
			response:   `{"error": "Resource not found"}`,
			expectErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check request method and path
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET method, got %s", r.Method)
				}
				expectedPath := "/preference/" + strconv.Itoa(int(tc.resourceID))
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Return response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.response))
			}))
			defer server.Close()

			// Create client pointing to test server
			client := givenPreferenceClient(server.URL)

			// Call the method
			ctx := context.Background()
			preference, err := client.GetPreference(ctx, tc.resourceID)

			// Validate results
			if tc.expectErr && err == nil {
				t.Error("Expected error but got none")
			} else if !tc.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tc.expectErr && preference != nil {
				// Validate response data
				if preference.ResourceID != tc.resourceID {
					t.Errorf("Expected resource ID %d, got %d", tc.resourceID, preference.ResourceID)
				}
				if preference.ContentType != "application/json" {
					t.Errorf("Expected content type application/json, got %s", preference.ContentType)
				}
			}
		})
	}
}

func TestUpdatePreference(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name       string
		resourceID int64
		preference Preference
		statusCode int
		expectErr  bool
	}{
		{
			name:       "successful_update",
			resourceID: 123,
			preference: Preference{
				ResourceID:  123,
				ContentType: "application/json",
				Description: "Updated preference",
				AvailabilityThresholds: AvailabilityThresholds{
					World: 96,
				},
				PerformanceFiltering: PerformanceFiltering{
					World: PerformanceConfig{
						Mode:              "relative",
						RelativeThreshold: toFloat64Pointer(1.3),
					},
				},
				EnabledSubdivisionCountries: EnabledSubdivisionCountries{},
			},
			statusCode: http.StatusOK,
			expectErr:  false,
		},
		{
			name:       "error_update",
			resourceID: 999,
			preference: Preference{
				ResourceID: 999,
			},
			statusCode: http.StatusNotFound,
			expectErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check request method and path
				if r.Method != http.MethodPut {
					t.Errorf("Expected PUT method, got %s", r.Method)
				}
				expectedPath := "/preference/" + strconv.Itoa(int(tc.resourceID))
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Decode request body
				var req Preference
				decoder := json.NewDecoder(r.Body)
				if err := decoder.Decode(&req); err != nil {
					t.Errorf("Failed to decode request body: %v", err)
				}

				// Validate request body
				if req.ResourceID != tc.preference.ResourceID {
					t.Errorf("Expected resource ID %d, got %d", tc.preference.ResourceID, req.ResourceID)
				}

				// Return response
				w.WriteHeader(tc.statusCode)
			}))
			defer server.Close()

			// Create client pointing to test server
			client := givenPreferenceClient(server.URL)

			// Call the method
			ctx := context.Background()
			err := client.UpdatePreference(ctx, tc.resourceID, &tc.preference)

			// Validate results
			if tc.expectErr && err == nil {
				t.Error("Expected error but got none")
			} else if !tc.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestDeletePreference(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name       string
		resourceID int64
		statusCode int
		expectErr  bool
	}{
		{
			name:       "successful_delete",
			resourceID: 123,
			statusCode: http.StatusNoContent,
			expectErr:  false,
		},
		{
			name:       "not_found",
			resourceID: 999,
			statusCode: http.StatusNotFound,
			expectErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check request method and path
				if r.Method != http.MethodDelete {
					t.Errorf("Expected DELETE method, got %s", r.Method)
				}
				expectedPath := "/preference/" + strconv.Itoa(int(tc.resourceID))
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Return response
				w.WriteHeader(tc.statusCode)
			}))
			defer server.Close()

			// Create client pointing to test server
			client := givenPreferenceClient(server.URL)

			// Call the method
			ctx := context.Background()
			err := client.DeletePreference(ctx, tc.resourceID)

			// Validate results
			if tc.expectErr && err == nil {
				t.Error("Expected error but got none")
			} else if !tc.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestGetAvailabilityThresholds(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name       string
		resourceID int64
		response   string
		statusCode int
		expectErr  bool
	}{
		{
			name:       "successful_get",
			resourceID: 123,
			statusCode: http.StatusOK,
			response: `{
				"world": 95,
				"continents": {
					"NA": {
						"default": 98,
						"countries": {
							"US": 99,
							"CA": 97
						}
					}
				}
			}`,
			expectErr: false,
		},
		{
			name:       "not_found",
			resourceID: 999,
			statusCode: http.StatusNotFound,
			response:   `{"error": "Resource not found"}`,
			expectErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check request method and path
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET method, got %s", r.Method)
				}
				expectedPath := "/preference/" + strconv.Itoa(int(tc.resourceID)) + "/availabilityThresholds"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Return response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.response))
			}))
			defer server.Close()

			// Create client pointing to test server
			client := givenPreferenceClient(server.URL)

			// Call the method
			ctx := context.Background()
			thresholds, err := client.GetAvailabilityThresholds(ctx, tc.resourceID)

			// Validate results
			if tc.expectErr && err == nil {
				t.Error("Expected error but got none")
			} else if !tc.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tc.expectErr && thresholds != nil {
				// Validate response data
				if thresholds.World != 95.0 {
					t.Errorf("Expected world threshold 95, got %d", thresholds.World)
				}
				if continent, ok := thresholds.Continents["NA"]; ok {
					if continent.Default != 98.0 {
						t.Errorf("Expected NA default threshold 98, got %d", continent.Default)
					}
					if country, ok := continent.Countries["US"]; ok {
						if country != 99 {
							t.Errorf("Expected US threshold 99, got %d", country)
						}
					} else {
						t.Error("Expected US country threshold but found none")
					}
				} else {
					t.Error("Expected NA continent threshold but found none")
				}
			}
		})
	}
}

func TestGetPerformanceFiltering(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name       string
		resourceID int64
		response   string
		statusCode int
		expectErr  bool
	}{
		{
			name:       "successful_get",
			resourceID: 123,
			statusCode: http.StatusOK,
			response: `{
				"world": {
					"mode": "relative",
					"relativeThreshold": 0.2
				},
				"continents": {
					"EU": {
						"mode": "relative",
						"relativeThreshold": 0.11,
						"countries": {
							"DE": {
								"mode": "relative",
								"relativeThreshold": 0.0
							}
						}
					}
				}
			}`,
			expectErr: false,
		},
		{
			name:       "not_found",
			resourceID: 999,
			statusCode: http.StatusNotFound,
			response:   `{"error": "Resource not found"}`,
			expectErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check request method and path
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET method, got %s", r.Method)
				}
				expectedPath := "/preference/" + strconv.Itoa(int(tc.resourceID)) + "/performanceFiltering"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Return response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.response))
			}))
			defer server.Close()

			// Create client pointing to test server
			client := givenPreferenceClient(server.URL)

			// Call the method
			ctx := context.Background()
			filtering, err := client.GetPerformanceFiltering(ctx, tc.resourceID)

			// Validate results
			if tc.expectErr && err == nil {
				t.Error("Expected error but got none")
			} else if !tc.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tc.expectErr && filtering != nil {
				// Validate response data
				if filtering.World.Mode != "relative" {
					t.Errorf("Expected world mode 'relative', got %s", filtering.World.Mode)
				}
				if filtering.World.RelativeThreshold == nil {
					t.Error("Expected world relative threshold but found nil")
				}
				if *filtering.World.RelativeThreshold != 0.2 {
					t.Errorf("Expected world threshold 0.2, got %f", *filtering.World.RelativeThreshold)
				}
				if continent, ok := filtering.Continents["EU"]; ok {
					if continent.Mode != "relative" {
						t.Errorf("Expected EU mode 'relative', got %s", continent.Mode)
					}
					if country, ok := continent.Countries["DE"]; ok {
						if country.RelativeThreshold == nil {
							t.Error("Expected DE country configuration but found nil")
						}
						if *country.RelativeThreshold != 0.0 {
							t.Errorf("Expected DE threshold 0.0, got %.1f", *country.RelativeThreshold)
						}
					} else {
						t.Error("Expected DE country configuration but found none")
					}
				} else {
					t.Error("Expected EU continent configuration but found none")
				}
			}
		})
	}
}

func TestGetEnabledSubdivisionCountries(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name       string
		resourceID int64
		response   string
		statusCode int
		expectErr  bool
	}{
		{
			name:       "successful_get",
			resourceID: 123,
			statusCode: http.StatusOK,
			response: `{
				"continents": {
					"NA": {
						"countries": ["US", "CA"]
					},
					"EU": {
						"countries": ["DE", "FR"]
					}
				}
			}`,
			expectErr: false,
		},
		{
			name:       "not_found",
			resourceID: 999,
			statusCode: http.StatusNotFound,
			response:   `{"error": "Resource not found"}`,
			expectErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check request method and path
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET method, got %s", r.Method)
				}
				expectedPath := "/preference/" + strconv.Itoa(int(tc.resourceID)) + "/enabledSubdivisionCountries"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Return response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.response))
			}))
			defer server.Close()

			// Create client pointing to test server
			client := givenPreferenceClient(server.URL)

			// Call the method
			ctx := context.Background()
			countries, err := client.GetEnabledSubdivisionCountries(ctx, tc.resourceID)

			// Validate results
			if tc.expectErr && err == nil {
				t.Error("Expected error but got none")
			} else if !tc.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tc.expectErr && countries != nil {
				// Validate response data
				if continent, ok := countries.Continents["NA"]; ok {
					if len(continent.Countries) != 2 {
						t.Errorf("Expected 2 countries for NA, got %d", len(continent.Countries))
					}
					if !containsString(continent.Countries, "US") {
						t.Error("Expected US in NA countries but not found")
					}
					if !containsString(continent.Countries, "CA") {
						t.Error("Expected CA in NA countries but not found")
					}
				} else {
					t.Error("Expected NA continent but found none")
				}

				if continent, ok := countries.Continents["EU"]; ok {
					if len(continent.Countries) != 2 {
						t.Errorf("Expected 2 countries for EU, got %d", len(continent.Countries))
					}
					if !containsString(continent.Countries, "DE") {
						t.Error("Expected DE in EU countries but not found")
					}
					if !containsString(continent.Countries, "FR") {
						t.Error("Expected FR in EU countries but not found")
					}
				} else {
					t.Error("Expected EU continent but found none")
				}
			}
		})
	}
}

// Helper function to check if a string slice contains a specific string
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
