package compute_test

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/compute/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type WindowsVirtualMachineScaleSetResource struct{}

func (r WindowsVirtualMachineScaleSetResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.VirtualMachineScaleSetID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Compute.VMScaleSetClient.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		return nil, fmt.Errorf("retrieving Compute Windows Virtual Machine Scale Set %q", id)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (WindowsVirtualMachineScaleSetResource) vmName(data acceptance.TestData) string {
	// windows VM names can be up to 15 chars, however the prefix can only be 9 chars
	return fmt.Sprintf("acctvm%s", fmt.Sprintf("%d", data.RandomInteger)[0:2])
}

func (r WindowsVirtualMachineScaleSetResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
locals {
  vm_name = "%s"
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctestnw-%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.2.0/24"]
}
`, r.vmName(data), data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
