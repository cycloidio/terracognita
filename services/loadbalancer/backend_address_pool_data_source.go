package loadbalancer

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/loadbalancer/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/loadbalancer/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func dataSourceArmLoadBalancerBackendAddressPool() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceArmLoadBalancerBackendAddressPoolRead,
		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"loadbalancer_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.LoadBalancerID,
			},

			"backend_address": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},

						"virtual_network_id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},

						"ip_address": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"backend_ip_configurations": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"load_balancing_rules": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"outbound_rules": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},
		},
	}
}

func dataSourceArmLoadBalancerBackendAddressPoolRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LoadBalancers.LoadBalancerBackendAddressPoolsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	loadBalancerId, err := parse.LoadBalancerID(d.Get("loadbalancer_id").(string))
	if err != nil {
		return err
	}
	id := parse.NewLoadBalancerBackendAddressPoolID(loadBalancerId.SubscriptionId, loadBalancerId.ResourceGroup, loadBalancerId.Name, name)

	resp, err := client.Get(ctx, id.ResourceGroup, id.LoadBalancerName, id.BackendAddressPoolName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Load Balancer Backend Address Pool %q was not found", id)
		}
		return fmt.Errorf("failed to retrieve Load Balancer Backend Address Pool %q: %+v", id, err)
	}

	d.SetId(id.ID())

	if props := resp.BackendAddressPoolPropertiesFormat; props != nil {
		if err := d.Set("backend_address", flattenArmLoadBalancerBackendAddresses(props.LoadBalancerBackendAddresses)); err != nil {
			return fmt.Errorf("setting `backend_address`: %v", err)
		}

		var backendIPConfigurations []interface{}
		if beipConfigs := props.BackendIPConfigurations; beipConfigs != nil {
			for _, config := range *beipConfigs {
				ipConfig := make(map[string]interface{})
				if id := config.ID; id != nil {
					ipConfig["id"] = *id
					backendIPConfigurations = append(backendIPConfigurations, ipConfig)
				}
			}
		}
		if err := d.Set("backend_ip_configurations", backendIPConfigurations); err != nil {
			return fmt.Errorf("setting `backend_ip_configurations`: %v", err)
		}

		var loadBalancingRules []string
		if rules := props.LoadBalancingRules; rules != nil {
			for _, rule := range *rules {
				if rule.ID == nil {
					continue
				}
				loadBalancingRules = append(loadBalancingRules, *rule.ID)
			}
		}
		if err := d.Set("load_balancing_rules", loadBalancingRules); err != nil {
			return fmt.Errorf("setting `load_balancing_rules`: %v", err)
		}

		var outboundRules []string
		if rules := props.OutboundRules; rules != nil {
			for _, rule := range *rules {
				if rule.ID == nil {
					continue
				}
				outboundRules = append(outboundRules, *rule.ID)
			}
		}
		if err := d.Set("outbound_rules", outboundRules); err != nil {
			return fmt.Errorf("setting `outbound_rules`: %v", err)
		}
	}

	return nil
}

func flattenArmLoadBalancerBackendAddresses(input *[]network.LoadBalancerBackendAddress) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	output := make([]interface{}, 0)

	for _, e := range *input {
		var name string
		if e.Name != nil {
			name = *e.Name
		}

		var (
			ipAddress string
			vnetId    string
		)
		if prop := e.LoadBalancerBackendAddressPropertiesFormat; prop != nil {
			if prop.IPAddress != nil {
				ipAddress = *prop.IPAddress
			}
			if prop.VirtualNetwork != nil && prop.VirtualNetwork.ID != nil {
				vnetId = *prop.VirtualNetwork.ID
			}
		}

		v := map[string]interface{}{
			"name":               name,
			"virtual_network_id": vnetId,
			"ip_address":         ipAddress,
		}
		output = append(output, v)
	}

	return output
}
