package cognitiveservicesaccounts

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

var _ resourceids.ResourceId = DeletedAccountId{}

// DeletedAccountId is a struct representing the Resource ID for a Deleted Account
type DeletedAccountId struct {
	SubscriptionId    string
	Location          string
	ResourceGroupName string
	AccountName       string
}

// NewDeletedAccountID returns a new DeletedAccountId struct
func NewDeletedAccountID(subscriptionId string, location string, resourceGroupName string, accountName string) DeletedAccountId {
	return DeletedAccountId{
		SubscriptionId:    subscriptionId,
		Location:          location,
		ResourceGroupName: resourceGroupName,
		AccountName:       accountName,
	}
}

// ParseDeletedAccountID parses 'input' into a DeletedAccountId
func ParseDeletedAccountID(input string) (*DeletedAccountId, error) {
	parser := resourceids.NewParserFromResourceIdType(DeletedAccountId{})
	parsed, err := parser.Parse(input, false)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	var ok bool
	id := DeletedAccountId{}

	if id.SubscriptionId, ok = parsed.Parsed["subscriptionId"]; !ok {
		return nil, fmt.Errorf("the segment 'subscriptionId' was not found in the resource id %q", input)
	}

	if id.Location, ok = parsed.Parsed["location"]; !ok {
		return nil, fmt.Errorf("the segment 'location' was not found in the resource id %q", input)
	}

	if id.ResourceGroupName, ok = parsed.Parsed["resourceGroupName"]; !ok {
		return nil, fmt.Errorf("the segment 'resourceGroupName' was not found in the resource id %q", input)
	}

	if id.AccountName, ok = parsed.Parsed["accountName"]; !ok {
		return nil, fmt.Errorf("the segment 'accountName' was not found in the resource id %q", input)
	}

	return &id, nil
}

// ParseDeletedAccountIDInsensitively parses 'input' case-insensitively into a DeletedAccountId
// note: this method should only be used for API response data and not user input
func ParseDeletedAccountIDInsensitively(input string) (*DeletedAccountId, error) {
	parser := resourceids.NewParserFromResourceIdType(DeletedAccountId{})
	parsed, err := parser.Parse(input, true)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	var ok bool
	id := DeletedAccountId{}

	if id.SubscriptionId, ok = parsed.Parsed["subscriptionId"]; !ok {
		return nil, fmt.Errorf("the segment 'subscriptionId' was not found in the resource id %q", input)
	}

	if id.Location, ok = parsed.Parsed["location"]; !ok {
		return nil, fmt.Errorf("the segment 'location' was not found in the resource id %q", input)
	}

	if id.ResourceGroupName, ok = parsed.Parsed["resourceGroupName"]; !ok {
		return nil, fmt.Errorf("the segment 'resourceGroupName' was not found in the resource id %q", input)
	}

	if id.AccountName, ok = parsed.Parsed["accountName"]; !ok {
		return nil, fmt.Errorf("the segment 'accountName' was not found in the resource id %q", input)
	}

	return &id, nil
}

// ValidateDeletedAccountID checks that 'input' can be parsed as a Deleted Account ID
func ValidateDeletedAccountID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if _, err := ParseDeletedAccountID(v); err != nil {
		errors = append(errors, err)
	}

	return
}

// ID returns the formatted Deleted Account ID
func (id DeletedAccountId) ID() string {
	fmtString := "/subscriptions/%s/providers/Microsoft.CognitiveServices/locations/%s/resourceGroups/%s/deletedAccounts/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.Location, id.ResourceGroupName, id.AccountName)
}

// Segments returns a slice of Resource ID Segments which comprise this Deleted Account ID
func (id DeletedAccountId) Segments() []resourceids.Segment {
	return []resourceids.Segment{
		resourceids.StaticSegment("subscriptions", "subscriptions", "subscriptions"),
		resourceids.SubscriptionIdSegment("subscriptionId", "12345678-1234-9876-4563-123456789012"),
		resourceids.StaticSegment("providers", "providers", "providers"),
		resourceids.ResourceProviderSegment("microsoftCognitiveServices", "Microsoft.CognitiveServices", "Microsoft.CognitiveServices"),
		resourceids.StaticSegment("locations", "locations", "locations"),
		resourceids.UserSpecifiedSegment("location", "locationValue"),
		resourceids.StaticSegment("resourceGroups", "resourceGroups", "resourceGroups"),
		resourceids.ResourceGroupSegment("resourceGroupName", "example-resource-group"),
		resourceids.StaticSegment("deletedAccounts", "deletedAccounts", "deletedAccounts"),
		resourceids.UserSpecifiedSegment("accountName", "accountValue"),
	}
}

// String returns a human-readable description of this Deleted Account ID
func (id DeletedAccountId) String() string {
	components := []string{
		fmt.Sprintf("Subscription: %q", id.SubscriptionId),
		fmt.Sprintf("Location: %q", id.Location),
		fmt.Sprintf("Resource Group Name: %q", id.ResourceGroupName),
		fmt.Sprintf("Account Name: %q", id.AccountName),
	}
	return fmt.Sprintf("Deleted Account (%s)", strings.Join(components, "\n"))
}
