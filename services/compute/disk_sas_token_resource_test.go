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

type ManagedDiskSASTokenResource struct{}

func TestAccManagedDiskSASToken_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_managed_disk_sas_token", "test")
	r := ManagedDiskSASTokenResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
	})
}

func (t ManagedDiskSASTokenResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.ManagedDiskID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Compute.DisksClient.Get(ctx, id.ResourceGroup, id.DiskName)
	if err != nil {
		return nil, fmt.Errorf("retrieving Compute Disk Export status %q", id.String())
	}

	if resp.DiskState != "ActiveSAS" {
		return nil, fmt.Errorf("Disk SAS token %s (resource group %s): %s", id.DiskName, id.ResourceGroup, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (r ManagedDiskSASTokenResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-revokedisk-%d"
  location = "%s"
}

resource "azurerm_managed_disk" "test" {
  name                 = "acctestsads%s"
  location             = azurerm_resource_group.test.location
  resource_group_name  = azurerm_resource_group.test.name
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "1"
}

resource "azurerm_managed_disk_sas_token" "test" {
  managed_disk_id     = azurerm_managed_disk.test.id
  duration_in_seconds = 300
  access_level        = "Read"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString)
}
