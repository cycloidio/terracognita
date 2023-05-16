package web_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
)

type AppServiceCertificateDataSource struct{}

func TestAccDataSourceAppServiceCertificate_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_app_service_certificate", "test")

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: AppServiceCertificateDataSource{}.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("id").Exists(),
				check.That(data.ResourceName).Key("subject_name").Exists(),
				check.That(data.ResourceName).Key("issue_date").Exists(),
				check.That(data.ResourceName).Key("expiration_date").Exists(),
				check.That(data.ResourceName).Key("thumbprint").Exists(),
			),
		},
	})
}

func (d AppServiceCertificateDataSource) basic(data acceptance.TestData) string {
	template := AppServiceCertificateResource{}.pfxNoPassword(data)
	return fmt.Sprintf(`
%s

data "azurerm_app_service_certificate" "test" {
  name                = azurerm_app_service_certificate.test.name
  resource_group_name = azurerm_app_service_certificate.test.resource_group_name
}
`, template)
}
