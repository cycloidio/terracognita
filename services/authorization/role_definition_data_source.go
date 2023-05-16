package authorization

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2020-04-01-preview/authorization"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
)

func dataSourceArmRoleDefinition() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceArmRoleDefinitionRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:          pluginsdk.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"role_definition_id"},
			},

			"role_definition_id": {
				Type:          pluginsdk.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validation.Any(validation.IsUUID, validation.StringIsEmpty),
			},

			"scope": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			// Computed

			"description": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"type": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"permissions": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"actions": {
							Type:     pluginsdk.TypeList,
							Computed: true,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
							},
						},

						"not_actions": {
							Type:     pluginsdk.TypeList,
							Computed: true,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
							},
						},

						"data_actions": {
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
							},
							Set: pluginsdk.HashString,
						},

						"not_data_actions": {
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
							},
							Set: pluginsdk.HashString,
						},
					},
				},
			},

			"assignable_scopes": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},
		},
	}
}

func dataSourceArmRoleDefinitionRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Authorization.RoleDefinitionsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	defId := d.Get("role_definition_id").(string)
	scope := d.Get("scope").(string)

	if name == "" && defId == "" {
		return fmt.Errorf("one of `name` or `role_definition_id` must be specified")
	}

	// search by name
	var role authorization.RoleDefinition
	if name != "" {
		// Accounting for eventual consistency
		err := pluginsdk.Retry(d.Timeout(pluginsdk.TimeoutRead), func() *pluginsdk.RetryError {

			roleDefinitions, err := client.List(ctx, scope, fmt.Sprintf("roleName eq '%s'", name))
			if err != nil {
				return pluginsdk.NonRetryableError(fmt.Errorf("loading Role Definition List: %+v", err))
			}
			if len(roleDefinitions.Values()) != 1 {
				return pluginsdk.RetryableError(fmt.Errorf("loading Role Definition List: could not find role '%s'", name))
			}
			if roleDefinitions.Values()[0].ID == nil {
				return pluginsdk.NonRetryableError(fmt.Errorf("loading Role Definition List: values[0].ID is nil '%s'", name))
			}

			defId = *roleDefinitions.Values()[0].ID
			role, err = client.GetByID(ctx, defId)
			if err != nil {
				return pluginsdk.NonRetryableError(fmt.Errorf("getting Role Definition by ID %s: %+v", defId, err))
			}
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		var err error
		role, err = client.Get(ctx, scope, defId)
		if err != nil {
			return fmt.Errorf("loading Role Definition: %+v", err)
		}
	}

	if role.ID == nil {
		return fmt.Errorf("returned role had a nil ID (id %q, scope %q, name %q)", defId, scope, name)
	}
	d.SetId(*role.ID)

	if props := role.RoleDefinitionProperties; props != nil {
		d.Set("name", props.RoleName)
		d.Set("role_definition_id", defId)
		d.Set("description", props.Description)
		d.Set("type", props.RoleType)

		permissions := flattenRoleDefinitionPermissions(props.Permissions)
		if err := d.Set("permissions", permissions); err != nil {
			return err
		}

		assignableScopes := flattenRoleDefinitionAssignableScopes(props.AssignableScopes)
		if err := d.Set("assignable_scopes", assignableScopes); err != nil {
			return err
		}
	}

	return nil
}
