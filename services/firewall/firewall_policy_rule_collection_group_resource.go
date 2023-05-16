package firewall

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	azValidate "github.com/hashicorp/terraform-provider-azurerm/helpers/validate"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/locks"
	"github.com/hashicorp/terraform-provider-azurerm/services/firewall/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/firewall/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceFirewallPolicyRuleCollectionGroup() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceFirewallPolicyRuleCollectionGroupCreateUpdate,
		Read:   resourceFirewallPolicyRuleCollectionGroupRead,
		Update: resourceFirewallPolicyRuleCollectionGroupCreateUpdate,
		Delete: resourceFirewallPolicyRuleCollectionGroupDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.FirewallPolicyRuleCollectionGroupID(id)
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
				ValidateFunc: validate.FirewallPolicyRuleCollectionGroupName(),
			},

			"firewall_policy_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.FirewallPolicyID,
			},

			"priority": {
				Type:         pluginsdk.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(100, 65000),
			},

			"application_rule_collection": {
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
						"priority": {
							Type:         pluginsdk.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(100, 65000),
						},
						"action": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(network.FirewallPolicyFilterRuleCollectionActionTypeAllow),
								string(network.FirewallPolicyFilterRuleCollectionActionTypeDeny),
							}, false),
						},
						"rule": {
							Type:     pluginsdk.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"name": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"description": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"protocols": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										Elem: &pluginsdk.Resource{
											Schema: map[string]*pluginsdk.Schema{
												"type": {
													Type:     pluginsdk.TypeString,
													Required: true,
													ValidateFunc: validation.StringInSlice([]string{
														string(network.FirewallPolicyRuleApplicationProtocolTypeHTTP),
														string(network.FirewallPolicyRuleApplicationProtocolTypeHTTPS),
													}, false),
												},
												"port": {
													Type:         pluginsdk.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(0, 64000),
												},
											},
										},
									},
									"source_addresses": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
											ValidateFunc: validation.Any(
												validation.IsIPAddress,
												validation.IsCIDR,
												validation.StringInSlice([]string{`*`}, false),
											),
										},
									},
									"source_ip_groups": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type:         pluginsdk.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
									},
									"destination_addresses": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
											ValidateFunc: validation.Any(
												validation.IsIPAddress,
												validation.IsCIDR,
												validation.StringInSlice([]string{`*`}, false),
											),
										},
									},
									"destination_fqdns": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type:         pluginsdk.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
									},
									"destination_urls": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type:         pluginsdk.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
									},
									"destination_fqdn_tags": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type:         pluginsdk.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
									},
									"terminate_tls": {
										Type:     pluginsdk.TypeBool,
										Optional: true,
									},
									"web_categories": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type:         pluginsdk.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
									},
								},
							},
						},
					},
				},
			},

			"network_rule_collection": {
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
						"priority": {
							Type:         pluginsdk.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(100, 65000),
						},
						"action": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(network.FirewallPolicyFilterRuleCollectionActionTypeAllow),
								string(network.FirewallPolicyFilterRuleCollectionActionTypeDeny),
							}, false),
						},
						"rule": {
							Type:     pluginsdk.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"name": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"protocols": {
										Type:     pluginsdk.TypeList,
										Required: true,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
											ValidateFunc: validation.StringInSlice([]string{
												string(network.FirewallPolicyRuleNetworkProtocolAny),
												string(network.FirewallPolicyRuleNetworkProtocolTCP),
												string(network.FirewallPolicyRuleNetworkProtocolUDP),
												string(network.FirewallPolicyRuleNetworkProtocolICMP),
											}, false),
										},
									},
									"source_addresses": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
											ValidateFunc: validation.Any(
												validation.IsIPAddress,
												validation.IsCIDR,
												validation.StringInSlice([]string{`*`}, false),
											),
										},
									},
									"source_ip_groups": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type:         pluginsdk.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
									},
									"destination_addresses": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
											// Can be IP address, CIDR, "*", or service tag
											ValidateFunc: validation.StringIsNotEmpty,
										},
									},
									"destination_ip_groups": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type:         pluginsdk.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
									},
									"destination_fqdns": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type:         pluginsdk.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
									},
									"destination_ports": {
										Type:     pluginsdk.TypeList,
										Required: true,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
											ValidateFunc: validation.Any(
												azValidate.PortOrPortRangeWithin(1, 65535),
												validation.StringInSlice([]string{`*`}, false),
											),
										},
									},
								},
							},
						},
					},
				},
			},

			"nat_rule_collection": {
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
						"priority": {
							Type:         pluginsdk.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(100, 65000),
						},
						"action": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								// Hardcode to using `Dnat` instead of the one defined in Swagger (i.e. network.DNAT) because of: https://github.com/Azure/azure-rest-api-specs/issues/9986
								// Setting `StateFunc: state.IgnoreCase` will cause other issues, as tracked by: https://github.com/hashicorp/terraform-plugin-sdk/issues/485
								"Dnat",
							}, false),
						},
						"rule": {
							Type:     pluginsdk.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"name": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"protocols": {
										Type:     pluginsdk.TypeList,
										Required: true,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
											ValidateFunc: validation.StringInSlice([]string{
												string(network.FirewallPolicyRuleNetworkProtocolTCP),
												string(network.FirewallPolicyRuleNetworkProtocolUDP),
											}, false),
										},
									},
									"source_addresses": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
											ValidateFunc: validation.Any(
												validation.IsIPAddress,
												validation.IsCIDR,
												validation.StringInSlice([]string{`*`}, false),
											),
										},
									},
									"source_ip_groups": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type:         pluginsdk.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
									},
									"destination_address": {
										Type:     pluginsdk.TypeString,
										Optional: true,
										ValidateFunc: validation.Any(
											validation.IsIPAddress,
											validation.IsCIDR,
										),
									},
									"destination_ports": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type:         pluginsdk.TypeString,
											ValidateFunc: azValidate.PortOrPortRangeWithin(1, 64000),
										},
									},
									"translated_address": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										ValidateFunc: validation.IsIPAddress,
									},
									"translated_port": {
										Type:         pluginsdk.TypeInt,
										Required:     true,
										ValidateFunc: validation.IsPortNumber,
									},
									"translated_fqdn": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceFirewallPolicyRuleCollectionGroupCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Firewall.FirewallPolicyRuleGroupClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	policyId, err := parse.FirewallPolicyID(d.Get("firewall_policy_id").(string))
	if err != nil {
		return err
	}

	if d.IsNewResource() {
		resp, err := client.Get(ctx, policyId.ResourceGroup, policyId.Name, name)
		if err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("checking for existing Firewall Policy Rule Collection Group %q (Resource Group %q / Policy %q): %+v", name, policyId.ResourceGroup, policyId.Name, err)
			}
		}

		if resp.ID != nil && *resp.ID != "" {
			return tf.ImportAsExistsError("azurerm_firewall_policy_rule_collection_group", *resp.ID)
		}
	}

	locks.ByName(policyId.Name, azureFirewallPolicyResourceName)
	defer locks.UnlockByName(policyId.Name, azureFirewallPolicyResourceName)

	param := network.FirewallPolicyRuleCollectionGroup{
		FirewallPolicyRuleCollectionGroupProperties: &network.FirewallPolicyRuleCollectionGroupProperties{
			Priority: utils.Int32(int32(d.Get("priority").(int))),
		},
	}
	var rulesCollections []network.BasicFirewallPolicyRuleCollection
	rulesCollections = append(rulesCollections, expandFirewallPolicyRuleCollectionApplication(d.Get("application_rule_collection").([]interface{}))...)
	rulesCollections = append(rulesCollections, expandFirewallPolicyRuleCollectionNetwork(d.Get("network_rule_collection").([]interface{}))...)

	natRules, err := expandFirewallPolicyRuleCollectionNat(d.Get("nat_rule_collection").([]interface{}))
	if err != nil {
		return fmt.Errorf("expanding NAT rule collection: %w", err)
	}
	rulesCollections = append(rulesCollections, natRules...)

	param.FirewallPolicyRuleCollectionGroupProperties.RuleCollections = &rulesCollections

	future, err := client.CreateOrUpdate(ctx, policyId.ResourceGroup, policyId.Name, name, param)
	if err != nil {
		return fmt.Errorf("creating Firewall Policy Rule Collection Group %q (Resource Group %q / Policy: %q): %+v", name, policyId.ResourceGroup, policyId.Name, err)
	}
	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting Firewall Policy Rule Collection Group %q (Resource Group %q / Policy: %q): %+v", name, policyId.ResourceGroup, policyId.Name, err)
	}

	resp, err := client.Get(ctx, policyId.ResourceGroup, policyId.Name, name)
	if err != nil {
		return fmt.Errorf("retrieving Firewall Policy Rule Collection Group %q (Resource Group %q / Policy: %q): %+v", name, policyId.ResourceGroup, policyId.Name, err)
	}
	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("empty or nil ID returned for Firewall Policy Rule Collection Group %q (Resource Group %q / Policy: %q) ID", name, policyId.ResourceGroup, policyId.Name)
	}
	id, err := parse.FirewallPolicyRuleCollectionGroupID(*resp.ID)
	if err != nil {
		return err
	}
	d.SetId(id.ID())

	return resourceFirewallPolicyRuleCollectionGroupRead(d, meta)
}

func resourceFirewallPolicyRuleCollectionGroupRead(d *pluginsdk.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).Firewall.FirewallPolicyRuleGroupClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FirewallPolicyRuleCollectionGroupID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.FirewallPolicyName, id.RuleCollectionGroupName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Firewall Policy Rule Collection Group %q was not found in Resource Group %q - removing from state!", id.RuleCollectionGroupName, id.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Firewall Policy Rule Collection Group %q (Resource Group %q / Policy: %q): %+v", id.RuleCollectionGroupName, id.ResourceGroup, id.FirewallPolicyName, err)
	}

	d.Set("name", resp.Name)
	d.Set("priority", resp.Priority)
	d.Set("firewall_policy_id", parse.NewFirewallPolicyID(subscriptionId, id.ResourceGroup, id.FirewallPolicyName).ID())

	applicationRuleCollections, networkRuleCollections, natRuleCollections, err := flattenFirewallPolicyRuleCollection(resp.RuleCollections)
	if err != nil {
		return fmt.Errorf("flattening Firewall Policy Rule Collections: %+v", err)
	}

	if err := d.Set("application_rule_collection", applicationRuleCollections); err != nil {
		return fmt.Errorf("setting `application_rule_collection`: %+v", err)
	}
	if err := d.Set("network_rule_collection", networkRuleCollections); err != nil {
		return fmt.Errorf("setting `network_rule_collection`: %+v", err)
	}
	if err := d.Set("nat_rule_collection", natRuleCollections); err != nil {
		return fmt.Errorf("setting `nat_rule_collection`: %+v", err)
	}

	return nil
}

func resourceFirewallPolicyRuleCollectionGroupDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Firewall.FirewallPolicyRuleGroupClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FirewallPolicyRuleCollectionGroupID(d.Id())
	if err != nil {
		return err
	}

	locks.ByName(id.FirewallPolicyName, azureFirewallPolicyResourceName)
	defer locks.UnlockByName(id.FirewallPolicyName, azureFirewallPolicyResourceName)

	future, err := client.Delete(ctx, id.ResourceGroup, id.FirewallPolicyName, id.RuleCollectionGroupName)
	if err != nil {
		return fmt.Errorf("deleting Firewall Policy Rule Collection Group %q (Resource Group %q / Policy: %q): %+v", id.RuleCollectionGroupName, id.ResourceGroup, id.FirewallPolicyName, err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("waiting for deleting %q (Resource Group %q / Policy: %q): %+v", id.RuleCollectionGroupName, id.ResourceGroup, id.FirewallPolicyName, err)
		}
	}

	return nil
}

func expandFirewallPolicyRuleCollectionApplication(input []interface{}) []network.BasicFirewallPolicyRuleCollection {
	return expandFirewallPolicyFilterRuleCollection(input, expandFirewallPolicyRuleApplication)
}

func expandFirewallPolicyRuleCollectionNetwork(input []interface{}) []network.BasicFirewallPolicyRuleCollection {
	return expandFirewallPolicyFilterRuleCollection(input, expandFirewallPolicyRuleNetwork)
}

func expandFirewallPolicyRuleCollectionNat(input []interface{}) ([]network.BasicFirewallPolicyRuleCollection, error) {
	result := make([]network.BasicFirewallPolicyRuleCollection, 0)
	for _, e := range input {
		rule := e.(map[string]interface{})
		rules, err := expandFirewallPolicyRuleNat(rule["rule"].([]interface{}))
		if err != nil {
			return nil, err
		}
		output := &network.FirewallPolicyNatRuleCollection{
			RuleCollectionType: network.RuleCollectionTypeFirewallPolicyNatRuleCollection,
			Name:               utils.String(rule["name"].(string)),
			Priority:           utils.Int32(int32(rule["priority"].(int))),
			Action: &network.FirewallPolicyNatRuleCollectionAction{
				Type: network.FirewallPolicyNatRuleCollectionActionType(rule["action"].(string)),
			},
			Rules: rules,
		}
		result = append(result, output)
	}
	return result, nil
}

func expandFirewallPolicyFilterRuleCollection(input []interface{}, f func(input []interface{}) *[]network.BasicFirewallPolicyRule) []network.BasicFirewallPolicyRuleCollection {
	result := make([]network.BasicFirewallPolicyRuleCollection, 0)
	for _, e := range input {
		rule := e.(map[string]interface{})
		output := &network.FirewallPolicyFilterRuleCollection{
			Action: &network.FirewallPolicyFilterRuleCollectionAction{
				Type: network.FirewallPolicyFilterRuleCollectionActionType(rule["action"].(string)),
			},
			Name:               utils.String(rule["name"].(string)),
			Priority:           utils.Int32(int32(rule["priority"].(int))),
			RuleCollectionType: network.RuleCollectionTypeFirewallPolicyFilterRuleCollection,
			Rules:              f(rule["rule"].([]interface{})),
		}
		result = append(result, output)
	}
	return result
}

func expandFirewallPolicyRuleApplication(input []interface{}) *[]network.BasicFirewallPolicyRule {
	result := make([]network.BasicFirewallPolicyRule, 0)
	for _, e := range input {
		condition := e.(map[string]interface{})
		var protocols []network.FirewallPolicyRuleApplicationProtocol
		for _, p := range condition["protocols"].([]interface{}) {
			proto := p.(map[string]interface{})
			protocols = append(protocols, network.FirewallPolicyRuleApplicationProtocol{
				ProtocolType: network.FirewallPolicyRuleApplicationProtocolType(proto["type"].(string)),
				Port:         utils.Int32(int32(proto["port"].(int))),
			})
		}
		output := &network.ApplicationRule{
			Name:                 utils.String(condition["name"].(string)),
			Description:          utils.String(condition["description"].(string)),
			RuleType:             network.RuleTypeApplicationRule,
			Protocols:            &protocols,
			SourceAddresses:      utils.ExpandStringSlice(condition["source_addresses"].([]interface{})),
			SourceIPGroups:       utils.ExpandStringSlice(condition["source_ip_groups"].([]interface{})),
			DestinationAddresses: utils.ExpandStringSlice(condition["destination_addresses"].([]interface{})),
			TargetFqdns:          utils.ExpandStringSlice(condition["destination_fqdns"].([]interface{})),
			TargetUrls:           utils.ExpandStringSlice(condition["destination_urls"].([]interface{})),
			FqdnTags:             utils.ExpandStringSlice(condition["destination_fqdn_tags"].([]interface{})),
			TerminateTLS:         utils.Bool(condition["terminate_tls"].(bool)),
			WebCategories:        utils.ExpandStringSlice(condition["web_categories"].([]interface{})),
		}
		result = append(result, output)
	}
	return &result
}

func expandFirewallPolicyRuleNetwork(input []interface{}) *[]network.BasicFirewallPolicyRule {
	result := make([]network.BasicFirewallPolicyRule, 0)
	for _, e := range input {
		condition := e.(map[string]interface{})
		var protocols []network.FirewallPolicyRuleNetworkProtocol
		for _, p := range condition["protocols"].([]interface{}) {
			protocols = append(protocols, network.FirewallPolicyRuleNetworkProtocol(p.(string)))
		}
		output := &network.Rule{
			Name:                 utils.String(condition["name"].(string)),
			RuleType:             network.RuleTypeNetworkRule,
			IPProtocols:          &protocols,
			SourceAddresses:      utils.ExpandStringSlice(condition["source_addresses"].([]interface{})),
			SourceIPGroups:       utils.ExpandStringSlice(condition["source_ip_groups"].([]interface{})),
			DestinationAddresses: utils.ExpandStringSlice(condition["destination_addresses"].([]interface{})),
			DestinationIPGroups:  utils.ExpandStringSlice(condition["destination_ip_groups"].([]interface{})),
			DestinationFqdns:     utils.ExpandStringSlice(condition["destination_fqdns"].([]interface{})),
			DestinationPorts:     utils.ExpandStringSlice(condition["destination_ports"].([]interface{})),
		}
		result = append(result, output)
	}
	return &result
}

func expandFirewallPolicyRuleNat(input []interface{}) (*[]network.BasicFirewallPolicyRule, error) {
	result := make([]network.BasicFirewallPolicyRule, 0)
	for _, e := range input {
		condition := e.(map[string]interface{})
		var protocols []network.FirewallPolicyRuleNetworkProtocol
		for _, p := range condition["protocols"].([]interface{}) {
			protocols = append(protocols, network.FirewallPolicyRuleNetworkProtocol(p.(string)))
		}
		destinationAddresses := []string{condition["destination_address"].(string)}

		// Exactly one of `translated_address` and `translated_fqdn` should be set.
		if condition["translated_address"].(string) != "" && condition["translated_fqdn"].(string) != "" {
			return nil, fmt.Errorf("can't specify both `translated_address` and `translated_fqdn` in rule %s", condition["name"].(string))
		}
		if condition["translated_address"].(string) == "" && condition["translated_fqdn"].(string) == "" {
			return nil, fmt.Errorf("should specify either `translated_address` or `translated_fqdn` in rule %s", condition["name"].(string))
		}
		output := &network.NatRule{
			Name:                 utils.String(condition["name"].(string)),
			RuleType:             network.RuleTypeNatRule,
			IPProtocols:          &protocols,
			SourceAddresses:      utils.ExpandStringSlice(condition["source_addresses"].([]interface{})),
			SourceIPGroups:       utils.ExpandStringSlice(condition["source_ip_groups"].([]interface{})),
			DestinationAddresses: &destinationAddresses,
			DestinationPorts:     utils.ExpandStringSlice(condition["destination_ports"].([]interface{})),
			TranslatedPort:       utils.String(strconv.Itoa(condition["translated_port"].(int))),
		}
		if condition["translated_address"].(string) != "" {
			output.TranslatedAddress = utils.String(condition["translated_address"].(string))
		}
		if condition["translated_fqdn"].(string) != "" {
			output.TranslatedFqdn = utils.String(condition["translated_fqdn"].(string))
		}
		result = append(result, output)
	}
	return &result, nil
}

func flattenFirewallPolicyRuleCollection(input *[]network.BasicFirewallPolicyRuleCollection) ([]interface{}, []interface{}, []interface{}, error) {
	var (
		applicationRuleCollection = []interface{}{}
		networkRuleCollection     = []interface{}{}
		natRuleCollection         = []interface{}{}
	)
	if input == nil {
		return applicationRuleCollection, networkRuleCollection, natRuleCollection, nil
	}

	for _, e := range *input {
		var result map[string]interface{}

		switch rule := e.(type) {
		case network.FirewallPolicyFilterRuleCollection:
			var name string
			if rule.Name != nil {
				name = *rule.Name
			}
			var priority int32
			if rule.Priority != nil {
				priority = *rule.Priority
			}

			var action string
			if rule.Action != nil {
				action = string(rule.Action.Type)
			}

			result = map[string]interface{}{
				"name":     name,
				"priority": priority,
				"action":   action,
			}

			if rule.Rules == nil || len(*rule.Rules) == 0 {
				continue
			}

			// Determine the rule type based on the first rule's type
			switch (*rule.Rules)[0].(type) {
			case network.ApplicationRule:
				appRules, err := flattenFirewallPolicyRuleApplication(rule.Rules)
				if err != nil {
					return nil, nil, nil, err
				}
				result["rule"] = appRules

				applicationRuleCollection = append(applicationRuleCollection, result)

			case network.Rule:
				networkRules, err := flattenFirewallPolicyRuleNetwork(rule.Rules)
				if err != nil {
					return nil, nil, nil, err
				}
				result["rule"] = networkRules

				networkRuleCollection = append(networkRuleCollection, result)

			default:
				return nil, nil, nil, fmt.Errorf("unknown rule condition type %+v", (*rule.Rules)[0])
			}
		case network.FirewallPolicyNatRuleCollection:
			var name string
			if rule.Name != nil {
				name = *rule.Name
			}
			var priority int32
			if rule.Priority != nil {
				priority = *rule.Priority
			}

			var action string
			if rule.Action != nil {
				action = string(rule.Action.Type)
			}

			rules, err := flattenFirewallPolicyRuleNat(rule.Rules)
			if err != nil {
				return nil, nil, nil, err
			}
			result = map[string]interface{}{
				"name":     name,
				"priority": priority,
				"action":   action,
				"rule":     rules,
			}

			natRuleCollection = append(natRuleCollection, result)

		default:
			return nil, nil, nil, fmt.Errorf("unknown rule type %+v", rule)
		}
	}
	return applicationRuleCollection, networkRuleCollection, natRuleCollection, nil
}

func flattenFirewallPolicyRuleApplication(input *[]network.BasicFirewallPolicyRule) ([]interface{}, error) {
	if input == nil {
		return []interface{}{}, nil
	}
	output := make([]interface{}, 0)
	for _, e := range *input {
		rule, ok := e.(network.ApplicationRule)
		if !ok {
			return nil, fmt.Errorf("unexpected non-application rule: %+v", e)
		}

		var name string
		if rule.Name != nil {
			name = *rule.Name
		}

		var description string
		if rule.Description != nil {
			description = *rule.Description
		}

		var terminate_tls bool
		if rule.TerminateTLS != nil {
			terminate_tls = *rule.TerminateTLS
		}

		protocols := make([]interface{}, 0)
		if rule.Protocols != nil {
			for _, protocol := range *rule.Protocols {
				var port int
				if protocol.Port != nil {
					port = int(*protocol.Port)
				}
				protocols = append(protocols, map[string]interface{}{
					"type": string(protocol.ProtocolType),
					"port": port,
				})
			}
		}

		output = append(output, map[string]interface{}{
			"name":                  name,
			"description":           description,
			"protocols":             protocols,
			"source_addresses":      utils.FlattenStringSlice(rule.SourceAddresses),
			"source_ip_groups":      utils.FlattenStringSlice(rule.SourceIPGroups),
			"destination_addresses": utils.FlattenStringSlice(rule.DestinationAddresses),
			"destination_urls":      utils.FlattenStringSlice(rule.TargetUrls),
			"destination_fqdns":     utils.FlattenStringSlice(rule.TargetFqdns),
			"destination_fqdn_tags": utils.FlattenStringSlice(rule.FqdnTags),
			"terminate_tls":         terminate_tls,
			"web_categories":        utils.FlattenStringSlice(rule.WebCategories),
		})
	}

	return output, nil
}

func flattenFirewallPolicyRuleNetwork(input *[]network.BasicFirewallPolicyRule) ([]interface{}, error) {
	if input == nil {
		return []interface{}{}, nil
	}
	output := make([]interface{}, 0)
	for _, e := range *input {
		rule, ok := e.(network.Rule)
		if !ok {
			return nil, fmt.Errorf("unexpected non-network rule: %+v", e)
		}

		var name string
		if rule.Name != nil {
			name = *rule.Name
		}

		protocols := make([]interface{}, 0)
		if rule.IPProtocols != nil {
			for _, protocol := range *rule.IPProtocols {
				protocols = append(protocols, string(protocol))
			}
		}

		output = append(output, map[string]interface{}{
			"name":                  name,
			"protocols":             protocols,
			"source_addresses":      utils.FlattenStringSlice(rule.SourceAddresses),
			"source_ip_groups":      utils.FlattenStringSlice(rule.SourceIPGroups),
			"destination_addresses": utils.FlattenStringSlice(rule.DestinationAddresses),
			"destination_ip_groups": utils.FlattenStringSlice(rule.DestinationIPGroups),
			"destination_fqdns":     utils.FlattenStringSlice(rule.DestinationFqdns),
			"destination_ports":     utils.FlattenStringSlice(rule.DestinationPorts),
		})
	}
	return output, nil
}

func flattenFirewallPolicyRuleNat(input *[]network.BasicFirewallPolicyRule) ([]interface{}, error) {
	if input == nil {
		return []interface{}{}, nil
	}
	output := make([]interface{}, 0)
	for _, e := range *input {
		rule, ok := e.(network.NatRule)
		if !ok {
			return nil, fmt.Errorf("unexpected non-nat rule: %+v", e)
		}

		var name string
		if rule.Name != nil {
			name = *rule.Name
		}

		protocols := make([]interface{}, 0)
		if rule.IPProtocols != nil {
			for _, protocol := range *rule.IPProtocols {
				protocols = append(protocols, string(protocol))
			}
		}
		destinationAddr := ""
		if rule.DestinationAddresses != nil && len(*rule.DestinationAddresses) != 0 {
			destinationAddr = (*rule.DestinationAddresses)[0]
		}

		translatedPort := 0
		if rule.TranslatedPort != nil {
			port, err := strconv.Atoi(*rule.TranslatedPort)
			if err != nil {
				return nil, fmt.Errorf(`The "translatedPort" property is not a valid integer (%s)`, *rule.TranslatedPort)
			}
			translatedPort = port
		}

		translatedAddress := ""
		if rule.TranslatedAddress != nil {
			translatedAddress = *rule.TranslatedAddress
		}

		translatedFQDN := ""
		if rule.TranslatedFqdn != nil {
			translatedFQDN = *rule.TranslatedFqdn
		}

		output = append(output, map[string]interface{}{
			"name":                name,
			"protocols":           protocols,
			"source_addresses":    utils.FlattenStringSlice(rule.SourceAddresses),
			"source_ip_groups":    utils.FlattenStringSlice(rule.SourceIPGroups),
			"destination_address": destinationAddr,
			"destination_ports":   utils.FlattenStringSlice(rule.DestinationPorts),
			"translated_address":  translatedAddress,
			"translated_port":     translatedPort,
			"translated_fqdn":     translatedFQDN,
		})
	}
	return output, nil
}
