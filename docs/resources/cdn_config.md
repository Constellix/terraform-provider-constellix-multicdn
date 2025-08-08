# multicdn_cdn_config (Resource)

Manages a CDN configuration document which defines your CDN providers, their enablement across regions, and traffic distribution rules.

## Example Usage

### Basic Configuration

```terraform
resource "multicdn_cdn_config" "website" {
  resource_id  = 12345
  content_type = "website"
  description  = "Main website CDN configuration"

  # Define your CDN providers
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

  # Define which CDNs are enabled in which regions
  cdn_enablement_map = {
    world_default = ["CF12345", "FY67890"]  # Default CDNs enabled globally
    continents = {
      "NA" = {  # North America
        default = ["CF12345"]
        countries = {
          "US" = {
            default = ["CF12345", "FY67890"]
          }
        }
      }
    }
  }

  # Define how traffic is distributed between enabled CDNs
  traffic_distribution = {
    world_default = {
      options = [
        {
          name = "standard-distribution"
          distribution = [
            {
              id     = "CF12345"
              weight = 60
            },
            {
              id     = "FY67890"
              weight = 40
            }
          ]
        }
      ]
    }
  }
}
```

### Advanced Configuration with Region-Specific Settings

```terraform
resource "multicdn_cdn_config" "video_streaming" {
  resource_id  = 67890
  content_type = "video"
  description  = "Video streaming CDN configuration"

  cdns = [
    {
      cdn_name     = "CloudFront"
      description  = "AWS CloudFront CDN"
      fqdn         = "d1234abcdef.cloudfront.net"
      client_cdn_id = "CF12345"
    },
    {
      cdn_name     = "Akamai"
      description  = "Akamai CDN"
      fqdn         = "example.akamaized.net"
      client_cdn_id = "AK54321"
    },
    {
      cdn_name     = "Fastly"
      description  = "Fastly CDN"
      fqdn         = "example.global.fastly.net"
      client_cdn_id = "FY67890"
    }
  ]

  # Complex enablement map with continent, country, and ASN-specific settings
  cdn_enablement_map = {
    world_default = ["AK54321", "FY67890", "CF12345"]
    
    # ASN-specific overrides
    asn_overrides = {
      "AS12345" = ["AK54321", "CF12345"]
      "AS67890" = ["FY67890"]
    }
    
    # Continental settings
    continents = {
      "NA" = {
        default = ["CF12345", "FY67890"]
        countries = {
          "US" = {
            default = ["CF12345"]
            # Country-specific ASN overrides
            asn_overrides = {
              "AS23456" = ["FY67890"]
            }
            # State/province level settings for US
            subdivisions = {
              "CA" = {
                asn_overrides = {
                  "AS34567" = ["AK54321"]
                }
              }
            }
          },
          "CA" = {
            default = ["AK54321", "FY67890"]
          }
        }
      },
      "EU" = {
        default = ["AK54321"]
        countries = {
          "DE" = {
            default = ["AK54321", "FY67890"]
          },
          "FR" = {
            default = ["FY67890"]
          }
        }
      }
    }
  }

  # Region-specific traffic distribution configurations
  traffic_distribution = {
    world_default = {
      options = [
        {
          name = "default-distribution"
          distribution = [
            { id = "AK54321", weight = 40 },
            { id = "FY67890", weight = 35 },
            { id = "CF12345", weight = 25 }
          ]
        }
      ]
    },
    continents = {
      "NA" = {
        default = {
          options = [
            {
              name = "na-distribution"
              distribution = [
                { id = "CF12345", weight = 70 },
                { id = "FY67890", weight = 30 }
              ]
            }
          ]
        },
        countries = {
          "US" = {
            default = {
              options = [
                {
                  name = "us-distribution"
                  distribution = [
                    { id = "CF12345", weight = 100 }
                  ]
                }
              ]
            }
          }
        }
      }
    }
  }
}
```

## Schema

### Required

- `cdn_enablement_map` (Attributes) CDN enablement configuration (see [below for nested schema](#nestedatt--cdn_enablement_map))
- `cdns` (Attributes List) List of CDN provider entries (see [below for nested schema](#nestedatt--cdns))
- `resource_id` (Number) Unique ID of the CDN configuration
- `traffic_distribution` (Attributes) Traffic distribution configuration (see [below for nested schema](#nestedatt--traffic_distribution))

### Optional

- `content_type` (String) Content type of the CDN configuration (e.g., "website", "video", "images")
- `description` (String) Description of the CDN configuration
- `last_updated` (String) Timestamp of when the configuration was last updated
- `version` (String) Version of the CDN configuration

<a id="nestedatt--cdn_enablement_map"></a>
### Nested Schema for `cdn_enablement_map`

Optional:

- `asn_overrides` (Map of List of String) ASN-specific CDN overrides
- `continents` (Attributes Map) Continent-specific enablement configurations (see [below for nested schema](#nestedatt--cdn_enablement_map--continents))
- `world_default` (List of String) Default CDNs enabled globally

<a id="nestedatt--cdn_enablement_map--continents"></a>
### Nested Schema for `cdn_enablement_map.continents`

Optional:

- `countries` (Attributes Map) Country-specific enablement configurations (see [below for nested schema](#nestedatt--cdn_enablement_map--continents--countries))
- `default` (List of String) Default CDNs enabled for the continent

<a id="nestedatt--cdn_enablement_map--continents--countries"></a>
### Nested Schema for `cdn_enablement_map.continents.countries`

Optional:

- `asn_overrides` (Map of List of String) ASN-specific CDN overrides for the country
- `default` (List of String) Default CDNs enabled for the country
- `subdivisions` (Attributes Map) Subdivision-specific enablement configurations (see [below for nested schema](#nestedatt--cdn_enablement_map--continents--countries--subdivisions))

<a id="nestedatt--cdn_enablement_map--continents--countries--subdivisions"></a>
### Nested Schema for `cdn_enablement_map.continents.countries.subdivisions`

Optional:

- `asn_overrides` (Map of List of String) ASN-specific CDN overrides for the subdivision





<a id="nestedatt--cdns"></a>
### Nested Schema for `cdns`

Required:

- `cdn_name` (String) Name of the CDN provider
- `client_cdn_id` (String) Client CDN identifier
- `fqdn` (String) Fully qualified domain name for the CDN

Optional:

- `description` (String) Description of the CDN provider entry


<a id="nestedatt--traffic_distribution"></a>
### Nested Schema for `traffic_distribution`

Optional:

- `continents` (Attributes Map) Continent-specific traffic distributions (see [below for nested schema](#nestedatt--traffic_distribution--continents))
- `world_default` (Attributes) Global default traffic distribution (see [below for nested schema](#nestedatt--traffic_distribution--world_default))

<a id="nestedatt--traffic_distribution--continents"></a>
### Nested Schema for `traffic_distribution.continents`

Optional:

- `countries` (Attributes Map) Country-specific traffic distributions (see [below for nested schema](#nestedatt--traffic_distribution--continents--countries))
- `default` (Attributes) Default traffic distribution for the continent (see [below for nested schema](#nestedatt--traffic_distribution--continents--default))

<a id="nestedatt--traffic_distribution--continents--countries"></a>
### Nested Schema for `traffic_distribution.continents.countries`

Required:

- `default` (Attributes) Default traffic distribution for the country (see [below for nested schema](#nestedatt--traffic_distribution--continents--countries--default))

<a id="nestedatt--traffic_distribution--continents--countries--default"></a>
### Nested Schema for `traffic_distribution.continents.countries.default`

Required:

- `options` (Attributes List) List of traffic distribution options (see [below for nested schema](#nestedatt--traffic_distribution--continents--countries--default--options))

<a id="nestedatt--traffic_distribution--continents--countries--default--options"></a>
### Nested Schema for `traffic_distribution.continents.countries.default.options`

Required:

- `distribution` (Attributes List) Traffic distribution weights (see [below for nested schema](#nestedatt--traffic_distribution--continents--countries--default--options--distribution))
- `name` (String) Name of the distribution option

Optional:

- `description` (String) Description of the traffic option
- `equal_weight` (Boolean) Whether traffic is distributed equally

<a id="nestedatt--traffic_distribution--continents--countries--default--options--distribution"></a>
### Nested Schema for `traffic_distribution.continents.countries.default.options.distribution`

Required:

- `id` (String) CDN identifier

Optional:

- `weight` (Number) Traffic weight percentage (0-100)


<a id="nestedatt--traffic_distribution--continents--default"></a>
### Nested Schema for `traffic_distribution.continents.default`

Required:

- `options` (Attributes List) List of traffic distribution options (see [below for nested schema](#nestedatt--traffic_distribution--continents--default--options))

<a id="nestedatt--traffic_distribution--continents--default--options"></a>
### Nested Schema for `traffic_distribution.continents.default.options`

Required:

- `distribution` (Attributes List) Traffic distribution weights (see [below for nested schema](#nestedatt--traffic_distribution--continents--default--options--distribution))
- `name` (String) Name of the distribution option

Optional:

- `description` (String) Description of the traffic option
- `equal_weight` (Boolean) Whether traffic is distributed equally

<a id="nestedatt--traffic_distribution--continents--default--options--distribution"></a>
### Nested Schema for `traffic_distribution.continents.default.options.distribution`

Required:

- `id` (String) CDN identifier

Optional:

- `weight` (Number) Traffic weight percentage (0-100)


<a id="nestedatt--traffic_distribution--world_default"></a>
### Nested Schema for `traffic_distribution.world_default`

Required:

- `options` (Attributes List) List of traffic distribution options (see [below for nested schema](#nestedatt--traffic_distribution--world_default--options))

<a id="nestedatt--traffic_distribution--world_default--options"></a>
### Nested Schema for `traffic_distribution.world_default.options`

Required:

- `distribution` (Attributes List) Traffic distribution weights (see [below for nested schema](#nestedatt--traffic_distribution--world_default--options--distribution))
- `name` (String) Name of the distribution option

Optional:

- `description` (String) Description of the traffic option
- `equal_weight` (Boolean) Whether traffic is distributed equally

<a id="nestedatt--traffic_distribution--world_default--options--distribution"></a>
### Nested Schema for `traffic_distribution.world_default.options.distribution`

Required:

- `id` (String) CDN identifier

Optional:

- `weight` (Number) Traffic weight percentage (0-100)
