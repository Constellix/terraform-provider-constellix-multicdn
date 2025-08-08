# multicdn_preference_config (Resource)

Manages a CDN preference configuration that controls how CDNs are selected based on availability thresholds, performance metrics, and region-specific settings.

## Example Usage

### Basic Configuration

```terraform
resource "multicdn_preference_config" "basic_preferences" {
  resource_id  = multicdn_cdn_config.website.resource_id  # Link to your CDN config
  description  = "Basic CDN preference settings"
  content_type = "website"
  
  # Define minimum availability thresholds (percentage)
  availability_thresholds = {
    world = 95  # Global default availability threshold (95%)
  }
  
  # Define performance filtering options
  performance_filtering = {
    world = {
      mode = "relative"  # Use relative performance filtering
    }
  }
  
  # Define countries where subdivision-level granularity is enabled
  enabled_subdivision_countries = {
    continents = {
      "NA" = {
        countries = ["US", "CA"]  # Enable subdivisions for US and Canada
      }
    }
  }
}
```

### Advanced Configuration with Region-Specific Settings

```terraform
resource "multicdn_preference_config" "advanced_preferences" {
  resource_id  = multicdn_cdn_config.video_streaming.resource_id
  description  = "Advanced CDN preference settings for video streaming"
  content_type = "video"
  
  # Define region-specific availability thresholds
  availability_thresholds = {
    world = 90  # Global default availability threshold (90%)
    
    continents = {
      "NA" = {
        default = 95  # Higher threshold for North America (95%)
        countries = {
          "US" = 98,  # Even higher for US (98%)
          "CA" = 96   # Higher for Canada (96%)
        }
      },
      "EU" = {
        default = 94  # Higher threshold for Europe (94%)
        countries = {
          "DE" = 96,  # Higher for Germany (96%)
          "FR" = 95   # Higher for France (95%)
        }
      },
      "AS" = {
        default = 85  # Lower threshold for Asia (85%)
      }
    }
  }
  
  # Define region-specific performance filtering rules
  performance_filtering = {
    world = {
      mode = "relative"  # Use relative performance filtering globally
    },
    
    continents = {
      "NA" = {
        mode = "relative",
        relative_threshold = 1.2,  # Accept CDNs within 20% of the fastest one in North America
        countries = {
          "US" = {
            mode = "relative",
            relative_threshold = 1.1  # Stricter in US (within 10% of fastest)
          }
        }
      },
      "EU" = {
        mode = "relative",
        relative_threshold = 1.3  # More lenient in Europe (within 30% of fastest)
      },
      "AS" = {
        mode = "relative",
        relative_threshold = 1.5  # Most lenient in Asia (within 50% of fastest)
      }
    }
  }
  
  # Define countries where subdivision-level granularity is enabled
  enabled_subdivision_countries = {
    continents = {
      "NA" = {
        countries = ["US", "CA", "MX"]
      },
      "EU" = {
        countries = ["DE", "FR", "UK", "IT", "ES"]
      },
      "AS" = {
        countries = ["JP", "CN", "IN", "SG"]
      }
    }
  }
}
```

### Configuration Linked with CDN Config

```terraform
# First define your CDN configuration
resource "multicdn_cdn_config" "api_gateway" {
  resource_id  = 12345
  content_type = "api"
  description  = "API Gateway CDN configuration"

  cdns = [
    {
      cdn_name     = "CloudFront"
      description  = "AWS CloudFront CDN"
      fqdn         = "d1234abcdef.cloudfront.net"
      client_cdn_id = "CF12345"
    },
    {
      cdn_name     = "Fastly"
      description  = "Fastly CDN"
      fqdn         = "example.global.fastly.net"
      client_cdn_id = "FY67890"
    }
  ]

  cdn_enablement_map = {
    world_default = ["CF12345", "FY67890"]
  }

  traffic_distribution = {
    world_default = {
      options = [
        {
          name = "standard-distribution"
          distribution = [
            { id = "CF12345", weight = 50 },
            { id = "FY67890", weight = 50 }
          ]
        }
      ]
    }
  }
}

# Then define preferences that reference the CDN config
resource "multicdn_preference_config" "api_preferences" {
  resource_id  = multicdn_cdn_config.api_gateway.resource_id
  content_type = "api"
  description  = "API Gateway preferences"
  
  availability_thresholds = {
    world = 99  # High availability requirement for API
    
    continents = {
      "NA" = {
        default = 99.5  # Even higher for North America
      },
      "EU" = {
        default = 99.5  # Even higher for Europe
      }
    }
  }
  
  performance_filtering = {
    world = {
      mode = "relative"
    },
    continents = {
      "NA" = {
        mode = "relative",
        relative_threshold = 1.1  # Strict performance requirements
      },
      "EU" = {
        mode = "relative",
        relative_threshold = 1.1  # Strict performance requirements
      }
    }
  }
  
  enabled_subdivision_countries = {
    continents = {
      "NA" = {
        countries = ["US"]  # Enable state-level granularity in US
      }
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `availability_thresholds` (Attributes) Availability thresholds configuration (see [below for nested schema](#nestedatt--availability_thresholds))
- `enabled_subdivision_countries` (Attributes) Configuration for countries with enabled subdivisions (see [below for nested schema](#nestedatt--enabled_subdivision_countries))
- `performance_filtering` (Attributes) Performance filtering configuration (see [below for nested schema](#nestedatt--performance_filtering))
- `resource_id` (Number) Unique ID of the CDN preference configuration

### Optional

- `content_type` (String) Content type of the CDN preference configuration
- `description` (String) Description of the CDN preference configuration
- `last_updated` (String) Timestamp of when the configuration was last updated
- `version` (String) Version of the CDN preference configuration

<a id="nestedatt--availability_thresholds"></a>
### Nested Schema for `availability_thresholds`

Optional:

- `continents` (Attributes Map) Continent-specific availability thresholds (see [below for nested schema](#nestedatt--availability_thresholds--continents))
- `world` (Number) Global availability threshold (0-100)

<a id="nestedatt--availability_thresholds--continents"></a>
### Nested Schema for `availability_thresholds.continents`

Required:

- `default` (Number) Default threshold for the continent (0-100)

Optional:

- `countries` (Map of Number) Country-specific thresholds (0-100)



<a id="nestedatt--enabled_subdivision_countries"></a>
### Nested Schema for `enabled_subdivision_countries`

Optional:

- `continents` (Attributes Map) Continent-specific subdivision configurations (see [below for nested schema](#nestedatt--enabled_subdivision_countries--continents))

<a id="nestedatt--enabled_subdivision_countries--continents"></a>
### Nested Schema for `enabled_subdivision_countries.continents`

Optional:

- `countries` (List of String) List of countries with enabled subdivisions



<a id="nestedatt--performance_filtering"></a>
### Nested Schema for `performance_filtering`

Optional:

- `continents` (Attributes Map) Continent-specific performance configurations (see [below for nested schema](#nestedatt--performance_filtering--continents))
- `world` (Attributes) Global performance filtering configuration (see [below for nested schema](#nestedatt--performance_filtering--world))

<a id="nestedatt--performance_filtering--continents"></a>
### Nested Schema for `performance_filtering.continents`

Optional:

- `countries` (Attributes Map) Country-specific performance configurations (see [below for nested schema](#nestedatt--performance_filtering--continents--countries))
- `mode` (String) Performance filtering mode for the continent (valid values: "relative", "absolute")
- `relative_threshold` (Number) Relative performance threshold for the continent (e.g., 1.2 means within 20% of fastest)

<a id="nestedatt--performance_filtering--continents--countries"></a>
### Nested Schema for `performance_filtering.continents.countries`

Optional:

- `mode` (String) Performance filtering mode for the country (valid values: "relative", "absolute")
- `relative_threshold` (Number) Relative performance threshold for the country (e.g., 1.1 means within 10% of fastest)



<a id="nestedatt--performance_filtering--world"></a>
### Nested Schema for `performance_filtering.world`

Optional:

- `mode` (String) Performance filtering mode (valid values: "relative", "absolute")
- `relative_threshold` (Number) Relative performance threshold for global filtering (e.g., 1.2 means within 20% of fastest)
