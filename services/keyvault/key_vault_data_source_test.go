package keyvault_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
)

type KeyVaultDataSource struct{}

func TestAccDataSourceKeyVault_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_key_vault", "test")
	r := KeyVaultDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("tenant_id").Exists(),
				check.That(data.ResourceName).Key("sku_name").Exists(),
				check.That(data.ResourceName).Key("access_policy.0.tenant_id").Exists(),
				check.That(data.ResourceName).Key("access_policy.0.object_id").Exists(),
				check.That(data.ResourceName).Key("access_policy.0.key_permissions.0").HasValue("Create"),
				check.That(data.ResourceName).Key("access_policy.0.secret_permissions.0").HasValue("Set"),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
	})
}

func TestAccDataSourceKeyVault_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_key_vault", "test")
	r := KeyVaultDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("tenant_id").Exists(),
				check.That(data.ResourceName).Key("sku_name").Exists(),
				check.That(data.ResourceName).Key("access_policy.0.tenant_id").Exists(),
				check.That(data.ResourceName).Key("access_policy.0.object_id").Exists(),
				check.That(data.ResourceName).Key("access_policy.0.key_permissions.0").HasValue("Get"),
				check.That(data.ResourceName).Key("access_policy.0.secret_permissions.0").HasValue("Get"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.environment").HasValue("Production"),
			),
		},
	})
}

func TestAccDataSourceKeyVault_networkAcls(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_key_vault", "test")
	r := KeyVaultDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.networkAcls(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("tenant_id").Exists(),
				check.That(data.ResourceName).Key("sku_name").Exists(),
				check.That(data.ResourceName).Key("access_policy.0.tenant_id").Exists(),
				check.That(data.ResourceName).Key("access_policy.0.object_id").Exists(),
				check.That(data.ResourceName).Key("access_policy.0.key_permissions.0").HasValue("Create"),
				check.That(data.ResourceName).Key("access_policy.0.secret_permissions.0").HasValue("Set"),
				check.That(data.ResourceName).Key("network_acls.#").HasValue("1"),
				check.That(data.ResourceName).Key("network_acls.0.default_action").HasValue("Allow"),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
	})
}

func TestAccDataSourceKeyVault_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_key_vault", "test")
	r := KeyVaultDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("tenant_id").Exists(),
				check.That(data.ResourceName).Key("sku_name").Exists(),
				check.That(data.ResourceName).Key("enable_rbac_authorization").HasValue("false"),
			),
		},
		{
			Config: r.update(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("enable_rbac_authorization").HasValue("true"),
			),
		},
	})
}

func (KeyVaultDataSource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurerm_key_vault" "test" {
  name                = azurerm_key_vault.test.name
  resource_group_name = azurerm_key_vault.test.resource_group_name
}
`, KeyVaultResource{}.basic(data))
}

func (KeyVaultDataSource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurerm_key_vault" "test" {
  name                = azurerm_key_vault.test.name
  resource_group_name = azurerm_key_vault.test.resource_group_name
}
`, KeyVaultResource{}.complete(data))
}

func (KeyVaultDataSource) update(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurerm_key_vault" "test" {
  name                = azurerm_key_vault.test.name
  resource_group_name = azurerm_key_vault.test.resource_group_name
}
`, KeyVaultResource{}.update(data))
}

func (KeyVaultDataSource) networkAcls(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurerm_key_vault" "test" {
  name                = azurerm_key_vault.test.name
  resource_group_name = azurerm_key_vault.test.resource_group_name
}
`, KeyVaultResource{}.networkAclsUpdated(data))
}

func (KeyVaultDataSource) enableSoftDelete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurerm_key_vault" "test" {
  name                = azurerm_key_vault.test.name
  resource_group_name = azurerm_key_vault.test.resource_group_name
}
`, KeyVaultResource{}.softDelete(data))
}
