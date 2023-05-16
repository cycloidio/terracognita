package client

import (
	"github.com/Azure/azure-sdk-for-go/services/digitaltwins/mgmt/2020-12-01/digitaltwins"
	"github.com/hashicorp/terraform-provider-azurerm/common"
)

type Client struct {
	EndpointClient *digitaltwins.EndpointClient
	InstanceClient *digitaltwins.Client
}

func NewClient(o *common.ClientOptions) *Client {
	endpointClient := digitaltwins.NewEndpointClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&endpointClient.Client, o.ResourceManagerAuthorizer)

	InstanceClient := digitaltwins.NewClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&InstanceClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		EndpointClient: &endpointClient,
		InstanceClient: &InstanceClient,
	}
}
