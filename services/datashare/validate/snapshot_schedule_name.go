package validate

import (
	"regexp"

	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
)

func SnapshotScheduleName() pluginsdk.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile(`^[^&%#/]{1,90}$`), `Data share snapshot schedule name should have length of 1 - 90, and cannot contain &%#/`,
	)
}
