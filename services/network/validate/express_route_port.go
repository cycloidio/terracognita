package validate

import (
	"regexp"

	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
)

func ExpressRoutePortName(i interface{}, k string) (warnings []string, errors []error) {
	return validation.StringMatch(regexp.MustCompile(`^[^\W_]([\w-.]{0,78}[\w])?$`), "")(i, k)
}
