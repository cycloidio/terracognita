package client

import (
	"github.com/Azure/azure-sdk-for-go/services/mariadb/mgmt/2018-06-01/mariadb"
	"github.com/hashicorp/terraform-provider-azurerm/common"
)

type Client struct {
	ConfigurationsClient      *mariadb.ConfigurationsClient
	DatabasesClient           *mariadb.DatabasesClient
	FirewallRulesClient       *mariadb.FirewallRulesClient
	ServersClient             *mariadb.ServersClient
	VirtualNetworkRulesClient *mariadb.VirtualNetworkRulesClient
}

func NewClient(o *common.ClientOptions) *Client {
	configurationsClient := mariadb.NewConfigurationsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&configurationsClient.Client, o.ResourceManagerAuthorizer)

	DatabasesClient := mariadb.NewDatabasesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&DatabasesClient.Client, o.ResourceManagerAuthorizer)

	FirewallRulesClient := mariadb.NewFirewallRulesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&FirewallRulesClient.Client, o.ResourceManagerAuthorizer)

	ServersClient := mariadb.NewServersClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&ServersClient.Client, o.ResourceManagerAuthorizer)

	VirtualNetworkRulesClient := mariadb.NewVirtualNetworkRulesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&VirtualNetworkRulesClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		ConfigurationsClient:      &configurationsClient,
		DatabasesClient:           &DatabasesClient,
		FirewallRulesClient:       &FirewallRulesClient,
		ServersClient:             &ServersClient,
		VirtualNetworkRulesClient: &VirtualNetworkRulesClient,
	}
}
