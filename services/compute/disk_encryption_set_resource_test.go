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

type DiskEncryptionSetResource struct{}

func TestAccDiskEncryptionSet_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_disk_encryption_set", "test")
	r := DiskEncryptionSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("encryption_type").HasValue("EncryptionAtRestWithCustomerKey"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccDiskEncryptionSet_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_disk_encryption_set", "test")
	r := DiskEncryptionSetResource{}

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

func TestAccDiskEncryptionSet_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_disk_encryption_set", "test")
	r := DiskEncryptionSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccDiskEncryptionSet_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_disk_encryption_set", "test")
	r := DiskEncryptionSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccDiskEncryptionSet_keyRotate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_disk_encryption_set", "test")
	r := DiskEncryptionSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config: r.keyRotate(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccDiskEncryptionSet_withEncryptionType(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_disk_encryption_set", "test")
	r := DiskEncryptionSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.withPlatformAndCustomerKeys(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (DiskEncryptionSetResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.DiskEncryptionSetID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Compute.DiskEncryptionSetsClient.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving Compute Disk Encryption Set %q", id.String())
	}

	return utils.Bool(resp.ID != nil), nil
}

func (DiskEncryptionSetResource) dependencies(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {
    key_vault {
      recover_soft_deleted_key_vaults    = false
      purge_soft_delete_on_destroy       = false
      purge_soft_deleted_keys_on_destroy = false
    }
  }
}

data "azurerm_client_config" "current" {}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_key_vault" "test" {
  name                        = "acctestkv-%s"
  location                    = azurerm_resource_group.test.location
  resource_group_name         = azurerm_resource_group.test.name
  tenant_id                   = data.azurerm_client_config.current.tenant_id
  sku_name                    = "standard"
  purge_protection_enabled    = true
  enabled_for_disk_encryption = true
}

resource "azurerm_key_vault_access_policy" "service-principal" {
  key_vault_id = azurerm_key_vault.test.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = data.azurerm_client_config.current.object_id

  key_permissions = [
    "Create",
    "Delete",
    "Get",
    "Purge",
    "Update",
  ]

  secret_permissions = [
    "Get",
    "Delete",
    "Purge",
    "Set",
  ]
}

resource "azurerm_key_vault_key" "test" {
  name         = "examplekey"
  key_vault_id = azurerm_key_vault.test.id
  key_type     = "RSA"
  key_size     = 2048

  key_opts = [
    "decrypt",
    "encrypt",
    "sign",
    "unwrapKey",
    "verify",
    "wrapKey",
  ]

  depends_on = ["azurerm_key_vault_access_policy.service-principal"]
}

resource "azurerm_key_vault_access_policy" "disk-encryption" {
  key_vault_id = azurerm_key_vault.test.id

  key_permissions = [
    "Get",
    "WrapKey",
    "UnwrapKey",
  ]

  tenant_id = azurerm_disk_encryption_set.test.identity.0.tenant_id
  object_id = azurerm_disk_encryption_set.test.identity.0.principal_id
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString)
}

func (r DiskEncryptionSetResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_disk_encryption_set" "test" {
  name                = "acctestDES-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  key_vault_key_id    = azurerm_key_vault_key.test.id

  identity {
    type = "SystemAssigned"
  }
}
`, r.dependencies(data), data.RandomInteger)
}

func (r DiskEncryptionSetResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_disk_encryption_set" "import" {
  name                = azurerm_disk_encryption_set.test.name
  resource_group_name = azurerm_disk_encryption_set.test.resource_group_name
  location            = azurerm_disk_encryption_set.test.location
  key_vault_key_id    = azurerm_disk_encryption_set.test.key_vault_key_id

  identity {
    type = "SystemAssigned"
  }
}
`, r.basic(data))
}

func (r DiskEncryptionSetResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_disk_encryption_set" "test" {
  name                      = "acctestDES-%d"
  resource_group_name       = azurerm_resource_group.test.name
  location                  = azurerm_resource_group.test.location
  key_vault_key_id          = azurerm_key_vault_key.test.id
  auto_key_rotation_enabled = true

  identity {
    type = "SystemAssigned"
  }

  tags = {
    Hello = "woRld"
  }
}
`, r.dependencies(data), data.RandomInteger)
}

func (r DiskEncryptionSetResource) keyRotate(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_key_vault_key" "new" {
  name         = "newKey"
  key_vault_id = azurerm_key_vault.test.id
  key_type     = "RSA"
  key_size     = 2048

  key_opts = [
    "decrypt",
    "encrypt",
    "sign",
    "unwrapKey",
    "verify",
    "wrapKey",
  ]

  depends_on = ["azurerm_key_vault_access_policy.service-principal"]
}

resource "azurerm_disk_encryption_set" "test" {
  name                      = "acctestDES-%d"
  resource_group_name       = azurerm_resource_group.test.name
  location                  = azurerm_resource_group.test.location
  key_vault_key_id          = azurerm_key_vault_key.new.id
  auto_key_rotation_enabled = true

  identity {
    type = "SystemAssigned"
  }
}
`, r.dependencies(data), data.RandomInteger)
}

func (r DiskEncryptionSetResource) withPlatformAndCustomerKeys(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_disk_encryption_set" "test" {
  name                = "acctestDES-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  key_vault_key_id    = azurerm_key_vault_key.test.id
  encryption_type     = "EncryptionAtRestWithPlatformAndCustomerKeys"

  identity {
    type = "SystemAssigned"
  }
}
`, r.dependencies(data), data.RandomInteger)
}
