package hybridconnections

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

var _ resourceids.ResourceId = HybridConnectionAuthorizationRuleId{}

// HybridConnectionAuthorizationRuleId is a struct representing the Resource ID for a Hybrid Connection Authorization Rule
type HybridConnectionAuthorizationRuleId struct {
	SubscriptionId        string
	ResourceGroupName     string
	NamespaceName         string
	HybridConnectionName  string
	AuthorizationRuleName string
}

// NewHybridConnectionAuthorizationRuleID returns a new HybridConnectionAuthorizationRuleId struct
func NewHybridConnectionAuthorizationRuleID(subscriptionId string, resourceGroupName string, namespaceName string, hybridConnectionName string, authorizationRuleName string) HybridConnectionAuthorizationRuleId {
	return HybridConnectionAuthorizationRuleId{
		SubscriptionId:        subscriptionId,
		ResourceGroupName:     resourceGroupName,
		NamespaceName:         namespaceName,
		HybridConnectionName:  hybridConnectionName,
		AuthorizationRuleName: authorizationRuleName,
	}
}

// ParseHybridConnectionAuthorizationRuleID parses 'input' into a HybridConnectionAuthorizationRuleId
func ParseHybridConnectionAuthorizationRuleID(input string) (*HybridConnectionAuthorizationRuleId, error) {
	parser := resourceids.NewParserFromResourceIdType(HybridConnectionAuthorizationRuleId{})
	parsed, err := parser.Parse(input, false)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	var ok bool
	id := HybridConnectionAuthorizationRuleId{}

	if id.SubscriptionId, ok = parsed.Parsed["subscriptionId"]; !ok {
		return nil, fmt.Errorf("the segment 'subscriptionId' was not found in the resource id %q", input)
	}

	if id.ResourceGroupName, ok = parsed.Parsed["resourceGroupName"]; !ok {
		return nil, fmt.Errorf("the segment 'resourceGroupName' was not found in the resource id %q", input)
	}

	if id.NamespaceName, ok = parsed.Parsed["namespaceName"]; !ok {
		return nil, fmt.Errorf("the segment 'namespaceName' was not found in the resource id %q", input)
	}

	if id.HybridConnectionName, ok = parsed.Parsed["hybridConnectionName"]; !ok {
		return nil, fmt.Errorf("the segment 'hybridConnectionName' was not found in the resource id %q", input)
	}

	if id.AuthorizationRuleName, ok = parsed.Parsed["authorizationRuleName"]; !ok {
		return nil, fmt.Errorf("the segment 'authorizationRuleName' was not found in the resource id %q", input)
	}

	return &id, nil
}

// ParseHybridConnectionAuthorizationRuleIDInsensitively parses 'input' case-insensitively into a HybridConnectionAuthorizationRuleId
// note: this method should only be used for API response data and not user input
func ParseHybridConnectionAuthorizationRuleIDInsensitively(input string) (*HybridConnectionAuthorizationRuleId, error) {
	parser := resourceids.NewParserFromResourceIdType(HybridConnectionAuthorizationRuleId{})
	parsed, err := parser.Parse(input, true)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	var ok bool
	id := HybridConnectionAuthorizationRuleId{}

	if id.SubscriptionId, ok = parsed.Parsed["subscriptionId"]; !ok {
		return nil, fmt.Errorf("the segment 'subscriptionId' was not found in the resource id %q", input)
	}

	if id.ResourceGroupName, ok = parsed.Parsed["resourceGroupName"]; !ok {
		return nil, fmt.Errorf("the segment 'resourceGroupName' was not found in the resource id %q", input)
	}

	if id.NamespaceName, ok = parsed.Parsed["namespaceName"]; !ok {
		return nil, fmt.Errorf("the segment 'namespaceName' was not found in the resource id %q", input)
	}

	if id.HybridConnectionName, ok = parsed.Parsed["hybridConnectionName"]; !ok {
		return nil, fmt.Errorf("the segment 'hybridConnectionName' was not found in the resource id %q", input)
	}

	if id.AuthorizationRuleName, ok = parsed.Parsed["authorizationRuleName"]; !ok {
		return nil, fmt.Errorf("the segment 'authorizationRuleName' was not found in the resource id %q", input)
	}

	return &id, nil
}

// ValidateHybridConnectionAuthorizationRuleID checks that 'input' can be parsed as a Hybrid Connection Authorization Rule ID
func ValidateHybridConnectionAuthorizationRuleID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if _, err := ParseHybridConnectionAuthorizationRuleID(v); err != nil {
		errors = append(errors, err)
	}

	return
}

// ID returns the formatted Hybrid Connection Authorization Rule ID
func (id HybridConnectionAuthorizationRuleId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Relay/namespaces/%s/hybridConnections/%s/authorizationRules/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroupName, id.NamespaceName, id.HybridConnectionName, id.AuthorizationRuleName)
}

// Segments returns a slice of Resource ID Segments which comprise this Hybrid Connection Authorization Rule ID
func (id HybridConnectionAuthorizationRuleId) Segments() []resourceids.Segment {
	return []resourceids.Segment{
		resourceids.StaticSegment("subscriptions", "subscriptions", "subscriptions"),
		resourceids.SubscriptionIdSegment("subscriptionId", "12345678-1234-9876-4563-123456789012"),
		resourceids.StaticSegment("resourceGroups", "resourceGroups", "resourceGroups"),
		resourceids.ResourceGroupSegment("resourceGroupName", "example-resource-group"),
		resourceids.StaticSegment("providers", "providers", "providers"),
		resourceids.ResourceProviderSegment("microsoftRelay", "Microsoft.Relay", "Microsoft.Relay"),
		resourceids.StaticSegment("namespaces", "namespaces", "namespaces"),
		resourceids.UserSpecifiedSegment("namespaceName", "namespaceValue"),
		resourceids.StaticSegment("hybridConnections", "hybridConnections", "hybridConnections"),
		resourceids.UserSpecifiedSegment("hybridConnectionName", "hybridConnectionValue"),
		resourceids.StaticSegment("authorizationRules", "authorizationRules", "authorizationRules"),
		resourceids.UserSpecifiedSegment("authorizationRuleName", "authorizationRuleValue"),
	}
}

// String returns a human-readable description of this Hybrid Connection Authorization Rule ID
func (id HybridConnectionAuthorizationRuleId) String() string {
	components := []string{
		fmt.Sprintf("Subscription: %q", id.SubscriptionId),
		fmt.Sprintf("Resource Group Name: %q", id.ResourceGroupName),
		fmt.Sprintf("Namespace Name: %q", id.NamespaceName),
		fmt.Sprintf("Hybrid Connection Name: %q", id.HybridConnectionName),
		fmt.Sprintf("Authorization Rule Name: %q", id.AuthorizationRuleName),
	}
	return fmt.Sprintf("Hybrid Connection Authorization Rule (%s)", strings.Join(components, "\n"))
}
