package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &multiCDNProvider{}
)

// multiCDNProvider is the provider implementation
type multiCDNProvider struct{}

// multiCDNProviderModel describes the provider data model
type multiCDNProviderModel struct {
	APIKey    types.String `tfsdk:"api_key"`
	APISecret types.String `tfsdk:"api_secret"`
	BaseURL   types.String `tfsdk:"base_url"`
}

// New creates a new instance of the provider
func New() provider.Provider {
	return &multiCDNProvider{}
}

// Metadata returns the provider metadata
func (p *multiCDNProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "multicdn"
}

// Schema defines the provider-level schema for configuration data
func (p *multiCDNProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provider for managing MultiCDN API resources, including CDN and preference configurations.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "API Key for MultiCDN API authentication",
				Required:    true,
				Sensitive:   true,
			},
			"api_secret": schema.StringAttribute{
				Description: "API Secret for MultiCDN API authentication",
				Required:    true,
				Sensitive:   true,
			},
			"base_url": schema.StringAttribute{
				Description: "Base URL for MultiCDN API",
				Required:    true,
			},
		},
	}
}

// Configure prepares a MultiCDN API client for data sources and resources
func (p *multiCDNProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config multiCDNProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for required configuration
	if config.BaseURL.IsNull() || config.BaseURL.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Missing Base URL",
			"The provider cannot create the MultiCDN API client without the base URL",
		)
		return
	}

	// Check for required configuration
	if config.APIKey.IsNull() || config.APIKey.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing API Key",
			"The provider cannot create the MultiCDN API client without an API key",
		)
		return
	}

	if config.APISecret.IsNull() || config.APISecret.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_secret"),
			"Missing API Secret",
			"The provider cannot create the MultiCDN API client without an API secret",
		)
		return
	}

	// Create the MultiCDN client
	client := NewAPIClient(
		config.BaseURL.ValueString(),
		config.APIKey.ValueString(),
		config.APISecret.ValueString(),
	)

	// Store the client in provider data for use in resources and data sources
	resp.ResourceData = client
	resp.DataSourceData = client
}

// Resources defines the resources implemented in the provider
func (p *multiCDNProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPreferenceResource,
		NewCdnResource,
	}
}

// DataSources defines the data sources implemented in the provider
func (p *multiCDNProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
