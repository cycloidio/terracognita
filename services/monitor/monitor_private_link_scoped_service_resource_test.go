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

type MonitorPrivateLinkScopedServiceResource struct{}

func TestAccMonitorPrivateLinkScopedService_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_private_link_scoped_service", "test")
	r := MonitorPrivateLinkScopedServiceResource{}

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

func TestAccMonitorPrivateLinkScopedService_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_private_link_scoped_service", "test")
	r := MonitorPrivateLinkScopedServiceResource{}

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

func (r MonitorPrivateLinkScopedServiceResource) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.PrivateLinkScopedServiceID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := client.Monitor.PrivateLinkScopedResourcesClient.Get(ctx, id.ResourceGroup, id.PrivateLinkScopeName, id.ScopedResourceName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return utils.Bool(false), nil
		}
		return nil, fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	return utils.Bool(resp.ScopedResourceProperties != nil), nil
}

func (r MonitorPrivateLinkScopedServiceResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-plss-%d"
  location = "%s"
}

resource "azurerm_monitor_private_link_scope" "test" {
  name                = "acctest-pls-%d"
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_application_insights" "test" {
  name                = "acctest-appinsights-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  application_type    = "web"
}

resource "azurerm_monitor_private_link_scoped_service" "test" {
  name                = "acctest-plss-%d"
  resource_group_name = azurerm_resource_group.test.name
  scope_name          = azurerm_monitor_private_link_scope.test.name
  linked_resource_id  = azurerm_application_insights.test.id
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (r MonitorPrivateLinkScopedServiceResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_private_link_scoped_service" "import" {
  name                = azurerm_monitor_private_link_scoped_service.test.name
  resource_group_name = azurerm_monitor_private_link_scoped_service.test.resource_group_name
  scope_name          = azurerm_monitor_private_link_scoped_service.test.scope_name
  linked_resource_id  = azurerm_monitor_private_link_scoped_service.test.linked_resource_id
}
`, r.basic(data))
}
