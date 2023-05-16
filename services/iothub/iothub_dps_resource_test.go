package iothub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/iothub/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type IotHubDPSResource struct{}

func TestAccIotHubDPS_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_iothub_dps", "test")
	r := IotHubDPSResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("allocation_policy").Exists(),
				check.That(data.ResourceName).Key("device_provisioning_host_name").Exists(),
				check.That(data.ResourceName).Key("id_scope").Exists(),
				check.That(data.ResourceName).Key("service_operations_host_name").Exists(),
			),
		},
		data.ImportStep(),
	})
}

func TestAccIotHubDPS_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_iothub_dps", "test")
	r := IotHubDPSResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_iothub_dps"),
		},
	})
}

func TestAccIotHubDPS_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_iothub_dps", "test")
	r := IotHubDPSResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("public_network_access_enabled").HasValue("true"),
			),
		},
		data.ImportStep(),
		{
			Config: r.update(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("public_network_access_enabled").HasValue("false"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccIotHubDPS_linkedHubs(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_iothub_dps", "test")
	r := IotHubDPSResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.linkedHubs(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("allocation_policy").Exists(),
			),
		},
		data.ImportStep(),
		{
			Config: r.linkedHubsUpdated(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccIotHubDPS_ipFilterRules(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_iothub_dps", "test")
	r := IotHubDPSResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.ipFilterRules(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("allocation_policy").Exists(),
				check.That(data.ResourceName).Key("device_provisioning_host_name").Exists(),
				check.That(data.ResourceName).Key("id_scope").Exists(),
				check.That(data.ResourceName).Key("service_operations_host_name").Exists(),
				check.That(data.ResourceName).Key("public_network_access_enabled").HasValue("false"),
				check.That(data.ResourceName).Key("ip_filter_rule.0.name").HasValue("test"),
				check.That(data.ResourceName).Key("ip_filter_rule.0.ip_mask").HasValue("10.0.0.0/31"),
				check.That(data.ResourceName).Key("ip_filter_rule.0.action").HasValue("Accept"),
				check.That(data.ResourceName).Key("ip_filter_rule.0.target").HasValue("All"),
				check.That(data.ResourceName).Key("ip_filter_rule.1.name").HasValue("test2"),
				check.That(data.ResourceName).Key("ip_filter_rule.1.ip_mask").HasValue("10.0.2.0/31"),
				check.That(data.ResourceName).Key("ip_filter_rule.1.action").HasValue("Accept"),
				check.That(data.ResourceName).Key("ip_filter_rule.1.target").HasValue("ServiceApi"),
			),
		},
		{
			Config: r.ipFilterRulesUpdated(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("ip_filter_rule.0.name").HasValue("test"),
				check.That(data.ResourceName).Key("ip_filter_rule.0.ip_mask").HasValue("10.0.0.0/31"),
				check.That(data.ResourceName).Key("ip_filter_rule.0.action").HasValue("Reject"),
				check.That(data.ResourceName).Key("ip_filter_rule.0.target").HasValue("All"),
				check.That(data.ResourceName).Key("ip_filter_rule.1.name").HasValue("test2"),
				check.That(data.ResourceName).Key("ip_filter_rule.1.ip_mask").HasValue("10.0.2.0/31"),
				check.That(data.ResourceName).Key("ip_filter_rule.1.action").HasValue("Reject"),
				check.That(data.ResourceName).Key("ip_filter_rule.1.target").HasValue("DeviceApi"),
				check.That(data.ResourceName).Key("public_network_access_enabled").HasValue("false"),
			),
		},
		data.ImportStep(),
	})
}

func (t IotHubDPSResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.IotHubDpsID(state.ID)
	if err != nil {
		return nil, err
	}
	resourceGroup := id.ResourceGroup
	name := id.ProvisioningServiceName

	resp, err := clients.IoTHub.DPSResourceClient.Get(ctx, name, resourceGroup)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %+v", id, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (IotHubDPSResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_iothub_dps" "test" {
  name                = "acctestIoTDPS-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location

  sku {
    name     = "S1"
    capacity = "1"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r IotHubDPSResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_iothub_dps" "import" {
  name                = azurerm_iothub_dps.test.name
  resource_group_name = azurerm_iothub_dps.test.resource_group_name
  location            = azurerm_iothub_dps.test.location

  sku {
    name     = "S1"
    capacity = "1"
  }
}
`, r.basic(data))
}

func (IotHubDPSResource) update(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_iothub_dps" "test" {
  name                          = "acctestIoTDPS-%d"
  resource_group_name           = azurerm_resource_group.test.name
  location                      = azurerm_resource_group.test.location
  public_network_access_enabled = false

  sku {
    name     = "S1"
    capacity = "1"
  }

  tags = {
    purpose = "testing"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (IotHubDPSResource) linkedHubs(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_iothub_dps" "test" {
  name                = "acctestIoTDPS-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location

  sku {
    name     = "S1"
    capacity = "1"
  }

  linked_hub {
    connection_string       = "HostName=test.azure-devices.net;SharedAccessKeyName=iothubowner;SharedAccessKey=booo"
    location                = azurerm_resource_group.test.location
    allocation_weight       = 15
    apply_allocation_policy = true
  }

  linked_hub {
    connection_string = "HostName=test2.azure-devices.net;SharedAccessKeyName=iothubowner2;SharedAccessKey=key2"
    location          = azurerm_resource_group.test.location
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (IotHubDPSResource) linkedHubsUpdated(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_iothub_dps" "test" {
  name                = "acctestIoTDPS-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location

  sku {
    name     = "S1"
    capacity = "1"
  }

  linked_hub {
    connection_string       = "HostName=test3.azure-devices.net;SharedAccessKeyName=iothubowner;SharedAccessKey=booo"
    location                = azurerm_resource_group.test.location
    allocation_weight       = 150
    apply_allocation_policy = true
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (IotHubDPSResource) ipFilterRules(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_iothub_dps" "test" {
  name                          = "acctestIoTDPS-%d"
  resource_group_name           = azurerm_resource_group.test.name
  location                      = azurerm_resource_group.test.location
  public_network_access_enabled = false

  ip_filter_rule {
    name    = "test"
    ip_mask = "10.0.0.0/31"
    action  = "Accept"
    target  = "All"
  }

  ip_filter_rule {
    name    = "test2"
    ip_mask = "10.0.2.0/31"
    action  = "Accept"
    target  = "ServiceApi"
  }

  ip_filter_rule {
    name    = "test3"
    ip_mask = "10.0.3.0/31"
    action  = "Accept"
  }

  sku {
    name     = "S1"
    capacity = "1"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (IotHubDPSResource) ipFilterRulesUpdated(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_iothub_dps" "test" {
  name                          = "acctestIoTDPS-%d"
  resource_group_name           = azurerm_resource_group.test.name
  location                      = azurerm_resource_group.test.location
  public_network_access_enabled = false

  ip_filter_rule {
    name    = "test"
    ip_mask = "10.0.0.0/31"
    action  = "Reject"
    target  = "All"
  }

  ip_filter_rule {
    name    = "test2"
    ip_mask = "10.0.2.0/31"
    action  = "Reject"
    target  = "DeviceApi"
  }

  ip_filter_rule {
    name    = "test3"
    ip_mask = "10.0.3.0/31"
    action  = "Accept"
  }

  sku {
    name     = "S1"
    capacity = "1"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
