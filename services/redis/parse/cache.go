package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

type CacheId struct {
	SubscriptionId string
	ResourceGroup  string
	RediName       string
}

func NewCacheID(subscriptionId, resourceGroup, rediName string) CacheId {
	return CacheId{
		SubscriptionId: subscriptionId,
		ResourceGroup:  resourceGroup,
		RediName:       rediName,
	}
}

func (id CacheId) String() string {
	segments := []string{
		fmt.Sprintf("Redi Name %q", id.RediName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Cache", segmentsStr)
}

func (id CacheId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Cache/Redis/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.RediName)
}

// CacheID parses a Cache ID into an CacheId struct
func CacheID(input string) (*CacheId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := CacheId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.RediName, err = id.PopSegment("Redis"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
