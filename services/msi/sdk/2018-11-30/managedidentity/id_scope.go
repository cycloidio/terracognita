package managedidentity

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

var _ resourceids.ResourceId = ScopeId{}

// ScopeId is a struct representing the Resource ID for a Scope
type ScopeId struct {
	Scope string
}

// NewScopeID returns a new ScopeId struct
func NewScopeID(scope string) ScopeId {
	return ScopeId{
		Scope: scope,
	}
}

// ParseScopeID parses 'input' into a ScopeId
func ParseScopeID(input string) (*ScopeId, error) {
	parser := resourceids.NewParserFromResourceIdType(ScopeId{})
	parsed, err := parser.Parse(input, false)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	var ok bool
	id := ScopeId{}

	if id.Scope, ok = parsed.Parsed["scope"]; !ok {
		return nil, fmt.Errorf("the segment 'scope' was not found in the resource id %q", input)
	}

	return &id, nil
}

// ParseScopeIDInsensitively parses 'input' case-insensitively into a ScopeId
// note: this method should only be used for API response data and not user input
func ParseScopeIDInsensitively(input string) (*ScopeId, error) {
	parser := resourceids.NewParserFromResourceIdType(ScopeId{})
	parsed, err := parser.Parse(input, true)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %+v", input, err)
	}

	var ok bool
	id := ScopeId{}

	if id.Scope, ok = parsed.Parsed["scope"]; !ok {
		return nil, fmt.Errorf("the segment 'scope' was not found in the resource id %q", input)
	}

	return &id, nil
}

// ValidateScopeID checks that 'input' can be parsed as a Scope ID
func ValidateScopeID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if _, err := ParseScopeID(v); err != nil {
		errors = append(errors, err)
	}

	return
}

// ID returns the formatted Scope ID
func (id ScopeId) ID() string {
	fmtString := "/%s"
	return fmt.Sprintf(fmtString, strings.TrimPrefix(id.Scope, "/"))
}

// Segments returns a slice of Resource ID Segments which comprise this Scope ID
func (id ScopeId) Segments() []resourceids.Segment {
	return []resourceids.Segment{
		resourceids.ScopeSegment("scope", "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/some-resource-group"),
	}
}

// String returns a human-readable description of this Scope ID
func (id ScopeId) String() string {
	components := []string{
		fmt.Sprintf("Scope: %q", id.Scope),
	}
	return fmt.Sprintf("Scope (%s)", strings.Join(components, "\n"))
}
