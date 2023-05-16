package privateendpointconnections

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

var _ resourceids.ResourceId = PrivateEndpointConnectionId{}

// PrivateEndpointConnectionId is a struct representing the Resource ID for a Private Endpoint Connection
type PrivateEndpointConnectionId struct {
	SubscriptionId                string
	ResourceGroupName             string
	ConfigStoreName               string
	PrivateEndpointConnectionName string
}

// NewPrivateEndpointConnectionID returns a new PrivateEndpointConnectionId struct
func NewPrivateEndpointConnectionID(subscriptionId string, resourceGroupName string, configStoreName string, privateEndpointConnectionName string) PrivateEndpointConnectionId {
	return PrivateEndpointConnectionId{
		SubscriptionId:                subscriptionId,
		ResourceGroupName:             resourceGroupName,
		ConfigStoreName:               configStoreName,
		PrivateEndpointConnectionName: privateEndpointConnectionName,
	}
}

// ParsePrivateEndpointConnectionID parses 'input' into a PrivateEndpointConnectionId
func ParsePrivateEndpointConnectionID(input string) (*PrivateEndpointConnectionId, error) {
	parser := resourceids.NewParserFromResourceIdType(PrivateEndpointConnectionId{})
	parsed, err := parser.Parse(input, false)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	var ok bool
	id := PrivateEndpointConnectionId{}

	if id.SubscriptionId, ok = parsed.Parsed["subscriptionId"]; !ok {
		return nil, fmt.Errorf("the segment 'subscriptionId' was not found in the resource id %q", input)
	}

	if id.ResourceGroupName, ok = parsed.Parsed["resourceGroupName"]; !ok {
		return nil, fmt.Errorf("the segment 'resourceGroupName' was not found in the resource id %q", input)
	}

	if id.ConfigStoreName, ok = parsed.Parsed["configStoreName"]; !ok {
		return nil, fmt.Errorf("the segment 'configStoreName' was not found in the resource id %q", input)
	}

	if id.PrivateEndpointConnectionName, ok = parsed.Parsed["privateEndpointConnectionName"]; !ok {
		return nil, fmt.Errorf("the segment 'privateEndpointConnectionName' was not found in the resource id %q", input)
	}

	return &id, nil
}

// ParsePrivateEndpointConnectionIDInsensitively parses 'input' case-insensitively into a PrivateEndpointConnectionId
// note: this method should only be used for API response data and not user input
func ParsePrivateEndpointConnectionIDInsensitively(input string) (*PrivateEndpointConnectionId, error) {
	parser := resourceids.NewParserFromResourceIdType(PrivateEndpointConnectionId{})
	parsed, err := parser.Parse(input, true)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	var ok bool
	id := PrivateEndpointConnectionId{}

	if id.SubscriptionId, ok = parsed.Parsed["subscriptionId"]; !ok {
		return nil, fmt.Errorf("the segment 'subscriptionId' was not found in the resource id %q", input)
	}

	if id.ResourceGroupName, ok = parsed.Parsed["resourceGroupName"]; !ok {
		return nil, fmt.Errorf("the segment 'resourceGroupName' was not found in the resource id %q", input)
	}

	if id.ConfigStoreName, ok = parsed.Parsed["configStoreName"]; !ok {
		return nil, fmt.Errorf("the segment 'configStoreName' was not found in the resource id %q", input)
	}

	if id.PrivateEndpointConnectionName, ok = parsed.Parsed["privateEndpointConnectionName"]; !ok {
		return nil, fmt.Errorf("the segment 'privateEndpointConnectionName' was not found in the resource id %q", input)
	}

	return &id, nil
}

// ValidatePrivateEndpointConnectionID checks that 'input' can be parsed as a Private Endpoint Connection ID
func ValidatePrivateEndpointConnectionID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if _, err := ParsePrivateEndpointConnectionID(v); err != nil {
		errors = append(errors, err)
	}

	return
}

// ID returns the formatted Private Endpoint Connection ID
func (id PrivateEndpointConnectionId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.AppConfiguration/configurationStores/%s/privateEndpointConnections/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroupName, id.ConfigStoreName, id.PrivateEndpointConnectionName)
}

// Segments returns a slice of Resource ID Segments which comprise this Private Endpoint Connection ID
func (id PrivateEndpointConnectionId) Segments() []resourceids.Segment {
	return []resourceids.Segment{
		resourceids.StaticSegment("subscriptions", "subscriptions", "subscriptions"),
		resourceids.SubscriptionIdSegment("subscriptionId", "12345678-1234-9876-4563-123456789012"),
		resourceids.StaticSegment("resourceGroups", "resourceGroups", "resourceGroups"),
		resourceids.ResourceGroupSegment("resourceGroupName", "example-resource-group"),
		resourceids.StaticSegment("providers", "providers", "providers"),
		resourceids.ResourceProviderSegment("microsoftAppConfiguration", "Microsoft.AppConfiguration", "Microsoft.AppConfiguration"),
		resourceids.StaticSegment("configurationStores", "configurationStores", "configurationStores"),
		resourceids.UserSpecifiedSegment("configStoreName", "configStoreValue"),
		resourceids.StaticSegment("privateEndpointConnections", "privateEndpointConnections", "privateEndpointConnections"),
		resourceids.UserSpecifiedSegment("privateEndpointConnectionName", "privateEndpointConnectionValue"),
	}
}

// String returns a human-readable description of this Private Endpoint Connection ID
func (id PrivateEndpointConnectionId) String() string {
	components := []string{
		fmt.Sprintf("Subscription: %q", id.SubscriptionId),
		fmt.Sprintf("Resource Group Name: %q", id.ResourceGroupName),
		fmt.Sprintf("Config Store Name: %q", id.ConfigStoreName),
		fmt.Sprintf("Private Endpoint Connection Name: %q", id.PrivateEndpointConnectionName),
	}
	return fmt.Sprintf("Private Endpoint Connection (%s)", strings.Join(components, "\n"))
}
