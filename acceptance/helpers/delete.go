package helpers

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/types"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
)

// DeleteResourceFunc returns a TestCheckFunc which deletes the resource within Azure
// this is only used within the Internal
func DeleteResourceFunc(client *clients.Client, testResource types.TestResourceVerifyingRemoved, resourceName string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		ctx := client.StopContext

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("%q was not found in the state", resourceName)
		}

		result, err := testResource.Destroy(ctx, client, rs.Primary)
		if err != nil {
			return fmt.Errorf("running destroy func for %q: %+v", resourceName, err)
		}
		if result == nil {
			return fmt.Errorf("received nil for destroy result for %q", resourceName)
		}

		if !*result {
			return fmt.Errorf("deleting %q but no error", resourceName)
		}

		return nil
	}
}
