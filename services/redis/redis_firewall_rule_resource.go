package redis

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/redis/mgmt/2021-06-01/redis"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/redis/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/redis/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceRedisFirewallRule() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceRedisFirewallRuleCreateUpdate,
		Read:   resourceRedisFirewallRuleRead,
		Update: resourceRedisFirewallRuleCreateUpdate,
		Delete: resourceRedisFirewallRuleDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.FirewallRuleID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.FirewallRuleName,
			},

			"redis_cache_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"start_ip": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.IsIPAddress,
					validation.StringIsNotEmpty,
				),
			},

			"end_ip": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.IsIPAddress,
					validation.StringIsNotEmpty,
				),
			},
		},
	}
}

func resourceRedisFirewallRuleCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Redis.FirewallRulesClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	log.Printf("[INFO] preparing arguments for AzureRM Redis Firewall Rule creation.")

	startIP := d.Get("start_ip").(string)
	endIP := d.Get("end_ip").(string)

	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	resourceId := parse.NewFirewallRuleID(subscriptionId, d.Get("resource_group_name").(string), d.Get("redis_cache_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceId.ResourceGroup, resourceId.RediName, resourceId.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Firewall Rule %q (Redis Cache %q / Resource Group %q): %+v", resourceId.Name, resourceId.RediName, resourceId.ResourceGroup, err)
			}
		}
		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_redis_firewall_rule", resourceId.ID())
		}
	}

	parameters := redis.FirewallRule{
		FirewallRuleProperties: &redis.FirewallRuleProperties{
			StartIP: utils.String(startIP),
			EndIP:   utils.String(endIP),
		},
	}

	if _, err := client.CreateOrUpdate(ctx, resourceId.ResourceGroup, resourceId.RediName, resourceId.Name, parameters); err != nil {
		return fmt.Errorf("creating Firewall Rule %q (Redis Cache %q / Resource Group %q): %+v", resourceId.Name, resourceId.RediName, resourceId.ResourceGroup, err)
	}

	d.SetId(resourceId.ID())
	return resourceRedisFirewallRuleRead(d, meta)
}

func resourceRedisFirewallRuleRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Redis.FirewallRulesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FirewallRuleID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.RediName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Firewall Rule %q was not found in Redis Cache %q / Resource Group %q - removing from state", id.Name, id.RediName, id.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Firewall Rule %q (Redis Cache %q / Resource Group %q): %+v", id.Name, id.RediName, id.ResourceGroup, err)
	}

	d.Set("name", id.Name)
	d.Set("redis_cache_name", id.RediName)
	d.Set("resource_group_name", id.ResourceGroup)
	if props := resp.FirewallRuleProperties; props != nil {
		d.Set("start_ip", props.StartIP)
		d.Set("end_ip", props.EndIP)
	}

	return nil
}

func resourceRedisFirewallRuleDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Redis.FirewallRulesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FirewallRuleID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Delete(ctx, id.ResourceGroup, id.RediName, id.Name)
	if err != nil {
		if !utils.ResponseWasNotFound(resp) {
			return fmt.Errorf("deleting Firewall Rule %q (Redis Cache %q / Resource Group %q): %+v", id.Name, id.RediName, id.ResourceGroup, err)
		}
	}

	return nil
}
