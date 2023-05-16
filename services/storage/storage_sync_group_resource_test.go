package storage_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/storage/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type StorageSyncGroupResource struct{}

func TestAccStorageSyncGroup_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_sync_group", "test")
	r := StorageSyncGroupResource{}

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

func TestAccStorageSyncGroup_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_sync_group", "test")
	r := StorageSyncGroupResource{}

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

func (r StorageSyncGroupResource) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.StorageSyncGroupID(state.ID)
	if err != nil {
		return nil, err
	}
	resp, err := client.Storage.SyncGroupsClient.Get(ctx, id.ResourceGroup, id.StorageSyncServiceName, id.SyncGroupName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return utils.Bool(false), nil
		}
		return nil, fmt.Errorf("retrieving Storage Sync Group %q (Service %q / Resource Group %q): %+v", id.SyncGroupName, id.StorageSyncServiceName, id.ResourceGroup, err)
	}
	return utils.Bool(true), nil
}

func (r StorageSyncGroupResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-SS-%[1]d"
  location = "%s"
}

resource "azurerm_storage_sync" "test" {
  name                = "acctest-StorageSync-%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
}

resource "azurerm_storage_sync_group" "test" {
  name            = "acctest-StorageSyncGroup-%[1]d"
  storage_sync_id = azurerm_storage_sync.test.id
}
`, data.RandomInteger, data.Locations.Primary)
}

func (r StorageSyncGroupResource) requiresImport(data acceptance.TestData) string {
	template := r.basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_storage_sync_group" "import" {
  name            = azurerm_storage_sync_group.test.name
  storage_sync_id = azurerm_storage_sync.test.id
}
`, template)
}
