package datashare

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/datashare/mgmt/2019-11-01/datashare"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/datashare/helper"
	"github.com/hashicorp/terraform-provider-azurerm/services/datashare/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/datashare/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceDataShareDataSetKustoCluster() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceDataShareDataSetKustoClusterCreate,
		Read:   resourceDataShareDataSetKustoClusterRead,
		Delete: resourceDataShareDataSetKustoClusterDelete,

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

			"kusto_cluster_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"display_name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"kusto_cluster_location": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDataShareDataSetKustoClusterCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataShare.DataSetClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	shareId, err := parse.ShareID(d.Get("share_id").(string))
	if err != nil {
		return err
	}
	id := parse.NewDataSetID(shareId.SubscriptionId, shareId.ResourceGroup, shareId.AccountName, shareId.Name, d.Get("name").(string))

	existingModel, err := client.Get(ctx, id.ResourceGroup, id.AccountName, id.ShareName, id.Name)
	if err != nil {
		if !utils.ResponseWasNotFound(existingModel.Response) {
			return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
		}
	}
	existingId := helper.GetAzurermDataShareDataSetId(existingModel.Value)
	if existingId != nil && *existingId != "" {
		return tf.ImportAsExistsError("azurerm_data_share_dataset_kusto_cluster", *existingId)
	}

	dataSet := datashare.KustoClusterDataSet{
		Kind: datashare.KindKustoCluster,
		KustoClusterDataSetProperties: &datashare.KustoClusterDataSetProperties{
			KustoClusterResourceID: utils.String(d.Get("kusto_cluster_id").(string)),
		},
	}

	if _, err := client.Create(ctx, id.ResourceGroup, id.AccountName, id.ShareName, id.Name, dataSet); err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceDataShareDataSetKustoClusterRead(d, meta)
}

func resourceDataShareDataSetKustoClusterRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataShare.DataSetClient
	shareClient := meta.(*clients.Client).DataShare.SharesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.DataSetID(d.Id())
	if err != nil {
		return err
	}

	respModel, err := client.Get(ctx, id.ResourceGroup, id.AccountName, id.ShareName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(respModel.Response) {
			log.Printf("[INFO] DataShare %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving DataShare Kusto Cluster DataSet %q (Resource Group %q / accountName %q / shareName %q): %+v", id.Name, id.ResourceGroup, id.AccountName, id.ShareName, err)
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

	resp, ok := respModel.Value.AsKustoClusterDataSet()
	if !ok {
		return fmt.Errorf("dataShare dataset %q (Resource Group %q / accountName %q / shareName %q) is not kusto cluster dataset", id.Name, id.ResourceGroup, id.AccountName, id.ShareName)
	}
	if props := resp.KustoClusterDataSetProperties; props != nil {
		d.Set("kusto_cluster_id", props.KustoClusterResourceID)
		d.Set("display_name", props.DataSetID)
		d.Set("kusto_cluster_location", props.Location)
	}

	return nil
}

func resourceDataShareDataSetKustoClusterDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataShare.DataSetClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.DataSetID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.AccountName, id.ShareName, id.Name)
	if err != nil {
		return fmt.Errorf("deleting DataShare Kusto Cluster DataSet %q (Resource Group %q / accountName %q / shareName %q): %+v", id.Name, id.ResourceGroup, id.AccountName, id.ShareName, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of DataShare Kusto Cluster DataSet %q (Resource Group %q / accountName %q / shareName %q): %+v", id.Name, id.ResourceGroup, id.AccountName, id.ShareName, err)
	}

	return nil
}
