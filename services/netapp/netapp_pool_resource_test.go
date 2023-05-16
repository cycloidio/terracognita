package netapp_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/netapp/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type NetAppPoolResource struct{}

func TestAccNetAppPool_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_netapp_pool", "test")
	r := NetAppPoolResource{}

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

func TestAccNetAppPool_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_netapp_pool", "test")
	r := NetAppPoolResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_netapp_pool"),
		},
	})
}

func TestAccNetAppPool_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_netapp_pool", "test")
	r := NetAppPoolResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("service_level").HasValue("Standard"),
				check.That(data.ResourceName).Key("size_in_tb").HasValue("15"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.FoO").HasValue("BaR"),
				check.That(data.ResourceName).Key("qos_type").HasValue("Auto"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccNetAppPool_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_netapp_pool", "test")
	r := NetAppPoolResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("size_in_tb").HasValue("4"),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
				check.That(data.ResourceName).Key("qos_type").HasValue("Auto"),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("size_in_tb").HasValue("15"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.FoO").HasValue("BaR"),
				check.That(data.ResourceName).Key("qos_type").HasValue("Auto"),
			),
		},
		data.ImportStep(),
		{
			Config: r.completeQosChange(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("size_in_tb").HasValue("15"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.FoO").HasValue("BaR"),
				check.That(data.ResourceName).Key("qos_type").HasValue("Manual"),
			),
		},
	})
}

func (t NetAppPoolResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.CapacityPoolID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.NetApp.PoolClient.Get(ctx, id.ResourceGroup, id.NetAppAccountName, id.Name)
	if err != nil {
		return nil, fmt.Errorf("reading Netapp Pool (%s): %+v", id.String(), err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (NetAppPoolResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-netapp-%d"
  location = "%s"
}

resource "azurerm_netapp_account" "test" {
  name                = "acctest-NetAppAccount-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_netapp_pool" "test" {
  name                = "acctest-NetAppPool-%d"
  account_name        = azurerm_netapp_account.test.name
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  service_level       = "Standard"
  size_in_tb          = 4
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (r NetAppPoolResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s
resource "azurerm_netapp_pool" "import" {
  name                = azurerm_netapp_pool.test.name
  location            = azurerm_netapp_pool.test.location
  resource_group_name = azurerm_netapp_pool.test.resource_group_name
  account_name        = azurerm_netapp_pool.test.account_name
  service_level       = azurerm_netapp_pool.test.service_level
  size_in_tb          = azurerm_netapp_pool.test.size_in_tb
}
`, r.basic(data))
}

func (NetAppPoolResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-netapp-%d"
  location = "%s"
}

resource "azurerm_netapp_account" "test" {
  name                = "acctest-NetAppAccount-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_netapp_pool" "test" {
  name                = "acctest-NetAppPool-%d"
  account_name        = azurerm_netapp_account.test.name
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  service_level       = "Standard"
  size_in_tb          = 15
  qos_type            = "Auto"

  tags = {
    "FoO" = "BaR"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (NetAppPoolResource) completeQosChange(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-netapp-%d"
  location = "%s"
}

resource "azurerm_netapp_account" "test" {
  name                = "acctest-NetAppAccount-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_netapp_pool" "test" {
  name                = "acctest-NetAppPool-%d"
  account_name        = azurerm_netapp_account.test.name
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  service_level       = "Standard"
  size_in_tb          = 15
  qos_type            = "Manual"

  tags = {
    "FoO" = "BaR"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}
