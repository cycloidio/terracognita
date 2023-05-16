package automation

import (
	"time"

	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
)

func dataSourceAutomationVariableBool() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceAutomationVariableBoolRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: datasourceAutomationVariableCommonSchema(pluginsdk.TypeBool),
	}
}

func dataSourceAutomationVariableBoolRead(d *pluginsdk.ResourceData, meta interface{}) error {
	return dataSourceAutomationVariableRead(d, meta, "Bool")
}
