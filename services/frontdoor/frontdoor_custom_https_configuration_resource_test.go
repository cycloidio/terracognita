package frontdoor_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/frontdoor/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/frontdoor/sdk/2020-05-01/frontdoors"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type FrontDoorCustomHttpsConfigurationResource struct{}

func TestAccFrontDoorCustomHttpsConfiguration_CustomHttps(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_frontdoor_custom_https_configuration", "test")
	r := FrontDoorCustomHttpsConfigurationResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.Enabled(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("custom_https_provisioning_enabled").HasValue("true"),
				check.That(data.ResourceName).Key("custom_https_configuration.0.certificate_source").HasValue("FrontDoor"),
			),
		},
		{
			Config: r.Disabled(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("custom_https_provisioning_enabled").HasValue("false"),
			),
		},
	})
}

func TestAccFrontDoorCustomHttpsConfiguration_DisabledWithConfigurationBlock(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_frontdoor_custom_https_configuration", "test")
	r := FrontDoorCustomHttpsConfigurationResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config:      r.DisabledWithConfigurationBlock(data),
			ExpectError: regexp.MustCompile(`"custom_https_provisioning_enabled" is set to "false". please remove the "custom_https_configuration" block from the configuration file`),
		},
	})
}

func TestAccFrontDoorCustomHttpsConfiguration_EnabledWithoutConfigurationBlock(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_frontdoor_custom_https_configuration", "test")
	r := FrontDoorCustomHttpsConfigurationResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config:      r.EnabledWithoutConfigurationBlock(data),
			ExpectError: regexp.MustCompile(`"custom_https_provisioning_enabled" is set to "true". please add a "custom_https_configuration" block to the configuration file`),
		},
	})
}

func TestAccFrontDoorCustomHttpsConfiguration_EnabledFrontdoorExtraAttributes(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_frontdoor_custom_https_configuration", "test")
	r := FrontDoorCustomHttpsConfigurationResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config:      r.EnabledFrontdoorExtraAttributes(data),
			ExpectError: regexp.MustCompile(`a Front Door managed "custom_https_configuration" block does not support the following keys.`),
		},
	})
}

func TestAccFrontDoorCustomHttpsConfiguration_EnabledKeyVaultLatest(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_frontdoor_custom_https_configuration", "test")
	r := FrontDoorCustomHttpsConfigurationResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config:      r.EnabledKeyVaultLatest(data),
			ExpectError: regexp.MustCompile(`"azure_key_vault_certificate_secret_version" can not be set to "latest" please remove this attribute from the configuration file.`),
		},
	})
}

func TestAccFrontDoorCustomHttpsConfiguration_EnabledKeyVaultLatestMissingAttributes(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_frontdoor_custom_https_configuration", "test")
	r := FrontDoorCustomHttpsConfigurationResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config:      r.EnabledKeyVaultLatestMissingAttributes(data),
			ExpectError: regexp.MustCompile(`a "AzureKeyVault" managed "custom_https_configuration" block must have values in the following fileds: "azure_key_vault_certificate_secret_name" and "azure_key_vault_certificate_vault_id"`),
		},
	})
}

func TestAccFrontDoorCustomHttpsConfiguration_EnabledKeyVaultMissingAttributes(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_frontdoor_custom_https_configuration", "test")
	r := FrontDoorCustomHttpsConfigurationResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config:      r.EnabledKeyVaultMissingAttributes(data),
			ExpectError: regexp.MustCompile(`a "AzureKeyVault" managed "custom_https_configuration" block must have values in the following fileds: "azure_key_vault_certificate_secret_name", "azure_key_vault_certificate_secret_version", and "azure_key_vault_certificate_vault_id"`),
		},
	})
}

func (FrontDoorCustomHttpsConfigurationResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.CustomHttpsConfigurationIDInsensitively(state.ID)
	if err != nil {
		return nil, err
	}

	frontdoorId, err := frontdoors.ParseFrontendEndpointID(state.Attributes["frontend_endpoint_id"])
	if err != nil {
		return nil, err
	}

	resp, err := clients.Frontdoor.FrontDoorsClient.FrontendEndpointsGet(ctx, *frontdoorId)
	if err != nil {
		return nil, fmt.Errorf("retrieving %s: %v", *id, err)
	}

	return utils.Bool(resp.Model != nil), nil
}

func (r FrontDoorCustomHttpsConfigurationResource) Enabled(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_frontdoor_custom_https_configuration" "test" {
  frontend_endpoint_id              = azurerm_frontdoor.test.frontend_endpoints[local.endpoint_name]
  custom_https_provisioning_enabled = true

  custom_https_configuration {
    certificate_source = "FrontDoor"
  }
}
`, r.template(data))
}

func (r FrontDoorCustomHttpsConfigurationResource) Disabled(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_frontdoor_custom_https_configuration" "test" {
  frontend_endpoint_id              = azurerm_frontdoor.test.frontend_endpoints[local.endpoint_name]
  custom_https_provisioning_enabled = false
}
`, r.template(data))
}

func (r FrontDoorCustomHttpsConfigurationResource) DisabledWithConfigurationBlock(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_frontdoor_custom_https_configuration" "test" {
  frontend_endpoint_id              = azurerm_frontdoor.test.frontend_endpoints[local.endpoint_name]
  custom_https_provisioning_enabled = false

  custom_https_configuration {
    certificate_source = "FrontDoor"
  }
}
`, r.template(data))
}

func (r FrontDoorCustomHttpsConfigurationResource) EnabledWithoutConfigurationBlock(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_frontdoor_custom_https_configuration" "test" {
  frontend_endpoint_id              = azurerm_frontdoor.test.frontend_endpoints[local.endpoint_name]
  custom_https_provisioning_enabled = true
}
`, r.template(data))
}

func (r FrontDoorCustomHttpsConfigurationResource) EnabledFrontdoorExtraAttributes(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_frontdoor_custom_https_configuration" "test" {
  frontend_endpoint_id              = azurerm_frontdoor.test.frontend_endpoints[local.endpoint_name]
  custom_https_provisioning_enabled = true

  custom_https_configuration {
    certificate_source                      = "FrontDoor"
    azure_key_vault_certificate_secret_name = "accTest"
  }
}
`, r.template(data))
}

func (r FrontDoorCustomHttpsConfigurationResource) EnabledKeyVaultLatest(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_frontdoor_custom_https_configuration" "test" {
  frontend_endpoint_id              = azurerm_frontdoor.test.frontend_endpoints[local.endpoint_name]
  custom_https_provisioning_enabled = true

  custom_https_configuration {
    certificate_source                         = "AzureKeyVault"
    azure_key_vault_certificate_secret_name    = "accTest"
    azure_key_vault_certificate_secret_version = "latest"
  }
}
`, r.template(data))
}

func (r FrontDoorCustomHttpsConfigurationResource) EnabledKeyVaultLatestMissingAttributes(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_frontdoor_custom_https_configuration" "test" {
  frontend_endpoint_id              = azurerm_frontdoor.test.frontend_endpoints[local.endpoint_name]
  custom_https_provisioning_enabled = true

  custom_https_configuration {
    certificate_source                      = "AzureKeyVault"
    azure_key_vault_certificate_secret_name = "accTest"
  }
}
`, r.template(data))
}

func (r FrontDoorCustomHttpsConfigurationResource) EnabledKeyVaultMissingAttributes(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_frontdoor_custom_https_configuration" "test" {
  frontend_endpoint_id              = azurerm_frontdoor.test.frontend_endpoints[local.endpoint_name]
  custom_https_provisioning_enabled = true

  custom_https_configuration {
    certificate_source                         = "AzureKeyVault"
    azure_key_vault_certificate_secret_name    = "accTest"
    azure_key_vault_certificate_secret_version = "accTest"
  }
}
`, r.template(data))
}

func (FrontDoorCustomHttpsConfigurationResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-frontdoor-%d"
  location = "%s"
}

locals {
  backend_name        = "backend-bing-custom"
  endpoint_name       = "frontend-endpoint-custom"
  health_probe_name   = "health-probe-custom"
  load_balancing_name = "load-balancing-setting-custom"
}

resource "azurerm_frontdoor" "test" {
  name                = "acctest-FD-%d"
  resource_group_name = azurerm_resource_group.test.name

  backend_pool_settings {
    enforce_backend_pools_certificate_name_check = false
  }

  routing_rule {
    name               = "routing-rule"
    accepted_protocols = ["Http", "Https"]
    patterns_to_match  = ["/*"]
    frontend_endpoints = [local.endpoint_name]

    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = local.backend_name
    }
  }

  backend_pool_load_balancing {
    name = local.load_balancing_name
  }

  backend_pool_health_probe {
    name = local.health_probe_name
  }

  backend_pool {
    name = local.backend_name
    backend {
      host_header = "www.bing.com"
      address     = "www.bing.com"
      http_port   = 80
      https_port  = 443
    }

    load_balancing_name = local.load_balancing_name
    health_probe_name   = local.health_probe_name
  }

  frontend_endpoint {
    name      = local.endpoint_name
    host_name = "acctest-FD-%d.azurefd.net"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}
