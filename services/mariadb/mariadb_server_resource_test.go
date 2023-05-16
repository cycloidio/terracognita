package mariadb_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/mariadb/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type MariaDbServerResource struct{}

func TestAccMariaDbServer_basicTenTwo(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mariadb_server", "test")
	r := MariaDbServerResource{}
	version := "10.2"

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data, version),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("version").HasValue(version),
			),
		},
		data.ImportStep("administrator_login_password"), // not returned as sensitive
	})
}

func TestAccMariaDbServer_basicTenThree(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mariadb_server", "test")
	r := MariaDbServerResource{}
	version := "10.3"

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data, version),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("version").HasValue(version),
			),
		},
		data.ImportStep("administrator_login_password"), // not returned as sensitive
	})
}

func TestAccMariaDbServer_autogrowOnly(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mariadb_server", "test")
	r := MariaDbServerResource{}
	version := "10.3"

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.autogrow(data, version),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("administrator_login_password"), // not returned as sensitive
	})
}

func TestAccMariaDbServer_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mariadb_server", "test")
	r := MariaDbServerResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data, "10.3"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func TestAccMariaDbServer_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mariadb_server", "test")
	r := MariaDbServerResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data, "10.3"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("administrator_login_password"), // not returned as sensitive
	})
}

func TestAccMariaDbServer_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mariadb_server", "test")
	r := MariaDbServerResource{}
	version := "10.3"

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data, version),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("administrator_login_password"), // not returned as sensitive
		{
			Config: r.complete(data, version),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("administrator_login_password"), // not returned as sensitive
		{
			Config: r.basic(data, version),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("administrator_login_password"), // not returned as sensitive
	})
}

func TestAccMariaDbServer_updateSKU(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mariadb_server", "test")
	r := MariaDbServerResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.sku(data, "GP_Gen5_32"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("administrator_login_password"), // not returned as sensitive
		{
			Config: r.sku(data, "MO_Gen5_16"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("administrator_login_password"), // not returned as sensitive
	})
}

func TestAccMariaDbServer_createReplica(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mariadb_server", "test")
	r := MariaDbServerResource{}
	version := "10.3"

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data, version),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("administrator_login_password"), // not returned as sensitive
		{
			Config: r.createReplica(data, version),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That("azurerm_mariadb_server.replica").ExistsInAzure(r),
			),
		},
		data.ImportStep("administrator_login_password"), // not returned as sensitive
	})
}

func TestAccMariaDbServer_createPointInTimeRestore(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mariadb_server", "test")
	r := MariaDbServerResource{}
	restoreTime := time.Now().Add(11 * time.Minute)
	version := "10.3"

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data, version),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("administrator_login_password"), // not returned as sensitive
		{
			PreConfig: func() { time.Sleep(restoreTime.Sub(time.Now().Add(-7 * time.Minute))) },
			Config:    r.createPointInTimeRestore(data, version, restoreTime.Format(time.RFC3339)),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That("azurerm_mariadb_server.restore").ExistsInAzure(r),
			),
		},
		data.ImportStep("administrator_login_password"), // not returned as sensitive
	})
}

func (MariaDbServerResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.ServerID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.MariaDB.ServersClient.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving MariaDB Server %q (Resource Group %q): %v", id.Name, id.ResourceGroup, err)
	}

	return utils.Bool(resp.ServerProperties != nil), nil
}

func (MariaDbServerResource) basic(data acceptance.TestData, version string) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_mariadb_server" "test" {
  name                = "acctestmariadbsvr-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku_name            = "B_Gen5_2"
  version             = "%s"

  administrator_login          = "acctestun"
  administrator_login_password = "H@Sh1CoR3!"
  ssl_enforcement_enabled      = true
  storage_mb                   = 51200
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, version)
}

func (MariaDbServerResource) complete(data acceptance.TestData, version string) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_mariadb_server" "test" {
  name                = "acctestmariadbsvr-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku_name            = "B_Gen5_2"
  version             = "%s"

  administrator_login           = "acctestun"
  administrator_login_password  = "H@Sh1CoR3!"
  auto_grow_enabled             = true
  backup_retention_days         = 14
  create_mode                   = "Default"
  geo_redundant_backup_enabled  = false
  ssl_enforcement_enabled       = true
  public_network_access_enabled = true
  storage_mb                    = 51200
  tags = {
    environment = "test"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, version)
}

func (MariaDbServerResource) autogrow(data acceptance.TestData, version string) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_mariadb_server" "test" {
  name                = "acctestmariadbsvr-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku_name            = "B_Gen5_2"
  version             = "%s"

  administrator_login          = "acctestun"
  administrator_login_password = "H@Sh1CoR3!"
  auto_grow_enabled            = true
  backup_retention_days        = 7
  geo_redundant_backup_enabled = false
  ssl_enforcement_enabled      = true
  storage_mb                   = 51200
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, version)
}

func (r MariaDbServerResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_mariadb_server" "import" {
  name                = azurerm_mariadb_server.test.name
  location            = azurerm_mariadb_server.test.location
  resource_group_name = azurerm_mariadb_server.test.resource_group_name
  sku_name            = "B_Gen5_2"
  version             = "10.3"

  administrator_login          = "acctestun"
  administrator_login_password = "H@Sh1CoR3!"
  backup_retention_days        = 7
  geo_redundant_backup_enabled = false
  ssl_enforcement_enabled      = true
  storage_mb                   = 51200
}
`, r.basic(data, "10.3"))
}

func (MariaDbServerResource) sku(data acceptance.TestData, sku string) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_mariadb_server" "test" {
  name                = "acctestmariadbsvr-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku_name            = "%s"
  version             = "10.2"

  administrator_login          = "acctestun"
  administrator_login_password = "H@Sh1CoR3!"
  backup_retention_days        = 7
  geo_redundant_backup_enabled = false
  ssl_enforcement_enabled      = true
  storage_mb                   = 640000
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, sku)
}

func (r MariaDbServerResource) createReplica(data acceptance.TestData, version string) string {
	return fmt.Sprintf(`
%s

resource "azurerm_mariadb_server" "replica" {
  name                      = "acctestmariadbsvr-%d-replica"
  location                  = azurerm_resource_group.test.location
  resource_group_name       = azurerm_resource_group.test.name
  sku_name                  = "B_Gen5_2"
  version                   = "%s"
  create_mode               = "Replica"
  creation_source_server_id = azurerm_mariadb_server.test.id
  ssl_enforcement_enabled   = true
  storage_mb                = 51200
}
`, r.basic(data, version), data.RandomInteger, version)
}

func (r MariaDbServerResource) createPointInTimeRestore(data acceptance.TestData, version, restoreTime string) string {
	return fmt.Sprintf(`
%s

resource "azurerm_mariadb_server" "restore" {
  name                      = "acctestmariadbsvr-%d-restore"
  location                  = azurerm_resource_group.test.location
  resource_group_name       = azurerm_resource_group.test.name
  sku_name                  = "B_Gen5_2"
  version                   = "%s"
  create_mode               = "PointInTimeRestore"
  creation_source_server_id = azurerm_mariadb_server.test.id
  restore_point_in_time     = "%s"
  ssl_enforcement_enabled   = true
  storage_mb                = 51200
}
`, r.basic(data, version), data.RandomInteger, version, restoreTime)
}
