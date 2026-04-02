package parse

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-sdk/resource-manager/search/2025-05-01/services"
)

type SearchIndexId struct {
	SearchServiceId services.SearchServiceId
	IndexName       string
}

func NewSearchIndexID(searchServiceId services.SearchServiceId, indexName string) SearchIndexId {
	return SearchIndexId{
		SearchServiceId: searchServiceId,
		IndexName:       indexName,
	}
}

func (id SearchIndexId) ID() string {
	return fmt.Sprintf("%s/indexes/%s", id.SearchServiceId.ID(), id.IndexName)
}

func (id SearchIndexId) String() string {
	return fmt.Sprintf("Search Index %q (Search Service %q / Resource Group %q)", id.IndexName, id.SearchServiceId.SearchServiceName, id.SearchServiceId.ResourceGroupName)
}

func SearchIndexID(input string) (*SearchIndexId, error) {
	parts := strings.Split(input, "/indexes/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid search index ID format: %s", input)
	}

	searchServiceId, err := services.ParseSearchServiceID(parts[0])
	if err != nil {
		return nil, fmt.Errorf("parsing search service ID: %+v", err)
	}

	return &SearchIndexId{
		SearchServiceId: *searchServiceId,
		IndexName:       parts[1],
	}, nil
}

func ValidateSearchIndexID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if _, err := SearchIndexID(v); err != nil {
		errors = append(errors, err)
	}

	return
}
