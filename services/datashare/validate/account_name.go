package validate

import (
	"regexp"

	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
)

func AccountName() pluginsdk.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile(`^[^<>%&:\\?/#*$^();,.\|+={}\[\]!~@]{3,90}$`), `Data share account name should have length of 3 - 90, and cannot contain <>%&:\?/#*$^();,.|+={}[]!~@.`,
	)
}
