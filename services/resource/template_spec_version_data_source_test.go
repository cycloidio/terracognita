package resource_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
)

type TemplateSpecVersionDataSource struct{}

func TestAccDataSourceTemplateSpecVersion(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_template_spec_version", "test")
	r := TemplateSpecVersionDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.basic(),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("name").HasValue("acctest-standing-data-empty"),
				check.That(data.ResourceName).Key("version").HasValue("v1.0.0"),
				check.That(data.ResourceName).Key("template_body").Exists(),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
			),
		},
	})
}

func (TemplateSpecVersionDataSource) basic() string {
	return `
provider "azurerm" {
  features {}
}

data "azurerm_template_spec_version" "test" {
  name                = "acctest-standing-data-empty"
  resource_group_name = "standing-data-for-acctest"
  version             = "v1.0.0"
}
`
}
