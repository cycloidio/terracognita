package network

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	commonValidate "github.com/hashicorp/terraform-provider-azurerm/helpers/validate"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/locks"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

var VPNGatewayResourceName = "azurerm_vpn_gateway"

func resourceVPNGateway() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceVPNGatewayCreate,
		Read:   resourceVPNGatewayRead,
		Update: resourceVPNGatewayUpdate,
		Delete: resourceVPNGatewayDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.VpnGatewayID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(90 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(90 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"virtual_hub_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.VirtualHubID,
			},

			"routing_preference": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Microsoft Network",
					"Internet",
				}, false),
			},

			"bgp_route_translation_for_nat_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			"bgp_settings": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"asn": {
							Type:     pluginsdk.TypeInt,
							Required: true,
							ForceNew: true,
						},

						"peer_weight": {
							Type:     pluginsdk.TypeInt,
							Required: true,
							ForceNew: true,
						},

						"bgp_peering_address": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},

						"instance_0_bgp_peering_address": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"custom_ips": {
										Type:     pluginsdk.TypeSet,
										Required: true,
										Elem: &pluginsdk.Schema{
											Type:         pluginsdk.TypeString,
											ValidateFunc: commonValidate.IPv4Address,
										},
									},

									"ip_configuration_id": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},

									"default_ips": {
										Type:     pluginsdk.TypeSet,
										Computed: true,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
										},
									},

									"tunnel_ips": {
										Type:     pluginsdk.TypeSet,
										Computed: true,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
										},
									},
								},
							},
						},

						"instance_1_bgp_peering_address": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"custom_ips": {
										Type:     pluginsdk.TypeSet,
										Required: true,
										Elem: &pluginsdk.Schema{
											Type:         pluginsdk.TypeString,
											ValidateFunc: commonValidate.IPv4Address,
										},
									},

									"ip_configuration_id": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},

									"default_ips": {
										Type:     pluginsdk.TypeSet,
										Computed: true,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
										},
									},

									"tunnel_ips": {
										Type:     pluginsdk.TypeSet,
										Computed: true,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
										},
									},
								},
							},
						},
					},
				},
			},

			"scale_unit": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntAtLeast(0),
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceVPNGatewayCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VpnGatewaysClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewVpnGatewayID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
		}
	}

	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurerm_vpn_gateway", id.ID())
	}

	bgpSettingsRaw := d.Get("bgp_settings").([]interface{})
	bgpSettings := expandVPNGatewayBGPSettings(bgpSettingsRaw)

	location := azure.NormalizeLocation(d.Get("location").(string))
	scaleUnit := d.Get("scale_unit").(int)
	virtualHubId := d.Get("virtual_hub_id").(string)
	t := d.Get("tags").(map[string]interface{})

	parameters := network.VpnGateway{
		Location: utils.String(location),
		VpnGatewayProperties: &network.VpnGatewayProperties{
			EnableBgpRouteTranslationForNat: utils.Bool(d.Get("bgp_route_translation_for_nat_enabled").(bool)),
			BgpSettings:                     bgpSettings,
			VirtualHub: &network.SubResource{
				ID: utils.String(virtualHubId),
			},
			VpnGatewayScaleUnit:         utils.Int32(int32(scaleUnit)),
			IsRoutingPreferenceInternet: utils.Bool(d.Get("routing_preference").(string) == "Internet"),
		},
		Tags: tags.Expand(t),
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, parameters)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation/update of %q: %+v", id, err)
	}
	if err := waitForCompletion(d, ctx, client, id.ResourceGroup, id.Name); err != nil {
		return err
	}
	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	// `vpnGatewayParameters.Properties.bgpSettings.bgpPeeringAddress` customer cannot provide this field during create. This will be set with default value once gateway is created.
	// it could only be updated
	if len(bgpSettingsRaw) > 0 && resp.VpnGatewayProperties != nil && resp.VpnGatewayProperties.BgpSettings != nil && resp.VpnGatewayProperties.BgpSettings.BgpPeeringAddresses != nil {
		val := bgpSettingsRaw[0].(map[string]interface{})
		input0 := val["instance_0_bgp_peering_address"].([]interface{})
		input1 := val["instance_1_bgp_peering_address"].([]interface{})

		if len(input0) > 0 || len(input1) > 0 {
			if len(input0) > 0 {
				val := input0[0].(map[string]interface{})
				(*resp.VpnGatewayProperties.BgpSettings.BgpPeeringAddresses)[0].CustomBgpIPAddresses = utils.ExpandStringSlice(val["custom_ips"].(*pluginsdk.Set).List())
			}
			if len(input1) > 0 {
				val := input1[0].(map[string]interface{})
				(*resp.VpnGatewayProperties.BgpSettings.BgpPeeringAddresses)[1].CustomBgpIPAddresses = utils.ExpandStringSlice(val["custom_ips"].(*pluginsdk.Set).List())
			}

			future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, resp)
			if err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
				return fmt.Errorf("waiting for creation/update of %q: %+v", id, err)
			}
			if err := waitForCompletion(d, ctx, client, id.ResourceGroup, id.Name); err != nil {
				return err
			}
		}
	}

	d.SetId(id.ID())

	return resourceVPNGatewayRead(d, meta)
}

func resourceVPNGatewayUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VpnGatewaysClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VpnGatewayID(d.Id())
	if err != nil {
		return err
	}

	locks.ByName(id.Name, VPNGatewayResourceName)
	defer locks.UnlockByName(id.Name, VPNGatewayResourceName)

	existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("retrieving for presence of existing %s: %+v", id, err)
	}

	if d.HasChange("scale_unit") {
		existing.VpnGatewayScaleUnit = utils.Int32(int32(d.Get("scale_unit").(int)))
	}
	if d.HasChange("tags") {
		existing.Tags = tags.Expand(d.Get("tags").(map[string]interface{}))
	}
	if d.HasChange("bgp_route_translation_for_nat_enabled") {
		existing.EnableBgpRouteTranslationForNat = utils.Bool(d.Get("bgp_route_translation_for_nat_enabled").(bool))
	}

	bgpSettingsRaw := d.Get("bgp_settings").([]interface{})
	if len(bgpSettingsRaw) > 0 {
		val := bgpSettingsRaw[0].(map[string]interface{})

		if d.HasChange("bgp_settings.0.instance_0_bgp_peering_address") {
			if input := val["instance_0_bgp_peering_address"].([]interface{}); len(input) > 0 {
				val := input[0].(map[string]interface{})
				(*existing.VpnGatewayProperties.BgpSettings.BgpPeeringAddresses)[0].CustomBgpIPAddresses = utils.ExpandStringSlice(val["custom_ips"].(*pluginsdk.Set).List())
			}
		}
		if d.HasChange("bgp_settings.0.instance_1_bgp_peering_address") {
			if input := val["instance_1_bgp_peering_address"].([]interface{}); len(input) > 0 {
				val := input[0].(map[string]interface{})
				(*existing.VpnGatewayProperties.BgpSettings.BgpPeeringAddresses)[1].CustomBgpIPAddresses = utils.ExpandStringSlice(val["custom_ips"].(*pluginsdk.Set).List())
			}
		}
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, existing)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation/update of %q: %+v", id, err)
	}
	if err := waitForCompletion(d, ctx, client, id.ResourceGroup, id.Name); err != nil {
		return err
	}

	return resourceVPNGatewayRead(d, meta)
}

func resourceVPNGatewayRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VpnGatewaysClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VpnGatewayID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] %s was not found - removing from state", *id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if props := resp.VpnGatewayProperties; props != nil {
		if err := d.Set("bgp_settings", flattenVPNGatewayBGPSettings(props.BgpSettings)); err != nil {
			return fmt.Errorf("setting `bgp_settings`: %+v", err)
		}

		bgpRouteTranslationForNatEnabled := false
		if props.EnableBgpRouteTranslationForNat != nil {
			bgpRouteTranslationForNatEnabled = *props.EnableBgpRouteTranslationForNat
		}
		d.Set("bgp_route_translation_for_nat_enabled", bgpRouteTranslationForNatEnabled)

		scaleUnit := 0
		if props.VpnGatewayScaleUnit != nil {
			scaleUnit = int(*props.VpnGatewayScaleUnit)
		}
		d.Set("scale_unit", scaleUnit)

		virtualHubId := ""
		if props.VirtualHub != nil && props.VirtualHub.ID != nil {
			virtualHubId = *props.VirtualHub.ID
		}
		d.Set("virtual_hub_id", virtualHubId)

		isRoutingPreferenceInternet := "Microsoft Network"
		if props.IsRoutingPreferenceInternet != nil && *props.IsRoutingPreferenceInternet {
			isRoutingPreferenceInternet = "Internet"
		}
		d.Set("routing_preference", isRoutingPreferenceInternet)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceVPNGatewayDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VpnGatewaysClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VpnGatewayID(d.Id())
	if err != nil {
		return err
	}

	deleteFuture, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if response.WasNotFound(deleteFuture.Response()) {
			return nil
		}

		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	err = deleteFuture.WaitForCompletionRef(ctx, client.Client)
	if err != nil {
		if response.WasNotFound(deleteFuture.Response()) {
			return nil
		}

		return fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
	}

	return nil
}

func waitForCompletion(d *pluginsdk.ResourceData, ctx context.Context, client *network.VpnGatewaysClient, resourceGroup, name string) error {
	log.Printf("[DEBUG] Waiting for Virtual Hub %q (Resource Group %q) to become available", name, resourceGroup)
	stateConf := &pluginsdk.StateChangeConf{
		Pending:                   []string{"pending"},
		Target:                    []string{"available"},
		Refresh:                   vpnGatewayWaitForCreatedRefreshFunc(ctx, client, resourceGroup, name),
		Delay:                     30 * time.Second,
		PollInterval:              10 * time.Second,
		ContinuousTargetOccurence: 3,
	}

	if d.IsNewResource() {
		stateConf.Timeout = d.Timeout(pluginsdk.TimeoutCreate)
	} else {
		stateConf.Timeout = d.Timeout(pluginsdk.TimeoutUpdate)
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for creation of Virtual Hub %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	return nil
}

func expandVPNGatewayBGPSettings(input []interface{}) *network.BgpSettings {
	if len(input) == 0 {
		return nil
	}

	val := input[0].(map[string]interface{})
	return &network.BgpSettings{
		Asn:        utils.Int64(int64(val["asn"].(int))),
		PeerWeight: utils.Int32(int32(val["peer_weight"].(int))),
	}
}

func flattenVPNGatewayBGPSettings(input *network.BgpSettings) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	asn := 0
	if input.Asn != nil {
		asn = int(*input.Asn)
	}

	bgpPeeringAddress := ""
	if input.BgpPeeringAddress != nil {
		bgpPeeringAddress = *input.BgpPeeringAddress
	}

	peerWeight := 0
	if input.PeerWeight != nil {
		peerWeight = int(*input.PeerWeight)
	}

	var instance0BgpPeeringAddress, instance1BgpPeeringAddress []interface{}
	if input.BgpPeeringAddresses != nil && len(*input.BgpPeeringAddresses) > 0 {
		instance0BgpPeeringAddress = flattenVPNGatewayIPConfigurationBgpPeeringAddress((*input.BgpPeeringAddresses)[0])
	}
	if input.BgpPeeringAddresses != nil && len(*input.BgpPeeringAddresses) > 1 {
		instance1BgpPeeringAddress = flattenVPNGatewayIPConfigurationBgpPeeringAddress((*input.BgpPeeringAddresses)[1])
	}

	return []interface{}{
		map[string]interface{}{
			"asn":                            asn,
			"bgp_peering_address":            bgpPeeringAddress,
			"instance_0_bgp_peering_address": instance0BgpPeeringAddress,
			"instance_1_bgp_peering_address": instance1BgpPeeringAddress,
			"peer_weight":                    peerWeight,
		},
	}
}

func flattenVPNGatewayIPConfigurationBgpPeeringAddress(input network.IPConfigurationBgpPeeringAddress) []interface{} {
	ipConfigurationID := ""
	if input.IpconfigurationID != nil {
		ipConfigurationID = *input.IpconfigurationID
	}

	return []interface{}{
		map[string]interface{}{
			"ip_configuration_id": ipConfigurationID,
			"custom_ips":          utils.FlattenStringSlice(input.CustomBgpIPAddresses),
			"default_ips":         utils.FlattenStringSlice(input.DefaultBgpIPAddresses),
			"tunnel_ips":          utils.FlattenStringSlice(input.TunnelIPAddresses),
		},
	}
}

func vpnGatewayWaitForCreatedRefreshFunc(ctx context.Context, client *network.VpnGatewaysClient, resourceGroup, name string) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Checking to see if VPN Gateway %q (Resource Group %q) has finished provisioning..", name, resourceGroup)

		resp, err := client.Get(ctx, resourceGroup, name)
		if err != nil {
			log.Printf("[DEBUG] Error retrieving VPN Gateway %q (Resource Group %q): %+v", name, resourceGroup, err)
			return nil, "error", fmt.Errorf("retrieving VPN Gateway %q (Resource Group %q): %+v", name, resourceGroup, err)
		}

		if resp.VpnGatewayProperties == nil {
			log.Printf("[DEBUG] Error retrieving VPN Gateway %q (Resource Group %q): `properties` was nil", name, resourceGroup)
			return nil, "error", fmt.Errorf("retrieving VPN Gateway %q (Resource Group %q): `properties` was nil", name, resourceGroup)
		}

		log.Printf("[DEBUG] VPN Gateway %q (Resource Group %q) is %q..", name, resourceGroup, string(resp.VpnGatewayProperties.ProvisioningState))
		switch resp.VpnGatewayProperties.ProvisioningState {
		case network.ProvisioningStateSucceeded:
			return "available", "available", nil

		case network.ProvisioningStateFailed:
			return "error", "error", fmt.Errorf("VPN Gateway %q (Resource Group %q) is in provisioningState `Failed`", name, resourceGroup)

		default:
			return "pending", "pending", nil
		}
	}
}
