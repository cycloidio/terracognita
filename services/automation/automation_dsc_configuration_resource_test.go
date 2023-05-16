package automation_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/automation/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type AutomationDscConfigurationResource struct{}

func TestAccAutomationDscConfiguration_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_automation_dsc_configuration", "test")
	r := AutomationDscConfigurationResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("location").Exists(),
				check.That(data.ResourceName).Key("description").HasValue("test"),
				check.That(data.ResourceName).Key("log_verbose").Exists(),
				check.That(data.ResourceName).Key("state").Exists(),
				check.That(data.ResourceName).Key("content_embedded").HasValue("configuration acctest {}"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.ENV").HasValue("prod"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAutomationDscConfiguration_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_automation_dsc_configuration", "test")
	r := AutomationDscConfigurationResource{}

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

func TestAccAutomationDscConfiguration_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_automation_dsc_configuration", "test")
	r := AutomationDscConfigurationResource{}
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

func (t AutomationDscConfigurationResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.ConfigurationID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Automation.DscConfigurationClient.Get(ctx, id.ResourceGroup, id.AutomationAccountName, id.Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving Automation Dsc Configuration %q (resource group: %q): %+v", id.Name, id.ResourceGroup, err)
	}

	return utils.Bool(resp.DscConfigurationProperties != nil), nil
}

func (AutomationDscConfigurationResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-auto-%d"
  location = "%s"
}

resource "azurerm_automation_account" "test" {
  name                = "acctest-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku_name            = "Basic"
}

resource "azurerm_automation_dsc_configuration" "test" {
  name                    = "acctest"
  resource_group_name     = azurerm_resource_group.test.name
  automation_account_name = azurerm_automation_account.test.name
  location                = azurerm_resource_group.test.location
  content_embedded        = "configuration acctest {}"
  description             = "test"

  tags = {
    ENV = "prod"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (AutomationDscConfigurationResource) requiresImport(data acceptance.TestData) string {
	template := AutomationDscConfigurationResource{}.basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_automation_dsc_configuration" "import" {
  name                    = azurerm_automation_dsc_configuration.test.name
  resource_group_name     = azurerm_automation_dsc_configuration.test.resource_group_name
  automation_account_name = azurerm_automation_dsc_configuration.test.automation_account_name
  location                = azurerm_automation_dsc_configuration.test.location
  content_embedded        = azurerm_automation_dsc_configuration.test.content_embedded
  description             = azurerm_automation_dsc_configuration.test.description
}
`, template)
}

func (AutomationDscConfigurationResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-auto-%d"
  location = "%s"
}

resource "azurerm_automation_account" "test" {
  name                = "acctest-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku_name            = "Basic"
}

resource "azurerm_automation_dsc_configuration" "test" {
  name                    = "acctest"
  resource_group_name     = azurerm_resource_group.test.name
  automation_account_name = azurerm_automation_account.test.name
  location                = azurerm_resource_group.test.location
  content_embedded        = "configuration acctest {}"
  description             = "test"
  log_verbose             = "true"
  tags = {
    ENV = "prod"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
