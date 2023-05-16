package subscription

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2021-01-01/subscriptions"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/subscription/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
)

func importSubscriptionByAlias() pluginsdk.ImporterFunc {
	return func(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) (data []*pluginsdk.ResourceData, err error) {
		aliasClient := meta.(*clients.Client).Subscription.AliasClient
		client := meta.(*clients.Client).Subscription.Client
		aliasId, err := parse.SubscriptionAliasID(d.Id())
		if err != nil {
			return []*pluginsdk.ResourceData{}, fmt.Errorf("failed parsing Subscription Alias ID for import")
		}
		alias, err := aliasClient.Get(ctx, aliasId.Name)
		if err != nil {
			return []*pluginsdk.ResourceData{}, fmt.Errorf("failed reading Subscription Alias: %+v", err)
		}
		if alias.Properties == nil || alias.Properties.SubscriptionID == nil {
			return []*pluginsdk.ResourceData{}, fmt.Errorf("failed reading Subscription Alias Properties, empty response or missing Subscription ID")
		}
		subscription, err := client.Get(ctx, *alias.Properties.SubscriptionID)
		if err != nil {
			return []*pluginsdk.ResourceData{}, fmt.Errorf("failed parsing Subscription details for import: %+v", err)
		}
		if subscription.State != subscriptions.StateEnabled {
			return []*pluginsdk.ResourceData{}, fmt.Errorf("cannot import a cancelled Subscription by Alias ID, please enable the subscription prior to import")
		}
		return []*pluginsdk.ResourceData{d}, nil
	}
}
