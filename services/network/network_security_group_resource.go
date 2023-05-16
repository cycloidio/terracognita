package network

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/locks"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/set"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

var networkSecurityGroupResourceName = "azurerm_network_security_group"

func resourceNetworkSecurityGroup() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceNetworkSecurityGroupCreateUpdate,
		Read:   resourceNetworkSecurityGroupRead,
		Update: resourceNetworkSecurityGroupCreateUpdate,
		Delete: resourceNetworkSecurityGroupDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.NetworkSecurityGroupID(id)
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

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"security_rule": {
				Type:       pluginsdk.TypeSet,
				ConfigMode: pluginsdk.SchemaConfigModeAttr,
				Optional:   true,
				Computed:   true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Required: true,
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
							Type:     pluginsdk.TypeString,
							Optional: true,
						},

						"source_port_ranges": {
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
							Set:      pluginsdk.HashString,
						},

						"destination_port_range": {
							Type:     pluginsdk.TypeString,
							Optional: true,
						},

						"destination_port_ranges": {
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
							Set:      pluginsdk.HashString,
						},

						"source_address_prefix": {
							Type:     pluginsdk.TypeString,
							Optional: true,
						},

						"source_address_prefixes": {
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
							Set:      pluginsdk.HashString,
						},

						"destination_address_prefix": {
							Type:     pluginsdk.TypeString,
							Optional: true,
						},

						"destination_address_prefixes": {
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
							Set:      pluginsdk.HashString,
						},

						"destination_application_security_group_ids": {
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
							Set:      pluginsdk.HashString,
						},

						"source_application_security_group_ids": {
							Type:     pluginsdk.TypeSet,
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
				},
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceNetworkSecurityGroupCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SecurityGroupClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewNetworkSecurityGroupID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %s", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_network_security_group", id.ID())
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	t := d.Get("tags").(map[string]interface{})

	sgRules, sgErr := expandAzureRmSecurityRules(d)
	if sgErr != nil {
		return fmt.Errorf("Building list of Network Security Group Rules: %+v", sgErr)
	}

	locks.ByName(id.Name, networkSecurityGroupResourceName)
	defer locks.UnlockByName(id.Name, networkSecurityGroupResourceName)

	sg := network.SecurityGroup{
		Name:     &id.Name,
		Location: &location,
		SecurityGroupPropertiesFormat: &network.SecurityGroupPropertiesFormat{
			SecurityRules: &sgRules,
		},
		Tags: tags.Expand(t),
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, sg)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the completion of %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceNetworkSecurityGroupRead(d, meta)
}

func resourceNetworkSecurityGroupRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SecurityGroupClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.NetworkSecurityGroupID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("making Read request on Network Security Group %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if props := resp.SecurityGroupPropertiesFormat; props != nil {
		flattenedRules := flattenNetworkSecurityRules(props.SecurityRules)
		if err := d.Set("security_rule", flattenedRules); err != nil {
			return fmt.Errorf("setting `security_rule`: %+v", err)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceNetworkSecurityGroupDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SecurityGroupClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.NetworkSecurityGroupID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting Network Security Group %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("deleting Network Security Group %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	return err
}

func expandAzureRmSecurityRules(d *pluginsdk.ResourceData) ([]network.SecurityRule, error) {
	sgRules := d.Get("security_rule").(*pluginsdk.Set).List()
	rules := make([]network.SecurityRule, 0)

	for _, sgRaw := range sgRules {
		sgRule := sgRaw.(map[string]interface{})

		if err := validateSecurityRule(sgRule); err != nil {
			return nil, err
		}

		name := sgRule["name"].(string)
		source_port_range := sgRule["source_port_range"].(string)
		destination_port_range := sgRule["destination_port_range"].(string)
		source_address_prefix := sgRule["source_address_prefix"].(string)
		destination_address_prefix := sgRule["destination_address_prefix"].(string)
		priority := int32(sgRule["priority"].(int))
		access := sgRule["access"].(string)
		direction := sgRule["direction"].(string)
		protocol := sgRule["protocol"].(string)

		properties := network.SecurityRulePropertiesFormat{
			SourcePortRange:          &source_port_range,
			DestinationPortRange:     &destination_port_range,
			SourceAddressPrefix:      &source_address_prefix,
			DestinationAddressPrefix: &destination_address_prefix,
			Priority:                 &priority,
			Access:                   network.SecurityRuleAccess(access),
			Direction:                network.SecurityRuleDirection(direction),
			Protocol:                 network.SecurityRuleProtocol(protocol),
		}

		if v := sgRule["description"].(string); v != "" {
			properties.Description = &v
		}

		if r, ok := sgRule["source_port_ranges"].(*pluginsdk.Set); ok && r.Len() > 0 {
			var sourcePortRanges []string
			for _, v := range r.List() {
				s := v.(string)
				sourcePortRanges = append(sourcePortRanges, s)
			}
			properties.SourcePortRanges = &sourcePortRanges
		}

		if r, ok := sgRule["destination_port_ranges"].(*pluginsdk.Set); ok && r.Len() > 0 {
			var destinationPortRanges []string
			for _, v := range r.List() {
				s := v.(string)
				destinationPortRanges = append(destinationPortRanges, s)
			}
			properties.DestinationPortRanges = &destinationPortRanges
		}

		if r, ok := sgRule["source_address_prefixes"].(*pluginsdk.Set); ok && r.Len() > 0 {
			var sourceAddressPrefixes []string
			for _, v := range r.List() {
				s := v.(string)
				sourceAddressPrefixes = append(sourceAddressPrefixes, s)
			}
			properties.SourceAddressPrefixes = &sourceAddressPrefixes
		}

		if r, ok := sgRule["destination_address_prefixes"].(*pluginsdk.Set); ok && r.Len() > 0 {
			var destinationAddressPrefixes []string
			for _, v := range r.List() {
				s := v.(string)
				destinationAddressPrefixes = append(destinationAddressPrefixes, s)
			}
			properties.DestinationAddressPrefixes = &destinationAddressPrefixes
		}

		if r, ok := sgRule["source_application_security_group_ids"].(*pluginsdk.Set); ok && r.Len() > 0 {
			var sourceApplicationSecurityGroups []network.ApplicationSecurityGroup
			for _, v := range r.List() {
				sg := network.ApplicationSecurityGroup{
					ID: utils.String(v.(string)),
				}
				sourceApplicationSecurityGroups = append(sourceApplicationSecurityGroups, sg)
			}
			properties.SourceApplicationSecurityGroups = &sourceApplicationSecurityGroups
		}

		if r, ok := sgRule["destination_application_security_group_ids"].(*pluginsdk.Set); ok && r.Len() > 0 {
			var destinationApplicationSecurityGroups []network.ApplicationSecurityGroup
			for _, v := range r.List() {
				sg := network.ApplicationSecurityGroup{
					ID: utils.String(v.(string)),
				}
				destinationApplicationSecurityGroups = append(destinationApplicationSecurityGroups, sg)
			}
			properties.DestinationApplicationSecurityGroups = &destinationApplicationSecurityGroups
		}

		rules = append(rules, network.SecurityRule{
			Name:                         &name,
			SecurityRulePropertiesFormat: &properties,
		})
	}

	return rules, nil
}

func flattenNetworkSecurityRules(rules *[]network.SecurityRule) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	if rules != nil {
		for _, rule := range *rules {
			sgRule := make(map[string]interface{})
			sgRule["name"] = *rule.Name

			if props := rule.SecurityRulePropertiesFormat; props != nil {
				if props.Description != nil {
					sgRule["description"] = *props.Description
				}

				if props.DestinationAddressPrefix != nil {
					sgRule["destination_address_prefix"] = *props.DestinationAddressPrefix
				}
				if props.DestinationAddressPrefixes != nil {
					sgRule["destination_address_prefixes"] = set.FromStringSlice(*props.DestinationAddressPrefixes)
				}
				if props.DestinationPortRange != nil {
					sgRule["destination_port_range"] = *props.DestinationPortRange
				}
				if props.DestinationPortRanges != nil {
					sgRule["destination_port_ranges"] = set.FromStringSlice(*props.DestinationPortRanges)
				}

				destinationApplicationSecurityGroups := make([]string, 0)
				if props.DestinationApplicationSecurityGroups != nil {
					for _, g := range *props.DestinationApplicationSecurityGroups {
						destinationApplicationSecurityGroups = append(destinationApplicationSecurityGroups, *g.ID)
					}
				}
				sgRule["destination_application_security_group_ids"] = set.FromStringSlice(destinationApplicationSecurityGroups)

				if props.SourceAddressPrefix != nil {
					sgRule["source_address_prefix"] = *props.SourceAddressPrefix
				}
				if props.SourceAddressPrefixes != nil {
					sgRule["source_address_prefixes"] = set.FromStringSlice(*props.SourceAddressPrefixes)
				}

				sourceApplicationSecurityGroups := make([]string, 0)
				if props.SourceApplicationSecurityGroups != nil {
					for _, g := range *props.SourceApplicationSecurityGroups {
						sourceApplicationSecurityGroups = append(sourceApplicationSecurityGroups, *g.ID)
					}
				}
				sgRule["source_application_security_group_ids"] = set.FromStringSlice(sourceApplicationSecurityGroups)

				if props.SourcePortRange != nil {
					sgRule["source_port_range"] = *props.SourcePortRange
				}
				if props.SourcePortRanges != nil {
					sgRule["source_port_ranges"] = set.FromStringSlice(*props.SourcePortRanges)
				}

				sgRule["protocol"] = string(props.Protocol)
				sgRule["priority"] = int(*props.Priority)
				sgRule["access"] = string(props.Access)
				sgRule["direction"] = string(props.Direction)
			}

			result = append(result, sgRule)
		}
	}

	return result
}

func validateSecurityRule(sgRule map[string]interface{}) error {
	var err *multierror.Error

	sourcePortRange := sgRule["source_port_range"].(string)
	sourcePortRanges := sgRule["source_port_ranges"].(*pluginsdk.Set)
	destinationPortRange := sgRule["destination_port_range"].(string)
	destinationPortRanges := sgRule["destination_port_ranges"].(*pluginsdk.Set)
	sourceAddressPrefix := sgRule["source_address_prefix"].(string)
	sourceAddressPrefixes := sgRule["source_address_prefixes"].(*pluginsdk.Set)
	destinationAddressPrefix := sgRule["destination_address_prefix"].(string)
	destinationAddressPrefixes := sgRule["destination_address_prefixes"].(*pluginsdk.Set)

	if sourcePortRange != "" && sourcePortRanges.Len() > 0 {
		err = multierror.Append(err, fmt.Errorf(
			"only one of \"source_port_range\" and \"source_port_ranges\" can be used per security rule"))
	}
	if destinationPortRange != "" && destinationPortRanges.Len() > 0 {
		err = multierror.Append(err, fmt.Errorf(
			"only one of \"destination_port_range\" and \"destination_port_ranges\" can be used per security rule"))
	}
	if sourceAddressPrefix != "" && sourceAddressPrefixes.Len() > 0 {
		err = multierror.Append(err, fmt.Errorf(
			"only one of \"source_address_prefix\" and \"source_address_prefixes\" can be used per security rule"))
	}
	if destinationAddressPrefix != "" && destinationAddressPrefixes.Len() > 0 {
		err = multierror.Append(err, fmt.Errorf(
			"only one of \"destination_address_prefix\" and \"destination_address_prefixes\" can be used per security rule"))
	}

	return err.ErrorOrNil()
}
