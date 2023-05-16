package monitor_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
)

type MonitorScheduledQueryRulesLogDataSource struct{}

func TestAccDataSourceMonitorScheduledQueryRules_LogToMetricAction(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_monitor_scheduled_query_rules_log", "test")
	r := MonitorScheduledQueryRulesLogDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.LogToMetricActionConfig(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("id").Exists(),
			),
		},
	})
}

func (r MonitorScheduledQueryRulesLogDataSource) LogToMetricActionConfig(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurerm_monitor_scheduled_query_rules_log" "test" {
  name                = basename(azurerm_monitor_scheduled_query_rules_log.test.id)
  resource_group_name = "${azurerm_resource_group.test.name}"
}
`, MonitorScheduledQueryRulesLogResource{}.LogToMetricActionConfigBasic(data))
}
