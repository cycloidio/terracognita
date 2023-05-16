package validate

import (
	"fmt"

	"github.com/hashicorp/terraform-provider-azurerm/services/portal/parse"
)

func PortalTenantConfigurationID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if _, err := parse.PortalTenantConfigurationID(v); err != nil {
		errors = append(errors, err)
	}

	return
}
