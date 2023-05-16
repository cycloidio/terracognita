package migration

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
)

func TestBlobV0ToV1(t *testing.T) {
	clouds := []azure.Environment{
		azure.ChinaCloud,
		azure.GermanCloud,
		azure.PublicCloud,
		azure.USGovernmentCloud,
	}

	for _, cloud := range clouds {
		t.Logf("[DEBUG] Testing with Cloud %q", cloud.Name)

		input := map[string]interface{}{
			"id":                     "old-id",
			"name":                   "some-name",
			"storage_container_name": "some-container",
			"storage_account_name":   "some-account",
		}
		meta := &clients.Client{
			Account: &clients.ResourceManagerAccount{
				Environment: cloud,
			},
		}
		expected := map[string]interface{}{
			"id":                     fmt.Sprintf("https://some-account.blob.%s/some-container/some-name", cloud.StorageEndpointSuffix),
			"name":                   "some-name",
			"storage_container_name": "some-container",
			"storage_account_name":   "some-account",
		}

		actual, err := BlobV0ToV1{}.UpgradeFunc()(context.TODO(), input, meta)
		if err != nil {
			t.Fatalf("Expected no error but got: %s", err)
		}

		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("Expected %+v. Got %+v. But expected them to be the same", expected, actual)
		}

		t.Logf("[DEBUG] Ok!")
	}
}
