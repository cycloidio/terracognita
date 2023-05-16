package network_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
)

type ExpressRouteCircuitDataSource struct{}

func testAccDataSourceExpressRoute_basicMetered(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_express_route_circuit", "test")
	r := ExpressRouteCircuitDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("service_provider_properties.0.service_provider_name").HasValue("Equinix"),
				check.That(data.ResourceName).Key("service_provider_properties.0.peering_location").HasValue("Silicon Valley"),
				check.That(data.ResourceName).Key("service_provider_properties.0.bandwidth_in_mbps").HasValue("50"),
				check.That(data.ResourceName).Key("sku.0.tier").HasValue("Standard"),
				check.That(data.ResourceName).Key("sku.0.family").HasValue("MeteredData"),
				check.That(data.ResourceName).Key("service_provider_provisioning_state").HasValue("NotProvisioned"),
			),
		},
	})
}

func (ExpressRouteCircuitDataSource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurerm_express_route_circuit" "test" {
  resource_group_name = azurerm_resource_group.test.name
  name                = azurerm_express_route_circuit.test.name
}
`, ExpressRouteCircuitResource{}.basicMeteredConfig(data))
}
