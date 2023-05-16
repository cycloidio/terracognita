package parse

import (
	"github.com/hashicorp/terraform-provider-azurerm/services/relay/sdk/2017-04-01/namespaces"
)

func NamespaceID(input string) (*namespaces.NamespaceId, error) {
	return namespaces.ParseNamespaceID(input)
}
