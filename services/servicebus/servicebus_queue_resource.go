package servicebus

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/servicebus/mgmt/2021-06-01-preview/servicebus"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/validate"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/servicebus/parse"
	azValidate "github.com/hashicorp/terraform-provider-azurerm/services/servicebus/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceServiceBusQueue() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceServiceBusQueueCreateUpdate,
		Read:   resourceServiceBusQueueRead,
		Update: resourceServiceBusQueueCreateUpdate,
		Delete: resourceServiceBusQueueDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.QueueID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: resourceServicebusQueueSchema(),
	}
}

func resourceServicebusQueueSchema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: azValidate.QueueName(),
		},

		//lintignore: S013
		"namespace_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: azValidate.NamespaceID,
		},

		// Optional
		"auto_delete_on_idle": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validate.ISO8601Duration,
		},

		"dead_lettering_on_message_expiration": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Default:  false,
		},

		"default_message_ttl": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validate.ISO8601Duration,
		},

		"duplicate_detection_history_time_window": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validate.ISO8601Duration,
		},

		// TODO 4.0: change this from enable_* to *_enabled
		"enable_batched_operations": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Default:  true,
		},

		// TODO 4.0: change this from enable_* to *_enabled
		"enable_express": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Default:  false,
		},

		// TODO 4.0: change this from enable_* to *_enabled
		"enable_partitioning": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Default:  false,
			ForceNew: true,
		},

		"forward_dead_lettered_messages_to": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: azValidate.QueueName(),
		},

		"forward_to": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: azValidate.QueueName(),
		},

		"lock_duration": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			Computed: true,
		},

		"max_delivery_count": {
			Type:         pluginsdk.TypeInt,
			Optional:     true,
			Default:      10,
			ValidateFunc: validation.IntAtLeast(1),
		},

		"max_message_size_in_kilobytes": {
			Type:         pluginsdk.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: azValidate.ServiceBusMaxMessageSizeInKilobytes(),
		},

		"max_size_in_megabytes": {
			Type:         pluginsdk.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: azValidate.ServiceBusMaxSizeInMegabytes(),
		},

		"requires_duplicate_detection": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Default:  false,
			ForceNew: true,
		},

		"requires_session": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Default:  false,
			ForceNew: true,
		},

		"status": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			Default:  string(servicebus.EntityStatusActive),
			ValidateFunc: validation.StringInSlice([]string{
				string(servicebus.EntityStatusActive),
				string(servicebus.EntityStatusCreating),
				string(servicebus.EntityStatusDeleting),
				string(servicebus.EntityStatusDisabled),
				string(servicebus.EntityStatusReceiveDisabled),
				string(servicebus.EntityStatusRenaming),
				string(servicebus.EntityStatusSendDisabled),
				string(servicebus.EntityStatusUnknown),
			}, false),
		},
	}
}

func resourceServiceBusQueueCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.QueuesClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	log.Printf("[INFO] preparing arguments for ServiceBus Queue creation/update.")

	var resourceId parse.QueueId
	if namespaceIdLit := d.Get("namespace_id").(string); namespaceIdLit != "" {
		namespaceId, _ := parse.NamespaceID(namespaceIdLit)
		resourceId = parse.NewQueueID(namespaceId.SubscriptionId, namespaceId.ResourceGroup, namespaceId.Name, d.Get("name").(string))
	}

	deadLetteringOnMessageExpiration := d.Get("dead_lettering_on_message_expiration").(bool)
	enableBatchedOperations := d.Get("enable_batched_operations").(bool)
	enableExpress := d.Get("enable_express").(bool)
	enablePartitioning := d.Get("enable_partitioning").(bool)
	maxDeliveryCount := int32(d.Get("max_delivery_count").(int))
	maxSizeInMegabytes := int32(d.Get("max_size_in_megabytes").(int))
	requiresDuplicateDetection := d.Get("requires_duplicate_detection").(bool)
	requiresSession := d.Get("requires_session").(bool)
	status := servicebus.EntityStatus(d.Get("status").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceId.ResourceGroup, resourceId.NamespaceName, resourceId.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of %s: %+v", resourceId, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_servicebus_queue", resourceId.ID())
		}
	}

	parameters := servicebus.SBQueue{
		Name: utils.String(resourceId.Name),
		SBQueueProperties: &servicebus.SBQueueProperties{
			DeadLetteringOnMessageExpiration: &deadLetteringOnMessageExpiration,
			EnableBatchedOperations:          &enableBatchedOperations,
			EnableExpress:                    &enableExpress,
			EnablePartitioning:               &enablePartitioning,
			MaxDeliveryCount:                 &maxDeliveryCount,
			MaxSizeInMegabytes:               &maxSizeInMegabytes,
			RequiresDuplicateDetection:       &requiresDuplicateDetection,
			RequiresSession:                  &requiresSession,
			Status:                           status,
		},
	}

	if autoDeleteOnIdle := d.Get("auto_delete_on_idle").(string); autoDeleteOnIdle != "" {
		parameters.SBQueueProperties.AutoDeleteOnIdle = &autoDeleteOnIdle
	}

	if defaultMessageTTL := d.Get("default_message_ttl").(string); defaultMessageTTL != "" {
		parameters.SBQueueProperties.DefaultMessageTimeToLive = &defaultMessageTTL
	}

	if duplicateDetectionHistoryTimeWindow := d.Get("duplicate_detection_history_time_window").(string); duplicateDetectionHistoryTimeWindow != "" {
		parameters.SBQueueProperties.DuplicateDetectionHistoryTimeWindow = &duplicateDetectionHistoryTimeWindow
	}

	if forwardDeadLetteredMessagesTo := d.Get("forward_dead_lettered_messages_to").(string); forwardDeadLetteredMessagesTo != "" {
		parameters.SBQueueProperties.ForwardDeadLetteredMessagesTo = &forwardDeadLetteredMessagesTo
	}

	if forwardTo := d.Get("forward_to").(string); forwardTo != "" {
		parameters.SBQueueProperties.ForwardTo = &forwardTo
	}

	if lockDuration := d.Get("lock_duration").(string); lockDuration != "" {
		parameters.SBQueueProperties.LockDuration = &lockDuration
	}

	// We need to retrieve the namespace because Premium namespace works differently from Basic and Standard,
	// so it needs different rules applied to it.
	namespacesClient := meta.(*clients.Client).ServiceBus.NamespacesClient
	namespace, err := namespacesClient.Get(ctx, resourceId.ResourceGroup, resourceId.NamespaceName)
	if err != nil {
		return fmt.Errorf("retrieving ServiceBus Namespace %q (Resource Group %q): %+v", resourceId.NamespaceName, resourceId.ResourceGroup, err)
	}

	// Enforce Premium namespace to have Express Entities disabled in Terraform since they are not supported for
	// Premium SKU.
	if namespace.Sku.Name == servicebus.SkuNamePremium && d.Get("enable_express").(bool) {
		return fmt.Errorf("ServiceBus Queue %q does not support Express Entities in Premium SKU and must be disabled", resourceId.Name)
	}

	// output of `max_message_size_in_kilobytes` is also set in non-Premium namespaces, with a value of 256
	if v, ok := d.GetOk("max_message_size_in_kilobytes"); ok && v.(int) != 256 {
		if namespace.Sku.Name != servicebus.SkuNamePremium {
			return fmt.Errorf("ServiceBus Queue %q does not support input on `max_message_size_in_kilobytes` in %s SKU and should be removed", resourceId.Name, namespace.Sku.Name)
		}
		parameters.SBQueueProperties.MaxMessageSizeInKilobytes = utils.Int64(int64(v.(int)))
	}

	if _, err = client.CreateOrUpdate(ctx, resourceId.ResourceGroup, resourceId.NamespaceName, resourceId.Name, parameters); err != nil {
		return err
	}

	d.SetId(resourceId.ID())
	return resourceServiceBusQueueRead(d, meta)
}

func resourceServiceBusQueueRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.QueuesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.QueueID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.NamespaceName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.Set("name", id.Name)
	d.Set("namespace_id", parse.NewNamespaceID(id.SubscriptionId, id.ResourceGroup, id.NamespaceName).ID())

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
		d.Set("max_message_size_in_kilobytes", props.MaxMessageSizeInKilobytes)
		d.Set("requires_duplicate_detection", props.RequiresDuplicateDetection)
		d.Set("requires_session", props.RequiresSession)
		d.Set("status", props.Status)

		if apiMaxSizeInMegabytes := props.MaxSizeInMegabytes; apiMaxSizeInMegabytes != nil {
			maxSizeInMegabytes := int(*apiMaxSizeInMegabytes)

			// If the queue is NOT in a premium namespace (ie. it is Basic or Standard) and partitioning is enabled
			// then the max size returned by the API will be 16 times greater than the value set.
			if *props.EnablePartitioning {
				namespacesClient := meta.(*clients.Client).ServiceBus.NamespacesClient
				namespace, err := namespacesClient.Get(ctx, id.ResourceGroup, id.NamespaceName)
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

func resourceServiceBusQueueDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.QueuesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.QueueID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Delete(ctx, id.ResourceGroup, id.NamespaceName, id.Name)
	if err != nil {
		if !utils.ResponseWasNotFound(resp) {
			return fmt.Errorf("deleting %s: %+v", id, err)
		}
	}

	return nil
}
