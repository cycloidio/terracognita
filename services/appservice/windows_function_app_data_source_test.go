package appservice_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
)

type WindowsFunctionAppDataSource struct{}

func TestAccWindowsFunctionAppDataSource_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_windows_function_app", "test")
	d := WindowsFunctionAppDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: d.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("location").HasValue(data.Locations.Primary),
			),
		},
	})
}

func (WindowsFunctionAppDataSource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data azurerm_windows_function_app test {
  name                = azurerm_windows_function_app.test.name
  resource_group_name = azurerm_windows_function_app.test.resource_group_name
}
`, WindowsFunctionAppResource{}.standardComplete(data))
}
