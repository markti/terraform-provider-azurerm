package search

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-sdk/data-plane/search/2025-09-01/indexes"
	"github.com/hashicorp/go-azure-sdk/resource-manager/search/2025-05-01/services"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/search/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
)

var _ indexes.SearchIndex

type SearchIndexResource struct{}

var _ sdk.ResourceWithUpdate = SearchIndexResource{}

type SearchIndexModel struct {
	Name                  string                `tfschema:"name"`
	SearchServiceId       string                `tfschema:"search_service_id"`
	Fields                []SearchIndexField    `tfschema:"fields"`
	Analyzers             []SearchIndexAnalyzer `tfschema:"analyzer"`
	CorsOptions           []CorsOptions         `tfschema:"cors_options"`
	ScoringProfiles       []ScoringProfile      `tfschema:"scoring_profile"`
	DefaultScoringProfile string                `tfschema:"default_scoring_profile"`
	ETag                  string                `tfschema:"etag"`
}

type SearchIndexField struct {
	Name           string             `tfschema:"name"`
	Type           string             `tfschema:"type"`
	Key            bool               `tfschema:"key"`
	Searchable     bool               `tfschema:"searchable"`
	Filterable     bool               `tfschema:"filterable"`
	Sortable       bool               `tfschema:"sortable"`
	Facetable      bool               `tfschema:"facetable"`
	Retrievable    bool               `tfschema:"retrievable"`
	Analyzer       string             `tfschema:"analyzer"`
	SearchAnalyzer string             `tfschema:"search_analyzer"`
	IndexAnalyzer  string             `tfschema:"index_analyzer"`
	SynonymMaps    []string           `tfschema:"synonym_maps"`
	Fields         []SearchIndexField `tfschema:"fields"` // For complex types
}

type SearchIndexAnalyzer struct {
	Name      string `tfschema:"name"`
	Type      string `tfschema:"type"`
	Tokenizer string `tfschema:"tokenizer"`
}

type CorsOptions struct {
	AllowedOrigins  []string `tfschema:"allowed_origins"`
	MaxAgeInSeconds int64    `tfschema:"max_age_in_seconds"`
}

type ScoringProfile struct {
	Name string `tfschema:"name"`
	// Add other scoring profile fields as needed
}

func (r SearchIndexResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsNotEmpty,
			Description:  "The name of the search index.",
		},

		"search_service_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: services.ValidateSearchServiceID,
			Description:  "The ID of the Search Service where this index should exist.",
		},

		"fields": {
			Type:        pluginsdk.TypeList,
			Required:    true,
			MinItems:    1,
			Description: "One or more field blocks as defined below.",
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"name": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.StringIsNotEmpty,
					},

					"type": {
						Type:     pluginsdk.TypeString,
						Required: true,
						ValidateFunc: validation.StringInSlice([]string{
							"Edm.String",
							"Edm.Int32",
							"Edm.Int64",
							"Edm.Double",
							"Edm.Boolean",
							"Edm.DateTimeOffset",
							"Edm.GeographyPoint",
							"Collection(Edm.String)",
							"Collection(Edm.Int32)",
							"Collection(Edm.Int64)",
							"Collection(Edm.Double)",
							"Collection(Edm.Boolean)",
							"Collection(Edm.DateTimeOffset)",
							"Collection(Edm.GeographyPoint)",
							"Edm.ComplexType",
						}, false),
					},

					"key": {
						Type:        pluginsdk.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "Whether the field is the key field.",
					},

					"searchable": {
						Type:        pluginsdk.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "Whether the field is searchable.",
					},

					"filterable": {
						Type:        pluginsdk.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "Whether the field is filterable.",
					},

					"sortable": {
						Type:        pluginsdk.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "Whether the field is sortable.",
					},

					"facetable": {
						Type:        pluginsdk.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "Whether the field is facetable.",
					},

					"retrievable": {
						Type:        pluginsdk.TypeBool,
						Optional:    true,
						Default:     true,
						Description: "Whether the field is retrievable.",
					},

					"analyzer": {
						Type:         pluginsdk.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringIsNotEmpty,
						Description:  "The name of the analyzer to use for the field.",
					},

					"search_analyzer": {
						Type:         pluginsdk.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringIsNotEmpty,
						Description:  "The name of the search analyzer to use for the field.",
					},

					"index_analyzer": {
						Type:         pluginsdk.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringIsNotEmpty,
						Description:  "The name of the index analyzer to use for the field.",
					},

					"synonym_maps": {
						Type:        pluginsdk.TypeList,
						Optional:    true,
						Description: "A list of synonym map names to associate with the field.",
						Elem: &pluginsdk.Schema{
							Type:         pluginsdk.TypeString,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},

					"fields": {
						Type:        pluginsdk.TypeList,
						Optional:    true,
						Description: "Nested fields for complex types.",
						Elem:        &pluginsdk.Resource{
							// Recursive schema - same as parent fields
							// You would need to define this recursively or limit depth
						},
					},
				},
			},
		},

		"cors_options": {
			Type:        pluginsdk.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "A cors_options block as defined below.",
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"allowed_origins": {
						Type:     pluginsdk.TypeList,
						Required: true,
						MinItems: 1,
						Elem: &pluginsdk.Schema{
							Type:         pluginsdk.TypeString,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},

					"max_age_in_seconds": {
						Type:         pluginsdk.TypeInt,
						Optional:     true,
						Default:      0,
						ValidateFunc: validation.IntAtLeast(0),
					},
				},
			},
		},

		"default_scoring_profile": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringIsNotEmpty,
			Description:  "The name of the default scoring profile.",
		},
	}
}

func (r SearchIndexResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"etag": {
			Type:        pluginsdk.TypeString,
			Computed:    true,
			Description: "The ETag of the search index.",
		},
	}
}

func (r SearchIndexResource) ModelObject() interface{} {
	return &SearchIndexModel{}
}

func (r SearchIndexResource) ResourceType() string {
	return "azurerm_search_index"
}

func (r SearchIndexResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	// You'll need to create a custom ID parser for data plane resources
	// Since data plane doesn't use ARM resource IDs
	return validation.StringIsNotEmpty
}

func (r SearchIndexResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			var model SearchIndexModel
			if err := metadata.Decode(&model); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			// Parse the search service ID
			searchServiceId, err := services.ParseSearchServiceID(model.SearchServiceId)
			if err != nil {
				return fmt.Errorf("parsing search service ID: %+v", err)
			}

			// Get the search service to retrieve endpoint and keys
			client := metadata.Client.Search.ServicesClient
			searchService, err := client.Get(ctx, *searchServiceId, services.DefaultGetOperationOptions())
			if err != nil {
				return fmt.Errorf("retrieving %s: %+v", searchServiceId, err)
			}

			if searchService.Model == nil {
				return fmt.Errorf("retrieving %s: model was nil", searchServiceId)
			}

			// TODO: Create data plane client using the endpoint
			// You'll need to:
			// 1. Get admin keys using client.ListAdminKeys(ctx, *searchServiceId)
			// 2. Construct the data plane endpoint: https://{serviceName}.search.windows.net
			// 3. Create an Azure AI Search data plane client
			// 4. Build the index schema from model.Fields
			// 5. Call the CreateOrUpdateIndex API

			// Example (you'll need to adapt based on the actual SDK):
			/*
			   keysResp, err := client.ListAdminKeys(ctx, *searchServiceId)
			   if err != nil {
			       return fmt.Errorf("retrieving admin keys: %+v", err)
			   }

			   endpoint := fmt.Sprintf("https://%s.search.windows.net", searchServiceId.SearchServiceName)

			   // Create data plane client with endpoint and key
			   dataPlaneClient, err := azsearch.NewIndexClient(endpoint, keysResp.Model.PrimaryKey, nil)
			   if err != nil {
			       return fmt.Errorf("creating data plane client: %+v", err)
			   }

			   // Build index definition
			   index := azsearch.Index{
			       Name:   pointer.To(model.Name),
			       Fields: expandSearchIndexFields(model.Fields),
			       // ... other properties
			   }

			   // Create the index
			   if _, err := dataPlaneClient.CreateOrUpdate(ctx, model.Name, index, nil); err != nil {
			       return fmt.Errorf("creating search index %q: %+v", model.Name, err)
			   }
			*/

			// Set the resource ID (custom format for data plane)
			id := parse.NewSearchIndexID(*searchServiceId, model.Name)
			metadata.SetID(id)

			return nil
		},
	}
}

func (r SearchIndexResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			// Parse custom ID
			// Get data plane client
			// Call Get Index API
			// Flatten response into model
			// Set state

			return nil
		},
	}
}

func (r SearchIndexResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			// Similar to Create, but handle updates
			return nil
		},
	}
}

func (r SearchIndexResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			// Get data plane client
			// Call Delete Index API
			return nil
		},
	}
}

// Helper functions for expand/flatten
func expandSearchIndexFields(input []SearchIndexField) []interface{} {
	// Convert Terraform model to SDK model
	return nil
}

func flattenSearchIndexFields(input []interface{}) []SearchIndexField {
	// Convert SDK model to Terraform model
	return nil
}
