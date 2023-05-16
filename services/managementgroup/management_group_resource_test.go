package managementgroup_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/managementgroup/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type ManagementGroupResource struct{}

func TestAccManagementGroup_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_management_group", "test")
	r := ManagementGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccManagementGroup_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_management_group", "test")
	r := ManagementGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(),
			ExpectError: acceptance.RequiresImportError("azurerm_management_group"),
		},
	})
}

func TestAccManagementGroup_nested(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_management_group", "parent")
	r := ManagementGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.nested(),
			Check: acceptance.ComposeTestCheckFunc(
				check.That("azurerm_management_group.parent").ExistsInAzure(r),
				check.That("azurerm_management_group.child").ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		data.ImportStepFor("azurerm_management_group.child"),
	})
}

func TestAccManagementGroup_multiLevel(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_management_group", "parent")
	r := ManagementGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.multiLevel(),
			Check: acceptance.ComposeTestCheckFunc(
				check.That("azurerm_management_group.grandparent").ExistsInAzure(r),
				check.That("azurerm_management_group.parent").ExistsInAzure(r),
				check.That("azurerm_management_group.child").ExistsInAzure(r),
			),
		},
		data.ImportStepFor("azurerm_management_group.grandparent"),
		data.ImportStep(),
		data.ImportStepFor("azurerm_management_group.child"),
	})
}

func TestAccManagementGroup_multiLevelUpdated(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_management_group", "parent")
	r := ManagementGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.nested(),
			Check: acceptance.ComposeTestCheckFunc(
				check.That("azurerm_management_group.parent").ExistsInAzure(r),
				check.That("azurerm_management_group.child").ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		data.ImportStepFor("azurerm_management_group.child"),
		{
			Config: r.multiLevel(),
			Check: acceptance.ComposeTestCheckFunc(
				check.That("azurerm_management_group.grandparent").ExistsInAzure(r),
				check.That("azurerm_management_group.parent").ExistsInAzure(r),
				check.That("azurerm_management_group.child").ExistsInAzure(r),
			),
		},
		data.ImportStepFor("azurerm_management_group.grandparent"),
		data.ImportStep(),
		data.ImportStepFor("azurerm_management_group.child"),
	})
}

func TestAccManagementGroup_withName(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_management_group", "test")
	r := ManagementGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.withName(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccManagementGroup_updateName(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_management_group", "test")
	r := ManagementGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config: r.withName(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("name").HasValue(fmt.Sprintf("acctestmg-%d", data.RandomInteger)),
			),
		},
		data.ImportStep(),
	})
}

func TestAccManagementGroup_withSubscriptions(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_management_group", "test")
	r := ManagementGroupResource{}
	subscriptionID := os.Getenv("ARM_SUBSCRIPTION_ID")

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("subscription_ids.#").HasValue("0"),
			),
		},
		{
			Config: r.withSubscriptions(subscriptionID),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("subscription_ids.#").HasValue("1"),
			),
		},
		{
			Config: r.removeSubscriptions(),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("subscription_ids.#").HasValue("0"),
			),
		},
	})
}

func (ManagementGroupResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.ManagementGroupID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.ManagementGroups.GroupsClient.Get(ctx, id.Name, "children", utils.Bool(true), "", "no-cache")
	if err != nil {
		return nil, fmt.Errorf("retrieving Management Group %s: %v", id.Name, err)
	}

	return utils.Bool(resp.Properties != nil), nil
}

func (r ManagementGroupResource) basic() string {
	return `
provider "azurerm" {
  features {}
}

resource "azurerm_management_group" "test" {
}
`
}

func (r ManagementGroupResource) requiresImport() string {
	return `
provider "azurerm" {
  features {}
}

resource "azurerm_management_group" "test" {
}

resource "azurerm_management_group" "import" {
  name = azurerm_management_group.test.name
}
`
}

func (r ManagementGroupResource) nested() string {
	return `
provider "azurerm" {
  features {}
}

resource "azurerm_management_group" "parent" {
}

resource "azurerm_management_group" "child" {
  parent_management_group_id = azurerm_management_group.parent.id
}
`
}

func (r ManagementGroupResource) multiLevel() string {
	return `
provider "azurerm" {
  features {}
}

resource "azurerm_management_group" "grandparent" {
}

resource "azurerm_management_group" "parent" {
  parent_management_group_id = azurerm_management_group.grandparent.id
}

resource "azurerm_management_group" "child" {
  parent_management_group_id = azurerm_management_group.parent.id
}
`
}

func (ManagementGroupResource) withName(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_management_group" "test" {
  name         = "acctestmg-%d"
  display_name = "accTestMG-%d"
}
`, data.RandomInteger, data.RandomInteger)
}

// TODO: switch this out for dynamically creating a subscription once that's supported in the future
func (r ManagementGroupResource) withSubscriptions(subscriptionID string) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_management_group" "test" {
  subscription_ids = [
    "%s",
  ]
}
`, subscriptionID)
}

func (r ManagementGroupResource) removeSubscriptions() string {
	return `
provider "azurerm" {
  features {}
}

resource "azurerm_management_group" "test" {
  subscription_ids = []
}
`
}
