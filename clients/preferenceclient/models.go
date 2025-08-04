package preferenceclient

import "time"

// PreferencePage represents a paginated list of CDN preferences
type PreferencePage struct {
	PreferenceConfigs []Preference `json:"preferenceConfigs"`
	TotalElements     int          `json:"totalElements"`
	TotalPages        int          `json:"totalPages"`
	PageNumber        int          `json:"pageNumber"`
	PageSize          int          `json:"pageSize"`
	NumberOfElements  int          `json:"numberOfElements"`
	First             bool         `json:"first"`
	Last              bool         `json:"last"`
	Empty             bool         `json:"empty"`
}

// Preference represents the complete CDN preference configuration
type Preference struct {
	ResourceID                  int                         `json:"resourceId"`
	ContentType                 string                      `json:"contentType,omitempty"`
	Description                 string                      `json:"description,omitempty"`
	Version                     string                      `json:"version,omitempty"`
	LastUpdated                 time.Time                   `json:"lastUpdated,omitempty"`
	AvailabilityThresholds      AvailabilityThresholds      `json:"availabilityThresholds"`
	PerformanceFiltering        PerformanceFiltering        `json:"performanceFiltering"`
	EnabledSubdivisionCountries EnabledSubdivisionCountries `json:"enabledSubdivisionCountries"`
}

// AvailabilityThresholds represents the thresholds for availability
type AvailabilityThresholds struct {
	World      float64                       `json:"world,omitempty"`
	Continents map[string]ContinentThreshold `json:"continents,omitempty"`
}

// ContinentThreshold represents availability thresholds for a continent
type ContinentThreshold struct {
	Default   float64            `json:"default,omitempty"`
	Countries map[string]float64 `json:"countries,omitempty"`
}

// PerformanceFiltering represents the performance filtering configuration
type PerformanceFiltering struct {
	World      PerformanceConfig                     `json:"world,omitempty"`
	Continents map[string]ContinentPerformanceConfig `json:"continents,omitempty"`
}

// PerformanceConfig represents a performance filtering configuration
type PerformanceConfig struct {
	Mode              string  `json:"mode,omitempty"` // Either "relative" or "absolute"
	RelativeThreshold float64 `json:"relativeThreshold,omitempty"`
}

// ContinentPerformanceConfig represents performance configuration for a continent
type ContinentPerformanceConfig struct {
	Mode              string                       `json:"mode,omitempty"`
	RelativeThreshold float64                      `json:"relativeThreshold,omitempty"`
	Countries         map[string]PerformanceConfig `json:"countries,omitempty"`
}

// EnabledSubdivisionCountries represents countries with enabled subdivisions
type EnabledSubdivisionCountries struct {
	Continents map[string]ContinentSubdivisions `json:"continents,omitempty"`
}

// ContinentSubdivisions represents subdivision countries within a continent
type ContinentSubdivisions struct {
	Countries []string `json:"countries,omitempty"`
}
