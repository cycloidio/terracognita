package loadbalancer

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/locks"
	"github.com/hashicorp/terraform-provider-azurerm/services/loadbalancer/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/loadbalancer/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceArmLoadBalancerOutboundRule() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceArmLoadBalancerOutboundRuleCreateUpdate,
		Read:   resourceArmLoadBalancerOutboundRuleRead,
		Update: resourceArmLoadBalancerOutboundRuleCreateUpdate,
		Delete: resourceArmLoadBalancerOutboundRuleDelete,

		Importer: loadBalancerSubResourceImporter(func(input string) (*parse.LoadBalancerId, error) {
			id, err := parse.LoadBalancerOutboundRuleID(input)
			if err != nil {
				return nil, err
			}

			lbId := parse.NewLoadBalancerID(id.SubscriptionId, id.ResourceGroup, id.LoadBalancerName)
			return &lbId, nil
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
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"loadbalancer_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.LoadBalancerID,
			},

			"frontend_ip_configuration": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},

						"id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"backend_address_pool_id": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"protocol": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.TransportProtocolAll),
					string(network.TransportProtocolTCP),
					string(network.TransportProtocolUDP),
				}, false),
			},

			// TODO 4.0: change this from enable_* to *_enabled
			"enable_tcp_reset": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			"allocated_outbound_ports": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Default:      1024,
				ValidateFunc: validation.IntAtLeast(0),
			},

			"idle_timeout_in_minutes": {
				Type:     pluginsdk.TypeInt,
				Optional: true,
				Default:  4,
			},
		},
	}
}

func resourceArmLoadBalancerOutboundRuleCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LoadBalancers.LoadBalancersClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	loadBalancerId, err := parse.LoadBalancerID(d.Get("loadbalancer_id").(string))
	if err != nil {
		return err
	}
	loadBalancerIDRaw := loadBalancerId.ID()
	id := parse.NewLoadBalancerOutboundRuleID(subscriptionId, loadBalancerId.ResourceGroup, loadBalancerId.Name, d.Get("name").(string))
	locks.ByID(loadBalancerIDRaw)
	defer locks.UnlockByID(loadBalancerIDRaw)

	loadBalancer, err := client.Get(ctx, loadBalancerId.ResourceGroup, loadBalancerId.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(loadBalancer.Response) {
			d.SetId("")
			log.Printf("[INFO] Load Balancer %q not found. Removing from state", id.LoadBalancerName)
			return nil
		}
		return fmt.Errorf("failed to retrieve Load Balancer %q (resource group %q) for Outbound Rule %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.OutboundRuleName, err)
	}

	newOutboundRule, err := expandAzureRmLoadBalancerOutboundRule(d, &loadBalancer)
	if err != nil {
		return fmt.Errorf("expanding Load Balancer Outbound Rule: %+v", err)
	}

	outboundRules := make([]network.OutboundRule, 0)

	if loadBalancer.LoadBalancerPropertiesFormat.OutboundRules != nil {
		outboundRules = *loadBalancer.LoadBalancerPropertiesFormat.OutboundRules
	}

	existingOutboundRule, existingOutboundRuleIndex, exists := FindLoadBalancerOutboundRuleByName(&loadBalancer, id.OutboundRuleName)
	if exists {
		if id.OutboundRuleName == *existingOutboundRule.Name {
			if d.IsNewResource() {
				return tf.ImportAsExistsError("azurerm_lb_outbound_rule", *existingOutboundRule.ID)
			}

			// this outbound rule is being updated/reapplied remove old copy from the slice
			outboundRules = append(outboundRules[:existingOutboundRuleIndex], outboundRules[existingOutboundRuleIndex+1:]...)
		}
	}

	outboundRules = append(outboundRules, *newOutboundRule)

	loadBalancer.LoadBalancerPropertiesFormat.OutboundRules = &outboundRules

	future, err := client.CreateOrUpdate(ctx, loadBalancerId.ResourceGroup, loadBalancerId.Name, loadBalancer)
	if err != nil {
		return fmt.Errorf("updating LoadBalancer %q (resource group %q) for Outbound Rule %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.OutboundRuleName, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for update of Load Balancer %q (resource group %q) for Outbound Rule %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.OutboundRuleName, err)
	}

	d.SetId(id.ID())

	return resourceArmLoadBalancerOutboundRuleRead(d, meta)
}

func resourceArmLoadBalancerOutboundRuleRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LoadBalancers.LoadBalancersClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LoadBalancerOutboundRuleID(d.Id())
	if err != nil {
		return err
	}

	loadBalancer, err := client.Get(ctx, id.ResourceGroup, id.LoadBalancerName, "")
	if err != nil {
		if utils.ResponseWasNotFound(loadBalancer.Response) {
			d.SetId("")
			log.Printf("[INFO] Load Balancer %q not found. Removing from state", id.LoadBalancerName)
			return nil
		}
		return fmt.Errorf("failed to retrieve Load Balancer %q (resource group %q) for Outbound Rule %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.OutboundRuleName, err)
	}

	config, _, exists := FindLoadBalancerOutboundRuleByName(&loadBalancer, id.OutboundRuleName)
	if !exists {
		d.SetId("")
		log.Printf("[INFO] Load Balancer Outbound Rule %q not found. Removing from state", id.OutboundRuleName)
		return nil
	}

	d.Set("name", config.Name)

	if props := config.OutboundRulePropertiesFormat; props != nil {
		allocatedOutboundPorts := 0
		if props.AllocatedOutboundPorts != nil {
			allocatedOutboundPorts = int(*props.AllocatedOutboundPorts)
		}
		d.Set("allocated_outbound_ports", allocatedOutboundPorts)

		backendAddressPoolId := ""
		if props.BackendAddressPool != nil && props.BackendAddressPool.ID != nil {
			bapid, err := parse.LoadBalancerBackendAddressPoolID(*props.BackendAddressPool.ID)
			if err != nil {
				return err
			}

			backendAddressPoolId = bapid.ID()
		}
		d.Set("backend_address_pool_id", backendAddressPoolId)
		d.Set("enable_tcp_reset", props.EnableTCPReset)

		frontendIpConfigurations := make([]interface{}, 0)
		if configs := props.FrontendIPConfigurations; configs != nil {
			for _, feConfig := range *configs {
				if feConfig.ID == nil {
					continue
				}
				feid, err := parse.LoadBalancerFrontendIpConfigurationID(*feConfig.ID)
				if err != nil {
					return err
				}

				frontendIpConfigurations = append(frontendIpConfigurations, map[string]interface{}{
					"id":   feid.ID(),
					"name": feid.FrontendIPConfigurationName,
				})
			}
		}
		d.Set("frontend_ip_configuration", frontendIpConfigurations)

		idleTimeoutInMinutes := 0
		if props.IdleTimeoutInMinutes != nil {
			idleTimeoutInMinutes = int(*props.IdleTimeoutInMinutes)
		}
		d.Set("idle_timeout_in_minutes", idleTimeoutInMinutes)
		d.Set("protocol", string(props.Protocol))
	}

	return nil
}

func resourceArmLoadBalancerOutboundRuleDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LoadBalancers.LoadBalancersClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LoadBalancerOutboundRuleID(d.Id())
	if err != nil {
		return err
	}

	loadBalancerId := parse.NewLoadBalancerID(id.SubscriptionId, id.ResourceGroup, id.LoadBalancerName)
	loadBalancerID := loadBalancerId.ID()
	locks.ByID(loadBalancerID)
	defer locks.UnlockByID(loadBalancerID)

	loadBalancer, err := client.Get(ctx, loadBalancerId.ResourceGroup, loadBalancerId.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(loadBalancer.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("failed to retrieve Load Balancer %q (resource group %q) for Outbound Rule %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.OutboundRuleName, err)
	}

	_, index, exists := FindLoadBalancerOutboundRuleByName(&loadBalancer, id.OutboundRuleName)
	if !exists {
		return nil
	}

	outboundRules := *loadBalancer.LoadBalancerPropertiesFormat.OutboundRules
	outboundRules = append(outboundRules[:index], outboundRules[index+1:]...)
	loadBalancer.LoadBalancerPropertiesFormat.OutboundRules = &outboundRules

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.LoadBalancerName, loadBalancer)
	if err != nil {
		return fmt.Errorf("updating Load Balancer %q (Resource Group %q) for Outbound Rule %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.OutboundRuleName, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for update of Load Balancer %q (Resource Group %q) for Outbound Rule %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.OutboundRuleName, err)
	}

	return nil
}

func expandAzureRmLoadBalancerOutboundRule(d *pluginsdk.ResourceData, lb *network.LoadBalancer) (*network.OutboundRule, error) {
	properties := network.OutboundRulePropertiesFormat{
		Protocol:               network.LoadBalancerOutboundRuleProtocol(d.Get("protocol").(string)),
		AllocatedOutboundPorts: utils.Int32(int32(d.Get("allocated_outbound_ports").(int))),
	}

	feConfigs := d.Get("frontend_ip_configuration").([]interface{})
	feConfigSubResources := make([]network.SubResource, 0)

	for _, raw := range feConfigs {
		v := raw.(map[string]interface{})
		rule, exists := FindLoadBalancerFrontEndIpConfigurationByName(lb, v["name"].(string))
		if !exists {
			return nil, fmt.Errorf("[ERROR] Cannot find FrontEnd IP Configuration with the name %s", v["name"])
		}

		feConfigSubResource := network.SubResource{
			ID: rule.ID,
		}

		feConfigSubResources = append(feConfigSubResources, feConfigSubResource)
	}

	properties.FrontendIPConfigurations = &feConfigSubResources

	if v := d.Get("backend_address_pool_id").(string); v != "" {
		properties.BackendAddressPool = &network.SubResource{
			ID: &v,
		}
	}

	if v, ok := d.GetOk("idle_timeout_in_minutes"); ok {
		properties.IdleTimeoutInMinutes = utils.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("enable_tcp_reset"); ok {
		properties.EnableTCPReset = utils.Bool(v.(bool))
	}

	return &network.OutboundRule{
		Name:                         utils.String(d.Get("name").(string)),
		OutboundRulePropertiesFormat: &properties,
	}, nil
}
