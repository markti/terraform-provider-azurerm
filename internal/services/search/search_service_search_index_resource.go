package search

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-sdk/data-plane/search/2025-09-01/indexes"
	"github.com/hashicorp/go-azure-sdk/resource-manager/search/2025-05-01/services"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/search/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
)

type SearchIndexResource struct{}

var _ sdk.ResourceWithUpdate = SearchIndexResource{}
var _ indexes.SearchIndex // Keep the import

type SearchIndexModel struct {
	Name                  string             `tfschema:"name"`
	SearchServiceId       string             `tfschema:"search_service_id"`
	Fields                []SearchIndexField `tfschema:"fields"`
	CorsOptions           []CorsOptions      `tfschema:"cors_options"`
	DefaultScoringProfile string             `tfschema:"default_scoring_profile"`
	ETag                  string             `tfschema:"etag"`
}

type SearchIndexField struct {
	Name           string   `tfschema:"name"`
	Type           string   `tfschema:"type"`
	Key            bool     `tfschema:"key"`
	Searchable     bool     `tfschema:"searchable"`
	Filterable     bool     `tfschema:"filterable"`
	Sortable       bool     `tfschema:"sortable"`
	Facetable      bool     `tfschema:"facetable"`
	Retrievable    bool     `tfschema:"retrievable"`
	Analyzer       string   `tfschema:"analyzer"`
	SearchAnalyzer string   `tfschema:"search_analyzer"`
	IndexAnalyzer  string   `tfschema:"index_analyzer"`
	SynonymMaps    []string `tfschema:"synonym_maps"`
	// Remove recursive Fields to avoid memory leaks
	// Complex types should be defined separately if needed
}

type CorsOptions struct {
	AllowedOrigins  []string `tfschema:"allowed_origins"`
	MaxAgeInSeconds int64    `tfschema:"max_age_in_seconds"`
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
	return parse.ValidateSearchIndexID
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

			// Get the endpoint from the environment
			domainSuffix, ok := metadata.Client.Account.Environment.Search.DomainSuffix()
			if !ok {
				return errors.New("could not determine Search domain suffix for the current environment")
			}
			endpoint := fmt.Sprintf("https://%s.%s", searchServiceId.SearchServiceName, *domainSuffix)

			// Use the pre-configured data plane client
			client := metadata.Client.Search.SearchDataPlaneClient.Indexes.Clone(endpoint)

			// Check if index already exists
			indexId := indexes.IndexId{
				IndexName: model.Name,
			}
			existing, err := client.Get(ctx, indexId, indexes.DefaultGetOperationOptions())
			if err != nil {
				if !response.WasNotFound(existing.HttpResponse) {
					return fmt.Errorf("checking for existing index %q: %+v", model.Name, err)
				}
			}
			if !response.WasNotFound(existing.HttpResponse) {
				return metadata.ResourceRequiresImport(r.ResourceType(), parse.NewSearchIndexID(*searchServiceId, model.Name))
			}

			// Build index definition
			indexDef := indexes.SearchIndex{
				Name:   model.Name,
				Fields: expandSearchIndexFields(model.Fields),
			}

			if len(model.CorsOptions) > 0 {
				indexDef.CorsOptions = expandCorsOptions(model.CorsOptions)
			}

			if model.DefaultScoringProfile != "" {
				indexDef.DefaultScoringProfile = &model.DefaultScoringProfile
			}

			// Create the index
			if _, err := client.CreateOrUpdate(ctx, indexId, indexDef, indexes.DefaultCreateOrUpdateOperationOptions()); err != nil {
				return fmt.Errorf("creating search index %q: %+v", model.Name, err)
			}

			// Set the resource ID
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
			id, err := parse.SearchIndexID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			// Get the endpoint
			domainSuffix, ok := metadata.Client.Account.Environment.Search.DomainSuffix()
			if !ok {
				return errors.New("could not determine Search domain suffix for the current environment")
			}
			endpoint := fmt.Sprintf("https://%s.%s", id.SearchServiceId.SearchServiceName, *domainSuffix)

			// Use the pre-configured data plane client
			client := metadata.Client.Search.SearchDataPlaneClient.Indexes.Clone(endpoint)

			// Get the index
			indexId := indexes.IndexId{
				IndexName: id.IndexName,
			}
			resp, err := client.Get(ctx, indexId, indexes.DefaultGetOperationOptions())
			if err != nil {
				if response.WasNotFound(resp.HttpResponse) {
					return metadata.MarkAsGone(id)
				}
				return fmt.Errorf("retrieving search index %q: %+v", id.IndexName, err)
			}

			if resp.Model == nil {
				return fmt.Errorf("retrieving search index %q: model was nil", id.IndexName)
			}

			// Flatten into state
			state := SearchIndexModel{
				Name:            resp.Model.Name,
				SearchServiceId: id.SearchServiceId.ID(),
				Fields:          flattenSearchIndexFields(resp.Model.Fields),
			}

			if resp.Model.CorsOptions != nil {
				state.CorsOptions = flattenCorsOptions(resp.Model.CorsOptions)
			}

			if resp.Model.DefaultScoringProfile != nil {
				state.DefaultScoringProfile = pointer.From(resp.Model.DefaultScoringProfile)
			}

			return metadata.Encode(&state)
		},
	}
}

func (r SearchIndexResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			id, err := parse.SearchIndexID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			var model SearchIndexModel
			if err := metadata.Decode(&model); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			// Get the endpoint
			domainSuffix, ok := metadata.Client.Account.Environment.Search.DomainSuffix()
			if !ok {
				return errors.New("could not determine Search domain suffix for the current environment")
			}
			endpoint := fmt.Sprintf("https://%s.%s", id.SearchServiceId.SearchServiceName, *domainSuffix)

			// Use the pre-configured data plane client
			client := metadata.Client.Search.SearchDataPlaneClient.Indexes.Clone(endpoint)

			// Build updated index
			indexDef := indexes.SearchIndex{
				Name:   model.Name,
				Fields: expandSearchIndexFields(model.Fields),
			}

			if len(model.CorsOptions) > 0 {
				indexDef.CorsOptions = expandCorsOptions(model.CorsOptions)
			}

			if model.DefaultScoringProfile != "" {
				indexDef.DefaultScoringProfile = &model.DefaultScoringProfile
			}

			// Update the index
			indexId := indexes.IndexId{
				IndexName: id.IndexName,
			}
			if _, err := client.CreateOrUpdate(ctx, indexId, indexDef, indexes.DefaultCreateOrUpdateOperationOptions()); err != nil {
				return fmt.Errorf("updating search index %q: %+v", id.IndexName, err)
			}

			return nil
		},
	}
}

func (r SearchIndexResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			id, err := parse.SearchIndexID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			// Get the endpoint
			domainSuffix, ok := metadata.Client.Account.Environment.Search.DomainSuffix()
			if !ok {
				return errors.New("could not determine Search domain suffix for the current environment")
			}
			endpoint := fmt.Sprintf("https://%s.%s", id.SearchServiceId.SearchServiceName, *domainSuffix)

			// Use the pre-configured data plane client
			client := metadata.Client.Search.SearchDataPlaneClient.Indexes.Clone(endpoint)

			// Delete the index
			indexId := indexes.IndexId{
				IndexName: id.IndexName,
			}
			if _, err := client.Delete(ctx, indexId, indexes.DefaultDeleteOperationOptions()); err != nil {
				return fmt.Errorf("deleting search index %q: %+v", id.IndexName, err)
			}

			return nil
		},
	}
}

// Helper functions
func expandSearchIndexFields(input []SearchIndexField) []indexes.SearchField {
	if len(input) == 0 {
		return nil
	}

	results := make([]indexes.SearchField, 0, len(input))
	for _, v := range input {
		field := indexes.SearchField{
			Name: v.Name,
			Type: indexes.SearchFieldDataType(v.Type),
		}

		if v.Key {
			field.Key = pointer.To(true)
		}
		if v.Searchable {
			field.Searchable = pointer.To(true)
		}
		if v.Filterable {
			field.Filterable = pointer.To(true)
		}
		if v.Sortable {
			field.Sortable = pointer.To(true)
		}
		if v.Facetable {
			field.Facetable = pointer.To(true)
		}
		if v.Retrievable {
			field.Retrievable = pointer.To(true)
		}
		if v.Analyzer != "" {
			analyzer := indexes.LexicalAnalyzerName(v.Analyzer)
			field.Analyzer = &analyzer
		}
		if v.SearchAnalyzer != "" {
			searchAnalyzer := indexes.LexicalAnalyzerName(v.SearchAnalyzer)
			field.SearchAnalyzer = &searchAnalyzer
		}
		if v.IndexAnalyzer != "" {
			indexAnalyzer := indexes.LexicalAnalyzerName(v.IndexAnalyzer)
			field.IndexAnalyzer = &indexAnalyzer
		}
		if len(v.SynonymMaps) > 0 {
			field.SynonymMaps = &v.SynonymMaps
		}

		results = append(results, field)
	}

	return results // Changed from &results
}

func expandCorsOptions(input []CorsOptions) *indexes.CorsOptions {
	if len(input) == 0 {
		return nil
	}

	cors := input[0]
	return &indexes.CorsOptions{
		AllowedOrigins:  cors.AllowedOrigins, // Changed from &cors.AllowedOrigins
		MaxAgeInSeconds: pointer.To(cors.MaxAgeInSeconds),
	}
}

func flattenSearchIndexFields(input []indexes.SearchField) []SearchIndexField {
	if len(input) == 0 { // Changed from input == nil || len(*input) == 0
		return []SearchIndexField{}
	}

	results := make([]SearchIndexField, 0, len(input))
	for _, v := range input {
		field := SearchIndexField{
			Name:        v.Name,
			Type:        string(v.Type),
			Key:         pointer.From(v.Key),
			Searchable:  pointer.From(v.Searchable),
			Filterable:  pointer.From(v.Filterable),
			Sortable:    pointer.From(v.Sortable),
			Facetable:   pointer.From(v.Facetable),
			Retrievable: pointer.From(v.Retrievable),
		}

		if v.Analyzer != nil {
			field.Analyzer = string(*v.Analyzer)
		}
		if v.SearchAnalyzer != nil {
			field.SearchAnalyzer = string(*v.SearchAnalyzer)
		}
		if v.IndexAnalyzer != nil {
			field.IndexAnalyzer = string(*v.IndexAnalyzer)
		}
		if v.SynonymMaps != nil {
			field.SynonymMaps = *v.SynonymMaps
		}

		results = append(results, field)
	}

	return results
}

func flattenCorsOptions(input *indexes.CorsOptions) []CorsOptions {
	if input == nil {
		return []CorsOptions{}
	}

	return []CorsOptions{
		{
			AllowedOrigins:  input.AllowedOrigins, // Changed from pointer.From(input.AllowedOrigins)
			MaxAgeInSeconds: pointer.From(input.MaxAgeInSeconds),
		},
	}
}
