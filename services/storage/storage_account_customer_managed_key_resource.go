package storage

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-04-01/storage"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/locks"
	keyVaultParse "github.com/hashicorp/terraform-provider-azurerm/services/keyvault/parse"
	keyVaultValidate "github.com/hashicorp/terraform-provider-azurerm/services/keyvault/validate"
	msivalidate "github.com/hashicorp/terraform-provider-azurerm/services/msi/validate"
	storageParse "github.com/hashicorp/terraform-provider-azurerm/services/storage/parse"
	storageValidate "github.com/hashicorp/terraform-provider-azurerm/services/storage/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceStorageAccountCustomerManagedKey() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceStorageAccountCustomerManagedKeyCreateUpdate,
		Read:   resourceStorageAccountCustomerManagedKeyRead,
		Update: resourceStorageAccountCustomerManagedKeyCreateUpdate,
		Delete: resourceStorageAccountCustomerManagedKeyDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := storageParse.StorageAccountID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"storage_account_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: storageValidate.StorageAccountID,
			},

			"key_vault_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: keyVaultValidate.VaultID,
			},

			"key_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"key_version": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"user_assigned_identity_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: msivalidate.UserAssignedIdentityID,
			},
		},
	}
}

func resourceStorageAccountCustomerManagedKeyCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	storageClient := meta.(*clients.Client).Storage.AccountsClient
	keyVaultsClient := meta.(*clients.Client).KeyVault
	vaultsClient := keyVaultsClient.VaultsClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	storageAccountIDRaw := d.Get("storage_account_id").(string)
	storageAccountID, err := storageParse.StorageAccountID(storageAccountIDRaw)
	if err != nil {
		return err
	}

	locks.ByName(storageAccountID.Name, storageAccountResourceName)
	defer locks.UnlockByName(storageAccountID.Name, storageAccountResourceName)

	storageAccount, err := storageClient.GetProperties(ctx, storageAccountID.ResourceGroup, storageAccountID.Name, "")
	if err != nil {
		return fmt.Errorf("retrieving Storage Account %q (Resource Group %q): %+v", storageAccountID.Name, storageAccountID.ResourceGroup, err)
	}
	if storageAccount.AccountProperties == nil {
		return fmt.Errorf("retrieving Storage Account %q (Resource Group %q): `properties` was nil", storageAccountID.Name, storageAccountID.ResourceGroup)
	}

	// since we're mutating the storage account here, we can use that as the ID
	resourceID := storageAccountIDRaw

	if d.IsNewResource() {
		// whilst this looks superfluous given encryption is enabled by default, due to the way
		// the Azure API works this technically can be nil
		if storageAccount.AccountProperties.Encryption != nil {
			if storageAccount.AccountProperties.Encryption.KeySource == storage.KeySourceMicrosoftKeyvault {
				return tf.ImportAsExistsError("azurerm_storage_account_customer_managed_key", resourceID)
			}
		}
	}

	keyVaultID, err := keyVaultParse.VaultID(d.Get("key_vault_id").(string))
	if err != nil {
		return err
	}

	// If the Keyvault is in another subscription we need to update the client
	if keyVaultID.SubscriptionId != vaultsClient.SubscriptionID {
		vaultsClient = meta.(*clients.Client).KeyVault.KeyVaultClientForSubscription(keyVaultID.SubscriptionId)
	}

	keyVault, err := vaultsClient.Get(ctx, keyVaultID.ResourceGroup, keyVaultID.Name)
	if err != nil {
		return fmt.Errorf("retrieving Key Vault %q (Resource Group %q): %+v", keyVaultID.Name, keyVaultID.ResourceGroup, err)
	}

	softDeleteEnabled := false
	purgeProtectionEnabled := false
	if props := keyVault.Properties; props != nil {
		if esd := props.EnableSoftDelete; esd != nil {
			softDeleteEnabled = *esd
		}
		if epp := props.EnablePurgeProtection; epp != nil {
			purgeProtectionEnabled = *epp
		}
	}
	if !softDeleteEnabled || !purgeProtectionEnabled {
		return fmt.Errorf("Key Vault %q (Resource Group %q) must be configured for both Purge Protection and Soft Delete", keyVaultID.Name, keyVaultID.ResourceGroup)
	}

	keyVaultBaseURL, err := keyVaultsClient.BaseUriForKeyVault(ctx, *keyVaultID)
	if err != nil {
		return fmt.Errorf("looking up Key Vault URI from Key Vault %q (Resource Group %q) (Subscription %q): %+v", keyVaultID.Name, keyVaultID.ResourceGroup, keyVaultsClient.VaultsClient.SubscriptionID, err)
	}

	keyName := d.Get("key_name").(string)
	keyVersion := d.Get("key_version").(string)
	userAssignedIdentity := d.Get("user_assigned_identity_id").(string)

	props := storage.AccountUpdateParameters{
		AccountPropertiesUpdateParameters: &storage.AccountPropertiesUpdateParameters{
			Encryption: &storage.Encryption{
				Services: &storage.EncryptionServices{
					Blob: &storage.EncryptionService{
						Enabled: utils.Bool(true),
					},
					File: &storage.EncryptionService{
						Enabled: utils.Bool(true),
					},
				},
				EncryptionIdentity: &storage.EncryptionIdentity{
					EncryptionUserAssignedIdentity: utils.String(userAssignedIdentity),
				},
				KeySource: storage.KeySourceMicrosoftKeyvault,
				KeyVaultProperties: &storage.KeyVaultProperties{
					KeyName:     utils.String(keyName),
					KeyVersion:  utils.String(keyVersion),
					KeyVaultURI: utils.String(*keyVaultBaseURL),
				},
			},
		},
	}

	if _, err = storageClient.Update(ctx, storageAccountID.ResourceGroup, storageAccountID.Name, props); err != nil {
		return fmt.Errorf("updating Customer Managed Key for Storage Account %q (Resource Group %q): %+v", storageAccountID.Name, storageAccountID.ResourceGroup, err)
	}

	d.SetId(resourceID)
	return resourceStorageAccountCustomerManagedKeyRead(d, meta)
}

func resourceStorageAccountCustomerManagedKeyRead(d *pluginsdk.ResourceData, meta interface{}) error {
	storageClient := meta.(*clients.Client).Storage.AccountsClient
	keyVaultsClient := meta.(*clients.Client).KeyVault
	resourcesClient := meta.(*clients.Client).Resource
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	storageAccountID, err := storageParse.StorageAccountID(d.Id())
	if err != nil {
		return err
	}

	storageAccount, err := storageClient.GetProperties(ctx, storageAccountID.ResourceGroup, storageAccountID.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(storageAccount.Response) {
			log.Printf("[DEBUG] Storage Account %q could not be found in Resource Group %q - removing from state!", storageAccountID.Name, storageAccountID.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Storage Account %q (Resource Group %q): %+v", storageAccountID.Name, storageAccountID.ResourceGroup, err)
	}
	if storageAccount.AccountProperties == nil {
		return fmt.Errorf("retrieving Storage Account %q (Resource Group %q): `properties` was nil", storageAccountID.Name, storageAccountID.ResourceGroup)
	}
	if storageAccount.AccountProperties.Encryption == nil || storageAccount.AccountProperties.Encryption.KeySource != storage.KeySourceMicrosoftKeyvault {
		log.Printf("[DEBUG] Customer Managed Key was not defined for Storage Account %q (Resource Group %q) - removing from state!", storageAccountID.Name, storageAccountID.ResourceGroup)
		d.SetId("")
		return nil
	}

	encryption := *storageAccount.AccountProperties.Encryption

	keyName := ""
	keyVaultURI := ""
	keyVersion := ""
	if props := encryption.KeyVaultProperties; props != nil {
		if props.KeyName != nil {
			keyName = *props.KeyName
		}
		if props.KeyVaultURI != nil {
			keyVaultURI = *props.KeyVaultURI
		}
		if props.KeyVersion != nil {
			keyVersion = *props.KeyVersion
		}
	}

	userAssignedIdentity := ""
	if props := encryption.EncryptionIdentity; props != nil {
		if props.EncryptionUserAssignedIdentity != nil {
			userAssignedIdentity = *props.EncryptionUserAssignedIdentity
		}
	}

	if keyVaultURI == "" {
		return fmt.Errorf("retrieving Storage Account %q (Resource Group %q): `properties.encryption.keyVaultProperties.keyVaultURI` was nil", storageAccountID.Name, storageAccountID.ResourceGroup)
	}

	keyVaultID, err := keyVaultsClient.KeyVaultIDFromBaseUrl(ctx, resourcesClient, keyVaultURI)
	if err != nil {
		return fmt.Errorf("retrieving Key Vault ID from the Base URI %q: %+v", keyVaultURI, err)
	}

	// now we have the key vault uri we can look up the ID

	d.Set("storage_account_id", d.Id())
	d.Set("key_vault_id", keyVaultID)
	d.Set("key_name", keyName)
	d.Set("key_version", keyVersion)
	d.Set("user_assigned_identity_id", userAssignedIdentity)

	return nil
}

func resourceStorageAccountCustomerManagedKeyDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	storageClient := meta.(*clients.Client).Storage.AccountsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	storageAccountID, err := storageParse.StorageAccountID(d.Id())
	if err != nil {
		return err
	}

	locks.ByName(storageAccountID.Name, storageAccountResourceName)
	defer locks.UnlockByName(storageAccountID.Name, storageAccountResourceName)

	// confirm it still exists prior to trying to update it, else we'll get an error
	storageAccount, err := storageClient.GetProperties(ctx, storageAccountID.ResourceGroup, storageAccountID.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(storageAccount.Response) {
			return nil
		}

		return fmt.Errorf("retrieving Storage Account %q (Resource Group %q): %+v", storageAccountID.Name, storageAccountID.ResourceGroup, err)
	}

	// Since this isn't a real object, just modifying an existing object
	// "Delete" doesn't really make sense it should really be a "Revert to Default"
	// So instead of the Delete func actually deleting the Storage Account I am
	// making it reset the Storage Account to its default state
	props := storage.AccountUpdateParameters{
		AccountPropertiesUpdateParameters: &storage.AccountPropertiesUpdateParameters{
			Encryption: &storage.Encryption{
				Services: &storage.EncryptionServices{
					Blob: &storage.EncryptionService{
						Enabled: utils.Bool(true),
					},
					File: &storage.EncryptionService{
						Enabled: utils.Bool(true),
					},
				},
				KeySource: storage.KeySourceMicrosoftStorage,
			},
		},
	}

	if _, err = storageClient.Update(ctx, storageAccountID.ResourceGroup, storageAccountID.Name, props); err != nil {
		return fmt.Errorf("removing Customer Managed Key for Storage Account %q (Resource Group %q): %+v", storageAccountID.Name, storageAccountID.ResourceGroup, err)
	}

	return nil
}
