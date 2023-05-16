package eventhub

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/locks"
	"github.com/hashicorp/terraform-provider-azurerm/services/eventhub/sdk/2021-01-01-preview/namespaces"
	keyVaultParse "github.com/hashicorp/terraform-provider-azurerm/services/keyvault/parse"
	keyVaultValidate "github.com/hashicorp/terraform-provider-azurerm/services/keyvault/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceEventHubNamespaceCustomerManagedKey() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceEventHubNamespaceCustomerManagedKeyCreateUpdate,
		Read:   resourceEventHubNamespaceCustomerManagedKeyRead,
		Update: resourceEventHubNamespaceCustomerManagedKeyCreateUpdate,
		Delete: resourceEventHubNamespaceCustomerManagedKeyDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceIdThen(func(id string) error {
			_, err := namespaces.ParseNamespaceID(id)
			return err
		}, func(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) ([]*pluginsdk.ResourceData, error) {
			client := meta.(*clients.Client).Eventhub.NamespacesClient

			var cancel context.CancelFunc
			ctx, cancel = timeouts.ForRead(ctx, d)
			defer cancel()

			id, err := namespaces.ParseNamespaceID(d.Id())
			if err != nil {
				return []*pluginsdk.ResourceData{d}, err
			}

			resp, err := client.Get(ctx, *id)
			if err != nil {
				return []*pluginsdk.ResourceData{d}, fmt.Errorf("retrieving %s: %+v", *id, err)
			}
			if resp.Model == nil || resp.Model.Properties == nil || resp.Model.Properties.Encryption == nil {
				return []*pluginsdk.ResourceData{d}, fmt.Errorf("retrieving %s: no customer managed key present", *id)
			}

			return []*pluginsdk.ResourceData{d}, nil
		}),

		Schema: map[string]*pluginsdk.Schema{
			"eventhub_namespace_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: namespaces.ValidateNamespaceID,
			},

			"key_vault_key_ids": {
				Type:     pluginsdk.TypeSet,
				Required: true,
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeString,
					ValidateFunc: keyVaultValidate.NestedItemIdWithOptionalVersion,
				},
			},
		},
	}
}

func resourceEventHubNamespaceCustomerManagedKeyCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Eventhub.NamespacesClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := namespaces.ParseNamespaceID(d.Get("eventhub_namespace_id").(string))
	if err != nil {
		return err
	}

	locks.ByName(id.NamespaceName, "azurerm_eventhub_namespace")
	defer locks.UnlockByName(id.NamespaceName, "azurerm_eventhub_namespace")

	resp, err := client.Get(ctx, *id)
	if err != nil {
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}
	if resp.Model == nil {
		return fmt.Errorf("retrieving %s: `model` was nil", *id)
	}

	if d.IsNewResource() {
		if resp.Model.Properties != nil && resp.Model.Properties.Encryption != nil {
			return tf.ImportAsExistsError("azurerm_eventhub_namespace_customer_managed_key", id.ID())
		}
	}

	namespace := resp.Model

	keySource := namespaces.KeySourceMicrosoftPointKeyVault
	namespace.Properties.Encryption = &namespaces.Encryption{
		KeySource: &keySource,
	}

	keyVaultProps, err := expandEventHubNamespaceKeyVaultKeyIds(d.Get("key_vault_key_ids").(*pluginsdk.Set).List())
	if err != nil {
		return err
	}
	namespace.Properties.Encryption.KeyVaultProperties = keyVaultProps

	if err := client.CreateOrUpdateThenPoll(ctx, *id, *namespace); err != nil {
		return fmt.Errorf("creating/updating %s: %+v", *id, err)
	}

	d.SetId(id.ID())

	return resourceEventHubNamespaceCustomerManagedKeyRead(d, meta)
}

func resourceEventHubNamespaceCustomerManagedKeyRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Eventhub.NamespacesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := namespaces.ParseNamespaceID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, *id)
	if err != nil {
		if response.WasNotFound(resp.HttpResponse) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}
	if resp.Model == nil {
		return fmt.Errorf("retrieving %s: `model` was nil", *id)
	}
	if resp.Model.Properties == nil && resp.Model.Properties.Encryption == nil {
		d.SetId("")
		return nil
	}

	d.Set("eventhub_namespace_id", id.ID())

	if props := resp.Model.Properties; props != nil {
		keyVaultKeyIds, err := flattenEventHubNamespaceKeyVaultKeyIds(props.Encryption)
		if err != nil {
			return err
		}

		d.Set("key_vault_key_ids", keyVaultKeyIds)
	}

	return nil
}

func resourceEventHubNamespaceCustomerManagedKeyDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	log.Printf(`[INFO] Customer Managed Keys cannot be removed from EventHub Namespaces once added. To remove the Customer Managed Key delete and recreate the parent EventHub Namespace")
`)
	return nil
}

func expandEventHubNamespaceKeyVaultKeyIds(input []interface{}) (*[]namespaces.KeyVaultProperties, error) {
	if len(input) == 0 {
		return nil, nil
	}

	results := make([]namespaces.KeyVaultProperties, 0)

	for _, item := range input {
		keyId, err := keyVaultParse.ParseOptionallyVersionedNestedItemID(item.(string))
		if err != nil {
			return nil, err
		}

		results = append(results, namespaces.KeyVaultProperties{
			KeyName:     utils.String(keyId.Name),
			KeyVaultUri: utils.String(keyId.KeyVaultBaseUrl),
			KeyVersion:  utils.String(keyId.Version),
		})
	}

	return &results, nil
}

func flattenEventHubNamespaceKeyVaultKeyIds(input *namespaces.Encryption) ([]interface{}, error) {
	results := make([]interface{}, 0)
	if input == nil || input.KeyVaultProperties == nil {
		return results, nil
	}

	for _, item := range *input.KeyVaultProperties {
		var keyName string
		if item.KeyName != nil {
			keyName = *item.KeyName
		}

		var keyVaultUri string
		if item.KeyVaultUri != nil {
			keyVaultUri = *item.KeyVaultUri
		}

		var keyVersion string
		if item.KeyVersion != nil {
			keyVersion = *item.KeyVersion
		}

		keyVaultKeyId, err := keyVaultParse.NewNestedItemID(keyVaultUri, "keys", keyName, keyVersion)
		if err != nil {
			return nil, err
		}

		results = append(results, keyVaultKeyId.ID())
	}

	return results, nil
}
