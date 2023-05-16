package monitor_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/monitor/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type MonitorPrivateLinkScopeResource struct{}

func TestAccMonitorPrivateLinkScope_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_private_link_scope", "test")
	r := MonitorPrivateLinkScopeResource{}

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

func TestAccMonitorPrivateLinkScope_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_private_link_scope", "test")
	r := MonitorPrivateLinkScopeResource{}

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

func TestAccMonitorPrivateLinkScope_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_private_link_scope", "test")
	r := MonitorPrivateLinkScopeResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data, "Test"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMonitorPrivateLinkScope_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_private_link_scope", "test")
	r := MonitorPrivateLinkScopeResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data, "Test1"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data, "Test2"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (r MonitorPrivateLinkScopeResource) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.PrivateLinkScopeID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := client.Monitor.PrivateLinkScopesClient.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return utils.Bool(false), nil
		}
		return nil, fmt.Errorf("retrieving %q %+v", id, err)
	}

	return utils.Bool(resp.AzureMonitorPrivateLinkScopeProperties != nil), nil
}

func (r MonitorPrivateLinkScopeResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-pls-%d"
  location = "%s"
}
`, data.RandomInteger, data.Locations.Primary)
}

func (r MonitorPrivateLinkScopeResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_private_link_scope" "test" {
  name                = "acctest-ampls-%d"
  resource_group_name = azurerm_resource_group.test.name
}
`, r.template(data), data.RandomInteger)
}

func (r MonitorPrivateLinkScopeResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_private_link_scope" "import" {
  name                = azurerm_monitor_private_link_scope.test.name
  resource_group_name = azurerm_monitor_private_link_scope.test.resource_group_name
}
`, r.basic(data))
}

func (r MonitorPrivateLinkScopeResource) complete(data acceptance.TestData, tag string) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_private_link_scope" "test" {
  name                = "acctest-AMPLS-%d"
  resource_group_name = azurerm_resource_group.test.name

  tags = {
    ENV = "%s"
  }
}
`, r.template(data), data.RandomInteger, tag)
}
