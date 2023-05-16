package storage

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/storagesync/mgmt/2020-03-01/storagesync"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/storage/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/storage/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceStorageSyncCloudEndpoint() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceStorageSyncCloudEndpointCreate,
		Read:   resourceStorageSyncCloudEndpointRead,
		Delete: resourceStorageSyncCloudEndpointDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.StorageSyncCloudEndpointID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(45 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(45 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.StorageSyncName,
			},

			"storage_sync_group_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.StorageSyncGroupID,
			},

			"file_share_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.StorageShareName,
			},

			"storage_account_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.StorageAccountID,
			},

			"storage_account_tenant_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
		},
	}
}

func resourceStorageSyncCloudEndpointCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage.CloudEndpointsClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	groupId, err := parse.StorageSyncGroupID(d.Get("storage_sync_group_id").(string))
	if err != nil {
		return err
	}

	id := parse.NewStorageSyncCloudEndpointID(groupId.SubscriptionId, groupId.ResourceGroup, groupId.StorageSyncServiceName, groupId.SyncGroupName, d.Get("name").(string))
	existing, err := client.Get(ctx, id.ResourceGroup, id.StorageSyncServiceName, id.SyncGroupName, id.CloudEndpointName)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
		}
	}
	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurerm_storage_sync_cloud_endpoint", id.ID())
	}

	parameters := storagesync.CloudEndpointCreateParameters{
		CloudEndpointCreateParametersProperties: &storagesync.CloudEndpointCreateParametersProperties{
			StorageAccountResourceID: utils.String(d.Get("storage_account_id").(string)),
			AzureFileShareName:       utils.String(d.Get("file_share_name").(string)),
		},
	}

	tenantId := meta.(*clients.Client).Account.TenantId
	if v, ok := d.GetOk("storage_account_tenant_id"); ok {
		tenantId = v.(string)
	}
	parameters.CloudEndpointCreateParametersProperties.StorageAccountTenantID = &tenantId

	future, err := client.Create(ctx, id.ResourceGroup, id.StorageSyncServiceName, id.SyncGroupName, id.CloudEndpointName, parameters)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of %s: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceStorageSyncCloudEndpointRead(d, meta)
}

func resourceStorageSyncCloudEndpointRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage.CloudEndpointsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.StorageSyncCloudEndpointID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.StorageSyncServiceName, id.SyncGroupName, id.CloudEndpointName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] %s does not exist - removing from state", *id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	d.Set("name", id.CloudEndpointName)

	groupId := parse.NewStorageSyncGroupID(id.SubscriptionId, id.ResourceGroup, id.StorageSyncServiceName, id.SyncGroupName)
	d.Set("storage_sync_group_id", groupId.ID())
	if props := resp.CloudEndpointProperties; props != nil {
		d.Set("file_share_name", props.AzureFileShareName)
		d.Set("storage_account_id", props.StorageAccountResourceID)
		d.Set("storage_account_tenant_id", props.StorageAccountTenantID)
	}

	return nil
}

func resourceStorageSyncCloudEndpointDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage.CloudEndpointsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.StorageSyncCloudEndpointID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.StorageSyncServiceName, id.SyncGroupName, id.CloudEndpointName)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
	}

	return nil
}
