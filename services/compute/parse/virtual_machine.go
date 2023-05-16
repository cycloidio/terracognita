package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

type VirtualMachineId struct {
	SubscriptionId string
	ResourceGroup  string
	Name           string
}

func NewVirtualMachineID(subscriptionId, resourceGroup, name string) VirtualMachineId {
	return VirtualMachineId{
		SubscriptionId: subscriptionId,
		ResourceGroup:  resourceGroup,
		Name:           name,
	}
}

func (id VirtualMachineId) String() string {
	segments := []string{
		fmt.Sprintf("Name %q", id.Name),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Virtual Machine", segmentsStr)
}

func (id VirtualMachineId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.Name)
}

// VirtualMachineID parses a VirtualMachine ID into an VirtualMachineId struct
func VirtualMachineID(input string) (*VirtualMachineId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := VirtualMachineId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.Name, err = id.PopSegment("virtualMachines"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
