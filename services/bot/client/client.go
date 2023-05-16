package client

import (
	"github.com/Azure/azure-sdk-for-go/services/healthbot/mgmt/2020-12-08/healthbot"
	"github.com/Azure/azure-sdk-for-go/services/preview/botservice/mgmt/2021-05-01-preview/botservice"
	"github.com/hashicorp/terraform-provider-azurerm/common"
)

type Client struct {
	BotClient        *botservice.BotsClient
	ConnectionClient *botservice.BotConnectionClient
	ChannelClient    *botservice.ChannelsClient
	HealthbotClient  *healthbot.BotsClient
}

func NewClient(o *common.ClientOptions) *Client {
	botClient := botservice.NewBotsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&botClient.Client, o.ResourceManagerAuthorizer)

	connectionClient := botservice.NewBotConnectionClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&connectionClient.Client, o.ResourceManagerAuthorizer)

	channelClient := botservice.NewChannelsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&channelClient.Client, o.ResourceManagerAuthorizer)

	healthBotClient := healthbot.NewBotsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&healthBotClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		BotClient:        &botClient,
		ChannelClient:    &channelClient,
		ConnectionClient: &connectionClient,
		HealthbotClient:  &healthBotClient,
	}
}
