package compute

import "github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"

// nolint: deadcode unused
func adminPasswordDiffSuppressFunc(_, old, new string, _ *pluginsdk.ResourceData) bool {
	// this is not the greatest hack in the world, this is just a tribute.
	if old == "ignored-as-imported" || new == "ignored-as-imported" {
		return true
	}

	return false
}
