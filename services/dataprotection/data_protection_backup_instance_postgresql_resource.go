package dataprotection

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/dataprotection/legacysdk/dataprotection"
	"github.com/hashicorp/terraform-provider-azurerm/services/dataprotection/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/dataprotection/validate"
	keyVaultValidate "github.com/hashicorp/terraform-provider-azurerm/services/keyvault/validate"
	postgresParse "github.com/hashicorp/terraform-provider-azurerm/services/postgres/parse"
	postgresValidate "github.com/hashicorp/terraform-provider-azurerm/services/postgres/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	azSchema "github.com/hashicorp/terraform-provider-azurerm/tf/schema"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceDataProtectionBackupInstancePostgreSQL() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceDataProtectionBackupInstancePostgreSQLCreateUpdate,
		Read:   resourceDataProtectionBackupInstancePostgreSQLRead,
		Update: resourceDataProtectionBackupInstancePostgreSQLCreateUpdate,
		Delete: resourceDataProtectionBackupInstancePostgreSQLDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: azSchema.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := parse.BackupInstanceID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"location": commonschema.Location(),

			"vault_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.BackupVaultID,
			},

			"database_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: postgresValidate.DatabaseID,
			},

			"backup_policy_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.BackupPolicyID,
			},

			"database_credential_key_vault_secret_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: keyVaultValidate.NestedItemIdWithOptionalVersion,
			},
		},
	}
}

func resourceDataProtectionBackupInstancePostgreSQLCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).DataProtection.BackupInstanceClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	vaultId, _ := parse.BackupVaultID(d.Get("vault_id").(string))

	id := parse.NewBackupInstanceID(subscriptionId, vaultId.ResourceGroup, vaultId.Name, name)

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.BackupVaultName, id.ResourceGroup, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for existing DataProtection BackupInstance (%q): %+v", id, err)
			}
		}
		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_data_protection_backup_instance_postgresql", id.ID())
		}
	}

	databaseId, _ := postgresParse.DatabaseID(d.Get("database_id").(string))
	location := location.Normalize(d.Get("location").(string))
	serverId := postgresParse.NewServerID(databaseId.SubscriptionId, databaseId.ResourceGroup, databaseId.ServerName)
	policyId, _ := parse.BackupPolicyID(d.Get("backup_policy_id").(string))

	parameters := dataprotection.BackupInstanceResource{
		Properties: &dataprotection.BackupInstance{
			DataSourceInfo: &dataprotection.Datasource{
				DatasourceType:   utils.String("Microsoft.DBforPostgreSQL/servers/databases"),
				ObjectType:       utils.String("Datasource"),
				ResourceID:       utils.String(databaseId.ID()),
				ResourceLocation: utils.String(location),
				ResourceName:     utils.String(databaseId.Name),
				ResourceType:     utils.String("Microsoft.DBforPostgreSQL/servers/databases"),
				ResourceURI:      utils.String(""),
			},
			DataSourceSetInfo: &dataprotection.DatasourceSet{
				DatasourceType:   utils.String("Microsoft.DBForPostgreSQL/servers"),
				ObjectType:       utils.String("DatasourceSet"),
				ResourceID:       utils.String(serverId.ID()),
				ResourceLocation: utils.String(location),
				ResourceName:     utils.String(serverId.Name),
				ResourceType:     utils.String("Microsoft.DBForPostgreSQL/servers"),
				ResourceURI:      utils.String(""),
			},
			FriendlyName: utils.String(id.Name),
			PolicyInfo: &dataprotection.PolicyInfo{
				PolicyID: utils.String(policyId.ID()),
			},
		},
	}

	if v, ok := d.GetOk("database_credential_key_vault_secret_id"); ok {
		parameters.Properties.DatasourceAuthCredentials = &dataprotection.SecretStoreBasedAuthCredentials{
			SecretStoreResource: &dataprotection.SecretStoreResource{
				URI:             utils.String(v.(string)),
				SecretStoreType: dataprotection.SecretStoreTypeAzureKeyVault,
			},
			ObjectType: dataprotection.ObjectTypeSecretStoreBasedAuthCredentials,
		}
	}

	future, err := client.CreateOrUpdate(ctx, id.BackupVaultName, id.ResourceGroup, id.Name, parameters)
	if err != nil {
		return fmt.Errorf("creating/updating DataProtection BackupInstance (%q): %+v", id, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation/update of the DataProtection BackupInstance (%q): %+v", id, err)
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		return fmt.Errorf("context had no deadline")
	}
	stateConf := &pluginsdk.StateChangeConf{
		Pending:    []string{string(dataprotection.StatusConfiguringProtection), "UpdatingProtection"},
		Target:     []string{string(dataprotection.StatusProtectionConfigured)},
		Refresh:    policyProtectionStateRefreshFunc(ctx, client, id),
		MinTimeout: 1 * time.Minute,
		Timeout:    time.Until(deadline),
	}

	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for BackupInstance(%q) policy protection to be completed: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceDataProtectionBackupInstancePostgreSQLRead(d, meta)
}

func resourceDataProtectionBackupInstancePostgreSQLRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataProtection.BackupInstanceClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.BackupInstanceID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.BackupVaultName, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] dataprotection %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving DataProtection BackupInstance (%q): %+v", id, err)
	}
	vaultId := parse.NewBackupVaultID(id.SubscriptionId, id.ResourceGroup, id.BackupVaultName)
	d.Set("name", id.Name)
	d.Set("vault_id", vaultId.ID())
	if props := resp.Properties; props != nil {
		if props.DataSourceInfo != nil {
			d.Set("database_id", props.DataSourceInfo.ResourceID)
			d.Set("location", props.DataSourceInfo.ResourceLocation)
		}
		if props.PolicyInfo != nil {
			d.Set("backup_policy_id", props.PolicyInfo.PolicyID)
		}
		if props.DatasourceAuthCredentials != nil {
			if credential, ok := props.DatasourceAuthCredentials.AsSecretStoreBasedAuthCredentials(); ok {
				if credential.SecretStoreResource != nil {
					d.Set("database_credential_key_vault_secret_id", credential.SecretStoreResource.URI)
				}
			} else {
				log.Printf("[DEBUG] Skipping setting database_credential_key_vault_secret_id since this DatasourceAuthCredentials is not supported")
			}
		}
	}
	return nil
}

func resourceDataProtectionBackupInstancePostgreSQLDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataProtection.BackupInstanceClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.BackupInstanceID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.BackupVaultName, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting DataProtection BackupInstance (%q): %+v", id, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of the DataProtection BackupInstance (%q): %+v", id.Name, err)
	}
	return nil
}

func policyProtectionStateRefreshFunc(ctx context.Context, client *dataprotection.BackupInstancesClient, id parse.BackupInstanceId) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		res, err := client.Get(ctx, id.BackupVaultName, id.ResourceGroup, id.Name)
		if err != nil {
			return nil, "", fmt.Errorf("retrieving DataProtection BackupInstance (%q): %+v", id, err)
		}
		if res.Properties == nil || res.Properties.ProtectionStatus == nil {
			return nil, "", fmt.Errorf("reading DataProtection BackupInstance (%q) protection status: %+v", id, err)
		}

		return res, string(res.Properties.ProtectionStatus.Status), nil
	}
}
