package helpers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-provider-azurerm/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/services/appservice/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/relay/sdk/2017-04-01/namespaces"
)

func GetSendKeyValue(ctx context.Context, metadata sdk.ResourceMetaData, id parse.AppHybridConnectionId, sendKeyName string) (*string, error) {
	relayClient := metadata.Client.Relay.NamespacesClient
	connectionId := namespaces.NewAuthorizationRuleID(id.SubscriptionId, id.ResourceGroup, id.HybridConnectionNamespaceName, sendKeyName)
	keys, err := relayClient.ListKeys(ctx, connectionId)
	if err != nil {
		return nil, fmt.Errorf("listing Send Keys for %s in %s: %+v", connectionId, id, err)
	}
	if err != nil || keys.Model == nil || keys.Model.PrimaryKey == nil {
		return nil, fmt.Errorf("reading Send Key Value for %s in %s", connectionId.AuthorizationRuleName, id)
	}
	return keys.Model.PrimaryKey, nil
}
