package migration

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
)

var _ pluginsdk.StateUpgrade = SubscriptionPricingV0ToV1{}

type SubscriptionPricingV0ToV1 struct{}

func (SubscriptionPricingV0ToV1) Schema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"tier": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},
	}
}

func (SubscriptionPricingV0ToV1) UpgradeFunc() pluginsdk.StateUpgraderFunc {
	return func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		log.Println("[DEBUG] Migrating ResourceType from v0 to v1 format")
		oldId := rawState["id"].(string)
		newId := strings.Replace(oldId, "/default", "/VirtualMachines", 1)

		log.Printf("[DEBUG] Updating ID from %q to %q", oldId, newId)

		rawState["id"] = newId

		return rawState, nil
	}
}
