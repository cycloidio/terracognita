package migration

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func TestLogProfileV0ToV1(t *testing.T) {
	testData := []struct {
		name     string
		input    map[string]interface{}
		expected *string
	}{
		{
			name: "old id",
			input: map[string]interface{}{
				"id": "/subscriptions/12345678-1234-5678-1234-123456789012/providers/microsoft.insights/logprofiles/profile1",
			},
			expected: utils.String("/subscriptions/12345678-1234-5678-1234-123456789012/providers/Microsoft.Insights/logProfiles/profile1"),
		},
		{
			name: "old id - mixed case",
			input: map[string]interface{}{
				"id": "/subscriptions/12345678-1234-5678-1234-123456789012/providers/microsoft.insights/LogProfiles/profile1",
			},
			expected: utils.String("/subscriptions/12345678-1234-5678-1234-123456789012/providers/Microsoft.Insights/logProfiles/profile1"),
		},
		{
			name: "new id",
			input: map[string]interface{}{
				"id": "/subscriptions/12345678-1234-5678-1234-123456789012/providers/Microsoft.Insights/logProfiles/profile1",
			},
			expected: utils.String("/subscriptions/12345678-1234-5678-1234-123456789012/providers/Microsoft.Insights/logProfiles/profile1"),
		},
	}
	for _, test := range testData {
		t.Logf("Testing %q...", test.name)
		result, err := LogProfileUpgradeV0ToV1{}.UpgradeFunc()(context.TODO(), test.input, nil)
		if err != nil && test.expected == nil {
			continue
		} else {
			if err == nil && test.expected == nil {
				t.Fatalf("Expected an error but didn't get one")
			} else if err != nil && test.expected != nil {
				t.Fatalf("Expected no error but got: %+v", err)
			}
		}

		actualId := result["id"].(string)
		if *test.expected != actualId {
			t.Fatalf("expected %q but got %q!", *test.expected, actualId)
		}
	}
}
