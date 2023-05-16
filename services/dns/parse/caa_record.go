package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

type CaaRecordId struct {
	SubscriptionId string
	ResourceGroup  string
	DnszoneName    string
	CAAName        string
}

func NewCaaRecordID(subscriptionId, resourceGroup, dnszoneName, cAAName string) CaaRecordId {
	return CaaRecordId{
		SubscriptionId: subscriptionId,
		ResourceGroup:  resourceGroup,
		DnszoneName:    dnszoneName,
		CAAName:        cAAName,
	}
}

func (id CaaRecordId) String() string {
	segments := []string{
		fmt.Sprintf("C A A Name %q", id.CAAName),
		fmt.Sprintf("Dnszone Name %q", id.DnszoneName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Caa Record", segmentsStr)
}

func (id CaaRecordId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/dnszones/%s/CAA/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.DnszoneName, id.CAAName)
}

// CaaRecordID parses a CaaRecord ID into an CaaRecordId struct
func CaaRecordID(input string) (*CaaRecordId, error) {
	id, err := resourceids.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := CaaRecordId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.DnszoneName, err = id.PopSegment("dnszones"); err != nil {
		return nil, err
	}
	if resourceId.CAAName, err = id.PopSegment("CAA"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
