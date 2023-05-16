package mssql

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/sql/mgmt/v5.0/sql"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/mssql/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/mssql/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func dataSourceMsSqlDatabase() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceMsSqlDatabaseRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.ValidateMsSqlDatabaseName,
			},

			"server_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.ServerID,
			},

			"collation": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"elastic_pool_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"license_type": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"max_size_gb": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"read_replica_count": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"read_scale": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"sku_name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"storage_account_type": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"zone_redundant": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"tags": tags.SchemaDataSource(),
		},
	}
}

func dataSourceMsSqlDatabaseRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).MSSQL.DatabasesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	mssqlServerId := d.Get("server_id").(string)
	serverId, err := parse.ServerID(mssqlServerId)
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, serverId.ResourceGroup, serverId.Name, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Database %q (Resource Group %q, SQL Server %q) was not found", name, serverId.ResourceGroup, serverId.Name)
		}

		return fmt.Errorf("making Read request on AzureRM Database %s (Resource Group %q, SQL Server %q): %+v", name, serverId.ResourceGroup, serverId.Name, err)
	}

	d.SetId(parse.NewDatabaseID(serverId.SubscriptionId, serverId.ResourceGroup, serverId.Name, name).ID())
	d.Set("name", name)
	d.Set("server_id", mssqlServerId)

	if props := resp.DatabaseProperties; props != nil {
		d.Set("collation", props.Collation)
		d.Set("elastic_pool_id", props.ElasticPoolID)
		d.Set("license_type", props.LicenseType)
		if props.MaxSizeBytes != nil {
			d.Set("max_size_gb", int32((*props.MaxSizeBytes)/int64(1073741824)))
		}
		d.Set("read_replica_count", props.HighAvailabilityReplicaCount)
		if props.ReadScale == sql.DatabaseReadScaleEnabled {
			d.Set("read_scale", true)
		} else if props.ReadScale == sql.DatabaseReadScaleDisabled {
			d.Set("read_scale", false)
		}
		d.Set("sku_name", props.CurrentServiceObjectiveName)
		d.Set("storage_account_type", sql.RequestedBackupStorageRedundancy(props.CurrentBackupStorageRedundancy))
		d.Set("zone_redundant", props.ZoneRedundant)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}
