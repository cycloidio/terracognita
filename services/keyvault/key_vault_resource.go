package keyvault

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/mgmt/2021-10-01/keyvault"
	KeyVaultMgmt "github.com/Azure/azure-sdk-for-go/services/keyvault/v7.1/keyvault"
	"github.com/gofrs/uuid"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	commonValidate "github.com/hashicorp/terraform-provider-azurerm/helpers/validate"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/locks"
	"github.com/hashicorp/terraform-provider-azurerm/services/keyvault/migration"
	"github.com/hashicorp/terraform-provider-azurerm/services/keyvault/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/keyvault/validate"
	"github.com/hashicorp/terraform-provider-azurerm/services/network"
	networkParse "github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/set"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

// As can be seen in the API definition, the Sku Family only supports the value
// `A` and is a required field
// https://github.com/Azure/azure-rest-api-specs/blob/master/arm-keyvault/2015-06-01/swagger/keyvault.json#L239
var armKeyVaultSkuFamily = "A"

var keyVaultResourceName = "azurerm_key_vault"

func resourceKeyVault() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceKeyVaultCreate,
		Read:   resourceKeyVaultRead,
		Update: resourceKeyVaultUpdate,
		Delete: resourceKeyVaultDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.VaultID(id)
			return err
		}),

		SchemaVersion: 2,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.KeyVaultV0ToV1{},
			1: migration.KeyVaultV1ToV2{},
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
				ValidateFunc: validate.VaultName,
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"sku_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(keyvault.SkuNameStandard),
					string(keyvault.SkuNamePremium),
				}, false),
			},

			"tenant_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},

			"access_policy": {
				Type:       pluginsdk.TypeList,
				ConfigMode: pluginsdk.SchemaConfigModeAttr,
				Optional:   true,
				Computed:   true,
				MaxItems:   1024,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"tenant_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.IsUUID,
						},
						"object_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.IsUUID,
						},
						"application_id": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: validate.IsUUIDOrEmpty,
						},
						"certificate_permissions": schemaCertificatePermissions(),
						"key_permissions":         schemaKeyPermissions(),
						"secret_permissions":      schemaSecretPermissions(),
						"storage_permissions":     schemaStoragePermissions(),
					},
				},
			},

			"enabled_for_deployment": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			"enabled_for_disk_encryption": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			"enabled_for_template_deployment": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			// TODO 4.0: change this from enable_* to *_enabled
			"enable_rbac_authorization": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			"network_acls": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"default_action": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(keyvault.NetworkRuleActionAllow),
								string(keyvault.NetworkRuleActionDeny),
							}, false),
						},
						"bypass": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(keyvault.NetworkRuleBypassOptionsNone),
								string(keyvault.NetworkRuleBypassOptionsAzureServices),
							}, false),
						},
						"ip_rules": {
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
								ValidateFunc: validation.Any(
									commonValidate.IPv4Address,
									commonValidate.CIDR,
								),
							},
							Set: set.HashIPv4AddressOrCIDR,
						},
						"virtual_network_subnet_ids": {
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
							Set:      set.HashStringIgnoreCase,
						},
					},
				},
			},

			"purge_protection_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			"soft_delete_retention_days": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Default:      90,
				ValidateFunc: validation.IntBetween(7, 90),
			},

			"contact": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"email": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
						"name": {
							Type:     pluginsdk.TypeString,
							Optional: true,
						},
						"phone": {
							Type:     pluginsdk.TypeString,
							Optional: true,
						},
					},
				},
			},

			"tags": tags.Schema(),

			// Computed
			"vault_uri": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceKeyVaultCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).KeyVault.VaultsClient
	dataPlaneClient := meta.(*clients.Client).KeyVault.ManagementClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewVaultID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	location := azure.NormalizeLocation(d.Get("location").(string))

	// Locking this resource so we don't make modifications to it at the same time if there is a
	// key vault access policy trying to update it as well
	locks.ByName(id.Name, keyVaultResourceName)
	defer locks.UnlockByName(id.Name, keyVaultResourceName)

	// check for the presence of an existing, live one which should be imported into the state
	existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
		}
	}

	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurerm_key_vault", id.ID())
	}

	// before creating check to see if the key vault exists in the soft delete state
	softDeletedKeyVault, err := client.GetDeleted(ctx, id.Name, location)
	if err != nil {
		// If Terraform lacks permission to read at the Subscription we'll get 409, not 404
		if !utils.ResponseWasNotFound(softDeletedKeyVault.Response) && !utils.ResponseWasForbidden(softDeletedKeyVault.Response) {
			return fmt.Errorf("checking for the presence of an existing Soft-Deleted Key Vault %q (Location %q): %+v", id.Name, location, err)
		}
	}

	// if so, does the user want us to recover it?

	recoverSoftDeletedKeyVault := false
	if !utils.ResponseWasNotFound(softDeletedKeyVault.Response) && !utils.ResponseWasForbidden(softDeletedKeyVault.Response) {
		if !meta.(*clients.Client).Features.KeyVault.RecoverSoftDeletedKeyVaults {
			// this exists but the users opted out so they must import this it out-of-band
			return fmt.Errorf(optedOutOfRecoveringSoftDeletedKeyVaultErrorFmt(id.Name, location))
		}

		recoverSoftDeletedKeyVault = true
	}

	tenantUUID := uuid.FromStringOrNil(d.Get("tenant_id").(string))
	enabledForDeployment := d.Get("enabled_for_deployment").(bool)
	enabledForDiskEncryption := d.Get("enabled_for_disk_encryption").(bool)
	enabledForTemplateDeployment := d.Get("enabled_for_template_deployment").(bool)
	enableRbacAuthorization := d.Get("enable_rbac_authorization").(bool)
	t := d.Get("tags").(map[string]interface{})

	policies := d.Get("access_policy").([]interface{})
	accessPolicies := expandAccessPolicies(policies)

	networkAclsRaw := d.Get("network_acls").([]interface{})
	networkAcls, subnetIds := expandKeyVaultNetworkAcls(networkAclsRaw)

	sku := keyvault.Sku{
		Family: &armKeyVaultSkuFamily,
		Name:   keyvault.SkuName(d.Get("sku_name").(string)),
	}

	parameters := keyvault.VaultCreateOrUpdateParameters{
		Location: &location,
		Properties: &keyvault.VaultProperties{
			TenantID:                     &tenantUUID,
			Sku:                          &sku,
			AccessPolicies:               accessPolicies,
			EnabledForDeployment:         &enabledForDeployment,
			EnabledForDiskEncryption:     &enabledForDiskEncryption,
			EnabledForTemplateDeployment: &enabledForTemplateDeployment,
			EnableRbacAuthorization:      &enableRbacAuthorization,
			NetworkAcls:                  networkAcls,

			// @tombuildsstuff: as of 2020-12-15 this is now defaulted on, and appears to be so in all regions
			// This has been confirmed in Azure Public and Azure China - but I couldn't find any more
			// documentation with further details
			EnableSoftDelete: utils.Bool(true),
		},
		Tags: tags.Expand(t),
	}

	if purgeProtectionEnabled := d.Get("purge_protection_enabled").(bool); purgeProtectionEnabled {
		parameters.Properties.EnablePurgeProtection = utils.Bool(purgeProtectionEnabled)
	}

	if v := d.Get("soft_delete_retention_days"); v != 90 {
		parameters.Properties.SoftDeleteRetentionInDays = utils.Int32(int32(v.(int)))
	}

	if recoverSoftDeletedKeyVault {
		parameters.Properties.CreateMode = keyvault.CreateModeRecover
	}

	// also lock on the Virtual Network ID's since modifications in the networking stack are exclusive
	virtualNetworkNames := make([]string, 0)
	for _, v := range subnetIds {
		id, err := networkParse.SubnetIDInsensitively(v)
		if err != nil {
			return err
		}
		if !utils.SliceContainsValue(virtualNetworkNames, id.VirtualNetworkName) {
			virtualNetworkNames = append(virtualNetworkNames, id.VirtualNetworkName)
		}
	}

	locks.MultipleByName(&virtualNetworkNames, network.VirtualNetworkResourceName)
	defer locks.UnlockMultipleByName(&virtualNetworkNames, network.VirtualNetworkResourceName)

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, parameters)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}
	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation/update of %q: %+v", id, err)
	}

	read, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}
	if read.Properties == nil || read.Properties.VaultURI == nil {
		return fmt.Errorf("retrieving %s: `properties.VaultUri` was nil", id)
	}
	d.SetId(id.ID())
	meta.(*clients.Client).KeyVault.AddToCache(id, *read.Properties.VaultURI)

	if props := read.Properties; props != nil {
		if vault := props.VaultURI; vault != nil {
			log.Printf("[DEBUG] Waiting for %s to become available", id)
			stateConf := &pluginsdk.StateChangeConf{
				Pending:                   []string{"pending"},
				Target:                    []string{"available"},
				Refresh:                   keyVaultRefreshFunc(*vault),
				Delay:                     30 * time.Second,
				PollInterval:              10 * time.Second,
				ContinuousTargetOccurence: 10,
				Timeout:                   d.Timeout(pluginsdk.TimeoutCreate),
			}

			if _, err := stateConf.WaitForStateContext(ctx); err != nil {
				return fmt.Errorf("waiting for %s to become available: %s", id, err)
			}
		}
	}

	if v, ok := d.GetOk("contact"); ok {
		contacts := KeyVaultMgmt.Contacts{
			ContactList: expandKeyVaultCertificateContactList(v.(*pluginsdk.Set).List()),
		}
		if read.Properties == nil || read.Properties.VaultURI == nil {
			return fmt.Errorf("failed to get vault base url for %s: %s", id, err)
		}
		if _, err := dataPlaneClient.SetCertificateContacts(ctx, *read.Properties.VaultURI, contacts); err != nil {
			return fmt.Errorf("failed to set Contacts for %s: %+v", id, err)
		}
	}

	return resourceKeyVaultRead(d, meta)
}

func resourceKeyVaultUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).KeyVault.VaultsClient
	managementClient := meta.(*clients.Client).KeyVault.ManagementClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VaultID(d.Id())
	if err != nil {
		return err
	}

	// Locking this resource so we don't make modifications to it at the same time if there is a
	// key vault access policy trying to update it as well
	locks.ByName(id.Name, keyVaultResourceName)
	defer locks.UnlockByName(id.Name, keyVaultResourceName)

	d.Partial(true)

	// first pull the existing key vault since we need to lock on several bits of its information
	existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}
	if existing.Properties == nil {
		return fmt.Errorf("retrieving %s: `properties` was nil", *id)
	}

	update := keyvault.VaultPatchParameters{}

	if d.HasChange("access_policy") {
		if update.Properties == nil {
			update.Properties = &keyvault.VaultPatchProperties{}
		}

		policiesRaw := d.Get("access_policy").([]interface{})
		accessPolicies := expandAccessPolicies(policiesRaw)
		update.Properties.AccessPolicies = accessPolicies
	}

	if d.HasChange("enabled_for_deployment") {
		if update.Properties == nil {
			update.Properties = &keyvault.VaultPatchProperties{}
		}

		update.Properties.EnabledForDeployment = utils.Bool(d.Get("enabled_for_deployment").(bool))
	}

	if d.HasChange("enabled_for_disk_encryption") {
		if update.Properties == nil {
			update.Properties = &keyvault.VaultPatchProperties{}
		}

		update.Properties.EnabledForDiskEncryption = utils.Bool(d.Get("enabled_for_disk_encryption").(bool))
	}

	if d.HasChange("enabled_for_template_deployment") {
		if update.Properties == nil {
			update.Properties = &keyvault.VaultPatchProperties{}
		}

		update.Properties.EnabledForTemplateDeployment = utils.Bool(d.Get("enabled_for_template_deployment").(bool))
	}

	if d.HasChange("enable_rbac_authorization") {
		if update.Properties == nil {
			update.Properties = &keyvault.VaultPatchProperties{}
		}

		update.Properties.EnableRbacAuthorization = utils.Bool(d.Get("enable_rbac_authorization").(bool))
	}

	if d.HasChange("network_acls") {
		if update.Properties == nil {
			update.Properties = &keyvault.VaultPatchProperties{}
		}

		networkAclsRaw := d.Get("network_acls").([]interface{})
		networkAcls, subnetIds := expandKeyVaultNetworkAcls(networkAclsRaw)

		// also lock on the Virtual Network ID's since modifications in the networking stack are exclusive
		virtualNetworkNames := make([]string, 0)
		for _, v := range subnetIds {
			id, err := networkParse.SubnetIDInsensitively(v)
			if err != nil {
				return err
			}

			if !utils.SliceContainsValue(virtualNetworkNames, id.VirtualNetworkName) {
				virtualNetworkNames = append(virtualNetworkNames, id.VirtualNetworkName)
			}
		}

		locks.MultipleByName(&virtualNetworkNames, network.VirtualNetworkResourceName)
		defer locks.UnlockMultipleByName(&virtualNetworkNames, network.VirtualNetworkResourceName)

		update.Properties.NetworkAcls = networkAcls
	}

	if d.HasChange("purge_protection_enabled") {
		if update.Properties == nil {
			update.Properties = &keyvault.VaultPatchProperties{}
		}

		newValue := d.Get("purge_protection_enabled").(bool)

		// existing.Properties guaranteed non-nil above
		oldValue := false
		if existing.Properties.EnablePurgeProtection != nil {
			oldValue = *existing.Properties.EnablePurgeProtection
		}

		// whilst this should have got caught in the customizeDiff this won't work if that fields interpolated
		// hence the double-checking here
		if oldValue && !newValue {
			return fmt.Errorf("updating %s: once Purge Protection has been Enabled it's not possible to disable it", *id)
		}

		update.Properties.EnablePurgeProtection = utils.Bool(newValue)
	}

	if d.HasChange("sku_name") {
		if update.Properties == nil {
			update.Properties = &keyvault.VaultPatchProperties{}
		}

		update.Properties.Sku = &keyvault.Sku{
			Family: &armKeyVaultSkuFamily,
			Name:   keyvault.SkuName(d.Get("sku_name").(string)),
		}
	}

	if d.HasChange("soft_delete_retention_days") {
		if update.Properties == nil {
			update.Properties = &keyvault.VaultPatchProperties{}
		}

		// existing.Properties guaranteed non-nil above
		var oldValue int32 = 0
		if existing.Properties.SoftDeleteRetentionInDays != nil {
			oldValue = *existing.Properties.SoftDeleteRetentionInDays
		}

		// whilst this should have got caught in the customizeDiff this won't work if that fields interpolated
		// hence the double-checking here
		if oldValue != 0 {
			// Code="BadRequest" Message="The property \"softDeleteRetentionInDays\" has been set already and it can't be modified."
			return fmt.Errorf("updating %s: once `soft_delete_retention_days` has been configured it cannot be modified", *id)
		}

		update.Properties.SoftDeleteRetentionInDays = utils.Int32(int32(d.Get("soft_delete_retention_days").(int)))
	}

	if d.HasChange("tenant_id") {
		if update.Properties == nil {
			update.Properties = &keyvault.VaultPatchProperties{}
		}

		tenantUUID := uuid.FromStringOrNil(d.Get("tenant_id").(string))
		update.Properties.TenantID = &tenantUUID
	}

	if d.HasChange("tags") {
		t := d.Get("tags").(map[string]interface{})
		update.Tags = tags.Expand(t)
	}

	if _, err := client.Update(ctx, id.ResourceGroup, id.Name, update); err != nil {
		return fmt.Errorf("updating %s: %+v", *id, err)
	}

	if d.HasChange("contact") {
		contacts := KeyVaultMgmt.Contacts{
			ContactList: expandKeyVaultCertificateContactList(d.Get("contact").(*pluginsdk.Set).List()),
		}
		if existing.Properties == nil || existing.Properties.VaultURI == nil {
			return fmt.Errorf("failed to get vault base url for %s: %s", *id, err)
		}

		var err error
		if len(*contacts.ContactList) == 0 {
			_, err = managementClient.DeleteCertificateContacts(ctx, *existing.Properties.VaultURI)
		} else {
			_, err = managementClient.SetCertificateContacts(ctx, *existing.Properties.VaultURI, contacts)
		}

		if err != nil {
			return fmt.Errorf("setting Contacts for %s: %+v", *id, err)
		}
	}

	d.Partial(false)

	return resourceKeyVaultRead(d, meta)
}

func resourceKeyVaultRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).KeyVault.VaultsClient
	managementClient := meta.(*clients.Client).KeyVault.ManagementClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VaultID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] %s was not found - removing from state!", *id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}
	if resp.Properties == nil {
		return fmt.Errorf("retrieving %s: `properties` was nil", *id)
	}
	if resp.Properties.VaultURI == nil {
		return fmt.Errorf("retrieving %s: `properties.VaultUri` was nil", *id)
	}

	props := *resp.Properties
	meta.(*clients.Client).KeyVault.AddToCache(*id, *resp.Properties.VaultURI)

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	d.Set("tenant_id", props.TenantID.String())
	d.Set("enabled_for_deployment", props.EnabledForDeployment)
	d.Set("enabled_for_disk_encryption", props.EnabledForDiskEncryption)
	d.Set("enabled_for_template_deployment", props.EnabledForTemplateDeployment)
	d.Set("enable_rbac_authorization", props.EnableRbacAuthorization)
	d.Set("purge_protection_enabled", props.EnablePurgeProtection)
	d.Set("vault_uri", props.VaultURI)

	// @tombuildsstuff: the API doesn't return this field if it's not configured
	// however https://docs.microsoft.com/en-us/azure/key-vault/general/soft-delete-overview
	// defaults this to 90 days, as such we're going to have to assume that for the moment
	// in lieu of anything being returned
	softDeleteRetentionDays := 90
	if props.SoftDeleteRetentionInDays != nil && *props.SoftDeleteRetentionInDays != 0 {
		softDeleteRetentionDays = int(*props.SoftDeleteRetentionInDays)
	}
	d.Set("soft_delete_retention_days", softDeleteRetentionDays)

	skuName := ""
	if sku := props.Sku; sku != nil {
		// the Azure API is inconsistent here, so rewrite this into the casing we expect
		for _, v := range keyvault.PossibleSkuNameValues() {
			if strings.EqualFold(string(v), string(sku.Name)) {
				skuName = string(v)
			}
		}
	}
	d.Set("sku_name", skuName)

	if err := d.Set("network_acls", flattenKeyVaultNetworkAcls(props.NetworkAcls)); err != nil {
		return fmt.Errorf("setting `network_acls` for KeyVault %q: %+v", *resp.Name, err)
	}

	flattenedPolicies := flattenAccessPolicies(props.AccessPolicies)
	if err := d.Set("access_policy", flattenedPolicies); err != nil {
		return fmt.Errorf("setting `access_policy` for KeyVault %q: %+v", *resp.Name, err)
	}

	contactsResp, err := managementClient.GetCertificateContacts(ctx, *props.VaultURI)
	if err != nil {
		if !utils.ResponseWasForbidden(contactsResp.Response) && !utils.ResponseWasNotFound(contactsResp.Response) {
			return fmt.Errorf("retrieving `contact` for KeyVault: %+v", err)
		}
	}
	if err := d.Set("contact", flattenKeyVaultCertificateContactList(contactsResp)); err != nil {
		return fmt.Errorf("setting `contact` for KeyVault: %+v", err)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceKeyVaultDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).KeyVault.VaultsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VaultID(d.Id())
	if err != nil {
		return err
	}

	locks.ByName(id.Name, keyVaultResourceName)
	defer locks.UnlockByName(id.Name, keyVaultResourceName)

	read, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(read.Response) {
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	if read.Properties == nil {
		return fmt.Errorf("retrieving %q: `properties` was nil", *id)
	}
	if read.Location == nil {
		return fmt.Errorf("retrieving %q: `location` was nil", *id)
	}

	// Check to see if purge protection is enabled or not...
	purgeProtectionEnabled := false
	if ppe := read.Properties.EnablePurgeProtection; ppe != nil {
		purgeProtectionEnabled = *ppe
	}
	softDeleteEnabled := false
	if sde := read.Properties.EnableSoftDelete; sde != nil {
		softDeleteEnabled = *sde
	}

	// ensure we lock on the latest network names, to ensure we handle Azure's networking layer being limited to one change at a time
	virtualNetworkNames := make([]string, 0)
	if props := read.Properties; props != nil {
		if acls := props.NetworkAcls; acls != nil {
			if rules := acls.VirtualNetworkRules; rules != nil {
				for _, v := range *rules {
					if v.ID == nil {
						continue
					}

					subnetId, err := networkParse.SubnetIDInsensitively(*v.ID)
					if err != nil {
						return err
					}

					if !utils.SliceContainsValue(virtualNetworkNames, subnetId.VirtualNetworkName) {
						virtualNetworkNames = append(virtualNetworkNames, subnetId.VirtualNetworkName)
					}
				}
			}
		}
	}

	locks.MultipleByName(&virtualNetworkNames, network.VirtualNetworkResourceName)
	defer locks.UnlockMultipleByName(&virtualNetworkNames, network.VirtualNetworkResourceName)

	resp, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if !response.WasNotFound(resp.Response) {
			return fmt.Errorf("retrieving %s: %+v", *id, err)
		}
	}

	// Purge the soft deleted key vault permanently if the feature flag is enabled
	if meta.(*clients.Client).Features.KeyVault.PurgeSoftDeleteOnDestroy && softDeleteEnabled {
		// KeyVaults with Purge Protection Enabled cannot be deleted unless done by Azure
		if purgeProtectionEnabled {
			deletedInfo, err := getSoftDeletedStateForKeyVault(ctx, client, id.Name, *read.Location)
			if err != nil {
				return fmt.Errorf("retrieving the Deletion Details for %s: %+v", *id, err)
			}

			// in the future it'd be nice to raise a warning, but this is the best we can do for now
			if deletedInfo != nil {
				log.Printf("[DEBUG] The Key Vault %q has Purge Protection Enabled and was deleted on %q. Azure will purge this on %q", id.Name, deletedInfo.deleteDate, deletedInfo.purgeDate)
			} else {
				log.Printf("[DEBUG] The Key Vault %q has Purge Protection Enabled and will be purged automatically by Azure", id.Name)
			}
			return nil
		}

		log.Printf("[DEBUG] KeyVault %q marked for purge - executing purge", id.Name)
		future, err := client.PurgeDeleted(ctx, id.Name, *read.Location)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] Waiting for purge of KeyVault %q..", id.Name)
		err = future.WaitForCompletionRef(ctx, client.Client)
		if err != nil {
			return fmt.Errorf("purging %s: %+v", *id, err)
		}
		log.Printf("[DEBUG] Purged KeyVault %q.", id.Name)
	}

	meta.(*clients.Client).KeyVault.Purge(*id)

	return nil
}

func keyVaultRefreshFunc(vaultUri string) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Checking to see if KeyVault %q is available..", vaultUri)

		client := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		}

		conn, err := client.Get(vaultUri)
		if err != nil {
			log.Printf("[DEBUG] Didn't find KeyVault at %q", vaultUri)
			return nil, "pending", fmt.Errorf("connecting to %q: %s", vaultUri, err)
		}

		defer conn.Body.Close()

		log.Printf("[DEBUG] Found KeyVault at %q", vaultUri)
		return "available", "available", nil
	}
}

func expandKeyVaultNetworkAcls(input []interface{}) (*keyvault.NetworkRuleSet, []string) {
	subnetIds := make([]string, 0)
	if len(input) == 0 {
		return nil, subnetIds
	}

	v := input[0].(map[string]interface{})

	bypass := v["bypass"].(string)
	defaultAction := v["default_action"].(string)

	ipRulesRaw := v["ip_rules"].(*pluginsdk.Set)
	ipRules := make([]keyvault.IPRule, 0)

	for _, v := range ipRulesRaw.List() {
		rule := keyvault.IPRule{
			Value: utils.String(v.(string)),
		}
		ipRules = append(ipRules, rule)
	}

	networkRulesRaw := v["virtual_network_subnet_ids"].(*pluginsdk.Set)
	networkRules := make([]keyvault.VirtualNetworkRule, 0)
	for _, v := range networkRulesRaw.List() {
		rawId := v.(string)
		subnetIds = append(subnetIds, rawId)
		rule := keyvault.VirtualNetworkRule{
			ID: utils.String(rawId),
		}
		networkRules = append(networkRules, rule)
	}

	ruleSet := keyvault.NetworkRuleSet{
		Bypass:              keyvault.NetworkRuleBypassOptions(bypass),
		DefaultAction:       keyvault.NetworkRuleAction(defaultAction),
		IPRules:             &ipRules,
		VirtualNetworkRules: &networkRules,
	}
	return &ruleSet, subnetIds
}

func expandKeyVaultCertificateContactList(input []interface{}) *[]KeyVaultMgmt.Contact {
	results := make([]KeyVaultMgmt.Contact, 0)
	if len(input) == 0 || input[0] == nil {
		return &results
	}

	for _, item := range input {
		v := item.(map[string]interface{})
		results = append(results, KeyVaultMgmt.Contact{
			Name:         utils.String(v["name"].(string)),
			EmailAddress: utils.String(v["email"].(string)),
			Phone:        utils.String(v["phone"].(string)),
		})
	}

	return &results
}

func flattenKeyVaultNetworkAcls(input *keyvault.NetworkRuleSet) []interface{} {
	if input == nil {
		return []interface{}{
			map[string]interface{}{
				"bypass":                     string(keyvault.NetworkRuleBypassOptionsAzureServices),
				"default_action":             string(keyvault.NetworkRuleActionAllow),
				"ip_rules":                   pluginsdk.NewSet(pluginsdk.HashString, []interface{}{}),
				"virtual_network_subnet_ids": pluginsdk.NewSet(pluginsdk.HashString, []interface{}{}),
			},
		}
	}

	output := make(map[string]interface{})

	output["bypass"] = string(input.Bypass)
	output["default_action"] = string(input.DefaultAction)

	ipRules := make([]interface{}, 0)
	if input.IPRules != nil {
		for _, v := range *input.IPRules {
			if v.Value == nil {
				continue
			}

			ipRules = append(ipRules, *v.Value)
		}
	}
	output["ip_rules"] = pluginsdk.NewSet(pluginsdk.HashString, ipRules)

	virtualNetworkRules := make([]interface{}, 0)
	if input.VirtualNetworkRules != nil {
		for _, v := range *input.VirtualNetworkRules {
			if v.ID == nil {
				continue
			}

			id := *v.ID
			subnetId, err := networkParse.SubnetIDInsensitively(*v.ID)
			if err == nil {
				id = subnetId.ID()
			}

			virtualNetworkRules = append(virtualNetworkRules, id)
		}
	}
	output["virtual_network_subnet_ids"] = pluginsdk.NewSet(pluginsdk.HashString, virtualNetworkRules)

	return []interface{}{output}
}

func flattenKeyVaultCertificateContactList(input KeyVaultMgmt.Contacts) []interface{} {
	results := make([]interface{}, 0)
	if input.ContactList == nil {
		return results
	}

	for _, contact := range *input.ContactList {
		emailAddress := ""
		if contact.EmailAddress != nil {
			emailAddress = *contact.EmailAddress
		}

		name := ""
		if contact.Name != nil {
			name = *contact.Name
		}

		phone := ""
		if contact.Phone != nil {
			phone = *contact.Phone
		}

		results = append(results, map[string]interface{}{
			"email": emailAddress,
			"name":  name,
			"phone": phone,
		})
	}

	return results
}

func optedOutOfRecoveringSoftDeletedKeyVaultErrorFmt(name, location string) string {
	return fmt.Sprintf(`
An existing soft-deleted Key Vault exists with the Name %q in the location %q, however
automatically recovering this KeyVault has been disabled via the "features" block.

Terraform can automatically recover the soft-deleted Key Vault when this behaviour is
enabled within the "features" block (located within the "provider" block) - more
information can be found here:

https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs#features

Alternatively you can manually recover this (e.g. using the Azure CLI) and then import
this into Terraform via "terraform import", or pick a different name/location.
`, name, location)
}

type keyVaultDeletionStatus struct {
	deleteDate string
	purgeDate  string
}

func getSoftDeletedStateForKeyVault(ctx context.Context, client *keyvault.VaultsClient, name string, location string) (*keyVaultDeletionStatus, error) {
	softDel, err := client.GetDeleted(ctx, name, location)
	if err != nil {
		return nil, err
	}

	// we found an existing key vault that is not soft deleted
	if softDel.Properties == nil {
		return nil, nil
	}

	// the logic is this way because the GetDeleted call will return an existing key vault
	// that is not soft deleted, but the Deleted Vault properties will be nil
	props := *softDel.Properties

	result := keyVaultDeletionStatus{}
	if props.DeletionDate != nil {
		result.deleteDate = props.DeletionDate.Format(time.RFC3339)
	}
	if props.ScheduledPurgeDate != nil {
		result.purgeDate = props.ScheduledPurgeDate.Format(time.RFC3339)
	}

	if result.deleteDate == "" && result.purgeDate == "" {
		return nil, nil
	}

	return &result, nil
}
