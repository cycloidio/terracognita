package servicebus

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourcegroups"

	"github.com/Azure/azure-sdk-for-go/services/preview/servicebus/mgmt/2021-06-01-preview/servicebus"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/servicebus/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/servicebus/validate"
	azValidate "github.com/hashicorp/terraform-provider-azurerm/services/servicebus/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func dataSourceServiceBusQueue() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceServiceBusQueueRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: azValidate.QueueName(),
			},

			"namespace_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validate.NamespaceID,
				AtLeastOneOf: []string{"namespace_id", "resource_group_name", "namespace_name"},
			},

			"namespace_name": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validate.NamespaceName,
				AtLeastOneOf: []string{"namespace_id", "resource_group_name", "namespace_name"},
			},

			"resource_group_name": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: resourcegroups.ValidateName,
				AtLeastOneOf: []string{"namespace_id", "resource_group_name", "namespace_name"},
			},

			"auto_delete_on_idle": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"dead_lettering_on_message_expiration": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"default_message_ttl": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"duplicate_detection_history_time_window": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			// TODO 4.0: change this from enable_* to *_enabled
			"enable_batched_operations": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			// TODO 4.0: change this from enable_* to *_enabled
			"enable_express": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			// TODO 4.0: change this from enable_* to *_enabled
			"enable_partitioning": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"forward_dead_lettered_messages_to": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"forward_to": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"lock_duration": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"max_delivery_count": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"max_size_in_megabytes": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"requires_duplicate_detection": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"requires_session": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"status": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceServiceBusQueueRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.QueuesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	defer cancel()

	name := d.Get("name").(string)

	var resourceGroup string
	var namespaceName string
	if v, ok := d.Get("namespace_id").(string); ok && v != "" {
		namespaceId, err := parse.NamespaceID(v)
		if err != nil {
			return fmt.Errorf("parsing topic ID %q: %+v", v, err)
		}
		resourceGroup = namespaceId.ResourceGroup
		namespaceName = namespaceId.Name
	} else {
		resourceGroup = d.Get("resource_group_name").(string)
		namespaceName = d.Get("namespace_name").(string)
	}
	id := parse.NewQueueID(subscriptionId, resourceGroup, namespaceName, name)

	resp, err := client.Get(ctx, resourceGroup, namespaceName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Service Bus Queue %q was not found in Resource Group %q Namespace %q", name, resourceGroup, namespaceName)
		}
		return fmt.Errorf("making Read request on Azure Service Bus Queue %q in Resource Group %q Namespace %q: %v", name, resourceGroup, namespaceName, err)
	}

	d.SetId(id.ID())

	if props := resp.SBQueueProperties; props != nil {
		d.Set("auto_delete_on_idle", props.AutoDeleteOnIdle)
		d.Set("dead_lettering_on_message_expiration", props.DeadLetteringOnMessageExpiration)
		d.Set("default_message_ttl", props.DefaultMessageTimeToLive)
		d.Set("duplicate_detection_history_time_window", props.DuplicateDetectionHistoryTimeWindow)
		d.Set("enable_batched_operations", props.EnableBatchedOperations)
		d.Set("enable_express", props.EnableExpress)
		d.Set("enable_partitioning", props.EnablePartitioning)
		d.Set("forward_dead_lettered_messages_to", props.ForwardDeadLetteredMessagesTo)
		d.Set("forward_to", props.ForwardTo)
		d.Set("lock_duration", props.LockDuration)
		d.Set("max_delivery_count", props.MaxDeliveryCount)
		d.Set("requires_duplicate_detection", props.RequiresDuplicateDetection)
		d.Set("requires_session", props.RequiresSession)
		d.Set("status", props.Status)

		if apiMaxSizeInMegabytes := props.MaxSizeInMegabytes; apiMaxSizeInMegabytes != nil {
			maxSizeInMegabytes := int(*apiMaxSizeInMegabytes)

			// If the queue is NOT in a premium namespace (ie. it is Basic or Standard) and partitioning is enabled
			// then the max size returned by the API will be 16 times greater than the value set.
			if *props.EnablePartitioning {
				namespacesClient := meta.(*clients.Client).ServiceBus.NamespacesClient
				namespace, err := namespacesClient.Get(ctx, resourceGroup, namespaceName)
				if err != nil {
					return err
				}

				if namespace.Sku.Name != servicebus.SkuNamePremium {
					const partitionCount = 16
					maxSizeInMegabytes = int(*apiMaxSizeInMegabytes / partitionCount)
				}
			}

			d.Set("max_size_in_megabytes", maxSizeInMegabytes)
		}
	}

	return nil
}
