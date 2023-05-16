package mariadb

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/mariadb/mgmt/2018-06-01/mariadb"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/mariadb/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/mariadb/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceMariaDbServer() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceMariaDbServerCreate,
		Read:   resourceMariaDbServerRead,
		Update: resourceMariaDbServerUpdate,
		Delete: resourceMariaDbServerDelete,

		Importer: pluginsdk.ImporterValidatingResourceIdThen(func(id string) error {
			_, err := parse.ServerID(id)
			return err
		}, func(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) ([]*pluginsdk.ResourceData, error) {
			d.Set("create_mode", "Default")
			if v, ok := d.GetOk("create_mode"); ok && v.(string) != "" {
				d.Set("create_mode", v)
			}

			return []*pluginsdk.ResourceData{d}, nil
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(60 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(60 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ServerName,
			},

			"administrator_login": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"administrator_login_password": {
				Type:      pluginsdk.TypeString,
				Optional:  true,
				Sensitive: true,
			},

			"auto_grow_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"backup_retention_days": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(7, 35),
			},

			"create_mode": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  string(mariadb.CreateModeDefault),
				ValidateFunc: validation.StringInSlice([]string{
					string(mariadb.CreateModeDefault),
					string(mariadb.CreateModeGeoRestore),
					string(mariadb.CreateModePointInTimeRestore),
					string(mariadb.CreateModeReplica),
				}, false),
			},

			"creation_source_server_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validate.ServerID,
			},

			"fqdn": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"geo_redundant_backup_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Computed: true,
			},

			"location": azure.SchemaLocation(),

			"public_network_access_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"restore_point_in_time": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},

			"sku_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"B_Gen5_1",
					"B_Gen5_2",
					"GP_Gen5_2",
					"GP_Gen5_4",
					"GP_Gen5_8",
					"GP_Gen5_16",
					"GP_Gen5_32",
					"MO_Gen5_2",
					"MO_Gen5_4",
					"MO_Gen5_8",
					"MO_Gen5_16",
				}, false),
			},

			"ssl_enforcement_enabled": {
				Type:     pluginsdk.TypeBool,
				Required: true,
			},

			"storage_mb": {
				Type:     pluginsdk.TypeInt,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.IntBetween(5120, 4194304),
					validation.IntDivisibleBy(1024),
				),
			},

			"tags": tags.Schema(),

			"version": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"10.2",
					"10.3",
				}, false),
			},
		},
	}
}

func resourceMariaDbServerCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).MariaDB.ServersClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewServerID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_mariadb_server", id.ID())
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	mode := mariadb.CreateMode(d.Get("create_mode").(string))
	source := d.Get("creation_source_server_id").(string)
	version := mariadb.ServerVersion(d.Get("version").(string))

	sku, err := expandServerSkuName(d.Get("sku_name").(string))
	if err != nil {
		return fmt.Errorf("expanding `sku_name`: %+v", err)
	}

	publicAccess := mariadb.PublicNetworkAccessEnumEnabled
	if v := d.Get("public_network_access_enabled"); !v.(bool) {
		publicAccess = mariadb.PublicNetworkAccessEnumDisabled
	}

	ssl := mariadb.SslEnforcementEnumEnabled
	if v := d.Get("ssl_enforcement_enabled").(bool); !v {
		ssl = mariadb.SslEnforcementEnumDisabled
	}

	storage := expandMariaDbStorageProfile(d)

	var props mariadb.BasicServerPropertiesForCreate
	switch mode {
	case mariadb.CreateModeDefault:
		admin := d.Get("administrator_login").(string)
		pass := d.Get("administrator_login_password").(string)

		if admin == "" {
			return fmt.Errorf("`administrator_login` must not be empty when `create_mode` is `default`")
		}
		if pass == "" {
			return fmt.Errorf("`administrator_login_password` must not be empty when `create_mode` is `default`")
		}

		if _, ok := d.GetOk("restore_point_in_time"); ok {
			return fmt.Errorf("`restore_point_in_time` cannot be set when `create_mode` is `default`")
		}

		props = &mariadb.ServerPropertiesForDefaultCreate{
			AdministratorLogin:         &admin,
			AdministratorLoginPassword: &pass,
			CreateMode:                 mode,
			PublicNetworkAccess:        publicAccess,
			SslEnforcement:             ssl,
			StorageProfile:             storage,
			Version:                    version,
		}
	case mariadb.CreateModePointInTimeRestore:
		v, ok := d.GetOk("restore_point_in_time")
		if !ok || v.(string) == "" {
			return fmt.Errorf("restore_point_in_time must be set when create_mode is PointInTimeRestore")
		}
		time, _ := time.Parse(time.RFC3339, v.(string)) // should be validated by the schema

		props = &mariadb.ServerPropertiesForRestore{
			CreateMode:     mode,
			SourceServerID: &source,
			RestorePointInTime: &date.Time{
				Time: time,
			},
			PublicNetworkAccess: publicAccess,
			SslEnforcement:      ssl,
			StorageProfile:      storage,
			Version:             version,
		}
	case mariadb.CreateModeGeoRestore:
		props = &mariadb.ServerPropertiesForGeoRestore{
			CreateMode:          mode,
			SourceServerID:      &source,
			PublicNetworkAccess: publicAccess,
			SslEnforcement:      ssl,
			StorageProfile:      storage,
			Version:             version,
		}
	case mariadb.CreateModeReplica:
		props = &mariadb.ServerPropertiesForReplica{
			CreateMode:          mode,
			SourceServerID:      &source,
			PublicNetworkAccess: publicAccess,
			SslEnforcement:      ssl,
			Version:             version,
		}
	}

	server := mariadb.ServerForCreate{
		Location:   &location,
		Properties: props,
		Sku:        sku,
		Tags:       tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	future, err := client.Create(ctx, id.ResourceGroup, id.Name, server)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the creation of %s: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceMariaDbServerRead(d, meta)
}

func resourceMariaDbServerUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).MariaDB.ServersClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for AzureRM MariaDB Server update.")

	id, err := parse.ServerID(d.Id())
	if err != nil {
		return err
	}

	sku, err := expandServerSkuName(d.Get("sku_name").(string))
	if err != nil {
		return fmt.Errorf("expanding `sku_name`: %+v", err)
	}

	publicAccess := mariadb.PublicNetworkAccessEnumEnabled
	if v := d.Get("public_network_access_enabled").(bool); !v {
		publicAccess = mariadb.PublicNetworkAccessEnumDisabled
	}

	ssl := mariadb.SslEnforcementEnumEnabled
	if v := d.Get("ssl_enforcement_enabled").(bool); !v {
		ssl = mariadb.SslEnforcementEnumDisabled
	}

	storageProfile := expandMariaDbStorageProfile(d)

	properties := mariadb.ServerUpdateParameters{
		ServerUpdateParametersProperties: &mariadb.ServerUpdateParametersProperties{
			AdministratorLoginPassword: utils.String(d.Get("administrator_login_password").(string)),
			PublicNetworkAccess:        publicAccess,
			SslEnforcement:             ssl,
			StorageProfile:             storageProfile,
			Version:                    mariadb.ServerVersion(d.Get("version").(string)),
		},
		Sku:  sku,
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	future, err := client.Update(ctx, id.ResourceGroup, id.Name, properties)
	if err != nil {
		return fmt.Errorf("updating %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for update of %s: %+v", *id, err)
	}

	return resourceMariaDbServerRead(d, meta)
}

func resourceMariaDbServerRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).MariaDB.ServersClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ServerID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[WARN] %s was not found - removing from state", *id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)

	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if sku := resp.Sku; sku != nil {
		d.Set("sku_name", sku.Name)
	}

	if props := resp.ServerProperties; props != nil {
		d.Set("administrator_login", props.AdministratorLogin)
		d.Set("public_network_access_enabled", props.PublicNetworkAccess == mariadb.PublicNetworkAccessEnumEnabled)
		d.Set("ssl_enforcement_enabled", props.SslEnforcement == mariadb.SslEnforcementEnumEnabled)
		d.Set("version", string(props.Version))

		if storage := props.StorageProfile; storage != nil {
			d.Set("auto_grow_enabled", storage.StorageAutogrow == mariadb.StorageAutogrowEnabled)
			d.Set("backup_retention_days", storage.BackupRetentionDays)
			d.Set("geo_redundant_backup_enabled", storage.GeoRedundantBackup == mariadb.Enabled)
			d.Set("storage_mb", storage.StorageMB)
		}

		// Computed
		d.Set("fqdn", props.FullyQualifiedDomainName)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceMariaDbServerDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).MariaDB.ServersClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ServerID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the deletion of %s: %+v", *id, err)
	}

	return nil
}

func expandServerSkuName(skuName string) (*mariadb.Sku, error) {
	parts := strings.Split(skuName, "_")
	if len(parts) != 3 {
		return nil, fmt.Errorf("sku_name (%s) has the wrong number of parts (%d) after splitting on _", skuName, len(parts))
	}

	var tier mariadb.SkuTier
	switch parts[0] {
	case "B":
		tier = mariadb.Basic
	case "GP":
		tier = mariadb.GeneralPurpose
	case "MO":
		tier = mariadb.MemoryOptimized
	default:
		return nil, fmt.Errorf("sku_name %s has unknown sku tier %s", skuName, parts[0])
	}

	capacity, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("cannot convert `sku_name` %q capacity %s to int", skuName, parts[2])
	}

	return &mariadb.Sku{
		Name:     utils.String(skuName),
		Tier:     tier,
		Capacity: utils.Int32(int32(capacity)),
		Family:   utils.String(parts[1]),
	}, nil
}

func expandMariaDbStorageProfile(d *pluginsdk.ResourceData) *mariadb.StorageProfile {
	storage := mariadb.StorageProfile{}
	// now override whatever we may have from the block with the top level properties
	if v, ok := d.GetOk("auto_grow_enabled"); ok {
		storage.StorageAutogrow = mariadb.StorageAutogrowDisabled
		if v.(bool) {
			storage.StorageAutogrow = mariadb.StorageAutogrowEnabled
		}
	}

	if v, ok := d.GetOk("backup_retention_days"); ok {
		storage.BackupRetentionDays = utils.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("geo_redundant_backup_enabled"); ok {
		storage.GeoRedundantBackup = mariadb.Disabled
		if v.(bool) {
			storage.GeoRedundantBackup = mariadb.Enabled
		}
	}

	if v, ok := d.GetOk("storage_mb"); ok {
		storage.StorageMB = utils.Int32(int32(v.(int)))
	}

	return &storage
}
