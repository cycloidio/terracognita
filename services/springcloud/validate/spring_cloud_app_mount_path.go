package validate

import (
	"fmt"

	"github.com/hashicorp/terraform-provider-azurerm/helpers/validate"
)

func MountPath(i interface{}, k string) (_ []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		return nil, append(errors, fmt.Errorf("expected type of %s to be string", k))
	}
	if len(v) < 2 || len(v) > 255 {
		errors = append(errors, fmt.Errorf("%s should equal or great than 2 and euqal or less than 255", k))
	} else if m, _ := validate.RegExHelper(i, k, `^(?:\/(?:[a-zA-Z][a-zA-Z0-9]*))+$`); !m {
		errors = append(errors, fmt.Errorf("%s is not valid, must match the regular expression ^(?:\\/(?:[a-zA-Z][a-zA-Z0-9]*))+$", k))
	}
	return nil, errors
}
