package network

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func dataSourceVirtualNetworkGateway() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceVirtualNetworkGatewayRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"resource_group_name": commonschema.ResourceGroupNameForDataSource(),

			"location": commonschema.LocationComputed(),

			"type": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"vpn_type": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			// TODO 4.0: change this from enable_* to *_enabled
			"enable_bgp": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"private_ip_address_enabled": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"active_active": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"sku": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"generation": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"ip_configuration": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},

						"name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},

						"private_ip_address_allocation": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},

						"subnet_id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},

						"public_ip_address_id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"vpn_client_configuration": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"address_space": {
							Type:     pluginsdk.TypeList,
							Computed: true,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
							},
						},

						"aad_tenant": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},

						"aad_audience": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},

						"aad_issuer": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},

						"root_certificate": {
							Type:     pluginsdk.TypeList,
							Computed: true,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"name": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},

									"public_cert_data": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},
								},
							},
						},

						"revoked_certificate": {
							Type:     pluginsdk.TypeList,
							Computed: true,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"name": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},

									"thumbprint": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},
								},
							},
						},

						"radius_server_address": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},

						"radius_server_secret": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},

						"vpn_client_protocols": {
							Type:     pluginsdk.TypeSet,
							Computed: true,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
							},
						},
					},
				},
			},

			"bgp_settings": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"asn": {
							Type:     pluginsdk.TypeInt,
							Computed: true,
						},

						"peering_address": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},

						"peer_weight": {
							Type:     pluginsdk.TypeInt,
							Computed: true,
						},
					},
				},
			},

			"custom_route": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"address_prefixes": {
							Type:     pluginsdk.TypeSet,
							Computed: true,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
							},
						},
					},
				},
			},

			"default_local_network_gateway_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"tags": tags.SchemaDataSource(),
		},
	}
}

func dataSourceVirtualNetworkGatewayRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VnetGatewayClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewVirtualNetworkGatewayID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("%s was not found", id)
		}

		return fmt.Errorf("making Read request on %s: %+v", id, err)
	}

	d.SetId(id.ID())

	d.Set("name", resp.Name)
	d.Set("resource_group_name", id.ResourceGroup)

	d.Set("location", location.NormalizeNilable(resp.Location))

	if gw := resp.VirtualNetworkGatewayPropertiesFormat; gw != nil {
		d.Set("type", string(gw.GatewayType))
		d.Set("enable_bgp", gw.EnableBgp)
		d.Set("private_ip_address_enabled", gw.EnablePrivateIPAddress)
		d.Set("active_active", gw.ActiveActive)
		d.Set("generation", string(gw.VpnGatewayGeneration))

		if string(gw.VpnType) != "" {
			d.Set("vpn_type", string(gw.VpnType))
		}

		if gw.GatewayDefaultSite != nil {
			d.Set("default_local_network_gateway_id", gw.GatewayDefaultSite.ID)
		}

		if gw.Sku != nil {
			d.Set("sku", string(gw.Sku.Name))
		}

		if err := d.Set("ip_configuration", flattenVirtualNetworkGatewayDataSourceIPConfigurations(gw.IPConfigurations)); err != nil {
			return fmt.Errorf("setting `ip_configuration`: %+v", err)
		}

		vpnConfigFlat := flattenVirtualNetworkGatewayDataSourceVpnClientConfig(gw.VpnClientConfiguration)
		if err := d.Set("vpn_client_configuration", vpnConfigFlat); err != nil {
			return fmt.Errorf("setting `vpn_client_configuration`: %+v", err)
		}

		bgpSettingsFlat := flattenVirtualNetworkGatewayDataSourceBgpSettings(gw.BgpSettings)
		if err := d.Set("bgp_settings", bgpSettingsFlat); err != nil {
			return fmt.Errorf("setting `bgp_settings`: %+v", err)
		}

		if err := d.Set("custom_route", flattenVirtualNetworkGatewayAddressSpace(gw.CustomRoutes)); err != nil {
			return fmt.Errorf("setting `custom_route`: %+v", err)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func flattenVirtualNetworkGatewayDataSourceIPConfigurations(ipConfigs *[]network.VirtualNetworkGatewayIPConfiguration) []interface{} {
	flat := make([]interface{}, 0)

	if ipConfigs != nil {
		for _, cfg := range *ipConfigs {
			props := cfg.VirtualNetworkGatewayIPConfigurationPropertiesFormat
			v := make(map[string]interface{})

			if id := cfg.ID; id != nil {
				v["id"] = *id
			}

			if name := cfg.Name; name != nil {
				v["name"] = *name
			}
			v["private_ip_address_allocation"] = string(props.PrivateIPAllocationMethod)

			if subnet := props.Subnet; subnet != nil {
				if id := subnet.ID; id != nil {
					v["subnet_id"] = *id
				}
			}

			if pip := props.PublicIPAddress; pip != nil {
				if id := pip.ID; id != nil {
					v["public_ip_address_id"] = *id
				}
			}

			flat = append(flat, v)
		}
	}

	return flat
}

func flattenVirtualNetworkGatewayDataSourceVpnClientConfig(cfg *network.VpnClientConfiguration) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	flat := make(map[string]interface{})

	if pool := cfg.VpnClientAddressPool; pool != nil {
		flat["address_space"] = utils.FlattenStringSlice(pool.AddressPrefixes)
	} else {
		flat["address_space"] = []interface{}{}
	}

	rootCerts := make([]interface{}, 0)
	if certs := cfg.VpnClientRootCertificates; certs != nil {
		for _, cert := range *certs {
			if cert.Name == nil || cert.VpnClientRootCertificatePropertiesFormat == nil || cert.VpnClientRootCertificatePropertiesFormat.PublicCertData == nil {
				continue
			}
			v := map[string]interface{}{
				"name":             *cert.Name,
				"public_cert_data": *cert.VpnClientRootCertificatePropertiesFormat.PublicCertData,
			}
			rootCerts = append(rootCerts, v)
		}
	}
	flat["root_certificate"] = rootCerts

	revokedCerts := make([]interface{}, 0)
	if certs := cfg.VpnClientRevokedCertificates; certs != nil {
		for _, cert := range *certs {
			if cert.Name == nil || cert.VpnClientRevokedCertificatePropertiesFormat == nil || cert.VpnClientRevokedCertificatePropertiesFormat.Thumbprint == nil {
				continue
			}
			v := map[string]interface{}{
				"name":       *cert.Name,
				"thumbprint": *cert.VpnClientRevokedCertificatePropertiesFormat.Thumbprint,
			}
			revokedCerts = append(revokedCerts, v)
		}
	}
	flat["revoked_certificate"] = revokedCerts

	vpnClientProtocols := make([]interface{}, 0)
	if vpnProtocols := cfg.VpnClientProtocols; vpnProtocols != nil {
		for _, protocol := range *vpnProtocols {
			vpnClientProtocols = append(vpnClientProtocols, string(protocol))
		}
	}
	flat["vpn_client_protocols"] = vpnClientProtocols

	if v := cfg.RadiusServerAddress; v != nil {
		flat["radius_server_address"] = *v
	}

	if v := cfg.RadiusServerSecret; v != nil {
		flat["radius_server_secret"] = *v
	}

	return []interface{}{flat}
}

func flattenVirtualNetworkGatewayDataSourceBgpSettings(settings *network.BgpSettings) []interface{} {
	output := make([]interface{}, 0)

	if settings != nil {
		flat := make(map[string]interface{})

		if asn := settings.Asn; asn != nil {
			flat["asn"] = int(*asn)
		}
		if address := settings.BgpPeeringAddress; address != nil {
			flat["peering_address"] = *address
		}
		if weight := settings.PeerWeight; weight != nil {
			flat["peer_weight"] = int(*weight)
		}

		output = append(output, flat)
	}

	return output
}
