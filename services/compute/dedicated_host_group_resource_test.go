package compute_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/compute/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type DedicatedHostGroupResource struct{}

func TestAccDedicatedHostGroup_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_dedicated_host_group", "test")
	r := DedicatedHostGroupResource{}

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

func TestAccDedicatedHostGroup_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_dedicated_host_group", "test")
	r := DedicatedHostGroupResource{}

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

func TestAccDedicatedHostGroup_automaticPlacementEnabled(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_dedicated_host_group", "test")
	r := DedicatedHostGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.automaticPlacementEnabled(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccDedicatedHostGroup_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_dedicated_host_group", "test")
	r := DedicatedHostGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("platform_fault_domain_count").HasValue("2"),
				check.That(data.ResourceName).Key("tags.ENV").HasValue("prod"),
			),
		},
		data.ImportStep(),
	})
}

func (r DedicatedHostGroupResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.HostGroupID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Compute.DedicatedHostGroupsClient.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		return nil, fmt.Errorf("retrieving Compute Dedicated Host Group %q", id)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (DedicatedHostGroupResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-compute-%d"
  location = "%s"
}

resource "azurerm_dedicated_host_group" "test" {
  name                        = "acctestDHG-compute-%d"
  resource_group_name         = azurerm_resource_group.test.name
  location                    = azurerm_resource_group.test.location
  platform_fault_domain_count = 2
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r DedicatedHostGroupResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s
resource "azurerm_dedicated_host_group" "import" {
  resource_group_name         = azurerm_dedicated_host_group.test.resource_group_name
  name                        = azurerm_dedicated_host_group.test.name
  location                    = azurerm_dedicated_host_group.test.location
  platform_fault_domain_count = 2
}
`, r.basic(data))
}

func (DedicatedHostGroupResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-compute-%d"
  location = "%s"
}

resource "azurerm_dedicated_host_group" "test" {
  name                        = "acctestDHG-compute-%d"
  resource_group_name         = azurerm_resource_group.test.name
  location                    = azurerm_resource_group.test.location
  platform_fault_domain_count = 2
  zone                        = "1"
  tags = {
    ENV = "prod"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (DedicatedHostGroupResource) automaticPlacementEnabled(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-compute-%d"
  location = "%s"
}

resource "azurerm_dedicated_host_group" "test" {
  name                        = "acctestDHG-compute-%d"
  resource_group_name         = azurerm_resource_group.test.name
  location                    = azurerm_resource_group.test.location
  platform_fault_domain_count = 2

  automatic_placement_enabled = true
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
