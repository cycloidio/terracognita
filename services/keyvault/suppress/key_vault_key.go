package suppress

import (
	keyVaultParse "github.com/hashicorp/terraform-provider-azurerm/services/keyvault/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
)

func DiffSuppressIgnoreKeyVaultKeyVersion(k, old, new string, d *pluginsdk.ResourceData) bool {
	oldKey, err := keyVaultParse.ParseOptionallyVersionedNestedItemID(old)
	if err != nil {
		return false
	}
	newKey, err := keyVaultParse.ParseOptionallyVersionedNestedItemID(new)
	if err != nil {
		return false
	}

	return (oldKey.KeyVaultBaseUrl == newKey.KeyVaultBaseUrl) && (oldKey.Name == newKey.Name)
}
