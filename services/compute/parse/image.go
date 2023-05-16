package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

type ImageId struct {
	SubscriptionId string
	ResourceGroup  string
	Name           string
}

func NewImageID(subscriptionId, resourceGroup, name string) ImageId {
	return ImageId{
		SubscriptionId: subscriptionId,
		ResourceGroup:  resourceGroup,
		Name:           name,
	}
}

func (id ImageId) String() string {
	segments := []string{
		fmt.Sprintf("Name %q", id.Name),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Image", segmentsStr)
}

func (id ImageId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/images/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.Name)
}

// ImageID parses a Image ID into an ImageId struct
func ImageID(input string) (*ImageId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := ImageId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.Name, err = id.PopSegment("images"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
