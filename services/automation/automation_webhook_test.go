package automation_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/automation/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type AutomationWebhookResource struct{}

func TestAccAutomationWebhook_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_automation_webhook", "test")
	r := AutomationWebhookResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.SimpleWebhook(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("name").Exists(),
				check.That(data.ResourceName).Key("resource_group_name").Exists(),
				check.That(data.ResourceName).Key("automation_account_name").Exists(),
				check.That(data.ResourceName).Key("expiry_time").Exists(),
				check.That(data.ResourceName).Key("enabled").HasValue("true"),
				check.That(data.ResourceName).Key("runbook_name").HasValue("Get-AzureVMTutorial"),
				check.That(data.ResourceName).Key("parameters").DoesNotExist(),
				check.That(data.ResourceName).Key("uri").Exists(),
			),
		},
		data.ImportStep("uri"),
	})
}

func TestAccAutomationWebhook_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_automation_webhook", "test")
	r := AutomationWebhookResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.SimpleWebhook(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_automation_webhook"),
		},
	})
}

func TestAccAutomationWebhook_WithParameters(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_automation_webhook", "test")
	r := AutomationWebhookResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.WebhookWithParameters(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("parameters.input").HasValue("parameter"),
				check.That(data.ResourceName).Key("uri").Exists(),
			),
		},
		data.ImportStep("uri"),
	})
}

func TestAccAutomationWebhook_ChangeUri(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_automation_webhook", "test")
	r := AutomationWebhookResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.WebhookURIChange(data, "https://12345678-9012-3456-7890-123456789012.webhook.we.azure-automation.net/webhooks?token=abcdefghijklmnoprstuwxyz1234567890abcdefghijklm"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("uri").HasValue("https://12345678-9012-3456-7890-123456789012.webhook.we.azure-automation.net/webhooks?token=abcdefghijklmnoprstuwxyz1234567890abcdefghijklm"),
			),
		},
		data.ImportStep("uri"),
		{
			Config: r.WebhookURIChange(data, "https://12345678-9012-3456-7890-123456789012.webhook.we.azure-automation.net/webhooks?token=abcdefghijklmnoprstuwxyz1234567890abcdefg313377"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("uri").HasValue("https://12345678-9012-3456-7890-123456789012.webhook.we.azure-automation.net/webhooks?token=abcdefghijklmnoprstuwxyz1234567890abcdefg313377"),
			),
		},
	})
}

func TestAccAutomationWebhook_WithWorkerGroup(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_automation_webhook", "test")
	r := AutomationWebhookResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config:      r.WebhookOnWorkerGroup(data),
			ExpectError: regexp.MustCompile("The Hybrid Runbook Worker Group given in RunOn parameter does not exist"),
		},
	})
}

func (t AutomationWebhookResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.WebhookID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Automation.WebhookClient.Get(ctx, id.ResourceGroup, id.AutomationAccountName, id.Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving Automation Webhook '%s' (resource group: '%s') does not exist", id.Name, id.ResourceGroup)
	}

	return utils.Bool(resp.WebhookProperties != nil), nil
}

func (AutomationWebhookResource) ParentResources(data acceptance.TestData) string {
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

resource "azurerm_automation_runbook" "test" {
  name                    = "Get-AzureVMTutorial"
  location                = azurerm_resource_group.test.location
  resource_group_name     = azurerm_resource_group.test.name
  automation_account_name = azurerm_automation_account.test.name

  log_verbose  = "true"
  log_progress = "true"
  description  = "This is a test runbook for terraform acceptance test"
  runbook_type = "PowerShell"

  content = <<CONTENT
# Some test content
# for Terraform acceptance test
CONTENT
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (AutomationWebhookResource) SimpleWebhook(data acceptance.TestData) string {
	template := AutomationWebhookResource{}.ParentResources(data)
	return fmt.Sprintf(`
%s

resource "azurerm_automation_webhook" "test" {
  name                    = "TestRunbook_webhook"
  resource_group_name     = azurerm_resource_group.test.name
  automation_account_name = azurerm_automation_account.test.name
  expiry_time             = "%s"
  enabled                 = true
  runbook_name            = azurerm_automation_runbook.test.name
}
`, template, time.Now().UTC().Add(time.Hour).Format(time.RFC3339))
}

func (AutomationWebhookResource) requiresImport(data acceptance.TestData) string {
	template := AutomationWebhookResource{}.SimpleWebhook(data)
	return fmt.Sprintf(`
%s

resource "azurerm_automation_webhook" "import" {
  name                    = azurerm_automation_webhook.test.name
  resource_group_name     = azurerm_automation_webhook.test.resource_group_name
  automation_account_name = azurerm_automation_webhook.test.automation_account_name
  expiry_time             = azurerm_automation_webhook.test.expiry_time
  enabled                 = azurerm_automation_webhook.test.enabled
  runbook_name            = azurerm_automation_webhook.test.runbook_name
}
`, template)
}

func (AutomationWebhookResource) WebhookWithParameters(data acceptance.TestData) string {
	template := AutomationWebhookResource{}.ParentResources(data)
	return fmt.Sprintf(`
%s

resource "azurerm_automation_webhook" "test" {
  name                    = "TestRunbook_webhook"
  resource_group_name     = azurerm_resource_group.test.name
  automation_account_name = azurerm_automation_account.test.name
  expiry_time             = "%s"
  enabled                 = true
  runbook_name            = azurerm_automation_runbook.test.name
  parameters = {
    input = "parameter"
  }
}
`, template, time.Now().UTC().Add(time.Hour).Format(time.RFC3339))
}

// requires creation of worker group
func (AutomationWebhookResource) WebhookOnWorkerGroup(data acceptance.TestData) string {
	template := AutomationWebhookResource{}.ParentResources(data)
	return fmt.Sprintf(`
%s

resource "azurerm_automation_webhook" "test" {
  name                    = "TestRunbook_webhook"
  resource_group_name     = azurerm_resource_group.test.name
  automation_account_name = azurerm_automation_account.test.name
  expiry_time             = timeadd(timestamp(), "10h")
  enabled                 = true
  runbook_name            = azurerm_automation_runbook.test.name
  run_on_worker_group     = "workergroup"
}
`, template)
}

func (AutomationWebhookResource) WebhookURIChange(data acceptance.TestData, uri string) string {
	template := AutomationWebhookResource{}.ParentResources(data)
	return fmt.Sprintf(`
%s

resource "azurerm_automation_webhook" "test" {
  name                    = "TestRunbook_webhook"
  resource_group_name     = azurerm_resource_group.test.name
  automation_account_name = azurerm_automation_account.test.name
  expiry_time             = "%s"
  enabled                 = true
  runbook_name            = azurerm_automation_runbook.test.name
  uri                     = "%s"
}
`, template, time.Now().UTC().Add(time.Hour).Format(time.RFC3339), uri)
}
