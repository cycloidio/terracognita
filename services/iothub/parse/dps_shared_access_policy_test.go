package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"testing"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

var _ resourceids.Id = DpsSharedAccessPolicyId{}

func TestDpsSharedAccessPolicyIDFormatter(t *testing.T) {
	actual := NewDpsSharedAccessPolicyID("12345678-1234-9876-4563-123456789012", "resGroup1", "provisioningService1", "sharedAccessPolicy1").ID()
	expected := "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Devices/provisioningServices/provisioningService1/keys/sharedAccessPolicy1"
	if actual != expected {
		t.Fatalf("Expected %q but got %q", expected, actual)
	}
}

func TestDpsSharedAccessPolicyID(t *testing.T) {
	testData := []struct {
		Input    string
		Error    bool
		Expected *DpsSharedAccessPolicyId
	}{

		{
			// empty
			Input: "",
			Error: true,
		},

		{
			// missing SubscriptionId
			Input: "/",
			Error: true,
		},

		{
			// missing value for SubscriptionId
			Input: "/subscriptions/",
			Error: true,
		},

		{
			// missing ResourceGroup
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/",
			Error: true,
		},

		{
			// missing value for ResourceGroup
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/",
			Error: true,
		},

		{
			// missing ProvisioningServiceName
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Devices/",
			Error: true,
		},

		{
			// missing value for ProvisioningServiceName
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Devices/provisioningServices/",
			Error: true,
		},

		{
			// missing KeyName
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Devices/provisioningServices/provisioningService1/",
			Error: true,
		},

		{
			// missing value for KeyName
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Devices/provisioningServices/provisioningService1/keys/",
			Error: true,
		},

		{
			// valid
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Devices/provisioningServices/provisioningService1/keys/sharedAccessPolicy1",
			Expected: &DpsSharedAccessPolicyId{
				SubscriptionId:          "12345678-1234-9876-4563-123456789012",
				ResourceGroup:           "resGroup1",
				ProvisioningServiceName: "provisioningService1",
				KeyName:                 "sharedAccessPolicy1",
			},
		},

		{
			// upper-cased
			Input: "/SUBSCRIPTIONS/12345678-1234-9876-4563-123456789012/RESOURCEGROUPS/RESGROUP1/PROVIDERS/MICROSOFT.DEVICES/PROVISIONINGSERVICES/PROVISIONINGSERVICE1/KEYS/SHAREDACCESSPOLICY1",
			Error: true,
		},
	}

	for _, v := range testData {
		t.Logf("[DEBUG] Testing %q", v.Input)

		actual, err := DpsSharedAccessPolicyID(v.Input)
		if err != nil {
			if v.Error {
				continue
			}

			t.Fatalf("Expect a value but got an error: %s", err)
		}
		if v.Error {
			t.Fatal("Expect an error but didn't get one")
		}

		if actual.SubscriptionId != v.Expected.SubscriptionId {
			t.Fatalf("Expected %q but got %q for SubscriptionId", v.Expected.SubscriptionId, actual.SubscriptionId)
		}
		if actual.ResourceGroup != v.Expected.ResourceGroup {
			t.Fatalf("Expected %q but got %q for ResourceGroup", v.Expected.ResourceGroup, actual.ResourceGroup)
		}
		if actual.ProvisioningServiceName != v.Expected.ProvisioningServiceName {
			t.Fatalf("Expected %q but got %q for ProvisioningServiceName", v.Expected.ProvisioningServiceName, actual.ProvisioningServiceName)
		}
		if actual.KeyName != v.Expected.KeyName {
			t.Fatalf("Expected %q but got %q for KeyName", v.Expected.KeyName, actual.KeyName)
		}
	}
}
