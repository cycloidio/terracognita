package databoxedge_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/databoxedge/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type DataboxEdgeOrderResource struct{}

func TestAccDataboxEdgeOrder_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_databox_edge_order", "test")
	r := DataboxEdgeOrderResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	},
	)
}

func TestAccDataboxEdgeOrder_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_databox_edge_order", "test")
	r := DataboxEdgeOrderResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_databox_edge_order"),
		},
	})
}

func TestAccDataboxEdgeOrder_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_databox_edge_order", "test")
	r := DataboxEdgeOrderResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	},
	)
}

func (DataboxEdgeOrderResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.OrderID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.DataboxEdge.OrderClient.Get(ctx, id.DataBoxEdgeDeviceName, id.ResourceGroup)
	if err != nil {
		return nil, fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	return utils.Bool(resp.OrderProperties != nil), nil
}

// Location has to be hard coded due to limited support of locations for this resource
func (DataboxEdgeOrderResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-databoxedge-%d"
  location = "%s"
}

resource "azurerm_databox_edge_device" "test" {
  name                = "acctest-dd-%s"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location

  sku_name = "EdgeP_Base-Standard"
}
`, data.RandomInteger, "eastus", data.RandomString)
}

func (r DataboxEdgeOrderResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_databox_edge_order" "test" {
  resource_group_name = azurerm_resource_group.test.name
  device_name         = azurerm_databox_edge_device.test.name

  contact {
    name         = "TerraForm Test"
    emails       = ["creator4983@FlynnsArcade.com"]
    company_name = "Microsoft"
    phone_number = "425-882-8080"
  }

  shipment_address {
    address     = ["One Microsoft Way"]
    city        = "Redmond"
    postal_code = "98052"
    state       = "WA"
    country     = "United States"
  }
}
`, r.template(data))
}

func (r DataboxEdgeOrderResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_databox_edge_order" "import" {
  resource_group_name = azurerm_databox_edge_order.test.resource_group_name
  device_name         = azurerm_databox_edge_device.test.name

  contact {
    name         = "TerraForm Test"
    emails       = ["creator4983@FlynnsArcade.com"]
    company_name = "Microsoft"
    phone_number = "425-882-8080"
  }

  shipment_address {
    address     = ["One Microsoft Way"]
    city        = "Redmond"
    postal_code = "98052"
    state       = "WA"
    country     = "United States"
  }
}
`, r.basic(data))
}

func (r DataboxEdgeOrderResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_databox_edge_order" "test" {
  resource_group_name = azurerm_resource_group.test.name
  device_name         = azurerm_databox_edge_device.test.name

  contact {
    name         = "TerraForm Test"
    emails       = ["creator4983@FlynnsArcade.com"]
    company_name = "Flynn's Arcade"
    phone_number = "(800) 555-1234"
  }

  shipment_address {
    address     = ["One Microsoft Way"]
    city        = "Redmond"
    postal_code = "98052"
    state       = "WA"
    country     = "United States"
  }
}
`, r.template(data))
}
