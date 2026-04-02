// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package search_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-sdk/data-plane/search/2025-09-01/indexes"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/search/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

type SearchIndexResource struct{}

func TestAccSearchIndex_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_search_index", "test")
	r := SearchIndexResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("name").HasValue(fmt.Sprintf("acctestindex-%d", data.RandomInteger)),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSearchIndex_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_search_index", "test")
	r := SearchIndexResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func TestAccSearchIndex_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_search_index", "test")
	r := SearchIndexResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("fields.#").HasValue("4"),
				check.That(data.ResourceName).Key("cors_options.#").HasValue("1"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSearchIndex_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_search_index", "test")
	r := SearchIndexResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("fields.#").HasValue("4"),
			),
		},
		data.ImportStep(),
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (r SearchIndexResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.SearchIndexID(state.ID)
	if err != nil {
		return nil, err
	}

	// Get the endpoint
	domainSuffix, ok := clients.Account.Environment.Search.DomainSuffix()
	if !ok {
		return nil, errors.New("could not determine Search domain suffix for the current environment")
	}
	endpoint := fmt.Sprintf("https://%s.%s", id.SearchServiceId.SearchServiceName, *domainSuffix)

	// Use the pre-configured data plane client
	client := clients.Search.SearchDataPlaneClient.Indexes.Clone(endpoint)

	// Check if the index exists
	indexId := indexes.IndexId{
		IndexName: id.IndexName,
	}
	resp, err := client.Get(ctx, indexId, indexes.DefaultGetOperationOptions())
	if err != nil {
		if response.WasNotFound(resp.HttpResponse) {
			return pointer.To(false), nil
		}
		return nil, fmt.Errorf("retrieving %s: %+v", id, err)
	}

	return pointer.To(true), nil
}

func (r SearchIndexResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-search-%d"
  location = "%s"
}

resource "azurerm_search_service" "test" {
  name                = "acctestsearchsvc%s"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku                 = "standard"
}

resource "azurerm_search_index" "test" {
  name              = "acctestindex-%d"
  search_service_id = azurerm_search_service.test.id

  fields {
    name        = "id"
    type        = "Edm.String"
    key         = true
    retrievable = true
  }

  fields {
    name        = "title"
    type        = "Edm.String"
    searchable  = true
    filterable  = true
    retrievable = true
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.RandomInteger)
}

func (r SearchIndexResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_search_index" "import" {
  name              = azurerm_search_index.test.name
  search_service_id = azurerm_search_index.test.search_service_id

  fields {
    name        = "id"
    type        = "Edm.String"
    key         = true
    retrievable = true
  }

  fields {
    name        = "title"
    type        = "Edm.String"
    searchable  = true
    filterable  = true
    retrievable = true
  }
}
`, r.basic(data))
}

func (r SearchIndexResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-search-%d"
  location = "%s"
}

resource "azurerm_search_service" "test" {
  name                = "accttestsearchsvc%s"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku                 = "standard"
}

resource "azurerm_search_index" "test" {
  name              = "acctestindex-%d"
  search_service_id = azurerm_search_service.test.id

  fields {
    name        = "id"
    type        = "Edm.String"
    key         = true
    retrievable = true
    filterable  = true
    sortable    = true
  }

  fields {
    name        = "title"
    type        = "Edm.String"
    searchable  = true
    filterable  = true
    sortable    = true
    retrievable = true
    analyzer    = "standard.lucene"
  }

  fields {
    name        = "description"
    type        = "Edm.String"
    searchable  = true
    retrievable = true
    analyzer    = "en.microsoft"
  }

  fields {
    name        = "category"
    type        = "Edm.String"
    filterable  = true
    facetable   = true
    retrievable = true
  }

  cors_options {
    allowed_origins     = ["https://example.com"]
    max_age_in_seconds  = 300
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.RandomInteger)
}
