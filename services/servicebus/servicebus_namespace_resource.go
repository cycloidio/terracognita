package servicebus

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/servicebus/mgmt/2021-06-01-preview/servicebus"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/identity"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	keyVaultParse "github.com/hashicorp/terraform-provider-azurerm/services/keyvault/parse"
	keyVaultValidate "github.com/hashicorp/terraform-provider-azurerm/services/keyvault/validate"
	msiValidate "github.com/hashicorp/terraform-provider-azurerm/services/msi/validate"
	"github.com/hashicorp/terraform-provider-azurerm/services/servicebus/migration"
	"github.com/hashicorp/terraform-provider-azurerm/services/servicebus/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/servicebus/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

// Default Authorization Rule/Policy created by Azure, used to populate the
// default connection strings and keys
var (
	serviceBusNamespaceDefaultAuthorizationRule = "RootManageSharedAccessKey"
	serviceBusNamespaceResourceName             = "azurerm_servicebus_namespace"
)

func resourceServiceBusNamespace() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceServiceBusNamespaceCreateUpdate,
		Read:   resourceServiceBusNamespaceRead,
		Update: resourceServiceBusNamespaceCreateUpdate,
		Delete: resourceServiceBusNamespaceDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.NamespaceID(id)
			return err
		}),

		SchemaVersion: 1,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.NamespaceV0ToV1{},
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
				ValidateFunc: validate.NamespaceName,
			},

			"location": commonschema.Location(),

			"resource_group_name": commonschema.ResourceGroupName(),

			"identity": commonschema.SystemAssignedUserAssignedIdentityOptional(),

			"sku": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(servicebus.SkuNameBasic),
					string(servicebus.SkuNameStandard),
					string(servicebus.SkuNamePremium),
				}, false),
			},

			"capacity": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntInSlice([]int{0, 1, 2, 4, 8, 16}),
			},

			"customer_managed_key": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"key_vault_key_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: keyVaultValidate.NestedItemIdWithOptionalVersion,
						},

						"identity_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: msiValidate.UserAssignedIdentityID,
						},

						"infrastructure_encryption_enabled": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},

			"local_auth_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"default_primary_connection_string": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"default_secondary_connection_string": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"default_primary_key": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"default_secondary_key": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"zone_redundant": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"tags": tags.Schema(),
		},

		CustomizeDiff: pluginsdk.CustomizeDiffShim(func(ctx context.Context, diff *pluginsdk.ResourceDiff, v interface{}) error {
			oldCustomerManagedKey, newCustomerManagedKey := diff.GetChange("customer_managed_key")
			if len(oldCustomerManagedKey.([]interface{})) != 0 && len(newCustomerManagedKey.([]interface{})) == 0 {
				diff.ForceNew("customer_managed_key")
			}

			oldSku, newSku := diff.GetChange("sku")
			if diff.HasChange("sku") {
				if strings.EqualFold(newSku.(string), string(servicebus.SkuNamePremium)) || strings.EqualFold(oldSku.(string), string(servicebus.SkuNamePremium)) {
					log.Printf("[DEBUG] cannot migrate a namespace from or to Premium SKU")
					diff.ForceNew("sku")
				}
			}
			return nil
		}),
	}
}

func resourceServiceBusNamespaceCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.NamespacesClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for ServiceBus Namespace create/update.")

	location := azure.NormalizeLocation(d.Get("location").(string))
	sku := d.Get("sku").(string)
	t := d.Get("tags").(map[string]interface{})

	resourceId := parse.NewNamespaceID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceId.ResourceGroup, resourceId.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", resourceId, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_servicebus_namespace", resourceId.ID())
		}
	}

	identity, err := expandServiceBusNamespaceIdentity(d.Get("identity").([]interface{}))
	if err != nil {
		return fmt.Errorf("expanding `identity`: %+v", err)
	}

	parameters := servicebus.SBNamespace{
		Location: &location,
		Identity: identity,
		Sku: &servicebus.SBSku{
			Name: servicebus.SkuName(sku),
			Tier: servicebus.SkuTier(sku),
		},
		SBNamespaceProperties: &servicebus.SBNamespaceProperties{
			ZoneRedundant:    utils.Bool(d.Get("zone_redundant").(bool)),
			Encryption:       expandServiceBusNamespaceEncryption(d.Get("customer_managed_key").([]interface{})),
			DisableLocalAuth: utils.Bool(!d.Get("local_auth_enabled").(bool)),
		},
		Tags: tags.Expand(t),
	}

	if capacity := d.Get("capacity"); capacity != nil {
		if !strings.EqualFold(sku, string(servicebus.SkuNamePremium)) && capacity.(int) > 0 {
			return fmt.Errorf("Service Bus SKU %q only supports `capacity` of 0", sku)
		}
		if strings.EqualFold(sku, string(servicebus.SkuNamePremium)) && capacity.(int) == 0 {
			return fmt.Errorf("Service Bus SKU %q only supports `capacity` of 1, 2, 4, 8 or 16", sku)
		}
		parameters.Sku.Capacity = utils.Int32(int32(capacity.(int)))
	}

	future, err := client.CreateOrUpdate(ctx, resourceId.ResourceGroup, resourceId.Name, parameters)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %+v", resourceId, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for create/update of %s: %+v", resourceId, err)
	}

	d.SetId(resourceId.ID())
	return resourceServiceBusNamespaceRead(d, meta)
}

func resourceServiceBusNamespaceRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.NamespacesClient
	clientStable := meta.(*clients.Client).ServiceBus.NamespacesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.NamespaceID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	identity, err := flattenServiceBusNamespaceIdentity(resp.Identity)
	if err != nil {
		return fmt.Errorf("flattening `identity`: %+v", err)
	}
	if err := d.Set("identity", identity); err != nil {
		return fmt.Errorf("setting `identity`: %+v", err)
	}

	if sku := resp.Sku; sku != nil {
		skuName := ""
		// the Azure API is inconsistent here, so rewrite this into the casing we expect
		for _, v := range servicebus.PossibleSkuNameValues() {
			if strings.EqualFold(string(v), string(sku.Name)) {
				skuName = string(v)
			}
		}
		d.Set("sku", skuName)
		d.Set("capacity", sku.Capacity)
	}

	if properties := resp.SBNamespaceProperties; properties != nil {
		d.Set("zone_redundant", properties.ZoneRedundant)
		if customerManagedKey, err := flattenServiceBusNamespaceEncryption(properties.Encryption); err == nil {
			d.Set("customer_managed_key", customerManagedKey)
		}
		localAuthEnabled := true
		if properties.DisableLocalAuth != nil {
			localAuthEnabled = !*properties.DisableLocalAuth
		}
		d.Set("local_auth_enabled", localAuthEnabled)
	}

	keys, err := clientStable.ListKeys(ctx, id.ResourceGroup, id.Name, serviceBusNamespaceDefaultAuthorizationRule)
	if err != nil {
		log.Printf("[WARN] listing default keys for %s: %+v", id, err)
	} else {
		d.Set("default_primary_connection_string", keys.PrimaryConnectionString)
		d.Set("default_secondary_connection_string", keys.SecondaryConnectionString)
		d.Set("default_primary_key", keys.PrimaryKey)
		d.Set("default_secondary_key", keys.SecondaryKey)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceServiceBusNamespaceDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.NamespacesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.NamespaceID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("waiting for deletion of %s: %+v", id, err)
		}
	}

	return nil
}

func expandServiceBusNamespaceIdentity(input []interface{}) (*servicebus.Identity, error) {
	expanded, err := identity.ExpandSystemAndUserAssignedMap(input)
	if err != nil {
		return nil, err
	}

	out := servicebus.Identity{
		Type: servicebus.ManagedServiceIdentityType(string(expanded.Type)),
	}

	if len(expanded.IdentityIds) > 0 {
		out.UserAssignedIdentities = map[string]*servicebus.UserAssignedIdentity{}
		for id := range expanded.IdentityIds {
			out.UserAssignedIdentities[id] = &servicebus.UserAssignedIdentity{
				// intentionally empty
			}
		}
	}
	return &out, nil
}

func expandServiceBusNamespaceEncryption(input []interface{}) *servicebus.Encryption {
	if len(input) == 0 || input[0] == nil {
		return nil
	}
	v := input[0].(map[string]interface{})
	keyId, _ := keyVaultParse.ParseOptionallyVersionedNestedItemID(v["key_vault_key_id"].(string))
	return &servicebus.Encryption{
		KeyVaultProperties: &[]servicebus.KeyVaultProperties{
			{
				KeyName:     utils.String(keyId.Name),
				KeyVersion:  utils.String(keyId.Version),
				KeyVaultURI: utils.String(keyId.KeyVaultBaseUrl),
				Identity: &servicebus.UserAssignedIdentityProperties{
					UserAssignedIdentity: utils.String(v["identity_id"].(string)),
				},
			},
		},
		KeySource:                       servicebus.KeySourceMicrosoftKeyVault,
		RequireInfrastructureEncryption: utils.Bool(v["infrastructure_encryption_enabled"].(bool)),
	}
}

func flattenServiceBusNamespaceIdentity(input *servicebus.Identity) (*[]interface{}, error) {
	var transform *identity.SystemAndUserAssignedMap

	if input != nil {
		transform = &identity.SystemAndUserAssignedMap{
			Type:        identity.Type(string(input.Type)),
			IdentityIds: make(map[string]identity.UserAssignedIdentityDetails),
		}
		if input.PrincipalID != nil {
			transform.PrincipalId = *input.PrincipalID
		}
		if input.TenantID != nil {
			transform.TenantId = *input.TenantID
		}
		for k, v := range input.UserAssignedIdentities {
			transform.IdentityIds[k] = identity.UserAssignedIdentityDetails{
				ClientId:    v.ClientID,
				PrincipalId: v.PrincipalID,
			}
		}
	}

	return identity.FlattenSystemAndUserAssignedMap(transform)
}

func flattenServiceBusNamespaceEncryption(encryption *servicebus.Encryption) ([]interface{}, error) {
	if encryption == nil {
		return []interface{}{}, nil
	}

	var keyId string
	var identityId string
	if keyVaultProperties := encryption.KeyVaultProperties; keyVaultProperties != nil && len(*keyVaultProperties) != 0 {
		props := (*keyVaultProperties)[0]
		keyVaultKeyId, err := keyVaultParse.NewNestedItemID(*props.KeyVaultURI, "keys", *props.KeyName, *props.KeyVersion)
		if err != nil {
			return nil, fmt.Errorf("parsing `key_vault_key_id`: %+v", err)
		}
		keyId = keyVaultKeyId.ID()
		if props.Identity != nil && props.Identity.UserAssignedIdentity != nil {
			identityId = *props.Identity.UserAssignedIdentity
		}
	}

	return []interface{}{
		map[string]interface{}{
			"infrastructure_encryption_enabled": encryption.RequireInfrastructureEncryption,
			"key_vault_key_id":                  keyId,
			"identity_id":                       identityId,
		},
	}, nil
}
