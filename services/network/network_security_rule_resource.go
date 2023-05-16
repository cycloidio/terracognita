package network

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceNetworkSecurityRule() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceNetworkSecurityRuleCreateUpdate,
		Read:   resourceNetworkSecurityRuleRead,
		Update: resourceNetworkSecurityRuleCreateUpdate,
		Delete: resourceNetworkSecurityRuleDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.SecurityRuleID(id)
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
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"network_security_group_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 140),
			},

			"protocol": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.SecurityRuleProtocolAsterisk),
					string(network.SecurityRuleProtocolTCP),
					string(network.SecurityRuleProtocolUDP),
					string(network.SecurityRuleProtocolIcmp),
					string(network.SecurityRuleProtocolAh),
					string(network.SecurityRuleProtocolEsp),
				}, false),
			},

			"source_port_range": {
				Type:          pluginsdk.TypeString,
				Optional:      true,
				ConflictsWith: []string{"source_port_ranges"},
			},

			"source_port_ranges": {
				Type:          pluginsdk.TypeSet,
				Optional:      true,
				Elem:          &pluginsdk.Schema{Type: pluginsdk.TypeString},
				Set:           pluginsdk.HashString,
				ConflictsWith: []string{"source_port_range"},
			},

			"destination_port_range": {
				Type:          pluginsdk.TypeString,
				Optional:      true,
				ConflictsWith: []string{"destination_port_ranges"},
			},

			"destination_port_ranges": {
				Type:          pluginsdk.TypeSet,
				Optional:      true,
				Elem:          &pluginsdk.Schema{Type: pluginsdk.TypeString},
				Set:           pluginsdk.HashString,
				ConflictsWith: []string{"destination_port_range"},
			},

			"source_address_prefix": {
				Type:          pluginsdk.TypeString,
				Optional:      true,
				ConflictsWith: []string{"source_address_prefixes"},
			},

			"source_address_prefixes": {
				Type:          pluginsdk.TypeSet,
				Optional:      true,
				Elem:          &pluginsdk.Schema{Type: pluginsdk.TypeString},
				Set:           pluginsdk.HashString,
				ConflictsWith: []string{"source_address_prefix"},
			},

			"destination_address_prefix": {
				Type:          pluginsdk.TypeString,
				Optional:      true,
				ConflictsWith: []string{"destination_address_prefixes"},
			},

			"destination_address_prefixes": {
				Type:          pluginsdk.TypeSet,
				Optional:      true,
				Elem:          &pluginsdk.Schema{Type: pluginsdk.TypeString},
				Set:           pluginsdk.HashString,
				ConflictsWith: []string{"destination_address_prefix"},
			},

			//lintignore:S018
			"source_application_security_group_ids": {
				Type:     pluginsdk.TypeSet,
				MaxItems: 10,
				Optional: true,
				Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
				Set:      pluginsdk.HashString,
			},

			//lintignore:S018
			"destination_application_security_group_ids": {
				Type:     pluginsdk.TypeSet,
				MaxItems: 10,
				Optional: true,
				Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
				Set:      pluginsdk.HashString,
			},

			"access": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.SecurityRuleAccessAllow),
					string(network.SecurityRuleAccessDeny),
				}, false),
			},

			"priority": {
				Type:         pluginsdk.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(100, 4096),
			},

			"direction": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.SecurityRuleDirectionInbound),
					string(network.SecurityRuleDirectionOutbound),
				}, false),
			},
		},
	}
}

func resourceNetworkSecurityRuleCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SecurityRuleClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewSecurityRuleID(subscriptionId, d.Get("resource_group_name").(string), d.Get("network_security_group_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.NetworkSecurityGroupName, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %s", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_network_security_rule", id.ID())
		}
	}

	sourcePortRange := d.Get("source_port_range").(string)
	destinationPortRange := d.Get("destination_port_range").(string)
	sourceAddressPrefix := d.Get("source_address_prefix").(string)
	destinationAddressPrefix := d.Get("destination_address_prefix").(string)
	priority := int32(d.Get("priority").(int))
	access := d.Get("access").(string)
	direction := d.Get("direction").(string)
	protocol := d.Get("protocol").(string)

	rule := network.SecurityRule{
		Name: &id.Name,
		SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
			SourcePortRange:          &sourcePortRange,
			DestinationPortRange:     &destinationPortRange,
			SourceAddressPrefix:      &sourceAddressPrefix,
			DestinationAddressPrefix: &destinationAddressPrefix,
			Priority:                 &priority,
			Access:                   network.SecurityRuleAccess(access),
			Direction:                network.SecurityRuleDirection(direction),
			Protocol:                 network.SecurityRuleProtocol(protocol),
		},
	}

	if v, ok := d.GetOk("description"); ok {
		description := v.(string)
		rule.SecurityRulePropertiesFormat.Description = &description
	}

	if r, ok := d.GetOk("source_port_ranges"); ok {
		var sourcePortRanges []string
		r := r.(*pluginsdk.Set).List()
		for _, v := range r {
			s := v.(string)
			sourcePortRanges = append(sourcePortRanges, s)
		}
		rule.SecurityRulePropertiesFormat.SourcePortRanges = &sourcePortRanges
	}

	if r, ok := d.GetOk("destination_port_ranges"); ok {
		var destinationPortRanges []string
		r := r.(*pluginsdk.Set).List()
		for _, v := range r {
			s := v.(string)
			destinationPortRanges = append(destinationPortRanges, s)
		}
		rule.SecurityRulePropertiesFormat.DestinationPortRanges = &destinationPortRanges
	}

	if r, ok := d.GetOk("source_address_prefixes"); ok {
		var sourceAddressPrefixes []string
		r := r.(*pluginsdk.Set).List()
		for _, v := range r {
			s := v.(string)
			sourceAddressPrefixes = append(sourceAddressPrefixes, s)
		}
		rule.SecurityRulePropertiesFormat.SourceAddressPrefixes = &sourceAddressPrefixes
	}

	if r, ok := d.GetOk("destination_address_prefixes"); ok {
		var destinationAddressPrefixes []string
		r := r.(*pluginsdk.Set).List()
		for _, v := range r {
			s := v.(string)
			destinationAddressPrefixes = append(destinationAddressPrefixes, s)
		}
		rule.SecurityRulePropertiesFormat.DestinationAddressPrefixes = &destinationAddressPrefixes
	}

	if r, ok := d.GetOk("source_application_security_group_ids"); ok {
		var sourceApplicationSecurityGroups []network.ApplicationSecurityGroup
		for _, v := range r.(*pluginsdk.Set).List() {
			sg := network.ApplicationSecurityGroup{
				ID: utils.String(v.(string)),
			}
			sourceApplicationSecurityGroups = append(sourceApplicationSecurityGroups, sg)
		}
		rule.SourceApplicationSecurityGroups = &sourceApplicationSecurityGroups
	}

	if r, ok := d.GetOk("destination_application_security_group_ids"); ok {
		var destinationApplicationSecurityGroups []network.ApplicationSecurityGroup
		for _, v := range r.(*pluginsdk.Set).List() {
			sg := network.ApplicationSecurityGroup{
				ID: utils.String(v.(string)),
			}
			destinationApplicationSecurityGroups = append(destinationApplicationSecurityGroups, sg)
		}
		rule.DestinationApplicationSecurityGroups = &destinationApplicationSecurityGroups
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.NetworkSecurityGroupName, id.Name, rule)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for completion of %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceNetworkSecurityRuleRead(d, meta)
}

func resourceNetworkSecurityRuleRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SecurityRuleClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SecurityRuleID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.NetworkSecurityGroupName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("making Read request on %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("network_security_group_name", id.NetworkSecurityGroupName)

	if props := resp.SecurityRulePropertiesFormat; props != nil {
		d.Set("description", props.Description)
		d.Set("protocol", string(props.Protocol))
		d.Set("destination_address_prefix", props.DestinationAddressPrefix)
		d.Set("destination_address_prefixes", props.DestinationAddressPrefixes)
		d.Set("destination_port_range", props.DestinationPortRange)
		d.Set("destination_port_ranges", props.DestinationPortRanges)
		d.Set("source_address_prefix", props.SourceAddressPrefix)
		d.Set("source_address_prefixes", props.SourceAddressPrefixes)
		d.Set("source_port_range", props.SourcePortRange)
		d.Set("source_port_ranges", props.SourcePortRanges)
		d.Set("access", string(props.Access))
		d.Set("priority", int(*props.Priority))
		d.Set("direction", string(props.Direction))

		if err := d.Set("source_application_security_group_ids", flattenApplicationSecurityGroupIds(props.SourceApplicationSecurityGroups)); err != nil {
			return fmt.Errorf("setting `source_application_security_group_ids`: %+v", err)
		}

		if err := d.Set("destination_application_security_group_ids", flattenApplicationSecurityGroupIds(props.DestinationApplicationSecurityGroups)); err != nil {
			return fmt.Errorf("setting `source_application_security_group_ids`: %+v", err)
		}
	}

	return nil
}

func resourceNetworkSecurityRuleDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SecurityRuleClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SecurityRuleID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.NetworkSecurityGroupName, id.Name)
	if err != nil {
		return fmt.Errorf("Deleting %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the deletion of %s: %+v", *id, err)
	}

	return nil
}

func flattenApplicationSecurityGroupIds(groups *[]network.ApplicationSecurityGroup) []string {
	ids := make([]string, 0)

	if groups != nil {
		for _, v := range *groups {
			ids = append(ids, *v.ID)
		}
	}

	return ids
}
