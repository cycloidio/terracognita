package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

type MediaServiceId struct {
	SubscriptionId string
	ResourceGroup  string
	Name           string
}

func NewMediaServiceID(subscriptionId, resourceGroup, name string) MediaServiceId {
	return MediaServiceId{
		SubscriptionId: subscriptionId,
		ResourceGroup:  resourceGroup,
		Name:           name,
	}
}

func (id MediaServiceId) String() string {
	segments := []string{
		fmt.Sprintf("Name %q", id.Name),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Media Service", segmentsStr)
}

func (id MediaServiceId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Media/mediaservices/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.Name)
}

// MediaServiceID parses a MediaService ID into an MediaServiceId struct
func MediaServiceID(input string) (*MediaServiceId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := MediaServiceId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.Name, err = id.PopSegment("mediaservices"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
