package apimanagement_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/apimanagement/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type ApiManagementAuthorizationServerResource struct{}

func TestAccApiManagementAuthorizationServer_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_api_management_authorization_server", "test")
	r := ApiManagementAuthorizationServerResource{}

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

func TestAccApiManagementAuthorizationServer_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_api_management_authorization_server", "test")
	r := ApiManagementAuthorizationServerResource{}

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

func TestAccApiManagementAuthorizationServer_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_api_management_authorization_server", "test")
	r := ApiManagementAuthorizationServerResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("client_secret", "resource_owner_password", "resource_owner_username"),
	})
}

func (ApiManagementAuthorizationServerResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.AuthorizationServerID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.ApiManagement.AuthorizationServersClient.Get(ctx, id.ResourceGroup, id.ServiceName, id.Name)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %+v", *id, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (r ApiManagementAuthorizationServerResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_api_management_authorization_server" "test" {
  name                         = "acctestauthsrv-%d"
  resource_group_name          = azurerm_resource_group.test.name
  api_management_name          = azurerm_api_management.test.name
  display_name                 = "Test Group"
  authorization_endpoint       = "https://azacceptance.hashicorptest.com/client/authorize"
  client_id                    = "42424242-4242-4242-4242-424242424242"
  client_registration_endpoint = "https://azacceptance.hashicorptest.com/client/register"

  grant_types = [
    "implicit",
  ]

  authorization_methods = [
    "GET",
  ]
}
`, r.template(data), data.RandomInteger)
}

func (r ApiManagementAuthorizationServerResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_api_management_authorization_server" "import" {
  name                         = azurerm_api_management_authorization_server.test.name
  resource_group_name          = azurerm_api_management_authorization_server.test.resource_group_name
  api_management_name          = azurerm_api_management_authorization_server.test.api_management_name
  display_name                 = azurerm_api_management_authorization_server.test.display_name
  authorization_endpoint       = azurerm_api_management_authorization_server.test.authorization_endpoint
  client_id                    = azurerm_api_management_authorization_server.test.client_id
  client_registration_endpoint = azurerm_api_management_authorization_server.test.client_registration_endpoint
  grant_types                  = azurerm_api_management_authorization_server.test.grant_types

  authorization_methods = [
    "GET",
  ]
}
`, r.basic(data))
}

func (r ApiManagementAuthorizationServerResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_api_management_authorization_server" "test" {
  name                         = "acctestauthsrv-%d"
  resource_group_name          = azurerm_resource_group.test.name
  api_management_name          = azurerm_api_management.test.name
  display_name                 = "Test Group"
  authorization_endpoint       = "https://azacceptance.hashicorptest.com/client/authorize"
  client_id                    = "42424242-4242-4242-4242-424242424242"
  client_registration_endpoint = "https://azacceptance.hashicorptest.com/client/register"
  description                  = "This is a test description"

  token_body_parameter {
    name  = "test"
    value = "token-body-parameter"
  }

  client_authentication_method = [
    "Basic",
  ]

  grant_types = [
    "authorizationCode",
  ]

  authorization_methods = [
    "GET",
    "POST",
  ]

  bearer_token_sending_methods = [
    "authorizationHeader",
  ]

  client_secret           = "n1n3-m0re-s3a5on5-m0r1y"
  default_scope           = "read write"
  token_endpoint          = "https://azacceptance.hashicorptest.com/client/token"
  resource_owner_username = "rick"
  resource_owner_password = "C-193P"
  support_state           = true
}
`, r.template(data), data.RandomInteger)
}

func (ApiManagementAuthorizationServerResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_api_management" "test" {
  name                = "acctestAM-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  publisher_name      = "pub1"
  publisher_email     = "pub1@email.com"
  sku_name            = "Consumption_0"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
