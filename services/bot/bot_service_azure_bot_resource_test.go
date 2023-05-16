package bot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/bot/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type BotServiceAzureBotResource struct{}

func testAccBotServiceAzureBot_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_bot_service_azure_bot", "test")
	r := BotServiceAzureBotResource{}

	data.ResourceSequentialTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("sku").HasValue("F0"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.environment").HasValue("test"),
			),
		},
		data.ImportStep(),
	})
}

func testAccBotServiceAzureBot_completeUpdate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_bot_service_azure_bot", "test")
	r := BotServiceAzureBotResource{}

	data.ResourceSequentialTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.update(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func testAccBotServiceAzureBot_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_bot_service_azure_bot", "test")
	r := BotServiceAzureBotResource{}

	data.ResourceSequentialTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_bot_service_azure_bot"),
		},
	})
}

func (t BotServiceAzureBotResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.BotServiceID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Bot.BotClient.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving %s: %v", *id, err)
	}

	return utils.Bool(resp.Properties != nil), nil
}

func (BotServiceAzureBotResource) basic(data acceptance.TestData) string {
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

resource "azurerm_bot_service_azure_bot" "test" {
  name                = "acctestdf%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = "global"
  sku                 = "F0"
  microsoft_app_id    = data.azurerm_client_config.current.client_id

  tags = {
    environment = "test"
  }
}
`, data.RandomInteger, data.Locations.Primary)
}

func (BotServiceAzureBotResource) update(data acceptance.TestData) string {
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

resource "azurerm_application_insights" "test" {
  name                = "acctestappinsights-%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  application_type    = "web"
}

resource "azurerm_application_insights_api_key" "test" {
  name                    = "acctestappinsightsapikey-%[1]d"
  application_insights_id = azurerm_application_insights.test.id
  read_permissions        = ["aggregate", "api", "draft", "extendqueries", "search"]
}

resource "azurerm_bot_service_azure_bot" "test" {
  name                = "acctestdf%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = "global"
  microsoft_app_id    = data.azurerm_client_config.current.client_id
  sku                 = "F0"

  endpoint                              = "https://example.com"
  developer_app_insights_api_key        = azurerm_application_insights_api_key.test.api_key
  developer_app_insights_application_id = azurerm_application_insights.test.app_id

  tags = {
    environment = "test"
  }
}
`, data.RandomInteger, data.Locations.Primary)
}

func (BotServiceAzureBotResource) requiresImport(data acceptance.TestData) string {
	template := BotServiceAzureBotResource{}.basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_bot_service_azure_bot" "import" {
  name                = azurerm_bot_service_azure_bot.test.name
  resource_group_name = azurerm_bot_service_azure_bot.test.resource_group_name
  location            = azurerm_bot_service_azure_bot.test.location
  sku                 = azurerm_bot_service_azure_bot.test.sku
  microsoft_app_id    = azurerm_bot_service_azure_bot.test.microsoft_app_id
}
`, template)
}
