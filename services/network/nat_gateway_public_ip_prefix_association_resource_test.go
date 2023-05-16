package network_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type NatGatewayPublicIpPrefixAssociationResource struct{}

func TestAccNatGatewayPublicIpPrefixAssociation_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_nat_gateway_public_ip_prefix_association", "test")
	r := NatGatewayPublicIpPrefixAssociationResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		// intentional as this is a Virtual Resource
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccNatGatewayPublicIpPrefixAssociation_updateNatGateway(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_nat_gateway_public_ip_prefix_association", "test")
	r := NatGatewayPublicIpPrefixAssociationResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		// intentional as this is a Virtual Resource
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.updateNatGateway(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccNatGatewayPublicIpPrefixAssociation_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_nat_gateway_public_ip_prefix_association", "test")
	r := NatGatewayPublicIpPrefixAssociationResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		// intentional as this is a Virtual Resource
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func TestAccNatGatewayPublicIpPrefixAssociation_deleted(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_nat_gateway_public_ip_prefix_association", "test")
	r := NatGatewayPublicIpPrefixAssociationResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		// intentional as this is a Virtual Resource
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       r.basic,
			TestResource: r,
		}),
	})
}

func (t NatGatewayPublicIpPrefixAssociationResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.NatGatewayPublicIPPrefixAssociationID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Network.NatGatewayClient.Get(ctx, id.NatGateway.ResourceGroup, id.NatGateway.Name, "")
	if err != nil {
		return nil, fmt.Errorf("reading Nat Gateway Public IP Prefix Association (%s): %+v", id, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (NatGatewayPublicIpPrefixAssociationResource) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.NatGatewayPublicIPPrefixAssociationID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := client.Network.NatGatewayClient.Get(ctx, id.NatGateway.ResourceGroup, id.NatGateway.Name, "")
	if err != nil {
		return nil, fmt.Errorf("reading Nat Gateway Public IP Prefix Association (%s): %+v", id, err)
	}

	updatedPrefixes := make([]network.SubResource, 0)
	if publicIpPrefixes := resp.PublicIPPrefixes; publicIpPrefixes != nil {
		for _, publicIpPrefix := range *publicIpPrefixes {
			if !strings.EqualFold(*publicIpPrefix.ID, id.PublicIPPrefixID) {
				updatedPrefixes = append(updatedPrefixes, publicIpPrefix)
			}
		}
	}
	resp.PublicIPPrefixes = &updatedPrefixes

	future, err := client.Network.NatGatewayClient.CreateOrUpdate(ctx, id.NatGateway.ResourceGroup, id.NatGateway.Name, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to remove Nat Gateway Public IP Prefix Association for Nat Gateway %q: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Network.NatGatewayClient.Client); err != nil {
		return nil, fmt.Errorf("failed to wait for removal of Nat Gateway Public IP Prefix Association for Nat Gateway %q: %+v", id, err)
	}

	return utils.Bool(true), nil
}

func (r NatGatewayPublicIpPrefixAssociationResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_nat_gateway" "test" {
  name                = "acctest-NatGateway-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku_name            = "Standard"
}

resource "azurerm_nat_gateway_public_ip_prefix_association" "test" {
  nat_gateway_id      = azurerm_nat_gateway.test.id
  public_ip_prefix_id = azurerm_public_ip_prefix.test.id
}
`, r.template(data), data.RandomInteger)
}

func (r NatGatewayPublicIpPrefixAssociationResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_nat_gateway_public_ip_prefix_association" "import" {
  nat_gateway_id      = azurerm_nat_gateway_public_ip_prefix_association.test.nat_gateway_id
  public_ip_prefix_id = azurerm_nat_gateway_public_ip_prefix_association.test.public_ip_prefix_id
}
`, r.basic(data))
}

func (r NatGatewayPublicIpPrefixAssociationResource) updateNatGateway(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_nat_gateway" "test" {
  name                = "acctest-NatGateway-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku_name            = "Standard"
  tags = {
    Hello = "World"
  }
}

resource "azurerm_nat_gateway_public_ip_prefix_association" "test" {
  nat_gateway_id      = azurerm_nat_gateway.test.id
  public_ip_prefix_id = azurerm_public_ip_prefix.test.id
}
`, r.template(data), data.RandomInteger)
}

func (NatGatewayPublicIpPrefixAssociationResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-ngpi-%d"
  location = "%s"
}

resource "azurerm_public_ip_prefix" "test" {
  name                = "acctestpublicIPPrefix-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  prefix_length       = 30
  zones               = ["1"]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
