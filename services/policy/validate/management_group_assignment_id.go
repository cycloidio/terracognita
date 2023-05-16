package validate

import (
	"fmt"

	"github.com/hashicorp/terraform-provider-azurerm/services/policy/parse"
)

func ManagementGroupAssignmentID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if _, err := parse.ManagementGroupAssignmentID(v); err != nil {
		errors = append(errors, err)
	}

	return
}
