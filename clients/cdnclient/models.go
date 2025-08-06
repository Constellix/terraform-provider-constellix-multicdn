// Package cdnclient provides a client for the CDN Configuration API
package cdnclient

import (
	"time"
)

// CdnConfigurationPage represents a paginated list of CDN configurations
type CdnConfigurationPage struct {
	Configs          []CdnConfigurationResponse `json:"configs"`
	TotalElements    int                        `json:"totalElements"`
	TotalPages       int                        `json:"totalPages"`
	PageNumber       int                        `json:"pageNumber"`
	PageSize         int                        `json:"pageSize"`
	NumberOfElements int                        `json:"numberOfElements"`
	First            bool                       `json:"first"`
	Last             bool                       `json:"last"`
	Empty            bool                       `json:"empty"`
}

// CdnConfiguration represents a CDN configuration document
type CdnConfiguration struct {
	ResourceID          int                 `json:"resourceId"`
	ContentType         *string             `json:"contentType,omitempty"`
	Description         *string             `json:"description,omitempty"`
	Version             *string             `json:"version,omitempty"`
	LastUpdated         *time.Time          `json:"lastUpdated,omitempty"`
	Cdns                []CdnEntry          `json:"cdns"`
	CdnEnablementMap    CdnEnablementMap    `json:"cdnEnablementMap"`
	TrafficDistribution TrafficDistribution `json:"trafficDistribution"`
}

// CdnConfigurationResponse represents a CDN configuration response from the API
type CdnConfigurationResponse struct {
	ID                  *string             `json:"id,omitempty"`
	AccountID           *int                `json:"accountId,omitempty"`
	ResourceID          int                 `json:"resourceId"`
	ContentType         *string             `json:"contentType,omitempty"`
	Description         *string             `json:"description,omitempty"`
	Version             *string             `json:"version,omitempty"`
	LastUpdated         *time.Time          `json:"lastUpdated,omitempty"`
	Cdns                []CdnEntry          `json:"cdns"`
	CdnEnablementMap    CdnEnablementMap    `json:"cdnEnablementMap"`
	TrafficDistribution TrafficDistribution `json:"trafficDistribution"`
}

// CdnEntry represents a CDN provider entry
type CdnEntry struct {
	CdnName     string  `json:"cdnName"`
	Description *string `json:"description,omitempty"`
	FQDN        string  `json:"fqdn"`
	ClientCdnID string  `json:"clientCdnId"`
}

// CdnEnablementMap represents the CDN enablement configuration
type CdnEnablementMap struct {
	WorldDefault map[string][]string            `json:"worldDefault"`
	ASNOverrides map[string][]string            `json:"asnOverrides"`
	Continents   map[string]ContinentEnablement `json:"continents"`
}

// ContinentEnablement represents enablement settings for a continent
type ContinentEnablement struct {
	Default   []string                     `json:"default,omitempty"`
	Countries map[string]CountryEnablement `json:"countries,omitempty"`
}

// CountryEnablement represents enablement settings for a country
type CountryEnablement struct {
	Default      []string                         `json:"default,omitempty"`
	ASNOverrides map[string][]string              `json:"asnOverrides,omitempty"`
	Subdivisions map[string]SubdivisionEnablement `json:"subdivisions,omitempty"`
}

// SubdivisionEnablement represents enablement settings for a subdivision
type SubdivisionEnablement struct {
	ASNOverrides map[string][]string `json:"asnOverrides,omitempty"`
}

// TrafficDistribution represents the traffic distribution configuration
type TrafficDistribution struct {
	WorldDefault *WorldDefault                    `json:"worldDefault,omitempty"`
	Continents   map[string]ContinentDistribution `json:"continents,omitempty"`
}

// WorldDefault represents the global default traffic distribution
type WorldDefault struct {
	Options []TrafficOption `json:"options,omitempty"`
}

// ContinentDistribution represents traffic distribution for a continent
type ContinentDistribution struct {
	Default   *TrafficOptionList             `json:"default,omitempty"`
	Countries map[string]CountryDistribution `json:"countries,omitempty"`
}

// CountryDistribution represents traffic distribution for a country
type CountryDistribution struct {
	Default *TrafficOptionList `json:"default,omitempty"`
}

// TrafficOptionList represents a list of traffic distribution options
type TrafficOptionList struct {
	Options []TrafficOption `json:"options,omitempty"`
}

// TrafficOption represents a traffic distribution configuration option
type TrafficOption struct {
	Name         string              `json:"name"`
	Description  *string             `json:"description,omitempty"`
	EqualWeight  *bool               `json:"equalWeight,omitempty"`
	Distribution []DistributionEntry `json:"distribution"`
}

// DistributionEntry represents a single entry in a traffic distribution
type DistributionEntry struct {
	ID     string `json:"id"`
	Weight *int   `json:"weight,omitempty"`
}
