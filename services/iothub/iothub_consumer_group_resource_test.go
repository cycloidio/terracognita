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

type IotHubConsumerGroupResource struct{}

func TestAccIotHubConsumerGroup_events(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_iothub_consumer_group", "test")
	r := IotHubConsumerGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data, "events"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("eventhub_endpoint_name").HasValue("events"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccIotHubConsumerGroup_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_iothub_consumer_group", "test")
	r := IotHubConsumerGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data, "events"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("eventhub_endpoint_name").HasValue("events"),
			),
		},
		{
			Config:      r.requiresImport(data, "events"),
			ExpectError: acceptance.RequiresImportError("azurerm_iothub_consumer_group"),
		},
	})
}

func TestAccIotHubConsumerGroup_operationsMonitoringEvents(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_iothub_consumer_group", "test")
	r := IotHubConsumerGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data, "operationsMonitoringEvents"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("eventhub_endpoint_name").HasValue("operationsMonitoringEvents"),
			),
		}, data.ImportStep(),
	})
}

func TestAccIotHubConsumerGroup_withSharedAccessPolicy(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_iothub_consumer_group", "test")
	r := IotHubConsumerGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.withSharedAccessPolicy(data, "events"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (t IotHubConsumerGroupResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.ConsumerGroupID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.IoTHub.ResourceClient.GetEventHubConsumerGroup(ctx, id.ResourceGroup, id.IotHubName, id.EventHubEndpointName, id.Name)
	if err != nil {
		return nil, fmt.Errorf("reading IotHuB Consumer Group (%s): %+v", id, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (IotHubConsumerGroupResource) basic(data acceptance.TestData, eventName string) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_iothub" "test" {
  name                = "acctestIoTHub-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location

  sku {
    name     = "B1"
    capacity = "1"
  }

  tags = {
    purpose = "testing"
  }
}

resource "azurerm_iothub_consumer_group" "test" {
  name                   = "test"
  iothub_name            = azurerm_iothub.test.name
  eventhub_endpoint_name = "%s"
  resource_group_name    = azurerm_resource_group.test.name
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, eventName)
}

func (r IotHubConsumerGroupResource) requiresImport(data acceptance.TestData, eventName string) string {
	return fmt.Sprintf(`
%s

resource "azurerm_iothub_consumer_group" "import" {
  name                   = azurerm_iothub_consumer_group.test.name
  iothub_name            = azurerm_iothub_consumer_group.test.iothub_name
  eventhub_endpoint_name = azurerm_iothub_consumer_group.test.eventhub_endpoint_name
  resource_group_name    = azurerm_iothub_consumer_group.test.resource_group_name
}
`, r.basic(data, eventName))
}

func (r IotHubConsumerGroupResource) withSharedAccessPolicy(data acceptance.TestData, eventName string) string {
	return fmt.Sprintf(`
%s

resource "azurerm_iothub_shared_access_policy" "test" {
  name                = "acctestSharedAccessPolicy"
  resource_group_name = azurerm_resource_group.test.name
  iothub_name         = azurerm_iothub.test.name
  service_connect     = true
}
`, r.basic(data, eventName))
}
