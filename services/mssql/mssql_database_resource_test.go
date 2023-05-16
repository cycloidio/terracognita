package mssql_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/mssql/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type MsSqlDatabaseResource struct{}

func TestAccMsSqlDatabase_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMsSqlDatabase_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

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

func TestAccMsSqlDatabase_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("collation").HasValue("SQL_AltDiction_CP850_CI_AI"),
				check.That(data.ResourceName).Key("collation").HasValue("SQL_AltDiction_CP850_CI_AI"),
				check.That(data.ResourceName).Key("license_type").HasValue("BasePrice"),
				check.That(data.ResourceName).Key("max_size_gb").HasValue("1"),
				check.That(data.ResourceName).Key("sku_name").HasValue("GP_Gen5_2"),
				check.That(data.ResourceName).Key("storage_account_type").HasValue("Local"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.ENV").HasValue("Test"),
			),
		},
		data.ImportStep("sample_name"),
		{
			Config: r.update(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("license_type").HasValue("LicenseIncluded"),
				check.That(data.ResourceName).Key("max_size_gb").HasValue("2"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.ENV").HasValue("Staging"),
			),
		},
		data.ImportStep("sample_name"),
	})
}

func TestAccMsSqlDatabase_elasticPool(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.elasticPool(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("elastic_pool_id").Exists(),
				check.That(data.ResourceName).Key("sku_name").HasValue("ElasticPool"),
			),
		},
		data.ImportStep(),
		{
			Config: r.elasticPoolDisassociation(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMsSqlDatabase_GP(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.gp(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("sku_name").HasValue("GP_Gen5_2"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMsSqlDatabase_GP_Serverless(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.gpServerless(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("auto_pause_delay_in_minutes").HasValue("70"),
				check.That(data.ResourceName).Key("min_capacity").HasValue("0.75"),
				check.That(data.ResourceName).Key("sku_name").HasValue("GP_S_Gen5_2"),
			),
		},
		data.ImportStep(),
		{
			Config: r.gpServerlessUpdate(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("auto_pause_delay_in_minutes").HasValue("90"),
				check.That(data.ResourceName).Key("min_capacity").HasValue("1.25"),
				check.That(data.ResourceName).Key("sku_name").HasValue("GP_S_Gen5_2"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMsSqlDatabase_BC(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	// Limited regional availability for BC
	data.Locations.Primary = "westeurope"

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.bc(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("read_scale").HasValue("true"),
				check.That(data.ResourceName).Key("sku_name").HasValue("BC_Gen5_2"),
				check.That(data.ResourceName).Key("zone_redundant").HasValue("true"),
			),
		},
		data.ImportStep(),
		{
			Config: r.bcUpdate(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("read_scale").HasValue("false"),
				check.That(data.ResourceName).Key("sku_name").HasValue("BC_Gen5_2"),
				check.That(data.ResourceName).Key("zone_redundant").HasValue("false"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMsSqlDatabase_HS(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.hs(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("read_replica_count").HasValue("2"),
				check.That(data.ResourceName).Key("sku_name").HasValue("HS_Gen5_2"),
			),
		},
		data.ImportStep(),
		{
			Config: r.hsUpdate(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("read_replica_count").HasValue("4"),
				check.That(data.ResourceName).Key("sku_name").HasValue("HS_Gen5_2"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMsSqlDatabase_createCopyMode(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "copy")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.createCopyMode(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("collation").HasValue("SQL_AltDiction_CP850_CI_AI"),
				check.That(data.ResourceName).Key("license_type").HasValue("BasePrice"),
				check.That(data.ResourceName).Key("sku_name").HasValue("GP_Gen5_2"),
			),
		},
		data.ImportStep("create_mode", "creation_source_database_id"),
	})
}

func TestAccMsSqlDatabase_createPITRMode(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),

		{
			PreConfig: func() { time.Sleep(11 * time.Minute) },
			Config:    r.createPITRMode(data, time.Now().Add(time.Duration(9)*time.Minute).UTC().Format(time.RFC3339)),
			Check: acceptance.ComposeTestCheckFunc(
				check.That("azurerm_mssql_database.pitr").ExistsInAzure(r),
			),
		},

		data.ImportStep("creation_source_database_id", "restore_point_in_time"),
	})
}

func TestAccMsSqlDatabase_createSecondaryMode(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "secondary")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.createSecondaryMode(data, "test1"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("collation").HasValue("SQL_AltDiction_CP850_CI_AI"),
				check.That(data.ResourceName).Key("license_type").HasValue("BasePrice"),
				check.That(data.ResourceName).Key("sku_name").HasValue("GP_Gen5_2"),
			),
		},
		data.ImportStep("sample_name"),
		{
			Config: r.createSecondaryMode(data, "test2"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("collation").HasValue("SQL_AltDiction_CP850_CI_AI"),
				check.That(data.ResourceName).Key("license_type").HasValue("BasePrice"),
				check.That(data.ResourceName).Key("sku_name").HasValue("GP_Gen5_2"),
			),
		},
		data.ImportStep("sample_name"),
	})
}

func TestAccMsSqlDatabase_createOnlineSecondaryMode(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "secondary")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.createOnlineSecondaryMode(data, "test1"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("collation").HasValue("SQL_AltDiction_CP850_CI_AI"),
				check.That(data.ResourceName).Key("license_type").HasValue("BasePrice"),
				check.That(data.ResourceName).Key("sku_name").HasValue("GP_Gen5_2"),
			),
		},
		data.ImportStep("sample_name", "create_mode"),
		{
			Config: r.createOnlineSecondaryMode(data, "test2"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("collation").HasValue("SQL_AltDiction_CP850_CI_AI"),
				check.That(data.ResourceName).Key("license_type").HasValue("BasePrice"),
				check.That(data.ResourceName).Key("sku_name").HasValue("GP_Gen5_2"),
			),
		},
		data.ImportStep("sample_name", "create_mode"),
	})
}

func TestAccMsSqlDatabase_scaleReplicaSet(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "primary")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.scaleReplicaSet(data, "GP_Gen5_2"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("sample_name"),
		{
			Config: r.scaleReplicaSet(data, "P2"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("sample_name"),
		{
			Config: r.scaleReplicaSet(data, "GP_Gen5_2"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("sample_name"),
		{
			Config: r.scaleReplicaSet(data, "BC_Gen5_2"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("sample_name"),
		{
			Config: r.scaleReplicaSet(data, "GP_Gen5_2"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("sample_name"),
		{
			Config: r.scaleReplicaSet(data, "S2"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("sample_name"),
		{
			Config: r.scaleReplicaSet(data, "Basic"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("sample_name"),
		{
			Config: r.scaleReplicaSet(data, "S1"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("sample_name"),
	})
}

func TestAccMsSqlDatabase_scaleReplicaSetWithFailovergroup(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "secondary")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.scaleReplicaSetWithFailovergroup(data, "GP_Gen5_2", 5),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("collation").HasValue("SQL_AltDiction_CP850_CI_AI"),
				check.That(data.ResourceName).Key("license_type").HasValue("BasePrice"),
				check.That(data.ResourceName).Key("sku_name").HasValue("GP_Gen5_2"),
			),
		},
		data.ImportStep(),
		{
			Config: r.scaleReplicaSetWithFailovergroup(data, "GP_Gen5_8", 25),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("collation").HasValue("SQL_AltDiction_CP850_CI_AI"),
				check.That(data.ResourceName).Key("license_type").HasValue("BasePrice"),
				check.That(data.ResourceName).Key("sku_name").HasValue("GP_Gen5_8"),
			),
		},
		data.ImportStep(),
		{
			Config: r.scaleReplicaSetWithFailovergroup(data, "GP_Gen5_2", 5),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("collation").HasValue("SQL_AltDiction_CP850_CI_AI"),
				check.That(data.ResourceName).Key("license_type").HasValue("BasePrice"),
				check.That(data.ResourceName).Key("sku_name").HasValue("GP_Gen5_2"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMsSqlDatabase_createRestoreMode(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.createRestoreMode(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("creation_source_database_id"),

		{
			PreConfig: func() { time.Sleep(8 * time.Minute) },
			Config:    r.createRestoreModeDBDeleted(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},

		data.ImportStep(),

		{
			PreConfig: func() { time.Sleep(8 * time.Minute) },
			Config:    r.createRestoreModeDBRestored(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That("azurerm_mssql_database.restore").ExistsInAzure(r),
			),
		},

		data.ImportStep("restore_dropped_database_id"),
	})
}

func TestAccMsSqlDatabase_storageAccountType(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.storageAccountTypeLocal(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("storage_account_type").HasValue("Local"),
			),
		},
		data.ImportStep("sample_name"),
	})
}

func TestAccMsSqlDatabase_threatDetectionPolicy(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.threatDetectionPolicy(data, "Enabled"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("threat_detection_policy.#").HasValue("1"),
				check.That(data.ResourceName).Key("threat_detection_policy.0.state").HasValue("Enabled"),
				check.That(data.ResourceName).Key("threat_detection_policy.0.retention_days").HasValue("15"),
				check.That(data.ResourceName).Key("threat_detection_policy.0.disabled_alerts.#").HasValue("1"),
				check.That(data.ResourceName).Key("threat_detection_policy.0.email_account_admins").HasValue("Enabled"),
			),
		},
		data.ImportStep("sample_name", "threat_detection_policy.0.storage_account_access_key"),
		{
			Config: r.threatDetectionPolicy(data, "Disabled"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("threat_detection_policy.#").HasValue("1"),
				check.That(data.ResourceName).Key("threat_detection_policy.0.state").HasValue("Disabled"),
			),
		},
		data.ImportStep("sample_name", "threat_detection_policy.0.storage_account_access_key"),
	})
}

func TestAccMsSqlDatabase_updateSku(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.updateSku(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.updateSku2(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMsSqlDatabase_minCapacity0(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.minCapacity0(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMsSqlDatabase_withLongTermRetentionPolicy(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.withLongTermRetentionPolicy(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.withLongTermRetentionPolicyUpdated(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMsSqlDatabase_withShortTermRetentionPolicy(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.withShortTermRetentionPolicy(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.withShortTermRetentionPolicyUpdated(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
	})
}

func TestAccMsSqlDatabase_geoBackupPolicy(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.withGeoBackupPoliciesDisabled(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("geo_backup_enabled").HasValue("false"),
			),
		},
		data.ImportStep(),
		{
			Config: r.withGeoBackupPoliciesEnabled(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("geo_backup_enabled").HasValue("true"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMsSqlDatabase_transitDataEncryption(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.withTransitDataEncryptionOnDwSku(data, true),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("transparent_data_encryption_enabled").HasValue("true"),
			),
		},
		data.ImportStep(),
		{
			Config: r.withTransitDataEncryptionOnDwSku(data, false),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("transparent_data_encryption_enabled").HasValue("false"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMsSqlDatabase_errorOnDisabledEncryption(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config:      r.errorOnDisabledEncryption(data),
			ExpectError: regexp.MustCompile("transparent data encryption can only be disabled on Data Warehouse SKUs"),
		},
	})
}

func TestAccMsSqlDatabase_ledgerEnabled(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mssql_database", "test")
	r := MsSqlDatabaseResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.ledgerEnabled(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (MsSqlDatabaseResource) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.DatabaseID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := client.MSSQL.DatabasesClient.Get(ctx, id.ResourceGroup, id.ServerName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return nil, fmt.Errorf("SQL Database %q (Server %q, Resource Group %q) does not exist", id.Name, id.ServerName, id.ResourceGroup)
		}

		return nil, fmt.Errorf("reading SQL Database %q (Server %q, Resource Group %q): %v", id.Name, id.ServerName, id.ResourceGroup, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (MsSqlDatabaseResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-mssql-%[1]d"
  location = "%[2]s"
}

resource "azurerm_mssql_server" "test" {
  name                         = "acctest-sqlserver-%[1]d"
  resource_group_name          = azurerm_resource_group.test.name
  location                     = azurerm_resource_group.test.location
  version                      = "12.0"
  administrator_login          = "mradministrator"
  administrator_login_password = "thisIsDog11"
}
`, data.RandomInteger, data.Locations.Primary)
}

func (r MsSqlDatabaseResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "test" {
  name      = "acctest-db-%[2]d"
  server_id = azurerm_mssql_server.test.id
}
`, r.template(data), data.RandomInteger)
}

func (r MsSqlDatabaseResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "import" {
  name      = azurerm_mssql_database.test.name
  server_id = azurerm_mssql_server.test.id
}
`, r.basic(data))
}

func (r MsSqlDatabaseResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "test" {
  name         = "acctest-db-%[2]d"
  server_id    = azurerm_mssql_server.test.id
  collation    = "SQL_AltDiction_CP850_CI_AI"
  license_type = "BasePrice"
  max_size_gb  = 1
  sample_name  = "AdventureWorksLT"
  sku_name     = "GP_Gen5_2"

  storage_account_type = "Local"

  tags = {
    ENV = "Test"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r MsSqlDatabaseResource) update(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "test" {
  name         = "acctest-db-%[2]d"
  server_id    = azurerm_mssql_server.test.id
  collation    = "SQL_AltDiction_CP850_CI_AI"
  license_type = "LicenseIncluded"
  max_size_gb  = 2
  sku_name     = "GP_Gen5_2"

  storage_account_type = "Zone"

  tags = {
    ENV = "Staging"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r MsSqlDatabaseResource) elasticPool(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_elasticpool" "test" {
  name                = "acctest-pool-%[2]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  server_name         = azurerm_mssql_server.test.name
  max_size_gb         = 5

  sku {
    name     = "GP_Gen5"
    tier     = "GeneralPurpose"
    capacity = 4
    family   = "Gen5"
  }

  per_database_settings {
    min_capacity = 0.25
    max_capacity = 4
  }
}

resource "azurerm_mssql_database" "test" {
  name            = "acctest-db-%[2]d"
  server_id       = azurerm_mssql_server.test.id
  elastic_pool_id = azurerm_mssql_elasticpool.test.id
  sku_name        = "ElasticPool"
}
`, r.template(data), data.RandomInteger)
}

func (r MsSqlDatabaseResource) elasticPoolDisassociation(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_elasticpool" "test" {
  name                = "acctest-pool-%[2]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  server_name         = azurerm_mssql_server.test.name
  max_size_gb         = 5

  sku {
    name     = "GP_Gen5"
    tier     = "GeneralPurpose"
    capacity = 4
    family   = "Gen5"
  }

  per_database_settings {
    min_capacity = 0.25
    max_capacity = 4
  }
}

resource "azurerm_mssql_database" "test" {
  name      = "acctest-db-%[2]d"
  server_id = azurerm_mssql_server.test.id
  sku_name  = "GP_Gen5_2"
}
`, r.template(data), data.RandomInteger)
}

func (r MsSqlDatabaseResource) gp(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "test" {
  name      = "acctest-db-%[2]d"
  server_id = azurerm_mssql_server.test.id
  sku_name  = "GP_Gen5_2"
}
`, r.template(data), data.RandomInteger)
}

func (r MsSqlDatabaseResource) gpServerless(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "test" {
  name                        = "acctest-db-%[2]d"
  server_id                   = azurerm_mssql_server.test.id
  auto_pause_delay_in_minutes = 70
  min_capacity                = 0.75
  sku_name                    = "GP_S_Gen5_2"
}
`, r.template(data), data.RandomInteger)
}

func (r MsSqlDatabaseResource) gpServerlessUpdate(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "test" {
  name                        = "acctest-db-%[2]d"
  server_id                   = azurerm_mssql_server.test.id
  auto_pause_delay_in_minutes = 90
  min_capacity                = 1.25
  sku_name                    = "GP_S_Gen5_2"
}
`, r.template(data), data.RandomInteger)
}

func (r MsSqlDatabaseResource) hs(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "test" {
  name               = "acctest-db-%[2]d"
  server_id          = azurerm_mssql_server.test.id
  read_replica_count = 2
  sku_name           = "HS_Gen5_2"
}
`, r.template(data), data.RandomInteger)
}

func (r MsSqlDatabaseResource) hsUpdate(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "test" {
  name               = "acctest-db-%[2]d"
  server_id          = azurerm_mssql_server.test.id
  read_replica_count = 4
  sku_name           = "HS_Gen5_2"
}
`, r.template(data), data.RandomInteger)
}

func (r MsSqlDatabaseResource) bc(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "test" {
  name           = "acctest-db-%[2]d"
  server_id      = azurerm_mssql_server.test.id
  read_scale     = true
  sku_name       = "BC_Gen5_2"
  zone_redundant = true
}
`, r.template(data), data.RandomInteger)
}

func (r MsSqlDatabaseResource) bcUpdate(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "test" {
  name           = "acctest-db-%[2]d"
  server_id      = azurerm_mssql_server.test.id
  read_scale     = false
  sku_name       = "BC_Gen5_2"
  zone_redundant = false
}
`, r.template(data), data.RandomInteger)
}

func (r MsSqlDatabaseResource) createCopyMode(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "copy" {
  name                        = "acctest-dbc-%[2]d"
  server_id                   = azurerm_mssql_server.test.id
  create_mode                 = "Copy"
  creation_source_database_id = azurerm_mssql_database.test.id
}
`, r.complete(data), data.RandomInteger)
}

func (r MsSqlDatabaseResource) createPITRMode(data acceptance.TestData, restorePointInTime string) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "pitr" {
  name                        = "acctest-dbp-%[2]d"
  server_id                   = azurerm_mssql_server.test.id
  create_mode                 = "PointInTimeRestore"
  restore_point_in_time       = "%[3]s"
  creation_source_database_id = azurerm_mssql_database.test.id

}
`, r.basic(data), data.RandomInteger, restorePointInTime)
}

func (r MsSqlDatabaseResource) createSecondaryMode(data acceptance.TestData, tag string) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_group" "second" {
  name     = "acctestRG-mssql2-%[2]d"
  location = "%[3]s"
}

resource "azurerm_mssql_server" "second" {
  name                         = "acctest-sqlserver2-%[2]d"
  resource_group_name          = azurerm_resource_group.second.name
  location                     = azurerm_resource_group.second.location
  version                      = "12.0"
  administrator_login          = "mradministrator"
  administrator_login_password = "thisIsDog11"
}

resource "azurerm_mssql_database" "secondary" {
  name                        = "acctest-dbs-%[2]d"
  server_id                   = azurerm_mssql_server.second.id
  create_mode                 = "Secondary"
  creation_source_database_id = azurerm_mssql_database.test.id

  tags = {
    tag = "%[4]s"
  }
}
`, r.complete(data), data.RandomInteger, data.Locations.Secondary, tag)
}

func (r MsSqlDatabaseResource) createOnlineSecondaryMode(data acceptance.TestData, tag string) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_group" "second" {
  name     = "acctestRG-mssql2-%[2]d"
  location = "%[3]s"
}

resource "azurerm_mssql_server" "second" {
  name                         = "acctest-sqlserver2-%[2]d"
  resource_group_name          = azurerm_resource_group.second.name
  location                     = azurerm_resource_group.second.location
  version                      = "12.0"
  administrator_login          = "mradministrator"
  administrator_login_password = "thisIsDog11"
}

resource "azurerm_mssql_database" "secondary" {
  name                        = "acctest-dbs-%[2]d"
  server_id                   = azurerm_mssql_server.second.id
  create_mode                 = "OnlineSecondary"
  creation_source_database_id = azurerm_mssql_database.test.id

  tags = {
    tag = "%[4]s"
  }
}
`, r.complete(data), data.RandomInteger, data.Locations.Secondary, tag)
}

func (r MsSqlDatabaseResource) scaleReplicaSet(data acceptance.TestData, sku string) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "primary" {
  name        = "acctest-db-%[2]d"
  server_id   = azurerm_mssql_server.test.id
  sample_name = "AdventureWorksLT"

  max_size_gb = "2"
  sku_name    = "%[4]s"
}

resource "azurerm_resource_group" "secondary" {
  name     = "acctestRG-mssql2-%[2]d"
  location = "%[3]s"
}

resource "azurerm_mssql_server" "secondary" {
  name                         = "acctest-sqlserver2-%[2]d"
  resource_group_name          = azurerm_resource_group.secondary.name
  location                     = azurerm_resource_group.secondary.location
  version                      = "12.0"
  administrator_login          = "mradministrator"
  administrator_login_password = "thisIsDog12"
}

resource "azurerm_mssql_database" "secondary" {
  name                        = "acctest-db-%[2]d"
  server_id                   = azurerm_mssql_server.secondary.id
  create_mode                 = "Secondary"
  creation_source_database_id = azurerm_mssql_database.primary.id

  sku_name = "%[4]s"
}
`, r.template(data), data.RandomInteger, data.Locations.Secondary, sku)
}

func (r MsSqlDatabaseResource) scaleReplicaSetWithFailovergroup(data acceptance.TestData, sku string, size int) string {
	return fmt.Sprintf(`
	%[1]s

resource "azurerm_mssql_database" "test" {
  name         = "acctest-db-%[2]d"
  server_id    = azurerm_mssql_server.test.id
  collation    = "SQL_AltDiction_CP850_CI_AI"
  license_type = "BasePrice"
  max_size_gb  = %[5]d
  sample_name  = "AdventureWorksLT"
  sku_name     = "%[4]s"

  tags = {
    ENV = "Test"
  }
}

resource "azurerm_resource_group" "second" {
  name     = "acctestRG-mssql2-%[2]d"
  location = "%[3]s"
}

resource "azurerm_mssql_server" "second" {
  name                         = "acctest-sqlserver2-%[2]d"
  resource_group_name          = azurerm_resource_group.second.name
  location                     = azurerm_resource_group.second.location
  version                      = "12.0"
  administrator_login          = "mradministrator"
  administrator_login_password = "thisIsDog11"
}

resource "azurerm_mssql_database" "secondary" {
  name                        = "acctest-db-%[2]d"
  server_id                   = azurerm_mssql_server.second.id
  create_mode                 = "Secondary"
  creation_source_database_id = azurerm_mssql_database.test.id
  sku_name                    = "%[4]s"
}

resource "azurerm_mssql_failover_group" "failover_group" {
  name      = "acctest-fog-%[2]d"
  server_id = azurerm_mssql_server.test.id
  databases = [azurerm_mssql_database.test.id]

  partner_server {
    id = azurerm_mssql_server.second.id
  }

  read_write_endpoint_failover_policy {
    mode          = "Automatic"
    grace_minutes = 60
  }

  depends_on = [
    azurerm_mssql_database.test,
    azurerm_mssql_database.secondary
  ]
}
`, r.template(data), data.RandomInteger, data.Locations.Secondary, sku, size)
}

func (MsSqlDatabaseResource) createRestoreMode(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-mssql-%[1]d"
  location = "%[2]s"
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

resource "azurerm_mssql_database" "copy" {
  name                        = "acctest-dbc-%[1]d"
  server_id                   = azurerm_mssql_server.test.id
  create_mode                 = "Copy"
  creation_source_database_id = azurerm_mssql_database.test.id
}
`, data.RandomInteger, data.Locations.Primary)
}

func (MsSqlDatabaseResource) createRestoreModeDBDeleted(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-mssql-%[1]d"
  location = "%[2]s"
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
`, data.RandomInteger, data.Locations.Primary)
}

func (MsSqlDatabaseResource) createRestoreModeDBRestored(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-mssql-%[1]d"
  location = "%[2]s"
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

resource "azurerm_mssql_database" "restore" {
  name                        = "acctest-dbr-%[1]d"
  server_id                   = azurerm_mssql_server.test.id
  create_mode                 = "Restore"
  restore_dropped_database_id = azurerm_mssql_server.test.restorable_dropped_database_ids[0]
}
`, data.RandomInteger, data.Locations.Primary)
}

func (r MsSqlDatabaseResource) storageAccountTypeLocal(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "test" {
  name      = "acctest-db-%[2]d"
  server_id = azurerm_mssql_server.test.id

  storage_account_type = "Local"
}
`, r.template(data), data.RandomInteger)
}

func (r MsSqlDatabaseResource) threatDetectionPolicy(data acceptance.TestData, state string) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_storage_account" "test" {
  name                     = "test%[2]d"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "GRS"
}

resource "azurerm_mssql_database" "test" {
  name         = "acctest-db-%[2]d"
  server_id    = azurerm_mssql_server.test.id
  collation    = "SQL_AltDiction_CP850_CI_AI"
  license_type = "BasePrice"
  max_size_gb  = 1
  sample_name  = "AdventureWorksLT"
  sku_name     = "GP_Gen5_2"

  threat_detection_policy {
    retention_days             = 15
    state                      = "%[3]s"
    disabled_alerts            = ["Sql_Injection"]
    email_account_admins       = "Enabled"
    storage_account_access_key = azurerm_storage_account.test.primary_access_key
    storage_endpoint           = azurerm_storage_account.test.primary_blob_endpoint
  }

  tags = {
    ENV = "Test"
  }
}
`, r.template(data), data.RandomInteger, state)
}

func (r MsSqlDatabaseResource) updateSku(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "test" {
  name      = "acctest-db-%[2]d"
  server_id = azurerm_mssql_server.test.id
  sku_name  = "HS_Gen5_2"
}
`, r.template(data), data.RandomInteger)
}

func (r MsSqlDatabaseResource) updateSku2(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "test" {
  name      = "acctest-db-%[2]d"
  server_id = azurerm_mssql_server.test.id
  sku_name  = "HS_Gen5_4"
}
`, r.template(data), data.RandomInteger)
}

func (r MsSqlDatabaseResource) minCapacity0(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "test" {
  name      = "acctest-db-%[2]d"
  server_id = azurerm_mssql_server.test.id

  min_capacity = 0
}
`, r.template(data), data.RandomInteger)
}

func (r MsSqlDatabaseResource) withLongTermRetentionPolicy(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_storage_account" "test" {
  name                     = "acctest%[2]d"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_storage_account" "test2" {
  name                     = "acctest2%[2]d"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_mssql_database" "test" {
  name      = "acctest-db-%[3]d"
  server_id = azurerm_mssql_server.test.id
  long_term_retention_policy {
    weekly_retention  = "P1W"
    monthly_retention = "P1M"
    yearly_retention  = "P1Y"
    week_of_year      = 1
  }
}
`, r.template(data), data.RandomIntOfLength(15), data.RandomInteger)
}

func (r MsSqlDatabaseResource) withLongTermRetentionPolicyUpdated(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_storage_account" "test" {
  name                     = "acctest%[2]d"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_storage_account" "test2" {
  name                     = "acctest2%[2]d"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_mssql_database" "test" {
  name      = "acctest-db-%[3]d"
  server_id = azurerm_mssql_server.test.id
  long_term_retention_policy {
    weekly_retention = "P1W"
    yearly_retention = "P1Y"
    week_of_year     = 2
  }
}
`, r.template(data), data.RandomIntOfLength(15), data.RandomInteger)
}

func (r MsSqlDatabaseResource) withShortTermRetentionPolicy(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_storage_account" "test" {
  name                     = "acctest%[2]d"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_storage_account" "test2" {
  name                     = "acctest2%[2]d"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_mssql_database" "test" {
  name      = "acctest-db-%[3]d"
  server_id = azurerm_mssql_server.test.id
  short_term_retention_policy {
    retention_days           = 8
    backup_interval_in_hours = 12
  }
}
`, r.template(data), data.RandomIntOfLength(15), data.RandomInteger)
}

func (r MsSqlDatabaseResource) withShortTermRetentionPolicyUpdated(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_storage_account" "test" {
  name                     = "acctest%[2]d"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_storage_account" "test2" {
  name                     = "acctest2%[2]d"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_mssql_database" "test" {
  name      = "acctest-db-%[3]d"
  server_id = azurerm_mssql_server.test.id
  short_term_retention_policy {
    retention_days           = 10
    backup_interval_in_hours = 24
  }
}
`, r.template(data), data.RandomIntOfLength(15), data.RandomInteger)
}

func (r MsSqlDatabaseResource) withGeoBackupPoliciesEnabled(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "test" {
  name               = "acctest-db-%[3]d"
  server_id          = azurerm_mssql_server.test.id
  sku_name           = "DW100c"
  geo_backup_enabled = true
}
`, r.template(data), data.RandomIntOfLength(15), data.RandomInteger)
}

func (r MsSqlDatabaseResource) withGeoBackupPoliciesDisabled(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "test" {
  name               = "acctest-db-%[3]d"
  server_id          = azurerm_mssql_server.test.id
  sku_name           = "DW100c"
  geo_backup_enabled = false
}
`, r.template(data), data.RandomIntOfLength(15), data.RandomInteger)
}

func (r MsSqlDatabaseResource) withTransitDataEncryptionOnDwSku(data acceptance.TestData, state bool) string {
	return fmt.Sprintf(`
%s

resource "azurerm_mssql_database" "test" {
  name                                = "acctest-db-%d"
  server_id                           = azurerm_mssql_server.test.id
  sku_name                            = "DW100c"
  transparent_data_encryption_enabled = %t
}
`, r.template(data), data.RandomInteger, state)
}

func (r MsSqlDatabaseResource) errorOnDisabledEncryption(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_mssql_database" "test" {
  name                                = "acctest-db-%d"
  server_id                           = azurerm_mssql_server.test.id
  transparent_data_encryption_enabled = false
}
`, r.template(data), data.RandomInteger)
}

func (r MsSqlDatabaseResource) ledgerEnabled(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_mssql_database" "test" {
  name           = "acctest-db-%[2]d"
  server_id      = azurerm_mssql_server.test.id
  ledger_enabled = true
}
`, r.template(data), data.RandomInteger)
}
