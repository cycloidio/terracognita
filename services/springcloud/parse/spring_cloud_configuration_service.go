package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

type SpringCloudConfigurationServiceId struct {
	SubscriptionId           string
	ResourceGroup            string
	SpringName               string
	ConfigurationServiceName string
}

func NewSpringCloudConfigurationServiceID(subscriptionId, resourceGroup, springName, configurationServiceName string) SpringCloudConfigurationServiceId {
	return SpringCloudConfigurationServiceId{
		SubscriptionId:           subscriptionId,
		ResourceGroup:            resourceGroup,
		SpringName:               springName,
		ConfigurationServiceName: configurationServiceName,
	}
}

func (id SpringCloudConfigurationServiceId) String() string {
	segments := []string{
		fmt.Sprintf("Configuration Service Name %q", id.ConfigurationServiceName),
		fmt.Sprintf("Spring Name %q", id.SpringName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Spring Cloud Configuration Service", segmentsStr)
}

func (id SpringCloudConfigurationServiceId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.AppPlatform/Spring/%s/configurationServices/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.SpringName, id.ConfigurationServiceName)
}

// SpringCloudConfigurationServiceID parses a SpringCloudConfigurationService ID into an SpringCloudConfigurationServiceId struct
func SpringCloudConfigurationServiceID(input string) (*SpringCloudConfigurationServiceId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := SpringCloudConfigurationServiceId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.SpringName, err = id.PopSegment("Spring"); err != nil {
		return nil, err
	}
	if resourceId.ConfigurationServiceName, err = id.PopSegment("configurationServices"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
