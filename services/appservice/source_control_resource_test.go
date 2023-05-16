package appservice_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/appservice/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

// (@jackofallops) Note: Some tests require a valid GitHub token for ARM_GITHUB_ACCESS_TOKEN. This token needs the `repo` and `workflow` permissions to be applicable on the referenced repositories.

type AppServiceSourceControlResource struct{}

func TestAccSourceControlResource_windowsExternalGit(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_app_service_source_control", "test")
	r := AppServiceSourceControlResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.windowsExternalGit(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSourceControlResource_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_app_service_source_control", "test")
	r := AppServiceSourceControlResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.windowsExternalGit(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func TestAccSourceControlResource_windowsLocalGit(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_app_service_source_control", "test")
	r := AppServiceSourceControlResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.windowsLocalGit(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("scm_type").HasValue("LocalGit"),
				check.That(data.ResourceName).Key("repo_url").HasValue(fmt.Sprintf("https://acctestwa-%d.scm.azurewebsites.net", data.RandomInteger)),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSourceControlResource_windowsGitHubAction(t *testing.T) {
	if ok := os.Getenv("ARM_GITHUB_ACCESS_TOKEN"); ok == "" {
		t.Skip("Skipping as `ARM_GITHUB_ACCESS_TOKEN` is not specified")
	}

	data := acceptance.BuildTestData(t, "azurerm_app_service_source_control", "test")
	r := AppServiceSourceControlResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.windowsGitHubAction(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSourceControlResource_windowsGitHub(t *testing.T) {
	if ok := os.Getenv("ARM_GITHUB_ACCESS_TOKEN"); ok == "" {
		t.Skip("Skipping as `ARM_GITHUB_ACCESS_TOKEN` is not specified")
	}

	data := acceptance.BuildTestData(t, "azurerm_app_service_source_control", "test")
	r := AppServiceSourceControlResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.windowsGitHub(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSourceControlResource_linuxExternalGit(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_app_service_source_control", "test")
	r := AppServiceSourceControlResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.linuxExternalGit(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSourceControlResource_linuxLocalGit(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_app_service_source_control", "test")
	r := AppServiceSourceControlResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.linuxLocalGit(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("scm_type").HasValue("LocalGit"),
				check.That(data.ResourceName).Key("repo_url").HasValue(fmt.Sprintf("https://acctestwa-%d.scm.azurewebsites.net", data.RandomInteger)),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSourceControlResource_linuxGitHubAction(t *testing.T) {
	if ok := os.Getenv("ARM_GITHUB_ACCESS_TOKEN"); ok == "" {
		t.Skip("Skipping as `ARM_GITHUB_ACCESS_TOKEN` is not specified")
	}

	data := acceptance.BuildTestData(t, "azurerm_app_service_source_control", "test")
	r := AppServiceSourceControlResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.linuxGitHubAction(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("uses_github_action").HasValue("true"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSourceControlResource_linuxGitHub(t *testing.T) {
	if ok := os.Getenv("ARM_GITHUB_ACCESS_TOKEN"); ok == "" {
		t.Skip("Skipping as `ARM_GITHUB_ACCESS_TOKEN` is not specified")
	}

	data := acceptance.BuildTestData(t, "azurerm_app_service_source_control", "test")
	r := AppServiceSourceControlResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.linuxGitHub(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (r AppServiceSourceControlResource) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.WebAppID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := client.AppService.WebAppsClient.GetSourceControl(ctx, id.ResourceGroup, id.SiteName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return utils.Bool(false), nil
		}
		return nil, fmt.Errorf("retrieving Source Control for %s: %v", id, err)
	}
	if resp.SiteSourceControlProperties == nil || resp.SiteSourceControlProperties.RepoURL == nil {
		return utils.Bool(false), nil
	}

	return utils.Bool(true), nil
}

func (r AppServiceSourceControlResource) windowsExternalGit(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

%s

resource "azurerm_app_service_source_control" "test" {
  app_id                 = azurerm_windows_web_app.test.id
  repo_url               = "https://github.com/Azure-Samples/app-service-web-dotnet-get-started.git"
  branch                 = "master"
  use_manual_integration = true
}
`, r.baseWindowsAppTemplate(data))
}

func (r AppServiceSourceControlResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_app_service_source_control" "import" {
  app_id                 = azurerm_app_service_source_control.test.app_id
  repo_url               = azurerm_app_service_source_control.test.repo_url
  branch                 = azurerm_app_service_source_control.test.branch
  use_manual_integration = azurerm_app_service_source_control.test.use_manual_integration
}
`, r.windowsExternalGit(data))
}

func (r AppServiceSourceControlResource) linuxExternalGit(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

%s

resource "azurerm_app_service_source_control" "test" {
  app_id                 = azurerm_linux_web_app.test.id
  repo_url               = "https://github.com/Azure-Samples/python-docs-hello-world.git"
  branch                 = "master"
  use_manual_integration = true
}
`, r.baseLinuxAppTemplate(data))
}

func (r AppServiceSourceControlResource) windowsLocalGit(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

%s

resource "azurerm_app_service_source_control" "test" {
  app_id        = azurerm_windows_web_app.test.id
  use_local_git = true
}
`, r.baseWindowsAppTemplate(data))
}

func (r AppServiceSourceControlResource) linuxLocalGit(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

%s

resource "azurerm_app_service_source_control" "test" {
  app_id        = azurerm_linux_web_app.test.id
  use_local_git = true
}
`, r.baseLinuxAppTemplate(data))
}

func (r AppServiceSourceControlResource) windowsGitHubAction(data acceptance.TestData) string {
	token := os.Getenv("ARM_GITHUB_ACCESS_TOKEN")
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

%s

resource azurerm_source_control_token test {
  type  = "GitHub"
  token = "%s"
}

resource "azurerm_app_service_source_control" "test" {
  app_id   = azurerm_windows_web_app.test.id
  repo_url = "https://github.com/jackofallops/app-service-web-dotnet-get-started.git"
  branch   = "master"

  github_action_configuration {
    generate_workflow_file = true

    code_configuration {
      runtime_stack   = "dotnetcore"
      runtime_version = "5.0.x"
    }
  }
}
`, r.baseWindowsAppTemplate(data), token)
}

func (r AppServiceSourceControlResource) windowsGitHub(data acceptance.TestData) string {
	token := os.Getenv("ARM_GITHUB_ACCESS_TOKEN")
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

%s

resource azurerm_source_control_token test {
  type  = "GitHub"
  token = "%s"
}

resource "azurerm_app_service_source_control" "test" {
  app_id   = azurerm_windows_web_app.test.id
  repo_url = "https://github.com/jackofallops/azure-app-service-static-site-tests.git"
  branch   = "main"

  depends_on = [
    azurerm_source_control_token.test,
  ]
}
`, r.baseWindowsAppTemplate(data), token)
}

func (r AppServiceSourceControlResource) linuxGitHubAction(data acceptance.TestData) string {
	token := os.Getenv("ARM_GITHUB_ACCESS_TOKEN")
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

%s

resource azurerm_source_control_token test {
  type  = "GitHub"
  token = "%s"
}

resource "azurerm_app_service_source_control" "test" {
  app_id   = azurerm_linux_web_app.test.id
  repo_url = "https://github.com/jackofallops/app-service-web-dotnet-get-started.git"
  branch   = "master"

  github_action_configuration {
    generate_workflow_file = true

    code_configuration {
      runtime_stack   = "dotnetcore"
      runtime_version = "5.0.x"
    }
  }
}
`, r.baseLinuxAppTemplate(data), token)
}

func (r AppServiceSourceControlResource) linuxGitHub(data acceptance.TestData) string {
	token := os.Getenv("ARM_GITHUB_ACCESS_TOKEN")
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

%s

resource "azurerm_source_control_token" "test" {
  type  = "GitHub"
  token = "%s"
}

resource "azurerm_app_service_source_control" "test" {
  app_id                 = azurerm_linux_web_app.test.id
  repo_url               = "https://github.com/Azure-Samples/python-docs-hello-world.git"
  branch                 = "master"
  use_manual_integration = true

  depends_on = [
    azurerm_source_control_token.test,
  ]
}
`, r.baseLinuxAppTemplate(data), token)
}

func (r AppServiceSourceControlResource) baseWindowsAppTemplate(data acceptance.TestData) string {
	return fmt.Sprintf(`

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-ASSC-%[1]d"
  location = "%s"
}

resource "azurerm_service_plan" "test" {
  name                = "acctest-SP-%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku_name            = "S1"
  os_type             = "Windows"
}

resource "azurerm_windows_web_app" "test" {
  name                = "acctestWA-%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  service_plan_id     = azurerm_service_plan.test.id

  site_config {}
}
`, data.RandomInteger, data.Locations.Primary)
}

func (r AppServiceSourceControlResource) baseLinuxAppTemplate(data acceptance.TestData) string {
	return fmt.Sprintf(`

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-ASSC-%[1]d"
  location = "%[2]s"
}

resource "azurerm_service_plan" "test" {
  name                = "acctest-SP-%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku_name            = "B1"
  os_type             = "Linux"
}

resource "azurerm_linux_web_app" "test" {
  name                = "acctestWA-%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  service_plan_id     = azurerm_service_plan.test.id

  site_config {
    application_stack {
      python_version = "3.9"
    }
  }
}
`, data.RandomInteger, data.Locations.Primary)
}
