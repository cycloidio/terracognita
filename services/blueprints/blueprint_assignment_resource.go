package blueprints

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/blueprint/mgmt/2018-11-01-preview/blueprint"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/blueprints/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/blueprints/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceBlueprintAssignment() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceBlueprintAssignmentCreateUpdate,
		Update: resourceBlueprintAssignmentCreateUpdate,
		Read:   resourceBlueprintAssignmentRead,
		Delete: resourceBlueprintAssignmentDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.AssignmentID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"target_subscription_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"location": commonschema.Location(),

			"identity": commonschema.UserAssignedIdentityRequired(),

			"version_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.VersionID,
			},

			"parameter_values": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				StateFunc:        normalizeAssignmentParameterValuesJSON,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: pluginsdk.SuppressJsonDiff,
			},

			"resource_groups": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				StateFunc:        normalizeAssignmentResourceGroupValuesJSON,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: pluginsdk.SuppressJsonDiff,
			},

			"lock_mode": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  string(blueprint.None),
				ValidateFunc: validation.StringInSlice([]string{
					string(blueprint.AssignmentLockModeNone),
					string(blueprint.AssignmentLockModeAllResourcesReadOnly),
					string(blueprint.AssignmentLockModeAllResourcesDoNotDelete),
				}, false),
				// The first character of value returned by the service is always in lower case.
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"lock_exclude_principals": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 5,
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeString,
					ValidateFunc: validation.IsUUID,
				},
			},

			"lock_exclude_actions": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 200,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"description": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"display_name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"blueprint_name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"type": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceBlueprintAssignmentCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Blueprints.AssignmentsClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	blueprintId := d.Get("version_id").(string)
	targetScope := d.Get("target_subscription_id").(string)

	if d.IsNewResource() {
		resp, err := client.Get(ctx, targetScope, name)
		if err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("failure checking for existing Blueprint Assignment %q in scope %q", name, targetScope)
			}
		}
		if !utils.ResponseWasNotFound(resp.Response) {
			return tf.ImportAsExistsError("azurerm_blueprint_assignment", *resp.ID)
		}
	}

	assignment := blueprint.Assignment{
		AssignmentProperties: &blueprint.AssignmentProperties{
			BlueprintID: utils.String(blueprintId), // This is mislabeled - The ID is that of the Published Version, not just the Blueprint
			Scope:       utils.String(targetScope),
		},
		Location: utils.String(azure.NormalizeLocation(d.Get("location"))),
	}

	if lockModeRaw, ok := d.GetOk("lock_mode"); ok {
		assignmentLockSettings := &blueprint.AssignmentLockSettings{}
		lockMode := lockModeRaw.(string)
		assignmentLockSettings.Mode = blueprint.AssignmentLockMode(lockMode)
		if lockMode != "None" {
			excludedPrincipalsRaw := d.Get("lock_exclude_principals").([]interface{})
			if len(excludedPrincipalsRaw) != 0 {
				assignmentLockSettings.ExcludedPrincipals = utils.ExpandStringSlice(excludedPrincipalsRaw)
			}

			excludedActionsRaw := d.Get("lock_exclude_actions").([]interface{})
			if len(excludedActionsRaw) != 0 {
				assignmentLockSettings.ExcludedActions = utils.ExpandStringSlice(excludedActionsRaw)
			}
		}
		assignment.AssignmentProperties.Locks = assignmentLockSettings
	}

	identity, err := expandArmBlueprintAssignmentIdentity(d.Get("identity").([]interface{}))
	if err != nil {
		return fmt.Errorf("expanding `identity`: %+v", err)
	}
	assignment.Identity = identity

	if paramValuesRaw := d.Get("parameter_values"); paramValuesRaw != "" {
		assignment.Parameters = expandArmBlueprintAssignmentParameters(paramValuesRaw.(string))
	} else {
		assignment.Parameters = expandArmBlueprintAssignmentParameters("{}")
	}

	if resourceGroupsRaw := d.Get("resource_groups"); resourceGroupsRaw != "" {
		assignment.ResourceGroups = expandArmBlueprintAssignmentResourceGroups(resourceGroupsRaw.(string))
	} else {
		assignment.ResourceGroups = expandArmBlueprintAssignmentResourceGroups("{}")
	}

	resp, err := client.CreateOrUpdate(ctx, targetScope, name, assignment)
	if err != nil {
		return err
	}

	stateConf := &pluginsdk.StateChangeConf{
		Pending: []string{
			string(blueprint.Waiting),
			string(blueprint.Validating),
			string(blueprint.Creating),
			string(blueprint.Deploying),
			string(blueprint.Locking),
		},
		Target:  []string{string(blueprint.Succeeded)},
		Refresh: blueprintAssignmentCreateStateRefreshFunc(ctx, client, targetScope, name),
		Timeout: d.Timeout(pluginsdk.TimeoutCreate),
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("failed waiting for Blueprint Assignment %q (Scope %q): %+v", name, targetScope, err)
	}

	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("could not read ID from Blueprint Assignment %q on scope %q", name, targetScope)
	}

	d.SetId(*resp.ID)

	return resourceBlueprintAssignmentRead(d, meta)
}

func resourceBlueprintAssignmentRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Blueprints.AssignmentsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.AssignmentID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.Scope, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] the Blueprint Assignment %q does not exist - removing from state", id.Name)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Read failed for Blueprint Assignment (%q): %+v", id.Name, err)
	}

	if resp.Name != nil {
		d.Set("name", resp.Name)
	}

	if resp.Scope != nil {
		d.Set("target_subscription_id", resp.Scope)
	}

	if resp.Location != nil {
		d.Set("location", azure.NormalizeLocation(*resp.Location))
	}

	identity, err := flattenArmBlueprintAssignmentIdentity(resp.Identity)
	if err != nil {
		return fmt.Errorf("flattening `identity`: %+v", err)
	}
	if err := d.Set("identity", identity); err != nil {
		return fmt.Errorf("setting `identity`: %+v", err)
	}

	if resp.AssignmentProperties != nil {
		if resp.AssignmentProperties.BlueprintID != nil {
			d.Set("version_id", resp.AssignmentProperties.BlueprintID)
		}

		if resp.AssignmentProperties.Parameters != nil {
			params, err := flattenArmBlueprintAssignmentParameters(resp.Parameters)
			if err != nil {
				return err
			}
			d.Set("parameter_values", params)
		}

		if resp.AssignmentProperties.ResourceGroups != nil {
			resourceGroups, err := flattenArmBlueprintAssignmentResourceGroups(resp.ResourceGroups)
			if err != nil {
				return err
			}
			d.Set("resource_groups", resourceGroups)
		}

		// Locks
		if locks := resp.Locks; locks != nil {
			d.Set("lock_mode", locks.Mode)
			if locks.ExcludedPrincipals != nil {
				d.Set("lock_exclude_principals", locks.ExcludedPrincipals)
			}
			if locks.ExcludedActions != nil {
				d.Set("lock_exclude_actions", locks.ExcludedActions)
			}
		}
	}

	if resp.DisplayName != nil {
		d.Set("display_name", resp.DisplayName)
	}

	if resp.Description != nil {
		d.Set("description", resp.Description)
	}

	return nil
}

func resourceBlueprintAssignmentDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Blueprints.AssignmentsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.AssignmentID(d.Id())
	if err != nil {
		return err
	}

	// We use none here to align the previous behaviour of the blueprint resource
	// TODO: we could add a features flag for the blueprint to empower terraform when deleting the blueprint to delete all the generated resources as well
	resp, err := client.Delete(ctx, id.Scope, id.Name, blueprint.None)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return nil
		}
		return fmt.Errorf("failed to delete Blueprint Assignment %q from scope %q: %+v", id.Name, id.Scope, err)
	}

	stateConf := &pluginsdk.StateChangeConf{
		Pending: []string{
			string(blueprint.Waiting),
			string(blueprint.Validating),
			string(blueprint.Locking),
			string(blueprint.Deleting),
			string(blueprint.Failed),
		},
		Target:  []string{"NotFound"},
		Refresh: blueprintAssignmentDeleteStateRefreshFunc(ctx, client, id.Scope, id.Name),
		Timeout: d.Timeout(pluginsdk.TimeoutDelete),
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("Failed waiting for Blueprint Assignment %q (Scope %q): %+v", id.Name, id.Scope, err)
	}

	return nil
}
