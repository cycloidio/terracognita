package desktopvirtualization_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/desktopvirtualization/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type VirtualDesktopHostPoolResource struct{}

func TestAccVirtualDesktopHostPool_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_virtual_desktop_host_pool", "test")
	r := VirtualDesktopHostPoolResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
	})
}

func TestAccVirtualDesktopHostPool_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_virtual_desktop_host_pool", "test")
	r := VirtualDesktopHostPoolResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
			),
		},
	})
}

func TestAccVirtualDesktopHostPool_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_virtual_desktop_host_pool", "test")
	r := VirtualDesktopHostPoolResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
			),
		},
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
	})
}

func TestAccVirtualDesktopHostPool_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_virtual_desktop_host_pool", "test")
	r := VirtualDesktopHostPoolResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func (VirtualDesktopHostPoolResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.HostPoolID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.DesktopVirtualization.HostPoolsClient.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving Virtual Desktop Host Pool %q (Resource Group: %q) does not exist", id.Name, id.ResourceGroup)
	}

	return utils.Bool(resp.HostPoolProperties != nil), nil
}

func (VirtualDesktopHostPoolResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-vdesktophp-%d"
  location = "%s"
}

resource "azurerm_virtual_desktop_host_pool" "test" {
  name                 = "acctestHP%s"
  location             = azurerm_resource_group.test.location
  resource_group_name  = azurerm_resource_group.test.name
  type                 = "Pooled"
  validate_environment = true
  load_balancer_type   = "BreadthFirst"
}
`, data.RandomInteger, data.Locations.Secondary, data.RandomString)
}

func (VirtualDesktopHostPoolResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-vdesktophp-%d"
  location = "%s"
}

resource "azurerm_virtual_desktop_host_pool" "test" {
  name                     = "acctestHP%s"
  location                 = azurerm_resource_group.test.location
  resource_group_name      = azurerm_resource_group.test.name
  type                     = "Pooled"
  friendly_name            = "A Friendly Name!"
  description              = "A Description!"
  validate_environment     = true
  start_vm_on_connect      = true
  load_balancer_type       = "BreadthFirst"
  maximum_sessions_allowed = 100
  preferred_app_group_type = "Desktop"
  custom_rdp_properties    = "audiocapturemode:i:1;audiomode:i:0;"

  tags = {
    Purpose = "Acceptance-Testing"
  }
}
`, data.RandomInteger, data.Locations.Secondary, data.RandomString)
}

func (r VirtualDesktopHostPoolResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_virtual_desktop_host_pool" "import" {
  name                 = azurerm_virtual_desktop_host_pool.test.name
  location             = azurerm_virtual_desktop_host_pool.test.location
  resource_group_name  = azurerm_virtual_desktop_host_pool.test.resource_group_name
  validate_environment = azurerm_virtual_desktop_host_pool.test.validate_environment
  type                 = azurerm_virtual_desktop_host_pool.test.type
  load_balancer_type   = azurerm_virtual_desktop_host_pool.test.load_balancer_type
}
`, r.basic(data))
}
