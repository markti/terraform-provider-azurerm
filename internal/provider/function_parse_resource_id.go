// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
		Summary:     "Parses an Azure Resource Identifier (ID) into its constituent parts",
		Description: "Given an Azure Resource Identifier (ID) string, will an object containing the parts of Azure Resource ID. ",

		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "resource_id",
				Description: "Azure Resource Identifier (ID) to parse",
			},
		},
		Return: function.DynamicReturn{},
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
	result := map[string]types.String{
		"SubscriptionId":       types.StringNull(),
		"ResourceGroupName":    types.StringNull(),
		"ResourceName":         types.StringNull(),
		"ResourceProviderName": types.StringNull(),
	}

	// Parse the resource ID
	parts := strings.Split(resourceId, "/")

	var subResources []map[string]types.String

	for i := 0; i < len(parts); i++ {
		switch parts[i] {
		case "subscriptions":
			if i+1 < len(parts) {
				result["SubscriptionId"] = types.StringValue(parts[i+1])
			}
		case "resourceGroups":
			if i+1 < len(parts) {
				result["ResourceGroupName"] = types.StringValue(parts[i+1])
			}
		case "providers":
			if i+1 < len(parts) {
				result["ResourceProviderName"] = types.StringValue(parts[i+1])
			}
		default:
			if i > 0 && parts[i-1] == "providers" {
				if i+1 < len(parts) {
					result["ResourceName"] = types.StringValue(parts[i+1])
				}
			} else if i > 0 && (parts[i-1] != "subscriptions" && parts[i-1] != "resourceGroups" && parts[i-1] != "providers") {
				// Capture sub-resource type and name
				if i+1 < len(parts) {
					subResources = append(subResources, map[string]types.String{"Type": types.StringValue(parts[i]), "Name": types.StringValue(parts[i+1])})
				}
			}
		}
	}

	// Convert sub-resources to types.ListType
	//subResourceList := make([]types.Object, len(subResources))
	//for i, subResource := range subResources {
	//subResourceList[i] = types.ObjectValue(map[string]types.Type{"Type": types.StringType, "Name": types.StringType}, subResource)
	//}

	resp.Result.Set(ctx, result)

	return
}
