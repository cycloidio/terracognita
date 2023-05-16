package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

type WorkspaceSecurityAlertPolicyId struct {
	SubscriptionId          string
	ResourceGroup           string
	WorkspaceName           string
	SecurityAlertPolicyName string
}

func NewWorkspaceSecurityAlertPolicyID(subscriptionId, resourceGroup, workspaceName, securityAlertPolicyName string) WorkspaceSecurityAlertPolicyId {
	return WorkspaceSecurityAlertPolicyId{
		SubscriptionId:          subscriptionId,
		ResourceGroup:           resourceGroup,
		WorkspaceName:           workspaceName,
		SecurityAlertPolicyName: securityAlertPolicyName,
	}
}

func (id WorkspaceSecurityAlertPolicyId) String() string {
	segments := []string{
		fmt.Sprintf("Security Alert Policy Name %q", id.SecurityAlertPolicyName),
		fmt.Sprintf("Workspace Name %q", id.WorkspaceName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Workspace Security Alert Policy", segmentsStr)
}

func (id WorkspaceSecurityAlertPolicyId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Synapse/workspaces/%s/securityAlertPolicies/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.WorkspaceName, id.SecurityAlertPolicyName)
}

// WorkspaceSecurityAlertPolicyID parses a WorkspaceSecurityAlertPolicy ID into an WorkspaceSecurityAlertPolicyId struct
func WorkspaceSecurityAlertPolicyID(input string) (*WorkspaceSecurityAlertPolicyId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := WorkspaceSecurityAlertPolicyId{
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
	if resourceId.SecurityAlertPolicyName, err = id.PopSegment("securityAlertPolicies"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
