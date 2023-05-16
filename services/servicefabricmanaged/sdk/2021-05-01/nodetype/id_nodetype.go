package nodetype

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

var _ resourceids.ResourceId = NodeTypeId{}

// NodeTypeId is a struct representing the Resource ID for a Node Type
type NodeTypeId struct {
	SubscriptionId    string
	ResourceGroupName string
	ClusterName       string
	NodeTypeName      string
}

// NewNodeTypeID returns a new NodeTypeId struct
func NewNodeTypeID(subscriptionId string, resourceGroupName string, clusterName string, nodeTypeName string) NodeTypeId {
	return NodeTypeId{
		SubscriptionId:    subscriptionId,
		ResourceGroupName: resourceGroupName,
		ClusterName:       clusterName,
		NodeTypeName:      nodeTypeName,
	}
}

// ParseNodeTypeID parses 'input' into a NodeTypeId
func ParseNodeTypeID(input string) (*NodeTypeId, error) {
	parser := resourceids.NewParserFromResourceIdType(NodeTypeId{})
	parsed, err := parser.Parse(input, false)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	var ok bool
	id := NodeTypeId{}

	if id.SubscriptionId, ok = parsed.Parsed["subscriptionId"]; !ok {
		return nil, fmt.Errorf("the segment 'subscriptionId' was not found in the resource id %q", input)
	}

	if id.ResourceGroupName, ok = parsed.Parsed["resourceGroupName"]; !ok {
		return nil, fmt.Errorf("the segment 'resourceGroupName' was not found in the resource id %q", input)
	}

	if id.ClusterName, ok = parsed.Parsed["clusterName"]; !ok {
		return nil, fmt.Errorf("the segment 'clusterName' was not found in the resource id %q", input)
	}

	if id.NodeTypeName, ok = parsed.Parsed["nodeTypeName"]; !ok {
		return nil, fmt.Errorf("the segment 'nodeTypeName' was not found in the resource id %q", input)
	}

	return &id, nil
}

// ParseNodeTypeIDInsensitively parses 'input' case-insensitively into a NodeTypeId
// note: this method should only be used for API response data and not user input
func ParseNodeTypeIDInsensitively(input string) (*NodeTypeId, error) {
	parser := resourceids.NewParserFromResourceIdType(NodeTypeId{})
	parsed, err := parser.Parse(input, true)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	var ok bool
	id := NodeTypeId{}

	if id.SubscriptionId, ok = parsed.Parsed["subscriptionId"]; !ok {
		return nil, fmt.Errorf("the segment 'subscriptionId' was not found in the resource id %q", input)
	}

	if id.ResourceGroupName, ok = parsed.Parsed["resourceGroupName"]; !ok {
		return nil, fmt.Errorf("the segment 'resourceGroupName' was not found in the resource id %q", input)
	}

	if id.ClusterName, ok = parsed.Parsed["clusterName"]; !ok {
		return nil, fmt.Errorf("the segment 'clusterName' was not found in the resource id %q", input)
	}

	if id.NodeTypeName, ok = parsed.Parsed["nodeTypeName"]; !ok {
		return nil, fmt.Errorf("the segment 'nodeTypeName' was not found in the resource id %q", input)
	}

	return &id, nil
}

// ValidateNodeTypeID checks that 'input' can be parsed as a Node Type ID
func ValidateNodeTypeID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if _, err := ParseNodeTypeID(v); err != nil {
		errors = append(errors, err)
	}

	return
}

// ID returns the formatted Node Type ID
func (id NodeTypeId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ServiceFabric/managedClusters/%s/nodeTypes/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroupName, id.ClusterName, id.NodeTypeName)
}

// Segments returns a slice of Resource ID Segments which comprise this Node Type ID
func (id NodeTypeId) Segments() []resourceids.Segment {
	return []resourceids.Segment{
		resourceids.StaticSegment("subscriptions", "subscriptions", "subscriptions"),
		resourceids.SubscriptionIdSegment("subscriptionId", "12345678-1234-9876-4563-123456789012"),
		resourceids.StaticSegment("resourceGroups", "resourceGroups", "resourceGroups"),
		resourceids.ResourceGroupSegment("resourceGroupName", "example-resource-group"),
		resourceids.StaticSegment("providers", "providers", "providers"),
		resourceids.ResourceProviderSegment("microsoftServiceFabric", "Microsoft.ServiceFabric", "Microsoft.ServiceFabric"),
		resourceids.StaticSegment("managedClusters", "managedClusters", "managedClusters"),
		resourceids.UserSpecifiedSegment("clusterName", "clusterValue"),
		resourceids.StaticSegment("nodeTypes", "nodeTypes", "nodeTypes"),
		resourceids.UserSpecifiedSegment("nodeTypeName", "nodeTypeValue"),
	}
}

// String returns a human-readable description of this Node Type ID
func (id NodeTypeId) String() string {
	components := []string{
		fmt.Sprintf("Subscription: %q", id.SubscriptionId),
		fmt.Sprintf("Resource Group Name: %q", id.ResourceGroupName),
		fmt.Sprintf("Cluster Name: %q", id.ClusterName),
		fmt.Sprintf("Node Type Name: %q", id.NodeTypeName),
	}
	return fmt.Sprintf("Node Type (%s)", strings.Join(components, "\n"))
}
