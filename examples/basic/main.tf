terraform {
  required_providers {
    multicdn = {
      source  = "constellix/constellix-multicdn"
      version = "0.0.4"
    }
  }
}

variable "api_key" {
  default = "<api-key>" # Replace with your actual API key
}
variable "api_secret" {
  default = "<secret-key>" # Replace with your actual API secret
}

variable "base_url" {
  default = "https://<multicdn-api-endpoint>" # Replace with your actual base URL
}

provider "multicdn" {
  api_key    = var.api_key
  api_secret = var.api_secret
  base_url   = var.base_url
}

# Create a CDN configuration resource
resource "multicdn_cdn_config" "example_website" {
  resource_id  = 123456
  content_type = "website"
  description  = "Main website CDN configuration"

  cdns = [
    {
      cdn_name      = "CloudFront5"
      description   = "AWS CloudFront CDN"
      fqdn          = "d1234abcdef.cloudfront.net"
      client_cdn_id = "CF12345"
    },
    {
      cdn_name      = "Fastly"
      description   = "Fastly CDN"
      fqdn          = "example.global.fastly.net"
      client_cdn_id = "FY67890"
    },
    {
      cdn_name      = "Akamai"
      description   = "Akamai CDN"
      fqdn          = "example.akamaized.net"
      client_cdn_id = "AK54321"
    }
  ]

  // Minimal required enablement map
  cdn_enablement_map = {
    world_default = ["AK54321", "FY67890", "CF12345"]
    asn_overrides = {
      "AS12345" = ["AK54321", "CF12345"]
      "AS67890" = ["FY67890"]
    }
    continents = {
      "NA" = {
        default = ["AK54321"]
      }
    }
  }

  // Minimal required traffic distribution
  traffic_distribution = {
    world_default = {
      options = [
        {
          name = "world-options"
          distribution = [
            {
              id     = "AK54321"
              weight = 40
            },
            {
              id     = "FY67890"
              weight = 45
            },
            {
              id     = "CF12345"
              weight = 15
            }
          ]
        }
      ]
    }
  }
}

# Create a preference resource
resource "multicdn_preference_config" "example_website_preferences" {
  resource_id  = multicdn_cdn_config.example_website.resource_id
  content_type = "website"
  description  = "CDN preference settings for main website"

  availability_thresholds = {
    world = 95
    continents = {
      "NA" = {
        default = 98
        countries = {
          "US" = 99
          "CA" = 97
          "MX" = 95
        }
      }
      "EU" = {
        default = 97
        countries = {
          "GB" = 98
          "DE" = 98
          "FR" = 96
        }
      }
      "AS" = {
        default = 90
        countries = {
          "JP" = 95
          "SG" = 95
          "IN" = 85
        }
      }
    }
  }

  performance_filtering = {
    world = {
      mode               = "relative"
      relative_threshold = 0.32
    },
    continents = {
      "EU" = {
        countries = {
          "DE" = {
            mode               = "relative"
            relative_threshold = 0.22
          },
        }
        mode               = "relative"
        relative_threshold = 0.4
      }
    }
  }

  enabled_subdivision_countries = {
    continents = {
      "NA" = {
        countries = [
          "US",
          "CA",
          "UK",
        ]
      },
    }
  }
}
