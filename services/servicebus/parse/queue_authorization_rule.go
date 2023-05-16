package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

type QueueAuthorizationRuleId struct {
	SubscriptionId        string
	ResourceGroup         string
	NamespaceName         string
	QueueName             string
	AuthorizationRuleName string
}

func NewQueueAuthorizationRuleID(subscriptionId, resourceGroup, namespaceName, queueName, authorizationRuleName string) QueueAuthorizationRuleId {
	return QueueAuthorizationRuleId{
		SubscriptionId:        subscriptionId,
		ResourceGroup:         resourceGroup,
		NamespaceName:         namespaceName,
		QueueName:             queueName,
		AuthorizationRuleName: authorizationRuleName,
	}
}

func (id QueueAuthorizationRuleId) String() string {
	segments := []string{
		fmt.Sprintf("Authorization Rule Name %q", id.AuthorizationRuleName),
		fmt.Sprintf("Queue Name %q", id.QueueName),
		fmt.Sprintf("Namespace Name %q", id.NamespaceName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Queue Authorization Rule", segmentsStr)
}

func (id QueueAuthorizationRuleId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ServiceBus/namespaces/%s/queues/%s/authorizationRules/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.NamespaceName, id.QueueName, id.AuthorizationRuleName)
}

// QueueAuthorizationRuleID parses a QueueAuthorizationRule ID into an QueueAuthorizationRuleId struct
func QueueAuthorizationRuleID(input string) (*QueueAuthorizationRuleId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := QueueAuthorizationRuleId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.NamespaceName, err = id.PopSegment("namespaces"); err != nil {
		return nil, err
	}
	if resourceId.QueueName, err = id.PopSegment("queues"); err != nil {
		return nil, err
	}
	if resourceId.AuthorizationRuleName, err = id.PopSegment("authorizationRules"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
