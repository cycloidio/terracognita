package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

type AccessPolicyId struct {
	SubscriptionId  string
	ResourceGroup   string
	EnvironmentName string
	Name            string
}

func NewAccessPolicyID(subscriptionId, resourceGroup, environmentName, name string) AccessPolicyId {
	return AccessPolicyId{
		SubscriptionId:  subscriptionId,
		ResourceGroup:   resourceGroup,
		EnvironmentName: environmentName,
		Name:            name,
	}
}

func (id AccessPolicyId) String() string {
	segments := []string{
		fmt.Sprintf("Name %q", id.Name),
		fmt.Sprintf("Environment Name %q", id.EnvironmentName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Access Policy", segmentsStr)
}

func (id AccessPolicyId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.TimeSeriesInsights/environments/%s/accessPolicies/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.EnvironmentName, id.Name)
}

// AccessPolicyID parses a AccessPolicy ID into an AccessPolicyId struct
func AccessPolicyID(input string) (*AccessPolicyId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := AccessPolicyId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.EnvironmentName, err = id.PopSegment("environments"); err != nil {
		return nil, err
	}
	if resourceId.Name, err = id.PopSegment("accessPolicies"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
