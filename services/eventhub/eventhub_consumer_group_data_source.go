package eventhub

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/eventhub/sdk/2017-04-01/consumergroups"
	"github.com/hashicorp/terraform-provider-azurerm/services/eventhub/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
)

func EventHubConsumerGroupDataSource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: EventHubConsumerGroupDataSourceRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.Any(
					validate.ValidateEventHubConsumerName(),
					validation.StringInSlice([]string{"$Default"}, false),
				),
			},

			"namespace_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.ValidateEventHubNamespaceName(),
			},

			"eventhub_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.ValidateEventHubName(),
			},

			"resource_group_name": commonschema.ResourceGroupNameForDataSource(),

			"user_metadata": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	}
}

func EventHubConsumerGroupDataSourceRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Eventhub.ConsumerGroupClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := consumergroups.NewConsumerGroupID(subscriptionId, d.Get("resource_group_name").(string), d.Get("namespace_name").(string), d.Get("eventhub_name").(string), d.Get("name").(string))

	resp, err := client.Get(ctx, id)
	if err != nil {
		if response.WasNotFound(resp.HttpResponse) {
			return fmt.Errorf("%s was not found", id)
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.SetId(id.ID())

	d.Set("name", id.ConsumerGroupName)
	d.Set("eventhub_name", id.EventHubName)
	d.Set("namespace_name", id.NamespaceName)
	d.Set("resource_group_name", id.ResourceGroupName)

	if model := resp.Model; model != nil {
		if props := model.Properties; props != nil {
			d.Set("user_metadata", props.UserMetadata)
		}
	}

	return nil
}
