package migration

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
)

var _ pluginsdk.StateUpgrade = GremlinDatabaseV0ToV1{}

type GremlinDatabaseV0ToV1 struct{}

func (GremlinDatabaseV0ToV1) Schema() map[string]*pluginsdk.Schema {
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

		"account_name": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},

		"throughput": {
			Type:     pluginsdk.TypeInt,
			Optional: true,
			Computed: true,
		},
	}
}

func (GremlinDatabaseV0ToV1) UpgradeFunc() pluginsdk.StateUpgraderFunc {
	return func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		oldId := rawState["id"].(string)
		newId := strings.Replace(rawState["id"].(string), "apis/gremlin/databases", "gremlinDatabases", 1)

		log.Printf("[DEBUG] Updating ID from %q to %q", oldId, newId)

		rawState["id"] = newId

		return rawState, nil
	}
}
