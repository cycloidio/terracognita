package loadbalancer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/loadbalancer/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type LoadBalancerNatPool struct{}

func TestAccAzureRMLoadBalancerNatPool_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_lb_nat_pool", "test")
	r := LoadBalancerNatPool{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data, "Basic"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAzureRMLoadBalancerNatPool_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_lb_nat_pool", "test")
	r := LoadBalancerNatPool{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data, "Standard"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAzureRMLoadBalancerNatPool_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_lb_nat_pool", "test")
	r := LoadBalancerNatPool{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data, "Standard"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data, "Standard"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config: r.basic(data, "Standard"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAzureRMLoadBalancerNatPool_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_lb_nat_pool", "test")
	r := LoadBalancerNatPool{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data, "Basic"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func TestAccAzureRMLoadBalancerNatPool_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_lb_nat_pool", "test")
	r := LoadBalancerNatPool{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config: func(data acceptance.TestData) string {
				return r.basic(data, "Basic")
			},
			TestResource: r,
		}),
	})
}

func TestAccAzureRMLoadBalancerNatPool_updateMultiplePools(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_lb_nat_pool", "test")
	data2 := acceptance.BuildTestData(t, "azurerm_lb_nat_pool", "test2")

	r := LoadBalancerNatPool{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.multiplePools(data, data2),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data2.ResourceName).ExistsInAzure(r),
				check.That(data2.ResourceName).Key("backend_port").HasValue("3390"),
			),
		},
		data.ImportStep(),
		{
			Config: r.multiplePoolsUpdate(data, data2),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data2.ResourceName).ExistsInAzure(r),
				check.That(data2.ResourceName).Key("backend_port").HasValue("3391"),
			),
		},
		data.ImportStep(),
	})
}

func (r LoadBalancerNatPool) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.LoadBalancerInboundNatPoolID(state.ID)
	if err != nil {
		return nil, err
	}

	lb, err := client.LoadBalancers.LoadBalancersClient.Get(ctx, id.ResourceGroup, id.LoadBalancerName, "")
	if err != nil {
		if utils.ResponseWasNotFound(lb.Response) {
			return nil, fmt.Errorf("Load Balancer %q (resource group %q) not found for Nat Pool %q", id.LoadBalancerName, id.ResourceGroup, id.InboundNatPoolName)
		}
		return nil, fmt.Errorf("failed reading Load Balancer %q (resource group %q) for Nat Pool %q", id.LoadBalancerName, id.ResourceGroup, id.InboundNatPoolName)
	}
	props := lb.LoadBalancerPropertiesFormat
	if props == nil || props.InboundNatPools == nil || len(*props.InboundNatPools) == 0 {
		return nil, fmt.Errorf("Nat Pool %q not found in Load Balancer %q (resource group %q)", id.InboundNatPoolName, id.LoadBalancerName, id.ResourceGroup)
	}

	found := false
	for _, v := range *props.InboundNatPools {
		if v.Name != nil && *v.Name == id.InboundNatPoolName {
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("Nat Pool %q not found in Load Balancer %q (resource group %q)", id.InboundNatPoolName, id.LoadBalancerName, id.ResourceGroup)
	}

	return utils.Bool(found), nil
}

func (r LoadBalancerNatPool) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.LoadBalancerInboundNatPoolID(state.ID)
	if err != nil {
		return nil, err
	}

	lb, err := client.LoadBalancers.LoadBalancersClient.Get(ctx, id.ResourceGroup, id.LoadBalancerName, "")
	if err != nil {
		return nil, fmt.Errorf("retrieving Load Balancer %q (Resource Group %q)", id.LoadBalancerName, id.ResourceGroup)
	}
	if lb.LoadBalancerPropertiesFormat == nil {
		return nil, fmt.Errorf("`properties` was nil")
	}
	if lb.LoadBalancerPropertiesFormat.InboundNatPools == nil {
		return nil, fmt.Errorf("`properties.InboundNatPools` was nil")
	}

	inboundNatPools := make([]network.InboundNatPool, 0)
	for _, inboundNatPool := range *lb.LoadBalancerPropertiesFormat.InboundNatPools {
		if inboundNatPool.Name == nil || *inboundNatPool.Name == id.InboundNatPoolName {
			continue
		}

		inboundNatPools = append(inboundNatPools, inboundNatPool)
	}
	lb.LoadBalancerPropertiesFormat.InboundNatPools = &inboundNatPools

	future, err := client.LoadBalancers.LoadBalancersClient.CreateOrUpdate(ctx, id.ResourceGroup, id.LoadBalancerName, lb)
	if err != nil {
		return nil, fmt.Errorf("updating Load Balancer %q (Resource Group %q): %+v", id.LoadBalancerName, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.LoadBalancers.LoadBalancersClient.Client); err != nil {
		return nil, fmt.Errorf("waiting for update of Load Balancer %q (Resource Group %q): %+v", id.LoadBalancerName, id.ResourceGroup, err)
	}

	return utils.Bool(true), nil
}

func (r LoadBalancerNatPool) basic(data acceptance.TestData, sku string) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurerm_public_ip" "test" {
  name                = "test-ip-%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  allocation_method   = "Static"
  sku                 = "%[3]s"
}

resource "azurerm_lb" "test" {
  name                = "arm-test-loadbalancer-%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku                 = "%[3]s"

  frontend_ip_configuration {
    name                 = "one-%[1]d"
    public_ip_address_id = azurerm_public_ip.test.id
  }
}

resource "azurerm_lb_nat_pool" "test" {
  resource_group_name            = azurerm_resource_group.test.name
  loadbalancer_id                = azurerm_lb.test.id
  name                           = "NatPool-%[1]d"
  protocol                       = "Tcp"
  frontend_port_start            = 80
  frontend_port_end              = 81
  backend_port                   = 3389
  frontend_ip_configuration_name = "one-%[1]d"
}
`, data.RandomInteger, data.Locations.Primary, sku)
}

func (r LoadBalancerNatPool) complete(data acceptance.TestData, sku string) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}
resource "azurerm_public_ip" "test" {
  name                = "test-ip-%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  allocation_method   = "Static"
  sku                 = "%[3]s"
}
resource "azurerm_lb" "test" {
  name                = "arm-test-loadbalancer-%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku                 = "%[3]s"

  frontend_ip_configuration {
    name                 = "one-%[1]d"
    public_ip_address_id = azurerm_public_ip.test.id
  }
}
resource "azurerm_lb_nat_pool" "test" {
  resource_group_name            = azurerm_resource_group.test.name
  loadbalancer_id                = azurerm_lb.test.id
  name                           = "NatPool-%[1]d"
  protocol                       = "Tcp"
  frontend_port_start            = 80
  frontend_port_end              = 81
  backend_port                   = 3389
  frontend_ip_configuration_name = "one-%[1]d"
  floating_ip_enabled            = true
  tcp_reset_enabled              = true
  idle_timeout_in_minutes        = 10
}
`, data.RandomInteger, data.Locations.Primary, sku)
}

func (r LoadBalancerNatPool) requiresImport(data acceptance.TestData) string {
	template := r.basic(data, "Basic")
	return fmt.Sprintf(`
%s

resource "azurerm_lb_nat_pool" "import" {
  name                           = azurerm_lb_nat_pool.test.name
  loadbalancer_id                = azurerm_lb_nat_pool.test.loadbalancer_id
  resource_group_name            = azurerm_lb_nat_pool.test.resource_group_name
  frontend_ip_configuration_name = azurerm_lb_nat_pool.test.frontend_ip_configuration_name
  protocol                       = "Tcp"
  frontend_port_start            = 80
  frontend_port_end              = 81
  backend_port                   = 3389
}
`, template)
}

func (r LoadBalancerNatPool) multiplePools(data, data2 acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_public_ip" "test" {
  name                = "test-ip-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  allocation_method = "Static"
}

resource "azurerm_lb" "test" {
  name                = "arm-test-loadbalancer-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = azurerm_public_ip.test.id
  }
}

resource "azurerm_lb_nat_pool" "test" {
  resource_group_name = azurerm_resource_group.test.name
  loadbalancer_id     = azurerm_lb.test.id
  name                = "NatPool-%d"
  protocol            = "Tcp"
  frontend_port_start = 80
  frontend_port_end   = 81
  backend_port        = 3389

  frontend_ip_configuration_name = "one-%d"
}

resource "azurerm_lb_nat_pool" "test2" {
  resource_group_name = azurerm_resource_group.test.name
  loadbalancer_id     = azurerm_lb.test.id
  name                = "NatPool-%d"
  protocol            = "Tcp"
  frontend_port_start = 82
  frontend_port_end   = 83
  backend_port        = 3390

  frontend_ip_configuration_name = "one-%d"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data2.RandomInteger, data.RandomInteger)
}

func (r LoadBalancerNatPool) multiplePoolsUpdate(data, data2 acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_public_ip" "test" {
  name                = "test-ip-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  allocation_method   = "Static"
}

resource "azurerm_lb" "test" {
  name                = "arm-test-loadbalancer-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = azurerm_public_ip.test.id
  }
}

resource "azurerm_lb_nat_pool" "test" {
  resource_group_name            = azurerm_resource_group.test.name
  loadbalancer_id                = azurerm_lb.test.id
  name                           = "NatPool-%d"
  protocol                       = "Tcp"
  frontend_port_start            = 80
  frontend_port_end              = 81
  backend_port                   = 3389
  frontend_ip_configuration_name = "one-%d"
}

resource "azurerm_lb_nat_pool" "test2" {
  resource_group_name            = azurerm_resource_group.test.name
  loadbalancer_id                = azurerm_lb.test.id
  name                           = "NatPool-%d"
  protocol                       = "Tcp"
  frontend_port_start            = 82
  frontend_port_end              = 83
  backend_port                   = 3391
  frontend_ip_configuration_name = "one-%d"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data2.RandomInteger, data.RandomInteger)
}
