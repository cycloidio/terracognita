package storage

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-04-01/storage"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	keyVaultValidate "github.com/hashicorp/terraform-provider-azurerm/services/keyvault/validate"
	"github.com/hashicorp/terraform-provider-azurerm/services/storage/parse"
	storageValidate "github.com/hashicorp/terraform-provider-azurerm/services/storage/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceStorageEncryptionScope() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceStorageEncryptionScopeCreate,
		Read:   resourceStorageEncryptionScopeRead,
		Update: resourceStorageEncryptionScopeUpdate,
		Delete: resourceStorageEncryptionScopeDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.EncryptionScopeID(id)
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
				ValidateFunc: storageValidate.StorageEncryptionScopeName,
			},

			"storage_account_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: storageValidate.StorageAccountID,
			},

			"source": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(storage.EncryptionScopeSourceMicrosoftKeyVault),
					string(storage.EncryptionScopeSourceMicrosoftStorage),
				}, false),
			},

			"key_vault_key_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: keyVaultValidate.KeyVaultChildIDWithOptionalVersion,
			},

			"infrastructure_encryption_required": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceStorageEncryptionScopeCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage.EncryptionScopesClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	accountId, err := parse.StorageAccountID(d.Get("storage_account_id").(string))
	if err != nil {
		return err
	}

	resourceId := parse.NewEncryptionScopeID(accountId.SubscriptionId, accountId.ResourceGroup, accountId.Name, name).ID()
	existing, err := client.Get(ctx, accountId.ResourceGroup, accountId.Name, name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for present of existing Storage Encryption Scope %q (Storage Account Name %q / Resource Group %q): %+v", name, accountId.Name, accountId.ResourceGroup, err)
		}
	}
	if existing.EncryptionScopeProperties != nil && strings.EqualFold(string(existing.EncryptionScopeProperties.State), string(storage.EncryptionScopeStateEnabled)) {
		return tf.ImportAsExistsError("azurerm_storage_encryption_scope", resourceId)
	}

	if d.Get("source").(string) == string(storage.EncryptionScopeSourceMicrosoftKeyVault) {
		if _, ok := d.GetOk("key_vault_key_id"); !ok {
			return fmt.Errorf("`key_vault_key_id` is required when source is `%s`", string(storage.KeySourceMicrosoftKeyvault))
		}
	}

	props := storage.EncryptionScope{
		EncryptionScopeProperties: &storage.EncryptionScopeProperties{
			Source: storage.EncryptionScopeSource(d.Get("source").(string)),
			State:  storage.EncryptionScopeStateEnabled,
			KeyVaultProperties: &storage.EncryptionScopeKeyVaultProperties{
				KeyURI: utils.String(d.Get("key_vault_key_id").(string)),
			},
		},
	}

	if v, ok := d.GetOk("infrastructure_encryption_required"); ok {
		props.EncryptionScopeProperties.RequireInfrastructureEncryption = utils.Bool(v.(bool))
	}

	if _, err := client.Put(ctx, accountId.ResourceGroup, accountId.Name, name, props); err != nil {
		return fmt.Errorf("creating Storage Encryption Scope %q (Storage Account Name %q / Resource Group %q): %+v", name, accountId.Name, accountId.ResourceGroup, err)
	}

	d.SetId(resourceId)
	return resourceStorageEncryptionScopeRead(d, meta)
}

func resourceStorageEncryptionScopeUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage.EncryptionScopesClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.EncryptionScopeID(d.Id())
	if err != nil {
		return err
	}

	if d.Get("source").(string) == string(storage.EncryptionScopeSourceMicrosoftKeyVault) {
		if _, ok := d.GetOk("key_vault_key_id"); !ok {
			return fmt.Errorf("`key_vault_key_id` is required when source is `%s`", string(storage.KeySourceMicrosoftKeyvault))
		}
	}

	props := storage.EncryptionScope{
		EncryptionScopeProperties: &storage.EncryptionScopeProperties{
			Source: storage.EncryptionScopeSource(d.Get("source").(string)),
			State:  storage.EncryptionScopeStateEnabled,
			KeyVaultProperties: &storage.EncryptionScopeKeyVaultProperties{
				KeyURI: utils.String(d.Get("key_vault_key_id").(string)),
			},
		},
	}
	if _, err := client.Patch(ctx, id.ResourceGroup, id.StorageAccountName, id.Name, props); err != nil {
		return fmt.Errorf("updating Storage Encryption Scope %q (Storage Account Name %q / Resource Group %q): %+v", id.Name, id.StorageAccountName, id.ResourceGroup, err)
	}

	return resourceStorageEncryptionScopeRead(d, meta)
}

func resourceStorageEncryptionScopeRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage.EncryptionScopesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.EncryptionScopeID(d.Id())
	if err != nil {
		return err
	}
	accountId := parse.NewStorageAccountID(id.SubscriptionId, id.ResourceGroup, id.StorageAccountName)

	resp, err := client.Get(ctx, id.ResourceGroup, id.StorageAccountName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Storage Encryption Scope %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Storage Encryption Scope %q (Storage Account Name %q / Resource Group %q): %+v", id.Name, id.StorageAccountName, id.ResourceGroup, err)
	}

	if resp.EncryptionScopeProperties == nil {
		return fmt.Errorf("retrieving Storage Encryption Scope %q (Storage Account Name %q / Resource Group %q): `properties` was nil", id.Name, id.StorageAccountName, id.ResourceGroup)
	}

	props := *resp.EncryptionScopeProperties
	if strings.EqualFold(string(props.State), string(storage.EncryptionScopeStateDisabled)) {
		log.Printf("[INFO] Storage Encryption Scope %q (Storage Account Name %q / Resource Group %q) does not exist - removing from state", id.Name, id.StorageAccountName, id.ResourceGroup)
		d.SetId("")
		return nil
	}

	d.Set("name", resp.Name)
	d.Set("storage_account_id", accountId.ID())
	if props := resp.EncryptionScopeProperties; props != nil {
		d.Set("source", flattenEncryptionScopeSource(props.Source))
		var keyId string
		if kv := props.KeyVaultProperties; kv != nil {
			if kv.KeyURI != nil {
				keyId = *kv.KeyURI
			}
		}
		d.Set("key_vault_key_id", keyId)
		if props.RequireInfrastructureEncryption != nil {
			d.Set("infrastructure_encryption_required", props.RequireInfrastructureEncryption)
		}
	}

	return nil
}

func resourceStorageEncryptionScopeDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage.EncryptionScopesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.EncryptionScopeID(d.Id())
	if err != nil {
		return err
	}

	props := storage.EncryptionScope{
		EncryptionScopeProperties: &storage.EncryptionScopeProperties{
			State: storage.EncryptionScopeStateDisabled,
		},
	}

	if _, err = client.Put(ctx, id.ResourceGroup, id.StorageAccountName, id.Name, props); err != nil {
		return fmt.Errorf("disabling Storage Encryption Scope %q (Storage Account Name %q / Resource Group %q): %+v", id.Name, id.StorageAccountName, id.ResourceGroup, err)
	}

	return nil
}

func flattenEncryptionScopeSource(input storage.EncryptionScopeSource) string {
	// TODO: file a bug
	// the Storage API differs from every other API in Azure in that these Enum's can be returned case-insensitively
	if strings.EqualFold(string(input), string(storage.EncryptionScopeSourceMicrosoftKeyVault)) {
		return string(storage.EncryptionScopeSourceMicrosoftKeyVault)
	}

	return string(storage.EncryptionScopeSourceMicrosoftStorage)
}
