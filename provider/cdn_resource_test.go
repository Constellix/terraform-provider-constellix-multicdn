package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/constellix/terraform-provider-multicdn/clients/cdnclient"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Test helper functions
func testAccCheckCdnConfigExists(resourceID int64, configs map[int64]*cdnclient.CdnConfigurationResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, exists := configs[resourceID]; !exists {
			return fmt.Errorf("CDN configuration with ID %d does not exist", resourceID)
		}
		return nil
	}
}

func testAccCheckCdnConfigDescription(resourceID int64, expectedDesc string, configs map[int64]*cdnclient.CdnConfigurationResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config, exists := configs[resourceID]
		if !exists {
			return fmt.Errorf("CDN configuration with ID %d does not exist", resourceID)
		}
		if config.Description == nil || *config.Description != expectedDesc {
			return fmt.Errorf("expected description '%s', got '%s'", expectedDesc, *config.Description)
		}
		return nil
	}
}

// Test configuration templates
func testAccCdnResourceConfig(serverURL, description string) string {
	return fmt.Sprintf(`
provider "multicdn" {
  api_key = "api_key"
  api_secret = "api_secret"
  base_url = "%s"
}

resource "multicdn_cdn_config" "test" {
  resource_id = 12345
  content_type = "application/json"
  description = "%s"
  
  cdns = [
    {
      cdn_name = "cdn1"
      description = "Primary CDN"
      fqdn = "cdn1.example.com"
      client_cdn_id = "cdn1_id"
    },
    {
      cdn_name = "cdn2"
      fqdn = "cdn2.example.com"
      client_cdn_id = "cdn2_id"
    }
  ]
  
  cdn_enablement_map = {
    world_default = {
      "default" = ["cdn1", "cdn2"]
    }
    
    continents = {
      "EU" = {
        default = ["cdn2"]
        
        countries = {
          "DE" = {
            default = ["cdn1", "cdn2"]
          }
        }
      }
    }
  }
  
  traffic_distribution = {
    world_default = {
      options = [
        {
          name = "default-option"
          description = "Default traffic option"
          distribution = [
            {
              id = "cdn1"
              weight = 70
            },
            {
              id = "cdn2"
              weight = 30
            }
          ]
        }
      ]
    }
    
    continents = {
      "EU" = {
        default = {
          options = [
            {
              name = "eu-option"
              equal_weight = true
              distribution = [
                {
                  id = "cdn1"
                },
                {
                  id = "cdn2"
                }
              ]
            }
          ]
        }
        
        countries = {
          "DE" = {
            default = {
              options = [
                {
                  name = "de-option"
                  distribution = [
                    {
                      id = "cdn2"
                      weight = 100
                    }
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
`, serverURL, description)
}

func testAccCdnResourceConfigWithInvalidID(serverURL string) string {
	return fmt.Sprintf(`
provider "multicdn" {
  api_key = "api_key"
  api_secret = "api_secret"
  base_url = "%s"
}

resource "multicdn_cdn_config" "test" {
  resource_id = -1
  content_type = "application/json"
  description = "Test Description"
  
  cdns = [
    {
      cdn_name = "cdn1"
      fqdn = "cdn1.example.com"
      client_cdn_id = "cdn1_id"
    }
  ]
  
  cdn_enablement_map = {
    world_default = {
      "default" = ["cdn1"]
    }
  }
  
  traffic_distribution = {
    world_default = {
      options = [
        {
          name = "default-option"
          distribution = [
            {
              id = "cdn1"
              weight = 100
            }
          ]
        }
      ]
    }
  }
}
`, serverURL)
}

// Basic acceptance test for CDN config resource
func TestAccCdnConfigResource_basic(t *testing.T) {
	mockServer, mockCdnConfigs, factories := setupCdnAccProtoV6ProviderFactories()
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCdnResourceConfig(mockServer.URL, "Test Description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("multicdn_cdn_config.test", "resource_id", "12345"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.test", "content_type", "application/json"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.test", "description", "Test Description"),
					resource.TestCheckNoResourceAttr("multicdn_cdn_config.test", "version"),
					resource.TestCheckNoResourceAttr("multicdn_cdn_config.test", "last_updated"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.test", "cdns.#", "2"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.test", "cdns.0.cdn_name", "cdn1"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.test", "cdns.0.description", "Primary CDN"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.test", "cdns.1.cdn_name", "cdn2"),
					testAccCheckCdnConfigExists(12345, mockCdnConfigs),
				),
			},
			// Update testing
			{
				Config: testAccCdnResourceConfig(mockServer.URL, "Updated Description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("multicdn_cdn_config.test", "description", "Updated Description"),
					testAccCheckCdnConfigDescription(12345, "Updated Description", mockCdnConfigs),
				),
			},
			// Import testing
			{
				ResourceName:                         "multicdn_cdn_config.test",
				ImportStateVerifyIdentifierAttribute: "resource_id",
				ImportStateId:                        "12345",
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

// Test for error cases
func TestAccCdnConfigResource_errors(t *testing.T) {
	mockServer, _, factories := setupCdnAccProtoV6ProviderFactories()
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			// Invalid resource ID (negative number)
			{
				Config:      testAccCdnResourceConfigWithInvalidID(mockServer.URL),
				ExpectError: regexp.MustCompile(`resource_id must be a positive integer`),
			},
		},
	})
}

// Test for complete workflow
func TestAccCdnConfigResource_completeLifecycle(t *testing.T) {
	mockServer, mockCdnConfigs, factories := setupCdnAccProtoV6ProviderFactories()
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			// Create initial configuration
			{
				Config: testAccCdnResourceConfig(mockServer.URL, "Initial Config"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCdnConfigExists(12345, mockCdnConfigs),
				),
			},
			// Verify import works
			{
				ResourceName:                         "multicdn_cdn_config.test",
				ImportStateVerifyIdentifierAttribute: "resource_id",
				ImportStateId:                        "12345", // Explicitly set the ID to import
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
			// Update configuration
			{
				Config: testAccCdnResourceConfig(mockServer.URL, "Updated Config"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("multicdn_cdn_config.test", "description", "Updated Config"),
				),
			},
			// Verify destroy works properly
			{
				Config: fmt.Sprintf(`
provider "multicdn" {
  api_key = "api_key"
  api_secret = "api_secret"
  base_url = "%s"
}
`, mockServer.URL),
				Check: func(s *terraform.State) error {
					if _, exists := mockCdnConfigs[12345]; exists {
						return fmt.Errorf("CDN configuration with ID 12345 still exists after destroy")
					}
					return nil
				},
			},
		},
	})
}

// Test for complete configuration with all nested fields
func TestAccCdnConfigResource_comprehensive(t *testing.T) {
	mockServer, mockCdnConfigs, factories := setupCdnAccProtoV6ProviderFactories()
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			// Create with comprehensive configuration
			{
				Config: testAccCdnResourceConfigComprehensive(mockServer.URL, "Comprehensive Config"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCdnConfigExists(54321, mockCdnConfigs),
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "resource_id", "54321"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "content_type", "application/json"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "description", "Comprehensive Config"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "version", "1.0"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "last_updated", "2025-08-01T00:00:00Z"),
					// Check CDN entries
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "cdns.#", "3"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "cdns.0.cdn_name", "cdn1"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "cdns.1.cdn_name", "cdn2"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "cdns.2.cdn_name", "cdn3"),
					// Check enablement map
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "cdn_enablement_map.world_default.default.#", "2"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "cdn_enablement_map.asn_overrides.12345.#", "1"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "cdn_enablement_map.asn_overrides.67890.#", "2"),
					// Check traffic distribution
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "traffic_distribution.world_default.options.#", "2"),
				),
			},
			// Update the comprehensive configuration
			{
				Config: testAccCdnResourceConfigComprehensiveUpdated(mockServer.URL, "Updated Comprehensive Config"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCdnConfigExists(54321, mockCdnConfigs),
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "description", "Updated Comprehensive Config"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "version", "1.1"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "last_updated", "2025-08-02T00:00:00Z"),
					// Check updated CDN entries
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "cdns.0.description", "Updated Primary CDN"),
					// Check updated enablement map
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "cdn_enablement_map.world_default.default.#", "3"),
					// Check updated traffic distribution
					resource.TestCheckResourceAttr("multicdn_cdn_config.comprehensive", "traffic_distribution.world_default.options.0.distribution.0.weight", "60"),
				),
			},
		},
	})
}

// Test for minimal configuration with only required fields
func TestAccCdnConfigResource_minimal(t *testing.T) {
	mockServer, mockCdnConfigs, factories := setupCdnAccProtoV6ProviderFactories()
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			// Create with minimal configuration
			{
				Config: testAccCdnResourceConfigMinimal(mockServer.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCdnConfigExists(98765, mockCdnConfigs),
					resource.TestCheckResourceAttr("multicdn_cdn_config.minimal", "resource_id", "98765"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.minimal", "cdns.#", "1"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.minimal", "cdns.0.cdn_name", "minimal-cdn"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.minimal", "cdns.0.fqdn", "minimal.example.com"),
					resource.TestCheckResourceAttr("multicdn_cdn_config.minimal", "cdns.0.client_cdn_id", "minimal-id"),
				),
			},
		},
	})
}

// Configuration for comprehensive test with all nested fields
func testAccCdnResourceConfigComprehensive(serverURL, description string) string {
	return fmt.Sprintf(`
provider "multicdn" {
  api_key = "api_key"
  api_secret = "api_secret"
  base_url = "%s"
}

resource "multicdn_cdn_config" "comprehensive" {
  resource_id = 54321
  content_type = "application/json"
  description = "%s"
  version = "1.0"
  last_updated = "2025-08-01T00:00:00Z"

  // Three CDN entries
  cdns = [
    {
      cdn_name = "cdn1"
      description = "Primary CDN"
      fqdn = "cdn1.example.com"
      client_cdn_id = "cdn1_id"
    },
    {
      cdn_name = "cdn2"
      description = "Secondary CDN"
      fqdn = "cdn2.example.com"
      client_cdn_id = "cdn2_id"
    },
    {
      cdn_name = "cdn3"
      description = "Tertiary CDN"
      fqdn = "cdn3.example.com"
      client_cdn_id = "cdn3_id"
    }
  ]
  
  // Comprehensive enablement map
  cdn_enablement_map = {
    world_default = {
      "default" = ["cdn1", "cdn2"]
    }
    
    asn_overrides = {
      "12345" = ["cdn1"]
      "67890" = ["cdn2", "cdn3"]
    }
    
    continents = {
      "NA" = {
        default = ["cdn1", "cdn3"]
        
        countries = {
          "US" = {
            default = ["cdn1", "cdn2"]
            
            asn_overrides = {
              "13579" = ["cdn3"]
              "24680" = ["cdn1", "cdn2"]
            }
            
            subdivisions = {
              "CA" = {
                asn_overrides = {
                  "11111" = ["cdn2"]
                  "22222" = ["cdn1"]
                }
              }
              "NY" = {
                asn_overrides = {
                  "33333" = ["cdn3"]
                }
              }
            }
          }
          "CA" = {
            default = ["cdn2", "cdn3"]
          }
        }
      }
      "EU" = {
        default = ["cdn2", "cdn3"]
        
        countries = {
          "DE" = {
            default = ["cdn1", "cdn3"]
          }
          "FR" = {
            default = ["cdn2"]
          }
        }
      }
    }
  }
  
  // Comprehensive traffic distribution
  traffic_distribution = {
    world_default = {
      options = [
        {
          name = "default-option-1"
          description = "Primary traffic option"
          equal_weight = false
          distribution = [
            {
              id = "cdn1"
              weight = 70
            },
            {
              id = "cdn2"
              weight = 20
            },
            {
              id = "cdn3"
              weight = 10
            }
          ]
        },
        {
          name = "default-option-2"
          description = "Secondary traffic option"
          equal_weight = true
          distribution = [
            {
              id = "cdn1"
            },
            {
              id = "cdn2"
            }
          ]
        }
      ]
    }
    
    continents = {
      "NA" = {
        default = {
          options = [
            {
              name = "na-option"
              equal_weight = false
              distribution = [
                {
                  id = "cdn1"
                  weight = 80
                },
                {
                  id = "cdn3"
                  weight = 20
                }
              ]
            }
          ]
        }
        
        countries = {
          "US" = {
            default = {
              options = [
                {
                  name = "us-option"
                  distribution = [
                    {
                      id = "cdn2"
                      weight = 100
                    }
                  ]
                }
              ]
            }
          }
          "CA" = {
            default = {
              options = [
                {
                  name = "ca-option"
                  equal_weight = true
                  distribution = [
                    {
                      id = "cdn1"
                    },
                    {
                      id = "cdn2"
                    },
                    {
                      id = "cdn3"
                    }
                  ]
                }
              ]
            }
          }
        }
      }
      "EU" = {
        default = {
          options = [
            {
              name = "eu-option"
              distribution = [
                {
                  id = "cdn2"
                  weight = 60
                },
                {
                  id = "cdn3"
                  weight = 40
                }
              ]
            }
          ]
        }
        
        countries = {
          "DE" = {
            default = {
              options = [
                {
                  name = "de-option"
                  distribution = [
                    {
                      id = "cdn1"
                      weight = 100
                    }
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
`, serverURL, description)
}

// Updated comprehensive configuration for testing updates
func testAccCdnResourceConfigComprehensiveUpdated(serverURL, description string) string {
	return fmt.Sprintf(`
provider "multicdn" {
  api_key = "api_key"
  api_secret = "api_secret"
  base_url = "%s"
}

resource "multicdn_cdn_config" "comprehensive" {
  resource_id = 54321
  content_type = "application/json"
  description = "%s"
  version = "1.1"
  last_updated = "2025-08-02T00:00:00Z"
  
  // Updated CDN entries
  cdns = [
    {
      cdn_name = "cdn1"
      description = "Updated Primary CDN"
      fqdn = "cdn1.example.com"
      client_cdn_id = "cdn1_id"
    },
    {
      cdn_name = "cdn2"
      description = "Updated Secondary CDN"
      fqdn = "cdn2.example.com"
      client_cdn_id = "cdn2_id"
    },
    {
      cdn_name = "cdn3"
      description = "Updated Tertiary CDN"
      fqdn = "cdn3.example.com"
      client_cdn_id = "cdn3_id"
    }
  ]
  
  // Updated enablement map
  cdn_enablement_map = {
    world_default = {
      "default" = ["cdn1", "cdn2", "cdn3"]  // Added cdn3
    }
    
    asn_overrides = {
      "12345" = ["cdn1", "cdn3"]  // Added cdn3
      "67890" = ["cdn2"]          // Removed cdn3
    }
    
    continents = {
      "NA" = {
        default = ["cdn2", "cdn3"]  // Changed from cdn1, cdn3
        
        countries = {
          "US" = {
            default = ["cdn1", "cdn3"]  // Changed from cdn1, cdn2
            
            asn_overrides = {
              "13579" = ["cdn1", "cdn2"]  // Changed from cdn3
              "24680" = ["cdn2", "cdn3"]  // Changed from cdn1, cdn2
            }
            
            subdivisions = {
              "CA" = {
                asn_overrides = {
                  "11111" = ["cdn3"]      // Changed from cdn2
                  "22222" = ["cdn1", "cdn3"]  // Added cdn3
                }
              }
              "TX" = {                    // New subdivision
                asn_overrides = {
                  "44444" = ["cdn1"]
                }
              }
            }
          }
          "CA" = {
            default = ["cdn1"]  // Changed from cdn2, cdn3
          }
        }
      }
      "EU" = {
        default = ["cdn1", "cdn2"]  // Changed from cdn2, cdn3
        
        countries = {
          "DE" = {
            default = ["cdn2", "cdn3"]  // Changed from cdn1, cdn3
          }
          "FR" = {
            default = ["cdn3"]  // Changed from cdn2
          }
        }
      }
    }
  }
  
  // Updated traffic distribution
  traffic_distribution = {
    world_default = {
      options = [
        {
          name = "default-option-1"
          description = "Updated Primary traffic option"
          equal_weight = false
          distribution = [
            {
              id = "cdn1"
              weight = 60  // Was 70
            },
            {
              id = "cdn2"
              weight = 30  // Was 20
            },
            {
              id = "cdn3"
              weight = 10  // Unchanged
            }
          ]
        },
        {
          name = "default-option-2"
          description = "Updated Secondary traffic option"
          equal_weight = false  // Was true
          distribution = [
            {
              id = "cdn1"
              weight = 40
            },
            {
              id = "cdn2"
              weight = 60
            }
          ]
        }
      ]
    }
    
    continents = {
      "NA" = {
        default = {
          options = [
            {
              name = "updated-na-option"  // Changed name
              equal_weight = true         // Was false
              distribution = [
                {
                  id = "cdn1"
                },
                {
                  id = "cdn3"
                }
              ]
            }
          ]
        }
        
        countries = {
          "US" = {
            default = {
              options = [
                {
                  name = "updated-us-option"  // Changed name
                  distribution = [
                    {
                      id = "cdn1"            // Was cdn2
                      weight = 100
                    }
                  ]
                }
              ]
            }
          }
          "CA" = {
            default = {
              options = [
                {
                  name = "updated-ca-option"  // Changed name
                  equal_weight = false        // Was true
                  distribution = [
                    {
                      id = "cdn1"
                      weight = 50
                    },
                    {
                      id = "cdn2"
                      weight = 50
                    }
                  ]
                }
              ]
            }
          }
        }
      }
      "EU" = {
        default = {
          options = [
            {
              name = "updated-eu-option"  // Changed name
              distribution = [
                {
                  id = "cdn1"            // Was cdn2
                  weight = 70            // Was 60
                },
                {
                  id = "cdn3"
                  weight = 30            // Was 40
                }
              ]
            }
          ]
        }
      }
    }
  }
}
`, serverURL, description)
}

// Configuration with only required fields
func testAccCdnResourceConfigMinimal(serverURL string) string {
	return fmt.Sprintf(`
provider "multicdn" {
  api_key = "api_key"
  api_secret = "api_secret"
  base_url = "%s"
}

resource "multicdn_cdn_config" "minimal" {
  resource_id = 98765
  
  // Minimal required CDN entries
  cdns = [
    {
      cdn_name = "minimal-cdn"
      fqdn = "minimal.example.com"
      client_cdn_id = "minimal-id"
    }
  ]
  
  // Minimal required enablement map
  cdn_enablement_map = {
    world_default = {
      "default" = ["minimal-cdn"]
    }
  }
  
  // Minimal required traffic distribution
  traffic_distribution = {
    world_default = {
      options = [
        {
          name = "minimal-option"
          distribution = [
            {
              id = "minimal-cdn"
              weight = 100
            }
          ]
        }
      ]
    }
  }
}
`, serverURL)
}
