package storage

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/storage/migration"
	"github.com/hashicorp/terraform-provider-azurerm/services/storage/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/storage/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/tombuildsstuff/giovanni/storage/2019-12-12/blob/containers"
)

func resourceStorageContainer() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceStorageContainerCreate,
		Read:   resourceStorageContainerRead,
		Delete: resourceStorageContainerDelete,
		Update: resourceStorageContainerUpdate,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.StorageContainerDataPlaneID(id)
			return err
		}),

		SchemaVersion: 1,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.ContainerV0ToV1{},
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
				ValidateFunc: validate.StorageContainerName,
			},

			"storage_account_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.StorageAccountName,
			},

			"container_access_type": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  "private",
				ValidateFunc: validation.StringInSlice([]string{
					string(containers.Blob),
					string(containers.Container),
					"private",
				}, false),
			},

			"metadata": MetaDataComputedSchema(),

			// TODO: support for ACL's, Legal Holds and Immutability Policies
			"has_immutability_policy": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"has_legal_hold": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"resource_manager_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceStorageContainerCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	storageClient := meta.(*clients.Client).Storage
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	containerName := d.Get("name").(string)
	accountName := d.Get("storage_account_name").(string)
	accessLevelRaw := d.Get("container_access_type").(string)
	accessLevel := expandStorageContainerAccessLevel(accessLevelRaw)

	metaDataRaw := d.Get("metadata").(map[string]interface{})
	metaData := ExpandMetaData(metaDataRaw)

	account, err := storageClient.FindAccount(ctx, accountName)
	if err != nil {
		return fmt.Errorf("retrieving Account %q for Container %q: %s", accountName, containerName, err)
	}
	if account == nil {
		return fmt.Errorf("Unable to locate Storage Account %q!", accountName)
	}

	client, err := storageClient.ContainersClient(ctx, *account)
	if err != nil {
		return fmt.Errorf("building storage client: %+v", err)
	}

	id := parse.NewStorageContainerDataPlaneId(accountName, storageClient.Environment.StorageEndpointSuffix, containerName).ID()
	exists, err := client.Exists(ctx, account.ResourceGroup, accountName, containerName)
	if err != nil {
		return err
	}
	if exists != nil && *exists {
		return tf.ImportAsExistsError("azurerm_storage_container", id)
	}

	log.Printf("[INFO] Creating Container %q in Storage Account %q", containerName, accountName)
	input := containers.CreateInput{
		AccessLevel: accessLevel,
		MetaData:    metaData,
	}

	if err := client.Create(ctx, account.ResourceGroup, accountName, containerName, input); err != nil {
		return fmt.Errorf("failed creating container: %+v", err)
	}

	d.SetId(id)
	return resourceStorageContainerRead(d, meta)
}

func resourceStorageContainerUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	storageClient := meta.(*clients.Client).Storage
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.StorageContainerDataPlaneID(d.Id())
	if err != nil {
		return err
	}

	account, err := storageClient.FindAccount(ctx, id.AccountName)
	if err != nil {
		return fmt.Errorf("retrieving Account %q for Container %q: %s", id.AccountName, id.Name, err)
	}
	if account == nil {
		return fmt.Errorf("Unable to locate Storage Account %q!", id.AccountName)
	}
	client, err := storageClient.ContainersClient(ctx, *account)
	if err != nil {
		return fmt.Errorf("building Containers Client for Storage Account %q (Resource Group %q): %s", id.AccountName, account.ResourceGroup, err)
	}

	if d.HasChange("container_access_type") {
		log.Printf("[DEBUG] Updating the Access Control for Container %q (Storage Account %q / Resource Group %q)..", id.Name, id.AccountName, account.ResourceGroup)
		accessLevelRaw := d.Get("container_access_type").(string)
		accessLevel := expandStorageContainerAccessLevel(accessLevelRaw)

		if err := client.UpdateAccessLevel(ctx, account.ResourceGroup, id.AccountName, id.Name, accessLevel); err != nil {
			return fmt.Errorf("updating the Access Control for Container %q (Storage Account %q / Resource Group %q): %s", id.Name, id.AccountName, account.ResourceGroup, err)
		}

		log.Printf("[DEBUG] Updated the Access Control for Container %q (Storage Account %q / Resource Group %q)", id.Name, id.AccountName, account.ResourceGroup)
	}

	if d.HasChange("metadata") {
		log.Printf("[DEBUG] Updating the MetaData for Container %q (Storage Account %q / Resource Group %q)..", id.Name, id.AccountName, account.ResourceGroup)
		metaDataRaw := d.Get("metadata").(map[string]interface{})
		metaData := ExpandMetaData(metaDataRaw)

		if err := client.UpdateMetaData(ctx, account.ResourceGroup, id.AccountName, id.Name, metaData); err != nil {
			return fmt.Errorf("updating the MetaData for Container %q (Storage Account %q / Resource Group %q): %s", id.Name, id.AccountName, account.ResourceGroup, err)
		}

		log.Printf("[DEBUG] Updated the MetaData for Container %q (Storage Account %q / Resource Group %q)", id.Name, id.AccountName, account.ResourceGroup)
	}

	return resourceStorageContainerRead(d, meta)
}

func resourceStorageContainerRead(d *pluginsdk.ResourceData, meta interface{}) error {
	storageClient := meta.(*clients.Client).Storage
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.StorageContainerDataPlaneID(d.Id())
	if err != nil {
		return err
	}

	account, err := storageClient.FindAccount(ctx, id.AccountName)
	if err != nil {
		return fmt.Errorf("retrieving Account %q for Container %q: %s", id.AccountName, id.Name, err)
	}
	if account == nil {
		log.Printf("[DEBUG] Unable to locate Account %q for Storage Container %q - assuming removed & removing from state", id.AccountName, id.Name)
		d.SetId("")
		return nil
	}
	client, err := storageClient.ContainersClient(ctx, *account)
	if err != nil {
		return fmt.Errorf("building Containers Client for Storage Account %q (Resource Group %q): %s", id.AccountName, account.ResourceGroup, err)
	}

	props, err := client.Get(ctx, account.ResourceGroup, id.AccountName, id.Name)
	if err != nil {
		return fmt.Errorf("retrieving Container %q (Account %q / Resource Group %q): %s", id.Name, id.AccountName, account.ResourceGroup, err)
	}
	if props == nil {
		log.Printf("[DEBUG] Container %q was not found in Account %q / Resource Group %q - assuming removed & removing from state", id.Name, id.AccountName, account.ResourceGroup)
		d.SetId("")
		return nil
	}

	d.Set("name", id.Name)
	d.Set("storage_account_name", id.AccountName)

	d.Set("container_access_type", flattenStorageContainerAccessLevel(props.AccessLevel))

	if err := d.Set("metadata", FlattenMetaData(props.MetaData)); err != nil {
		return fmt.Errorf("setting `metadata`: %+v", err)
	}

	d.Set("has_immutability_policy", props.HasImmutabilityPolicy)
	d.Set("has_legal_hold", props.HasLegalHold)

	resourceManagerId := parse.NewStorageContainerResourceManagerID(subscriptionId, account.ResourceGroup, id.AccountName, "default", id.Name)
	d.Set("resource_manager_id", resourceManagerId.ID())

	return nil
}

func resourceStorageContainerDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	storageClient := meta.(*clients.Client).Storage
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.StorageContainerDataPlaneID(d.Id())
	if err != nil {
		return err
	}

	account, err := storageClient.FindAccount(ctx, id.AccountName)
	if err != nil {
		return fmt.Errorf("retrieving Account %q for Container %q: %s", id.AccountName, id.Name, err)
	}
	if account == nil {
		return fmt.Errorf("Unable to locate Storage Account %q!", id.AccountName)
	}
	client, err := storageClient.ContainersClient(ctx, *account)
	if err != nil {
		return fmt.Errorf("building Containers Client for Storage Account %q (Resource Group %q): %s", id.AccountName, account.ResourceGroup, err)
	}

	if err := client.Delete(ctx, account.ResourceGroup, id.AccountName, id.Name); err != nil {
		return fmt.Errorf("deleting Container %q (Storage Account %q / Resource Group %q): %s", id.Name, id.AccountName, account.ResourceGroup, err)
	}

	return nil
}

func expandStorageContainerAccessLevel(input string) containers.AccessLevel {
	// for historical reasons, "private" above is an empty string in the API
	// so the enum doesn't 1:1 match. You could argue the SDK should handle this
	// but this is suitable for now
	if input == "private" {
		return containers.Private
	}

	return containers.AccessLevel(input)
}

func flattenStorageContainerAccessLevel(input containers.AccessLevel) string {
	// for historical reasons, "private" above is an empty string in the API
	if input == containers.Private {
		return "private"
	}

	return string(input)
}
