package validate

import (
	"regexp"

	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
)

// RunbookName validates Automation Account Runbook names
func RunbookName() pluginsdk.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile(`^[0-9a-zA-Z][-_0-9a-zA-Z]{0,62}$`),
		`The name can contain only letters, numbers, underscores and dashes. The name must begin with a letter. The name must be less than 64 characters.`,
	)
}
