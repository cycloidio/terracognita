package sql_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/sql/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type SqlAdministratorResource struct{}

func TestAccSqlAdministrator_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_sql_active_directory_administrator", "test")
	r := SqlAdministratorResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("login").HasValue("sqladmin"),
			),
		},
		data.ImportStep(),
		{
			Config: r.withUpdates(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("login").HasValue("sqladmin2"),
			),
		},
		data.ImportStep(),
		{
			Config: r.aadOnlyAuthEnabled(data, true),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("azuread_authentication_only").HasValue("true"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSqlAdministrator_aadOnlyAuth(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_sql_active_directory_administrator", "test")
	r := SqlAdministratorResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.aadOnlyAuthEnabled(data, true),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("azuread_authentication_only").HasValue("true"),
			),
		},
		data.ImportStep(),
		{
			Config: r.aadOnlyAuthEnabled(data, false),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("azuread_authentication_only").HasValue("false"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSqlAdministrator_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_sql_active_directory_administrator", "test")
	r := SqlAdministratorResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("login").HasValue("sqladmin"),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func TestAccSqlAdministrator_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_sql_active_directory_administrator", "test")
	r := SqlAdministratorResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       r.basic,
			TestResource: r,
		}),
	})
}

func (r SqlAdministratorResource) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.AzureActiveDirectoryAdministratorID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := client.Sql.ServerAzureADAdministratorsClient.Get(ctx, id.ResourceGroup, id.ServerName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return utils.Bool(false), nil
		}
		return nil, fmt.Errorf("retrieving AAD Administrator %q (Server %q / Resource Group %q): %+v", id.AdministratorName, id.ServerName, id.ResourceGroup, err)
	}
	return utils.Bool(true), nil
}

func (r SqlAdministratorResource) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.AzureActiveDirectoryAdministratorID(state.ID)
	if err != nil {
		return nil, err
	}
	if _, err := client.Sql.ServerAzureADAdministratorsClient.Delete(ctx, id.ResourceGroup, id.ServerName); err != nil {
		return nil, err
	}
	return utils.Bool(true), nil
}

func (r SqlAdministratorResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

data "azurerm_client_config" "current" {
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_sql_server" "test" {
  name                         = "acctestsqlserver%d"
  resource_group_name          = azurerm_resource_group.test.name
  location                     = azurerm_resource_group.test.location
  version                      = "12.0"
  administrator_login          = "mradministrator"
  administrator_login_password = "thisIsDog11"
}

resource "azurerm_sql_active_directory_administrator" "test" {
  server_name         = azurerm_sql_server.test.name
  resource_group_name = azurerm_resource_group.test.name
  login               = "sqladmin"
  tenant_id           = data.azurerm_client_config.current.tenant_id
  object_id           = data.azurerm_client_config.current.client_id
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r SqlAdministratorResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_sql_active_directory_administrator" "import" {
  server_name         = azurerm_sql_active_directory_administrator.test.server_name
  resource_group_name = azurerm_sql_active_directory_administrator.test.resource_group_name
  login               = azurerm_sql_active_directory_administrator.test.login
  tenant_id           = azurerm_sql_active_directory_administrator.test.tenant_id
  object_id           = azurerm_sql_active_directory_administrator.test.object_id
}
`, r.basic(data))
}

func (r SqlAdministratorResource) withUpdates(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

data "azurerm_client_config" "current" {
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_sql_server" "test" {
  name                         = "acctestsqlserver%d"
  resource_group_name          = azurerm_resource_group.test.name
  location                     = azurerm_resource_group.test.location
  version                      = "12.0"
  administrator_login          = "mradministrator"
  administrator_login_password = "thisIsDog11"
}

resource "azurerm_sql_active_directory_administrator" "test" {
  server_name         = azurerm_sql_server.test.name
  resource_group_name = azurerm_resource_group.test.name
  login               = "sqladmin2"
  tenant_id           = data.azurerm_client_config.current.tenant_id
  object_id           = data.azurerm_client_config.current.client_id
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r SqlAdministratorResource) aadOnlyAuthEnabled(data acceptance.TestData, enabled bool) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

data "azurerm_client_config" "current" {
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_sql_server" "test" {
  name                         = "acctestsqlserver%d"
  resource_group_name          = azurerm_resource_group.test.name
  location                     = azurerm_resource_group.test.location
  version                      = "12.0"
  administrator_login          = "mradministrator"
  administrator_login_password = "thisIsDog11"
}

resource "azurerm_sql_active_directory_administrator" "test" {
  server_name                 = azurerm_sql_server.test.name
  resource_group_name         = azurerm_resource_group.test.name
  login                       = "sqladmin"
  tenant_id                   = data.azurerm_client_config.current.tenant_id
  object_id                   = data.azurerm_client_config.current.client_id
  azuread_authentication_only = %t
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, enabled)
}
