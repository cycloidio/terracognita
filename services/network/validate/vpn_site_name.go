package validate

import (
	"regexp"

	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
)

func VpnSiteName() func(i interface{}, k string) (warnings []string, errors []error) {
	return validation.StringMatch(regexp.MustCompile(`^[^'<>%&:?/+]+$`), "The value must not contain characters from '<>%&:?/+.")
}
