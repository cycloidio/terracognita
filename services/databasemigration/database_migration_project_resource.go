package databasemigration

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/datamigration/mgmt/2018-04-19/datamigration"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/databasemigration/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/databasemigration/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceDatabaseMigrationProject() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceDatabaseMigrationProjectCreateUpdate,
		Read:   resourceDatabaseMigrationProjectRead,
		Update: resourceDatabaseMigrationProjectCreateUpdate,
		Delete: resourceDatabaseMigrationProjectDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ProjectID(id)
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
				ValidateFunc: validate.ProjectName,
			},

			"service_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ServiceName,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"source_platform": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					// Now that go sdk only export SQL as source platform type, we only allow it here.
					string(datamigration.ProjectSourcePlatformSQL),
				}, false),
			},

			"target_platform": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					// Now that go sdk only export SQL as source platform type, we only allow it here.
					string(datamigration.ProjectTargetPlatformSQLDB),
				}, false),
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceDatabaseMigrationProjectCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DatabaseMigration.ProjectsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewProjectID(subscriptionId, d.Get("resource_group_name").(string), d.Get("service_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.ServiceName, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
		}
		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_database_migration_project", id.ID())
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	sourcePlatform := d.Get("source_platform").(string)
	targetPlatform := d.Get("target_platform").(string)
	t := d.Get("tags").(map[string]interface{})

	parameters := datamigration.Project{
		Location: utils.String(location),
		ProjectProperties: &datamigration.ProjectProperties{
			SourcePlatform: datamigration.ProjectSourcePlatform(sourcePlatform),
			TargetPlatform: datamigration.ProjectTargetPlatform(targetPlatform),
		},
		Tags: tags.Expand(t),
	}

	if _, err := client.CreateOrUpdate(ctx, parameters, id.ResourceGroup, id.ServiceName, id.Name); err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceDatabaseMigrationProjectRead(d, meta)
}

func resourceDatabaseMigrationProjectRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DatabaseMigration.ProjectsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ProjectID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.ServiceName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] %s does not exist - removing from state", *id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("service_name", id.ServiceName)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	if prop := resp.ProjectProperties; prop != nil {
		d.Set("source_platform", string(prop.SourcePlatform))
		d.Set("target_platform", string(prop.TargetPlatform))
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceDatabaseMigrationProjectDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DatabaseMigration.ProjectsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ProjectID(d.Id())
	if err != nil {
		return err
	}

	deleteRunningTasks := false
	if _, err := client.Delete(ctx, id.ResourceGroup, id.ServiceName, id.Name, &deleteRunningTasks); err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	return nil
}
