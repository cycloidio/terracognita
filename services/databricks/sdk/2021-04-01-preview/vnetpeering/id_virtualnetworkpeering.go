package vnetpeering

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

var _ resourceids.ResourceId = VirtualNetworkPeeringId{}

// VirtualNetworkPeeringId is a struct representing the Resource ID for a Virtual Network Peering
type VirtualNetworkPeeringId struct {
	SubscriptionId    string
	ResourceGroupName string
	WorkspaceName     string
	PeeringName       string
}

// NewVirtualNetworkPeeringID returns a new VirtualNetworkPeeringId struct
func NewVirtualNetworkPeeringID(subscriptionId string, resourceGroupName string, workspaceName string, peeringName string) VirtualNetworkPeeringId {
	return VirtualNetworkPeeringId{
		SubscriptionId:    subscriptionId,
		ResourceGroupName: resourceGroupName,
		WorkspaceName:     workspaceName,
		PeeringName:       peeringName,
	}
}

// ParseVirtualNetworkPeeringID parses 'input' into a VirtualNetworkPeeringId
func ParseVirtualNetworkPeeringID(input string) (*VirtualNetworkPeeringId, error) {
	parser := resourceids.NewParserFromResourceIdType(VirtualNetworkPeeringId{})
	parsed, err := parser.Parse(input, false)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	var ok bool
	id := VirtualNetworkPeeringId{}

	if id.SubscriptionId, ok = parsed.Parsed["subscriptionId"]; !ok {
		return nil, fmt.Errorf("the segment 'subscriptionId' was not found in the resource id %q", input)
	}

	if id.ResourceGroupName, ok = parsed.Parsed["resourceGroupName"]; !ok {
		return nil, fmt.Errorf("the segment 'resourceGroupName' was not found in the resource id %q", input)
	}

	if id.WorkspaceName, ok = parsed.Parsed["workspaceName"]; !ok {
		return nil, fmt.Errorf("the segment 'workspaceName' was not found in the resource id %q", input)
	}

	if id.PeeringName, ok = parsed.Parsed["peeringName"]; !ok {
		return nil, fmt.Errorf("the segment 'peeringName' was not found in the resource id %q", input)
	}

	return &id, nil
}

// ParseVirtualNetworkPeeringIDInsensitively parses 'input' case-insensitively into a VirtualNetworkPeeringId
// note: this method should only be used for API response data and not user input
func ParseVirtualNetworkPeeringIDInsensitively(input string) (*VirtualNetworkPeeringId, error) {
	parser := resourceids.NewParserFromResourceIdType(VirtualNetworkPeeringId{})
	parsed, err := parser.Parse(input, true)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	var ok bool
	id := VirtualNetworkPeeringId{}

	if id.SubscriptionId, ok = parsed.Parsed["subscriptionId"]; !ok {
		return nil, fmt.Errorf("the segment 'subscriptionId' was not found in the resource id %q", input)
	}

	if id.ResourceGroupName, ok = parsed.Parsed["resourceGroupName"]; !ok {
		return nil, fmt.Errorf("the segment 'resourceGroupName' was not found in the resource id %q", input)
	}

	if id.WorkspaceName, ok = parsed.Parsed["workspaceName"]; !ok {
		return nil, fmt.Errorf("the segment 'workspaceName' was not found in the resource id %q", input)
	}

	if id.PeeringName, ok = parsed.Parsed["peeringName"]; !ok {
		return nil, fmt.Errorf("the segment 'peeringName' was not found in the resource id %q", input)
	}

	return &id, nil
}

// ValidateVirtualNetworkPeeringID checks that 'input' can be parsed as a Virtual Network Peering ID
func ValidateVirtualNetworkPeeringID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if _, err := ParseVirtualNetworkPeeringID(v); err != nil {
		errors = append(errors, err)
	}

	return
}

// ID returns the formatted Virtual Network Peering ID
func (id VirtualNetworkPeeringId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Databricks/workspaces/%s/virtualNetworkPeerings/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroupName, id.WorkspaceName, id.PeeringName)
}

// Segments returns a slice of Resource ID Segments which comprise this Virtual Network Peering ID
func (id VirtualNetworkPeeringId) Segments() []resourceids.Segment {
	return []resourceids.Segment{
		resourceids.StaticSegment("staticSubscriptions", "subscriptions", "subscriptions"),
		resourceids.SubscriptionIdSegment("subscriptionId", "12345678-1234-9876-4563-123456789012"),
		resourceids.StaticSegment("staticResourceGroups", "resourceGroups", "resourceGroups"),
		resourceids.ResourceGroupSegment("resourceGroupName", "example-resource-group"),
		resourceids.StaticSegment("staticProviders", "providers", "providers"),
		resourceids.ResourceProviderSegment("staticMicrosoftDatabricks", "Microsoft.Databricks", "Microsoft.Databricks"),
		resourceids.StaticSegment("staticWorkspaces", "workspaces", "workspaces"),
		resourceids.UserSpecifiedSegment("workspaceName", "workspaceValue"),
		resourceids.StaticSegment("staticVirtualNetworkPeerings", "virtualNetworkPeerings", "virtualNetworkPeerings"),
		resourceids.UserSpecifiedSegment("peeringName", "peeringValue"),
	}
}

// String returns a human-readable description of this Virtual Network Peering ID
func (id VirtualNetworkPeeringId) String() string {
	components := []string{
		fmt.Sprintf("Subscription: %q", id.SubscriptionId),
		fmt.Sprintf("Resource Group Name: %q", id.ResourceGroupName),
		fmt.Sprintf("Workspace Name: %q", id.WorkspaceName),
		fmt.Sprintf("Peering Name: %q", id.PeeringName),
	}
	return fmt.Sprintf("Virtual Network Peering (%s)", strings.Join(components, "\n"))
}
