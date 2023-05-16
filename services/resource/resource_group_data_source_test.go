package resource_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
)

type ResourceGroupDataSource struct{}

func TestAccDataSourceAzureRMResourceGroup_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_resource_group", "test")
	r := ResourceGroupDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("name").HasValue(fmt.Sprintf("acctestRg-%d", data.RandomInteger)),
				check.That(data.ResourceName).Key("location").HasValue(azure.NormalizeLocation(data.Locations.Primary)),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.env").HasValue("test"),
			),
		},
	})
}

func (ResourceGroupDataSource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRg-%d"
  location = "%s"

  tags = {
    env = "test"
  }
}

data "azurerm_resource_group" "test" {
  name = azurerm_resource_group.test.name
}
`, data.RandomInteger, data.Locations.Primary)
}
