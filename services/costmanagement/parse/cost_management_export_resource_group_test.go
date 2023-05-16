package parse

import (
	"testing"
)

func TestCostManagementExportResourceGroupId(t *testing.T) {
	testData := []struct {
		Name     string
		Input    string
		Expected *CostManagementExportResourceGroupId
	}{
		{
			Name:     "Empty",
			Input:    "",
			Expected: nil,
		},
		{
			Name:     "No Resource Groups Segment",
			Input:    "/subscriptions/00000000-0000-0000-0000-000000000000",
			Expected: nil,
		},
		{
			Name:     "No Resource Groups Value",
			Input:    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/",
			Expected: nil,
		},
		{
			Name:     "Resource Group ID",
			Input:    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/foo/",
			Expected: nil,
		},
		{
			Name:     "Missing Cost Management Export Value",
			Input:    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resGroup1/providers/Microsoft.CostManagement/exports/",
			Expected: nil,
		},
		{
			Name:  "Cost Management Export ID",
			Input: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resGroup1/providers/Microsoft.CostManagement/exports/Export1",
			Expected: &CostManagementExportResourceGroupId{
				Name:       "Export1",
				ResourceId: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resGroup1",
			},
		},
		{
			Name:     "Wrong Casing",
			Input:    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resGroup1/providers/Microsoft.CostManagement/Exports/Service1",
			Expected: nil,
		},
	}

	for _, v := range testData {
		t.Logf("[DEBUG] Testing %q", v.Name)

		actual, err := CostManagementExportResourceGroupID(v.Input)
		if err != nil {
			if v.Expected == nil {
				continue
			}

			t.Fatalf("Expected a value but got an error: %s", err)
		}

		if actual.Name != v.Expected.Name {
			t.Fatalf("Expected %q but got %q for Name", v.Expected.Name, actual.Name)
		}

		if actual.ResourceId != v.Expected.ResourceId {
			t.Fatalf("Expected %q but got %q for Resource Group", v.Expected.ResourceId, actual.ResourceId)
		}
	}
}
