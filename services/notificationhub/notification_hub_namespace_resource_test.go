package notificationhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/notificationhub/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type NotificationHubNamespaceResource struct{}

func TestAccNotificationHubNamespace_free(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_notification_hub_namespace", "test")
	r := NotificationHubNamespaceResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.free(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("namespace_type"),
	})
}

func TestAccNotificationHubNamespace_updateTag(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_notification_hub_namespace", "test")
	r := NotificationHubNamespaceResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.free(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
			),
		},
		data.ImportStep("namespace_type"),
		{
			Config: r.withoutTag(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
		data.ImportStep("namespace_type"),
		{
			Config: r.free(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
			),
		},
		data.ImportStep("namespace_type"),
	})
}

func TestAccNotificationHubNamespace_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_notification_hub_namespace", "test")
	r := NotificationHubNamespaceResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.free(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func (NotificationHubNamespaceResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.NamespaceID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.NotificationHubs.NamespacesClient.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving %s: %v", id.String(), err)
	}

	return utils.Bool(resp.NamespaceProperties != nil), nil
}

func (NotificationHubNamespaceResource) free(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_notification_hub_namespace" "test" {
  name                = "acctestnhn-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  namespace_type      = "NotificationHub"

  sku_name = "Free"

  tags = {
    env = "Test"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (NotificationHubNamespaceResource) withoutTag(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_notification_hub_namespace" "test" {
  name                = "acctestnhn-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  namespace_type      = "NotificationHub"

  sku_name = "Free"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (NotificationHubNamespaceResource) requiresImport(data acceptance.TestData) string {
	template := NotificationHubNamespaceResource{}.free(data)
	return fmt.Sprintf(`
%s

resource "azurerm_notification_hub_namespace" "import" {
  name                = azurerm_notification_hub_namespace.test.name
  resource_group_name = azurerm_notification_hub_namespace.test.resource_group_name
  location            = azurerm_notification_hub_namespace.test.location
  namespace_type      = azurerm_notification_hub_namespace.test.namespace_type

  sku_name = "Free"
}
`, template)
}
