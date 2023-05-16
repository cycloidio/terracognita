package validate

import (
	"fmt"

	"github.com/hashicorp/terraform-provider-azurerm/services/relay/sdk/2017-04-01/hybridconnections"
)

func HybridConnectionID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if _, err := hybridconnections.ParseHybridConnectionID(v); err != nil {
		errors = append(errors, err)
	}

	return
}
