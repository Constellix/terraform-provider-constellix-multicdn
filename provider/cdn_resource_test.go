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
func testAccCheckCdnConfigExists(resourceID int, configs map[int]*cdnclient.CdnConfigurationResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, exists := configs[resourceID]; !exists {
			return fmt.Errorf("CDN configuration with ID %d does not exist", resourceID)
		}
		return nil
	}
}

func testAccCheckCdnConfigDescription(resourceID int, expectedDesc string, configs map[int]*cdnclient.CdnConfigurationResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config, exists := configs[resourceID]
		if !exists {
			return fmt.Errorf("CDN configuration with ID %d does not exist", resourceID)
		}
		if config.Description == nil || *config.Description != expectedDesc {
			return fmt.Errorf("Expected description '%s', got '%s'", expectedDesc, *config.Description)
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
					resource.TestCheckResourceAttrSet("multicdn_cdn_config.test", "version"),
					resource.TestCheckResourceAttrSet("multicdn_cdn_config.test", "last_updated"),
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
