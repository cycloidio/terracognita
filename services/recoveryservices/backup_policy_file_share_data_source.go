package recoveryservices

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/recoveryservices/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func dataSourceBackupPolicyFileShare() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceBackupPolicyFileShareRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: dataSourceBackupPolicyFileShareSchema(),
	}
}

func dataSourceBackupPolicyFileShareRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).RecoveryServices.ProtectionPoliciesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	vaultName := d.Get("recovery_vault_name").(string)

	log.Printf("[DEBUG] Reading Recovery Service Policy %q (resource group %q)", name, resourceGroup)

	protectionPolicy, err := client.Get(ctx, vaultName, resourceGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(protectionPolicy.Response) {
			return fmt.Errorf("Error: Backup Policy %q (Resource Group %q) was not found", name, resourceGroup)
		}

		return fmt.Errorf("making Read request on Backup Policy %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	id := strings.Replace(*protectionPolicy.ID, "Subscriptions", "subscriptions", 1)
	d.SetId(id)

	return nil
}

func dataSourceBackupPolicyFileShareSchema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"recovery_vault_name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ValidateFunc: validate.RecoveryServicesVaultName,
		},

		"resource_group_name": commonschema.ResourceGroupNameForDataSource(),
	}
}
