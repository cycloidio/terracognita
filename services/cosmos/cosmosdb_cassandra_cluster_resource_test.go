package cosmos_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/cosmos/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type CassandraClusterResource struct{}

func testAccCassandraCluster_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_cosmosdb_cassandra_cluster", "test")
	r := CassandraClusterResource{}

	data.ResourceSequentialTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeAggregateTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("default_admin_password"),
	})
}

func testAccCassandraCluster_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_cosmosdb_cassandra_cluster", "test")
	r := CassandraClusterResource{}

	data.ResourceSequentialTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("default_admin_password"),
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_cosmosdb_cassandra_cluster"),
		},
	})
}

func testAccCassandraCluster_tags(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_cosmosdb_cassandra_cluster", "test")
	r := CassandraClusterResource{}

	data.ResourceSequentialTest(t, r, []acceptance.TestStep{
		{
			Config: r.tags(data, "Test"),
			Check: acceptance.ComposeAggregateTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("default_admin_password"),
		{
			Config: r.tags(data, "Test2"),
			Check: acceptance.ComposeAggregateTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("default_admin_password"),
	})
}

func (t CassandraClusterResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.CassandraClusterID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Cosmos.CassandraClustersClient.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %+v", id, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (r CassandraClusterResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_cosmosdb_cassandra_cluster" "test" {
  name                           = "acctca-mi-cluster-%d"
  resource_group_name            = azurerm_resource_group.test.name
  location                       = azurerm_resource_group.test.location
  delegated_management_subnet_id = azurerm_subnet.test.id
  default_admin_password         = "Password1234"

  depends_on = [azurerm_role_assignment.test]
}
`, r.template(data), data.RandomInteger)
}

func (r CassandraClusterResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_cosmosdb_cassandra_cluster" "import" {
  name                           = azurerm_cosmosdb_cassandra_cluster.test.name
  resource_group_name            = azurerm_cosmosdb_cassandra_cluster.test.resource_group_name
  location                       = azurerm_cosmosdb_cassandra_cluster.test.location
  delegated_management_subnet_id = azurerm_subnet.test.id
  default_admin_password         = "Password1234"
}
`, r.basic(data))
}

func (r CassandraClusterResource) tags(data acceptance.TestData, tag string) string {
	return fmt.Sprintf(`
%s

resource "azurerm_cosmosdb_cassandra_cluster" "test" {
  name                           = "acctca-mi-cluster-%d"
  resource_group_name            = azurerm_resource_group.test.name
  location                       = azurerm_resource_group.test.location
  delegated_management_subnet_id = azurerm_subnet.test.id
  default_admin_password         = "Password1234"

  tags = {
    Env = "%s"
  }

  depends_on = [azurerm_role_assignment.test]
}
`, r.template(data), data.RandomInteger, tag)
}

func (r CassandraClusterResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctest-ca-%d"
  location = "%s"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctvn-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  address_space       = ["10.0.0.0/16"]
}

resource "azurerm_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.1.0/24"]
}

data "azuread_service_principal" "test" {
  display_name = "Azure Cosmos DB"
}

resource "azurerm_role_assignment" "test" {
  scope                = azurerm_virtual_network.test.id
  role_definition_name = "Network Contributor"
  principal_id         = data.azuread_service_principal.test.object_id
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}
