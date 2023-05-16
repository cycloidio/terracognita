package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

type ApiId struct {
	SubscriptionId string
	ResourceGroup  string
	ServiceName    string
	Name           string
}

func NewApiID(subscriptionId, resourceGroup, serviceName, name string) ApiId {
	return ApiId{
		SubscriptionId: subscriptionId,
		ResourceGroup:  resourceGroup,
		ServiceName:    serviceName,
		Name:           name,
	}
}

func (id ApiId) String() string {
	segments := []string{
		fmt.Sprintf("Name %q", id.Name),
		fmt.Sprintf("Service Name %q", id.ServiceName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Api", segmentsStr)
}

func (id ApiId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ApiManagement/service/%s/apis/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.ServiceName, id.Name)
}

// ApiID parses a Api ID into an ApiId struct
func ApiID(input string) (*ApiId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := ApiId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.ServiceName, err = id.PopSegment("service"); err != nil {
		return nil, err
	}
	if resourceId.Name, err = id.PopSegment("apis"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
