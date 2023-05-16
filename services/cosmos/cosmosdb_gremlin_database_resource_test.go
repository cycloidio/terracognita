package cosmos_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2021-10-15/documentdb"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/cosmos/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type CosmosGremlinDatabaseResource struct{}

func TestAccCosmosGremlinDatabase_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_cosmosdb_gremlin_database", "test")
	r := CosmosGremlinDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeAggregateTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccCosmosGremlinDatabase_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_cosmosdb_gremlin_database", "test")
	r := CosmosGremlinDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_cosmosdb_gremlin_database"),
		},
	})
}

func TestAccCosmosGremlinDatabase_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_cosmosdb_gremlin_database", "test")
	r := CosmosGremlinDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data, 700),
			Check: acceptance.ComposeAggregateTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("throughput").HasValue("700"),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data, 1700),
			Check: acceptance.ComposeAggregateTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("throughput").HasValue("1700"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccCosmosGremlinDatabase_autoscale(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_cosmosdb_gremlin_database", "test")
	r := CosmosGremlinDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.autoscale(data, 4000),
			Check: acceptance.ComposeAggregateTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("autoscale_settings.0.max_throughput").HasValue("4000"),
			),
		},
		data.ImportStep(),
		{
			Config: r.autoscale(data, 5000),
			Check: acceptance.ComposeAggregateTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("autoscale_settings.0.max_throughput").HasValue("5000"),
			),
		},
		data.ImportStep(),
		{
			Config: r.autoscale(data, 4000),
			Check: acceptance.ComposeAggregateTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("autoscale_settings.0.max_throughput").HasValue("4000"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccCosmosGremlinDatabase_serverless(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_cosmosdb_gremlin_database", "test")
	r := CosmosGremlinDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.serverless(data),
			Check: acceptance.ComposeAggregateTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (t CosmosGremlinDatabaseResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.GremlinDatabaseID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Cosmos.GremlinClient.GetGremlinDatabase(ctx, id.ResourceGroup, id.DatabaseAccountName, id.Name)
	if err != nil {
		return nil, fmt.Errorf("reading Cosmos Gremlin Database (%s): %+v", id.String(), err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (CosmosGremlinDatabaseResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_cosmosdb_gremlin_database" "test" {
  name                = "acctest-%[2]d"
  resource_group_name = azurerm_cosmosdb_account.test.resource_group_name
  account_name        = azurerm_cosmosdb_account.test.name
}
`, CosmosDBAccountResource{}.capabilities(data, documentdb.DatabaseAccountKindGlobalDocumentDB, []string{"EnableGremlin"}), data.RandomInteger)
}

func (r CosmosGremlinDatabaseResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_cosmosdb_gremlin_database" "import" {
  name                = azurerm_cosmosdb_gremlin_database.test.name
  resource_group_name = azurerm_cosmosdb_gremlin_database.test.resource_group_name
  account_name        = azurerm_cosmosdb_gremlin_database.test.account_name
}
`, r.basic(data))
}

func (CosmosGremlinDatabaseResource) complete(data acceptance.TestData, throughput int) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_cosmosdb_gremlin_database" "test" {
  name                = "acctest-%[2]d"
  resource_group_name = azurerm_cosmosdb_account.test.resource_group_name
  account_name        = azurerm_cosmosdb_account.test.name
  throughput          = %[3]d
}
`, CosmosDBAccountResource{}.capabilities(data, documentdb.DatabaseAccountKindGlobalDocumentDB, []string{"EnableGremlin"}), data.RandomInteger, throughput)
}

func (CosmosGremlinDatabaseResource) autoscale(data acceptance.TestData, maxThroughput int) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_cosmosdb_gremlin_database" "test" {
  name                = "acctest-%[2]d"
  resource_group_name = azurerm_cosmosdb_account.test.resource_group_name
  account_name        = azurerm_cosmosdb_account.test.name
  autoscale_settings {
    max_throughput = %[3]d
  }
}
`, CosmosDBAccountResource{}.capabilities(data, documentdb.DatabaseAccountKindGlobalDocumentDB, []string{"EnableGremlin"}), data.RandomInteger, maxThroughput)
}

func (CosmosGremlinDatabaseResource) serverless(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_cosmosdb_gremlin_database" "test" {
  name                = "acctest-%[2]d"
  resource_group_name = azurerm_cosmosdb_account.test.resource_group_name
  account_name        = azurerm_cosmosdb_account.test.name
}
`, CosmosDBAccountResource{}.capabilities(data, documentdb.DatabaseAccountKindGlobalDocumentDB, []string{"EnableGremlin", "EnableServerless"}), data.RandomInteger)
}
