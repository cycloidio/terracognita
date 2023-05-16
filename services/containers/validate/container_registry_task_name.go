package validate

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
)

func ContainerRegistryTaskName(v interface{}, k string) (warnings []string, errors []error) {
	return validation.StringMatch(regexp.MustCompile(`^[\w-]*$`), fmt.Sprintf("only alpha numeric characters (optionally separated by dash) are allowed in %q", k))(v, k)
}
