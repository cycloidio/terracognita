package validate

import (
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
)

func ResourceAssignmentId() pluginsdk.SchemaValidateFunc {
	return validation.All(
		validation.None(
			map[string]func(interface{}, string) ([]string, []error){
				"Management Group ID": commonids.ValidateManagementGroupID,
				"Resource Group ID":   commonids.ValidateResourceGroupID,
				"Subscription ID":     commonids.ValidateSubscriptionID,
			},
		),
		azure.ValidateResourceID,
	)
}
