package kusto

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/kusto/mgmt/2021-08-27/kusto"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/locks"
	"github.com/hashicorp/terraform-provider-azurerm/services/kusto/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/kusto/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceKustoDatabaseScript() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceKustoDatabaseScriptCreateUpdate,
		Read:   resourceKustoDatabaseScriptRead,
		Update: resourceKustoDatabaseScriptCreateUpdate,
		Delete: resourceKustoDatabaseScriptDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ScriptID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"database_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.DatabaseID,
			},

			"url": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"sas_token": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"continue_on_errors_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			"force_an_update_when_value_changed": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func resourceKustoDatabaseScriptCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Kusto.ScriptsClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	databaseId, _ := parse.DatabaseID(d.Get("database_id").(string))
	id := parse.NewScriptID(databaseId.SubscriptionId, databaseId.ResourceGroup, databaseId.ClusterName, databaseId.Name, d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.ClusterName, id.DatabaseName, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for existing %q: %+v", id, err)
			}
		}
		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_kusto_script", id.ID())
		}
	}

	clusterId := parse.NewClusterID(databaseId.SubscriptionId, databaseId.ResourceGroup, databaseId.ClusterName)
	locks.ByID(clusterId.ID())
	defer locks.UnlockByID(clusterId.ID())

	forceUpdateTag := d.Get("force_an_update_when_value_changed").(string)
	if len(forceUpdateTag) == 0 {
		forceUpdateTag, _ = uuid.GenerateUUID()
	}

	parameters := kusto.Script{
		ScriptProperties: &kusto.ScriptProperties{
			ContinueOnErrors:  utils.Bool(d.Get("continue_on_errors_enabled").(bool)),
			ForceUpdateTag:    utils.String(forceUpdateTag),
			ScriptURL:         utils.String(d.Get("url").(string)),
			ScriptURLSasToken: utils.String(d.Get("sas_token").(string)),
		},
	}
	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.ClusterName, id.DatabaseName, id.Name, parameters)
	if err != nil {
		return fmt.Errorf("creating %q: %+v", id, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of %q: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceKustoDatabaseScriptRead(d, meta)
}

func resourceKustoDatabaseScriptRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Kusto.ScriptsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ScriptID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.ClusterName, id.DatabaseName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] %s does not exist - removing from state", *id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}
	d.Set("name", id.Name)
	d.Set("database_id", parse.NewDatabaseID(id.SubscriptionId, id.ResourceGroup, id.ClusterName, id.DatabaseName).ID())
	if props := resp.ScriptProperties; props != nil {
		d.Set("continue_on_errors_enabled", props.ContinueOnErrors)
		d.Set("force_an_update_when_value_changed", props.ForceUpdateTag)
		d.Set("url", props.ScriptURL)
	}
	return nil
}

func resourceKustoDatabaseScriptDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Kusto.ScriptsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ScriptID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.ClusterName, id.DatabaseName, id.Name)
	if err != nil {
		return fmt.Errorf("deleting %q: %+v", id, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of %q: %+v", id, err)
	}
	return nil
}
