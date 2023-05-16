package blueprints

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/blueprints/validate"
	mgValidate "github.com/hashicorp/terraform-provider-azurerm/services/managementgroup/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func dataSourceBlueprintDefinition() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceBlueprintDefinitionRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.DefinitionName,
			},

			"scope_id": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.Any(
					azure.ValidateResourceID,
					mgValidate.ManagementGroupID,
				),
			},

			// Computed
			"description": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"display_name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"last_modified": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"target_scope": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"time_created": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"versions": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},
		},
	}
}

func dataSourceBlueprintDefinitionRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Blueprints.BlueprintsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	scope := d.Get("scope_id").(string)

	resp, err := client.Get(ctx, scope, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Blueprint Definition %q not found in Scope (%q): %+v", name, scope, err)
		}

		return fmt.Errorf("Read failed for Blueprint Definition (%q) in Sccope (%q): %+v", name, scope, err)
	}

	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("Failed to retrieve ID for Blueprint %q", name)
	} else {
		d.SetId(*resp.ID)
	}

	if resp.Description != nil {
		d.Set("description", resp.Description)
	}

	if resp.DisplayName != nil {
		d.Set("display_name", resp.DisplayName)
	}

	d.Set("last_modified", resp.Status.LastModified.String())

	d.Set("time_created", resp.Status.TimeCreated.String())

	d.Set("target_scope", resp.TargetScope)

	if resp.Versions != nil {
		d.Set("versions", resp.Versions)
	}

	return nil
}
