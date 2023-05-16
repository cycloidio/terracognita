package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

type ApiKeyId struct {
	SubscriptionId string
	ResourceGroup  string
	ComponentName  string
	Name           string
}

func NewApiKeyID(subscriptionId, resourceGroup, componentName, name string) ApiKeyId {
	return ApiKeyId{
		SubscriptionId: subscriptionId,
		ResourceGroup:  resourceGroup,
		ComponentName:  componentName,
		Name:           name,
	}
}

func (id ApiKeyId) String() string {
	segments := []string{
		fmt.Sprintf("Name %q", id.Name),
		fmt.Sprintf("Component Name %q", id.ComponentName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Api Key", segmentsStr)
}

func (id ApiKeyId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Insights/components/%s/apiKeys/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.ComponentName, id.Name)
}

// ApiKeyID parses a ApiKey ID into an ApiKeyId struct
func ApiKeyID(input string) (*ApiKeyId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := ApiKeyId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.ComponentName, err = id.PopSegment("components"); err != nil {
		return nil, err
	}
	if resourceId.Name, err = id.PopSegment("apiKeys"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}

// ApiKeyIDInsensitively parses an ApiKey ID into an ApiKeyId struct, insensitively
// This should only be used to parse an ID for rewriting, the ApiKeyID
// method should be used instead for validation etc.
//
// Whilst this may seem strange, this enables Terraform have consistent casing
// which works around issues in Core, whilst handling broken API responses.
func ApiKeyIDInsensitively(input string) (*ApiKeyId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := ApiKeyId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	// find the correct casing for the 'components' segment
	componentsKey := "components"
	for key := range id.Path {
		if strings.EqualFold(key, componentsKey) {
			componentsKey = key
			break
		}
	}
	if resourceId.ComponentName, err = id.PopSegment(componentsKey); err != nil {
		return nil, err
	}

	// find the correct casing for the 'apiKeys' segment
	apiKeysKey := "apiKeys"
	for key := range id.Path {
		if strings.EqualFold(key, apiKeysKey) {
			apiKeysKey = key
			break
		}
	}
	if resourceId.Name, err = id.PopSegment(apiKeysKey); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
