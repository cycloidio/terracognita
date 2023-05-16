package sentinel_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
)

type SentinelAlertRuleTemplateDataSource struct{}

func TestAccSentinelAlertRuleTemplateDataSource_fusion(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_sentinel_alert_rule_template", "test")
	r := SentinelAlertRuleTemplateDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.byDisplayName(data, "Advanced Multistage Attack Detection"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("id").Exists(),
				check.That(data.ResourceName).Key("display_name").Exists(),
				check.That(data.ResourceName).Key("name").Exists(),
				check.That(data.ResourceName).Key("scheduled_template").DoesNotExist(),
				check.That(data.ResourceName).Key("security_incident_template").DoesNotExist(),
			),
		},
	})
}

func TestAccSentinelAlertRuleTemplateDataSource_securityIncident(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_sentinel_alert_rule_template", "test")
	r := SentinelAlertRuleTemplateDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.byDisplayName(data, "Create incidents based on Azure Defender alerts"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("id").Exists(),
				check.That(data.ResourceName).Key("display_name").Exists(),
				check.That(data.ResourceName).Key("name").Exists(),
				check.That(data.ResourceName).Key("security_incident_template.0.description").Exists(),
				check.That(data.ResourceName).Key("security_incident_template.0.product_filter").Exists(),
				check.That(data.ResourceName).Key("scheduled_template").DoesNotExist(),
			),
		},
	})
}

func TestAccSentinelAlertRuleTemplateDataSource_scheduled(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_sentinel_alert_rule_template", "test")
	r := SentinelAlertRuleTemplateDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.byDisplayName(data, "Malware in the recycle bin"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("id").Exists(),
				check.That(data.ResourceName).Key("display_name").Exists(),
				check.That(data.ResourceName).Key("name").Exists(),
				check.That(data.ResourceName).Key("scheduled_template.0.description").Exists(),
				check.That(data.ResourceName).Key("scheduled_template.0.severity").Exists(),
				check.That(data.ResourceName).Key("scheduled_template.0.query").Exists(),
				check.That(data.ResourceName).Key("scheduled_template.0.query_frequency").Exists(),
				check.That(data.ResourceName).Key("scheduled_template.0.query_period").Exists(),
				check.That(data.ResourceName).Key("scheduled_template.0.trigger_operator").Exists(),
				check.That(data.ResourceName).Key("scheduled_template.0.trigger_threshold").Exists(),
				check.That(data.ResourceName).Key("security_incident_template").DoesNotExist(),
			),
		},
	})
}

func (SentinelAlertRuleTemplateDataSource) byDisplayName(data acceptance.TestData, displayName string) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-sentinel-%d"
  location = "%s"
}

resource "azurerm_log_analytics_workspace" "test" {
  name                = "acctestLAW-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku                 = "PerGB2018"
}

resource "azurerm_log_analytics_solution" "test" {
  solution_name         = "SecurityInsights"
  location              = azurerm_resource_group.test.location
  resource_group_name   = azurerm_resource_group.test.name
  workspace_resource_id = azurerm_log_analytics_workspace.test.id
  workspace_name        = azurerm_log_analytics_workspace.test.name

  plan {
    publisher = "Microsoft"
    product   = "OMSGallery/SecurityInsights"
  }
}

data "azurerm_sentinel_alert_rule_template" "test" {
  display_name               = "%s"
  log_analytics_workspace_id = azurerm_log_analytics_solution.test.workspace_resource_id
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, displayName)
}
