package datashare

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/datashare/mgmt/2019-11-01/datashare"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/datashare/helper"
	"github.com/hashicorp/terraform-provider-azurerm/services/datashare/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/datashare/validate"
	storageParsers "github.com/hashicorp/terraform-provider-azurerm/services/storage/parse"
	storageValidate "github.com/hashicorp/terraform-provider-azurerm/services/storage/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceDataShareDataSetDataLakeGen2() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceDataShareDataSetDataLakeGen2Create,
		Read:   resourceDataShareDataSetDataLakeGen2Read,
		Delete: resourceDataShareDataSetDataLakeGen2Delete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.DataSetID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.DataSetName(),
			},

			"share_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ShareID,
			},

			"storage_account_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: storageValidate.StorageAccountID,
			},

			"file_system_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"file_path": {
				Type:          pluginsdk.TypeString,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  validation.StringIsNotEmpty,
				ConflictsWith: []string{"folder_path"},
			},

			"folder_path": {
				Type:          pluginsdk.TypeString,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  validation.StringIsNotEmpty,
				ConflictsWith: []string{"file_path"},
			},

			"display_name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDataShareDataSetDataLakeGen2Create(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataShare.DataSetClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	shareId, err := parse.ShareID(d.Get("share_id").(string))
	if err != nil {
		return err
	}
	id := parse.NewDataSetID(shareId.SubscriptionId, shareId.ResourceGroup, shareId.AccountName, shareId.Name, d.Get("name").(string))

	existing, err := client.Get(ctx, id.ResourceGroup, id.AccountName, id.ShareName, id.Name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for presence of %s: %+v", id, err)
		}
	}
	existingId := helper.GetAzurermDataShareDataSetId(existing.Value)
	if existingId != nil && *existingId != "" {
		return tf.ImportAsExistsError("azurerm_data_share_dataset_data_lake_gen2", *existingId)
	}

	strId, err := storageParsers.StorageAccountID(d.Get("storage_account_id").(string))
	if err != nil {
		return err
	}

	var dataSet datashare.BasicDataSet

	if filePath, ok := d.GetOk("file_path"); ok {
		dataSet = datashare.ADLSGen2FileDataSet{
			Kind: datashare.KindAdlsGen2File,
			ADLSGen2FileProperties: &datashare.ADLSGen2FileProperties{
				StorageAccountName: utils.String(strId.Name),
				ResourceGroup:      utils.String(strId.ResourceGroup),
				SubscriptionID:     utils.String(strId.SubscriptionId),
				FileSystem:         utils.String(d.Get("file_system_name").(string)),
				FilePath:           utils.String(filePath.(string)),
			},
		}
	} else if folderPath, ok := d.GetOk("folder_path"); ok {
		dataSet = datashare.ADLSGen2FolderDataSet{
			Kind: datashare.KindAdlsGen2Folder,
			ADLSGen2FolderProperties: &datashare.ADLSGen2FolderProperties{
				StorageAccountName: utils.String(strId.Name),
				ResourceGroup:      utils.String(strId.ResourceGroup),
				SubscriptionID:     utils.String(strId.SubscriptionId),
				FileSystem:         utils.String(d.Get("file_system_name").(string)),
				FolderPath:         utils.String(folderPath.(string)),
			},
		}
	} else {
		dataSet = datashare.ADLSGen2FileSystemDataSet{
			Kind: datashare.KindAdlsGen2FileSystem,
			ADLSGen2FileSystemProperties: &datashare.ADLSGen2FileSystemProperties{
				StorageAccountName: utils.String(strId.Name),
				ResourceGroup:      utils.String(strId.ResourceGroup),
				SubscriptionID:     utils.String(strId.SubscriptionId),
				FileSystem:         utils.String(d.Get("file_system_name").(string)),
			},
		}
	}

	if _, err := client.Create(ctx, id.ResourceGroup, id.AccountName, id.ShareName, id.Name, dataSet); err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceDataShareDataSetDataLakeGen2Read(d, meta)
}

func resourceDataShareDataSetDataLakeGen2Read(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataShare.DataSetClient
	shareClient := meta.(*clients.Client).DataShare.SharesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.DataSetID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.AccountName, id.ShareName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] DataShare %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving DataShare DataSet %q (Resource Group %q / accountName %q / shareName %q): %+v", id.Name, id.ResourceGroup, id.AccountName, id.ShareName, err)
	}
	d.Set("name", id.Name)
	shareResp, err := shareClient.Get(ctx, id.ResourceGroup, id.AccountName, id.ShareName)
	if err != nil {
		return fmt.Errorf("retrieving DataShare %q (Resource Group %q / accountName %q): %+v", id.ShareName, id.ResourceGroup, id.AccountName, err)
	}
	if shareResp.ID == nil || *shareResp.ID == "" {
		return fmt.Errorf("empty or nil ID returned for DataShare %q (Resource Group %q / accountName %q)", id.ShareName, id.ResourceGroup, id.AccountName)
	}
	d.Set("share_id", shareResp.ID)

	switch resp := resp.Value.(type) {
	case datashare.ADLSGen2FileDataSet:
		if props := resp.ADLSGen2FileProperties; props != nil {
			d.Set("storage_account_id", fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Storage/storageAccounts/%s", *props.SubscriptionID, *props.ResourceGroup, *props.StorageAccountName))
			d.Set("file_system_name", props.FileSystem)
			d.Set("file_path", props.FilePath)
			d.Set("display_name", props.DataSetID)
		}

	case datashare.ADLSGen2FolderDataSet:
		if props := resp.ADLSGen2FolderProperties; props != nil {
			d.Set("storage_account_id", fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Storage/storageAccounts/%s", *props.SubscriptionID, *props.ResourceGroup, *props.StorageAccountName))
			d.Set("file_system_name", props.FileSystem)
			d.Set("folder_path", props.FolderPath)
			d.Set("display_name", props.DataSetID)
		}

	case datashare.ADLSGen2FileSystemDataSet:
		if props := resp.ADLSGen2FileSystemProperties; props != nil {
			d.Set("storage_account_id", fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Storage/storageAccounts/%s", *props.SubscriptionID, *props.ResourceGroup, *props.StorageAccountName))
			d.Set("file_system_name", props.FileSystem)
			d.Set("display_name", props.DataSetID)
		}

	default:
		return fmt.Errorf("data share dataset %q (Resource Group %q / accountName %q / shareName %q) is not a datalake store gen2 dataset", id.Name, id.ResourceGroup, id.AccountName, id.ShareName)
	}

	return nil
}

func resourceDataShareDataSetDataLakeGen2Delete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataShare.DataSetClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.DataSetID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.AccountName, id.ShareName, id.Name)
	if err != nil {
		return fmt.Errorf("deleting DataShare DataSet %q (Resource Group %q / accountName %q / shareName %q): %+v", id.Name, id.ResourceGroup, id.AccountName, id.ShareName, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of DataShare Data Lake Gen2 DataSet %q (Resource Group %q / accountName %q / shareName %q): %+v", id.Name, id.ResourceGroup, id.AccountName, id.ShareName, err)
	}

	return nil
}
