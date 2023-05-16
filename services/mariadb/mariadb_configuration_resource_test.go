package mariadb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/mariadb/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type MariaDbConfigurationResource struct{}

func TestAccMariaDbConfiguration_characterSetServer(t *testing.T) {
	srv := acceptance.BuildTestData(t, "azurerm_mariadb_server", "test")
	data := acceptance.BuildTestData(t, "azurerm_mariadb_configuration", "test")
	r := MariaDbConfigurationResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.characterSetServer(data),
			Check: acceptance.ComposeTestCheckFunc(
				data.CheckWithClient(checkValueIs("hebrew")),
			),
		},
		data.ImportStep(),
		{
			Config: r.empty(data),
			Check: acceptance.ComposeTestCheckFunc(
				// "delete" resets back to the default value
				srv.CheckWithClient(checkValueIsReset("character_set_server")),
			),
		},
	})
}

func TestAccMariaDbConfiguration_interactiveTimeout(t *testing.T) {
	srv := acceptance.BuildTestData(t, "azurerm_mariadb_server", "test")
	data := acceptance.BuildTestData(t, "azurerm_mariadb_configuration", "test")
	r := MariaDbConfigurationResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.interactiveTimeout(data),
			Check: acceptance.ComposeTestCheckFunc(
				data.CheckWithClient(checkValueIs("30")),
			),
		},
		data.ImportStep(),
		{
			Config: r.empty(data),
			Check: acceptance.ComposeTestCheckFunc(
				// "delete" resets back to the default value
				srv.CheckWithClient(checkValueIsReset("interactive_timeout")),
			),
		},
	})
}

func TestAccMariaDbConfiguration_logSlowAdminStatements(t *testing.T) {
	srv := acceptance.BuildTestData(t, "azurerm_mariadb_server", "test")
	data := acceptance.BuildTestData(t, "azurerm_mariadb_configuration", "test")
	r := MariaDbConfigurationResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.logSlowAdminStatements(data),
			Check: acceptance.ComposeTestCheckFunc(
				data.CheckWithClient(checkValueIs("On")),
			),
		},
		data.ImportStep(),
		{
			Config: r.empty(data),
			Check: acceptance.ComposeTestCheckFunc(
				// "delete" resets back to the default value
				srv.CheckWithClient(checkValueIsReset("log_slow_admin_statements")),
			),
		},
	})
}

func (MariaDbConfigurationResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.MariaDBConfigurationID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.MariaDB.ConfigurationsClient.Get(ctx, id.ResourceGroup, id.ServerName, id.ConfigurationName)
	if err != nil {
		return nil, fmt.Errorf("retrieving %s: %v", *id, err)
	}

	return utils.Bool(resp.ConfigurationProperties != nil), nil
}

func checkValueIs(value string) acceptance.ClientCheckFunc {
	return func(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) error {
		id, err := parse.MariaDBConfigurationID(state.ID)
		if err != nil {
			return err
		}

		resp, err := clients.MariaDB.ConfigurationsClient.Get(ctx, id.ResourceGroup, id.ServerName, id.ConfigurationName)
		if err != nil {
			return fmt.Errorf("retrieving %s: %v", *id, err)
		}

		if resp.Value == nil {
			return fmt.Errorf("%s Value is nil", *id)
		}

		actualValue := *resp.Value

		if value != actualValue {
			return fmt.Errorf("%s Value (%s) != expected (%s)", *id, actualValue, value)
		}

		return nil
	}
}

func checkValueIsReset(configurationName string) acceptance.ClientCheckFunc {
	return func(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) error {
		id, err := parse.ServerID(state.ID)
		if err != nil {
			return err
		}

		resp, err := clients.MariaDB.ConfigurationsClient.Get(ctx, id.ResourceGroup, id.Name, configurationName)
		if err != nil {
			return fmt.Errorf("retrieving MariaDB Configuration %q (Server %q / Resource Group %q): %v", configurationName, id.Name, id.ResourceGroup, err)
		}

		if resp.Value == nil {
			return fmt.Errorf("MariaDB Configuration %q (Server %q / Resource Group %q) Value is nil", configurationName, id.Name, id.ResourceGroup)
		}

		if resp.DefaultValue == nil {
			return fmt.Errorf("MariaDB Configuration %q (Server %q / Resource Group %q) Default Value is nil", configurationName, id.Name, id.ResourceGroup)
		}
		actualValue := *resp.Value
		defaultValue := *resp.DefaultValue

		if defaultValue != actualValue {
			return fmt.Errorf("MariaDB Configuration %q (Server %q / Resource Group %q) Value (%s) != Default (%s)", configurationName, id.Name, id.ResourceGroup, actualValue, defaultValue)
		}

		return nil
	}
}

func (r MariaDbConfigurationResource) characterSetServer(data acceptance.TestData) string {
	return r.template(data, "character_set_server", "hebrew")
}

func (r MariaDbConfigurationResource) interactiveTimeout(data acceptance.TestData) string {
	return r.template(data, "interactive_timeout", "30")
}

func (r MariaDbConfigurationResource) logSlowAdminStatements(data acceptance.TestData) string {
	return r.template(data, "log_slow_admin_statements", "On")
}

func (r MariaDbConfigurationResource) template(data acceptance.TestData, name string, value string) string {
	server := r.empty(data)
	config := fmt.Sprintf(`
resource "azurerm_mariadb_configuration" "test" {
  name                = "%s"
  resource_group_name = "${azurerm_resource_group.test.name}"
  server_name         = "${azurerm_mariadb_server.test.name}"
  value               = "%s"
}
`, name, value)
	return server + config
}

func (MariaDbConfigurationResource) empty(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_mariadb_server" "test" {
  name                = "acctestmariadbsvr-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku_name            = "GP_Gen5_2"
  version             = "10.2"

  administrator_login          = "acctestun"
  administrator_login_password = "H@Sh1CoR3!"
  backup_retention_days        = 7
  geo_redundant_backup_enabled = false
  ssl_enforcement_enabled      = true
  storage_mb                   = 51200
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
