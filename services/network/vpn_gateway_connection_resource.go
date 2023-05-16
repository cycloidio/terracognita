package network

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/locks"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceVPNGatewayConnection() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceVpnGatewayConnectionResourceCreateUpdate,
		Read:   resourceVpnGatewayConnectionResourceRead,
		Update: resourceVpnGatewayConnectionResourceCreateUpdate,
		Delete: resourceVpnGatewayConnectionResourceDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.VpnConnectionID(id)
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
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"vpn_gateway_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.VpnGatewayID,
			},

			"remote_vpn_site_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.VpnSiteID,
			},

			"internet_security_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			// Service will create a route table for the user if this is not specified.
			"routing": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"associated_route_table": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validate.HubRouteTableID,
						},
						"propagated_route_table": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"route_table_ids": {
										Type:     pluginsdk.TypeList,
										Required: true,
										Elem: &pluginsdk.Schema{
											Type:         pluginsdk.TypeString,
											ValidateFunc: validate.HubRouteTableID,
										},
									},

									"labels": {
										Type:     pluginsdk.TypeSet,
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

			"vpn_link": {
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

						"vpn_site_link_id": {
							Type:     pluginsdk.TypeString,
							Required: true,
							// The vpn site link associated with one link connection can not be updated
							ForceNew:     true,
							ValidateFunc: validate.VpnSiteLinkID,
						},

						"egress_nat_rule_ids": {
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Elem: &pluginsdk.Schema{
								Type:         pluginsdk.TypeString,
								ValidateFunc: validate.VpnGatewayNatRuleID,
							},
						},

						"ingress_nat_rule_ids": {
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Elem: &pluginsdk.Schema{
								Type:         pluginsdk.TypeString,
								ValidateFunc: validate.VpnGatewayNatRuleID,
							},
						},

						"connection_mode": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(network.VpnLinkConnectionModeDefault),
								string(network.VpnLinkConnectionModeInitiatorOnly),
								string(network.VpnLinkConnectionModeResponderOnly),
							}, false),
							Default: string(network.VpnLinkConnectionModeDefault),
						},

						"route_weight": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
							Default:      0,
						},

						"protocol": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(network.VirtualNetworkGatewayConnectionProtocolIKEv1),
								string(network.VirtualNetworkGatewayConnectionProtocolIKEv2),
							}, false),
							Default: string(network.VirtualNetworkGatewayConnectionProtocolIKEv2),
						},

						"bandwidth_mbps": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
							Default:      10,
						},

						"shared_key": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},

						"bgp_enabled": {
							Type:     pluginsdk.TypeBool,
							ForceNew: true,
							Optional: true,
							Default:  false,
						},

						"ipsec_policy": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"sa_lifetime_sec": {
										Type:         pluginsdk.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(300, 172799),
									},
									"sa_data_size_kb": {
										Type:         pluginsdk.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1024, 2147483647),
									},
									"encryption_algorithm": {
										Type:     pluginsdk.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											string(network.IpsecEncryptionAES128),
											string(network.IpsecEncryptionAES192),
											string(network.IpsecEncryptionAES256),
											string(network.IpsecEncryptionDES),
											string(network.IpsecEncryptionDES3),
											string(network.IpsecEncryptionGCMAES128),
											string(network.IpsecEncryptionGCMAES192),
											string(network.IpsecEncryptionGCMAES256),
											string(network.IpsecEncryptionNone),
										}, false),
									},
									"integrity_algorithm": {
										Type:     pluginsdk.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											string(network.IpsecIntegrityMD5),
											string(network.IpsecIntegritySHA1),
											string(network.IpsecIntegritySHA256),
											string(network.IpsecIntegrityGCMAES128),
											string(network.IpsecIntegrityGCMAES192),
											string(network.IpsecIntegrityGCMAES256),
										}, false),
									},

									"ike_encryption_algorithm": {
										Type:     pluginsdk.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											string(network.IkeEncryptionDES),
											string(network.IkeEncryptionDES3),
											string(network.IkeEncryptionAES128),
											string(network.IkeEncryptionAES192),
											string(network.IkeEncryptionAES256),
											string(network.IkeEncryptionGCMAES128),
											string(network.IkeEncryptionGCMAES256),
										}, false),
									},

									"ike_integrity_algorithm": {
										Type:     pluginsdk.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											string(network.IkeIntegrityMD5),
											string(network.IkeIntegritySHA1),
											string(network.IkeIntegritySHA256),
											string(network.IkeIntegritySHA384),
											string(network.IkeIntegrityGCMAES128),
											string(network.IkeIntegrityGCMAES256),
										}, false),
									},

									"dh_group": {
										Type:     pluginsdk.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											string(network.DhGroupNone),
											string(network.DhGroupDHGroup1),
											string(network.DhGroupDHGroup2),
											string(network.DhGroupDHGroup14),
											string(network.DhGroupDHGroup24),
											string(network.DhGroupDHGroup2048),
											string(network.DhGroupECP256),
											string(network.DhGroupECP384),
										}, false),
									},

									"pfs_group": {
										Type:     pluginsdk.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											string(network.PfsGroupNone),
											string(network.PfsGroupPFS1),
											string(network.PfsGroupPFS2),
											string(network.PfsGroupPFS14),
											string(network.PfsGroupPFS24),
											string(network.PfsGroupPFS2048),
											string(network.PfsGroupPFSMM),
											string(network.PfsGroupECP256),
											string(network.PfsGroupECP384),
										}, false),
									},
								},
							},
						},

						"ratelimit_enabled": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
							Default:  false,
						},

						"local_azure_ip_address_enabled": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
							Default:  false,
						},

						"policy_based_traffic_selector_enabled": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
							Default:  false,
						},

						"custom_bgp_address": {
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"ip_address": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.IsIPv4Address,
									},

									"ip_configuration_id": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
					},
				},
			},

			"traffic_selector_policy": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"local_address_ranges": {
							Type:     pluginsdk.TypeSet,
							Required: true,
							Elem: &pluginsdk.Schema{
								Type:         pluginsdk.TypeString,
								ValidateFunc: validation.IsCIDR,
							},
						},

						"remote_address_ranges": {
							Type:     pluginsdk.TypeSet,
							Required: true,
							Elem: &pluginsdk.Schema{
								Type:         pluginsdk.TypeString,
								ValidateFunc: validation.IsCIDR,
							},
						},
					},
				},
			},
		},
	}
}

func resourceVpnGatewayConnectionResourceCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VpnConnectionsClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	gatewayId, err := parse.VpnGatewayID(d.Get("vpn_gateway_id").(string))
	if err != nil {
		return err
	}

	if d.IsNewResource() {
		resp, err := client.Get(ctx, gatewayId.ResourceGroup, gatewayId.Name, name)
		if err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("checking for existing Vpn Gateway Connection Resource %q (Resource Group %q / VPN Gateway %q): %+v", name, gatewayId.ResourceGroup, gatewayId.Name, err)
			}
		}

		if resp.ID != nil && *resp.ID != "" {
			return tf.ImportAsExistsError("azurerm_vpn_gateway_connection", *resp.ID)
		}
	}

	locks.ByName(gatewayId.Name, VPNGatewayResourceName)
	defer locks.UnlockByName(gatewayId.Name, VPNGatewayResourceName)

	param := network.VpnConnection{
		Name: &name,
		VpnConnectionProperties: &network.VpnConnectionProperties{
			EnableInternetSecurity: utils.Bool(d.Get("internet_security_enabled").(bool)),
			RemoteVpnSite: &network.SubResource{
				ID: utils.String(d.Get("remote_vpn_site_id").(string)),
			},
			VpnLinkConnections:   expandVpnGatewayConnectionVpnSiteLinkConnections(d.Get("vpn_link").([]interface{})),
			RoutingConfiguration: expandVpnGatewayConnectionRoutingConfiguration(d.Get("routing").([]interface{})),
		},
	}

	if v, ok := d.GetOk("traffic_selector_policy"); ok {
		param.VpnConnectionProperties.TrafficSelectorPolicies = expandVpnGatewayConnectionTrafficSelectorPolicy(v.(*pluginsdk.Set).List())
	}

	future, err := client.CreateOrUpdate(ctx, gatewayId.ResourceGroup, gatewayId.Name, name, param)
	if err != nil {
		return fmt.Errorf("creating Vpn Gateway Connection Resource %q (Resource Group %q / VPN Gateway %q): %+v", name, gatewayId.ResourceGroup, gatewayId.Name, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of Vpn Gateway Connection Resource %q (Resource Group %q / VPN Gateway %q): %+v", name, gatewayId.ResourceGroup, gatewayId.Name, err)
	}

	resp, err := client.Get(ctx, gatewayId.ResourceGroup, gatewayId.Name, name)
	if err != nil {
		return fmt.Errorf("retrieving Vpn Gateway Connection Resource %q (Resource Group %q / VPN Gateway: %q): %+v", name, gatewayId.ResourceGroup, gatewayId.Name, err)
	}
	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("empty or nil ID returned for Vpn Gateway Connection Resource %q (Resource Group %q / VPN Gateway: %q) ID", name, gatewayId.ResourceGroup, gatewayId.Name)
	}

	id, err := parse.VpnConnectionID(*resp.ID)
	if err != nil {
		return err
	}
	d.SetId(id.ID())

	return resourceVpnGatewayConnectionResourceRead(d, meta)
}

func resourceVpnGatewayConnectionResourceRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VpnConnectionsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VpnConnectionID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.VpnGatewayName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Vpn Gateway Connection Resource %q was not found in VPN Gateway %q in Resource Group %q - removing from state!", id.Name, id.VpnGatewayName, id.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Vpn Gateway Connection Resource %q (Resource Group %q / VPN Gateway %q): %+v", id.Name, id.ResourceGroup, id.VpnGatewayName, err)
	}

	d.Set("name", id.Name)

	gatewayId := parse.NewVpnGatewayID(id.SubscriptionId, id.ResourceGroup, id.VpnGatewayName)
	d.Set("vpn_gateway_id", gatewayId.ID())

	if prop := resp.VpnConnectionProperties; prop != nil {
		vpnSiteId := ""
		if site := prop.RemoteVpnSite; site != nil {
			if id := site.ID; id != nil {
				theVpnSiteId, err := parse.VpnSiteID(*id)
				if err != nil {
					return err
				}
				vpnSiteId = theVpnSiteId.ID()
			}
		}
		d.Set("remote_vpn_site_id", vpnSiteId)

		enableInternetSecurity := false
		if prop.EnableInternetSecurity != nil {
			enableInternetSecurity = *prop.EnableInternetSecurity
		}
		d.Set("internet_security_enabled", enableInternetSecurity)

		if err := d.Set("routing", flattenVpnGatewayConnectionRoutingConfiguration(prop.RoutingConfiguration)); err != nil {
			return fmt.Errorf(`setting "routing": %v`, err)
		}

		if err := d.Set("vpn_link", flattenVpnGatewayConnectionVpnSiteLinkConnections(prop.VpnLinkConnections)); err != nil {
			return fmt.Errorf(`setting "vpn_link": %v`, err)
		}

		if err := d.Set("traffic_selector_policy", flattenVpnGatewayConnectionTrafficSelectorPolicy(prop.TrafficSelectorPolicies)); err != nil {
			return fmt.Errorf("setting `traffic_selector_policy`: %+v", err)
		}
	}

	return nil
}

func resourceVpnGatewayConnectionResourceDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VpnConnectionsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VpnConnectionID(d.Id())
	if err != nil {
		return err
	}

	locks.ByName(id.VpnGatewayName, VPNGatewayResourceName)
	defer locks.UnlockByName(id.VpnGatewayName, VPNGatewayResourceName)

	future, err := client.Delete(ctx, id.ResourceGroup, id.VpnGatewayName, id.Name)
	if err != nil {
		return fmt.Errorf("deleting Vpn Gateway Connection Resource %q (Resource Group %q / VPN Gateway %q): %+v", id.Name, id.ResourceGroup, id.VpnGatewayName, err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("waiting for the deletion of VPN Gateway Connection %q (Resource Group %q / VPN Gateway %q): %+v", id.Name, id.ResourceGroup, id.VpnGatewayName, err)
		}
	}

	return nil
}

func expandVpnGatewayConnectionVpnSiteLinkConnections(input []interface{}) *[]network.VpnSiteLinkConnection {
	if len(input) == 0 {
		return nil
	}

	result := make([]network.VpnSiteLinkConnection, 0)

	for _, e := range input {
		e := e.(map[string]interface{})
		v := network.VpnSiteLinkConnection{
			Name: utils.String(e["name"].(string)),
			VpnSiteLinkConnectionProperties: &network.VpnSiteLinkConnectionProperties{
				VpnSiteLink:                    &network.SubResource{ID: utils.String(e["vpn_site_link_id"].(string))},
				RoutingWeight:                  utils.Int32(int32(e["route_weight"].(int))),
				VpnConnectionProtocolType:      network.VirtualNetworkGatewayConnectionProtocol(e["protocol"].(string)),
				VpnLinkConnectionMode:          network.VpnLinkConnectionMode(e["connection_mode"].(string)),
				ConnectionBandwidth:            utils.Int32(int32(e["bandwidth_mbps"].(int))),
				EnableBgp:                      utils.Bool(e["bgp_enabled"].(bool)),
				IpsecPolicies:                  expandVpnGatewayConnectionIpSecPolicies(e["ipsec_policy"].([]interface{})),
				EnableRateLimiting:             utils.Bool(e["ratelimit_enabled"].(bool)),
				UseLocalAzureIPAddress:         utils.Bool(e["local_azure_ip_address_enabled"].(bool)),
				UsePolicyBasedTrafficSelectors: utils.Bool(e["policy_based_traffic_selector_enabled"].(bool)),
				VpnGatewayCustomBgpAddresses:   expandVpnGatewayConnectionCustomBgpAddresses(e["custom_bgp_address"].(*pluginsdk.Set).List()),
			},
		}

		if egressNatRuleIds := e["egress_nat_rule_ids"].(*pluginsdk.Set).List(); len(egressNatRuleIds) != 0 {
			v.VpnSiteLinkConnectionProperties.EgressNatRules = expandVpnGatewayConnectionNatRuleIds(egressNatRuleIds)
		}

		if ingressNatRuleIds := e["ingress_nat_rule_ids"].(*pluginsdk.Set).List(); len(ingressNatRuleIds) != 0 {
			v.VpnSiteLinkConnectionProperties.IngressNatRules = expandVpnGatewayConnectionNatRuleIds(ingressNatRuleIds)
		}

		if sharedKey := e["shared_key"]; sharedKey != "" {
			sharedKey := sharedKey.(string)
			v.VpnSiteLinkConnectionProperties.SharedKey = &sharedKey
		}
		result = append(result, v)
	}

	return &result
}

func flattenVpnGatewayConnectionVpnSiteLinkConnections(input *[]network.VpnSiteLinkConnection) interface{} {
	if input == nil {
		return []interface{}{}
	}

	output := make([]interface{}, 0)

	for _, e := range *input {
		name := ""
		if e.Name != nil {
			name = *e.Name
		}

		vpnSiteLinkId := ""
		if e.VpnSiteLink != nil && e.VpnSiteLink.ID != nil {
			vpnSiteLinkId = *e.VpnSiteLink.ID
		}

		routeWeight := 0
		if e.RoutingWeight != nil {
			routeWeight = int(*e.RoutingWeight)
		}

		bandwidth := 0
		if e.ConnectionBandwidth != nil {
			bandwidth = int(*e.ConnectionBandwidth)
		}

		sharedKey := ""
		if e.SharedKey != nil {
			sharedKey = *e.SharedKey
		}

		bgpEnabled := false
		if e.EnableBgp != nil {
			bgpEnabled = *e.EnableBgp
		}

		usePolicyBased := false
		if e.UsePolicyBasedTrafficSelectors != nil {
			usePolicyBased = *e.UsePolicyBasedTrafficSelectors
		}

		rateLimitEnabled := false
		if e.EnableRateLimiting != nil {
			rateLimitEnabled = *e.EnableRateLimiting
		}

		useLocalAzureIpAddress := false
		if e.UseLocalAzureIPAddress != nil {
			useLocalAzureIpAddress = *e.UseLocalAzureIPAddress
		}

		v := map[string]interface{}{
			"name":                                  name,
			"egress_nat_rule_ids":                   flattenVpnGatewayConnectionNatRuleIds(e.VpnSiteLinkConnectionProperties.EgressNatRules),
			"ingress_nat_rule_ids":                  flattenVpnGatewayConnectionNatRuleIds(e.VpnSiteLinkConnectionProperties.IngressNatRules),
			"vpn_site_link_id":                      vpnSiteLinkId,
			"route_weight":                          routeWeight,
			"protocol":                              string(e.VpnConnectionProtocolType),
			"connection_mode":                       string(e.VpnLinkConnectionMode),
			"bandwidth_mbps":                        bandwidth,
			"shared_key":                            sharedKey,
			"bgp_enabled":                           bgpEnabled,
			"ipsec_policy":                          flattenVpnGatewayConnectionIpSecPolicies(e.IpsecPolicies),
			"ratelimit_enabled":                     rateLimitEnabled,
			"local_azure_ip_address_enabled":        useLocalAzureIpAddress,
			"policy_based_traffic_selector_enabled": usePolicyBased,
			"custom_bgp_address":                    flattenVpnGatewayConnectionCustomBgpAddresses(e.VpnGatewayCustomBgpAddresses),
		}

		output = append(output, v)
	}

	return output
}

func expandVpnGatewayConnectionIpSecPolicies(input []interface{}) *[]network.IpsecPolicy {
	if len(input) == 0 {
		return nil
	}

	result := make([]network.IpsecPolicy, 0)

	for _, e := range input {
		e := e.(map[string]interface{})
		v := network.IpsecPolicy{
			SaLifeTimeSeconds:   utils.Int32(int32(e["sa_lifetime_sec"].(int))),
			SaDataSizeKilobytes: utils.Int32(int32(e["sa_data_size_kb"].(int))),
			IpsecEncryption:     network.IpsecEncryption(e["encryption_algorithm"].(string)),
			IpsecIntegrity:      network.IpsecIntegrity(e["integrity_algorithm"].(string)),
			IkeEncryption:       network.IkeEncryption(e["ike_encryption_algorithm"].(string)),
			IkeIntegrity:        network.IkeIntegrity(e["ike_integrity_algorithm"].(string)),
			DhGroup:             network.DhGroup(e["dh_group"].(string)),
			PfsGroup:            network.PfsGroup(e["pfs_group"].(string)),
		}
		result = append(result, v)
	}

	return &result
}

func flattenVpnGatewayConnectionIpSecPolicies(input *[]network.IpsecPolicy) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	output := make([]interface{}, 0)

	for _, e := range *input {
		saLifetimeSec := 0
		if e.SaLifeTimeSeconds != nil {
			saLifetimeSec = int(*e.SaLifeTimeSeconds)
		}

		saDataSizeKb := 0
		if e.SaDataSizeKilobytes != nil {
			saDataSizeKb = int(*e.SaDataSizeKilobytes)
		}

		v := map[string]interface{}{
			"sa_lifetime_sec":          saLifetimeSec,
			"sa_data_size_kb":          saDataSizeKb,
			"encryption_algorithm":     string(e.IpsecEncryption),
			"integrity_algorithm":      string(e.IpsecIntegrity),
			"ike_encryption_algorithm": string(e.IkeEncryption),
			"ike_integrity_algorithm":  string(e.IkeIntegrity),
			"dh_group":                 string(e.DhGroup),
			"pfs_group":                string(e.PfsGroup),
		}

		output = append(output, v)
	}

	return output
}

func expandVpnGatewayConnectionRoutingConfiguration(input []interface{}) *network.RoutingConfiguration {
	if len(input) == 0 || input[0] == nil {
		return nil
	}
	raw := input[0].(map[string]interface{})
	output := &network.RoutingConfiguration{
		AssociatedRouteTable: &network.SubResource{ID: utils.String(raw["associated_route_table"].(string))},
	}

	if v := raw["propagated_route_table"].([]interface{}); len(v) != 0 {
		output.PropagatedRouteTables = expandVpnGatewayConnectionPropagatedRouteTable(v)
	}

	return output
}

func flattenVpnGatewayConnectionRoutingConfiguration(input *network.RoutingConfiguration) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	associateRouteTable := ""
	if input.AssociatedRouteTable != nil && input.AssociatedRouteTable.ID != nil {
		associateRouteTable = *input.AssociatedRouteTable.ID
	}

	return []interface{}{
		map[string]interface{}{
			"propagated_route_table": flattenVpnGatewayConnectionPropagatedRouteTable(input.PropagatedRouteTables),
			"associated_route_table": associateRouteTable,
		},
	}
}

func flattenVpnGatewayConnectionPropagatedRouteTable(input *network.PropagatedRouteTable) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	labels := make([]interface{}, 0)
	if input.Labels != nil {
		labels = utils.FlattenStringSlice(input.Labels)
	}

	routeTableIds := make([]interface{}, 0)
	if input.Ids != nil {
		routeTableIds = flattenSubResourcesToIDs(input.Ids)
	}

	return []interface{}{
		map[string]interface{}{
			"labels":          labels,
			"route_table_ids": routeTableIds,
		},
	}
}

func expandVpnGatewayConnectionTrafficSelectorPolicy(input []interface{}) *[]network.TrafficSelectorPolicy {
	results := make([]network.TrafficSelectorPolicy, 0)

	for _, item := range input {
		v := item.(map[string]interface{})

		results = append(results, network.TrafficSelectorPolicy{
			LocalAddressRanges:  utils.ExpandStringSlice(v["local_address_ranges"].(*pluginsdk.Set).List()),
			RemoteAddressRanges: utils.ExpandStringSlice(v["remote_address_ranges"].(*pluginsdk.Set).List()),
		})
	}

	return &results
}

func flattenVpnGatewayConnectionTrafficSelectorPolicy(input *[]network.TrafficSelectorPolicy) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	for _, item := range *input {
		results = append(results, map[string]interface{}{
			"local_address_ranges":  utils.FlattenStringSlice(item.LocalAddressRanges),
			"remote_address_ranges": utils.FlattenStringSlice(item.RemoteAddressRanges),
		})
	}

	return results
}

func expandVpnGatewayConnectionPropagatedRouteTable(input []interface{}) *network.PropagatedRouteTable {
	if len(input) == 0 {
		return &network.PropagatedRouteTable{}
	}

	v := input[0].(map[string]interface{})

	result := network.PropagatedRouteTable{
		Ids: expandIDsToSubResources(v["route_table_ids"].([]interface{})),
	}

	if labels := v["labels"].(*pluginsdk.Set).List(); len(labels) != 0 {
		result.Labels = utils.ExpandStringSlice(labels)
	}

	return &result
}

func expandVpnGatewayConnectionNatRuleIds(input []interface{}) *[]network.SubResource {
	results := make([]network.SubResource, 0)

	for _, item := range input {
		results = append(results, network.SubResource{
			ID: utils.String(item.(string)),
		})
	}

	return &results
}

func flattenVpnGatewayConnectionNatRuleIds(input *[]network.SubResource) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	for _, item := range *input {
		var id string
		if item.ID != nil {
			id = *item.ID
		}

		results = append(results, id)
	}

	return results
}

func expandVpnGatewayConnectionCustomBgpAddresses(input []interface{}) *[]network.GatewayCustomBgpIPAddressIPConfiguration {
	results := make([]network.GatewayCustomBgpIPAddressIPConfiguration, 0)

	for _, item := range input {
		v := item.(map[string]interface{})

		results = append(results, network.GatewayCustomBgpIPAddressIPConfiguration{
			CustomBgpIPAddress: utils.String(v["ip_address"].(string)),
			IPConfigurationID:  utils.String(v["ip_configuration_id"].(string)),
		})
	}

	return &results
}

func flattenVpnGatewayConnectionCustomBgpAddresses(input *[]network.GatewayCustomBgpIPAddressIPConfiguration) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	for _, item := range *input {
		var customBgpIpAddress string
		if item.CustomBgpIPAddress != nil {
			customBgpIpAddress = *item.CustomBgpIPAddress
		}

		var ipConfigurationId string
		if item.IPConfigurationID != nil {
			ipConfigurationId = *item.IPConfigurationID
		}

		results = append(results, map[string]interface{}{
			"ip_address":          customBgpIpAddress,
			"ip_configuration_id": ipConfigurationId,
		})
	}

	return results
}
