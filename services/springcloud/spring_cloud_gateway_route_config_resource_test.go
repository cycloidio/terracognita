package springcloud_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/springcloud/parse"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type SpringCloudGatewayRouteConfigResource struct{}

func TestAccSpringCloudGatewayRouteConfig_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_spring_cloud_gateway_route_config", "test")
	r := SpringCloudGatewayRouteConfigResource{}
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

func TestAccSpringCloudGatewayRouteConfig_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_spring_cloud_gateway_route_config", "test")
	r := SpringCloudGatewayRouteConfigResource{}
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

func TestAccSpringCloudGatewayRouteConfig_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_spring_cloud_gateway_route_config", "test")
	r := SpringCloudGatewayRouteConfigResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSpringCloudGatewayRouteConfig_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_spring_cloud_gateway_route_config", "test")
	r := SpringCloudGatewayRouteConfigResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (r SpringCloudGatewayRouteConfigResource) Exists(ctx context.Context, client *clients.Client, state *terraform.InstanceState) (*bool, error) {
	id, err := parse.SpringCloudGatewayRouteConfigID(state.ID)
	if err != nil {
		return nil, err
	}
	resp, err := client.AppPlatform.GatewayRouteConfigClient.Get(ctx, id.ResourceGroup, id.SpringName, id.GatewayName, id.RouteConfigName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return utils.Bool(false), nil
		}
		return nil, fmt.Errorf("retrieving %s: %+v", id, err)
	}
	return utils.Bool(true), nil
}

func (r SpringCloudGatewayRouteConfigResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-spring-%[2]d"
  location = "%[1]s"
}

resource "azurerm_spring_cloud_service" "test" {
  name                = "acctest-sc-%[2]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku_name            = "E0"
}

resource "azurerm_spring_cloud_gateway" "test" {
  name                    = "default"
  spring_cloud_service_id = azurerm_spring_cloud_service.test.id
}

resource "azurerm_spring_cloud_app" "test" {
  name                = "acctest-sca-%[2]d"
  resource_group_name = azurerm_spring_cloud_service.test.resource_group_name
  service_name        = azurerm_spring_cloud_service.test.name
}
`, data.Locations.Primary, data.RandomInteger)
}

func (r SpringCloudGatewayRouteConfigResource) basic(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_spring_cloud_gateway_route_config" "test" {
  name                    = "acctest-agrc-%d"
  spring_cloud_gateway_id = azurerm_spring_cloud_gateway.test.id
  spring_cloud_app_id     = azurerm_spring_cloud_app.test.id
}
`, template, data.RandomInteger)
}

func (r SpringCloudGatewayRouteConfigResource) requiresImport(data acceptance.TestData) string {
	config := r.basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_spring_cloud_gateway_route_config" "import" {
  name                    = azurerm_spring_cloud_gateway_route_config.test.name
  spring_cloud_gateway_id = azurerm_spring_cloud_gateway_route_config.test.spring_cloud_gateway_id
  spring_cloud_app_id     = azurerm_spring_cloud_gateway_route_config.test.spring_cloud_app_id
}
`, config)
}

func (r SpringCloudGatewayRouteConfigResource) complete(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_spring_cloud_gateway_route_config" "test" {
  name                    = "acctest-agrc-%d"
  spring_cloud_gateway_id = azurerm_spring_cloud_gateway.test.id
  spring_cloud_app_id     = azurerm_spring_cloud_app.test.id
  route {
    description            = "test description"
    filters                = ["StripPrefix=2", "RateLimit=1,1s"]
    order                  = 1
    predicates             = ["Path=/api5/customer/**"]
    sso_validation_enabled = true
    title                  = "myApp route config"
    token_relay            = true
    uri                    = "https://www.test.com"
    classification_tags    = ["tag1", "tag2"]
  }
}
`, template, data.RandomInteger)
}
