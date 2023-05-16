package apimanagement_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/apimanagement/mgmt/2021-08-01/apimanagement"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/apimanagement/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type ApiManagementIdentityProviderTwitterResource struct{}

func TestAccApiManagementIdentityProviderTwitter_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_api_management_identity_provider_twitter", "test")
	r := ApiManagementIdentityProviderTwitterResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("api_secret_key"),
	})
}

func TestAccApiManagementIdentityProviderTwitter_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_api_management_identity_provider_twitter", "test")
	r := ApiManagementIdentityProviderTwitterResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("api_key").HasValue("00000000000000000000000000000000"),
			),
		},
		data.ImportStep("api_secret_key"),
		{
			Config: r.update(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("api_key").HasValue("11111111111111111111111111111111"),
			),
		},
		data.ImportStep("api_secret_key"),
	})
}

func TestAccApiManagementIdentityProviderTwitter_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_api_management_identity_provider_twitter", "test")
	r := ApiManagementIdentityProviderTwitterResource{}

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

func (ApiManagementIdentityProviderTwitterResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.IdentityProviderID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.ApiManagement.IdentityProviderClient.Get(ctx, id.ResourceGroup, id.ServiceName, apimanagement.IdentityProviderType(id.Name))
	if err != nil {
		return nil, fmt.Errorf("reading ApiManagement Identity Provider Twitter (%s): %+v", id, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (ApiManagementIdentityProviderTwitterResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-api-%d"
  location = "%s"
}

resource "azurerm_api_management" "test" {
  name                = "acctestAM-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  publisher_name      = "pub1"
  publisher_email     = "pub1@email.com"
  sku_name            = "Developer_1"
}

resource "azurerm_api_management_identity_provider_twitter" "test" {
  resource_group_name = azurerm_resource_group.test.name
  api_management_name = azurerm_api_management.test.name
  api_key             = "00000000000000000000000000000000"
  api_secret_key      = "00000000000000000000000000000000"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (ApiManagementIdentityProviderTwitterResource) update(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-api-%d"
  location = "%s"
}

resource "azurerm_api_management" "test" {
  name                = "acctestAM-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  publisher_name      = "pub1"
  publisher_email     = "pub1@email.com"
  sku_name            = "Developer_1"
}

resource "azurerm_api_management_identity_provider_twitter" "test" {
  resource_group_name = azurerm_resource_group.test.name
  api_management_name = azurerm_api_management.test.name
  api_key             = "11111111111111111111111111111111"
  api_secret_key      = "11111111111111111111111111111111"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r ApiManagementIdentityProviderTwitterResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_api_management_identity_provider_twitter" "import" {
  resource_group_name = azurerm_api_management_identity_provider_twitter.test.resource_group_name
  api_management_name = azurerm_api_management_identity_provider_twitter.test.api_management_name
  api_key             = azurerm_api_management_identity_provider_twitter.test.api_key
  api_secret_key      = azurerm_api_management_identity_provider_twitter.test.api_secret_key
}
`, r.basic(data))
}
