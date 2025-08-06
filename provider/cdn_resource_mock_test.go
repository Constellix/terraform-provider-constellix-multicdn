package provider_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/constellix/terraform-provider-multicdn/clients/cdnclient"
	"github.com/constellix/terraform-provider-multicdn/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// setupMockCdnServer creates a mock HTTP server for CDN config tests
func setupMockCdnServer() (*httptest.Server, map[int]*cdnclient.CdnConfigurationResponse) {
	// Map to store configurations by resource ID
	mockCdnConfigs := make(map[int]*cdnclient.CdnConfigurationResponse)

	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for authentication
		authHeader := r.Header.Get("x-cns-security-token")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_, err := w.Write([]byte(`{"error": "Unauthorized"}`))
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			return
		}

		// Parse the URL path
		pathParts := strings.Split(r.URL.Path, "/")

		// Handle different API endpoints
		switch {
		case r.URL.Path == "/cdn-configs" && r.Method == http.MethodGet:
			// List configurations (paginated)
			pageStr := r.URL.Query().Get("page")
			sizeStr := r.URL.Query().Get("size")
			page, _ := strconv.Atoi(pageStr)
			size, _ := strconv.Atoi(sizeStr)
			if size == 0 {
				size = 50
			}

			var configs []cdnclient.CdnConfigurationResponse
			for _, config := range mockCdnConfigs {
				configs = append(configs, *config)
			}

			response := cdnclient.CdnConfigurationPage{
				Configs:          configs,
				TotalElements:    len(configs),
				TotalPages:       1,
				PageNumber:       page,
				PageSize:         size,
				NumberOfElements: len(configs),
				First:            true,
				Last:             true,
				Empty:            len(configs) == 0,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)

		case r.URL.Path == "/cdn-configs" && r.Method == http.MethodPost:
			// Create a new configuration
			var newConfig cdnclient.CdnConfiguration
			if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Validate the request
			if newConfig.ResourceID <= 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("resource_id must be a positive integer"))
				return
			}

			// Create a new configuration response
			now := time.Now()
			version := "v1.0"
			id := "12345-abcde-67890"
			accountId := 1001

			configResponse := &cdnclient.CdnConfigurationResponse{
				ID:                  &id,
				AccountID:           &accountId,
				ResourceID:          newConfig.ResourceID,
				ContentType:         newConfig.ContentType,
				Description:         newConfig.Description,
				Version:             &version,
				LastUpdated:         &now,
				Cdns:                newConfig.Cdns,
				CdnEnablementMap:    newConfig.CdnEnablementMap,
				TrafficDistribution: newConfig.TrafficDistribution,
			}

			// Save to mock store
			mockCdnConfigs[newConfig.ResourceID] = configResponse

			// Return the created configuration
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(configResponse)

		case strings.HasPrefix(r.URL.Path, "/cdn-configs/") && r.Method == http.MethodGet:
			// Get a specific configuration by ID
			if len(pathParts) != 3 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			resourceID, err := strconv.Atoi(pathParts[2])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			config, exists := mockCdnConfigs[resourceID]
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(config)

		case strings.HasPrefix(r.URL.Path, "/cdn-configs/") && r.Method == http.MethodPut:
			// Update a specific configuration
			if len(pathParts) != 3 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			resourceID, err := strconv.Atoi(pathParts[2])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			_, exists := mockCdnConfigs[resourceID]
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			// Parse the update request
			var updateConfig cdnclient.CdnConfiguration
			if err := json.NewDecoder(r.Body).Decode(&updateConfig); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Update the configuration
			now := time.Now()
			version := "v1.1"

			// Get existing values
			id := mockCdnConfigs[resourceID].ID
			accountId := mockCdnConfigs[resourceID].AccountID

			updatedConfig := &cdnclient.CdnConfigurationResponse{
				ID:                  id,
				AccountID:           accountId,
				ResourceID:          resourceID,
				ContentType:         updateConfig.ContentType,
				Description:         updateConfig.Description,
				Version:             &version,
				LastUpdated:         &now,
				Cdns:                updateConfig.Cdns,
				CdnEnablementMap:    updateConfig.CdnEnablementMap,
				TrafficDistribution: updateConfig.TrafficDistribution,
			}

			// Save to mock store
			mockCdnConfigs[resourceID] = updatedConfig

			// Return the updated configuration
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(updatedConfig)

		case strings.HasPrefix(r.URL.Path, "/cdn-configs/") && r.Method == http.MethodDelete:
			// Delete a specific configuration
			if len(pathParts) != 3 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			resourceID, err := strconv.Atoi(pathParts[2])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			_, exists := mockCdnConfigs[resourceID]
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			// Delete from mock store
			delete(mockCdnConfigs, resourceID)

			// Return success
			w.WriteHeader(http.StatusNoContent)

		case strings.HasPrefix(r.URL.Path, "/cdn-configs/") && strings.HasSuffix(r.URL.Path, "/cdns") && r.Method == http.MethodGet:
			// Get CDN entries for a configuration
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) != 4 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			resourceID, err := strconv.Atoi(parts[2])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			config, exists := mockCdnConfigs[resourceID]
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(config.Cdns)

		case strings.HasPrefix(r.URL.Path, "/cdn-configs/") && strings.HasSuffix(r.URL.Path, "/enablement") && r.Method == http.MethodGet:
			// Get enablement map for a configuration
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) != 4 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			resourceID, err := strconv.Atoi(parts[2])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			config, exists := mockCdnConfigs[resourceID]
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(config.CdnEnablementMap)

		case strings.HasPrefix(r.URL.Path, "/cdn-configs/") && strings.HasSuffix(r.URL.Path, "/trafficDistribution") && r.Method == http.MethodGet:
			// Get traffic distribution for a configuration
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) != 4 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			resourceID, err := strconv.Atoi(parts[2])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			config, exists := mockCdnConfigs[resourceID]
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(config.TrafficDistribution)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	return mockServer, mockCdnConfigs
}

// setupAccProtoV6ProviderFactories creates provider factories with a mock server
func setupCdnAccProtoV6ProviderFactories() (*httptest.Server, map[int]*cdnclient.CdnConfigurationResponse, map[string]func() (tfprotov6.ProviderServer, error)) {
	// Create the mock server
	mockServer, mockCdnConfigs := setupMockCdnServer()

	// Create provider factories using the mock server URL
	factories := map[string]func() (tfprotov6.ProviderServer, error){
		"multicdn": func() (tfprotov6.ProviderServer, error) {
			testProvider := provider.New()
			return providerserver.NewProtocol6(testProvider)(), nil
		},
	}

	return mockServer, mockCdnConfigs, factories
}
