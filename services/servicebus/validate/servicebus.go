package validate

import (
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
)

func ServiceBusMaxSizeInMegabytes() pluginsdk.SchemaValidateFunc {
	return validation.IntInSlice([]int{
		1024,
		2048,
		3072,
		4096,
		5120,
		10240,
		20480,
		40960,
		81920,
	})
}

func ServiceBusMaxMessageSizeInKilobytes() pluginsdk.SchemaValidateFunc {
	return validation.IntBetween(1024, 102400)
}
