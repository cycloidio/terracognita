package client

import (
	"github.com/Azure/azure-sdk-for-go/services/preview/portal/mgmt/2019-01-01-preview/portal"
	"github.com/hashicorp/terraform-provider-azurerm/common"
)

type Client struct {
	DashboardsClient           *portal.DashboardsClient
	TenantConfigurationsClient *portal.TenantConfigurationsClient
}

func NewClient(o *common.ClientOptions) *Client {
	dashboardsClient := portal.NewDashboardsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&dashboardsClient.Client, o.ResourceManagerAuthorizer)

	tenantConfigurationsClient := portal.NewTenantConfigurationsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&tenantConfigurationsClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		DashboardsClient:           &dashboardsClient,
		TenantConfigurationsClient: &tenantConfigurationsClient,
	}
}
