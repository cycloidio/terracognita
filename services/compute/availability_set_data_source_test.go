package compute_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
)

type AvailabilitySetDataSource struct{}

func TestAccDataSourceAvailabilitySet_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_availability_set", "test")
	r := AvailabilitySetDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("location").Exists(),
				check.That(data.ResourceName).Key("name").Exists(),
				check.That(data.ResourceName).Key("resource_group_name").Exists(),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
			),
		},
	})
}

func (AvailabilitySetDataSource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurerm_availability_set" "test" {
  name                = "acctestavset-%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  tags = {
    "foo" = "bar"
  }
}

data "azurerm_availability_set" "test" {
  resource_group_name = azurerm_resource_group.test.name
  name                = azurerm_availability_set.test.name
}
`, data.RandomInteger, data.Locations.Primary)
}
