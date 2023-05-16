package springcloud_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/springcloud/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type SpringCloudContainerDeploymentResource struct{}

func TestAccSpringCloudContainerDeployment_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_spring_cloud_container_deployment", "test")
	r := SpringCloudContainerDeploymentResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSpringCloudContainerDeployment_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_spring_cloud_container_deployment", "test")
	r := SpringCloudContainerDeploymentResource{}

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

func TestAccSpringCloudContainerDeployment_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_spring_cloud_container_deployment", "test")
	r := SpringCloudContainerDeploymentResource{}

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

func TestAccSpringCloudContainerDeployment_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_spring_cloud_container_deployment", "test")
	r := SpringCloudContainerDeploymentResource{}

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
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (r SpringCloudContainerDeploymentResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.SpringCloudDeploymentID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.AppPlatform.DeploymentsClient.Get(ctx, id.ResourceGroup, id.SpringName, id.AppName, id.DeploymentName)
	if err != nil {
		return nil, fmt.Errorf("reading Spring Cloud Deployment %q (Spring Cloud Service %q / App %q / Resource Group %q): %+v", id.DeploymentName, id.SpringName, id.AppName, id.ResourceGroup, err)
	}

	return utils.Bool(resp.Properties != nil), nil
}

func (r SpringCloudContainerDeploymentResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_spring_cloud_container_deployment" "test" {
  name                = "acctest-scjd%s"
  spring_cloud_app_id = azurerm_spring_cloud_app.test.id
  server              = "docker.io"
  image               = "springio/gs-spring-boot-docker"
}
`, r.template(data), data.RandomString)
}

func (r SpringCloudContainerDeploymentResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_spring_cloud_container_deployment" "import" {
  name                = azurerm_spring_cloud_container_deployment.test.name
  spring_cloud_app_id = azurerm_spring_cloud_container_deployment.test.spring_cloud_app_id
  server              = "docker.io"
  image               = "springio/gs-spring-boot-docker"
}
`, r.basic(data))
}

func (r SpringCloudContainerDeploymentResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_spring_cloud_container_deployment" "test" {
  name                = "acctest-scjd%s"
  spring_cloud_app_id = azurerm_spring_cloud_app.test.id
  instance_count      = 2
  arguments           = ["-c", "echo hello"]
  commands            = ["/bin/sh"]
  environment_variables = {
    "Foo" : "Bar"
    "Env" : "Staging"
  }
  server             = "docker.io"
  image              = "springio/gs-spring-boot-docker"
  language_framework = "springboot"
}
`, r.template(data), data.RandomString)
}

func (SpringCloudContainerDeploymentResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-spring-%d"
  location = "%s"
}

resource "azurerm_spring_cloud_service" "test" {
  name                = "acctest-sc-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku_name            = "E0"
}

resource "azurerm_spring_cloud_app" "test" {
  name                = "acctest-sca-%d"
  resource_group_name = azurerm_spring_cloud_service.test.resource_group_name
  service_name        = azurerm_spring_cloud_service.test.name
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}
