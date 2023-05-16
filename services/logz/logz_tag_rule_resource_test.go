package logz_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/logz/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type LogzTagRuleResource struct{}

func TestAccLogzTagRule_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_logz_tag_rule", "test")
	r := LogzTagRuleResource{}
	email := uuid.New().String()
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data, email),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLogzTagRule_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_logz_tag_rule", "test")
	r := LogzTagRuleResource{}
	email := uuid.New().String()
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data, email),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data, email),
			ExpectError: acceptance.RequiresImportError(data.ResourceType),
		},
	})
}

func TestAccLogzTagRule_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_logz_tag_rule", "test")
	r := LogzTagRuleResource{}
	email := uuid.New().String()
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data, email),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLogzTagRule_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_logz_tag_rule", "test")
	r := LogzTagRuleResource{}
	email := uuid.New().String()
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data, email),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data, email),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.basic(data, email),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (r LogzTagRuleResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.LogzTagRuleID(state.ID)
	if err != nil {
		return nil, err
	}
	resp, err := clients.Logz.TagRuleClient.Get(ctx, id.ResourceGroup, id.MonitorName, id.TagRuleName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return utils.Bool(false), nil
		}
		return nil, fmt.Errorf("retrieving %s: %+v", id, err)
	}
	return utils.Bool(true), nil
}

func (r LogzTagRuleResource) template(data acceptance.TestData, email string) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctest-logz-%d"
  location = "%s"
}

resource "azurerm_logz_monitor" "test" {
  name                = "%s"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  plan {
    billing_cycle  = "MONTHLY"
    effective_date = "%s"
    plan_id        = "100gb14days"
    usage_type     = "COMMITTED"
  }

  user {
    email        = "%s@example.com"
    first_name   = "first"
    last_name    = "last"
    phone_number = "123456"
  }
}
`, data.RandomInteger, data.Locations.Primary, getLogzInstanceName(data.RandomInteger), getEffectiveDate(), email)
}

func (r LogzTagRuleResource) basic(data acceptance.TestData, email string) string {
	template := r.template(data, email)
	return fmt.Sprintf(`
%s

resource "azurerm_logz_tag_rule" "test" {
  logz_monitor_id = azurerm_logz_monitor.test.id
}
`, template)
}

func (r LogzTagRuleResource) requiresImport(data acceptance.TestData, email string) string {
	config := r.basic(data, email)
	return fmt.Sprintf(`
%s

resource "azurerm_logz_tag_rule" "import" {
  logz_monitor_id = azurerm_logz_tag_rule.test.logz_monitor_id
}
`, config)
}

func (r LogzTagRuleResource) complete(data acceptance.TestData, email string) string {
	template := r.template(data, email)
	return fmt.Sprintf(`
%s

resource "azurerm_logz_tag_rule" "test" {
  logz_monitor_id = azurerm_logz_monitor.test.id
  tag_filter {
    name   = "ccc"
    action = "Include"
    value  = "ccc"
  }

  tag_filter {
    name   = "bbb"
    action = "Exclude"
    value  = ""
  }

  tag_filter {
    name   = "ccc"
    action = "Include"
    value  = "ccc"
  }
  send_aad_logs          = true
  send_activity_logs     = true
  send_subscription_logs = true
}
`, template)
}

func getEffectiveDate() string {
	effectiveDate := time.Now().Add(time.Hour * 72)
	year, month, day := effectiveDate.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, effectiveDate.Location()).Format(time.RFC3339)
}
