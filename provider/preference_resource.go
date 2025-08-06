package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/constellix/terraform-provider-multicdn/clients/preferenceclient"
)

// Ensure resource implements required interfaces
var (
	_ resource.Resource                = &preferenceResource{}
	_ resource.ResourceWithImportState = &preferenceResource{}
)

// preferenceResource is the resource implementation
type preferenceResource struct {
	client *APIClient
}

// preferenceResourceModel maps the resource schema to the API client model
type preferenceResourceModel struct {
	ResourceID                  types.Int64                       `tfsdk:"resource_id"`
	ContentType                 types.String                      `tfsdk:"content_type"`
	Description                 types.String                      `tfsdk:"description"`
	Version                     types.String                      `tfsdk:"version"`
	LastUpdated                 types.String                      `tfsdk:"last_updated"`
	AvailabilityThresholds      *availabilityThresholdsModel      `tfsdk:"availability_thresholds"`
	PerformanceFiltering        *performanceFilteringModel        `tfsdk:"performance_filtering"`
	EnabledSubdivisionCountries *enabledSubdivisionCountriesModel `tfsdk:"enabled_subdivision_countries"`
}

// availabilityThresholdsModel maps the AvailabilityThresholds schema
type availabilityThresholdsModel struct {
	World      types.Float64                       `tfsdk:"world"`
	Continents map[string]*continentThresholdModel `tfsdk:"continents"`
}

// continentThresholdModel maps the ContinentThreshold schema
type continentThresholdModel struct {
	Default   types.Float64            `tfsdk:"default"`
	Countries map[string]types.Float64 `tfsdk:"countries"`
}

// performanceFilteringModel maps the PerformanceFiltering schema
type performanceFilteringModel struct {
	World      *performanceConfigModel                     `tfsdk:"world"`
	Continents map[string]*continentPerformanceConfigModel `tfsdk:"continents"`
}

// performanceConfigModel maps the PerformanceConfig schema
type performanceConfigModel struct {
	Mode              types.String  `tfsdk:"mode"`
	RelativeThreshold types.Float64 `tfsdk:"relative_threshold"`
}

// continentPerformanceConfigModel maps the ContinentPerformanceConfig schema
type continentPerformanceConfigModel struct {
	Mode              types.String                       `tfsdk:"mode"`
	RelativeThreshold types.Float64                      `tfsdk:"relative_threshold"`
	Countries         map[string]*performanceConfigModel `tfsdk:"countries"`
}

// enabledSubdivisionCountriesModel maps the EnabledSubdivisionCountries schema
type enabledSubdivisionCountriesModel struct {
	Continents map[string]*continentSubdivisionsModel `tfsdk:"continents"`
}

// continentSubdivisionsModel maps the ContinentSubdivisions schema
type continentSubdivisionsModel struct {
	Countries []types.String `tfsdk:"countries"`
}

// NewPreferenceResource creates a new preference resource
func NewPreferenceResource() resource.Resource {
	return &preferenceResource{}
}

// Metadata returns the resource metadata
func (r *preferenceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_preference"
}

// Schema defines the schema for the resource
func (r *preferenceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CDN preference configuration",
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.Int64Attribute{
				Description: "Unique ID of the CDN preference configuration",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"content_type": schema.StringAttribute{
				Description: "Content type of the CDN preference configuration",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the CDN preference configuration",
				Optional:    true,
			},
			"version": schema.StringAttribute{
				Description: "Version of the CDN preference configuration",
				Optional:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of when the configuration was last updated",
				Optional:    true,
			},
			"availability_thresholds": schema.SingleNestedAttribute{
				Description: "Availability thresholds configuration",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"world": schema.Float64Attribute{
						Description: "Global availability threshold",
						Optional:    true,
					},
					"continents": schema.MapNestedAttribute{
						Description: "Continent-specific availability thresholds",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"default": schema.Float64Attribute{
									Description: "Default threshold for the continent",
									Required:    true,
								},
								"countries": schema.MapAttribute{
									Description: "Country-specific thresholds",
									Optional:    true,
									ElementType: types.Float64Type,
								},
							},
						},
					},
				},
			},
			"performance_filtering": schema.SingleNestedAttribute{
				Description: "Performance filtering configuration",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"world": schema.SingleNestedAttribute{
						Description: "Global performance filtering configuration",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"mode": schema.StringAttribute{
								Description: "Performance filtering mode (relative or absolute)",
								Optional:    true,
							},
							"relative_threshold": schema.Float64Attribute{
								Description: "Relative performance threshold",
								Optional:    true,
							},
						},
					},
					"continents": schema.MapNestedAttribute{
						Description: "Continent-specific performance configurations",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"mode": schema.StringAttribute{
									Description: "Performance filtering mode for the continent",
									Optional:    true,
								},
								"relative_threshold": schema.Float64Attribute{
									Description: "Relative performance threshold for the continent",
									Optional:    true,
								},
								"countries": schema.MapNestedAttribute{
									Description: "Country-specific performance configurations",
									Optional:    true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"mode": schema.StringAttribute{
												Description: "Performance filtering mode for the country",
												Optional:    true,
											},
											"relative_threshold": schema.Float64Attribute{
												Description: "Relative performance threshold for the country",
												Optional:    true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"enabled_subdivision_countries": schema.SingleNestedAttribute{
				Description: "Configuration for countries with enabled subdivisions",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"continents": schema.MapNestedAttribute{
						Description: "Continent-specific subdivision configurations",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"countries": schema.ListAttribute{
									Description: "List of countries with enabled subdivisions",
									Optional:    true,
									ElementType: types.StringType,
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure configures the resource with the provider client
func (r *preferenceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *APIClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates a new preference configuration
func (r *preferenceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read the plan data
	var plan preferenceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	apiPreference := r.convertToAPIModel(&plan)

	// Call the API client to create the preference
	err := r.client.preference.CreatePreference(ctx, apiPreference)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Preference",
			fmt.Sprintf("Unable to create preference, got error: %s", err),
		)
		return
	}

	// Fetch the created resource to get all properties including computed ones
	apiPreference, err = r.client.preference.GetPreference(ctx, int(plan.ResourceID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Preference After Create",
			fmt.Sprintf("Unable to read created preference, got error: %s", err),
		)
		return
	}

	// Convert API model back to Terraform model
	r.convertFromAPIModel(apiPreference, &plan)

	// Save the data into Terraform state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read reads the preference configuration from the API
func (r *preferenceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read the current state
	var state preferenceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the resource ID from state
	resourceID := int(state.ResourceID.ValueInt64())

	// Call the API client to get the preference
	preference, err := r.client.preference.GetPreference(ctx, resourceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Preference",
			fmt.Sprintf("Unable to read preference ID %d: %s", resourceID, err),
		)
		return
	}

	// Convert API model to Terraform model
	r.convertFromAPIModel(preference, &state)

	// Save the updated data into Terraform state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the preference configuration in the API
func (r *preferenceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read the plan data
	var plan preferenceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the resource ID from plan
	resourceID := int(plan.ResourceID.ValueInt64())

	// Convert Terraform model to API model
	apiPreference := r.convertToAPIModel(&plan)

	// Call the API client to update the preference
	err := r.client.preference.UpdatePreference(ctx, resourceID, apiPreference)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Preference",
			fmt.Sprintf("Unable to update preference ID %d: %s", resourceID, err),
		)
		return
	}

	// Fetch the updated resource to get all properties including computed ones
	apiPreference, err = r.client.preference.GetPreference(ctx, resourceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Preference After Update",
			fmt.Sprintf("Unable to read updated preference ID %d: %s", resourceID, err),
		)
		return
	}

	// Convert API model back to Terraform model
	r.convertFromAPIModel(apiPreference, &plan)

	// Save the updated data into Terraform state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the preference configuration from the API
func (r *preferenceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read the current state
	var state preferenceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the resource ID from state
	resourceID := int(state.ResourceID.ValueInt64())

	// Call the API client to delete the preference
	err := r.client.preference.DeletePreference(ctx, resourceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Preference",
			fmt.Sprintf("Unable to delete preference ID %d: %s", resourceID, err),
		)
		return
	}
}

// ImportState imports an existing preference configuration into Terraform state
func (r *preferenceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse the resource ID from the import ID
	resourceID, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Preference",
			fmt.Sprintf("Invalid resource ID format: %s. Expected a numeric ID", req.ID),
		)
		return
	}

	// Set the resource ID in the state
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("resource_id"), resourceID)...)
}

// Helper functions to convert between Terraform and API models
func (r *preferenceResource) convertToAPIModel(tfModel *preferenceResourceModel) *preferenceclient.Preference {
	apiModel := &preferenceclient.Preference{
		ResourceID: int(tfModel.ResourceID.ValueInt64()),
	}

	if !tfModel.ContentType.IsNull() {
		apiModel.ContentType = tfModel.ContentType.ValueString()
	}

	if !tfModel.Description.IsNull() {
		apiModel.Description = tfModel.Description.ValueString()
	}

	if !tfModel.Version.IsNull() {
		apiModel.Version = tfModel.Version.ValueString()
	}

	if !tfModel.LastUpdated.IsNull() {
		parsedTime, err := time.Parse(time.RFC3339, tfModel.LastUpdated.ValueString())
		if err == nil {
			apiModel.LastUpdated = parsedTime
		}
	}

	// Convert AvailabilityThresholds
	if tfModel.AvailabilityThresholds != nil {
		if !tfModel.AvailabilityThresholds.World.IsNull() {
			apiModel.AvailabilityThresholds.World = tfModel.AvailabilityThresholds.World.ValueFloat64()
		}

		if tfModel.AvailabilityThresholds.Continents != nil {
			apiModel.AvailabilityThresholds.Continents = make(map[string]preferenceclient.ContinentThreshold)

			for continent, tfContinent := range tfModel.AvailabilityThresholds.Continents {
				apiContinent := preferenceclient.ContinentThreshold{}

				if !tfContinent.Default.IsNull() {
					apiContinent.Default = tfContinent.Default.ValueFloat64()
				}

				if tfContinent.Countries != nil {
					apiContinent.Countries = make(map[string]float64)
					for country, threshold := range tfContinent.Countries {
						if !threshold.IsNull() {
							apiContinent.Countries[country] = threshold.ValueFloat64()
						}
					}
				}

				apiModel.AvailabilityThresholds.Continents[continent] = apiContinent
			}
		}
	}

	// Convert PerformanceFiltering
	if tfModel.PerformanceFiltering != nil {
		// Convert World performance config
		if tfModel.PerformanceFiltering.World != nil {
			if !tfModel.PerformanceFiltering.World.Mode.IsNull() {
				apiModel.PerformanceFiltering.World.Mode = tfModel.PerformanceFiltering.World.Mode.ValueString()
			}

			if !tfModel.PerformanceFiltering.World.RelativeThreshold.IsNull() {
				apiModel.PerformanceFiltering.World.RelativeThreshold = tfModel.PerformanceFiltering.World.RelativeThreshold.ValueFloat64()
			}
		}

		// Convert Continents performance config
		if tfModel.PerformanceFiltering.Continents != nil {
			apiModel.PerformanceFiltering.Continents = make(map[string]preferenceclient.ContinentPerformanceConfig)

			for continent, tfContinent := range tfModel.PerformanceFiltering.Continents {
				apiContinent := preferenceclient.ContinentPerformanceConfig{}

				if !tfContinent.Mode.IsNull() {
					apiContinent.Mode = tfContinent.Mode.ValueString()
				}

				if !tfContinent.RelativeThreshold.IsNull() {
					apiContinent.RelativeThreshold = tfContinent.RelativeThreshold.ValueFloat64()
				}

				// Convert Countries performance config
				if tfContinent.Countries != nil {
					apiContinent.Countries = make(map[string]preferenceclient.PerformanceConfig)

					for country, tfCountry := range tfContinent.Countries {
						apiCountry := preferenceclient.PerformanceConfig{}

						if !tfCountry.Mode.IsNull() {
							apiCountry.Mode = tfCountry.Mode.ValueString()
						}

						if !tfCountry.RelativeThreshold.IsNull() {
							apiCountry.RelativeThreshold = tfCountry.RelativeThreshold.ValueFloat64()
						}

						apiContinent.Countries[country] = apiCountry
					}
				}

				apiModel.PerformanceFiltering.Continents[continent] = apiContinent
			}
		}
	}

	// Convert EnabledSubdivisionCountries
	if tfModel.EnabledSubdivisionCountries != nil && tfModel.EnabledSubdivisionCountries.Continents != nil {
		apiModel.EnabledSubdivisionCountries.Continents = make(map[string]preferenceclient.ContinentSubdivisions)

		for continent, tfContinent := range tfModel.EnabledSubdivisionCountries.Continents {
			apiContinent := preferenceclient.ContinentSubdivisions{}

			if tfContinent.Countries != nil {
				apiContinent.Countries = make([]string, 0, len(tfContinent.Countries))

				for _, country := range tfContinent.Countries {
					if !country.IsNull() {
						apiContinent.Countries = append(apiContinent.Countries, country.ValueString())
					}
				}
			}

			apiModel.EnabledSubdivisionCountries.Continents[continent] = apiContinent
		}
	}

	return apiModel
}

func (r *preferenceResource) convertFromAPIModel(apiModel *preferenceclient.Preference, tfModel *preferenceResourceModel) {
	tfModel.ResourceID = types.Int64Value(int64(apiModel.ResourceID))

	// Fix for empty strings - use null values instead of empty strings
	if apiModel.ContentType != "" {
		tfModel.ContentType = types.StringValue(apiModel.ContentType)
	} else {
		tfModel.ContentType = types.StringNull()
	}

	if apiModel.Description != "" {
		tfModel.Description = types.StringValue(apiModel.Description)
	} else {
		tfModel.Description = types.StringNull()
	}

	if apiModel.Version != "" {
		tfModel.Version = types.StringValue(apiModel.Version)
	} else {
		tfModel.Version = types.StringNull()
	}

	if !apiModel.LastUpdated.IsZero() {
		tfModel.LastUpdated = types.StringValue(apiModel.LastUpdated.Format(time.RFC3339))
	} else {
		tfModel.LastUpdated = types.StringNull()
	}

	// Convert AvailabilityThresholds
	tfModel.AvailabilityThresholds = &availabilityThresholdsModel{
		World:      types.Float64Value(apiModel.AvailabilityThresholds.World),
		Continents: nil, // Initialize as nil, not empty map
	}

	// Only initialize continents map if there are actual continents
	if len(apiModel.AvailabilityThresholds.Continents) > 0 {
		tfModel.AvailabilityThresholds.Continents = make(map[string]*continentThresholdModel)

		for continent, apiContinent := range apiModel.AvailabilityThresholds.Continents {
			tfContinent := &continentThresholdModel{
				Default:   types.Float64Value(apiContinent.Default),
				Countries: nil, // Initialize as nil, not empty map
			}

			// Only initialize countries map if there are actual countries
			if len(apiContinent.Countries) > 0 {
				tfContinent.Countries = make(map[string]types.Float64)

				for country, threshold := range apiContinent.Countries {
					tfContinent.Countries[country] = types.Float64Value(threshold)
				}
			}

			tfModel.AvailabilityThresholds.Continents[continent] = tfContinent
		}
	}

	// Convert PerformanceFiltering
	tfModel.PerformanceFiltering = &performanceFilteringModel{
		World:      &performanceConfigModel{},
		Continents: nil, // Initialize as nil, not empty map
	}

	// Convert World performance config
	tfModel.PerformanceFiltering.World.Mode = types.StringValue(apiModel.PerformanceFiltering.World.Mode)
	tfModel.PerformanceFiltering.World.RelativeThreshold = types.Float64Value(apiModel.PerformanceFiltering.World.RelativeThreshold)

	// Only initialize continents map if there are actual continents
	if len(apiModel.PerformanceFiltering.Continents) > 0 {
		tfModel.PerformanceFiltering.Continents = make(map[string]*continentPerformanceConfigModel)

		// Convert Continents performance config
		for continent, apiContinent := range apiModel.PerformanceFiltering.Continents {
			tfContinent := &continentPerformanceConfigModel{
				Mode:              types.StringValue(apiContinent.Mode),
				RelativeThreshold: types.Float64Value(apiContinent.RelativeThreshold),
				Countries:         nil, // Initialize as nil, not empty map
			}

			// Only initialize countries map if there are actual countries
			if len(apiContinent.Countries) > 0 {
				tfContinent.Countries = make(map[string]*performanceConfigModel)

				// Convert Countries performance config
				for country, apiCountry := range apiContinent.Countries {
					tfCountry := &performanceConfigModel{
						Mode:              types.StringValue(apiCountry.Mode),
						RelativeThreshold: types.Float64Value(apiCountry.RelativeThreshold),
					}

					tfContinent.Countries[country] = tfCountry
				}
			}

			tfModel.PerformanceFiltering.Continents[continent] = tfContinent
		}
	}

	// Convert EnabledSubdivisionCountries - FIX: handle empty case properly
	tfModel.EnabledSubdivisionCountries = &enabledSubdivisionCountriesModel{
		Continents: make(map[string]*continentSubdivisionsModel), // Always initialize to empty map, not nil
	}

	// Add any continents if they exist
	if len(apiModel.EnabledSubdivisionCountries.Continents) > 0 {

		for continent, apiContinent := range apiModel.EnabledSubdivisionCountries.Continents {
			// Only create a model if there are actual countries
			if len(apiContinent.Countries) > 0 {
				tfContinent := &continentSubdivisionsModel{
					Countries: make([]types.String, 0, len(apiContinent.Countries)),
				}

				for _, country := range apiContinent.Countries {
					tfContinent.Countries = append(tfContinent.Countries, types.StringValue(country))
				}

				tfModel.EnabledSubdivisionCountries.Continents[continent] = tfContinent
			}
		}
	}
}
