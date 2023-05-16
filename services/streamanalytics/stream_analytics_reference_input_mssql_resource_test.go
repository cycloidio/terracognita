package streamanalytics_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/streamanalytics/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type StreamAnalyticsReferenceInputMsSqlResource struct{}

func TestAccStreamAnalyticsReferenceInputMsSql_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_reference_input_mssql", "test")
	r := StreamAnalyticsReferenceInputMsSqlResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("password"),
	})
}

func TestAccStreamAnalyticsReferenceInputMsSql_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_reference_input_mssql", "test")
	r := StreamAnalyticsReferenceInputMsSqlResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config: r.updated(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("password"),
	})
}

func TestAccStreamAnalyticsReferenceInputMsSql_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_reference_input_mssql", "test")
	r := StreamAnalyticsReferenceInputMsSqlResource{}

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

func (r StreamAnalyticsReferenceInputMsSqlResource) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.StreamInputID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := client.StreamAnalytics.InputsClient.Get(ctx, id.ResourceGroup, id.StreamingjobName, id.InputName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return utils.Bool(false), nil
		}
		return nil, fmt.Errorf("reading (%s): %+v", *id, err)
	}
	return utils.Bool(true), nil
}

func (r StreamAnalyticsReferenceInputMsSqlResource) basic(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_stream_analytics_reference_input_mssql" "test" {
  name                      = "acctestinput-%d"
  stream_analytics_job_name = azurerm_stream_analytics_job.test.name
  resource_group_name       = azurerm_stream_analytics_job.test.resource_group_name
  server                    = azurerm_mssql_server.test.fully_qualified_domain_name
  database                  = azurerm_mssql_database.test.name
  username                  = "maurice"
  password                  = "ludicrousdisplay"
  refresh_type              = "RefreshPeriodicallyWithFull"
  refresh_interval_duration = "00:10:00"
  full_snapshot_query       = <<QUERY
    SELECT *
    INTO [YourOutputAlias]
    FROM [YourInputAlias]
QUERY

}
`, template, data.RandomInteger)
}

func (r StreamAnalyticsReferenceInputMsSqlResource) updated(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_stream_analytics_reference_input_mssql" "test" {
  name                      = "acctestinput-%d"
  stream_analytics_job_name = azurerm_stream_analytics_job.test.name
  resource_group_name       = azurerm_stream_analytics_job.test.resource_group_name
  server                    = azurerm_mssql_server.test.fully_qualified_domain_name
  database                  = azurerm_mssql_database.test.name
  username                  = "maurice"
  password                  = "ludicrousdisplay"
  refresh_type              = "RefreshPeriodicallyWithDelta"
  refresh_interval_duration = "00:20:00"
  full_snapshot_query       = <<QUERY
    SELECT *
    INTO [YourOutputAlias]
    FROM [YourInputAlias]
QUERY
  delta_snapshot_query      = <<QUERY
    SELECT *
    INTO [YourOutputAlias]
    FROM [YourInputAlias]
QUERY

}
`, template, data.RandomInteger)
}

func (r StreamAnalyticsReferenceInputMsSqlResource) requiresImport(data acceptance.TestData) string {
	template := r.basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_stream_analytics_reference_input_mssql" "import" {
  name                      = azurerm_stream_analytics_reference_input_mssql.test.name
  stream_analytics_job_name = azurerm_stream_analytics_job.test.name
  resource_group_name       = azurerm_stream_analytics_job.test.resource_group_name
  server                    = azurerm_stream_analytics_reference_input_mssql.test.server
  database                  = azurerm_stream_analytics_reference_input_mssql.test.database
  username                  = azurerm_stream_analytics_reference_input_mssql.test.username
  password                  = azurerm_stream_analytics_reference_input_mssql.test.password
  refresh_type              = azurerm_stream_analytics_reference_input_mssql.test.refresh_type
  full_snapshot_query       = azurerm_stream_analytics_reference_input_mssql.test.full_snapshot_query

}
`, template)
}

func (r StreamAnalyticsReferenceInputMsSqlResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%s"
}

resource "azurerm_stream_analytics_job" "test" {
  name                                     = "acctestjob-%[1]d"
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

resource "azurerm_mssql_server" "test" {
  name                         = "acctest-sqlserver-%[1]d"
  resource_group_name          = azurerm_resource_group.test.name
  location                     = azurerm_resource_group.test.location
  version                      = "12.0"
  administrator_login          = "mradministrator"
  administrator_login_password = "thisIsDog11"
}

resource "azurerm_mssql_database" "test" {
  name      = "acctest-db-%[1]d"
  server_id = azurerm_mssql_server.test.id
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString)
}
