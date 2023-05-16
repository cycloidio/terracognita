package client

import (
	"github.com/hashicorp/terraform-provider-azurerm/common"
	"github.com/hashicorp/terraform-provider-azurerm/services/frontdoor/sdk/2020-04-01/webapplicationfirewallpolicies"
	"github.com/hashicorp/terraform-provider-azurerm/services/frontdoor/sdk/2020-05-01/frontdoors"
)

// NOTE: @tombuildsstuff: we cannot upgrade the "old" FrontDoor resources past 2020-11-01
// however the "new" FrontDoor resources require a newer API version (or specifying `sku`
// for the older API version). Since the older resources will be deprecated when the new
// ones are available, we should leave the older resources on the older API version.

type Client struct {
	FrontDoorsClient       *frontdoors.FrontDoorsClient
	FrontDoorsPolicyClient *webapplicationfirewallpolicies.WebApplicationFirewallPoliciesClient
}

func NewClient(o *common.ClientOptions) *Client {
	frontDoorsClient := frontdoors.NewFrontDoorsClientWithBaseURI(o.ResourceManagerEndpoint)
	o.ConfigureClient(&frontDoorsClient.Client, o.ResourceManagerAuthorizer)

	frontDoorsPolicyClient := webapplicationfirewallpolicies.NewWebApplicationFirewallPoliciesClientWithBaseURI(o.ResourceManagerEndpoint)
	o.ConfigureClient(&frontDoorsPolicyClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		FrontDoorsClient:       &frontDoorsClient,
		FrontDoorsPolicyClient: &frontDoorsPolicyClient,
	}
}
