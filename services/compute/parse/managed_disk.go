package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

type ManagedDiskId struct {
	SubscriptionId string
	ResourceGroup  string
	DiskName       string
}

func NewManagedDiskID(subscriptionId, resourceGroup, diskName string) ManagedDiskId {
	return ManagedDiskId{
		SubscriptionId: subscriptionId,
		ResourceGroup:  resourceGroup,
		DiskName:       diskName,
	}
}

func (id ManagedDiskId) String() string {
	segments := []string{
		fmt.Sprintf("Disk Name %q", id.DiskName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Managed Disk", segmentsStr)
}

func (id ManagedDiskId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/disks/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.DiskName)
}

// ManagedDiskID parses a ManagedDisk ID into an ManagedDiskId struct
func ManagedDiskID(input string) (*ManagedDiskId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := ManagedDiskId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.DiskName, err = id.PopSegment("disks"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
