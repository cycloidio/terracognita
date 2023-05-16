package containers

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/containerregistry/mgmt/2021-08-01-preview/containerregistry"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/containers/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/containers/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceContainerRegistryScopeMap() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceContainerRegistryScopeMapCreate,
		Read:   resourceContainerRegistryScopeMapRead,
		Update: resourceContainerRegistryScopeMapUpdate,
		Delete: resourceContainerRegistryScopeMapDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ContainerRegistryScopeMapID(id)
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
				ValidateFunc: validate.ContainerRegistryScopeMapName,
			},

			"description": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"container_registry_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ContainerRegistryName,
			},

			"actions": {
				Type:     pluginsdk.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
			},
		},
	}
}

func resourceContainerRegistryScopeMapCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Containers.ScopeMapsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewContainerRegistryScopeMapID(subscriptionId, d.Get("resource_group_name").(string), d.Get("container_registry_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.RegistryName, id.ScopeMapName)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %s", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_container_registry_scope_map", id.ID())
		}
	}

	description := d.Get("description").(string)
	actions := d.Get("actions").([]interface{})

	parameters := containerregistry.ScopeMap{
		ScopeMapProperties: &containerregistry.ScopeMapProperties{
			Description: utils.String(description),
			Actions:     utils.ExpandStringSlice(actions),
		},
	}

	future, err := client.Create(ctx, id.ResourceGroup, id.RegistryName, id.ScopeMapName, parameters)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceContainerRegistryScopeMapRead(d, meta)
}

func resourceContainerRegistryScopeMapUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Containers.ScopeMapsClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for AzureRM Container Registry scope map update.")
	id, err := parse.ContainerRegistryScopeMapID(d.Id())
	if err != nil {
		return err
	}
	description := d.Get("description").(string)
	actions := d.Get("actions").([]interface{})

	parameters := containerregistry.ScopeMapUpdateParameters{
		ScopeMapPropertiesUpdateParameters: &containerregistry.ScopeMapPropertiesUpdateParameters{
			Description: utils.String(description),
			Actions:     utils.ExpandStringSlice(actions),
		},
	}

	future, err := client.Update(ctx, id.ResourceGroup, id.RegistryName, id.ScopeMapName, parameters)
	if err != nil {
		return fmt.Errorf("updating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for update of %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceContainerRegistryScopeMapRead(d, meta)
}

func resourceContainerRegistryScopeMapRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Containers.ScopeMapsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ContainerRegistryScopeMapID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.RegistryName, id.ScopeMapName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Scope Map %q was not found in Container Registry %q in Resource Group %q", id.ScopeMapName, id.RegistryName, id.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("making Read request on scope map %q in Azure Container Registry %q (Resource Group %q): %+v", id.ScopeMapName, id.RegistryName, id.ResourceGroup, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("container_registry_name", id.RegistryName)
	d.Set("description", resp.Description)
	d.Set("actions", utils.FlattenStringSlice(resp.Actions))

	return nil
}

func resourceContainerRegistryScopeMapDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Containers.ScopeMapsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ContainerRegistryScopeMapID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.RegistryName, id.ScopeMapName)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
	}

	return nil
}
