package databoxedge

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-provider-azurerm/services/databoxedge/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
)

func databoxEdgeCustomizeDiff(ctx context.Context, d *pluginsdk.ResourceDiff, v interface{}) error {
	if value, ok := d.GetOk("shipment_address"); ok {
		shippingInfo := (value.([]interface{}))[0].(map[string]interface{})

		_, err := validate.DataboxEdgeStreetAddress(shippingInfo["address"].([]interface{}), "address")
		if err != nil {
			return fmt.Errorf("%s", err)
		}
	}

	return nil
}
