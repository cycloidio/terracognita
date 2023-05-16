package validate

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
)

func IntegrationAccountPartnerBusinessIdentityValue() pluginsdk.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected %q to be a string", k))
			return
		}

		if len(v) > 128 {
			errors = append(errors, fmt.Errorf("length should be equal to or less than %d, got %q", 128, v))
			return
		}

		if !regexp.MustCompile(`^[A-Za-z0-9-() .]+$`).MatchString(v) {
			errors = append(errors, fmt.Errorf("%q contains only letters, numbers, dots, parentheses and hyphens", k))
			return
		}

		return
	}
}
