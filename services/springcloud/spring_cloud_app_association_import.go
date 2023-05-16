package springcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/springcloud/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
)

const (
	springCloudAppAssociationTypeCosmosDb = "Microsoft.DocumentDB"
	springCloudAppAssociationTypeMysql    = "Microsoft.DBforMySQL"
	springCloudAppAssociationTypeRedis    = "Microsoft.Cache"
)

func importSpringCloudAppAssociation(resourceType string) pluginsdk.ImporterFunc {
	return func(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) (data []*pluginsdk.ResourceData, err error) {
		id, err := parse.SpringCloudAppAssociationID(d.Id())
		if err != nil {
			return []*pluginsdk.ResourceData{}, err
		}

		client := meta.(*clients.Client).AppPlatform.BindingsClient
		resp, err := client.Get(ctx, id.ResourceGroup, id.SpringName, id.AppName, id.BindingName)
		if err != nil {
			return []*pluginsdk.ResourceData{}, fmt.Errorf("retrieving %s: %+v", id, err)
		}

		if resp.Properties == nil || resp.Properties.ResourceType == nil {
			return []*pluginsdk.ResourceData{}, fmt.Errorf("retrieving %s: `properties` or `properties.resourceType` was nil", id)
		}

		if *resp.Properties.ResourceType != resourceType {
			return []*pluginsdk.ResourceData{}, fmt.Errorf(`spring Cloud App Association "type" mismatch, expected "%s", got "%s"`, resourceType, *resp.Properties.ResourceType)
		}

		return []*pluginsdk.ResourceData{d}, nil
	}
}
