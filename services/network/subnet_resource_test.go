package network_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type SubnetResource struct{}

func TestAccSubnet_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_subnet", "test")
	r := SubnetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSubnet_basic_addressPrefixes(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_subnet", "test")
	r := SubnetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic_addressPrefixes(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSubnet_complete_addressPrefixes(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_subnet", "test")
	r := SubnetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete_addressPrefixes(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSubnet_update_addressPrefixes(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_subnet", "test")
	r := SubnetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic_addressPrefixes(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete_addressPrefixes(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.basic_addressPrefixes(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSubnet_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_subnet", "test")
	r := SubnetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_subnet"),
		},
	})
}

func TestAccSubnet_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_subnet", "test")
	r := SubnetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       r.basic,
			TestResource: r,
		}),
	})
}

func TestAccSubnet_delegation(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_subnet", "test")
	r := SubnetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.delegation(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.delegationUpdated(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.delegation(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSubnet_enforcePrivateLinkEndpointNetworkPolicies(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_subnet", "test")
	r := SubnetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.enforcePrivateLinkEndpointNetworkPolicies(data, true),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.enforcePrivateLinkEndpointNetworkPolicies(data, false),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.enforcePrivateLinkEndpointNetworkPolicies(data, true),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSubnet_enforcePrivateLinkServiceNetworkPolicies(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_subnet", "test")
	r := SubnetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.enforcePrivateLinkServiceNetworkPolicies(data, true),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.enforcePrivateLinkServiceNetworkPolicies(data, false),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.enforcePrivateLinkServiceNetworkPolicies(data, true),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSubnet_serviceEndpoints(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_subnet", "test")
	r := SubnetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.serviceEndpoints(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.serviceEndpointsUpdated(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			// remove them
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.serviceEndpoints(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSubnet_serviceEndpointPolicy(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_subnet", "test")
	r := SubnetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.serviceEndpointPolicyBasic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.serviceEndpointPolicyUpdate(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.serviceEndpointPolicyBasic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
	})
}

func TestAccSubnet_updateAddressPrefix(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_subnet", "test")
	r := SubnetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.updatedAddressPrefix(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (t SubnetResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.SubnetID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Network.SubnetsClient.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name, "")
	if err != nil {
		return nil, fmt.Errorf("reading Subnet (%s): %+v", id, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (SubnetResource) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.SubnetID(state.ID)
	if err != nil {
		return nil, err
	}

	future, err := client.Network.SubnetsClient.Delete(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name)
	if err != nil {
		return nil, fmt.Errorf("deleting Subnet %q: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Network.SubnetsClient.Client); err != nil {
		return nil, fmt.Errorf("waiting for deletion of Subnet %q: %+v", id, err)
	}

	return utils.Bool(true), nil
}

func (SubnetResource) hasNoNatGateway(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) error {
	id, err := parse.SubnetID(state.ID)
	if err != nil {
		return err
	}

	subnet, err := client.Network.SubnetsClient.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(subnet.Response) {
			return fmt.Errorf("Bad: Subnet %q (Virtual Network %q / Resource Group: %q) does not exist", id.Name, id.VirtualNetworkName, id.ResourceGroup)
		}
		return fmt.Errorf("Bad: Get on subnetClient: %+v", err)
	}

	props := subnet.SubnetPropertiesFormat
	if props == nil {
		return fmt.Errorf("Properties was nil for Subnet %q (Virtual Network %q / Resource Group: %q)", id.Name, id.VirtualNetworkName, id.ResourceGroup)
	}

	if props.NatGateway != nil && ((props.NatGateway.ID == nil) || (props.NatGateway.ID != nil && *props.NatGateway.ID == "")) {
		return fmt.Errorf("No Route Table should exist for Subnet %q (Virtual Network %q / Resource Group: %q) but got %q", id.Name, id.VirtualNetworkName, id.ResourceGroup, *props.RouteTable.ID)
	}
	return nil
}

func (SubnetResource) hasNoNetworkSecurityGroup(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) error {
	id, err := parse.SubnetID(state.ID)
	if err != nil {
		return err
	}

	resp, err := client.Network.SubnetsClient.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Bad: Subnet %q (Virtual Network %q / Resource Group: %q) does not exist", id.Name, id.VirtualNetworkName, id.ResourceGroup)
		}

		return fmt.Errorf("Bad: Get on subnetClient: %+v", err)
	}

	props := resp.SubnetPropertiesFormat
	if props == nil {
		return fmt.Errorf("Properties was nil for Subnet %q (Virtual Network %q / Resource Group: %q)", id.Name, id.VirtualNetworkName, id.ResourceGroup)
	}

	if props.NetworkSecurityGroup != nil && ((props.NetworkSecurityGroup.ID == nil) || (props.NetworkSecurityGroup.ID != nil && *props.NetworkSecurityGroup.ID == "")) {
		return fmt.Errorf("No Network Security Group should exist for Subnet %q (Virtual Network %q / Resource Group: %q) but got %q", id.Name, id.VirtualNetworkName, id.ResourceGroup, *props.NetworkSecurityGroup.ID)
	}

	return nil
}

func (SubnetResource) hasNoRouteTable(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) error {
	id, err := parse.SubnetID(state.ID)
	if err != nil {
		return err
	}

	resp, err := client.Network.SubnetsClient.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Bad: Subnet %q (Virtual Network %q / Resource Group: %q) does not exist", id.Name, id.VirtualNetworkName, id.ResourceGroup)
		}

		return fmt.Errorf("Bad: Get on subnetClient: %+v", err)
	}

	props := resp.SubnetPropertiesFormat
	if props == nil {
		return fmt.Errorf("Properties was nil for Subnet %q (Virtual Network %q / Resource Group: %q)", id.Name, id.VirtualNetworkName, id.ResourceGroup)
	}

	if props.RouteTable != nil && ((props.RouteTable.ID == nil) || (props.RouteTable.ID != nil && *props.RouteTable.ID == "")) {
		return fmt.Errorf("No Route Table should exist for Subnet %q (Virtual Network %q / Resource Group: %q) but got %q", id.Name, id.VirtualNetworkName, id.ResourceGroup, *props.RouteTable.ID)
	}

	return nil
}

func (r SubnetResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_subnet" "test2" {
  name                 = "internal2"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.3.0/24"]
}
`, r.template(data))
}

func (r SubnetResource) delegation(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.2.0/24"]

  delegation {
    name = "first"

    service_delegation {
      name = "Microsoft.ContainerInstance/containerGroups"
      actions = [
        "Microsoft.Network/virtualNetworks/subnets/action",
      ]
    }
  }
}
`, r.template(data))
}

func (r SubnetResource) delegationUpdated(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.2.0/24"]

  delegation {
    name = "first"

    service_delegation {
      name = "Microsoft.Databricks/workspaces"
      actions = [
        "Microsoft.Network/virtualNetworks/subnets/join/action",
        "Microsoft.Network/virtualNetworks/subnets/prepareNetworkPolicies/action",
        "Microsoft.Network/virtualNetworks/subnets/unprepareNetworkPolicies/action",
      ]
    }
  }
}
`, r.template(data))
}

func (r SubnetResource) enforcePrivateLinkEndpointNetworkPolicies(data acceptance.TestData, enabled bool) string {
	return fmt.Sprintf(`
%s

resource "azurerm_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.2.0/24"]

  enforce_private_link_endpoint_network_policies = %t
}
`, r.template(data), enabled)
}

func (r SubnetResource) enforcePrivateLinkServiceNetworkPolicies(data acceptance.TestData, enabled bool) string {
	return fmt.Sprintf(`
%s

resource "azurerm_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.2.0/24"]

  enforce_private_link_service_network_policies = %t
}
`, r.template(data), enabled)
}

func (SubnetResource) basic_addressPrefixes(data acceptance.TestData) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-n-%d"
  location = "%s"
}
resource "azurerm_virtual_network" "test" {
  name                = "acctestvirtnet%d"
  address_space       = ["10.0.0.0/16", "ace:cab:deca::/48"]
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
}
resource "azurerm_subnet" "test" {
  name                 = "acctestsubnet%d"
  resource_group_name  = "${azurerm_resource_group.test.name}"
  virtual_network_name = "${azurerm_virtual_network.test.name}"
  address_prefixes     = ["10.0.0.0/24"]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (SubnetResource) complete_addressPrefixes(data acceptance.TestData) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-n-%d"
  location = "%s"
}
resource "azurerm_virtual_network" "test" {
  name                = "acctestvirtnet%d"
  address_space       = ["10.0.0.0/16", "ace:cab:deca::/48"]
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
}
resource "azurerm_subnet" "test" {
  name                 = "acctestsubnet%d"
  resource_group_name  = "${azurerm_resource_group.test.name}"
  virtual_network_name = "${azurerm_virtual_network.test.name}"
  address_prefixes     = ["10.0.0.0/24", "ace:cab:deca:deed::/64"]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (r SubnetResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_subnet" "import" {
  name                 = azurerm_subnet.test.name
  resource_group_name  = azurerm_subnet.test.resource_group_name
  virtual_network_name = azurerm_subnet.test.virtual_network_name
  address_prefixes     = azurerm_subnet.test.address_prefixes
}
`, r.basic(data))
}

func (r SubnetResource) serviceEndpoints(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.2.0/24"]
  service_endpoints    = ["Microsoft.Sql"]
}

resource "azurerm_subnet" "test2" {
  name                 = "internal2"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.3.0/24"]
  service_endpoints    = ["Microsoft.Sql"]
}
`, r.template(data))
}

func (r SubnetResource) serviceEndpointsUpdated(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.2.0/24"]
  service_endpoints    = ["Microsoft.Sql", "Microsoft.Storage"]
}

resource "azurerm_subnet" "test2" {
  name                 = "internal2"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.3.0/24"]
  service_endpoints    = ["Microsoft.Sql", "Microsoft.Storage"]
}
`, r.template(data))
}

func (r SubnetResource) serviceEndpointPolicyBasic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_subnet_service_endpoint_storage_policy" "test" {
  name                = "acctestSEP-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
}

resource "azurerm_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.2.0/24"]
}
`, r.template(data), data.RandomInteger)
}

func (r SubnetResource) serviceEndpointPolicyUpdate(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_subnet_service_endpoint_storage_policy" "test" {
  name                = "acctestSEP-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
}

resource "azurerm_subnet" "test" {
  name                        = "internal"
  resource_group_name         = azurerm_resource_group.test.name
  virtual_network_name        = azurerm_virtual_network.test.name
  address_prefixes            = ["10.0.2.0/24"]
  service_endpoints           = ["Microsoft.Sql"]
  service_endpoint_policy_ids = [azurerm_subnet_service_endpoint_storage_policy.test.id]
}
`, r.template(data), data.RandomInteger)
}

func (r SubnetResource) updatedAddressPrefix(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.3.0/24"]
}
`, r.template(data))
}

func (SubnetResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctestvirtnet%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
