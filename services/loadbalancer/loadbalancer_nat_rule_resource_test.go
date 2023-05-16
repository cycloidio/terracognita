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

type LoadBalancerNatRule struct{}

func TestAccAzureRMLoadBalancerNatRule_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_lb_nat_rule", "test")
	r := LoadBalancerNatRule{}

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

func TestAccAzureRMLoadBalancerNatRule_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_lb_nat_rule", "test")
	r := LoadBalancerNatRule{}

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

func TestAccAzureRMLoadBalancerNatRule_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_lb_nat_rule", "test")
	r := LoadBalancerNatRule{}

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
		data.ImportStep(),
		{
			Config: r.basic(data, "Standard"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAzureRMLoadBalancerNatRule_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_lb_nat_rule", "test")
	r := LoadBalancerNatRule{}

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

func TestAccAzureRMLoadBalancerNatRule_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_lb_nat_rule", "test")
	r := LoadBalancerNatRule{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config: func(data acceptance.TestData) string {
				return r.basic(data, "Basic")
			},
			TestResource: r,
		}),
	})
}

func TestAccAzureRMLoadBalancerNatRule_updateMultipleRules(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_lb_nat_rule", "test")
	data2 := acceptance.BuildTestData(t, "azurerm_lb_nat_rule", "test2")

	r := LoadBalancerNatRule{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.multipleRules(data, data2),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data2.ResourceName).ExistsInAzure(r),
				check.That(data2.ResourceName).Key("frontend_port").HasValue("3390"),
				check.That(data2.ResourceName).Key("backend_port").HasValue("3390"),
			),
		},
		data.ImportStep(),
		{
			Config: r.multipleRulesUpdate(data, data2),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data2.ResourceName).ExistsInAzure(r),
				check.That(data2.ResourceName).Key("frontend_port").HasValue("3391"),
				check.That(data2.ResourceName).Key("backend_port").HasValue("3391"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAzureRMLoadBalancerNatRule_zeroPortNumber(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_lb_nat_rule", "test")
	r := LoadBalancerNatRule{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.zeroPortNumber(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (r LoadBalancerNatRule) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.LoadBalancerInboundNatRuleID(state.ID)
	if err != nil {
		return nil, err
	}

	lb, err := client.LoadBalancers.LoadBalancersClient.Get(ctx, id.ResourceGroup, id.LoadBalancerName, "")
	if err != nil {
		if utils.ResponseWasNotFound(lb.Response) {
			return nil, fmt.Errorf("Load Balancer %q (resource group %q) not found for Nat Rule %q", id.LoadBalancerName, id.ResourceGroup, id.InboundNatRuleName)
		}
		return nil, fmt.Errorf("failed reading Load Balancer %q (resource group %q) for Nat Rule %q", id.LoadBalancerName, id.ResourceGroup, id.InboundNatRuleName)
	}
	props := lb.LoadBalancerPropertiesFormat
	if props == nil || props.InboundNatRules == nil || len(*props.InboundNatRules) == 0 {
		return nil, fmt.Errorf("Nat Rule %q not found in Load Balancer %q (resource group %q)", id.InboundNatRuleName, id.LoadBalancerName, id.ResourceGroup)
	}

	found := false
	for _, v := range *props.InboundNatRules {
		if v.Name != nil && *v.Name == id.InboundNatRuleName {
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("Nat Rule %q not found in Load Balancer %q (resource group %q)", id.InboundNatRuleName, id.LoadBalancerName, id.ResourceGroup)
	}
	return utils.Bool(found), nil
}

func (r LoadBalancerNatRule) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.LoadBalancerInboundNatRuleID(state.ID)
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
	if lb.LoadBalancerPropertiesFormat.InboundNatRules == nil {
		return nil, fmt.Errorf("`properties.InboundNatRules` was nil")
	}

	inboundNatRules := make([]network.InboundNatRule, 0)
	for _, inboundNatRule := range *lb.LoadBalancerPropertiesFormat.InboundNatRules {
		if inboundNatRule.Name == nil || *inboundNatRule.Name == id.InboundNatRuleName {
			continue
		}

		inboundNatRules = append(inboundNatRules, inboundNatRule)
	}
	lb.LoadBalancerPropertiesFormat.InboundNatRules = &inboundNatRules

	future, err := client.LoadBalancers.LoadBalancersClient.CreateOrUpdate(ctx, id.ResourceGroup, id.LoadBalancerName, lb)
	if err != nil {
		return nil, fmt.Errorf("updating Load Balancer %q (Resource Group %q): %+v", id.LoadBalancerName, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.LoadBalancers.LoadBalancersClient.Client); err != nil {
		return nil, fmt.Errorf("waiting for update of Load Balancer %q (Resource Group %q): %+v", id.LoadBalancerName, id.ResourceGroup, err)
	}

	return utils.Bool(true), nil
}

func (r LoadBalancerNatRule) template(data acceptance.TestData, sku string) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-lb-%[1]d"
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
`, data.RandomInteger, data.Locations.Primary, sku)
}

func (r LoadBalancerNatRule) basic(data acceptance.TestData, sku string) string {
	template := r.template(data, sku)
	return fmt.Sprintf(`
%s

resource "azurerm_lb_nat_rule" "test" {
  resource_group_name            = "${azurerm_resource_group.test.name}"
  loadbalancer_id                = "${azurerm_lb.test.id}"
  name                           = "NatRule-%d"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = azurerm_lb.test.frontend_ip_configuration.0.name
}
`, template, data.RandomInteger)
}

func (r LoadBalancerNatRule) complete(data acceptance.TestData, sku string) string {
	template := r.template(data, sku)
	return fmt.Sprintf(`
%s

resource "azurerm_lb_nat_rule" "test" {
  name                = "NatRule-%d"
  resource_group_name = "${azurerm_resource_group.test.name}"
  loadbalancer_id     = "${azurerm_lb.test.id}"

  protocol      = "Tcp"
  frontend_port = 3389
  backend_port  = 3389

  enable_floating_ip      = true
  enable_tcp_reset        = true
  idle_timeout_in_minutes = 10

  frontend_ip_configuration_name = azurerm_lb.test.frontend_ip_configuration.0.name
}
`, template, data.RandomInteger)
}

func (r LoadBalancerNatRule) requiresImport(data acceptance.TestData) string {
	template := r.basic(data, "Basic")
	return fmt.Sprintf(`
%s

resource "azurerm_lb_nat_rule" "import" {
  name                           = azurerm_lb_nat_rule.test.name
  loadbalancer_id                = azurerm_lb_nat_rule.test.loadbalancer_id
  resource_group_name            = azurerm_lb_nat_rule.test.resource_group_name
  frontend_ip_configuration_name = azurerm_lb_nat_rule.test.frontend_ip_configuration_name
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
}
`, template)
}

func (r LoadBalancerNatRule) multipleRules(data, data2 acceptance.TestData) string {
	template := r.template(data, "Basic")
	return fmt.Sprintf(`
%s

resource "azurerm_lb_nat_rule" "test" {
  resource_group_name            = "${azurerm_resource_group.test.name}"
  loadbalancer_id                = "${azurerm_lb.test.id}"
  name                           = "NatRule-%d"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = azurerm_lb.test.frontend_ip_configuration.0.name
}

resource "azurerm_lb_nat_rule" "test2" {
  resource_group_name            = "${azurerm_resource_group.test.name}"
  loadbalancer_id                = "${azurerm_lb.test.id}"
  name                           = "NatRule-%d"
  protocol                       = "Tcp"
  frontend_port                  = 3390
  backend_port                   = 3390
  frontend_ip_configuration_name = azurerm_lb.test.frontend_ip_configuration.0.name
}
`, template, data.RandomInteger, data2.RandomInteger)
}

func (r LoadBalancerNatRule) multipleRulesUpdate(data, data2 acceptance.TestData) string {
	template := r.template(data, "Basic")
	return fmt.Sprintf(`
%s
resource "azurerm_lb_nat_rule" "test" {
  resource_group_name            = "${azurerm_resource_group.test.name}"
  loadbalancer_id                = "${azurerm_lb.test.id}"
  name                           = "NatRule-%d"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = azurerm_lb.test.frontend_ip_configuration.0.name
}

resource "azurerm_lb_nat_rule" "test2" {
  resource_group_name            = "${azurerm_resource_group.test.name}"
  loadbalancer_id                = "${azurerm_lb.test.id}"
  name                           = "NatRule-%d"
  protocol                       = "Tcp"
  frontend_port                  = 3391
  backend_port                   = 3391
  frontend_ip_configuration_name = azurerm_lb.test.frontend_ip_configuration.0.name
}
`, template, data.RandomInteger, data2.RandomInteger)
}

func (r LoadBalancerNatRule) zeroPortNumber(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-lb-%[1]d"
  location = "%[2]s"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctest-vnet-%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_subnet" "test" {
  name                 = "acctest-subnet-%[1]d"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_lb" "test" {
  name                = "acctest-lb-%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku                 = "Standard"

  frontend_ip_configuration {
    name                          = "Internal"
    private_ip_address_allocation = "Static"
    private_ip_address_version    = "IPv4"
    private_ip_address            = "10.0.2.7"
    subnet_id                     = azurerm_subnet.test.id
  }
}

resource "azurerm_lb_nat_rule" "test" {
  resource_group_name            = azurerm_resource_group.test.name
  loadbalancer_id                = azurerm_lb.test.id
  name                           = "NatRule-%[1]d"
  protocol                       = "All"
  frontend_port                  = 0
  backend_port                   = 0
  idle_timeout_in_minutes        = 4
  enable_floating_ip             = false
  enable_tcp_reset               = false
  frontend_ip_configuration_name = azurerm_lb.test.frontend_ip_configuration.0.name
}
`, data.RandomInteger, data.Locations.Primary)
}
