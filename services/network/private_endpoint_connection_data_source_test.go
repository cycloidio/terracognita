package network_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
)

type PrivateEndpointConnectionDataSource struct{}

func TestAccDataSourcePrivateEndpointConnection_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_private_endpoint_connection", "test")
	r := PrivateEndpointConnectionDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("network_interface.0.id").Exists(),
				check.That(data.ResourceName).Key("network_interface.0.name").Exists(),
				check.That(data.ResourceName).Key("private_service_connection.0.status").HasValue("Approved"),
			),
		},
	})
}

func (PrivateEndpointConnectionDataSource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurerm_private_endpoint_connection" "test" {
  name                = azurerm_private_endpoint.test.name
  resource_group_name = azurerm_resource_group.test.name
}
`, PrivateEndpointResource{}.basic(data))
}
