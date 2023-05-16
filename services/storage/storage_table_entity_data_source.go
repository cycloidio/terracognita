package storage

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/storage/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/tombuildsstuff/giovanni/storage/2019-12-12/table/entities"
)

func dataSourceStorageTableEntity() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceStorageTableEntityRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"table_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.StorageTableName,
			},

			"storage_account_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.StorageAccountName,
			},

			"partition_key": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"row_key": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"entity": {
				Type:     pluginsdk.TypeMap,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},
		},
	}
}

func dataSourceStorageTableEntityRead(d *pluginsdk.ResourceData, meta interface{}) error {
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()
	storageClient := meta.(*clients.Client).Storage

	storageAccountName := d.Get("storage_account_name").(string)
	tableName := d.Get("table_name").(string)
	partitionKey := d.Get("partition_key").(string)
	rowKey := d.Get("row_key").(string)

	account, err := storageClient.FindAccount(ctx, storageAccountName)
	if err != nil {
		return fmt.Errorf("retrieving Account %q for Table %q: %s", storageAccountName, tableName, err)
	}
	if account == nil {
		log.Printf("[WARN] Unable to determine Resource Group for Storage Table %q (Account %s) - assuming removed & removing from state", tableName, storageAccountName)
		d.SetId("")
		return nil
	}

	client, err := storageClient.TableEntityClient(ctx, *account)
	if err != nil {
		return fmt.Errorf("building Table Entity Client for Storage Account %q (Resource Group %q): %s", storageAccountName, account.ResourceGroup, err)
	}

	id := client.GetResourceID(storageAccountName, tableName, partitionKey, rowKey)

	input := entities.GetEntityInput{
		PartitionKey:  partitionKey,
		RowKey:        rowKey,
		MetaDataLevel: entities.NoMetaData,
	}

	result, err := client.Get(ctx, storageAccountName, tableName, input)
	if err != nil {
		return fmt.Errorf("retrieving Entity (Partition Key %q / Row Key %q) (Table %q / Storage Account %q / Resource Group %q): %s", partitionKey, rowKey, tableName, storageAccountName, account.ResourceGroup, err)
	}

	d.Set("storage_account_name", storageAccountName)
	d.Set("table_name", tableName)
	d.Set("partition_key", partitionKey)
	d.Set("row_key", rowKey)
	if err := d.Set("entity", flattenEntity(result.Entity)); err != nil {
		return fmt.Errorf("setting `entity` for Entity (Partition Key %q / Row Key %q) (Table %q / Storage Account %q / Resource Group %q): %s", partitionKey, rowKey, tableName, storageAccountName, account.ResourceGroup, err)
	}
	d.SetId(id)

	return nil
}
