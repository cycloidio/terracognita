package validate

import (
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
)

func DiskPoolSku() pluginsdk.SchemaValidateFunc {
	return validation.StringInSlice(
		[]string{
			"Basic_B1",
			"Standard_S1",
			"Premium_P1",
		}, false,
	)
}
