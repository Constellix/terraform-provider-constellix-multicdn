// Package provider implements the terraform provider resources and data sources
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

	"github.com/constellix/terraform-provider-constellix-multicdn/clients/cdnclient"
)

// Ensure resource implements required interfaces
var (
	_ resource.Resource                = &cdnResource{}
	_ resource.ResourceWithImportState = &cdnResource{}
)

// cdnResource is the resource implementation
type cdnResource struct {
	client *APIClient
}

// cdnResourceModel maps the resource schema to the API client model
type cdnResourceModel struct {
	ResourceID          types.Int64               `tfsdk:"resource_id"`
	ContentType         types.String              `tfsdk:"content_type"`
	Description         types.String              `tfsdk:"description"`
	Version             types.String              `tfsdk:"version"`
	LastUpdated         types.String              `tfsdk:"last_updated"`
	Cdns                []cdnEntryModel           `tfsdk:"cdns"`
	CdnEnablementMap    *cdnEnablementMapModel    `tfsdk:"cdn_enablement_map"`
	TrafficDistribution *trafficDistributionModel `tfsdk:"traffic_distribution"`
}

// cdnEntryModel maps the CDN entry schema
type cdnEntryModel struct {
	CdnName     types.String `tfsdk:"cdn_name"`
	Description types.String `tfsdk:"description"`
	FQDN        types.String `tfsdk:"fqdn"`
	ClientCdnID types.String `tfsdk:"client_cdn_id"`
}

// cdnEnablementMapModel maps the CDN enablement map schema
type cdnEnablementMapModel struct {
	WorldDefault []types.String                       `tfsdk:"world_default"`
	ASNOverrides map[string][]types.String            `tfsdk:"asn_overrides"`
	Continents   map[string]*continentEnablementModel `tfsdk:"continents"`
}

// continentEnablementModel maps the continent enablement schema
type continentEnablementModel struct {
	Default   []types.String                     `tfsdk:"default"`
	Countries map[string]*countryEnablementModel `tfsdk:"countries"`
}

// countryEnablementModel maps the country enablement schema
type countryEnablementModel struct {
	Default      []types.String                         `tfsdk:"default"`
	ASNOverrides map[string][]types.String              `tfsdk:"asn_overrides"`
	Subdivisions map[string]*subdivisionEnablementModel `tfsdk:"subdivisions"`
}

// subdivisionEnablementModel maps the subdivision enablement schema
type subdivisionEnablementModel struct {
	ASNOverrides map[string][]types.String `tfsdk:"asn_overrides"`
}

// trafficDistributionModel maps the traffic distribution schema
type trafficDistributionModel struct {
	WorldDefault *worldDefaultModel                     `tfsdk:"world_default"`
	Continents   map[string]*continentDistributionModel `tfsdk:"continents"`
}

// worldDefaultModel maps the world default schema
type worldDefaultModel struct {
	Options []trafficOptionModel `tfsdk:"options"`
}

// continentDistributionModel maps the continent distribution schema
type continentDistributionModel struct {
	Default   *trafficOptionListModel              `tfsdk:"default"`
	Countries map[string]*countryDistributionModel `tfsdk:"countries"`
}

// countryDistributionModel maps the country distribution schema
type countryDistributionModel struct {
	Default *trafficOptionListModel `tfsdk:"default"`
}

// trafficOptionListModel maps the traffic option list schema
type trafficOptionListModel struct {
	Options []trafficOptionModel `tfsdk:"options"`
}

// trafficOptionModel maps the traffic option schema
type trafficOptionModel struct {
	Name         types.String             `tfsdk:"name"`
	Description  types.String             `tfsdk:"description"`
	EqualWeight  types.Bool               `tfsdk:"equal_weight"`
	Distribution []distributionEntryModel `tfsdk:"distribution"`
}

// distributionEntryModel maps the distribution entry schema
type distributionEntryModel struct {
	ID     types.String `tfsdk:"id"`
	Weight types.Int64  `tfsdk:"weight"`
}

// NewCdnResource creates a new CDN resource
func NewCdnResource() resource.Resource {
	return &cdnResource{}
}

// Metadata returns the resource metadata
func (r *cdnResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "multicdn_cdn_config"
}

// Schema defines the schema for the resource
func (r *cdnResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CDN configuration document",
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.Int64Attribute{
				Description: "Unique ID of the CDN configuration",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"content_type": schema.StringAttribute{
				Description: "Content type of the CDN configuration",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the CDN configuration",
				Optional:    true,
			},
			"version": schema.StringAttribute{
				Description: "Version of the CDN configuration",
				Optional:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of when the configuration was last updated",
				Optional:    true,
			},
			"cdns": schema.ListNestedAttribute{
				Description: "List of CDN provider entries",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cdn_name": schema.StringAttribute{
							Description: "Name of the CDN provider",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the CDN provider entry",
							Optional:    true,
						},
						"fqdn": schema.StringAttribute{
							Description: "Fully qualified domain name for the CDN",
							Required:    true,
						},
						"client_cdn_id": schema.StringAttribute{
							Description: "Client CDN identifier",
							Required:    true,
						},
					},
				},
			},
			"cdn_enablement_map": schema.SingleNestedAttribute{
				Description: "CDN enablement configuration",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"world_default": schema.ListAttribute{
						Description: "Default CDNs enabled globally",
						Required:    true,
						ElementType: types.StringType,
					},
					"asn_overrides": schema.MapAttribute{
						Description: "ASN-specific CDN overrides",
						Required:    true,
						ElementType: types.ListType{
							ElemType: types.StringType,
						},
					},
					"continents": schema.MapNestedAttribute{
						Description: "Continent-specific enablement configurations",
						Required:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"default": schema.ListAttribute{
									Description: "Default CDNs enabled for the continent",
									Required:    true,
									ElementType: types.StringType,
								},
								"countries": schema.MapNestedAttribute{
									Description: "Country-specific enablement configurations",
									Optional:    true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"default": schema.ListAttribute{
												Description: "Default CDNs enabled for the country",
												Required:    true,
												ElementType: types.StringType,
											},
											"asn_overrides": schema.MapAttribute{
												Description: "ASN-specific CDN overrides for the country",
												Required:    true,
												ElementType: types.ListType{
													ElemType: types.StringType,
												},
											},
											"subdivisions": schema.MapNestedAttribute{
												Description: "Subdivision-specific enablement configurations",
												Optional:    true,
												NestedObject: schema.NestedAttributeObject{
													Attributes: map[string]schema.Attribute{
														"asn_overrides": schema.MapAttribute{
															Description: "ASN-specific CDN overrides for the subdivision",
															Required:    true,
															ElementType: types.ListType{
																ElemType: types.StringType,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"traffic_distribution": schema.SingleNestedAttribute{
				Description: "Traffic distribution configuration",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"world_default": schema.SingleNestedAttribute{
						Description: "Global default traffic distribution",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"options": schema.ListNestedAttribute{
								Description: "Traffic distribution options",
								Required:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "Name of the traffic option",
											Required:    true,
										},
										"description": schema.StringAttribute{
											Description: "Description of the traffic option",
											Optional:    true,
										},
										"equal_weight": schema.BoolAttribute{
											Description: "Whether traffic is distributed equally",
											Optional:    true,
										},
										"distribution": schema.ListNestedAttribute{
											Description: "Distribution entries",
											Required:    true,
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"id": schema.StringAttribute{
														Description: "CDN identifier",
														Required:    true,
													},
													"weight": schema.Int64Attribute{
														Description: "Traffic weight",
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
					"continents": schema.MapNestedAttribute{
						Description: "Continent-specific traffic distributions",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"default": schema.SingleNestedAttribute{
									Description: "Default traffic distribution for the continent",
									Optional:    true,
									Attributes: map[string]schema.Attribute{
										"options": schema.ListNestedAttribute{
											Description: "Traffic distribution options",
											Required:    true,
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Name of the traffic option",
														Required:    true,
													},
													"description": schema.StringAttribute{
														Description: "Description of the traffic option",
														Optional:    true,
													},
													"equal_weight": schema.BoolAttribute{
														Description: "Whether traffic is distributed equally",
														Optional:    true,
													},
													"distribution": schema.ListNestedAttribute{
														Description: "Distribution entries",
														Required:    true,
														NestedObject: schema.NestedAttributeObject{
															Attributes: map[string]schema.Attribute{
																"id": schema.StringAttribute{
																	Description: "CDN identifier",
																	Required:    true,
																},
																"weight": schema.Int64Attribute{
																	Description: "Traffic weight",
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
								"countries": schema.MapNestedAttribute{
									Description: "Country-specific traffic distributions",
									Optional:    true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"default": schema.SingleNestedAttribute{
												Description: "Default traffic distribution for the country",
												Optional:    true,
												Attributes: map[string]schema.Attribute{
													"options": schema.ListNestedAttribute{
														Description: "Traffic distribution options",
														Required:    true,
														NestedObject: schema.NestedAttributeObject{
															Attributes: map[string]schema.Attribute{
																"name": schema.StringAttribute{
																	Description: "Name of the traffic option",
																	Required:    true,
																},
																"description": schema.StringAttribute{
																	Description: "Description of the traffic option",
																	Optional:    true,
																},
																"equal_weight": schema.BoolAttribute{
																	Description: "Whether traffic is distributed equally",
																	Optional:    true,
																},
																"distribution": schema.ListNestedAttribute{
																	Description: "Distribution entries",
																	Required:    true,
																	NestedObject: schema.NestedAttributeObject{
																		Attributes: map[string]schema.Attribute{
																			"id": schema.StringAttribute{
																				Description: "CDN identifier",
																				Required:    true,
																			},
																			"weight": schema.Int64Attribute{
																				Description: "Traffic weight",
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
										},
									},
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
func (r *cdnResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new CDN configuration
func (r *cdnResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read the plan data
	var plan cdnResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	apiConfig := r.convertToAPIModel(&plan)

	// Call the API client to create the CDN configuration
	createdConfig, err := r.client.cdn.CreateCdnConfig(ctx, apiConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating CDN Configuration",
			fmt.Sprintf("Unable to create CDN configuration, got error: %s", err),
		)
		return
	}

	// // Fetch the created resource to get all properties including computed ones
	// apiConfig, err = r.client.cdn.GetCdnConfig(ctx, int(plan.ResourceID.ValueInt64()))
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Error Reading CDN Configuration After Create",
	// 		fmt.Sprintf("Unable to read created CDN configuration, got error: %s", err),
	// 	)
	// 	return
	// }

	// Convert API model back to Terraform model
	r.convertFromAPIModel(createdConfig, &plan)

	// Save the data into Terraform state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read reads the CDN configuration from the API
func (r *cdnResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read the current state
	var state cdnResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the resource ID from state
	resourceID := state.ResourceID.ValueInt64()

	// Call the API client to get the CDN configuration
	config, err := r.client.cdn.GetCdnConfig(ctx, resourceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading CDN Configuration",
			fmt.Sprintf("Unable to read CDN configuration ID %d: %s", resourceID, err),
		)
		return
	}

	// Convert API model to Terraform model
	r.convertFromAPIModel(config, &state)

	// Save the updated data into Terraform state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the CDN configuration in the API
func (r *cdnResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read the plan data
	var plan cdnResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the resource ID from plan
	resourceID := plan.ResourceID.ValueInt64()

	// Convert Terraform model to API model
	apiConfig := r.convertToAPIModel(&plan)

	// Call the API client to update the CDN configuration
	updatedConfig, err := r.client.cdn.UpdateCdnConfig(ctx, resourceID, apiConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating CDN Configuration",
			fmt.Sprintf("Unable to update CDN configuration ID %d: %s", resourceID, err),
		)
		return
	}

	// Convert API model back to Terraform model
	r.convertFromAPIModel(updatedConfig, &plan)

	// Save the updated data into Terraform state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the CDN configuration from the API
func (r *cdnResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read the current state
	var state cdnResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the resource ID from state
	resourceID := state.ResourceID.ValueInt64()

	// Call the API client to delete the CDN configuration
	err := r.client.cdn.DeleteCdnConfig(ctx, resourceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting CDN Configuration",
			fmt.Sprintf("Unable to delete CDN configuration ID %d: %s", resourceID, err),
		)
		return
	}
}

// ImportState imports an existing CDN configuration into Terraform state
func (r *cdnResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse the resource ID from the import ID
	resourceID, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing CDN Configuration",
			fmt.Sprintf("Invalid resource ID format: %s. Expected a numeric ID", req.ID),
		)
		return
	}

	// Set the resource ID in the state
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("resource_id"), resourceID)...)
}

// Helper functions to convert between Terraform and API models
func (r *cdnResource) convertToAPIModel(tfModel *cdnResourceModel) *cdnclient.CdnConfiguration {
	apiModel := &cdnclient.CdnConfiguration{
		ResourceID: tfModel.ResourceID.ValueInt64(),
	}

	// Handle optional fields with pointers
	if !tfModel.ContentType.IsNull() && tfModel.ContentType.ValueString() != "" {
		contentType := tfModel.ContentType.ValueString()
		apiModel.ContentType = &contentType
	}

	if !tfModel.Description.IsNull() && tfModel.Description.ValueString() != "" {
		description := tfModel.Description.ValueString()
		apiModel.Description = &description
	}

	if !tfModel.Version.IsNull() && tfModel.Version.ValueString() != "" {
		version := tfModel.Version.ValueString()
		apiModel.Version = &version
	}

	if !tfModel.LastUpdated.IsNull() && tfModel.LastUpdated.ValueString() != "" {
		parsedTime, err := time.Parse(time.RFC3339, tfModel.LastUpdated.ValueString())
		if err == nil {
			apiModel.LastUpdated = &parsedTime
		}
	}

	// Convert CDN entries
	if len(tfModel.Cdns) > 0 {
		apiModel.Cdns = make([]cdnclient.CdnEntry, 0, len(tfModel.Cdns))
		for _, tfEntry := range tfModel.Cdns {
			apiEntry := cdnclient.CdnEntry{
				CdnName:     tfEntry.CdnName.ValueString(),
				FQDN:        tfEntry.FQDN.ValueString(),
				ClientCdnID: tfEntry.ClientCdnID.ValueString(),
			}

			if !tfEntry.Description.IsNull() && tfEntry.Description.ValueString() != "" {
				description := tfEntry.Description.ValueString()
				apiEntry.Description = &description
			}

			apiModel.Cdns = append(apiModel.Cdns, apiEntry)
		}
	}

	// Convert CDN enablement map
	if tfModel.CdnEnablementMap != nil {
		// World default
		if tfModel.CdnEnablementMap.WorldDefault != nil && len(tfModel.CdnEnablementMap.WorldDefault) > 0 {
			apiModel.CdnEnablementMap.WorldDefault = make([]string, 0, len(tfModel.CdnEnablementMap.WorldDefault))
			for _, cdn := range tfModel.CdnEnablementMap.WorldDefault {
				apiModel.CdnEnablementMap.WorldDefault = append(apiModel.CdnEnablementMap.WorldDefault, cdn.ValueString())
			}
		}

		// ASN overrides
		if tfModel.CdnEnablementMap.ASNOverrides != nil && len(tfModel.CdnEnablementMap.ASNOverrides) > 0 {
			apiModel.CdnEnablementMap.ASNOverrides = make(map[string][]string)
			for asn, cdnList := range tfModel.CdnEnablementMap.ASNOverrides {
				apiCdns := make([]string, 0, len(cdnList))
				for _, cdn := range cdnList {
					apiCdns = append(apiCdns, cdn.ValueString())
				}
				apiModel.CdnEnablementMap.ASNOverrides[asn] = apiCdns
			}
		}

		// Continents
		if tfModel.CdnEnablementMap.Continents != nil && len(tfModel.CdnEnablementMap.Continents) > 0 {
			apiModel.CdnEnablementMap.Continents = make(map[string]cdnclient.ContinentEnablement)

			for continent, tfContinent := range tfModel.CdnEnablementMap.Continents {
				apiContinent := cdnclient.ContinentEnablement{}

				// Default
				if tfContinent.Default != nil && len(tfContinent.Default) > 0 {
					apiContinent.Default = make([]string, 0, len(tfContinent.Default))
					for _, cdn := range tfContinent.Default {
						apiContinent.Default = append(apiContinent.Default, cdn.ValueString())
					}
				}

				// Countries
				if tfContinent.Countries != nil && len(tfContinent.Countries) > 0 {
					apiContinent.Countries = make(map[string]cdnclient.CountryEnablement)

					for country, tfCountry := range tfContinent.Countries {
						apiCountry := cdnclient.CountryEnablement{}

						// Country default
						if tfCountry.Default != nil && len(tfCountry.Default) > 0 {
							apiCountry.Default = make([]string, 0, len(tfCountry.Default))
							for _, cdn := range tfCountry.Default {
								apiCountry.Default = append(apiCountry.Default, cdn.ValueString())
							}
						}

						// Country ASN overrides
						if tfCountry.ASNOverrides != nil && len(tfCountry.ASNOverrides) > 0 {
							apiCountry.ASNOverrides = make(map[string][]string)
							for asn, cdnList := range tfCountry.ASNOverrides {
								apiCdns := make([]string, 0, len(cdnList))
								for _, cdn := range cdnList {
									apiCdns = append(apiCdns, cdn.ValueString())
								}
								apiCountry.ASNOverrides[asn] = apiCdns
							}
						}

						// Subdivisions
						if tfCountry.Subdivisions != nil && len(tfCountry.Subdivisions) > 0 {
							apiCountry.Subdivisions = make(map[string]cdnclient.SubdivisionEnablement)

							for subdivision, tfSubdivision := range tfCountry.Subdivisions {
								apiSubdivision := cdnclient.SubdivisionEnablement{}

								// Subdivision ASN overrides
								if tfSubdivision.ASNOverrides != nil && len(tfSubdivision.ASNOverrides) > 0 {
									apiSubdivision.ASNOverrides = make(map[string][]string)
									for asn, cdnList := range tfSubdivision.ASNOverrides {
										apiCdns := make([]string, 0, len(cdnList))
										for _, cdn := range cdnList {
											apiCdns = append(apiCdns, cdn.ValueString())
										}
										apiSubdivision.ASNOverrides[asn] = apiCdns
									}
								}

								apiCountry.Subdivisions[subdivision] = apiSubdivision
							}
						}

						apiContinent.Countries[country] = apiCountry
					}
				}

				apiModel.CdnEnablementMap.Continents[continent] = apiContinent
			}
		}
	}

	// Convert Traffic Distribution
	if tfModel.TrafficDistribution != nil {
		// World default
		if tfModel.TrafficDistribution.WorldDefault != nil && len(tfModel.TrafficDistribution.WorldDefault.Options) > 0 {
			apiWorldDefault := &cdnclient.WorldDefault{
				Options: make([]cdnclient.TrafficOption, 0, len(tfModel.TrafficDistribution.WorldDefault.Options)),
			}

			for _, tfOption := range tfModel.TrafficDistribution.WorldDefault.Options {
				apiOption := cdnclient.TrafficOption{
					Name:         tfOption.Name.ValueString(),
					Distribution: make([]cdnclient.DistributionEntry, 0, len(tfOption.Distribution)),
				}

				if !tfOption.Description.IsNull() && tfOption.Description.ValueString() != "" {
					description := tfOption.Description.ValueString()
					apiOption.Description = &description
				}

				if !tfOption.EqualWeight.IsNull() {
					equalWeight := tfOption.EqualWeight.ValueBool()
					apiOption.EqualWeight = &equalWeight
				}

				for _, tfDist := range tfOption.Distribution {
					apiDist := cdnclient.DistributionEntry{
						ID: tfDist.ID.ValueString(),
					}

					if !tfDist.Weight.IsNull() {
						weight := tfDist.Weight.ValueInt64()
						apiDist.Weight = &weight
					}

					apiOption.Distribution = append(apiOption.Distribution, apiDist)
				}

				apiWorldDefault.Options = append(apiWorldDefault.Options, apiOption)
			}

			apiModel.TrafficDistribution.WorldDefault = apiWorldDefault
		}

		// Continents
		if tfModel.TrafficDistribution.Continents != nil && len(tfModel.TrafficDistribution.Continents) > 0 {
			apiModel.TrafficDistribution.Continents = make(map[string]cdnclient.ContinentDistribution)

			for continent, tfContinent := range tfModel.TrafficDistribution.Continents {
				apiContinent := cdnclient.ContinentDistribution{}

				// Continent default
				if tfContinent.Default != nil && len(tfContinent.Default.Options) > 0 {
					apiOptions := make([]cdnclient.TrafficOption, 0, len(tfContinent.Default.Options))

					for _, tfOption := range tfContinent.Default.Options {
						apiOption := cdnclient.TrafficOption{
							Name:         tfOption.Name.ValueString(),
							Distribution: make([]cdnclient.DistributionEntry, 0, len(tfOption.Distribution)),
						}

						if !tfOption.Description.IsNull() && tfOption.Description.ValueString() != "" {
							description := tfOption.Description.ValueString()
							apiOption.Description = &description
						}

						if !tfOption.EqualWeight.IsNull() {
							equalWeight := tfOption.EqualWeight.ValueBool()
							apiOption.EqualWeight = &equalWeight
						}

						for _, tfDist := range tfOption.Distribution {
							apiDist := cdnclient.DistributionEntry{
								ID: tfDist.ID.ValueString(),
							}

							if !tfDist.Weight.IsNull() {
								weight := tfDist.Weight.ValueInt64()
								apiDist.Weight = &weight
							}

							apiOption.Distribution = append(apiOption.Distribution, apiDist)
						}

						apiOptions = append(apiOptions, apiOption)
					}

					apiContinent.Default = &cdnclient.TrafficOptionList{
						Options: apiOptions,
					}
				}

				// Countries
				if tfContinent.Countries != nil && len(tfContinent.Countries) > 0 {
					apiContinent.Countries = make(map[string]cdnclient.CountryDistribution)

					for country, tfCountry := range tfContinent.Countries {
						apiCountry := cdnclient.CountryDistribution{}

						if tfCountry.Default != nil && len(tfCountry.Default.Options) > 0 {
							apiOptions := make([]cdnclient.TrafficOption, 0, len(tfCountry.Default.Options))

							for _, tfOption := range tfCountry.Default.Options {
								apiOption := cdnclient.TrafficOption{
									Name:         tfOption.Name.ValueString(),
									Distribution: make([]cdnclient.DistributionEntry, 0, len(tfOption.Distribution)),
								}

								if !tfOption.Description.IsNull() && tfOption.Description.ValueString() != "" {
									description := tfOption.Description.ValueString()
									apiOption.Description = &description
								}

								if !tfOption.EqualWeight.IsNull() {
									equalWeight := tfOption.EqualWeight.ValueBool()
									apiOption.EqualWeight = &equalWeight
								}

								for _, tfDist := range tfOption.Distribution {
									apiDist := cdnclient.DistributionEntry{
										ID: tfDist.ID.ValueString(),
									}

									if !tfDist.Weight.IsNull() {
										weight := tfDist.Weight.ValueInt64()
										apiDist.Weight = &weight
									}

									apiOption.Distribution = append(apiOption.Distribution, apiDist)
								}

								apiOptions = append(apiOptions, apiOption)
							}

							apiCountry.Default = &cdnclient.TrafficOptionList{
								Options: apiOptions,
							}
						}

						apiContinent.Countries[country] = apiCountry
					}
				}

				apiModel.TrafficDistribution.Continents[continent] = apiContinent
			}
		}
	}

	return apiModel
}

func (r *cdnResource) convertFromAPIModel(apiModel *cdnclient.CdnConfigurationResponse, tfModel *cdnResourceModel) {
	tfModel.ResourceID = types.Int64Value(apiModel.ResourceID)

	// Handle optional fields with pointers
	if apiModel.ContentType != nil {
		tfModel.ContentType = types.StringValue(*apiModel.ContentType)
	} else {
		tfModel.ContentType = types.StringNull()
	}

	if apiModel.Description != nil {
		tfModel.Description = types.StringValue(*apiModel.Description)
	} else {
		tfModel.Description = types.StringNull()
	}

	if apiModel.Version != nil {
		tfModel.Version = types.StringValue(*apiModel.Version)
	} else {
		tfModel.Version = types.StringNull()
	}

	if apiModel.LastUpdated != nil {
		tfModel.LastUpdated = types.StringValue(apiModel.LastUpdated.Format(time.RFC3339))
	} else {
		tfModel.LastUpdated = types.StringNull()
	}

	// Convert CDN entries
	tfModel.Cdns = make([]cdnEntryModel, 0, len(apiModel.Cdns))
	for _, apiEntry := range apiModel.Cdns {
		tfEntry := cdnEntryModel{
			CdnName:     types.StringValue(apiEntry.CdnName),
			FQDN:        types.StringValue(apiEntry.FQDN),
			ClientCdnID: types.StringValue(apiEntry.ClientCdnID),
		}

		if apiEntry.Description != nil {
			tfEntry.Description = types.StringValue(*apiEntry.Description)
		} else {
			tfEntry.Description = types.StringNull()
		}

		tfModel.Cdns = append(tfModel.Cdns, tfEntry)
	}

	// Convert CDN enablement map
	tfModel.CdnEnablementMap = &cdnEnablementMapModel{
		WorldDefault: make([]types.String, 0), // Initialize as empty slice
		ASNOverrides: nil,                     // Initialize as nil, not empty map
		Continents:   make(map[string]*continentEnablementModel),
	}

	// World default
	if len(apiModel.CdnEnablementMap.WorldDefault) > 0 {
		tfModel.CdnEnablementMap.WorldDefault = make([]types.String, 0, len(apiModel.CdnEnablementMap.WorldDefault))
		for _, cdn := range apiModel.CdnEnablementMap.WorldDefault {
			tfModel.CdnEnablementMap.WorldDefault = append(tfModel.CdnEnablementMap.WorldDefault, types.StringValue(cdn))
		}
	} else {
		tfModel.CdnEnablementMap.WorldDefault = nil // Empty slice becomes nil
	}

	// ASN overrides
	if len(apiModel.CdnEnablementMap.ASNOverrides) > 0 {
		tfModel.CdnEnablementMap.ASNOverrides = make(map[string][]types.String)
		for asn, cdnList := range apiModel.CdnEnablementMap.ASNOverrides {
			tfCdns := make([]types.String, 0, len(cdnList))
			for _, cdn := range cdnList {
				tfCdns = append(tfCdns, types.StringValue(cdn))
			}
			tfModel.CdnEnablementMap.ASNOverrides[asn] = tfCdns
		}
	}

	// Continents
	if len(apiModel.CdnEnablementMap.Continents) > 0 {
		for continent, apiContinent := range apiModel.CdnEnablementMap.Continents {
			tfContinent := &continentEnablementModel{
				Default:   nil, // Initialize as nil instead of empty slice
				Countries: make(map[string]*countryEnablementModel),
			}

			// Default
			if len(apiContinent.Default) > 0 {
				tfContinent.Default = make([]types.String, 0, len(apiContinent.Default))
				for _, cdn := range apiContinent.Default {
					tfContinent.Default = append(tfContinent.Default, types.StringValue(cdn))
				}
			}

			// Countries
			if len(apiContinent.Countries) > 0 {
				for country, apiCountry := range apiContinent.Countries {
					tfCountry := &countryEnablementModel{
						Default:      nil, // Initialize as nil
						ASNOverrides: nil, // Initialize as nil
						Subdivisions: make(map[string]*subdivisionEnablementModel),
					}

					// Country default
					if len(apiCountry.Default) > 0 {
						tfCountry.Default = make([]types.String, 0, len(apiCountry.Default))
						for _, cdn := range apiCountry.Default {
							tfCountry.Default = append(tfCountry.Default, types.StringValue(cdn))
						}
					}

					// Country ASN overrides
					if len(apiCountry.ASNOverrides) > 0 {
						tfCountry.ASNOverrides = make(map[string][]types.String)
						for asn, cdnList := range apiCountry.ASNOverrides {
							tfCdns := make([]types.String, 0, len(cdnList))
							for _, cdn := range cdnList {
								tfCdns = append(tfCdns, types.StringValue(cdn))
							}
							tfCountry.ASNOverrides[asn] = tfCdns
						}
					}

					// Subdivisions
					if len(apiCountry.Subdivisions) > 0 {
						for subdivision, apiSubdivision := range apiCountry.Subdivisions {
							tfSubdivision := &subdivisionEnablementModel{
								ASNOverrides: nil, // Initialize as nil
							}

							// Subdivision ASN overrides
							if len(apiSubdivision.ASNOverrides) > 0 {
								tfSubdivision.ASNOverrides = make(map[string][]types.String)
								for asn, cdnList := range apiSubdivision.ASNOverrides {
									tfCdns := make([]types.String, 0, len(cdnList))
									for _, cdn := range cdnList {
										tfCdns = append(tfCdns, types.StringValue(cdn))
									}
									tfSubdivision.ASNOverrides[asn] = tfCdns
								}
							}

							tfCountry.Subdivisions[subdivision] = tfSubdivision
						}
					} else {
						tfCountry.Subdivisions = nil // Convert empty map to nil
					}

					tfContinent.Countries[country] = tfCountry
				}
			} else {
				tfContinent.Countries = nil // Convert empty map to nil
			}

			tfModel.CdnEnablementMap.Continents[continent] = tfContinent
		}
	} else {
		tfModel.CdnEnablementMap.Continents = nil // Convert empty map to nil
	}

	// Convert Traffic Distribution
	tfModel.TrafficDistribution = &trafficDistributionModel{
		WorldDefault: nil,
		Continents:   nil, // Initialize as nil, not empty map
	}

	// World default
	if apiModel.TrafficDistribution.WorldDefault != nil && len(apiModel.TrafficDistribution.WorldDefault.Options) > 0 {
		tfWorldDefault := &worldDefaultModel{
			Options: make([]trafficOptionModel, 0, len(apiModel.TrafficDistribution.WorldDefault.Options)),
		}

		for _, apiOption := range apiModel.TrafficDistribution.WorldDefault.Options {
			tfOption := trafficOptionModel{
				Name:         types.StringValue(apiOption.Name),
				Distribution: make([]distributionEntryModel, 0, len(apiOption.Distribution)),
			}

			if apiOption.Description != nil {
				tfOption.Description = types.StringValue(*apiOption.Description)
			} else {
				tfOption.Description = types.StringNull()
			}

			if apiOption.EqualWeight != nil {
				tfOption.EqualWeight = types.BoolValue(*apiOption.EqualWeight)
			} else {
				tfOption.EqualWeight = types.BoolNull()
			}

			for _, apiDist := range apiOption.Distribution {
				tfDist := distributionEntryModel{
					ID: types.StringValue(apiDist.ID),
				}

				if apiDist.Weight != nil {
					tfDist.Weight = types.Int64Value(*apiDist.Weight)
				} else {
					tfDist.Weight = types.Int64Null()
				}

				tfOption.Distribution = append(tfOption.Distribution, tfDist)
			}

			tfWorldDefault.Options = append(tfWorldDefault.Options, tfOption)
		}

		tfModel.TrafficDistribution.WorldDefault = tfWorldDefault
	}

	// Continents
	if len(apiModel.TrafficDistribution.Continents) > 0 {
		tfModel.TrafficDistribution.Continents = make(map[string]*continentDistributionModel)

		for continent, apiContinent := range apiModel.TrafficDistribution.Continents {
			tfContinent := &continentDistributionModel{
				Default:   nil,
				Countries: nil, // Initialize as nil
			}

			// Continent default
			if apiContinent.Default != nil && len(apiContinent.Default.Options) > 0 {
				tfOptions := make([]trafficOptionModel, 0, len(apiContinent.Default.Options))

				for _, apiOption := range apiContinent.Default.Options {
					tfOption := trafficOptionModel{
						Name:         types.StringValue(apiOption.Name),
						Distribution: make([]distributionEntryModel, 0, len(apiOption.Distribution)),
					}

					if apiOption.Description != nil {
						tfOption.Description = types.StringValue(*apiOption.Description)
					} else {
						tfOption.Description = types.StringNull()
					}

					if apiOption.EqualWeight != nil {
						tfOption.EqualWeight = types.BoolValue(*apiOption.EqualWeight)
					} else {
						tfOption.EqualWeight = types.BoolNull()
					}

					for _, apiDist := range apiOption.Distribution {
						tfDist := distributionEntryModel{
							ID: types.StringValue(apiDist.ID),
						}

						if apiDist.Weight != nil {
							tfDist.Weight = types.Int64Value(*apiDist.Weight)
						} else {
							tfDist.Weight = types.Int64Null()
						}

						tfOption.Distribution = append(tfOption.Distribution, tfDist)
					}

					tfOptions = append(tfOptions, tfOption)
				}

				tfContinent.Default = &trafficOptionListModel{
					Options: tfOptions,
				}
			}

			// Countries
			if len(apiContinent.Countries) > 0 {
				tfContinent.Countries = make(map[string]*countryDistributionModel)

				for country, apiCountry := range apiContinent.Countries {
					tfCountry := &countryDistributionModel{
						Default: nil,
					}

					if apiCountry.Default != nil && len(apiCountry.Default.Options) > 0 {
						tfOptions := make([]trafficOptionModel, 0, len(apiCountry.Default.Options))

						for _, apiOption := range apiCountry.Default.Options {
							tfOption := trafficOptionModel{
								Name:         types.StringValue(apiOption.Name),
								Distribution: make([]distributionEntryModel, 0, len(apiOption.Distribution)),
							}

							if apiOption.Description != nil {
								tfOption.Description = types.StringValue(*apiOption.Description)
							} else {
								tfOption.Description = types.StringNull()
							}

							if apiOption.EqualWeight != nil {
								tfOption.EqualWeight = types.BoolValue(*apiOption.EqualWeight)
							} else {
								tfOption.EqualWeight = types.BoolNull()
							}

							for _, apiDist := range apiOption.Distribution {
								tfDist := distributionEntryModel{
									ID: types.StringValue(apiDist.ID),
								}

								if apiDist.Weight != nil {
									tfDist.Weight = types.Int64Value(*apiDist.Weight)
								} else {
									tfDist.Weight = types.Int64Null()
								}

								tfOption.Distribution = append(tfOption.Distribution, tfDist)
							}

							tfOptions = append(tfOptions, tfOption)
						}

						tfCountry.Default = &trafficOptionListModel{
							Options: tfOptions,
						}
					}

					tfContinent.Countries[country] = tfCountry
				}
			}

			tfModel.TrafficDistribution.Continents[continent] = tfContinent
		}
	}
}
