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

type AutomationCredentialResource struct{}

func TestAccAutomationCredential_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_automation_credential", "test")
	r := AutomationCredentialResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("username").HasValue("test_user"),
			),
		},
		data.ImportStep("password"),
	})
}

func TestAccAutomationCredential_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_automation_credential", "test")
	r := AutomationCredentialResource{}

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

func TestAccAutomationCredential_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_automation_credential", "test")
	r := AutomationCredentialResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("username").HasValue("test_user"),
				check.That(data.ResourceName).Key("description").HasValue("This is a test credential for terraform acceptance test"),
			),
		},
		data.ImportStep("password"),
	})
}

func (t AutomationCredentialResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.CredentialID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Automation.CredentialClient.Get(ctx, id.ResourceGroup, id.AutomationAccountName, id.Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving Automation Credential %q (resource group: %q): %+v", id.Name, id.ResourceGroup, err)
	}

	return utils.Bool(resp.CredentialProperties != nil), nil
}

func (AutomationCredentialResource) basic(data acceptance.TestData) string {
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

resource "azurerm_automation_credential" "test" {
  name                    = "acctest-%d"
  resource_group_name     = azurerm_resource_group.test.name
  automation_account_name = azurerm_automation_account.test.name
  username                = "test_user"
  password                = "test_pwd"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (AutomationCredentialResource) requiresImport(data acceptance.TestData) string {
	template := AutomationCredentialResource{}.basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_automation_credential" "import" {
  name                    = azurerm_automation_credential.test.name
  resource_group_name     = azurerm_automation_credential.test.resource_group_name
  automation_account_name = azurerm_automation_credential.test.automation_account_name
  username                = azurerm_automation_credential.test.username
  password                = azurerm_automation_credential.test.password
}
`, template)
}

func (AutomationCredentialResource) complete(data acceptance.TestData) string {
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

resource "azurerm_automation_credential" "test" {
  name                    = "acctest-%d"
  resource_group_name     = azurerm_resource_group.test.name
  automation_account_name = azurerm_automation_account.test.name
  username                = "test_user"
  password                = "test_pwd"
  description             = "This is a test credential for terraform acceptance test"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}
