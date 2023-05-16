package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

type LogAnalyticsLinkedStorageAccountId struct {
	SubscriptionId           string
	ResourceGroup            string
	WorkspaceName            string
	LinkedStorageAccountName string
}

func NewLogAnalyticsLinkedStorageAccountID(subscriptionId, resourceGroup, workspaceName, linkedStorageAccountName string) LogAnalyticsLinkedStorageAccountId {
	return LogAnalyticsLinkedStorageAccountId{
		SubscriptionId:           subscriptionId,
		ResourceGroup:            resourceGroup,
		WorkspaceName:            workspaceName,
		LinkedStorageAccountName: linkedStorageAccountName,
	}
}

func (id LogAnalyticsLinkedStorageAccountId) String() string {
	segments := []string{
		fmt.Sprintf("Linked Storage Account Name %q", id.LinkedStorageAccountName),
		fmt.Sprintf("Workspace Name %q", id.WorkspaceName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Log Analytics Linked Storage Account", segmentsStr)
}

func (id LogAnalyticsLinkedStorageAccountId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.OperationalInsights/workspaces/%s/linkedStorageAccounts/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.WorkspaceName, id.LinkedStorageAccountName)
}

// LogAnalyticsLinkedStorageAccountID parses a LogAnalyticsLinkedStorageAccount ID into an LogAnalyticsLinkedStorageAccountId struct
func LogAnalyticsLinkedStorageAccountID(input string) (*LogAnalyticsLinkedStorageAccountId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := LogAnalyticsLinkedStorageAccountId{
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
	if resourceId.LinkedStorageAccountName, err = id.PopSegment("linkedStorageAccounts"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
