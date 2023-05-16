package client

import (
	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2021-10-15/documentdb"
	"github.com/hashicorp/terraform-provider-azurerm/common"
)

type Client struct {
	CassandraClient                  *documentdb.CassandraResourcesClient
	CassandraClustersClient          *documentdb.CassandraClustersClient
	CassandraDatacentersClient       *documentdb.CassandraDataCentersClient
	DatabaseClient                   *documentdb.DatabaseAccountsClient
	GremlinClient                    *documentdb.GremlinResourcesClient
	MongoDbClient                    *documentdb.MongoDBResourcesClient
	NotebookWorkspaceClient          *documentdb.NotebookWorkspacesClient
	RestorableDatabaseAccountsClient *documentdb.RestorableDatabaseAccountsClient
	SqlClient                        *documentdb.SQLResourcesClient
	SqlResourceClient                *documentdb.SQLResourcesClient
	TableClient                      *documentdb.TableResourcesClient
}

func NewClient(o *common.ClientOptions) *Client {
	cassandraClient := documentdb.NewCassandraResourcesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&cassandraClient.Client, o.ResourceManagerAuthorizer)

	cassandraClustersClient := documentdb.NewCassandraClustersClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&cassandraClustersClient.Client, o.ResourceManagerAuthorizer)

	cassandraDatacentersClient := documentdb.NewCassandraDataCentersClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&cassandraDatacentersClient.Client, o.ResourceManagerAuthorizer)

	databaseClient := documentdb.NewDatabaseAccountsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&databaseClient.Client, o.ResourceManagerAuthorizer)

	gremlinClient := documentdb.NewGremlinResourcesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&gremlinClient.Client, o.ResourceManagerAuthorizer)

	mongoDbClient := documentdb.NewMongoDBResourcesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&mongoDbClient.Client, o.ResourceManagerAuthorizer)

	notebookWorkspaceClient := documentdb.NewNotebookWorkspacesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&notebookWorkspaceClient.Client, o.ResourceManagerAuthorizer)

	restorableDatabaseAccountsClient := documentdb.NewRestorableDatabaseAccountsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&restorableDatabaseAccountsClient.Client, o.ResourceManagerAuthorizer)

	sqlClient := documentdb.NewSQLResourcesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&sqlClient.Client, o.ResourceManagerAuthorizer)

	sqlResourceClient := documentdb.NewSQLResourcesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&sqlResourceClient.Client, o.ResourceManagerAuthorizer)

	tableClient := documentdb.NewTableResourcesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&tableClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		CassandraClient:                  &cassandraClient,
		CassandraClustersClient:          &cassandraClustersClient,
		CassandraDatacentersClient:       &cassandraDatacentersClient,
		DatabaseClient:                   &databaseClient,
		GremlinClient:                    &gremlinClient,
		MongoDbClient:                    &mongoDbClient,
		NotebookWorkspaceClient:          &notebookWorkspaceClient,
		RestorableDatabaseAccountsClient: &restorableDatabaseAccountsClient,
		SqlClient:                        &sqlClient,
		SqlResourceClient:                &sqlResourceClient,
		TableClient:                      &tableClient,
	}
}
