package containers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/containers/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type ContainerRegistryAgentPoolResource struct {
}

func TestAccContainerRegistryAgentPool_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_container_registry_agent_pool", "test")
	r := ContainerRegistryAgentPoolResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccContainerRegistryAgentPool_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_container_registry_agent_pool", "test")
	r := ContainerRegistryAgentPoolResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_container_registry_agent_pool"),
		},
	})
}

func TestAccContainerRegistryAgentPool_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_container_registry_agent_pool", "test")
	r := ContainerRegistryAgentPoolResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (t ContainerRegistryAgentPoolResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.ContainerRegistryAgentPoolID(state.ID)
	if err != nil {
		return nil, err
	}
	resourceGroup := id.ResourceGroup
	name := id.AgentPoolName
	registryName := id.RegistryName

	resp, err := clients.Containers.ContainerRegistryAgentPoolsClient.Get(ctx, resourceGroup, registryName, name)
	if err != nil {
		return nil, fmt.Errorf("reading Container Registry Agent Pool (%s): %+v", id, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (ContainerRegistryAgentPoolResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-acragent_pool-%d"
  location = "%s"
}

resource "azurerm_container_registry" "test" {
  name                = "testacccr%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku                 = "Premium"
}

resource "azurerm_container_registry_agent_pool" "test" {
  name                    = "ap%d"
  resource_group_name     = azurerm_resource_group.test.name
  location                = azurerm_resource_group.test.location
  container_registry_name = azurerm_container_registry.test.name
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomIntOfLength(15))
}

func (ContainerRegistryAgentPoolResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-acragent_pool-%d"
  location = "%s"
}

resource "azurerm_container_registry" "test" {
  name                = "testacccr%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku                 = "Premium"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctestvirtnet%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_container_registry_agent_pool" "test" {
  name                      = "ap%d"
  resource_group_name       = azurerm_resource_group.test.name
  location                  = azurerm_resource_group.test.location
  container_registry_name   = azurerm_container_registry.test.name
  instance_count            = 2
  tier                      = "S2"
  virtual_network_subnet_id = azurerm_subnet.test.id
}

`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomIntOfLength(15))
}

func (r ContainerRegistryAgentPoolResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_container_registry_agent_pool" "import" {
  name                    = azurerm_container_registry_agent_pool.test.name
  resource_group_name     = azurerm_container_registry_agent_pool.test.resource_group_name
  location                = azurerm_container_registry_agent_pool.test.location
  container_registry_name = azurerm_container_registry_agent_pool.test.container_registry_name
}

`, r.basic(data))
}
