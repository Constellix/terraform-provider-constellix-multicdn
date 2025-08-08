# Constellix MultiCDN Provider

The Constellix MultiCDN Provider allows you to manage content delivery network (CDN) configurations and preferences through Terraform. This provider enables you to define and orchestrate multi-CDN setups with various configuration options including CDN enablement, traffic distribution, availability thresholds, and performance filtering.

## Example Usage

```terraform
terraform {
  required_providers {
    multicdn = {
      source  = "constellix/constellix-multicdn"
      version = "0.0.1"
    }
  }
}

# Configure the MultiCDN Provider
provider "multicdn" {
  api_key    = var.api_key     # API Key for MultiCDN authentication
  api_secret = var.api_secret  # API Secret for MultiCDN authentication
  base_url   = var.base_url    # Base URL for the MultiCDN API
}

# Define variables for sensitive information
variable "api_key" {
  type        = string
  description = "API Key for MultiCDN authentication"
  sensitive   = true
}

variable "api_secret" {
  type        = string
  description = "API Secret for MultiCDN authentication"
  sensitive   = true
}

variable "base_url" {
  type        = string
  description = "Base URL for the MultiCDN API"
  default     = "https://api.multicdn.example.com"
}
```

## Authentication

The MultiCDN provider requires both an API key and an API secret for authentication. These credentials should be handled securely, preferably using environment variables or Terraform variables stored in a secure backend.

### Using Environment Variables

```terraform
provider "multicdn" {
  api_key    = var.MULTICDN_API_KEY
  api_secret = var.MULTICDN_API_SECRET
  base_url   = var.MULTICDN_BASE_URL
}
```

```shell
export TF_VAR_MULTICDN_API_KEY="your-api-key"
export TF_VAR_MULTICDN_API_SECRET="your-api-secret"
export TF_VAR_MULTICDN_BASE_URL="https://api.multicdn.example.com"
```
