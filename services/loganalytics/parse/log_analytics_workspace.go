package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

type LogAnalyticsWorkspaceId struct {
	SubscriptionId string
	ResourceGroup  string
	WorkspaceName  string
}

func NewLogAnalyticsWorkspaceID(subscriptionId, resourceGroup, workspaceName string) LogAnalyticsWorkspaceId {
	return LogAnalyticsWorkspaceId{
		SubscriptionId: subscriptionId,
		ResourceGroup:  resourceGroup,
		WorkspaceName:  workspaceName,
	}
}

func (id LogAnalyticsWorkspaceId) String() string {
	segments := []string{
		fmt.Sprintf("Workspace Name %q", id.WorkspaceName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Log Analytics Workspace", segmentsStr)
}

func (id LogAnalyticsWorkspaceId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.OperationalInsights/workspaces/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.WorkspaceName)
}

// LogAnalyticsWorkspaceID parses a LogAnalyticsWorkspace ID into an LogAnalyticsWorkspaceId struct
func LogAnalyticsWorkspaceID(input string) (*LogAnalyticsWorkspaceId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := LogAnalyticsWorkspaceId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.WorkspaceName, err = id.PopSegment("workspaces"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
