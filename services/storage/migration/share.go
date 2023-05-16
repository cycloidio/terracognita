package migration

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/tombuildsstuff/giovanni/storage/2020-08-04/file/shares"
)

var _ pluginsdk.StateUpgrade = ShareV0ToV1{}

type ShareV0ToV1 struct{}

func (ShareV0ToV1) Schema() map[string]*pluginsdk.Schema {
	return shareSchemaForV0AndV1()
}

func (ShareV0ToV1) UpgradeFunc() pluginsdk.StateUpgraderFunc {
	// this should have been applied from pre-0.12 migration system; backporting just in-case
	return func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		shareName := rawState["name"].(string)
		resourceGroup := rawState["resource_group_name"].(string)
		accountName := rawState["storage_account_name"].(string)

		id := rawState["id"].(string)
		newResourceID := fmt.Sprintf("%s/%s/%s", shareName, resourceGroup, accountName)
		log.Printf("[DEBUG] Updating ID from %q to %q", id, newResourceID)

		rawState["id"] = newResourceID
		return rawState, nil
	}
}

var _ pluginsdk.StateUpgrade = ShareV1ToV2{}

type ShareV1ToV2 struct{}

func (s ShareV1ToV2) Schema() map[string]*pluginsdk.Schema {
	return shareSchemaForV0AndV1()
}

func (s ShareV1ToV2) UpgradeFunc() pluginsdk.StateUpgraderFunc {
	return func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		id := rawState["id"].(string)

		// name/resourceGroup/accountName
		parsedId := strings.Split(id, "/")
		if len(parsedId) != 3 {
			return rawState, fmt.Errorf("Expected 3 segments in the ID but got %d", len(parsedId))
		}

		shareName := parsedId[0]
		accountName := parsedId[2]

		environment := meta.(*clients.Client).Account.Environment
		client := shares.NewWithEnvironment(environment)

		newResourceId := client.GetResourceID(accountName, shareName)
		log.Printf("[DEBUG] Updating Resource ID from %q to %q", id, newResourceId)

		rawState["id"] = newResourceId

		return rawState, nil
	}
}

// the schema schema was used for both V0 and V1
func shareSchemaForV0AndV1() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},
		"resource_group_name": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},
		"storage_account_name": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},
		"quota": {
			Type:         pluginsdk.TypeInt,
			Optional:     true,
			Default:      5120,
			ValidateFunc: validation.IntBetween(1, 5120),
		},
		"url": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},
	}
}
