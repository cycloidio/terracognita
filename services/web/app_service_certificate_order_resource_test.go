package web_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/web/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type AppServiceCertificateOrderResource struct{}

func TestAccAppServiceCertificateOrder_basic(t *testing.T) {
	if os.Getenv("ARM_RUN_TEST_APP_SERVICE_CERTIFICATE") == "" {
		t.Skip("Skipping as ARM_RUN_TEST_APP_SERVICE_CERTIFICATE is not specified")
		return
	}
	data := acceptance.BuildTestData(t, "azurerm_app_service_certificate_order", "test")
	r := AppServiceCertificateOrderResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("csr").Exists(),
				check.That(data.ResourceName).Key("domain_verification_token").Exists(),
				check.That(data.ResourceName).Key("distinguished_name").HasValue("CN=example.com"),
				check.That(data.ResourceName).Key("product_type").HasValue("Standard"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAppServiceCertificateOrder_wildcard(t *testing.T) {
	if os.Getenv("ARM_RUN_TEST_APP_SERVICE_CERTIFICATE") == "" {
		t.Skip("Skipping as ARM_RUN_TEST_APP_SERVICE_CERTIFICATE is not specified")
		return
	}

	data := acceptance.BuildTestData(t, "azurerm_app_service_certificate_order", "test")
	r := AppServiceCertificateOrderResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.wildcard(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("csr").Exists(),
				check.That(data.ResourceName).Key("domain_verification_token").Exists(),
				check.That(data.ResourceName).Key("distinguished_name").HasValue("CN=*.example.com"),
				check.That(data.ResourceName).Key("product_type").HasValue("WildCard"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAppServiceCertificateOrder_requiresImport(t *testing.T) {
	if os.Getenv("ARM_RUN_TEST_APP_SERVICE_CERTIFICATE") == "" {
		t.Skip("Skipping as ARM_RUN_TEST_APP_SERVICE_CERTIFICATE is not specified")
		return
	}

	data := acceptance.BuildTestData(t, "azurerm_app_service_certificate_order", "test")
	r := AppServiceCertificateOrderResource{}

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

func TestAccAppServiceCertificateOrder_complete(t *testing.T) {
	if os.Getenv("ARM_RUN_TEST_APP_SERVICE_CERTIFICATE") == "" {
		t.Skip("Skipping as ARM_RUN_TEST_APP_SERVICE_CERTIFICATE is not specified")
		return
	}
	data := acceptance.BuildTestData(t, "azurerm_app_service_certificate_order", "test")
	r := AppServiceCertificateOrderResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data, 4096),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("csr").Exists(),
				check.That(data.ResourceName).Key("domain_verification_token").Exists(),
				check.That(data.ResourceName).Key("distinguished_name").HasValue("CN=example.com"),
				check.That(data.ResourceName).Key("product_type").HasValue("Standard"),
				check.That(data.ResourceName).Key("validity_in_years").HasValue("1"),
				check.That(data.ResourceName).Key("auto_renew").HasValue("false"),
				check.That(data.ResourceName).Key("key_size").HasValue("4096"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAppServiceCertificateOrder_update(t *testing.T) {
	if os.Getenv("ARM_RUN_TEST_APP_SERVICE_CERTIFICATE") == "" {
		t.Skip("Skipping as ARM_RUN_TEST_APP_SERVICE_CERTIFICATE is not specified")
		return
	}

	data := acceptance.BuildTestData(t, "azurerm_app_service_certificate_order", "test")
	r := AppServiceCertificateOrderResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("csr").Exists(),
				check.That(data.ResourceName).Key("domain_verification_token").Exists(),
				check.That(data.ResourceName).Key("distinguished_name").HasValue("CN=example.com"),
				check.That(data.ResourceName).Key("product_type").HasValue("Standard"),
			),
		},
		{
			Config: r.complete(data, 2048), // keySize cannot be updated
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("domain_verification_token").Exists(),
				check.That(data.ResourceName).Key("distinguished_name").HasValue("CN=example.com"),
				check.That(data.ResourceName).Key("auto_renew").HasValue("false"),
			),
		},
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("domain_verification_token").Exists(),
				check.That(data.ResourceName).Key("distinguished_name").HasValue("CN=example.com"),
				check.That(data.ResourceName).Key("auto_renew").HasValue("true"),
				check.That(data.ResourceName).Key("key_size").HasValue("2048"),
			),
		},
		data.ImportStep(),
	})
}

func (r AppServiceCertificateOrderResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.CertificateOrderID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Web.CertificatesOrderClient.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return utils.Bool(false), nil
		}
		return nil, fmt.Errorf("retrieving App Service Certificate Order %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	return utils.Bool(true), nil
}

func (r AppServiceCertificateOrderResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_app_service_certificate_order" "test" {
  name                = "acctestASCO-%d"
  location            = "global"
  resource_group_name = azurerm_resource_group.test.name
  distinguished_name  = "CN=example.com"
  product_type        = "Standard"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r AppServiceCertificateOrderResource) wildcard(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_app_service_certificate_order" "test" {
  name                = "acctestASCO-%d"
  location            = "global"
  resource_group_name = azurerm_resource_group.test.name
  distinguished_name  = "CN=*.example.com"
  product_type        = "WildCard"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r AppServiceCertificateOrderResource) requiresImport(data acceptance.TestData) string {
	template := r.basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_app_service_certificate_order" "import" {
  name                = azurerm_app_service_certificate_order.test.name
  location            = azurerm_app_service_certificate_order.test.location
  resource_group_name = azurerm_app_service_certificate_order.test.resource_group_name
  distinguished_name  = azurerm_app_service_certificate_order.test.distinguished_name
  product_type        = azurerm_app_service_certificate_order.test.product_type
}
`, template)
}

func (r AppServiceCertificateOrderResource) complete(data acceptance.TestData, keySize int) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_app_service_certificate_order" "test" {
  name                = "acctestASCO-%d"
  location            = "global"
  resource_group_name = azurerm_resource_group.test.name
  distinguished_name  = "CN=example.com"
  product_type        = "Standard"
  auto_renew          = false
  validity_in_years   = 1
  key_size            = %d
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, keySize)
}
