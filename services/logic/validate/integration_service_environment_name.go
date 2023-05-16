package validate

import (
	"regexp"

	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
)

// https://docs.microsoft.com/en-us/azure/azure-resource-manager/management/resource-name-rules
// 1-80	Alphanumerics, hyphens, periods, and underscores.
func IntegrationServiceEnvironmentName() pluginsdk.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile(`^[\w-.]{1,80}$`), `Integration Service Environment names must be between 1 and 80 characters in length, contain only letters, numbers, underscores, hyphens and periods.`,
	)
}
