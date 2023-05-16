package vmware_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/vmware/sdk/2020-03-20/privateclouds"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type VmwarePrivateCloudResource struct{}

func TestAccVmwarePrivateCloud_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_vmware_private_cloud", "test")
	r := VmwarePrivateCloudResource{}

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

func TestAccVmwarePrivateCloud_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_vmware_private_cloud", "test")
	r := VmwarePrivateCloudResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func TestAccVmwarePrivateCloud_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_vmware_private_cloud", "test")
	r := VmwarePrivateCloudResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("nsxt_password", "vcenter_password"),
	})
}

// Internet availability, cluster size, identity sources, vcenter password or nsxt password cannot be updated at the same time
func TestAccVmwarePrivateCloud_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_vmware_private_cloud", "test")
	r := VmwarePrivateCloudResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("nsxt_password", "vcenter_password"),
		{
			Config: r.update(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("nsxt_password", "vcenter_password"),
		{
			Config: r.update2(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("nsxt_password", "vcenter_password"),
		{
			Config: r.update3(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("nsxt_password", "vcenter_password"),
	})
}

func (VmwarePrivateCloudResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := privateclouds.ParsePrivateCloudID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Vmware.PrivateCloudClient.Get(ctx, *id)
	if err != nil {
		return nil, fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	return utils.Bool(resp.Model != nil), nil
}

func (VmwarePrivateCloudResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
  # In Vmware acctest, please disable correlation request id, else the continuous operations like update or delete will not be triggered
  # issue https://github.com/Azure/azure-rest-api-specs/issues/14086 
  disable_correlation_request_id = true
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-Vmware-%d"
  location = "%s"
}
`, data.RandomInteger, data.Locations.Primary)
}

func (r VmwarePrivateCloudResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_vmware_private_cloud" "test" {
  name                = "acctest-PC-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku_name            = "av36"

  management_cluster {
    size = 3
  }
  network_subnet_cidr = "192.168.48.0/22"
}
`, r.template(data), data.RandomInteger)
}

func (r VmwarePrivateCloudResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_vmware_private_cloud" "import" {
  name                = azurerm_vmware_private_cloud.test.name
  resource_group_name = azurerm_vmware_private_cloud.test.resource_group_name
  location            = azurerm_vmware_private_cloud.test.location
  sku_name            = azurerm_vmware_private_cloud.test.sku_name

  management_cluster {
    size = azurerm_vmware_private_cloud.test.management_cluster.0.size
  }
  network_subnet_cidr = azurerm_vmware_private_cloud.test.network_subnet_cidr
}
`, r.basic(data))
}

func (r VmwarePrivateCloudResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_vmware_private_cloud" "test" {
  name                = "acctest-PC-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku_name            = "av36"

  management_cluster {
    size = 3
  }
  network_subnet_cidr         = "192.168.48.0/22"
  internet_connection_enabled = false
  nsxt_password               = "QazWsx13$Edc"
  vcenter_password            = "WsxEdc23$Rfv"
  tags = {
    ENV = "Test"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r VmwarePrivateCloudResource) update(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_vmware_private_cloud" "test" {
  name                = "acctest-PC-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku_name            = "av36"

  management_cluster {
    size = 4
  }
  network_subnet_cidr         = "192.168.48.0/22"
  internet_connection_enabled = false
  nsxt_password               = "QazWsx13$Edc"
  vcenter_password            = "WsxEdc23$Rfv"
  tags = {
    ENV = "Stage"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r VmwarePrivateCloudResource) update2(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_vmware_private_cloud" "test" {
  name                = "acctest-PC-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku_name            = "av36"

  management_cluster {
    size = 4
  }
  network_subnet_cidr         = "192.168.48.0/22"
  internet_connection_enabled = true
  nsxt_password               = "QazWsx13$Edc"
  vcenter_password            = "WsxEdc23$Rfv"
  tags = {
    ENV = "Test"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r VmwarePrivateCloudResource) update3(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_vmware_private_cloud" "test" {
  name                = "acctest-PC-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku_name            = "av36"

  management_cluster {
    size = 3
  }
  network_subnet_cidr         = "192.168.48.0/22"
  internet_connection_enabled = true
  nsxt_password               = "QazWsx13$Edc"
  vcenter_password            = "WsxEdc23$Rfv"
  tags = {
    ENV = "Test"
  }
}
`, r.template(data), data.RandomInteger)
}
