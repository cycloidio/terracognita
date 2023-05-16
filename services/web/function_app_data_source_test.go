package web_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
)

type FunctionAppDataSource struct{}

func TestAccFunctionAppDataSource_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_function_app", "test")

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: FunctionAppDataSource{}.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("outbound_ip_addresses").Exists(),
				check.That(data.ResourceName).Key("possible_outbound_ip_addresses").Exists(),
				check.That(data.ResourceName).Key("custom_domain_verification_id").Exists(),
			),
		},
	})
}

func TestAccFunctionAppDataSource_appSettings(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_function_app", "test")

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: FunctionAppDataSource{}.appSettings(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("app_settings.hello").HasValue("world"),
			),
		},
	})
}

func TestAccFunctionAppDataSource_connectionStrings(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_function_app", "test")

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: FunctionAppDataSource{}.connectionStrings(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("connection_string.0.name").HasValue("Example"),
				check.That(data.ResourceName).Key("connection_string.0.value").HasValue("some-postgresql-connection-string"),
				check.That(data.ResourceName).Key("connection_string.0.type").HasValue("PostgreSQL"),
			),
		},
	})
}

func TestAccFunctionAppDataSource_withSourceControl(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_function_app", "test")

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: FunctionAppDataSource{}.withSourceControl(data, "main"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("source_control.0.branch").HasValue("main"),
			),
		},
	})
}

func TestAccFunctionAppDataSource_siteConfig(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_function_app", "test")

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: FunctionAppDataSource{}.withSiteConfig(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("site_config.0.ip_restriction.0.ip_address").HasValue("10.10.10.10/32"),
				check.That(data.ResourceName).Key("site_config.0.ip_restriction.1.ip_address").HasValue("20.20.20.0/24"),
				check.That(data.ResourceName).Key("site_config.0.ip_restriction.2.ip_address").HasValue("30.30.0.0/16"),
				check.That(data.ResourceName).Key("site_config.0.ip_restriction.3.ip_address").HasValue("192.168.1.2/24"),
			),
		},
	})
}

func TestAccFunctionAppDataSource_clientCertMode(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_function_app", "test")

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: FunctionAppDataSource{}.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("client_cert_mode").HasValue(""),
			),
		},
		{
			Config: FunctionAppDataSource{}.certClientMode(data, "Optional"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("client_cert_mode").HasValue("Optional"),
			),
		},
		{
			Config: FunctionAppDataSource{}.certClientMode(data, "Required"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("client_cert_mode").HasValue("Required"),
			),
		},
	})
}

func TestAccFunctionAppDataSource_siteConfigVnetRouteAllEnabled(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_function_app", "test")

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: FunctionAppDataSource{}.siteConfigVnetRouteAllEnabled(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("site_config.0.vnet_route_all_enabled").HasValue("true"),
			),
		},
	})
}

func (d FunctionAppDataSource) basic(data acceptance.TestData) string {
	template := FunctionAppResource{}.basic(data)
	return fmt.Sprintf(`
%s

data "azurerm_function_app" "test" {
  name                = azurerm_function_app.test.name
  resource_group_name = azurerm_resource_group.test.name
}
`, template)
}

func (d FunctionAppDataSource) connectionStrings(data acceptance.TestData) string {
	template := FunctionAppResource{}.connectionStrings(data)
	return fmt.Sprintf(`
%s

data "azurerm_function_app" "test" {
  name                = azurerm_function_app.test.name
  resource_group_name = azurerm_resource_group.test.name
}
`, template)
}

func (d FunctionAppDataSource) appSettings(data acceptance.TestData) string {
	template := FunctionAppResource{}.appSettings(data)
	return fmt.Sprintf(`
%s

data "azurerm_function_app" "test" {
  name                = azurerm_function_app.test.name
  resource_group_name = azurerm_resource_group.test.name
}
`, template)
}

func (d FunctionAppDataSource) withSourceControl(data acceptance.TestData, branch string) string {
	config := FunctionAppResource{}.withSourceControl(data, branch)
	return fmt.Sprintf(`
%s

data "azurerm_function_app" "test" {
  name                = azurerm_function_app.test.name
  resource_group_name = azurerm_resource_group.test.name
}
`, config)
}

func (d FunctionAppDataSource) withSiteConfig(data acceptance.TestData) string {
	config := FunctionAppResource{}.manyIpRestrictions(data)
	return fmt.Sprintf(`
%s

data "azurerm_function_app" "test" {
  name                = azurerm_function_app.test.name
  resource_group_name = azurerm_resource_group.test.name
}
`, config)
}

func (d FunctionAppDataSource) certClientMode(data acceptance.TestData, modeValue string) string {
	template := FunctionAppResource{}.clientCertMode(data, modeValue)
	return fmt.Sprintf(`
%s

data "azurerm_function_app" "test" {
  name                = azurerm_function_app.test.name
  resource_group_name = azurerm_resource_group.test.name
}
`, template)
}

func (d FunctionAppDataSource) siteConfigVnetRouteAllEnabled(data acceptance.TestData) string {
	template := FunctionAppResource{}.siteConfigVnetRouteAllEnabled(data)
	return fmt.Sprintf(`
%s

data "azurerm_function_app" "test" {
  name                = azurerm_function_app.test.name
  resource_group_name = azurerm_resource_group.test.name
}
`, template)
}
