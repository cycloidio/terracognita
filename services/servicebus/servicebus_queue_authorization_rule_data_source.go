package servicebus

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourcegroups"

	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/servicebus/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/servicebus/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func dataSourceServiceBusQueueAuthorizationRule() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceServiceBusQueueAuthorizationRuleRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.AuthorizationRuleName(),
			},

			"queue_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validate.QueueID,
				AtLeastOneOf: []string{"queue_id", "resource_group_name", "namespace_name", "queue_name"},
			},

			"namespace_name": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validate.NamespaceName,
				AtLeastOneOf: []string{"queue_id", "resource_group_name", "namespace_name", "queue_name"},
			},

			"queue_name": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validate.QueueName(),
				AtLeastOneOf: []string{"queue_id", "resource_group_name", "namespace_name", "queue_name"},
			},

			"resource_group_name": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: resourcegroups.ValidateName,
				AtLeastOneOf: []string{"queue_id", "resource_group_name", "namespace_name", "queue_name"},
			},

			"listen": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"send": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"manage": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"primary_key": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"primary_connection_string": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"secondary_key": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"secondary_connection_string": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"primary_connection_string_alias": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"secondary_connection_string_alias": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceServiceBusQueueAuthorizationRuleRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.QueuesClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	var rgName string
	var nsName string
	var queueName string
	if v, ok := d.Get("queue_id").(string); ok && v != "" {
		queueId, err := parse.QueueID(v)
		if err != nil {
			return fmt.Errorf("parsing topic ID %q: %+v", v, err)
		}
		rgName = queueId.ResourceGroup
		nsName = queueId.NamespaceName
		queueName = queueId.Name
	} else {
		rgName = d.Get("resource_group_name").(string)
		nsName = d.Get("namespace_name").(string)
		queueName = d.Get("queue_name").(string)
	}

	id := parse.NewQueueAuthorizationRuleID(subscriptionId, rgName, nsName, queueName, d.Get("name").(string))
	resp, err := client.GetAuthorizationRule(ctx, id.ResourceGroup, id.NamespaceName, id.QueueName, id.AuthorizationRuleName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("%s was not found", id)
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	keysResp, err := client.ListKeys(ctx, id.ResourceGroup, id.NamespaceName, id.QueueName, id.AuthorizationRuleName)
	if err != nil {
		return fmt.Errorf("listing keys for %s: %+v", id, err)
	}

	d.SetId(id.ID())
	d.Set("name", id.AuthorizationRuleName)
	d.Set("queue_name", id.QueueName)
	d.Set("namespace_name", id.NamespaceName)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("queue_id", parse.NewQueueID(id.SubscriptionId, id.ResourceGroup, id.NamespaceName, id.QueueName).ID())

	if properties := resp.SBAuthorizationRuleProperties; properties != nil {
		listen, send, manage := flattenAuthorizationRuleRights(properties.Rights)
		d.Set("listen", listen)
		d.Set("send", send)
		d.Set("manage", manage)
	}

	d.Set("primary_key", keysResp.PrimaryKey)
	d.Set("primary_connection_string", keysResp.PrimaryConnectionString)
	d.Set("secondary_key", keysResp.SecondaryKey)
	d.Set("secondary_connection_string", keysResp.SecondaryConnectionString)
	d.Set("primary_connection_string_alias", keysResp.AliasPrimaryConnectionString)
	d.Set("secondary_connection_string_alias", keysResp.AliasSecondaryConnectionString)

	return nil
}
