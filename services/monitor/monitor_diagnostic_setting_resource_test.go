package monitor_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/monitor"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type MonitorDiagnosticSettingResource struct{}

func TestAccMonitorDiagnosticSetting_eventhub(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_diagnostic_setting", "test")
	r := MonitorDiagnosticSettingResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.eventhub(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("eventhub_name").Exists(),
				check.That(data.ResourceName).Key("eventhub_authorization_rule_id").Exists(),
				check.That(data.ResourceName).Key("log.#").HasValue("2"),
				check.That(data.ResourceName).Key("metric.#").HasValue("1"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMonitorDiagnosticSetting_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_diagnostic_setting", "test")
	r := MonitorDiagnosticSettingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.eventhub(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_monitor_diagnostic_setting"),
		},
	})
}

func TestAccMonitorDiagnosticSetting_logAnalyticsWorkspace(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_diagnostic_setting", "test")
	r := MonitorDiagnosticSettingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.logAnalyticsWorkspace(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("log_analytics_workspace_id").Exists(),
				check.That(data.ResourceName).Key("log.#").HasValue("2"),
				check.That(data.ResourceName).Key("metric.#").HasValue("1"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMonitorDiagnosticSetting_logAnalyticsWorkspaceDedicated(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_diagnostic_setting", "test")
	r := MonitorDiagnosticSettingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.logAnalyticsWorkspaceDedicated(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMonitorDiagnosticSetting_storageAccount(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_diagnostic_setting", "test")
	r := MonitorDiagnosticSettingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.storageAccount(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("storage_account_id").Exists(),
				check.That(data.ResourceName).Key("log.#").HasValue("2"),
				check.That(data.ResourceName).Key("metric.#").HasValue("1"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMonitorDiagnosticSetting_activityLog(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_monitor_diagnostic_setting", "test")
	r := MonitorDiagnosticSettingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.activityLog(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (t MonitorDiagnosticSettingResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := monitor.ParseMonitorDiagnosticId(state.ID)
	if err != nil {
		return nil, err
	}
	actualResourceId := id.ResourceID
	targetResourceId := strings.TrimPrefix(actualResourceId, "/")

	resp, err := clients.Monitor.DiagnosticSettingsClient.Get(ctx, targetResourceId, id.Name)
	if err != nil {
		return nil, fmt.Errorf("reading diagnostic setting (%s): %+v", id, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (MonitorDiagnosticSettingResource) eventhub(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

data "azurerm_client_config" "current" {
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurerm_eventhub_namespace" "test" {
  name                = "acctest-EHN-%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku                 = "Basic"
}

resource "azurerm_eventhub" "test" {
  name                = "acctest-EH-%[1]d"
  namespace_name      = azurerm_eventhub_namespace.test.name
  resource_group_name = azurerm_resource_group.test.name
  partition_count     = 2
  message_retention   = 1
}

resource "azurerm_eventhub_namespace_authorization_rule" "test" {
  name                = "example"
  namespace_name      = azurerm_eventhub_namespace.test.name
  resource_group_name = azurerm_resource_group.test.name
  listen              = true
  send                = true
  manage              = true
}

resource "azurerm_key_vault" "test" {
  name                = "acctest%[3]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  tenant_id           = data.azurerm_client_config.current.tenant_id
  sku_name            = "standard"
}

resource "azurerm_monitor_diagnostic_setting" "test" {
  name                           = "acctest-DS-%[1]d"
  target_resource_id             = azurerm_key_vault.test.id
  eventhub_authorization_rule_id = azurerm_eventhub_namespace_authorization_rule.test.id
  eventhub_name                  = azurerm_eventhub.test.name

  log {
    category = "AuditEvent"
    enabled  = false

    retention_policy {
      enabled = false
    }
  }

  log {
    category = "AzurePolicyEvaluationDetails"
    enabled  = false

    retention_policy {
      days    = 0
      enabled = false
    }
  }

  metric {
    category = "AllMetrics"
    enabled  = true

    retention_policy {
      enabled = false
      days    = 7
    }
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomIntOfLength(17))
}

func (r MonitorDiagnosticSettingResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_monitor_diagnostic_setting" "import" {
  name                           = azurerm_monitor_diagnostic_setting.test.name
  target_resource_id             = azurerm_monitor_diagnostic_setting.test.target_resource_id
  eventhub_authorization_rule_id = azurerm_monitor_diagnostic_setting.test.eventhub_authorization_rule_id
  eventhub_name                  = azurerm_monitor_diagnostic_setting.test.eventhub_name

  log {
    category = "AuditEvent"
    enabled  = false

    retention_policy {
      enabled = false
    }
  }

  metric {
    category = "AllMetrics"

    retention_policy {
      enabled = false
    }
  }
}
`, r.eventhub(data))
}

func (MonitorDiagnosticSettingResource) logAnalyticsWorkspace(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

data "azurerm_client_config" "current" {
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurerm_log_analytics_workspace" "test" {
  name                = "acctest-LAW-%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

resource "azurerm_key_vault" "test" {
  name                = "acctest%[3]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  tenant_id           = data.azurerm_client_config.current.tenant_id
  sku_name            = "standard"
}

resource "azurerm_monitor_diagnostic_setting" "test" {
  name                       = "acctest-DS-%[1]d"
  target_resource_id         = azurerm_key_vault.test.id
  log_analytics_workspace_id = azurerm_log_analytics_workspace.test.id

  log {
    category = "AuditEvent"
    enabled  = false

    retention_policy {
      enabled = false
    }
  }

  log {
    category = "AzurePolicyEvaluationDetails"
    enabled  = false

    retention_policy {
      days    = 0
      enabled = false
    }
  }

  metric {
    category = "AllMetrics"

    retention_policy {
      enabled = false
    }
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomIntOfLength(17))
}

func (MonitorDiagnosticSettingResource) logAnalyticsWorkspaceDedicated(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

data "azurerm_client_config" "current" {
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurerm_log_analytics_workspace" "test" {
  name                = "acctest-LAW-%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

resource "azurerm_data_factory" "test" {
  name                = "acctest-DF-%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_monitor_diagnostic_setting" "test" {
  name                       = "acctest-DS-%[1]d"
  target_resource_id         = azurerm_data_factory.test.id
  log_analytics_workspace_id = azurerm_log_analytics_workspace.test.id

  log_analytics_destination_type = "Dedicated"

  log {
    category = "ActivityRuns"
    retention_policy {
      enabled = false
      days    = 0
    }
  }

  log {
    category = "PipelineRuns"
    retention_policy {
      enabled = false
      days    = 0
    }
  }

  log {
    category = "TriggerRuns"
    retention_policy {
      days    = 0
      enabled = false
    }
  }

  log {
    category = "SSISIntegrationRuntimeLogs"
    retention_policy {
      days    = 0
      enabled = false
    }
  }

  log {
    category = "SSISPackageEventMessageContext"
    retention_policy {
      days    = 0
      enabled = false
    }
  }

  log {
    category = "SSISPackageEventMessages"
    retention_policy {
      days    = 0
      enabled = false
    }
  }

  log {
    category = "SSISPackageExecutableStatistics"
    retention_policy {
      days    = 0
      enabled = false
    }
  }

  log {
    category = "SSISPackageExecutionComponentPhases"
    retention_policy {
      days    = 0
      enabled = false
    }
  }

  log {
    category = "SSISPackageExecutionDataStatistics"
    retention_policy {
      days    = 0
      enabled = false
    }
  }

  log {
    category = "SandboxActivityRuns"
    retention_policy {
      days    = 0
      enabled = false
    }
  }

  log {
    category = "SandboxPipelineRuns"
    retention_policy {
      days    = 0
      enabled = false
    }
  }

  log {
    category = "AirflowDagProcessingLogs"
    enabled  = false

    retention_policy {
      days    = 0
      enabled = false
    }
  }

  log {
    category = "AirflowSchedulerLogs"
    enabled  = false

    retention_policy {
      days    = 0
      enabled = false
    }
  }

  log {
    category = "AirflowTaskLogs"
    enabled  = false

    retention_policy {
      days    = 0
      enabled = false
    }
  }

  log {
    category = "AirflowWebLogs"
    enabled  = false

    retention_policy {
      days    = 0
      enabled = false
    }
  }

  log {
    category = "AirflowWorkerLogs"
    enabled  = false

    retention_policy {
      days    = 0
      enabled = false
    }
  }

  metric {
    category = "AllMetrics"
    retention_policy {
      enabled = false
    }
  }
}
`, data.RandomInteger, data.Locations.Primary)
}

func (MonitorDiagnosticSettingResource) storageAccount(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

data "azurerm_client_config" "current" {
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurerm_storage_account" "test" {
  name                     = "acctest%[3]d"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_replication_type = "LRS"
  account_tier             = "Standard"
}

resource "azurerm_key_vault" "test" {
  name                = "acctest%[3]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  tenant_id           = data.azurerm_client_config.current.tenant_id
  sku_name            = "standard"
}

resource "azurerm_monitor_diagnostic_setting" "test" {
  name               = "acctest-DS-%[1]d"
  target_resource_id = azurerm_key_vault.test.id
  storage_account_id = azurerm_storage_account.test.id

  log {
    category = "AuditEvent"
    enabled  = false

    retention_policy {
      enabled = false
    }
  }

  log {
    category = "AzurePolicyEvaluationDetails"
    enabled  = false

    retention_policy {
      days    = 0
      enabled = false
    }
  }

  metric {
    category = "AllMetrics"

    retention_policy {
      enabled = false
    }
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomIntOfLength(17))
}

func (MonitorDiagnosticSettingResource) activityLog(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

data "azurerm_client_config" "current" {
}


data "azurerm_subscription" "current" {
}


resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurerm_storage_account" "test" {
  name                     = "acctest%[3]d"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_replication_type = "LRS"
  account_tier             = "Standard"
}


resource "azurerm_monitor_diagnostic_setting" "test" {
  name               = "acctest-DS-%[1]d"
  target_resource_id = data.azurerm_subscription.current.id
  storage_account_id = azurerm_storage_account.test.id

  log {
    category = "Administrative"
    enabled  = true
  }

  log {
    category = "Alert"
    enabled  = true
  }

  log {
    category = "Autoscale"
    enabled  = true
  }

  log {
    category = "Policy"
    enabled  = true
  }

  log {
    category = "Recommendation"
    enabled  = true
  }

  log {
    category = "ResourceHealth"
    enabled  = true
  }

  log {
    category = "Security"
    enabled  = true
  }

  log {
    category = "ServiceHealth"
    enabled  = true
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomIntOfLength(17))
}
