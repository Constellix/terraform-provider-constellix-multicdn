package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"

	"github.com/constellix/terraform-provider-multicdn/clients/preferenceclient"
)

// setupMockPreferenceServer creates a mock HTTP server that simulates the CDN Preference API
func setupMockPreferenceServer() (*httptest.Server, map[int]*preferenceclient.Preference) {
	// Store for preference resources, keyed by resourceID
	preferences := make(map[int]*preferenceclient.Preference)

	// Set up the mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		// Match routes and handle appropriately
		// GET /preference (list preferences)
		if r.Method == http.MethodGet && r.URL.Path == "/preference" {
			handleListPreferences(w, r, preferences)
			return
		}

		// POST /preference (create preference)
		if r.Method == http.MethodPost && r.URL.Path == "/preference" {
			handleCreatePreference(w, r, preferences)
			return
		}

		// Path with resource ID: /preference/{resourceId}
		resourcePathRegex := regexp.MustCompile(`^/preference/(\d+)$`)
		if matches := resourcePathRegex.FindStringSubmatch(r.URL.Path); len(matches) > 1 {
			resourceID, _ := strconv.Atoi(matches[1])

			// GET /preference/{resourceId}
			if r.Method == http.MethodGet {
				handleGetPreference(w, r, preferences, resourceID)
				return
			}

			// PUT /preference/{resourceId}
			if r.Method == http.MethodPut {
				handleUpdatePreference(w, r, preferences, resourceID)
				return
			}

			// DELETE /preference/{resourceId}
			if r.Method == http.MethodDelete {
				handleDeletePreference(w, r, preferences, resourceID)
				return
			}
		}

		// Handle sub-resources
		thresholdsRegex := regexp.MustCompile(`^/preference/(\d+)/availabilityThresholds$`)
		performanceRegex := regexp.MustCompile(`^/preference/(\d+)/performanceFiltering$`)
		countriesRegex := regexp.MustCompile(`^/preference/(\d+)/enabledSubdivisionCountries$`)

		if matches := thresholdsRegex.FindStringSubmatch(r.URL.Path); len(matches) > 1 && r.Method == http.MethodGet {
			resourceID, _ := strconv.Atoi(matches[1])
			handleGetAvailabilityThresholds(w, r, preferences, resourceID)
			return
		}

		if matches := performanceRegex.FindStringSubmatch(r.URL.Path); len(matches) > 1 && r.Method == http.MethodGet {
			resourceID, _ := strconv.Atoi(matches[1])
			handleGetPerformanceFiltering(w, r, preferences, resourceID)
			return
		}

		if matches := countriesRegex.FindStringSubmatch(r.URL.Path); len(matches) > 1 && r.Method == http.MethodGet {
			resourceID, _ := strconv.Atoi(matches[1])
			handleGetEnabledSubdivisionCountries(w, r, preferences, resourceID)
			return
		}

		// Default: route not found
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte(`{"error": "Route not found"}`))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}))

	return server, preferences
}

// Handler for GET /preference
func handleListPreferences(w http.ResponseWriter, r *http.Request, preferences map[int]*preferenceclient.Preference) {
	// Extract pagination parameters
	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")

	page := 0
	size := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 0 {
			page = p
		}
	}

	if sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
			size = s
		}
	}

	// Convert map to slice for pagination
	allPrefs := make([]*preferenceclient.Preference, 0, len(preferences))
	for _, pref := range preferences {
		allPrefs = append(allPrefs, pref)
	}

	// Calculate pagination
	totalElements := len(allPrefs)
	totalPages := (totalElements + size - 1) / size
	start := page * size
	end := (page + 1) * size
	if end > totalElements {
		end = totalElements
	}

	// Create paginated response
	var pagedPrefs []preferenceclient.Preference
	if start < totalElements {
		pagedPrefs = dereferencePreferences(allPrefs[start:end])
	} else {
		pagedPrefs = []preferenceclient.Preference{}
	}

	prefPage := preferenceclient.PreferencePage{
		PreferenceConfigs: pagedPrefs,
		TotalElements:     totalElements,
		TotalPages:        totalPages,
		PageNumber:        page,
		PageSize:          size,
		NumberOfElements:  len(pagedPrefs),
		First:             page == 0,
		Last:              page >= totalPages-1,
		Empty:             len(pagedPrefs) == 0,
	}

	// Respond with the page
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode([]preferenceclient.PreferencePage{prefPage})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func dereferencePreferences(prefs []*preferenceclient.Preference) []preferenceclient.Preference {
	result := make([]preferenceclient.Preference, len(prefs))
	for i, p := range prefs {
		result[i] = *p
	}
	return result
}

// Handler for POST /preference
func handleCreatePreference(w http.ResponseWriter, r *http.Request, preferences map[int]*preferenceclient.Preference) {
	var preference preferenceclient.Preference

	err := json.NewDecoder(r.Body).Decode(&preference)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{"error": "Invalid request body"}`))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Check required fields
	if preference.ResourceID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{"error": "resource_id must be a positive integer"}`))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Check if resource already exists
	if _, exists := preferences[preference.ResourceID]; exists {
		w.WriteHeader(http.StatusConflict)
		_, err := w.Write([]byte(`{"error": "Resource already exists"}`))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Store the resource
	preferences[preference.ResourceID] = &preference

	w.WriteHeader(http.StatusCreated)
}

// Handler for GET /preference/{resourceId}
func handleGetPreference(w http.ResponseWriter, _ *http.Request, preferences map[int]*preferenceclient.Preference, resourceID int) {
	preference, exists := preferences[resourceID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte(fmt.Sprintf(`{"error": "Preference with ID %d not found"}`, resourceID)))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(preference)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// Handler for PUT /preference/{resourceId}
func handleUpdatePreference(w http.ResponseWriter, r *http.Request, preferences map[int]*preferenceclient.Preference, resourceID int) {
	// Check if resource exists
	_, exists := preferences[resourceID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte(fmt.Sprintf(`{"error": "Preference with ID %d not found"}`, resourceID)))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	var updatedPreference preferenceclient.Preference
	err := json.NewDecoder(r.Body).Decode(&updatedPreference)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{"error": "Invalid request body"}`))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Ensure resourceID in body matches URL
	if updatedPreference.ResourceID != resourceID {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{"error": "Resource ID in body does not match URL"}`))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Update the resource
	preferences[resourceID] = &updatedPreference

	w.WriteHeader(http.StatusOK)
}

// Handler for DELETE /preference/{resourceId}
func handleDeletePreference(w http.ResponseWriter, _ *http.Request, preferences map[int]*preferenceclient.Preference, resourceID int) {
	_, exists := preferences[resourceID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte(fmt.Sprintf(`{"error": "Preference with ID %d not found"}`, resourceID)))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Delete the resource
	delete(preferences, resourceID)

	w.WriteHeader(http.StatusNoContent)
}

// Handler for GET /preference/{resourceId}/availabilityThresholds
func handleGetAvailabilityThresholds(w http.ResponseWriter, _ *http.Request, preferences map[int]*preferenceclient.Preference, resourceID int) {
	preference, exists := preferences[resourceID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte(fmt.Sprintf(`{"error": "Preference with ID %d not found"}`, resourceID)))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(preference.AvailabilityThresholds)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// Handler for GET /preference/{resourceId}/performanceFiltering
func handleGetPerformanceFiltering(w http.ResponseWriter, _ *http.Request, preferences map[int]*preferenceclient.Preference, resourceID int) {
	preference, exists := preferences[resourceID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte(fmt.Sprintf(`{"error": "Preference with ID %d not found"}`, resourceID)))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(preference.PerformanceFiltering)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// Handler for GET /preference/{resourceId}/enabledSubdivisionCountries
func handleGetEnabledSubdivisionCountries(w http.ResponseWriter, _ *http.Request, preferences map[int]*preferenceclient.Preference, resourceID int) {
	preference, exists := preferences[resourceID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte(fmt.Sprintf(`{"error": "Preference with ID %d not found"}`, resourceID)))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(preference.EnabledSubdivisionCountries)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
