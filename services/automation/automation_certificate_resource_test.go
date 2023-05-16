package automation_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/automation/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type AutomationCertificateResource struct{}

var (
	testCertThumbprintRaw, _ = os.ReadFile(filepath.Join("testdata", "automation_certificate_test.thumb"))
	testCertRaw, _           = os.ReadFile(filepath.Join("testdata", "automation_certificate_test.pfx"))
)

var testCertBase64 = base64.StdEncoding.EncodeToString(testCertRaw)

func TestAccAutomationCertificate_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_automation_certificate", "test")
	r := AutomationCertificateResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("base64"),
	})
}

func TestAccAutomationCertificate_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_automation_certificate", "test")
	r := AutomationCertificateResource{}

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

func TestAccAutomationCertificate_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_automation_certificate", "test")
	r := AutomationCertificateResource{}
	testCertThumbprint := strings.TrimSpace(string(testCertThumbprintRaw))

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("base64").HasValue(testCertBase64),
				check.That(data.ResourceName).Key("thumbprint").HasValue(testCertThumbprint),
			),
		},
		data.ImportStep("base64"),
	})
}

func TestAccAutomationCertificate_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_automation_certificate", "test")
	r := AutomationCertificateResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("description").HasValue(""),
			),
		},
		data.ImportStep("base64"),
		{
			Config: r.update(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("description").HasValue("This is a test certificate for terraform acceptance test"),
			),
		},
		data.ImportStep("base64"),
	})
}

func (t AutomationCertificateResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.CertificateID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Automation.CertificateClient.Get(ctx, id.ResourceGroup, id.AutomationAccountName, id.Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving Automation Certificate %q (resource group: %q): %+v", id.Name, id.ResourceGroup, err)
	}

	return utils.Bool(resp.CertificateProperties != nil), nil
}

func (AutomationCertificateResource) basic(data acceptance.TestData) string {
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

resource "azurerm_automation_certificate" "test" {
  name                    = "acctest-%d"
  resource_group_name     = azurerm_resource_group.test.name
  automation_account_name = azurerm_automation_account.test.name
  base64                  = "%s"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, testCertBase64)
}

func (AutomationCertificateResource) requiresImport(data acceptance.TestData) string {
	template := AutomationCertificateResource{}.basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_automation_certificate" "import" {
  name                    = azurerm_automation_certificate.test.name
  resource_group_name     = azurerm_automation_certificate.test.resource_group_name
  automation_account_name = azurerm_automation_certificate.test.automation_account_name
  base64                  = azurerm_automation_certificate.test.base64
}
`, template)
}

func (AutomationCertificateResource) complete(data acceptance.TestData) string {
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

resource "azurerm_automation_certificate" "test" {
  name                    = "acctest-%d"
  resource_group_name     = azurerm_resource_group.test.name
  automation_account_name = azurerm_automation_account.test.name
  base64                  = "%s"
  description             = "This is a test certificate for terraform acceptance test"
  exportable              = true
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, testCertBase64)
}

func (AutomationCertificateResource) update(data acceptance.TestData) string {
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

resource "azurerm_automation_certificate" "test" {
  name                    = "acctest-%d"
  resource_group_name     = azurerm_resource_group.test.name
  automation_account_name = azurerm_automation_account.test.name
  base64                  = "%s"
  description             = "This is a test certificate for terraform acceptance test"
  exportable              = false
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, testCertBase64)
}
