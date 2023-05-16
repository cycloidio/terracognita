package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"testing"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
)

var _ resourceids.Id = LogzTagRuleId{}

func TestLogzTagRuleIDFormatter(t *testing.T) {
	actual := NewLogzTagRuleID("12345678-1234-9876-4563-123456789012", "resourceGroup1", "monitor1", "ruleSet1").ID()
	expected := "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resourceGroup1/providers/Microsoft.Logz/monitors/monitor1/tagRules/ruleSet1"
	if actual != expected {
		t.Fatalf("Expected %q but got %q", expected, actual)
	}
}

func TestLogzTagRuleID(t *testing.T) {
	testData := []struct {
		Input    string
		Error    bool
		Expected *LogzTagRuleId
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
			// missing MonitorName
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resourceGroup1/providers/Microsoft.Logz/",
			Error: true,
		},

		{
			// missing value for MonitorName
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resourceGroup1/providers/Microsoft.Logz/monitors/",
			Error: true,
		},

		{
			// missing TagRuleName
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resourceGroup1/providers/Microsoft.Logz/monitors/monitor1/",
			Error: true,
		},

		{
			// missing value for TagRuleName
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resourceGroup1/providers/Microsoft.Logz/monitors/monitor1/tagRules/",
			Error: true,
		},

		{
			// valid
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resourceGroup1/providers/Microsoft.Logz/monitors/monitor1/tagRules/ruleSet1",
			Expected: &LogzTagRuleId{
				SubscriptionId: "12345678-1234-9876-4563-123456789012",
				ResourceGroup:  "resourceGroup1",
				MonitorName:    "monitor1",
				TagRuleName:    "ruleSet1",
			},
		},

		{
			// upper-cased
			Input: "/SUBSCRIPTIONS/12345678-1234-9876-4563-123456789012/RESOURCEGROUPS/RESOURCEGROUP1/PROVIDERS/MICROSOFT.LOGZ/MONITORS/MONITOR1/TAGRULES/RULESET1",
			Error: true,
		},
	}

	for _, v := range testData {
		t.Logf("[DEBUG] Testing %q", v.Input)

		actual, err := LogzTagRuleID(v.Input)
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
		if actual.MonitorName != v.Expected.MonitorName {
			t.Fatalf("Expected %q but got %q for MonitorName", v.Expected.MonitorName, actual.MonitorName)
		}
		if actual.TagRuleName != v.Expected.TagRuleName {
			t.Fatalf("Expected %q but got %q for TagRuleName", v.Expected.TagRuleName, actual.TagRuleName)
		}
	}
}
