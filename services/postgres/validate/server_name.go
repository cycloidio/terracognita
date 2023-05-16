package validate

import (
	"fmt"

	"github.com/hashicorp/terraform-provider-azurerm/helpers/validate"
)

func ServerName(i interface{}, k string) (_ []string, errors []error) {
	if m, regexErrs := validate.RegExHelper(i, k, `^[0-9a-z][-0-9a-z]{1,61}[0-9a-z]$`); !m {
		return nil, append(regexErrs, fmt.Errorf("%q can contain only lowercase letters, numbers, and '-', but can't start or end with '-'. And must be at least 3 characters and at most 63 characters", k))
	}

	return nil, nil
}
