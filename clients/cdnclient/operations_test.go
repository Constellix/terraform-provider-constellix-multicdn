package cdnclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/constellix/terraform-provider-multicdn/clients/httpclient"
)

// setupMockServer creates a test server that simulates the CDN Configuration API
func setupMockServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authentication header
		token := r.Header.Get("x-cns-security-token")
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Different responses based on the path and method
		switch {
		case r.URL.Path == "/cdn-configs" && r.Method == http.MethodGet:
			// Get CDN configs page
			pageStr := r.URL.Query().Get("page")
			sizeStr := r.URL.Query().Get("size")

			page, _ := strconv.Atoi(pageStr)
			size, _ := strconv.Atoi(sizeStr)

			if page < 0 || size <= 0 {
				w.WriteHeader(http.StatusBadRequest)
				writeResponse(t, w, []byte(`{"error":"Invalid page or size"}`))
				return
			}

			w.WriteHeader(http.StatusOK)
			writeResponse(t, w, []byte(`{
				"configs": [{
					"id": "d290f1ee-6c54-4b01-90e6-d701748f0851",
					"accountId": 12345,
					"resourceId": 123,
					"contentType": "application/json",
					"description": "Test CDN config",
					"version": "1.0",
					"lastUpdated": "2023-01-01T00:00:00Z",
					"cdns": [
						{
							"cdnName": "cdn1",
							"description": "CDN Provider 1",
							"fqdn": "cdn1.example.com",
							"clientCdnId": "CDN1"
						}
					],
					"cdnEnablementMap": {
						"worldDefault": ["cdn1"],
						"asnOverrides": {},
						"continents": {}
					},
					"trafficDistribution": {
						"worldDefault": {
							"options": [
								{
									"name": "default",
									"distribution": [
										{
											"id": "cdn1",
											"weight": 100
										}
									]
								}
							]
						}
					}
				}],
				"totalElements": 1,
				"totalPages": 1,
				"pageNumber": 0,
				"pageSize": 10,
				"numberOfElements": 1,
				"first": true,
				"last": true,
				"empty": false
			}`))

		case r.URL.Path == "/cdn-configs" && r.Method == http.MethodPost:
			// Create CDN config
			var config CdnConfiguration
			if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeResponse(t, w, []byte(`{"error":"Invalid request body"}`))
				return
			}

			w.WriteHeader(http.StatusCreated)
			writeResponse(t, w, []byte(`{
				"id": "d290f1ee-6c54-4b01-90e6-d701748f0851",
				"accountId": 12345,
				"resourceId": 123,
				"contentType": "application/json",
				"description": "Test CDN config",
				"version": "1.0",
				"lastUpdated": "2023-01-01T00:00:00Z",
				"cdns": [
					{
						"cdnName": "cdn1",
						"description": "CDN Provider 1",
						"fqdn": "cdn1.example.com",
						"clientCdnId": "CDN1"
					}
				],
				"cdnEnablementMap": {
					"worldDefault": ["cdn1"],
					"asnOverrides": {},
					"continents": {}
				},
				"trafficDistribution": {
					"worldDefault": {
						"options": [
							{
								"name": "default",
								"distribution": [
									{
										"id": "cdn1",
										"weight": 100
									}
								]
							}
						]
					}
				}
			}`))

		case r.URL.Path == "/cdn-configs/123" && r.Method == http.MethodGet:
			// Get CDN config
			w.WriteHeader(http.StatusOK)
			writeResponse(t, w, []byte(`{
				"id": "d290f1ee-6c54-4b01-90e6-d701748f0851",
				"accountId": 12345,
				"resourceId": 123,
				"contentType": "application/json",
				"description": "Test CDN config",
				"version": "1.0",
				"lastUpdated": "2023-01-01T00:00:00Z",
				"cdns": [
					{
						"cdnName": "cdn1",
						"description": "CDN Provider 1",
						"fqdn": "cdn1.example.com",
						"clientCdnId": "CDN1"
					}
				],
				"cdnEnablementMap": {
					"worldDefault": ["cdn1"],
					"asnOverrides": {},
					"continents": {}
				},
				"trafficDistribution": {
					"worldDefault": {
						"options": [
							{
								"name": "default",
								"distribution": [
									{
										"id": "cdn1",
										"weight": 100
									}
								]
							}
						]
					}
				}
			}`))

		case r.URL.Path == "/cdn-configs/123" && r.Method == http.MethodPut:
			// Update CDN config
			var config CdnConfiguration
			if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeResponse(t, w, []byte(`{"error":"Invalid request body"}`))
				return
			}

			w.WriteHeader(http.StatusOK)
			writeResponse(t, w, []byte(`{
				"id": "d290f1ee-6c54-4b01-90e6-d701748f0851",
				"accountId": 12345,
				"resourceId": 123,
				"contentType": "application/json",
				"description": "Updated CDN config",
				"version": "1.1",
				"lastUpdated": "2023-01-02T00:00:00Z",
				"cdns": [
					{
						"cdnName": "cdn1",
						"description": "CDN Provider 1 Updated",
						"fqdn": "cdn1.example.com",
						"clientCdnId": "CDN1"
					}
				],
				"cdnEnablementMap": {
					"worldDefault": ["cdn1"],
					"asnOverrides": {},
					"continents": {}
				},
				"trafficDistribution": {
					"worldDefault": {
						"options": [
							{
								"name": "default",
								"distribution": [
									{
										"id": "cdn1",
										"weight": 100
									}
								]
							}
						]
					}
				}
			}`))

		case r.URL.Path == "/cdn-configs/123" && r.Method == http.MethodDelete:
			// Delete CDN config
			w.WriteHeader(http.StatusNoContent)

		case r.URL.Path == "/cdn-configs/123/cdns" && r.Method == http.MethodGet:
			// Get CDN entries
			w.WriteHeader(http.StatusOK)
			writeResponse(t, w, []byte(`[
				{
					"cdnName": "cdn1",
					"description": "CDN Provider 1",
					"fqdn": "cdn1.example.com",
					"clientCdnId": "CDN1"
				},
				{
					"cdnName": "cdn2",
					"description": "CDN Provider 2",
					"fqdn": "cdn2.example.com",
					"clientCdnId": "CDN2"
				}
			]`))

		case r.URL.Path == "/cdn-configs/123/enablement" && r.Method == http.MethodGet:
			// Get CDN enablement map
			w.WriteHeader(http.StatusOK)
			writeResponse(t, w, []byte(`{
				"worldDefault": ["cdn1", "cdn2"],
				"asnOverrides": {
					"12345": ["cdn1"]
				},
				"continents": {
					"NA": {
						"default": ["cdn1"],
						"countries": {
							"US": {
								"default": ["cdn1", "cdn2"],
								"asnOverrides": {
									"54321": ["cdn2"]
								},
								"subdivisions": {
									"CA": {
										"asnOverrides": {
											"98765": ["cdn1"]
										}
									}
								}
							}
						}
					}
				}
			}`))

		case r.URL.Path == "/cdn-configs/123/trafficDistribution" && r.Method == http.MethodGet:
			// Get traffic distribution
			w.WriteHeader(http.StatusOK)
			writeResponse(t, w, []byte(`{
				"worldDefault": {
					"options": [
						{
							"name": "default",
							"equalWeight": false,
							"distribution": [
								{
									"id": "cdn1",
									"weight": 70
								},
								{
									"id": "cdn2",
									"weight": 30
								}
							]
						}
					]
				},
				"continents": {
					"EU": {
						"default": {
							"options": [
								{
									"name": "eu-option",
									"distribution": [
										{
											"id": "cdn1",
											"weight": 50
										},
										{
											"id": "cdn2",
											"weight": 50
										}
									]
								}
							]
						},
						"countries": {
							"DE": {
								"default": {
									"options": [
										{
											"name": "de-option",
											"distribution": [
												{
													"id": "cdn2",
													"weight": 100
												}
											]
										}
									]
								}
							}
						}
					}
				}
			}`))

		case r.URL.Path == "/cdn-configs/404" && r.Method == http.MethodGet:
			// Not found
			w.WriteHeader(http.StatusNotFound)
			writeResponse(t, w, []byte(`{"error":"Resource not found"}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestGetCdnConfigsPage(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name       string
		pageNumber int
		pageSize   int
		expectErr  bool
	}{
		{
			name:       "successful_response",
			pageNumber: 0,
			pageSize:   10,
			expectErr:  false,
		},
		{
			name:       "invalid_parameters",
			pageNumber: -1,
			pageSize:   0,
			expectErr:  true,
		},
	}

	// Setup mock server
	server := setupMockServer(t)
	defer server.Close()

	client := givenCdnClient(server.URL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPage, err := client.GetCdnConfigsPage(context.Background(), tt.pageNumber, tt.pageSize)
			if (err != nil) != tt.expectErr {
				t.Errorf("GetCdnConfigsPage() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			if !tt.expectErr && (configPage == nil || len(configPage.Configs) == 0) {
				t.Errorf("GetCdnConfigsPage() expected configs, got none")
			}
		})
	}
}

func TestCRUDOperations(t *testing.T) {
	// Setup mock server
	server := setupMockServer(t)
	defer server.Close()

	client := givenCdnClient(server.URL)
	ctx := context.Background()

	// Test CDN config for create and update operations
	contentType := "application/json"
	description := "Test CDN config"
	cdnDescription := "CDN Provider 1"
	equalWeight := false
	var weight int64 = 100

	config := &CdnConfiguration{
		ResourceID:  123,
		ContentType: &contentType,
		Description: &description,
		Cdns: []CdnEntry{
			{
				CdnName:     "cdn1",
				Description: &cdnDescription,
				FQDN:        "cdn1.example.com",
				ClientCdnID: "CDN1",
			},
		},
		CdnEnablementMap: CdnEnablementMap{
			WorldDefault: []string{"cdn1"},
		},
		TrafficDistribution: TrafficDistribution{
			WorldDefault: &WorldDefault{
				Options: []TrafficOption{
					{
						Name:        "default",
						EqualWeight: &equalWeight,
						Distribution: []DistributionEntry{
							{
								ID:     "cdn1",
								Weight: &weight,
							},
						},
					},
				},
			},
		},
	}

	// Test Create
	createdConfig, err := client.CreateCdnConfig(ctx, config)
	if err != nil {
		t.Errorf("CreateCdnConfig() error = %v", err)
	}
	if createdConfig.ResourceID != 123 {
		t.Errorf("CreateCdnConfig() expected resourceID 123, got %d", createdConfig.ResourceID)
	}

	// Test Get
	gotConfig, err := client.GetCdnConfig(ctx, 123)
	if err != nil {
		t.Errorf("GetCdnConfig() error = %v", err)
	}
	if gotConfig.ResourceID != 123 {
		t.Errorf("GetCdnConfig() expected resourceID 123, got %d", gotConfig.ResourceID)
	}
	if len(gotConfig.Cdns) == 0 {
		t.Errorf("GetCdnConfig() expected CDN entries, got none")
	}

	// Test Update
	updatedDescription := "Updated CDN config"
	config.Description = &updatedDescription
	updatedConfig, err := client.UpdateCdnConfig(ctx, 123, config)
	if err != nil {
		t.Errorf("UpdateCdnConfig() error = %v", err)
	}
	if updatedConfig.Description == nil || *updatedConfig.Description != "Updated CDN config" {
		t.Errorf("UpdateCdnConfig() expected description 'Updated CDN config', got '%v'", updatedConfig.Description)
	}

	// Test Delete
	err = client.DeleteCdnConfig(ctx, 123)
	if err != nil {
		t.Errorf("DeleteCdnConfig() error = %v", err)
	}

	// Test Get non-existent config
	_, err = client.GetCdnConfig(ctx, 404)
	if err == nil {
		t.Error("GetCdnConfig() expected error for non-existent resource")
	}
}

func TestGetCdnComponents(t *testing.T) {
	// Setup mock server
	server := setupMockServer(t)
	defer server.Close()

	client := givenCdnClient(server.URL)
	ctx := context.Background()

	// Test GetCdnEntries
	entries, err := client.GetCdnEntries(ctx, 123)
	if err != nil {
		t.Errorf("GetCdnEntries() error = %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("GetCdnEntries() expected 2 entries, got %d", len(entries))
	}
	if entries[0].CdnName != "cdn1" {
		t.Errorf("GetCdnEntries() expected cdnName 'cdn1', got '%s'", entries[0].CdnName)
	}

	// Test GetCdnEnablementMap
	enablementMap, err := client.GetCdnEnablementMap(ctx, 123)
	if err != nil {
		t.Errorf("GetCdnEnablementMap() error = %v", err)
	}
	if len(enablementMap.WorldDefault) == 0 {
		t.Errorf("GetCdnEnablementMap() expected worldDefault, got none")
	}
	if len(enablementMap.Continents) == 0 {
		t.Errorf("GetCdnEnablementMap() expected continents, got none")
	}

	// Test GetTrafficDistribution
	distribution, err := client.GetTrafficDistribution(ctx, 123)
	if err != nil {
		t.Errorf("GetTrafficDistribution() error = %v", err)
	}
	if distribution.WorldDefault == nil || len(distribution.WorldDefault.Options) == 0 {
		t.Errorf("GetTrafficDistribution() expected worldDefault options, got none")
	}
	if len(distribution.Continents) == 0 {
		t.Errorf("GetTrafficDistribution() expected continents, got none")
	}
}

func givenCdnClient(serverURL string) *Client {
	return New(httpclient.New(serverURL, "test-key", "test-secret"))
}

func writeResponse(t *testing.T, w http.ResponseWriter, body []byte) {
	if _, err := w.Write(body); err != nil {
		t.Fatalf("Failed to write response: %v", err)
	}
}
