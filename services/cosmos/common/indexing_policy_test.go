package common

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2021-10-15/documentdb"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func TestValidateAzureRmCosmosDbIndexingPolicy(t *testing.T) {
	cases := []struct {
		Name        string
		Value       *documentdb.IndexingPolicy
		ExpectError bool
	}{
		{
			Name:        "nil",
			Value:       nil,
			ExpectError: false,
		},
		{
			Name: "no included_path or excluded_path with Consistent indexing_mode",
			Value: &documentdb.IndexingPolicy{
				IndexingMode: documentdb.IndexingModeConsistent,
			},
			ExpectError: false,
		},
		{
			Name: "no included_path or excluded_path with None indexing_mode",
			Value: &documentdb.IndexingPolicy{
				IndexingMode: documentdb.IndexingModeNone,
			},
			ExpectError: false,
		},
		{
			Name: "included_path with /*",
			Value: &documentdb.IndexingPolicy{
				IndexingMode: documentdb.IndexingModeConsistent,
				IncludedPaths: &[]documentdb.IncludedPath{
					{
						Path: utils.String("/*"),
					},
					{
						Path: utils.String("/foo/?"),
					},
				},
			},
			ExpectError: false,
		},
		{
			Name: "excluded_path with /*",
			Value: &documentdb.IndexingPolicy{
				IndexingMode: documentdb.IndexingModeConsistent,
				ExcludedPaths: &[]documentdb.ExcludedPath{
					{
						Path: utils.String("/*"),
					},
					{
						Path: utils.String("/foo/?"),
					},
				},
			},
			ExpectError: false,
		},
		{
			Name: "included_path with /* and excluded_path",
			Value: &documentdb.IndexingPolicy{
				IndexingMode: documentdb.IndexingModeConsistent,
				IncludedPaths: &[]documentdb.IncludedPath{
					{
						Path: utils.String("/*"),
					},
					{
						Path: utils.String("/foo/?"),
					},
				},
				ExcludedPaths: &[]documentdb.ExcludedPath{
					{
						Path: utils.String("/testing/?"),
					},
					{
						Path: utils.String("/bar/?"),
					},
				},
			},
			ExpectError: false,
		},
		{
			Name: "included_path and excluded_path with /*",
			Value: &documentdb.IndexingPolicy{
				IndexingMode: documentdb.IndexingModeConsistent,
				IncludedPaths: &[]documentdb.IncludedPath{
					{
						Path: utils.String("/*"),
					},
					{
						Path: utils.String("/foo/?"),
					},
				},
				ExcludedPaths: &[]documentdb.ExcludedPath{
					{
						Path: utils.String("/*"),
					},
					{
						Path: utils.String("/testing/?"),
					},
					{
						Path: utils.String("/bar/?"),
					},
				},
			},
			ExpectError: true,
		},
		{
			Name: "missing /* from included_path",
			Value: &documentdb.IndexingPolicy{
				IndexingMode: documentdb.IndexingModeConsistent,
				IncludedPaths: &[]documentdb.IncludedPath{
					{
						Path: utils.String("/testing/?"),
					},
					{
						Path: utils.String("/foo/?"),
					},
				},
			},
			ExpectError: true,
		},
		{
			Name: "missing /* with included_path and excluded_path",
			Value: &documentdb.IndexingPolicy{
				IndexingMode: documentdb.IndexingModeConsistent,
				IncludedPaths: &[]documentdb.IncludedPath{
					{
						Path: utils.String("/foo/?"),
					},
					{
						Path: utils.String("/foo/?"),
					},
				},
				ExcludedPaths: &[]documentdb.ExcludedPath{
					{
						Path: utils.String("/bar/?"),
					},
					{
						Path: utils.String("/foo/?"),
					},
				},
			},
			ExpectError: true,
		},
		{
			Name: "indexing_mode None with included_path",
			Value: &documentdb.IndexingPolicy{
				IndexingMode: documentdb.IndexingModeNone,
				IncludedPaths: &[]documentdb.IncludedPath{
					{
						Path: utils.String("/*"),
					},
				},
			},
			ExpectError: true,
		},
		{
			Name: "indexing_mode None with excluded_path",
			Value: &documentdb.IndexingPolicy{
				IndexingMode: documentdb.IndexingModeNone,
				ExcludedPaths: &[]documentdb.ExcludedPath{
					{
						Path: utils.String("/*"),
					},
				},
			},
			ExpectError: true,
		},
		{
			Name: "indexing_mode None with included_path and excluded_path",
			Value: &documentdb.IndexingPolicy{
				IndexingMode: documentdb.IndexingModeNone,
				IncludedPaths: &[]documentdb.IncludedPath{
					{
						Path: utils.String("/*"),
					},
				},
				ExcludedPaths: &[]documentdb.ExcludedPath{
					{
						Path: utils.String("/testing/?"),
					},
				},
			},
			ExpectError: true,
		},
	}

	for _, tc := range cases {
		err := ValidateAzureRmCosmosDbIndexingPolicy(tc.Value)
		if tc.ExpectError && err == nil {
			t.Fatalf("Expected an error but didn't get one for %q", tc.Name)
		}

		if !tc.ExpectError && err != nil {
			t.Fatalf("Expected to get no errors for %q but got error: %+v", tc.Name, err)
		}
	}
}
