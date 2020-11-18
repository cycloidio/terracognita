package azurerm

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
)

// ListVirtualMachineExtensions returns a list of VirtualMachineExtensions within a subscription and a resource group
// it needs to be manually written since it does not follow the same pattern as the other resources. This "List" method
// does not return a "...ListResultPage", but directly a slice holding the values
func (ar *AzureReader) ListVirtualMachineExtensions(ctx context.Context, VMName string, expand string) ([]compute.VirtualMachineExtension, error) {
	client := compute.NewVirtualMachineExtensionsClient(ar.config.SubscriptionID)
	client.Authorizer = ar.authorizer

	output, err := client.List(ctx, ar.GetResourceGroupName(), VMName, expand)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list compute.VirtualMachineExtension from Azure APIs")
	}

	return *output.Value, nil
}
