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
	"github.com/tombuildsstuff/giovanni/storage/2019-12-12/table/tables"
)

func resourceStorageTable() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceStorageTableCreate,
		Read:   resourceStorageTableRead,
		Delete: resourceStorageTableDelete,
		Update: resourceStorageTableUpdate,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.StorageTableDataPlaneID(id)
			return err
		}),

		SchemaVersion: 2,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.TableV0ToV1{},
			1: migration.TableV1ToV2{},
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
				ValidateFunc: validate.StorageTableName,
			},

			"storage_account_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.StorageAccountName,
			},

			"acl": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 64),
						},
						"access_policy": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"start": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"expiry": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"permissions": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceStorageTableCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	storageClient := meta.(*clients.Client).Storage

	tableName := d.Get("name").(string)
	accountName := d.Get("storage_account_name").(string)
	aclsRaw := d.Get("acl").(*pluginsdk.Set).List()
	acls := expandStorageTableACLs(aclsRaw)

	account, err := storageClient.FindAccount(ctx, accountName)
	if err != nil {
		return fmt.Errorf("retrieving Account %q for Table %q: %s", accountName, tableName, err)
	}
	if account == nil {
		return fmt.Errorf("unable to locate Storage Account %q!", accountName)
	}

	client, err := storageClient.TablesClient(ctx, *account)
	if err != nil {
		return fmt.Errorf("building Table Client: %s", err)
	}

	id := parse.NewStorageTableDataPlaneId(accountName, storageClient.Environment.StorageEndpointSuffix, tableName).ID()

	exists, err := client.Exists(ctx, account.ResourceGroup, accountName, tableName)
	if err != nil {
		return fmt.Errorf("checking for existence of existing Storage Table %q (Account %q / Resource Group %q): %+v", tableName, accountName, account.ResourceGroup, err)
	}
	if exists != nil && *exists {
		return tf.ImportAsExistsError("azurerm_storage_table", id)
	}

	log.Printf("[DEBUG] Creating Table %q in Storage Account %q.", tableName, accountName)
	if err := client.Create(ctx, account.ResourceGroup, accountName, tableName); err != nil {
		return fmt.Errorf("creating Table %q within Storage Account %q: %s", tableName, accountName, err)
	}

	d.SetId(id)
	if err := client.UpdateACLs(ctx, account.ResourceGroup, accountName, tableName, acls); err != nil {
		return fmt.Errorf("setting ACL's for Storage Table %q (Account %q / Resource Group %q): %+v", tableName, accountName, account.ResourceGroup, err)
	}

	return resourceStorageTableRead(d, meta)
}

func resourceStorageTableRead(d *pluginsdk.ResourceData, meta interface{}) error {
	storageClient := meta.(*clients.Client).Storage
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.StorageTableDataPlaneID(d.Id())
	if err != nil {
		return err
	}

	account, err := storageClient.FindAccount(ctx, id.AccountName)
	if err != nil {
		return fmt.Errorf("retrieving Account %q for Table %q: %s", id.AccountName, id.Name, err)
	}
	if account == nil {
		log.Printf("Unable to determine Resource Group for Storage Storage Table %q (Account %s) - assuming removed & removing from state", id.Name, id.AccountName)
		d.SetId("")
		return nil
	}

	client, err := storageClient.TablesClient(ctx, *account)
	if err != nil {
		return fmt.Errorf("building Table Client: %s", err)
	}

	exists, err := client.Exists(ctx, account.ResourceGroup, id.AccountName, id.Name)
	if err != nil {
		return fmt.Errorf("retrieving Table %q (Storage Account %q / Resource Group %q): %s", id.Name, id.AccountName, account.ResourceGroup, err)
	}
	if exists == nil || !*exists {
		log.Printf("[DEBUG] Storage Account %q not found, removing table %q from state", id.AccountName, id.Name)
		d.SetId("")
		return nil
	}

	acls, err := client.GetACLs(ctx, account.ResourceGroup, id.AccountName, id.Name)
	if err != nil {
		return fmt.Errorf("retrieving ACL's %q in Storage Account %q: %s", id.Name, id.AccountName, err)
	}

	d.Set("name", id.Name)
	d.Set("storage_account_name", id.AccountName)

	if err := d.Set("acl", flattenStorageTableACLs(acls)); err != nil {
		return fmt.Errorf("flattening `acl`: %+v", err)
	}

	return nil
}

func resourceStorageTableDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	storageClient := meta.(*clients.Client).Storage
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.StorageTableDataPlaneID(d.Id())
	if err != nil {
		return err
	}

	account, err := storageClient.FindAccount(ctx, id.AccountName)
	if err != nil {
		return fmt.Errorf("retrieving Account %q for Table %q: %s", id.AccountName, id.Name, err)
	}
	if account == nil {
		return fmt.Errorf("Unable to locate Storage Account %q!", id.AccountName)
	}

	client, err := storageClient.TablesClient(ctx, *account)
	if err != nil {
		return fmt.Errorf("building Table Client: %s", err)
	}

	log.Printf("[INFO] Deleting Table %q in Storage Account %q", id.Name, id.AccountName)
	if err := client.Delete(ctx, account.ResourceGroup, id.AccountName, id.Name); err != nil {
		return fmt.Errorf("deleting Table %q from Storage Account %q: %s", id.Name, id.AccountName, err)
	}

	return nil
}

func resourceStorageTableUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	storageClient := meta.(*clients.Client).Storage
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.StorageTableDataPlaneID(d.Id())
	if err != nil {
		return err
	}

	account, err := storageClient.FindAccount(ctx, id.AccountName)
	if err != nil {
		return fmt.Errorf("retrieving Account %q for Table %q: %s", id.AccountName, id.Name, err)
	}
	if account == nil {
		return fmt.Errorf("unable to locate Storage Account %q!", id.AccountName)
	}

	client, err := storageClient.TablesClient(ctx, *account)
	if err != nil {
		return fmt.Errorf("building Table Client: %s", err)
	}

	if d.HasChange("acl") {
		log.Printf("[DEBUG] Updating the ACL's for Storage Table %q (Storage Account %q)", id.Name, id.AccountName)

		aclsRaw := d.Get("acl").(*pluginsdk.Set).List()
		acls := expandStorageTableACLs(aclsRaw)

		if err := client.UpdateACLs(ctx, account.ResourceGroup, id.AccountName, id.Name, acls); err != nil {
			return fmt.Errorf("updating ACL's for Table %q (Storage Account %q): %s", id.Name, id.AccountName, err)
		}

		log.Printf("[DEBUG] Updated the ACL's for Storage Table %q (Storage Account %q)", id.Name, id.AccountName)
	}

	return resourceStorageTableRead(d, meta)
}

func expandStorageTableACLs(input []interface{}) []tables.SignedIdentifier {
	results := make([]tables.SignedIdentifier, 0)

	for _, v := range input {
		vals := v.(map[string]interface{})

		policies := vals["access_policy"].([]interface{})
		policy := policies[0].(map[string]interface{})

		identifier := tables.SignedIdentifier{
			Id: vals["id"].(string),
			AccessPolicy: tables.AccessPolicy{
				Start:      policy["start"].(string),
				Expiry:     policy["expiry"].(string),
				Permission: policy["permissions"].(string),
			},
		}
		results = append(results, identifier)
	}

	return results
}

func flattenStorageTableACLs(input *[]tables.SignedIdentifier) []interface{} {
	result := make([]interface{}, 0)
	if input == nil {
		return result
	}

	for _, v := range *input {
		output := map[string]interface{}{
			"id": v.Id,
			"access_policy": []interface{}{
				map[string]interface{}{
					"start":       v.AccessPolicy.Start,
					"expiry":      v.AccessPolicy.Expiry,
					"permissions": v.AccessPolicy.Permission,
				},
			},
		}

		result = append(result, output)
	}

	return result
}
