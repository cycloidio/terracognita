package network

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceVpnSite() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceVpnSiteCreateUpdate,
		Read:   resourceVpnSiteRead,
		Update: resourceVpnSiteCreateUpdate,
		Delete: resourceVpnSiteDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.VpnSiteID(id)
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
				ValidateFunc: validate.VpnSiteName(),
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": commonschema.Location(),

			"virtual_wan_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.VirtualWanID,
			},

			"address_cidrs": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeString,
					ValidateFunc: validation.IsCIDR,
				},
			},

			"device_vendor": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"device_model": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"link": {
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
						"provider_name": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"speed_in_mbps": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
							Default:      0,
						},
						"ip_address": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsIPAddress,
						},
						"fqdn": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"bgp": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"asn": {
										Type:         pluginsdk.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"peering_address": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.IsIPAddress,
									},
								},
							},
						},
						"id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceVpnSiteCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VpnSitesClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	location := azure.NormalizeLocation(d.Get("location").(string))

	if d.IsNewResource() {
		resp, err := client.Get(ctx, resourceGroup, name)
		if err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("checking for existing Vpn Site %q (Resource Group %q): %+v", name, resourceGroup, err)
			}
		}

		if resp.ID != nil && *resp.ID != "" {
			return tf.ImportAsExistsError("azurerm_vpn_site", *resp.ID)
		}
	}

	param := network.VpnSite{
		Name:     &name,
		Location: &location,
		VpnSiteProperties: &network.VpnSiteProperties{
			VirtualWan:       &network.SubResource{ID: utils.String(d.Get("virtual_wan_id").(string))},
			DeviceProperties: expandVpnSiteDeviceProperties(d),
			AddressSpace:     expandVpnSiteAddressSpace(d.Get("address_cidrs").(*pluginsdk.Set).List()),
			VpnSiteLinks:     expandVpnSiteLinks(d.Get("link").([]interface{})),
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	future, err := client.CreateOrUpdate(ctx, resourceGroup, name, param)
	if err != nil {
		return fmt.Errorf("creating  %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of  %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		return fmt.Errorf("retrieving Vpn Site %q (Resource Group %q): %+v", name, resourceGroup, err)
	}
	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("empty or nil ID returned for Vpn Site %q (Resource Group %q) ID", name, resourceGroup)
	}

	id, err := parse.VpnSiteID(*resp.ID)
	if err != nil {
		return err
	}
	d.SetId(id.ID())

	return resourceVpnSiteRead(d, meta)
}

func resourceVpnSiteRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VpnSitesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VpnSiteID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Vpn Site %q was not found in Resource Group %q - removing from state!", id.Name, id.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Vpn Site %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)

	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if prop := resp.VpnSiteProperties; prop != nil {
		if deviceProp := prop.DeviceProperties; deviceProp != nil {
			d.Set("device_vendor", deviceProp.DeviceVendor)
			d.Set("device_model", deviceProp.DeviceModel)
		}
		if prop.VirtualWan != nil {
			d.Set("virtual_wan_id", prop.VirtualWan.ID)
		}
		if err := d.Set("address_cidrs", flattenVpnSiteAddressSpace(prop.AddressSpace)); err != nil {
			return fmt.Errorf("setting `address_cidrs`")
		}
		if err := d.Set("link", flattenVpnSiteLinks(prop.VpnSiteLinks)); err != nil {
			return fmt.Errorf("setting `link`")
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceVpnSiteDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VpnSitesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VpnSiteID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting Vpn Site %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("waiting for deleting Vpn Site %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}
	}

	return nil
}

func expandVpnSiteDeviceProperties(d *pluginsdk.ResourceData) *network.DeviceProperties {
	vendor, model := d.Get("device_vendor").(string), d.Get("device_model").(string)
	if vendor == "" && model == "" {
		return nil
	}
	output := &network.DeviceProperties{}
	if vendor != "" {
		output.DeviceVendor = &vendor
	}
	if model != "" {
		output.DeviceModel = &model
	}

	return output
}

func expandVpnSiteAddressSpace(input []interface{}) *network.AddressSpace {
	if len(input) == 0 {
		return nil
	}

	addressPrefixes := []string{}
	for _, addr := range input {
		addressPrefixes = append(addressPrefixes, addr.(string))
	}

	return &network.AddressSpace{
		AddressPrefixes: &addressPrefixes,
	}
}

func flattenVpnSiteAddressSpace(input *network.AddressSpace) []interface{} {
	if input == nil {
		return nil
	}
	return utils.FlattenStringSlice(input.AddressPrefixes)
}

func expandVpnSiteLinks(input []interface{}) *[]network.VpnSiteLink {
	if len(input) == 0 {
		return nil
	}

	result := make([]network.VpnSiteLink, 0)
	for _, e := range input {
		if e == nil {
			continue
		}
		e := e.(map[string]interface{})
		link := network.VpnSiteLink{
			Name: utils.String(e["name"].(string)),
			VpnSiteLinkProperties: &network.VpnSiteLinkProperties{
				LinkProperties: &network.VpnLinkProviderProperties{
					LinkSpeedInMbps: utils.Int32(int32(e["speed_in_mbps"].(int))),
				},
			},
		}

		if v, ok := e["provider_name"]; ok {
			link.VpnSiteLinkProperties.LinkProperties.LinkProviderName = utils.String(v.(string))
		}
		if v, ok := e["ip_address"]; ok {
			link.VpnSiteLinkProperties.IPAddress = utils.String(v.(string))
		}
		if v, ok := e["fqdn"]; ok {
			link.VpnSiteLinkProperties.Fqdn = utils.String(v.(string))
		}
		if v, ok := e["bgp"]; ok {
			link.VpnSiteLinkProperties.BgpProperties = expandVpnSiteVpnLinkBgpSettings(v.([]interface{}))
		}

		result = append(result, link)
	}

	return &result
}

func flattenVpnSiteLinks(input *[]network.VpnSiteLink) []interface{} {
	if input == nil {
		return nil
	}

	output := make([]interface{}, 0)

	for _, e := range *input {
		var name string
		if e.Name != nil {
			name = *e.Name
		}

		var id string
		if e.ID != nil {
			id = *e.ID
		}

		var (
			ipAddress        string
			fqdn             string
			linkProviderName string
			linkSpeed        int
			bgpProperty      []interface{}
		)

		if prop := e.VpnSiteLinkProperties; prop != nil {
			if prop.IPAddress != nil {
				ipAddress = *prop.IPAddress
			}

			if prop.Fqdn != nil {
				fqdn = *prop.Fqdn
			}

			if linkProp := prop.LinkProperties; linkProp != nil {
				if linkProp.LinkProviderName != nil {
					linkProviderName = *linkProp.LinkProviderName
				}
				if linkProp.LinkSpeedInMbps != nil {
					linkSpeed = int(*linkProp.LinkSpeedInMbps)
				}
			}

			bgpProperty = flattenVpnSiteVpnSiteBgpSettings(prop.BgpProperties)
		}

		link := map[string]interface{}{
			"name":          name,
			"id":            id,
			"provider_name": linkProviderName,
			"speed_in_mbps": linkSpeed,
			"ip_address":    ipAddress,
			"fqdn":          fqdn,
			"bgp":           bgpProperty,
		}

		output = append(output, link)
	}

	return output
}

func expandVpnSiteVpnLinkBgpSettings(input []interface{}) *network.VpnLinkBgpSettings {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	v := input[0].(map[string]interface{})

	return &network.VpnLinkBgpSettings{
		Asn:               utils.Int64(int64(v["asn"].(int))),
		BgpPeeringAddress: utils.String(v["peering_address"].(string)),
	}
}

func flattenVpnSiteVpnSiteBgpSettings(input *network.VpnLinkBgpSettings) []interface{} {
	if input == nil {
		return nil
	}

	var asn int
	if input.Asn != nil {
		asn = int(*input.Asn)
	}

	var peerAddress string
	if input.BgpPeeringAddress != nil {
		peerAddress = *input.BgpPeeringAddress
	}

	return []interface{}{
		map[string]interface{}{
			"asn":             asn,
			"peering_address": peerAddress,
		},
	}
}
