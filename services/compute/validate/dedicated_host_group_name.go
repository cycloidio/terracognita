package validate

import (
	"regexp"

	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
)

func DedicatedHostGroupName() func(i interface{}, k string) (warnings []string, errors []error) {
	return validation.StringMatch(regexp.MustCompile(`^[^_\W][\w-.]{0,78}[\w]$`), "")
}
