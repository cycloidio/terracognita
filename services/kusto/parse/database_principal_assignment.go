package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

type DatabasePrincipalAssignmentId struct {
	SubscriptionId          string
	ResourceGroup           string
	ClusterName             string
	DatabaseName            string
	PrincipalAssignmentName string
}

func NewDatabasePrincipalAssignmentID(subscriptionId, resourceGroup, clusterName, databaseName, principalAssignmentName string) DatabasePrincipalAssignmentId {
	return DatabasePrincipalAssignmentId{
		SubscriptionId:          subscriptionId,
		ResourceGroup:           resourceGroup,
		ClusterName:             clusterName,
		DatabaseName:            databaseName,
		PrincipalAssignmentName: principalAssignmentName,
	}
}

func (id DatabasePrincipalAssignmentId) String() string {
	segments := []string{
		fmt.Sprintf("Principal Assignment Name %q", id.PrincipalAssignmentName),
		fmt.Sprintf("Database Name %q", id.DatabaseName),
		fmt.Sprintf("Cluster Name %q", id.ClusterName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Database Principal Assignment", segmentsStr)
}

func (id DatabasePrincipalAssignmentId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Kusto/Clusters/%s/Databases/%s/PrincipalAssignments/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.ClusterName, id.DatabaseName, id.PrincipalAssignmentName)
}

// DatabasePrincipalAssignmentID parses a DatabasePrincipalAssignment ID into an DatabasePrincipalAssignmentId struct
func DatabasePrincipalAssignmentID(input string) (*DatabasePrincipalAssignmentId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := DatabasePrincipalAssignmentId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.ClusterName, err = id.PopSegment("Clusters"); err != nil {
		return nil, err
	}
	if resourceId.DatabaseName, err = id.PopSegment("Databases"); err != nil {
		return nil, err
	}
	if resourceId.PrincipalAssignmentName, err = id.PopSegment("PrincipalAssignments"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
