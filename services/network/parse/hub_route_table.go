package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

type HubRouteTableId struct {
	SubscriptionId string
	ResourceGroup  string
	VirtualHubName string
	Name           string
}

func NewHubRouteTableID(subscriptionId, resourceGroup, virtualHubName, name string) HubRouteTableId {
	return HubRouteTableId{
		SubscriptionId: subscriptionId,
		ResourceGroup:  resourceGroup,
		VirtualHubName: virtualHubName,
		Name:           name,
	}
}

func (id HubRouteTableId) String() string {
	segments := []string{
		fmt.Sprintf("Name %q", id.Name),
		fmt.Sprintf("Virtual Hub Name %q", id.VirtualHubName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Hub Route Table", segmentsStr)
}

func (id HubRouteTableId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualHubs/%s/hubRouteTables/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.VirtualHubName, id.Name)
}

// HubRouteTableID parses a HubRouteTable ID into an HubRouteTableId struct
func HubRouteTableID(input string) (*HubRouteTableId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := HubRouteTableId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.VirtualHubName, err = id.PopSegment("virtualHubs"); err != nil {
		return nil, err
	}
	if resourceId.Name, err = id.PopSegment("hubRouteTables"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
