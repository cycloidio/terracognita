package client

import (
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2020-05-01/managementgroups"
	"github.com/hashicorp/terraform-provider-azurerm/common"
)

type Client struct {
	GroupsClient       *managementgroups.Client
	SubscriptionClient *managementgroups.SubscriptionsClient
}

func NewClient(o *common.ClientOptions) *Client {
	GroupsClient := managementgroups.NewClientWithBaseURI(o.ResourceManagerEndpoint)
	o.ConfigureClient(&GroupsClient.Client, o.ResourceManagerAuthorizer)

	SubscriptionClient := managementgroups.NewSubscriptionsClientWithBaseURI(o.ResourceManagerEndpoint)
	o.ConfigureClient(&SubscriptionClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		GroupsClient:       &GroupsClient,
		SubscriptionClient: &SubscriptionClient,
	}
}
