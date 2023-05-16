package logic_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/logic/parse"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type LogicAppIntegrationAccountPartnerResource struct{}

func TestAccLogicAppIntegrationAccountPartner_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_logic_app_integration_account_partner", "test")
	r := LogicAppIntegrationAccountPartnerResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLogicAppIntegrationAccountPartner_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_logic_app_integration_account_partner", "test")
	r := LogicAppIntegrationAccountPartnerResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func TestAccLogicAppIntegrationAccountPartner_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_logic_app_integration_account_partner", "test")
	r := LogicAppIntegrationAccountPartnerResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.complete(data, "DUNS", "FabrikamNY", "bar"),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLogicAppIntegrationAccountPartner_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_logic_app_integration_account_partner", "test")
	r := LogicAppIntegrationAccountPartnerResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.complete(data, "DUNS", "FabrikamNY", "bar"),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data, "AS2Identity", "FabrikamDC", "bar2"),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (r LogicAppIntegrationAccountPartnerResource) Exists(ctx context.Context, client *clients.Client, state *terraform.InstanceState) (*bool, error) {
	id, err := parse.IntegrationAccountPartnerID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := client.Logic.IntegrationAccountPartnerClient.Get(ctx, id.ResourceGroup, id.IntegrationAccountName, id.PartnerName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return utils.Bool(false), nil
		}
		return nil, fmt.Errorf("retrieving %q %+v", id, err)
	}

	return utils.Bool(resp.IntegrationAccountPartnerProperties != nil), nil
}

func (r LogicAppIntegrationAccountPartnerResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-logic-%d"
  location = "%s"
}

resource "azurerm_logic_app_integration_account" "test" {
  name                = "acctest-ia-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku_name            = "Standard"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r LogicAppIntegrationAccountPartnerResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_logic_app_integration_account_partner" "test" {
  name                     = "acctest-iap-%d"
  resource_group_name      = azurerm_resource_group.test.name
  integration_account_name = azurerm_logic_app_integration_account.test.name

  business_identity {
    qualifier = "DUNS"
    value     = "FabrikamNY"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r LogicAppIntegrationAccountPartnerResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_logic_app_integration_account_partner" "import" {
  name                     = azurerm_logic_app_integration_account_partner.test.name
  resource_group_name      = azurerm_logic_app_integration_account_partner.test.resource_group_name
  integration_account_name = azurerm_logic_app_integration_account_partner.test.integration_account_name

  business_identity {
    qualifier = "DUNS"
    value     = "FabrikamNY"
  }
}
`, r.basic(data))
}

func (r LogicAppIntegrationAccountPartnerResource) complete(data acceptance.TestData, qualifier string, value string, metdataContent string) string {
	return fmt.Sprintf(`
%s

resource "azurerm_logic_app_integration_account_partner" "test" {
  name                     = "acctest-iap-%d"
  resource_group_name      = azurerm_resource_group.test.name
  integration_account_name = azurerm_logic_app_integration_account.test.name

  business_identity {
    qualifier = "%s"
    value     = "%s"
  }

  metadata = <<METADATA
    {
        "foo": "%s"
    }
METADATA
}
`, r.template(data), data.RandomInteger, qualifier, value, metdataContent)
}
