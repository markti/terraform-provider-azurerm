// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
)

var _ function.Function = &ParseResourceIdentifierFunction{}

type ParseResourceIdentifierFunction struct{}

func NewParseResourceIdentifierFunction() function.Function {
	return &ParseResourceIdentifierFunction{}
}

func (f *ParseResourceIdentifierFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "parse_resource_id"
}

func (f *ParseResourceIdentifierFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "parse_resource_id Function",
		Description: "Given an Azure Resource Identifier (ID) string, will an object containing the parts of Azure Resource ID. ",

		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "resource_id",
				Description: "Azure Resource Identifier (ID) to parse",
			},
		},
		Return: function.ObjectReturn{
			AttributeTypes: *azure.ResourceID,
		},
	}
}

// Subscription /subscriptions/{subscription-id}
// Resource Group /subscriptions/{subscription-id}/resourceGroups/{resource-group-name}
// Virtual Machine /subscriptions/{subscription-id}/resourceGroups/{resource-group-name}/providers/Microsoft.Compute/virtualMachines/{vm-name}
// Virtual Network /subscriptions/{subscription-id}/resourceGroups/{resource-group-name}/providers/Microsoft.Network/virtualNetworks/{vnet-name}
// Multiple Sub-resources /subscriptions/{subscription-id}/resourceGroups/{resource-group-name}/providers/Microsoft.Network/virtualNetworks/{vnet-name}/subnets/{subnet-name}/networkInterfaces/{nic-name}
func (f *ParseResourceIdentifierFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var resourceId string

	resp.Error = req.Arguments.Get(ctx, &resourceId)
	if resp.Error != nil {
		return
	}

	// Initialize the response object
	result, err := azure.ParseAzureResourceID(resourceId)
	if err != nil {
		resp.Error = err
		return
	}

	// Convert sub-resources to types.ListType
	//subResourceList := make([]types.Object, len(subResources))
	//for i, subResource := range subResources {
	//subResourceList[i] = types.ObjectValue(map[string]types.Type{"Type": types.StringType, "Name": types.StringType}, subResource)
	//}

	resp.Result.Set(ctx, result)

	return
}
