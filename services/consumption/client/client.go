package client

import (
	"github.com/Azure/azure-sdk-for-go/services/consumption/mgmt/2019-10-01/consumption"
	"github.com/hashicorp/terraform-provider-azurerm/common"
)

type Client struct {
	BudgetsClient *consumption.BudgetsClient
}

func NewClient(o *common.ClientOptions) *Client {
	budgetsClient := consumption.NewBudgetsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&budgetsClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		BudgetsClient: &budgetsClient,
	}
}
