package servicebus

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/servicebus/mgmt/2021-06-01-preview/servicebus"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/servicebus/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/servicebus/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceServiceBusTopicAuthorizationRule() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceServiceBusTopicAuthorizationRuleCreateUpdate,
		Read:   resourceServiceBusTopicAuthorizationRuleRead,
		Update: resourceServiceBusTopicAuthorizationRuleCreateUpdate,
		Delete: resourceServiceBusTopicAuthorizationRuleDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.TopicAuthorizationRuleID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: authorizationRuleSchemaFrom(resourceServiceBusTopicAuthorizationRuleSchema()),

		CustomizeDiff: pluginsdk.CustomizeDiffShim(authorizationRuleCustomizeDiff),
	}
}

func resourceServiceBusTopicAuthorizationRuleSchema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.AuthorizationRuleName(),
		},

		//lintignore: S013
		"topic_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.TopicID,
		},
	}
}

func resourceServiceBusTopicAuthorizationRuleCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.TopicsClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	log.Printf("[INFO] preparing arguments for AzureRM ServiceBus Topic Authorization Rule creation.")

	var resourceId parse.TopicAuthorizationRuleId
	if topicIdLit := d.Get("topic_id").(string); topicIdLit != "" {
		topicId, _ := parse.TopicID(topicIdLit)
		resourceId = parse.NewTopicAuthorizationRuleID(topicId.SubscriptionId, topicId.ResourceGroup, topicId.NamespaceName, topicId.Name, d.Get("name").(string))
	}

	if d.IsNewResource() {
		existing, err := client.GetAuthorizationRule(ctx, resourceId.ResourceGroup, resourceId.NamespaceName, resourceId.TopicName, resourceId.AuthorizationRuleName)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", resourceId, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_servicebus_topic_authorization_rule", resourceId.ID())
		}
	}

	parameters := servicebus.SBAuthorizationRule{
		Name: utils.String(resourceId.AuthorizationRuleName),
		SBAuthorizationRuleProperties: &servicebus.SBAuthorizationRuleProperties{
			Rights: expandAuthorizationRuleRights(d),
		},
	}

	if _, err := client.CreateOrUpdateAuthorizationRule(ctx, resourceId.ResourceGroup, resourceId.NamespaceName, resourceId.TopicName, resourceId.AuthorizationRuleName, parameters); err != nil {
		return fmt.Errorf("creating/updating %s: %+v", resourceId, err)
	}

	d.SetId(resourceId.ID())

	if err := waitForPairedNamespaceReplication(ctx, meta, resourceId.ResourceGroup, resourceId.NamespaceName, d.Timeout(pluginsdk.TimeoutUpdate)); err != nil {
		return fmt.Errorf("waiting for replication to complete for Service Bus Namespace Disaster Recovery Configs (Namespace %q / Resource Group %q): %s", resourceId.NamespaceName, resourceId.ResourceGroup, err)
	}

	return resourceServiceBusTopicAuthorizationRuleRead(d, meta)
}

func resourceServiceBusTopicAuthorizationRuleRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.TopicsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.TopicAuthorizationRuleID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.GetAuthorizationRule(ctx, id.ResourceGroup, id.NamespaceName, id.TopicName, id.AuthorizationRuleName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.Set("name", id.AuthorizationRuleName)
	d.Set("topic_id", parse.NewTopicID(id.SubscriptionId, id.ResourceGroup, id.NamespaceName, id.TopicName).ID())

	if properties := resp.SBAuthorizationRuleProperties; properties != nil {
		listen, send, manage := flattenAuthorizationRuleRights(properties.Rights)
		d.Set("listen", listen)
		d.Set("send", send)
		d.Set("manage", manage)
	}

	keysResp, err := client.ListKeys(ctx, id.ResourceGroup, id.NamespaceName, id.TopicName, id.AuthorizationRuleName)
	if err != nil {
		return fmt.Errorf("listing keys for %s: %+v", id, err)
	}

	d.Set("primary_key", keysResp.PrimaryKey)
	d.Set("primary_connection_string", keysResp.PrimaryConnectionString)
	d.Set("secondary_key", keysResp.SecondaryKey)
	d.Set("secondary_connection_string", keysResp.SecondaryConnectionString)
	d.Set("primary_connection_string_alias", keysResp.AliasPrimaryConnectionString)
	d.Set("secondary_connection_string_alias", keysResp.AliasSecondaryConnectionString)

	return nil
}

func resourceServiceBusTopicAuthorizationRuleDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.TopicsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.TopicAuthorizationRuleID(d.Id())
	if err != nil {
		return err
	}

	if _, err = client.DeleteAuthorizationRule(ctx, id.ResourceGroup, id.NamespaceName, id.TopicName, id.AuthorizationRuleName); err != nil {
		return fmt.Errorf("deleting %s: %+v", id, err)
	}

	if err := waitForPairedNamespaceReplication(ctx, meta, id.ResourceGroup, id.NamespaceName, d.Timeout(pluginsdk.TimeoutUpdate)); err != nil {
		return fmt.Errorf("waiting for replication to complete for Service Bus Namespace Disaster Recovery Configs (Namespace %q / Resource Group %q): %s", id.NamespaceName, id.ResourceGroup, err)
	}

	return nil
}
