package servicebus

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/servicebus/mgmt/2021-06-01-preview/servicebus"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/servicebus/migration"
	"github.com/hashicorp/terraform-provider-azurerm/services/servicebus/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/servicebus/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceServiceBusSubscription() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceServiceBusSubscriptionCreateUpdate,
		Read:   resourceServiceBusSubscriptionRead,
		Update: resourceServiceBusSubscriptionCreateUpdate,
		Delete: resourceServiceBusSubscriptionDelete,

		SchemaVersion: 1,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.ServiceBusSubscriptionV0ToV1{},
		}),

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.SubscriptionID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: resourceServicebusSubscriptionSchema(),
	}
}

func resourceServicebusSubscriptionSchema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.SubscriptionName(),
		},

		//lintignore: S013
		"topic_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.TopicID,
		},

		"auto_delete_on_idle": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			Computed: true,
		},

		"default_message_ttl": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			Computed: true,
		},

		"lock_duration": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			Computed: true,
		},

		"dead_lettering_on_message_expiration": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
		},

		"dead_lettering_on_filter_evaluation_error": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Default:  true,
		},

		// TODO 4.0: change this from enable_* to *_enabled
		"enable_batched_operations": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
		},

		"max_delivery_count": {
			Type:     pluginsdk.TypeInt,
			Required: true,
		},

		"requires_session": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			// cannot be modified
			ForceNew: true,
		},

		"forward_to": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},

		"forward_dead_lettered_messages_to": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},

		"status": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			Default:  string(servicebus.EntityStatusActive),
			ValidateFunc: validation.StringInSlice([]string{
				string(servicebus.EntityStatusActive),
				string(servicebus.EntityStatusDisabled),
				string(servicebus.EntityStatusReceiveDisabled),
			}, false),
		},
	}
}

func resourceServiceBusSubscriptionCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.SubscriptionsClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	log.Printf("[INFO] preparing arguments for ServiceBus Subscription creation.")

	var resourceId parse.SubscriptionId
	if topicIdLit := d.Get("topic_id").(string); topicIdLit != "" {
		topicId, _ := parse.TopicID(topicIdLit)
		resourceId = parse.NewSubscriptionID(topicId.SubscriptionId, topicId.ResourceGroup, topicId.NamespaceName, topicId.Name, d.Get("name").(string))
	}

	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceId.ResourceGroup, resourceId.NamespaceName, resourceId.TopicName, resourceId.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", resourceId, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_servicebus_subscription", resourceId.ID())
		}
	}

	parameters := servicebus.SBSubscription{
		SBSubscriptionProperties: &servicebus.SBSubscriptionProperties{
			DeadLetteringOnMessageExpiration:          utils.Bool(d.Get("dead_lettering_on_message_expiration").(bool)),
			DeadLetteringOnFilterEvaluationExceptions: utils.Bool(d.Get("dead_lettering_on_filter_evaluation_error").(bool)),
			EnableBatchedOperations:                   utils.Bool(d.Get("enable_batched_operations").(bool)),
			MaxDeliveryCount:                          utils.Int32(int32(d.Get("max_delivery_count").(int))),
			RequiresSession:                           utils.Bool(d.Get("requires_session").(bool)),
			Status:                                    servicebus.EntityStatus(d.Get("status").(string)),
		},
	}

	if autoDeleteOnIdle := d.Get("auto_delete_on_idle").(string); autoDeleteOnIdle != "" {
		parameters.SBSubscriptionProperties.AutoDeleteOnIdle = &autoDeleteOnIdle
	}

	if lockDuration := d.Get("lock_duration").(string); lockDuration != "" {
		parameters.SBSubscriptionProperties.LockDuration = &lockDuration
	}

	if forwardTo := d.Get("forward_to").(string); forwardTo != "" {
		parameters.SBSubscriptionProperties.ForwardTo = &forwardTo
	}

	if forwardDeadLetteredMessagesTo := d.Get("forward_dead_lettered_messages_to").(string); forwardDeadLetteredMessagesTo != "" {
		parameters.SBSubscriptionProperties.ForwardDeadLetteredMessagesTo = &forwardDeadLetteredMessagesTo
	}

	if defaultMessageTtl := d.Get("default_message_ttl").(string); defaultMessageTtl != "" {
		parameters.DefaultMessageTimeToLive = &defaultMessageTtl
	}

	if _, err := client.CreateOrUpdate(ctx, resourceId.ResourceGroup, resourceId.NamespaceName, resourceId.TopicName, resourceId.Name, parameters); err != nil {
		return fmt.Errorf("creating/updating %s: %v", resourceId, err)
	}

	d.SetId(resourceId.ID())
	return resourceServiceBusSubscriptionRead(d, meta)
}

func resourceServiceBusSubscriptionRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.SubscriptionsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SubscriptionID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.NamespaceName, id.TopicName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.Set("name", id.Name)
	d.Set("topic_id", parse.NewTopicID(id.SubscriptionId, id.ResourceGroup, id.NamespaceName, id.TopicName).ID())

	if props := resp.SBSubscriptionProperties; props != nil {
		d.Set("auto_delete_on_idle", props.AutoDeleteOnIdle)
		d.Set("default_message_ttl", props.DefaultMessageTimeToLive)
		d.Set("lock_duration", props.LockDuration)
		d.Set("dead_lettering_on_message_expiration", props.DeadLetteringOnMessageExpiration)
		d.Set("dead_lettering_on_filter_evaluation_error", props.DeadLetteringOnFilterEvaluationExceptions)
		d.Set("enable_batched_operations", props.EnableBatchedOperations)
		d.Set("requires_session", props.RequiresSession)
		d.Set("forward_to", props.ForwardTo)
		d.Set("forward_dead_lettered_messages_to", props.ForwardDeadLetteredMessagesTo)
		d.Set("status", utils.String(string(props.Status)))

		if count := props.MaxDeliveryCount; count != nil {
			d.Set("max_delivery_count", int(*count))
		}
	}

	return nil
}

func resourceServiceBusSubscriptionDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.SubscriptionsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SubscriptionID(d.Id())
	if err != nil {
		return err
	}

	if _, err = client.Delete(ctx, id.ResourceGroup, id.NamespaceName, id.TopicName, id.Name); err != nil {
		return fmt.Errorf("deleting %s: %+v", id, err)
	}

	return nil
}
