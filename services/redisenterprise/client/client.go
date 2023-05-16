package client

import (
	"github.com/hashicorp/terraform-provider-azurerm/common"
	"github.com/hashicorp/terraform-provider-azurerm/services/redisenterprise/sdk/2022-01-01/databases"
	"github.com/hashicorp/terraform-provider-azurerm/services/redisenterprise/sdk/2022-01-01/redisenterprise"
)

type Client struct {
	Client         *redisenterprise.RedisEnterpriseClient
	DatabaseClient *databases.DatabasesClient
}

func NewClient(o *common.ClientOptions) *Client {
	client := redisenterprise.NewRedisEnterpriseClientWithBaseURI(o.ResourceManagerEndpoint)
	o.ConfigureClient(&client.Client, o.ResourceManagerAuthorizer)

	databaseClient := databases.NewDatabasesClientWithBaseURI(o.ResourceManagerEndpoint)
	o.ConfigureClient(&databaseClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		Client:         &client,
		DatabaseClient: &databaseClient,
	}
}
