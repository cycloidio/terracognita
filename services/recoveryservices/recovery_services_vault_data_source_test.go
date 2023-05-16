package recoveryservices_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
)

type RecoveryServicesVaultDataSource struct{}

func TestAccDataSourceAzureRMRecoveryServicesVault_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_recovery_services_vault", "test")
	r := RecoveryServicesVaultDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("name").Exists(),
				check.That(data.ResourceName).Key("location").Exists(),
				check.That(data.ResourceName).Key("resource_group_name").Exists(),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
				check.That(data.ResourceName).Key("sku").HasValue("Standard"),
			),
		},
	})
}

func (RecoveryServicesVaultDataSource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurerm_recovery_services_vault" "test" {
  name                = azurerm_recovery_services_vault.test.name
  resource_group_name = azurerm_resource_group.test.name
}
`, RecoveryServicesVaultResource{}.basic(data))
}
