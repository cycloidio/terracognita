package client

import (
	"github.com/Azure/azure-sdk-for-go/services/webpubsub/mgmt/2021-10-01/webpubsub"
	"github.com/hashicorp/terraform-provider-azurerm/common"
	"github.com/hashicorp/terraform-provider-azurerm/services/signalr/sdk/2020-05-01/signalr"
)

type Client struct {
	SignalRClient       *signalr.SignalRClient
	WebPubsubClient     *webpubsub.Client
	WebPubsubHubsClient *webpubsub.HubsClient
}

func NewClient(o *common.ClientOptions) *Client {
	signalRClient := signalr.NewSignalRClientWithBaseURI(o.ResourceManagerEndpoint)
	o.ConfigureClient(&signalRClient.Client, o.ResourceManagerAuthorizer)

	webpubsubClient := webpubsub.NewClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&webpubsubClient.Client, o.ResourceManagerAuthorizer)

	webpubsubHubsClient := webpubsub.NewHubsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&webpubsubHubsClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		SignalRClient:       &signalRClient,
		WebPubsubClient:     &webpubsubClient,
		WebPubsubHubsClient: &webpubsubHubsClient,
	}
}
