package logic

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/services/logic/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
)

func resourceLogicAppActionCustom() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceLogicAppActionCustomCreateUpdate,
		Read:   resourceLogicAppActionCustomRead,
		Update: resourceLogicAppActionCustomCreateUpdate,
		Delete: resourceLogicAppActionCustomDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ActionID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"logic_app_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"body": {
				Type:             pluginsdk.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: pluginsdk.SuppressJsonDiff,
			},
		},
	}
}

func resourceLogicAppActionCustomCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	workflowId, err := parse.WorkflowID(d.Get("logic_app_id").(string))
	if err != nil {
		return err
	}

	id := parse.NewActionID(workflowId.SubscriptionId, workflowId.ResourceGroup, workflowId.Name, d.Get("name").(string))

	bodyRaw := d.Get("body").(string)

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(bodyRaw), &body); err != nil {
		return fmt.Errorf("unmarshalling JSON for Custom Action %q: %+v", id.Name, err)
	}

	if err := resourceLogicAppActionUpdate(d, meta, *workflowId, id, body, "azurerm_logic_app_action_custom"); err != nil {
		return err
	}

	return resourceLogicAppActionCustomRead(d, meta)
}

func resourceLogicAppActionCustomRead(d *pluginsdk.ResourceData, meta interface{}) error {
	id, err := parse.ActionID(d.Id())
	if err != nil {
		return err
	}

	t, app, err := retrieveLogicAppAction(d, meta, id.ResourceGroup, id.WorkflowName, id.Name)
	if err != nil {
		return err
	}

	if t == nil {
		log.Printf("[DEBUG] Logic App %q (Resource Group %q) does not contain Action %q - removing from state", id.WorkflowName, id.ResourceGroup, id.Name)
		d.SetId("")
		return nil
	}

	action := *t

	d.Set("name", id.Name)
	d.Set("logic_app_id", app.ID)

	body, err := json.Marshal(action)
	if err != nil {
		return fmt.Errorf("serializing `body` for Action %q: %+v", id.Name, err)
	}

	if err := d.Set("body", string(body)); err != nil {
		return fmt.Errorf("setting `body` for Action %q: %+v", id.Name, err)
	}

	return nil
}

func resourceLogicAppActionCustomDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	id, err := parse.ActionID(d.Id())
	if err != nil {
		return err
	}

	err = resourceLogicAppActionRemove(d, meta, id.ResourceGroup, id.WorkflowName, id.Name)
	if err != nil {
		return fmt.Errorf("removing Action %s: %+v", id, err)
	}

	return nil
}
