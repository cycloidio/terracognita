package migration

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-provider-azurerm/services/notificationhub/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
)

var _ pluginsdk.StateUpgrade = NotificationHubNamespaceResourceV0ToV1{}

type NotificationHubNamespaceResourceV0ToV1 struct{}

func (NotificationHubNamespaceResourceV0ToV1) Schema() map[string]*pluginsdk.Schema {
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

		"location": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},

		"sku_name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"enabled": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Default:  true,
		},

		"namespace_type": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"tags": {
			Type:     pluginsdk.TypeMap,
			Optional: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},

		"servicebus_endpoint": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},
	}
}

func (NotificationHubNamespaceResourceV0ToV1) UpgradeFunc() pluginsdk.StateUpgraderFunc {
	return func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		oldIdRaw := rawState["id"].(string)
		oldId, err := parse.NamespaceIDInsensitively(oldIdRaw)
		if err != nil {
			return rawState, fmt.Errorf("parsing ID %q to upgrade: %+v", oldIdRaw, err)
		}

		rawState["id"] = oldId.ID()
		return rawState, nil
	}
}
