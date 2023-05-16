package client

import (
	"github.com/Azure/azure-sdk-for-go/services/timeseriesinsights/mgmt/2020-05-15/timeseriesinsights"
	"github.com/hashicorp/terraform-provider-azurerm/common"
)

type Client struct {
	AccessPoliciesClient    *timeseriesinsights.AccessPoliciesClient
	EnvironmentsClient      *timeseriesinsights.EnvironmentsClient
	EventSourcesClient      *timeseriesinsights.EventSourcesClient
	ReferenceDataSetsClient *timeseriesinsights.ReferenceDataSetsClient
}

func NewClient(o *common.ClientOptions) *Client {
	AccessPoliciesClient := timeseriesinsights.NewAccessPoliciesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&AccessPoliciesClient.Client, o.ResourceManagerAuthorizer)

	EnvironmentsClient := timeseriesinsights.NewEnvironmentsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&EnvironmentsClient.Client, o.ResourceManagerAuthorizer)

	EventSourcesClient := timeseriesinsights.NewEventSourcesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&EventSourcesClient.Client, o.ResourceManagerAuthorizer)

	ReferenceDataSetsClient := timeseriesinsights.NewReferenceDataSetsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&ReferenceDataSetsClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		AccessPoliciesClient:    &AccessPoliciesClient,
		EnvironmentsClient:      &EnvironmentsClient,
		EventSourcesClient:      &EventSourcesClient,
		ReferenceDataSetsClient: &ReferenceDataSetsClient,
	}
}
