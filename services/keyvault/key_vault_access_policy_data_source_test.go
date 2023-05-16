package keyvault_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
)

type KeyVaultAccessPolicyDataSource struct{}

func TestAccDataSourceKeyVaultAccessPolicy_key(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_key_vault_access_policy", "test")
	r := KeyVaultAccessPolicyDataSource{}
	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.testAccDataSourceKeyVaultAccessPolicy("Key Management"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("key_permissions.#").HasValue("9"),
				acceptance.TestCheckNoResourceAttr(data.ResourceName, "secret_permissions"),
				acceptance.TestCheckNoResourceAttr(data.ResourceName, "certificate_permissions"),
			),
		},
	})
}

func TestAccDataSourceKeyVaultAccessPolicy_secret(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_key_vault_access_policy", "test")
	r := KeyVaultAccessPolicyDataSource{}
	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.testAccDataSourceKeyVaultAccessPolicy("Secret Management"),
			Check: acceptance.ComposeTestCheckFunc(
				acceptance.TestCheckNoResourceAttr(data.ResourceName, "key_permissions"),
				check.That(data.ResourceName).Key("secret_permissions.#").HasValue("7"),
				acceptance.TestCheckNoResourceAttr(data.ResourceName, "certificate_permissions"),
			),
		},
	})
}

func TestAccDataSourceKeyVaultAccessPolicy_certificate(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_key_vault_access_policy", "test")
	r := KeyVaultAccessPolicyDataSource{}
	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.testAccDataSourceKeyVaultAccessPolicy("Certificate Management"),
			Check: acceptance.ComposeTestCheckFunc(
				acceptance.TestCheckNoResourceAttr(data.ResourceName, "key_permissions"),
				acceptance.TestCheckNoResourceAttr(data.ResourceName, "secret_permissions"),
				check.That(data.ResourceName).Key("certificate_permissions.#").HasValue("12"),
			),
		},
	})
}

func TestAccDataSourceKeyVaultAccessPolicy_keySecret(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_key_vault_access_policy", "test")
	r := KeyVaultAccessPolicyDataSource{}
	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.testAccDataSourceKeyVaultAccessPolicy("Key & Secret Management"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("key_permissions.#").HasValue("9"),
				check.That(data.ResourceName).Key("secret_permissions.#").HasValue("7"),
				acceptance.TestCheckNoResourceAttr(data.ResourceName, "certificate_permissions"),
			),
		},
	})
}

func TestAccDataSourceKeyVaultAccessPolicy_keyCertificate(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_key_vault_access_policy", "test")
	r := KeyVaultAccessPolicyDataSource{}
	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.testAccDataSourceKeyVaultAccessPolicy("Key & Certificate Management"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("key_permissions.#").HasValue("9"),
				acceptance.TestCheckNoResourceAttr(data.ResourceName, "secret_permissions"),
				check.That(data.ResourceName).Key("certificate_permissions.#").HasValue("12"),
			),
		},
	})
}

func TestAccDataSourceKeyVaultAccessPolicy_secretCertificate(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_key_vault_access_policy", "test")
	r := KeyVaultAccessPolicyDataSource{}
	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.testAccDataSourceKeyVaultAccessPolicy("Secret & Certificate Management"),
			Check: acceptance.ComposeTestCheckFunc(
				acceptance.TestCheckNoResourceAttr(data.ResourceName, "key_permissions"),
				check.That(data.ResourceName).Key("secret_permissions.#").HasValue("7"),
				check.That(data.ResourceName).Key("certificate_permissions.#").HasValue("12"),
			),
		},
	})
}

func TestAccDataSourceKeyVaultAccessPolicy_keySecretCertificate(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_key_vault_access_policy", "test")
	r := KeyVaultAccessPolicyDataSource{}
	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.testAccDataSourceKeyVaultAccessPolicy("Key, Secret, & Certificate Management"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("key_permissions.#").HasValue("9"),
				check.That(data.ResourceName).Key("secret_permissions.#").HasValue("7"),
				check.That(data.ResourceName).Key("certificate_permissions.#").HasValue("12"),
			),
		},
	})
}

func (r KeyVaultAccessPolicyDataSource) testAccDataSourceKeyVaultAccessPolicy(name string) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

data "azurerm_key_vault_access_policy" "test" {
  name = "%s"
}
`, name)
}
