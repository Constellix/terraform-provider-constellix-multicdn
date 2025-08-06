package provider_test

import (
	"fmt"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/constellix/terraform-provider-multicdn/clients/preferenceclient"
	"github.com/constellix/terraform-provider-multicdn/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccProtoV6ProviderFactories creates provider factories with a mock server
func setupAccProtoV6ProviderFactories() (*httptest.Server, map[int]*preferenceclient.Preference, map[string]func() (tfprotov6.ProviderServer, error)) {
	// Create the mock server
	mockServer, mockPreferences := setupMockPreferenceServer()

	// Create provider factories using the mock server URL
	factories := map[string]func() (tfprotov6.ProviderServer, error){
		"multicdn": func() (tfprotov6.ProviderServer, error) {
			testProvider := provider.New()
			return providerserver.NewProtocol6(testProvider)(), nil
		},
	}

	return mockServer, mockPreferences, factories
}

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
					resource.TestCheckResourceAttrSet("multicdn_preference.test", "version"),
					resource.TestCheckResourceAttrSet("multicdn_preference.test", "last_updated"),
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
func testAccCheckPreferenceExists(resourceID int, mockPreferences map[int]*preferenceclient.Preference) resource.TestCheckFunc {
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
func testAccCheckPreferenceDescription(resourceID int, expected string, mockPreferences map[int]*preferenceclient.Preference) resource.TestCheckFunc {
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
func testAccCheckPreferenceWorld(resourceID int, expected float64, mockPreferences map[int]*preferenceclient.Preference) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		pref, exists := mockPreferences[resourceID]
		if !exists {
			return fmt.Errorf("preference with ID %d does not exist in mock store", resourceID)
		}
		if pref.AvailabilityThresholds.World != expected {
			return fmt.Errorf("preference world threshold is %f, expected %f",
				pref.AvailabilityThresholds.World, expected)
		}
		return nil
	}
}

// Helper to check that the resource is destroyed
func testAccCheckPreferenceDestroyed(mockPreferences map[int]*preferenceclient.Preference) resource.TestCheckFunc {
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
