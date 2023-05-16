package client

import (
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/terraform-provider-azurerm/common"
)

type Client struct {
	LoadBalancersClient                   *network.LoadBalancersClient
	LoadBalancerBackendAddressPoolsClient *network.LoadBalancerBackendAddressPoolsClient
	LoadBalancingRulesClient              *network.LoadBalancerLoadBalancingRulesClient
}

func NewClient(o *common.ClientOptions) *Client {
	loadBalancersClient := network.NewLoadBalancersClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&loadBalancersClient.Client, o.ResourceManagerAuthorizer)

	loadBalancerBackendAddressPoolsClient := network.NewLoadBalancerBackendAddressPoolsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&loadBalancerBackendAddressPoolsClient.Client, o.ResourceManagerAuthorizer)

	loadBalancingRulesClient := network.NewLoadBalancerLoadBalancingRulesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&loadBalancingRulesClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		LoadBalancersClient:                   &loadBalancersClient,
		LoadBalancerBackendAddressPoolsClient: &loadBalancerBackendAddressPoolsClient,
		LoadBalancingRulesClient:              &loadBalancingRulesClient,
	}
}
