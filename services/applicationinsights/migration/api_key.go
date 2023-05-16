package migration

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-provider-azurerm/services/applicationinsights/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
)

var _ pluginsdk.StateUpgrade = ApiKeyUpgradeV0ToV1{}

type ApiKeyUpgradeV0ToV1 struct{}

func (ApiKeyUpgradeV0ToV1) Schema() map[string]*pluginsdk.Schema {
	return apiKeySchemaForV0AndV1()
}

func (ApiKeyUpgradeV0ToV1) UpgradeFunc() pluginsdk.StateUpgraderFunc {
	return func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		// old:
		// 	/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/microsoft.insights/components/component1/apikeys/key1
		// new:
		// 	/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Insights/components/component1/apiKeys/key1
		oldIdRaw := rawState["id"].(string)
		id, err := parse.ApiKeyIDInsensitively(oldIdRaw)
		if err != nil {
			return rawState, err
		}

		newId := id.ID()
		log.Printf("[DEBUG] Updating ID from %q to %q", oldIdRaw, newId)
		rawState["id"] = newId

		return rawState, nil
	}
}

func apiKeySchemaForV0AndV1() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"application_insights_id": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"read_permissions": {
			Type:     pluginsdk.TypeSet,
			Optional: true,
			Set:      pluginsdk.HashString,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},

		"write_permissions": {
			Type:     pluginsdk.TypeSet,
			Optional: true,
			Set:      pluginsdk.HashString,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},

		"api_key": {
			Type:      pluginsdk.TypeString,
			Computed:  true,
			Sensitive: true,
		},
	}
}
