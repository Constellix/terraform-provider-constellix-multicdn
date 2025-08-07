package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/constellix/terraform-provider-multicdn/clients/preferenceclient"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Basic acceptance test for preference resource
func TestAccPreferenceResource_basic(t *testing.T) {
	mockServer, mockPreferences, factories := setupAccProtoV6ProviderFactories()
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPreferenceResourceConfig(mockServer.URL, "Test Description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("multicdn_preference.test", "resource_id", "12345"),
					resource.TestCheckResourceAttr("multicdn_preference.test", "content_type", "application/json"),
					resource.TestCheckResourceAttr("multicdn_preference.test", "description", "Test Description"),
					resource.TestCheckNoResourceAttr("multicdn_preference.test", "version"),
					resource.TestCheckNoResourceAttr("multicdn_preference.test", "last_updated"),
					// Check that the preference exists in our mock store
					testAccCheckPreferenceExists(12345, mockPreferences),
				),
			},
			// Update testing
			{
				Config: testAccPreferenceResourceConfig(mockServer.URL, "Updated Description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("multicdn_preference.test", "description", "Updated Description"),
					testAccCheckPreferenceDescription(12345, "Updated Description", mockPreferences),
				),
			},
			// Import testing
			{
				ResourceName:                         "multicdn_preference.test",
				ImportStateVerifyIdentifierAttribute: "resource_id",
				ImportStateId:                        "12345",
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

// Test for error cases
func TestAccPreferenceResource_errors(t *testing.T) {
	mockServer, _, factories := setupAccProtoV6ProviderFactories()
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			// Invalid resource ID (negative number)
			{
				Config:      testAccPreferenceResourceConfigWithInvalidID(mockServer.URL),
				ExpectError: regexp.MustCompile(`resource_id must be a positive integer`),
			},
		},
	})
}

// Test for complete workflow
func TestAccPreferenceResource_completeLifecycle(t *testing.T) {
	mockServer, mockPreferences, factories := setupAccProtoV6ProviderFactories()
	defer mockServer.Close()

	var resourceID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			// Create a new resource
			{
				Config: testAccPreferenceResourceConfig(mockServer.URL, "Initial Resource"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPreferenceResourceExists("multicdn_preference.test", &resourceID),
					resource.TestCheckResourceAttr("multicdn_preference.test", "description", "Initial Resource"),
					testAccCheckPreferenceExists(12345, mockPreferences),
				),
			},
			// Update the resource
			{
				Config: testAccPreferenceResourceUpdateConfig(mockServer.URL, "Updated Resource"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPreferenceResourceExists("multicdn_preference.test", &resourceID),
					resource.TestCheckResourceAttr("multicdn_preference.test", "description", "Updated Resource"),
					// Check updated availability thresholds
					resource.TestCheckResourceAttr("multicdn_preference.test", "availability_thresholds.world", "96"),
					// Check updated performance filtering
					resource.TestCheckResourceAttr("multicdn_preference.test", "performance_filtering.world.relative_threshold", "0.3"),
					testAccCheckPreferenceWorld(12345, 96, mockPreferences),
				),
			},
			// Import the resource
			{
				ResourceName:                         "multicdn_preference.test",
				ImportStateVerifyIdentifierAttribute: "resource_id",
				ImportStateId:                        "12345", // Explicitly set the ID to import
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
			// Delete happens automatically after all steps
			// Let's verify it's gone from our mock store after the test
		},
		CheckDestroy: testAccCheckPreferenceDestroyed(mockPreferences),
	})
}

// Test for configuration with all nested fields and multiple values
func TestAccPreferenceResource_comprehensive(t *testing.T) {
	mockServer, mockPreferences, factories := setupAccProtoV6ProviderFactories()
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			// Create with comprehensive configuration
			{
				Config: testAccPreferenceResourceConfigComprehensive(mockServer.URL, "Comprehensive Config"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPreferenceExists(54321, mockPreferences),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "resource_id", "54321"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "content_type", "application/json"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "description", "Comprehensive Config"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "version", "1.0"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "last_updated", "2025-08-01T00:00:00Z"),
					// Check availability thresholds - Using the exact string representation without trailing zeros
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "availability_thresholds.world", "95"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "availability_thresholds.continents.NA.default", "98"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "availability_thresholds.continents.EU.default", "97"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "availability_thresholds.continents.NA.countries.US", "99"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "availability_thresholds.continents.NA.countries.CA", "99"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "availability_thresholds.continents.EU.countries.DE", "98"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "availability_thresholds.continents.EU.countries.FR", "97"),
					// Check performance filtering
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "performance_filtering.world.mode", "relative"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "performance_filtering.world.relative_threshold", "0.25"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "performance_filtering.continents.NA.mode", "relative"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "performance_filtering.continents.NA.relative_threshold", "0.2"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "performance_filtering.continents.EU.mode", "absolute"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "performance_filtering.continents.EU.relative_threshold", "0.15"),
					// Check enabled subdivision countries
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "enabled_subdivision_countries.continents.NA.countries.#", "2"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "enabled_subdivision_countries.continents.EU.countries.#", "2"),
				),
			},
			// Update with modifications to nested fields
			{
				Config: testAccPreferenceResourceConfigComprehensiveUpdated(mockServer.URL, "Updated Comprehensive Config"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPreferenceExists(54321, mockPreferences),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "description", "Updated Comprehensive Config"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "version", "1.1"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "last_updated", "2025-08-02T00:00:00Z"),
					// Check updated availability thresholds
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "availability_thresholds.world", "96"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "availability_thresholds.continents.NA.default", "99"),
					// Check updated performance filtering
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "performance_filtering.world.relative_threshold", "0.3"),
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "performance_filtering.continents.EU.mode", "relative"),
					// Check updated enabled subdivision countries
					resource.TestCheckResourceAttr("multicdn_preference.comprehensive", "enabled_subdivision_countries.continents.NA.countries.#", "3"),
				),
			},
		},
	})
}

// Test for minimal configuration with only required fields
func TestAccPreferenceResource_minimal(t *testing.T) {
	mockServer, mockPreferences, factories := setupAccProtoV6ProviderFactories()
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			// Create with minimal configuration
			{
				Config: testAccPreferenceResourceConfigMinimal(mockServer.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPreferenceExists(98765, mockPreferences),
					resource.TestCheckResourceAttr("multicdn_preference.minimal", "resource_id", "98765"),
					// Check minimal availability thresholds - using exact string representation
					resource.TestCheckResourceAttr("multicdn_preference.minimal", "availability_thresholds.world", "95"),
					// Check minimal performance filtering
					resource.TestCheckResourceAttr("multicdn_preference.minimal", "performance_filtering.world.mode", "relative"),
					resource.TestCheckResourceAttr("multicdn_preference.minimal", "performance_filtering.world.relative_threshold", "0.1"),
					// Check empty continents map explicitly using % wildcard for map
					resource.TestCheckResourceAttr("multicdn_preference.minimal", "enabled_subdivision_countries.continents.%", "0"),
				),
			},
		},
	})
}

// Helper function to check if the resource exists in Terraform state
func testAccCheckPreferenceResourceExists(resourceName string, resourceID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Find the resource in the Terraform state
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		*resourceID = rs.Primary.ID
		if *resourceID == "" {
			return fmt.Errorf("resource ID not set")
		}

		return nil
	}
}

// Helper function to verify preference exists in mock store
func testAccCheckPreferenceExists(resourceID int64, mockPreferences map[int64]*preferenceclient.Preference) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if pref, exists := mockPreferences[resourceID]; !exists {
			return fmt.Errorf("preference with ID %d does not exist in mock store", resourceID)
		} else if pref == nil {
			return fmt.Errorf("preference with ID %d is nil", resourceID)
		}
		return nil
	}
}

// Helper function to verify description was updated
func testAccCheckPreferenceDescription(resourceID int64, expected string, mockPreferences map[int64]*preferenceclient.Preference) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		pref, exists := mockPreferences[resourceID]
		if !exists {
			return fmt.Errorf("preference with ID %d does not exist in mock store", resourceID)
		}
		if pref.Description != expected {
			return fmt.Errorf("preference description is %s, expected %s", pref.Description, expected)
		}
		return nil
	}
}

// Helper function to verify world threshold was updated
func testAccCheckPreferenceWorld(resourceID int64, expected int64, mockPreferences map[int64]*preferenceclient.Preference) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		pref, exists := mockPreferences[resourceID]
		if !exists {
			return fmt.Errorf("preference with ID %d does not exist in mock store", resourceID)
		}
		if pref.AvailabilityThresholds.World != expected {
			return fmt.Errorf("preference world threshold is %d, expected %d",
				pref.AvailabilityThresholds.World, expected)
		}
		return nil
	}
}

// Helper to check that the resource is destroyed
func testAccCheckPreferenceDestroyed(mockPreferences map[int64]*preferenceclient.Preference) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if pref, exists := mockPreferences[12345]; exists {
			return fmt.Errorf("preference with ID 12345 still exists after destroy: %v", pref)
		}
		return nil
	}
}

// Base configuration for preference resource testing with mockServer URL
func testAccPreferenceResourceConfig(mockServerURL, description string) string {
	return fmt.Sprintf(`
provider "multicdn" {
  api_key    = "test-key"
  api_secret = "test-secret"
  base_url   = "%s"  # Use the mock server URL
}

resource "multicdn_preference" "test" {
  resource_id = 12345
  content_type = "application/json"
  description = "%s"
  
  availability_thresholds = {
    world = 95
    continents = {
      "NA" = {
        default = 98
        countries = {
          "US" = 99
          "CA" = 97
        }
      }
    }
  }
  
  performance_filtering = {
    world = {
      mode = "relative"
      relative_threshold = 0.2
    }
    continents = {
      "EU" = {
        mode = "relative"
        relative_threshold = 0.1
        countries = {
          "DE" = {
            mode = "relative"
            relative_threshold = 1.0
          }
        }
      }
    }
  }
  
  enabled_subdivision_countries = {
    continents = {
      "NA" = {
        countries = ["US", "CA"]
      }
    }
  }
}
`, mockServerURL, description)
}

// Configuration for testing updates with mockServer URL
func testAccPreferenceResourceUpdateConfig(mockServerURL, description string) string {
	return fmt.Sprintf(`
provider "multicdn" {
  api_key    = "test-key"
  api_secret = "test-secret"
  base_url   = "%s"
}

resource "multicdn_preference" "test" {
  resource_id = 12345
  content_type = "application/json"
  description = "%s"
  
  availability_thresholds = {
    world = 96  # Updated value
    continents = {
      "NA" = {
        default = 98
        countries = {
          "US" = 99
          "CA" = 97
        }
      }
    }
  }
  
  performance_filtering = {
    world = {
      mode = "relative"
      relative_threshold = 0.3  # Updated value
    }
    continents = {
      "EU" = {
        mode = "relative"
        relative_threshold = 0.1
        countries = {
          "DE" = {
            mode = "relative"
            relative_threshold = 0.05
          }
        }
      }
    }
  }
  
  enabled_subdivision_countries ={
    continents = {
      "NA" = {
        countries = ["US", "CA"]
      }
    }
  }
}
`, mockServerURL, description)
}

// Configuration with invalid resource ID and mockServer URL
func testAccPreferenceResourceConfigWithInvalidID(mockServerURL string) string {
	return fmt.Sprintf(`
provider "multicdn" {
  api_key    = "test-key"
  api_secret = "test-secret"
  base_url   = "%s"
}

resource "multicdn_preference" "test" {
  resource_id = -1  # Invalid negative ID
  content_type = "application/json"
  description = "Invalid Resource"
  
  availability_thresholds = {
    world = 95
  }
  
  performance_filtering = {
    world = {
      mode = "relative"
      relative_threshold = 0.5
    }
  }
  
  enabled_subdivision_countries = {
    continents = {}
  }
}
`, mockServerURL)
}

// Configuration for comprehensive test with all nested fields and multiple values
func testAccPreferenceResourceConfigComprehensive(serverURL, description string) string {
	return fmt.Sprintf(`
provider "multicdn" {
  api_key = "test-key"
  api_secret = "test-secret"
  base_url = "%s"
}

resource "multicdn_preference" "comprehensive" {
  resource_id = 54321
  content_type = "application/json"
  description = "%s"
  version = "1.0"
  last_updated = "2025-08-01T00:00:00Z"
  
  // Comprehensive availability thresholds with multiple continents and countries
  availability_thresholds = {
    world = 95
    
    continents = {
      "NA" = {
        default = 98
        countries = {
          "US" = 99
          "CA" = 99
          "MX" = 98
        }
      }
      "EU" = {
        default = 97
        countries = {
          "DE" = 98
          "FR" = 97
          "UK" = 96
          "IT" = 96
        }
      }
      "AS" = {
        default = 96
        countries = {
          "JP" = 97
          "CN" = 96
        }
      }
    }
  }
  
  // Comprehensive performance filtering with multiple continents and countries
  performance_filtering = {
    world = {
      mode = "relative"
      relative_threshold = 0.25
    }
    
    continents = {
      "NA" = {
        mode = "relative"
        relative_threshold = 0.2
        
        countries = {
          "US" = {
            mode = "relative"
            relative_threshold = 0.15
          }
          "CA" = {
            mode = "absolute"
            relative_threshold = 0.1
          }
        }
      }
      "EU" = {
        mode = "absolute"
        relative_threshold = 0.15
        
        countries = {
          "DE" = {
            mode = "relative"
            relative_threshold = 0.12
          }
          "FR" = {
            mode = "relative"
            relative_threshold = 0.13
          }
        }
      }
      "AS" = {
        mode = "relative"
        relative_threshold = 0.18
        
        countries = {
          "JP" = {
            mode = "relative"
            relative_threshold = 0.14
          }
          "CN" = {
            mode = "absolute"
            relative_threshold = 0.2
          }
        }
      }
    }
  }
  
  // Comprehensive enabled subdivision countries with multiple continents and countries
  enabled_subdivision_countries = {
    continents = {
      "NA" = {
        countries = ["US", "CA"]
      }
      "EU" = {
        countries = ["DE", "FR"]
      }
    }
  }
}
`, serverURL, description)
}

// Updated comprehensive configuration for testing updates
func testAccPreferenceResourceConfigComprehensiveUpdated(serverURL, description string) string {
	return fmt.Sprintf(`
provider "multicdn" {
  api_key = "test-key"
  api_secret = "test-secret"
  base_url = "%s"
}

resource "multicdn_preference" "comprehensive" {
  resource_id = 54321
  content_type = "application/json"
  description = "%s"
  version = "1.1"
  last_updated = "2025-08-02T00:00:00Z"
  
  // Updated availability thresholds
  availability_thresholds = {
    world = 96  // Changed from 95
    
    continents = {
      "NA" = {
        default = 99  // Changed from 98
        countries = {
          "US" = 99
          "CA" = 98  // Changed from 99
          "MX" = 97  // Changed from 98
        }
      }
      "EU" = {
        default = 96  // Changed from 97
        countries = {
          "DE" = 95  // Changed from 98
          "FR" = 94  // Changed from 97
          "UK" = 93  // Changed from 96
          "ES" = 92  // Added new country
        }
      }
      "AS" = {
        default = 91  // Changed from 96
        countries = {
          "JP" = 90   // Changed from 97
          "CN" = 91   // Changed from 96
          "KR" = 92   // Added new country
        }
      }
    }
  }
  
  // Updated performance filtering
  performance_filtering = {
    world = {
      mode = "relative"
      relative_threshold = 0.3  // Changed from 0.25
    }
    
    continents = {
      "NA" = {
        mode = "relative"
        relative_threshold = 0.25  // Changed from 0.2
        
        countries = {
          "US" = {
            mode = "relative"
            relative_threshold = 0.2  // Changed from 0.15
          }
          "CA" = {
            mode = "relative"  // Changed from absolute
            relative_threshold = 0.15  // Changed from 0.1
          }
          "MX" = {  // Added new country
            mode = "relative"
            relative_threshold = 0.1
          }
        }
      }
      "EU" = {
        mode = "relative"  // Changed from absolute
        relative_threshold = 0.2  // Changed from 0.15
        
        countries = {
          "DE" = {
            mode = "relative"
            relative_threshold = 0.15  // Changed from 0.12
          }
          "FR" = {
            mode = "absolute"  // Changed from relative
            relative_threshold = 0.15  // Changed from 0.13
          }
        }
      }
      "AS" = {
        mode = "absolute"  // Changed from relative
        relative_threshold = 0.2  // Changed from 0.18
        
        countries = {
          "JP" = {
            mode = "relative"
            relative_threshold = 0.15  // Changed from 0.14
          }
          "CN" = {
            mode = "absolute"
            relative_threshold = 0.25  // Changed from 0.2
          }
        }
      }
    }
  }
  
  // Updated enabled subdivision countries
  enabled_subdivision_countries = {
    continents = {
      "NA" = {
        countries = ["US", "CA", "MX"]  // Added MX
      }
      "EU" = {
        countries = ["DE", "FR", "ES"]  // Added ES
      }
    }
  }
}
`, serverURL, description)
}

// Configuration with only required fields
func testAccPreferenceResourceConfigMinimal(serverURL string) string {
	return fmt.Sprintf(`
provider "multicdn" {
  api_key = "test-key"
  api_secret = "test-secret"
  base_url = "%s"
}

resource "multicdn_preference" "minimal" {
  resource_id = 98765
  
  // Minimal availability thresholds
  availability_thresholds = {
    world = 95
  }
  
  // Minimal performance filtering
  performance_filtering = {
    world = {
      mode = "relative"
      relative_threshold = 0.1
    }
  }
  
  // Minimal enabled subdivision countries (empty but required)
  enabled_subdivision_countries = {
    continents = {}
  }
}
`, serverURL)
}
