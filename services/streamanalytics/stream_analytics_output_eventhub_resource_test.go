package streamanalytics_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type StreamAnalyticsOutputEventhubResource struct{}

func TestAccStreamAnalyticsOutputEventHub_avro(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_output_eventhub", "test")
	r := StreamAnalyticsOutputEventhubResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.avro(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("serialization.0.type").HasValue("Avro"),
			),
		},
		data.ImportStep("shared_access_policy_key"),
	})
}

func TestAccStreamAnalyticsOutputEventHub_csv(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_output_eventhub", "test")
	r := StreamAnalyticsOutputEventhubResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.csv(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("serialization.0.type").HasValue("Csv"),
				check.That(data.ResourceName).Key("serialization.0.field_delimiter").HasValue(","),
				check.That(data.ResourceName).Key("serialization.0.encoding").HasValue("UTF8"),
			),
		},
		data.ImportStep("shared_access_policy_key"),
	})
}

func TestAccStreamAnalyticsOutputEventHub_json(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_output_eventhub", "test")
	r := StreamAnalyticsOutputEventhubResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.json(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("serialization.0.format").HasValue("LineSeparated"),
			),
		},
		data.ImportStep("shared_access_policy_key"),
	})
}

func TestAccStreamAnalyticsOutputEventHub_jsonArrayFormat(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_output_eventhub", "test")
	r := StreamAnalyticsOutputEventhubResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.jsonArrayFormat(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("serialization.0.format").HasValue("Array"),
				check.That(data.ResourceName).Key("serialization.0.type").HasValue("Json"),
				check.That(data.ResourceName).Key("serialization.0.encoding").HasValue("UTF8"),
			),
		},
		data.ImportStep("shared_access_policy_key"),
	})
}

func TestAccStreamAnalyticsOutputEventHub_propertyColumns(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_output_eventhub", "test")
	r := StreamAnalyticsOutputEventhubResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.propertyColumns(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("property_columns.0").HasValue("col1"),
				check.That(data.ResourceName).Key("property_columns.1").HasValue("col2"),
			),
		},
		data.ImportStep("shared_access_policy_key"),
	})
}

func TestAccStreamAnalyticsOutputEventHub_partitionKey(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_output_eventhub", "test")
	r := StreamAnalyticsOutputEventhubResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.partitionKey(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("partition_key").HasValue("partitionKey"),
			),
		},
		data.ImportStep("shared_access_policy_key"),
	})
}

func TestAccStreamAnalyticsOutputEventHub_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_output_eventhub", "test")
	r := StreamAnalyticsOutputEventhubResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.json(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config: r.updated(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("serialization.0.type").HasValue("Avro"),
			),
		},
		data.ImportStep("shared_access_policy_key"),
	})
}

func TestAccStreamAnalyticsOutputEventHub_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_output_eventhub", "test")
	r := StreamAnalyticsOutputEventhubResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.json(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func (r StreamAnalyticsOutputEventhubResource) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	name := state.Attributes["name"]
	jobName := state.Attributes["stream_analytics_job_name"]
	resourceGroup := state.Attributes["resource_group_name"]

	resp, err := client.StreamAnalytics.OutputsClient.Get(ctx, resourceGroup, jobName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return utils.Bool(false), nil
		}
		return nil, fmt.Errorf("retrieving Stream Output %q (Stream Analytics Job %q / Resource Group %q): %+v", name, jobName, resourceGroup, err)
	}
	return utils.Bool(true), nil
}

func (r StreamAnalyticsOutputEventhubResource) avro(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_stream_analytics_output_eventhub" "test" {
  name                      = "acctestinput-%d"
  stream_analytics_job_name = azurerm_stream_analytics_job.test.name
  resource_group_name       = azurerm_stream_analytics_job.test.resource_group_name
  eventhub_name             = azurerm_eventhub.test.name
  servicebus_namespace      = azurerm_eventhub_namespace.test.name
  shared_access_policy_key  = azurerm_eventhub_namespace.test.default_primary_key
  shared_access_policy_name = "RootManageSharedAccessKey"

  serialization {
    type = "Avro"
  }
}
`, template, data.RandomInteger)
}

func (r StreamAnalyticsOutputEventhubResource) csv(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_stream_analytics_output_eventhub" "test" {
  name                      = "acctestinput-%d"
  stream_analytics_job_name = azurerm_stream_analytics_job.test.name
  resource_group_name       = azurerm_stream_analytics_job.test.resource_group_name
  eventhub_name             = azurerm_eventhub.test.name
  servicebus_namespace      = azurerm_eventhub_namespace.test.name
  shared_access_policy_key  = azurerm_eventhub_namespace.test.default_primary_key
  shared_access_policy_name = "RootManageSharedAccessKey"

  serialization {
    type            = "Csv"
    encoding        = "UTF8"
    field_delimiter = ","
  }
}
`, template, data.RandomInteger)
}

func (r StreamAnalyticsOutputEventhubResource) propertyColumns(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_stream_analytics_output_eventhub" "test" {
  name                      = "acctestinput-%d"
  stream_analytics_job_name = azurerm_stream_analytics_job.test.name
  resource_group_name       = azurerm_stream_analytics_job.test.resource_group_name
  eventhub_name             = azurerm_eventhub.test.name
  servicebus_namespace      = azurerm_eventhub_namespace.test.name
  shared_access_policy_key  = azurerm_eventhub_namespace.test.default_primary_key
  shared_access_policy_name = "RootManageSharedAccessKey"
  property_columns          = ["col1", "col2"]

  serialization {
    type            = "Csv"
    encoding        = "UTF8"
    field_delimiter = ","
  }
}
`, template, data.RandomInteger)
}

func (r StreamAnalyticsOutputEventhubResource) partitionKey(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_stream_analytics_output_eventhub" "test" {
  name                      = "acctestinput-%d"
  stream_analytics_job_name = azurerm_stream_analytics_job.test.name
  resource_group_name       = azurerm_stream_analytics_job.test.resource_group_name
  eventhub_name             = azurerm_eventhub.test.name
  servicebus_namespace      = azurerm_eventhub_namespace.test.name
  shared_access_policy_key  = azurerm_eventhub_namespace.test.default_primary_key
  shared_access_policy_name = "RootManageSharedAccessKey"
  partition_key             = "partitionKey"

  serialization {
    type            = "Csv"
    encoding        = "UTF8"
    field_delimiter = ","
  }
}
`, template, data.RandomInteger)
}

func (r StreamAnalyticsOutputEventhubResource) json(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_stream_analytics_output_eventhub" "test" {
  name                      = "acctestinput-%d"
  stream_analytics_job_name = azurerm_stream_analytics_job.test.name
  resource_group_name       = azurerm_stream_analytics_job.test.resource_group_name
  eventhub_name             = azurerm_eventhub.test.name
  servicebus_namespace      = azurerm_eventhub_namespace.test.name
  shared_access_policy_key  = azurerm_eventhub_namespace.test.default_primary_key
  shared_access_policy_name = "RootManageSharedAccessKey"

  serialization {
    type     = "Json"
    encoding = "UTF8"
    format   = "LineSeparated"
  }
}
`, template, data.RandomInteger)
}

func (r StreamAnalyticsOutputEventhubResource) jsonArrayFormat(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_stream_analytics_output_eventhub" "test" {
  name                      = "acctestinput-%d"
  stream_analytics_job_name = azurerm_stream_analytics_job.test.name
  resource_group_name       = azurerm_stream_analytics_job.test.resource_group_name
  eventhub_name             = azurerm_eventhub.test.name
  servicebus_namespace      = azurerm_eventhub_namespace.test.name
  shared_access_policy_key  = azurerm_eventhub_namespace.test.default_primary_key
  shared_access_policy_name = "RootManageSharedAccessKey"

  serialization {
    type     = "Json"
    encoding = "UTF8"
    format   = "Array"
  }
}
`, template, data.RandomInteger)
}

func (r StreamAnalyticsOutputEventhubResource) updated(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_eventhub_namespace" "updated" {
  name                = "acctestehn2-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku                 = "Standard"
  capacity            = 1
}

resource "azurerm_eventhub" "updated" {
  name                = "acctesteh2-%d"
  namespace_name      = azurerm_eventhub_namespace.updated.name
  resource_group_name = azurerm_resource_group.test.name
  partition_count     = 2
  message_retention   = 1
}

resource "azurerm_stream_analytics_output_eventhub" "test" {
  name                      = "acctestinput-%d"
  stream_analytics_job_name = azurerm_stream_analytics_job.test.name
  resource_group_name       = azurerm_stream_analytics_job.test.resource_group_name
  eventhub_name             = azurerm_eventhub.updated.name
  servicebus_namespace      = azurerm_eventhub_namespace.updated.name
  shared_access_policy_key  = azurerm_eventhub_namespace.updated.default_primary_key
  shared_access_policy_name = "RootManageSharedAccessKey"

  serialization {
    type = "Avro"
  }
}
`, template, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (r StreamAnalyticsOutputEventhubResource) requiresImport(data acceptance.TestData) string {
	template := r.json(data)
	return fmt.Sprintf(`
%s

resource "azurerm_stream_analytics_output_eventhub" "import" {
  name                      = azurerm_stream_analytics_output_eventhub.test.name
  stream_analytics_job_name = azurerm_stream_analytics_output_eventhub.test.stream_analytics_job_name
  resource_group_name       = azurerm_stream_analytics_output_eventhub.test.resource_group_name
  eventhub_name             = azurerm_stream_analytics_output_eventhub.test.eventhub_name
  servicebus_namespace      = azurerm_stream_analytics_output_eventhub.test.servicebus_namespace
  shared_access_policy_key  = azurerm_stream_analytics_output_eventhub.test.shared_access_policy_key
  shared_access_policy_name = azurerm_stream_analytics_output_eventhub.test.shared_access_policy_name

  serialization {
    type     = azurerm_stream_analytics_output_eventhub.test.serialization.0.type
    encoding = azurerm_stream_analytics_output_eventhub.test.serialization.0.encoding
    format   = azurerm_stream_analytics_output_eventhub.test.serialization.0.format
  }
}
`, template)
}

func (r StreamAnalyticsOutputEventhubResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_eventhub_namespace" "test" {
  name                = "acctestehn-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku                 = "Standard"
  capacity            = 1
}

resource "azurerm_eventhub" "test" {
  name                = "acctesteh-%d"
  namespace_name      = azurerm_eventhub_namespace.test.name
  resource_group_name = azurerm_resource_group.test.name
  partition_count     = 2
  message_retention   = 1
}

resource "azurerm_stream_analytics_job" "test" {
  name                                     = "acctestjob-%d"
  resource_group_name                      = azurerm_resource_group.test.name
  location                                 = azurerm_resource_group.test.location
  compatibility_level                      = "1.0"
  data_locale                              = "en-GB"
  events_late_arrival_max_delay_in_seconds = 60
  events_out_of_order_max_delay_in_seconds = 50
  events_out_of_order_policy               = "Adjust"
  output_error_policy                      = "Drop"
  streaming_units                          = 3

  transformation_query = <<QUERY
    SELECT *
    INTO [YourOutputAlias]
    FROM [YourInputAlias]
QUERY

}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}
