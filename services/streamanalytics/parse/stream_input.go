package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

type StreamInputId struct {
	SubscriptionId   string
	ResourceGroup    string
	StreamingjobName string
	InputName        string
}

func NewStreamInputID(subscriptionId, resourceGroup, streamingjobName, inputName string) StreamInputId {
	return StreamInputId{
		SubscriptionId:   subscriptionId,
		ResourceGroup:    resourceGroup,
		StreamingjobName: streamingjobName,
		InputName:        inputName,
	}
}

func (id StreamInputId) String() string {
	segments := []string{
		fmt.Sprintf("Input Name %q", id.InputName),
		fmt.Sprintf("Streamingjob Name %q", id.StreamingjobName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Stream Input", segmentsStr)
}

func (id StreamInputId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.StreamAnalytics/streamingjobs/%s/inputs/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.StreamingjobName, id.InputName)
}

// StreamInputID parses a StreamInput ID into an StreamInputId struct
func StreamInputID(input string) (*StreamInputId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := StreamInputId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.StreamingjobName, err = id.PopSegment("streamingjobs"); err != nil {
		return nil, err
	}
	if resourceId.InputName, err = id.PopSegment("inputs"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
