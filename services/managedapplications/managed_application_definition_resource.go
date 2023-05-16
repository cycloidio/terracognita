package managedapplications

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-07-01/managedapplications"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/managedapplications/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/managedapplications/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceManagedApplicationDefinition() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceManagedApplicationDefinitionCreateUpdate,
		Read:   resourceManagedApplicationDefinitionRead,
		Update: resourceManagedApplicationDefinitionCreateUpdate,
		Delete: resourceManagedApplicationDefinitionDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ApplicationDefinitionID(id)
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
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ApplicationDefinitionName,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"display_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.ApplicationDefinitionDisplayName,
			},

			"lock_level": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(managedapplications.CanNotDelete),
					string(managedapplications.None),
					string(managedapplications.ReadOnly),
				}, false),
			},

			"authorization": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"role_definition_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.IsUUID,
						},
						"service_principal_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.IsUUID,
						},
					},
				},
			},

			"create_ui_definition": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: pluginsdk.SuppressJsonDiff,
				ConflictsWith:    []string{"package_file_uri"},
			},

			"description": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validate.ApplicationDefinitionDescription,
			},

			"main_template": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: pluginsdk.SuppressJsonDiff,
				ConflictsWith:    []string{"package_file_uri"},
			},

			"package_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"package_file_uri": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceManagedApplicationDefinitionCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ManagedApplication.ApplicationDefinitionClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewApplicationDefinitionID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("failed to check for presence of existing %s: %+v", id, err)
			}
		}
		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_managed_application_definition", id.ID())
		}
	}

	parameters := managedapplications.ApplicationDefinition{
		Location: utils.String(azure.NormalizeLocation(d.Get("location"))),
		ApplicationDefinitionProperties: &managedapplications.ApplicationDefinitionProperties{
			Authorizations: expandManagedApplicationDefinitionAuthorization(d.Get("authorization").(*pluginsdk.Set).List()),
			Description:    utils.String(d.Get("description").(string)),
			DisplayName:    utils.String(d.Get("display_name").(string)),
			IsEnabled:      utils.Bool(d.Get("package_enabled").(bool)),
			LockLevel:      managedapplications.ApplicationLockLevel(d.Get("lock_level").(string)),
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	if v, ok := d.GetOk("create_ui_definition"); ok {
		parameters.CreateUIDefinition = utils.String(v.(string))
	}

	if v, ok := d.GetOk("main_template"); ok {
		parameters.MainTemplate = utils.String(v.(string))
	}

	if (parameters.CreateUIDefinition != nil && parameters.MainTemplate == nil) || (parameters.CreateUIDefinition == nil && parameters.MainTemplate != nil) {
		return fmt.Errorf("if either `create_ui_definition` or `main_template` is set the other one must be too")
	}

	if v, ok := d.GetOk("package_file_uri"); ok {
		parameters.PackageFileURI = utils.String(v.(string))
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, parameters)
	if err != nil {
		return fmt.Errorf("failed to create %s: %+v", id, err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("failed to wait for creation of %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceManagedApplicationDefinitionRead(d, meta)
}

func resourceManagedApplicationDefinitionRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ManagedApplication.ApplicationDefinitionClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ApplicationDefinitionID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Managed Application Definition %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("failed to read Managed Application Definition %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	if props := resp.ApplicationDefinitionProperties; props != nil {
		if err := d.Set("authorization", flattenManagedApplicationDefinitionAuthorization(props.Authorizations)); err != nil {
			return fmt.Errorf("setting `authorization`: %+v", err)
		}
		d.Set("description", props.Description)
		d.Set("display_name", props.DisplayName)
		d.Set("package_enabled", props.IsEnabled)
		d.Set("lock_level", string(props.LockLevel))
	}

	// the following are not returned from the API so lets pull it from state
	if v, ok := d.GetOk("create_ui_definition"); ok {
		d.Set("create_ui_definition", v.(string))
	}
	if v, ok := d.GetOk("main_template"); ok {
		d.Set("main_template", v.(string))
	}
	if v, ok := d.GetOk("package_file_uri"); ok {
		d.Set("package_file_uri", v.(string))
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceManagedApplicationDefinitionDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ManagedApplication.ApplicationDefinitionClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ApplicationDefinitionID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("failed to delete Managed Application Definition %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("failed to wait for deleting Managed Application Definition (Managed Application Definition Name %q / Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	return nil
}

func expandManagedApplicationDefinitionAuthorization(input []interface{}) *[]managedapplications.ApplicationAuthorization {
	results := make([]managedapplications.ApplicationAuthorization, 0)
	for _, item := range input {
		v := item.(map[string]interface{})
		result := managedapplications.ApplicationAuthorization{
			RoleDefinitionID: utils.String(v["role_definition_id"].(string)),
			PrincipalID:      utils.String(v["service_principal_id"].(string)),
		}

		results = append(results, result)
	}
	return &results
}

func flattenManagedApplicationDefinitionAuthorization(input *[]managedapplications.ApplicationAuthorization) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	for _, item := range *input {
		servicePrincipalId := ""
		if item.PrincipalID != nil {
			servicePrincipalId = *item.PrincipalID
		}

		roleDefinitionId := ""
		if item.RoleDefinitionID != nil {
			roleDefinitionId = *item.RoleDefinitionID
		}

		results = append(results, map[string]interface{}{
			"role_definition_id":   roleDefinitionId,
			"service_principal_id": servicePrincipalId,
		})
	}

	return results
}
