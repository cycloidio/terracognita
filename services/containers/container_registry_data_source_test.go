package containers_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
)

type ContainerRegistryDataSource struct{}

func TestAccDataSourceAzureRMContainerRegistry_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_container_registry", "test")
	r := ContainerRegistryDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("name").Exists(),
				check.That(data.ResourceName).Key("resource_group_name").Exists(),
				check.That(data.ResourceName).Key("location").Exists(),
				check.That(data.ResourceName).Key("admin_enabled").Exists(),
				check.That(data.ResourceName).Key("login_server").Exists(),
			),
		},
	})
}

func (ContainerRegistryDataSource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurerm_container_registry" "test" {
  name                = azurerm_container_registry.test.name
  resource_group_name = azurerm_container_registry.test.resource_group_name
}
`, ContainerRegistryResource{}.basicManaged(data, "Basic"))
}
