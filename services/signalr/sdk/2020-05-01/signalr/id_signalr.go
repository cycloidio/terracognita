package signalr

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

var _ resourceids.ResourceId = SignalRId{}

// SignalRId is a struct representing the Resource ID for a Signal R
type SignalRId struct {
	SubscriptionId    string
	ResourceGroupName string
	ResourceName      string
}

// NewSignalRID returns a new SignalRId struct
func NewSignalRID(subscriptionId string, resourceGroupName string, resourceName string) SignalRId {
	return SignalRId{
		SubscriptionId:    subscriptionId,
		ResourceGroupName: resourceGroupName,
		ResourceName:      resourceName,
	}
}

// ParseSignalRID parses 'input' into a SignalRId
func ParseSignalRID(input string) (*SignalRId, error) {
	parser := resourceids.NewParserFromResourceIdType(SignalRId{})
	parsed, err := parser.Parse(input, false)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	var ok bool
	id := SignalRId{}

	if id.SubscriptionId, ok = parsed.Parsed["subscriptionId"]; !ok {
		return nil, fmt.Errorf("the segment 'subscriptionId' was not found in the resource id %q", input)
	}

	if id.ResourceGroupName, ok = parsed.Parsed["resourceGroupName"]; !ok {
		return nil, fmt.Errorf("the segment 'resourceGroupName' was not found in the resource id %q", input)
	}

	if id.ResourceName, ok = parsed.Parsed["resourceName"]; !ok {
		return nil, fmt.Errorf("the segment 'resourceName' was not found in the resource id %q", input)
	}

	return &id, nil
}

// ParseSignalRIDInsensitively parses 'input' case-insensitively into a SignalRId
// note: this method should only be used for API response data and not user input
func ParseSignalRIDInsensitively(input string) (*SignalRId, error) {
	parser := resourceids.NewParserFromResourceIdType(SignalRId{})
	parsed, err := parser.Parse(input, true)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	var ok bool
	id := SignalRId{}

	if id.SubscriptionId, ok = parsed.Parsed["subscriptionId"]; !ok {
		return nil, fmt.Errorf("the segment 'subscriptionId' was not found in the resource id %q", input)
	}

	if id.ResourceGroupName, ok = parsed.Parsed["resourceGroupName"]; !ok {
		return nil, fmt.Errorf("the segment 'resourceGroupName' was not found in the resource id %q", input)
	}

	if id.ResourceName, ok = parsed.Parsed["resourceName"]; !ok {
		return nil, fmt.Errorf("the segment 'resourceName' was not found in the resource id %q", input)
	}

	return &id, nil
}

// ValidateSignalRID checks that 'input' can be parsed as a Signal R ID
func ValidateSignalRID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if _, err := ParseSignalRID(v); err != nil {
		errors = append(errors, err)
	}

	return
}

// ID returns the formatted Signal R ID
func (id SignalRId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.SignalRService/signalR/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroupName, id.ResourceName)
}

// Segments returns a slice of Resource ID Segments which comprise this Signal R ID
func (id SignalRId) Segments() []resourceids.Segment {
	return []resourceids.Segment{
		resourceids.StaticSegment("subscriptions", "subscriptions", "subscriptions"),
		resourceids.SubscriptionIdSegment("subscriptionId", "12345678-1234-9876-4563-123456789012"),
		resourceids.StaticSegment("resourceGroups", "resourceGroups", "resourceGroups"),
		resourceids.ResourceGroupSegment("resourceGroupName", "example-resource-group"),
		resourceids.StaticSegment("providers", "providers", "providers"),
		resourceids.ResourceProviderSegment("microsoftSignalRService", "Microsoft.SignalRService", "Microsoft.SignalRService"),
		resourceids.StaticSegment("signalR", "signalR", "signalR"),
		resourceids.UserSpecifiedSegment("resourceName", "resourceValue"),
	}
}

// String returns a human-readable description of this Signal R ID
func (id SignalRId) String() string {
	components := []string{
		fmt.Sprintf("Subscription: %q", id.SubscriptionId),
		fmt.Sprintf("Resource Group Name: %q", id.ResourceGroupName),
		fmt.Sprintf("Resource Name: %q", id.ResourceName),
	}
	return fmt.Sprintf("Signal R (%s)", strings.Join(components, "\n"))
}
