package network_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type VirtualHubIPResource struct{}

func TestAccVirtualHubIP_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_virtual_hub_ip", "test")
	r := VirtualHubIPResource{}

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

func TestAccVirtualHubIP_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_virtual_hub_ip", "test")
	r := VirtualHubIPResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_virtual_hub_ip"),
		},
	})
}

func TestAccVirtualHubIP_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_virtual_hub_ip", "test")
	r := VirtualHubIPResource{}

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

func TestAccVirtualHubIP_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_virtual_hub_ip", "test")
	r := VirtualHubIPResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (t VirtualHubIPResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.VirtualHubIpConfigurationID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Network.VirtualHubIPClient.Get(ctx, id.ResourceGroup, id.VirtualHubName, id.IpConfigurationName)
	if err != nil {
		return nil, fmt.Errorf("reading Virtual Hub IP (%s): %+v", id, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (r VirtualHubIPResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_virtual_hub_ip" "test" {
  name                 = "acctest-vhubipconfig-%d"
  virtual_hub_id       = azurerm_virtual_hub.test.id
  public_ip_address_id = azurerm_public_ip.test.id
  subnet_id            = azurerm_subnet.test.id
}
`, r.template(data), data.RandomInteger)
}

func (r VirtualHubIPResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_virtual_hub_ip" "import" {
  name                 = azurerm_virtual_hub_ip.test.name
  virtual_hub_id       = azurerm_virtual_hub_ip.test.virtual_hub_id
  public_ip_address_id = azurerm_virtual_hub_ip.test.public_ip_address_id
  subnet_id            = azurerm_virtual_hub_ip.test.subnet_id
}
`, r.basic(data))
}

func (r VirtualHubIPResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_virtual_hub_ip" "test" {
  name                         = "acctest-vhubipconfig-%d"
  virtual_hub_id               = azurerm_virtual_hub.test.id
  private_ip_address           = "10.5.1.18"
  private_ip_allocation_method = "Static"
  public_ip_address_id         = azurerm_public_ip.test.id
  subnet_id                    = azurerm_subnet.test.id
}
`, r.template(data), data.RandomInteger)
}

func (VirtualHubIPResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-vhub-%d"
  location = "%s"
}

resource "azurerm_virtual_hub" "test" {
  name                = "acctest-vhub-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku                 = "Standard"
}

resource "azurerm_public_ip" "test" {
  name                = "acctest-pip-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  allocation_method   = "Static"
  sku                 = "Standard"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctest-vnet-%d"
  address_space       = ["10.5.0.0/16"]
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_subnet" "test" {
  name                 = "RouteServerSubnet"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.5.1.0/24"]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}
