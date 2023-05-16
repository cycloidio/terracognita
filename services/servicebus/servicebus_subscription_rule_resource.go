package servicebus

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/servicebus/mgmt/2021-06-01-preview/servicebus"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/servicebus/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/servicebus/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceServiceBusSubscriptionRule() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceServiceBusSubscriptionRuleCreateUpdate,
		Read:   resourceServiceBusSubscriptionRuleRead,
		Update: resourceServiceBusSubscriptionRuleCreateUpdate,
		Delete: resourceServiceBusSubscriptionRuleDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.SubscriptionRuleID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: resourceServicebusSubscriptionRuleSchema(),
	}
}

func resourceServicebusSubscriptionRuleSchema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringLenBetween(1, 50),
		},

		//lintignore: S013
		"subscription_id": {
			Type:             pluginsdk.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateFunc:     validate.SubscriptionID,
			DiffSuppressFunc: suppress.CaseDifference,
		},

		"filter_type": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ValidateFunc: validation.StringInSlice([]string{
				string(servicebus.FilterTypeSQLFilter),
				string(servicebus.FilterTypeCorrelationFilter),
			}, false),
		},

		"action": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},

		"sql_filter": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: validate.SqlFilter,
		},

		"correlation_filter": {
			Type:          pluginsdk.TypeList,
			Optional:      true,
			MaxItems:      1,
			ConflictsWith: []string{"sql_filter"},
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"correlation_id": {
						Type:     pluginsdk.TypeString,
						Optional: true,
						AtLeastOneOf: []string{
							"correlation_filter.0.correlation_id", "correlation_filter.0.message_id", "correlation_filter.0.to",
							"correlation_filter.0.reply_to", "correlation_filter.0.label", "correlation_filter.0.session_id",
							"correlation_filter.0.reply_to_session_id", "correlation_filter.0.content_type", "correlation_filter.0.properties",
						},
					},
					"message_id": {
						Type:     pluginsdk.TypeString,
						Optional: true,
						AtLeastOneOf: []string{
							"correlation_filter.0.correlation_id", "correlation_filter.0.message_id", "correlation_filter.0.to",
							"correlation_filter.0.reply_to", "correlation_filter.0.label", "correlation_filter.0.session_id",
							"correlation_filter.0.reply_to_session_id", "correlation_filter.0.content_type", "correlation_filter.0.properties",
						},
					},
					"to": {
						Type:     pluginsdk.TypeString,
						Optional: true,
						AtLeastOneOf: []string{
							"correlation_filter.0.correlation_id", "correlation_filter.0.message_id", "correlation_filter.0.to",
							"correlation_filter.0.reply_to", "correlation_filter.0.label", "correlation_filter.0.session_id",
							"correlation_filter.0.reply_to_session_id", "correlation_filter.0.content_type", "correlation_filter.0.properties",
						},
					},
					"reply_to": {
						Type:     pluginsdk.TypeString,
						Optional: true,
						AtLeastOneOf: []string{
							"correlation_filter.0.correlation_id", "correlation_filter.0.message_id", "correlation_filter.0.to",
							"correlation_filter.0.reply_to", "correlation_filter.0.label", "correlation_filter.0.session_id",
							"correlation_filter.0.reply_to_session_id", "correlation_filter.0.content_type", "correlation_filter.0.properties",
						},
					},
					"label": {
						Type:     pluginsdk.TypeString,
						Optional: true,
						AtLeastOneOf: []string{
							"correlation_filter.0.correlation_id", "correlation_filter.0.message_id", "correlation_filter.0.to",
							"correlation_filter.0.reply_to", "correlation_filter.0.label", "correlation_filter.0.session_id",
							"correlation_filter.0.reply_to_session_id", "correlation_filter.0.content_type", "correlation_filter.0.properties",
						},
					},
					"session_id": {
						Type:     pluginsdk.TypeString,
						Optional: true,
						AtLeastOneOf: []string{
							"correlation_filter.0.correlation_id", "correlation_filter.0.message_id", "correlation_filter.0.to",
							"correlation_filter.0.reply_to", "correlation_filter.0.label", "correlation_filter.0.session_id",
							"correlation_filter.0.reply_to_session_id", "correlation_filter.0.content_type", "correlation_filter.0.properties",
						},
					},
					"reply_to_session_id": {
						Type:     pluginsdk.TypeString,
						Optional: true,
						AtLeastOneOf: []string{
							"correlation_filter.0.correlation_id", "correlation_filter.0.message_id", "correlation_filter.0.to",
							"correlation_filter.0.reply_to", "correlation_filter.0.label", "correlation_filter.0.session_id",
							"correlation_filter.0.reply_to_session_id", "correlation_filter.0.content_type", "correlation_filter.0.properties",
						},
					},
					"content_type": {
						Type:     pluginsdk.TypeString,
						Optional: true,
						AtLeastOneOf: []string{
							"correlation_filter.0.correlation_id", "correlation_filter.0.message_id", "correlation_filter.0.to",
							"correlation_filter.0.reply_to", "correlation_filter.0.label", "correlation_filter.0.session_id",
							"correlation_filter.0.reply_to_session_id", "correlation_filter.0.content_type", "correlation_filter.0.properties",
						},
					},
					"properties": {
						Type:     pluginsdk.TypeMap,
						Optional: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
						AtLeastOneOf: []string{
							"correlation_filter.0.correlation_id", "correlation_filter.0.message_id", "correlation_filter.0.to",
							"correlation_filter.0.reply_to", "correlation_filter.0.label", "correlation_filter.0.session_id",
							"correlation_filter.0.reply_to_session_id", "correlation_filter.0.content_type", "correlation_filter.0.properties",
						},
					},
				},
			},
		},
	}
}

func resourceServiceBusSubscriptionRuleCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.SubscriptionRulesClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	log.Printf("[INFO] preparing arguments for Azure Service Bus Subscription Rule creation.")

	filterType := d.Get("filter_type").(string)

	var resourceId parse.SubscriptionRuleId
	if subscriptionIdLit := d.Get("subscription_id").(string); subscriptionIdLit != "" {
		subscriptionId, _ := parse.SubscriptionID(subscriptionIdLit)
		resourceId = parse.NewSubscriptionRuleID(subscriptionId.SubscriptionId,
			subscriptionId.ResourceGroup,
			subscriptionId.NamespaceName,
			subscriptionId.TopicName,
			subscriptionId.Name,
			d.Get("name").(string),
		)
	}

	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceId.ResourceGroup, resourceId.NamespaceName, resourceId.TopicName, resourceId.SubscriptionName, resourceId.RuleName)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", resourceId, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_servicebus_subscription_rule", resourceId.ID())
		}
	}

	rule := servicebus.Rule{
		Ruleproperties: &servicebus.Ruleproperties{
			FilterType: servicebus.FilterType(filterType),
		},
	}

	if action := d.Get("action").(string); action != "" {
		rule.Ruleproperties.Action = &servicebus.Action{
			SQLExpression: &action,
		}
	}

	if rule.Ruleproperties.FilterType == servicebus.FilterTypeCorrelationFilter {
		correlationFilter, err := expandAzureRmServiceBusCorrelationFilter(d)
		if err != nil {
			return fmt.Errorf("expanding `correlation_filter`: %+v", err)
		}

		rule.Ruleproperties.CorrelationFilter = correlationFilter
	}

	if rule.Ruleproperties.FilterType == servicebus.FilterTypeSQLFilter {
		sqlFilter := d.Get("sql_filter").(string)
		rule.Ruleproperties.SQLFilter = &servicebus.SQLFilter{
			SQLExpression: &sqlFilter,
		}
	}

	if _, err := client.CreateOrUpdate(ctx, resourceId.ResourceGroup, resourceId.NamespaceName, resourceId.TopicName, resourceId.SubscriptionName, resourceId.RuleName, rule); err != nil {
		return fmt.Errorf("creating/updating %s: %+v", resourceId, err)
	}

	d.SetId(resourceId.ID())
	return resourceServiceBusSubscriptionRuleRead(d, meta)
}

func resourceServiceBusSubscriptionRuleRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.SubscriptionRulesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SubscriptionRuleID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.NamespaceName, id.TopicName, id.SubscriptionName, id.RuleName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.Set("subscription_id", parse.NewSubscriptionID(id.SubscriptionId, id.ResourceGroup, id.NamespaceName, id.TopicName, id.SubscriptionName).ID())
	d.Set("name", id.RuleName)

	if properties := resp.Ruleproperties; properties != nil {
		d.Set("filter_type", properties.FilterType)

		if properties.Action != nil {
			d.Set("action", properties.Action.SQLExpression)
		}

		if properties.SQLFilter != nil {
			d.Set("sql_filter", properties.SQLFilter.SQLExpression)
		}

		if err := d.Set("correlation_filter", flattenAzureRmServiceBusCorrelationFilter(properties.CorrelationFilter)); err != nil {
			return fmt.Errorf("setting `correlation_filter`: %+v", err)
		}
	}

	return nil
}

func resourceServiceBusSubscriptionRuleDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.SubscriptionRulesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SubscriptionRuleID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Delete(ctx, id.ResourceGroup, id.NamespaceName, id.TopicName, id.SubscriptionName, id.RuleName)
	if err != nil {
		if !response.WasNotFound(resp.Response) {
			return fmt.Errorf("deleting %s: %+v", id, err)
		}
	}

	return nil
}

func expandAzureRmServiceBusCorrelationFilter(d *pluginsdk.ResourceData) (*servicebus.CorrelationFilter, error) {
	configs := d.Get("correlation_filter").([]interface{})
	if len(configs) == 0 {
		return nil, fmt.Errorf("`correlation_filter` is required when `filter_type` is set to `CorrelationFilter`")
	}

	config := configs[0].(map[string]interface{})

	contentType := config["content_type"].(string)
	correlationID := config["correlation_id"].(string)
	label := config["label"].(string)
	messageID := config["message_id"].(string)
	replyTo := config["reply_to"].(string)
	replyToSessionID := config["reply_to_session_id"].(string)
	sessionID := config["session_id"].(string)
	to := config["to"].(string)
	properties := utils.ExpandMapStringPtrString(config["properties"].(map[string]interface{}))

	if contentType == "" && correlationID == "" && label == "" && messageID == "" && replyTo == "" && replyToSessionID == "" && sessionID == "" && to == "" && len(properties) == 0 {
		return nil, fmt.Errorf("At least one property must be set in the `correlation_filter` block")
	}

	correlationFilter := servicebus.CorrelationFilter{}

	if correlationID != "" {
		correlationFilter.CorrelationID = utils.String(correlationID)
	}

	if messageID != "" {
		correlationFilter.MessageID = utils.String(messageID)
	}

	if to != "" {
		correlationFilter.To = utils.String(to)
	}

	if replyTo != "" {
		correlationFilter.ReplyTo = utils.String(replyTo)
	}

	if label != "" {
		correlationFilter.Label = utils.String(label)
	}

	if sessionID != "" {
		correlationFilter.SessionID = utils.String(sessionID)
	}

	if replyToSessionID != "" {
		correlationFilter.ReplyToSessionID = utils.String(replyToSessionID)
	}

	if contentType != "" {
		correlationFilter.ContentType = utils.String(contentType)
	}

	if len(properties) > 0 {
		correlationFilter.Properties = properties
	}

	return &correlationFilter, nil
}

func flattenAzureRmServiceBusCorrelationFilter(input *servicebus.CorrelationFilter) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	filter := make(map[string]interface{})

	if input.CorrelationID != nil {
		filter["correlation_id"] = *input.CorrelationID
	}

	if input.MessageID != nil {
		filter["message_id"] = *input.MessageID
	}

	if input.To != nil {
		filter["to"] = *input.To
	}

	if input.ReplyTo != nil {
		filter["reply_to"] = *input.ReplyTo
	}

	if input.Label != nil {
		filter["label"] = *input.Label
	}

	if input.SessionID != nil {
		filter["session_id"] = *input.SessionID
	}

	if input.ReplyToSessionID != nil {
		filter["reply_to_session_id"] = *input.ReplyToSessionID
	}

	if input.ContentType != nil {
		filter["content_type"] = *input.ContentType
	}

	if input.Properties != nil {
		filter["properties"] = utils.FlattenMapStringPtrString(input.Properties)
	}

	return []interface{}{filter}
}
