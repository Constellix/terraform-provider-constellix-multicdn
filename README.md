# Constellix MultiCDN Terraform Provider

This Terraform provider allows for the management of MultiCDN configurations through Terraform. It currently supports two resource types:

- CDN Configuration Resources
- Preference Resources

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24.5 (for development/building from source)

## Building The Provider

### Building

Clone the repository and build the provider:

```shell
git clone <repository-url>
cd terraform-provider-constellix-multicdn
make build
```

### Installing

To install the provider for local development:

```shell
make install
```

This will build and install the provider to `~/.terraform.d/plugins/registry.terraform.io/constellix/multicdn/0.0.1/darwin_arm64/`.

### Testing

Run the unit tests:

```shell
make test
```

Run the acceptance tests (requires API access):

```shell
make testacc
```

## Provider Configuration

The provider requires authentication to access the MultiCDN API:

```hcl
provider "multicdn" {
  api_key    = "your-api-key"
  api_secret = "your-api-secret"
  base_url   = "https://config.myserver.com"
}
```

## Usage

## Example Files

Example configurations are provided in the `examples` directory. To use them:

1. Configure your API credentials:

```shell
export TF_VAR_api_key="your-api-key"
export TF_VAR_api_secret="your-api-secret"
```

2. Navigate to the example directory:

```shell
cd examples/basic
```

3. Initialize Terraform:

```shell
terraform init
```

### CRUD Operations

#### Create Resources

```shell
terraform plan
terraform apply
```

#### Read Resource Information

```shell
terraform show
```

#### Update Resources

Modify the configuration file, then:

```shell
terraform plan
terraform apply
```

#### Delete Resources

```shell
terraform destroy
```

### Import Existing Resources

To import existing CDN configurations:

```shell
terraform import multicdn_cdn_config.example [resource_id]
```

To import existing preference settings:

```shell
terraform import multicdn_preference_config.example [resource_id]
```

## Development

### Adding New Features

1. Implement the feature in the `provider/` directory
2. Add tests in the respective `*_test.go` files
3. Run tests: `make test testacc`
