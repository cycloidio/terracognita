package validate

import (
	"fmt"

	"github.com/hashicorp/terraform-provider-azurerm/services/blueprints/parse"
)

func VersionID(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return
	}

	if _, err := parse.VersionID(v); err != nil {
		errors = append(errors, fmt.Errorf("Can not parse %q as a Blueprint Version id: %v", k, err))
		return
	}

	return warnings, errors
}
