package network_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	network2 "github.com/hashicorp/terraform-provider-azurerm/services/network"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type NetworkInterfaceApplicationGatewayBackendAddressPoolAssociationResource struct{}

func TestAccNetworkInterfaceApplicationGatewayBackendAddressPoolAssociation_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_network_interface_application_gateway_backend_address_pool_association", "test")
	r := NetworkInterfaceApplicationGatewayBackendAddressPoolAssociationResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		// intentional as this is a Virtual Resource
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccNetworkInterfaceApplicationGatewayBackendAddressPoolAssociation_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_network_interface_application_gateway_backend_address_pool_association", "test")
	r := NetworkInterfaceApplicationGatewayBackendAddressPoolAssociationResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		// intentional as this is a Virtual Resource
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_network_interface_application_gateway_backend_address_pool_association"),
		},
	})
}

func TestAccNetworkInterfaceApplicationGatewayBackendAddressPoolAssociation_deleted(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_network_interface_application_gateway_backend_address_pool_association", "test")
	r := NetworkInterfaceApplicationGatewayBackendAddressPoolAssociationResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		// intentionally not using a DisappearsStep as this is a Virtual Resource
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				data.CheckWithClient(r.destroy),
			),
			ExpectNonEmptyPlan: true,
		},
	})
}

func TestAccNetworkInterfaceApplicationGatewayBackendAddressPoolAssociation_updateNIC(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_network_interface_application_gateway_backend_address_pool_association", "test")
	r := NetworkInterfaceApplicationGatewayBackendAddressPoolAssociationResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		// intentional as this is a Virtual Resource
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.updateNIC(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (NetworkInterfaceApplicationGatewayBackendAddressPoolAssociationResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	splitId := strings.Split(state.ID, "|")
	if len(splitId) != 2 {
		return nil, fmt.Errorf("expected ID to be in the format {networkInterfaceId}/ipConfigurations/{ipConfigurationName}|{backendAddressPoolId} but got %q", state.ID)
	}

	id, err := parse.NetworkInterfaceIpConfigurationID(splitId[0])
	if err != nil {
		return nil, err
	}

	backendAddressPoolId := splitId[1]

	read, err := clients.Network.InterfacesClient.Get(ctx, id.ResourceGroup, id.NetworkInterfaceName, "")
	if err != nil {
		return nil, fmt.Errorf("reading NetworkInterfaceApplicationGatewayBackendAddressPoolAssociation (%s): %+v", id, err)
	}

	nicProps := read.InterfacePropertiesFormat
	if nicProps == nil {
		return nil, fmt.Errorf("`properties` was nil for (%s): %+v", id, err)
	}

	ipConfigs := nicProps.IPConfigurations
	if ipConfigs == nil {
		return nil, fmt.Errorf("`properties.IPConfigurations` was nil for  (%s): %+v", id, err)
	}

	c := network2.FindNetworkInterfaceIPConfiguration(nicProps.IPConfigurations, id.IpConfigurationName)
	if c == nil {
		return nil, fmt.Errorf("IP configuration was nil for (%s): %+v", id, err)
	}
	config := *c

	found := false
	if props := config.InterfaceIPConfigurationPropertiesFormat; props != nil {
		if backendPools := props.ApplicationGatewayBackendAddressPools; backendPools != nil {
			for _, pool := range *backendPools {
				if pool.ID == nil {
					continue
				}

				if *pool.ID == backendAddressPoolId {
					found = true
					break
				}
			}
		}
	}

	return utils.Bool(found), nil
}

func (NetworkInterfaceApplicationGatewayBackendAddressPoolAssociationResource) destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) error {
	nicID, err := parse.NetworkInterfaceID(state.Attributes["network_interface_id"])
	if err != nil {
		return err
	}

	backendAddressPoolId := state.Attributes["backend_address_pool_id"]
	ipConfigurationName := state.Attributes["ip_configuration_name"]

	read, err := client.Network.InterfacesClient.Get(ctx, nicID.ResourceGroup, nicID.Name, "")
	if err != nil {
		return fmt.Errorf("retrieving %s: %+v", nicID, err)
	}

	c := network2.FindNetworkInterfaceIPConfiguration(read.InterfacePropertiesFormat.IPConfigurations, ipConfigurationName)
	if c == nil {
		return fmt.Errorf("IP Configuration %q wasn't found for %s", ipConfigurationName, nicID)
	}
	config := *c

	updatedPools := make([]network.ApplicationGatewayBackendAddressPool, 0)
	if config.InterfaceIPConfigurationPropertiesFormat.ApplicationGatewayBackendAddressPools != nil {
		for _, pool := range *config.InterfaceIPConfigurationPropertiesFormat.ApplicationGatewayBackendAddressPools {
			if *pool.ID != backendAddressPoolId {
				updatedPools = append(updatedPools, pool)
			}
		}
	}
	config.InterfaceIPConfigurationPropertiesFormat.ApplicationGatewayBackendAddressPools = &updatedPools

	future, err := client.Network.InterfacesClient.CreateOrUpdate(ctx, nicID.ResourceGroup, nicID.Name, read)
	if err != nil {
		return fmt.Errorf("removing Application Gateway Backend Address Pool Association for %s: %+v", nicID, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Network.InterfacesClient.Client); err != nil {
		return fmt.Errorf("waiting for removal of Application Gateway Backend Address Pool Association for %s: %+v", nicID, err)
	}

	return nil
}

func (r NetworkInterfaceApplicationGatewayBackendAddressPoolAssociationResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_network_interface" "test" {
  name                = "acctestni-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurerm_subnet.backend.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_network_interface_application_gateway_backend_address_pool_association" "test" {
  network_interface_id    = azurerm_network_interface.test.id
  ip_configuration_name   = "testconfiguration1"
  backend_address_pool_id = tolist(azurerm_application_gateway.test.backend_address_pool).0.id
}
`, r.template(data), data.RandomInteger)
}

func (r NetworkInterfaceApplicationGatewayBackendAddressPoolAssociationResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_network_interface_application_gateway_backend_address_pool_association" "import" {
  network_interface_id    = azurerm_network_interface_application_gateway_backend_address_pool_association.test.network_interface_id
  ip_configuration_name   = azurerm_network_interface_application_gateway_backend_address_pool_association.test.ip_configuration_name
  backend_address_pool_id = tolist(azurerm_application_gateway.test.backend_address_pool).0.id
}
`, r.basic(data))
}

func (r NetworkInterfaceApplicationGatewayBackendAddressPoolAssociationResource) updateNIC(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_network_interface" "test" {
  name                = "acctestni-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurerm_subnet.backend.id}"
    private_ip_address_allocation = "Dynamic"
    primary                       = true
  }

  ip_configuration {
    name                          = "testconfiguration2"
    private_ip_address_version    = "IPv6"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_network_interface_application_gateway_backend_address_pool_association" "test" {
  network_interface_id    = azurerm_network_interface.test.id
  ip_configuration_name   = "testconfiguration1"
  backend_address_pool_id = tolist(azurerm_application_gateway.test.backend_address_pool).0.id
}
`, r.template(data), data.RandomInteger)
}

func (NetworkInterfaceApplicationGatewayBackendAddressPoolAssociationResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctestvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_subnet" "frontend" {
  name                 = "frontend"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_subnet" "backend" {
  name                 = "backend"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.4.0/24"]
}

resource "azurerm_public_ip" "test" {
  name                = "acctestpip%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  allocation_method   = "Dynamic"
}

# since these variables are re-used - a locals block makes this more maintainable
locals {
  backend_address_pool_name      = "${azurerm_virtual_network.test.name}-beap"
  frontend_port_name             = "${azurerm_virtual_network.test.name}-feport"
  frontend_ip_configuration_name = "${azurerm_virtual_network.test.name}-feip"
  http_setting_name              = "${azurerm_virtual_network.test.name}-be-htst"
  listener_name                  = "${azurerm_virtual_network.test.name}-httplstn"
  request_routing_rule_name      = "${azurerm_virtual_network.test.name}-rqrt"
}

resource "azurerm_application_gateway" "test" {
  name                = "apptestag%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location

  sku {
    name     = "Standard_Small"
    tier     = "Standard"
    capacity = 2
  }

  gateway_ip_configuration {
    name      = "my-gateway-ip-configuration"
    subnet_id = azurerm_subnet.frontend.id
  }

  frontend_port {
    name = local.frontend_port_name
    port = 80
  }

  frontend_ip_configuration {
    name                 = local.frontend_ip_configuration_name
    public_ip_address_id = azurerm_public_ip.test.id
  }

  backend_address_pool {
    name = local.backend_address_pool_name
  }

  backend_http_settings {
    name                  = local.http_setting_name
    cookie_based_affinity = "Disabled"
    port                  = 80
    protocol              = "Http"
    request_timeout       = 1
  }

  http_listener {
    name                           = local.listener_name
    frontend_ip_configuration_name = local.frontend_ip_configuration_name
    frontend_port_name             = local.frontend_port_name
    protocol                       = "Http"
  }

  request_routing_rule {
    name                       = local.request_routing_rule_name
    rule_type                  = "Basic"
    http_listener_name         = local.listener_name
    backend_address_pool_name  = local.backend_address_pool_name
    backend_http_settings_name = local.http_setting_name
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}
