package loganalytics

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/operationalinsights/mgmt/2020-08-01/operationalinsights"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	azValidate "github.com/hashicorp/terraform-provider-azurerm/helpers/validate"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/loganalytics/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/loganalytics/validate"
	storageValidate "github.com/hashicorp/terraform-provider-azurerm/services/storage/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceLogAnalyticsStorageInsights() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceLogAnalyticsStorageInsightsCreateUpdate,
		Read:   resourceLogAnalyticsStorageInsightsRead,
		Update: resourceLogAnalyticsStorageInsightsCreateUpdate,
		Delete: resourceLogAnalyticsStorageInsightsDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceIdThen(func(id string) error {
			_, err := parse.LogAnalyticsStorageInsightsID(id)
			return err
		}, func(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) ([]*pluginsdk.ResourceData, error) {
			if v, ok := d.GetOk("storage_account_key"); ok && v.(string) != "" {
				d.Set("storage_account_key", v)
			}

			return []*pluginsdk.ResourceData{d}, nil
		}),

		Schema: resourceLogAnalyticsStorageInsightsSchema(),
	}
}

func resourceLogAnalyticsStorageInsightsCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LogAnalytics.StorageInsightsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	storageAccountId := d.Get("storage_account_id").(string)
	storageAccountKey := d.Get("storage_account_key").(string)

	workspace, err := parse.LogAnalyticsWorkspaceID(d.Get("workspace_id").(string))
	if err != nil {
		return err
	}
	id := parse.NewLogAnalyticsStorageInsightsID(subscriptionId, resourceGroup, workspace.WorkspaceName, name)

	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceGroup, id.WorkspaceName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for present of existing Log Analytics Storage Insights %q (Resource Group %q / workspaceName %q): %+v", name, resourceGroup, id.WorkspaceName, err)
			}
		}
		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_log_analytics_storage_insights", id.ID())
		}
	}

	parameters := operationalinsights.StorageInsight{
		StorageInsightProperties: &operationalinsights.StorageInsightProperties{
			StorageAccount: expandStorageInsightConfigStorageAccount(storageAccountId, storageAccountKey),
		},
	}

	if _, ok := d.GetOk("table_names"); ok {
		parameters.StorageInsightProperties.Tables = utils.ExpandStringSlice(d.Get("table_names").(*pluginsdk.Set).List())
	}

	if _, ok := d.GetOk("blob_container_names"); ok {
		parameters.StorageInsightProperties.Containers = utils.ExpandStringSlice(d.Get("blob_container_names").(*pluginsdk.Set).List())
	}

	if _, err := client.CreateOrUpdate(ctx, resourceGroup, id.WorkspaceName, name, parameters); err != nil {
		return fmt.Errorf("creating/updating Log Analytics Storage Insights %q (Resource Group %q / workspaceName %q): %+v", name, resourceGroup, id.WorkspaceName, err)
	}

	d.SetId(id.ID())
	return resourceLogAnalyticsStorageInsightsRead(d, meta)
}

func resourceLogAnalyticsStorageInsightsRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LogAnalytics.StorageInsightsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LogAnalyticsStorageInsightsID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.WorkspaceName, id.StorageInsightConfigName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Log Analytics Storage Insights %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Log Analytics Storage Insights %q (Resource Group %q / workspaceName %q): %+v", id.StorageInsightConfigName, id.ResourceGroup, id.WorkspaceName, err)
	}

	d.Set("name", id.StorageInsightConfigName)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("workspace_id", parse.NewLogAnalyticsWorkspaceID(id.SubscriptionId, id.ResourceGroup, id.WorkspaceName).ID())

	if props := resp.StorageInsightProperties; props != nil {
		d.Set("blob_container_names", utils.FlattenStringSlice(props.Containers))
		storageAccountId := ""
		if props.StorageAccount != nil && props.StorageAccount.ID != nil {
			storageAccountId = *props.StorageAccount.ID
		}
		d.Set("storage_account_id", storageAccountId)
		d.Set("table_names", utils.FlattenStringSlice(props.Tables))
	}

	return nil
}

func resourceLogAnalyticsStorageInsightsDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LogAnalytics.StorageInsightsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LogAnalyticsStorageInsightsID(d.Id())
	if err != nil {
		return err
	}

	if _, err := client.Delete(ctx, id.ResourceGroup, id.WorkspaceName, id.StorageInsightConfigName); err != nil {
		return fmt.Errorf("deleting LogAnalytics Storage Insight Config %q (Resource Group %q / workspaceName %q): %+v", id.StorageInsightConfigName, id.ResourceGroup, id.WorkspaceName, err)
	}
	return nil
}

func expandStorageInsightConfigStorageAccount(id string, key string) *operationalinsights.StorageAccount {
	return &operationalinsights.StorageAccount{
		ID:  utils.String(id),
		Key: utils.String(key),
	}
}

func resourceLogAnalyticsStorageInsightsSchema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.LogAnalyticsStorageInsightsName,
		},

		"resource_group_name": azure.SchemaResourceGroupName(),

		"workspace_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.LogAnalyticsWorkspaceID,
		},

		"storage_account_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ValidateFunc: storageValidate.StorageAccountID,
		},

		"storage_account_key": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			Sensitive:    true,
			ValidateFunc: azValidate.Base64EncodedString,
		},

		"blob_container_names": {
			Type:     pluginsdk.TypeSet,
			Optional: true,
			Elem: &pluginsdk.Schema{
				Type:         pluginsdk.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
		},

		"table_names": {
			Type:     pluginsdk.TypeSet,
			Optional: true,
			MinItems: 1,
			Elem: &pluginsdk.Schema{
				Type:         pluginsdk.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}
