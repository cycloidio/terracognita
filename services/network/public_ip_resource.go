package network

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/zones"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func resourcePublicIp() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourcePublicIpCreateUpdate,
		Read:   resourcePublicIpRead,
		Update: resourcePublicIpCreateUpdate,
		Delete: resourcePublicIpDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.PublicIpAddressID(id)
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

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"allocation_method": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.IPAllocationMethodStatic),
					string(network.IPAllocationMethodDynamic),
				}, false),
			},

			// Optional
			"edge_zone": commonschema.EdgeZoneOptionalForceNew(),

			"ip_version": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  string(network.IPVersionIPv4),
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.IPVersionIPv4),
					string(network.IPVersionIPv6),
				}, false),
			},

			"sku": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  string(network.PublicIPAddressSkuNameBasic),
				ValidateFunc: validation.StringInSlice([]string{
					string(network.PublicIPAddressSkuNameBasic),
					string(network.PublicIPAddressSkuNameStandard),
				}, false),
			},

			"sku_tier": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  string(network.PublicIPAddressSkuTierRegional),
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.PublicIPAddressSkuTierGlobal),
					string(network.PublicIPAddressSkuTierRegional),
				}, false),
			},

			"idle_timeout_in_minutes": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Default:      4,
				ValidateFunc: validation.IntBetween(4, 30),
			},

			"domain_name_label": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validate.PublicIpDomainNameLabel,
			},

			"fqdn": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"reverse_fqdn": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"ip_address": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"public_ip_prefix_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"ip_tags": {
				Type:     pluginsdk.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"zones": commonschema.ZonesMultipleOptionalForceNew(),

			"tags": tags.Schema(),
		},
	}
}

func resourcePublicIpCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.PublicIPsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for AzureRM Public IP creation.")

	id := parse.NewPublicIpAddressID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_public_ip", id.ID())
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	sku := d.Get("sku").(string)
	sku_tier := d.Get("sku_tier").(string)
	t := d.Get("tags").(map[string]interface{})

	idleTimeout := d.Get("idle_timeout_in_minutes").(int)
	ipVersion := network.IPVersion(d.Get("ip_version").(string))
	ipAllocationMethod := d.Get("allocation_method").(string)

	if strings.EqualFold(sku, "standard") {
		if !strings.EqualFold(ipAllocationMethod, "static") {
			return fmt.Errorf("Static IP allocation must be used when creating Standard SKU public IP addresses.")
		}
	}

	publicIp := network.PublicIPAddress{
		Name:             utils.String(id.Name),
		ExtendedLocation: expandEdgeZone(d.Get("edge_zone").(string)),
		Location:         &location,
		Sku: &network.PublicIPAddressSku{
			Name: network.PublicIPAddressSkuName(sku),
			Tier: network.PublicIPAddressSkuTier(sku_tier),
		},
		PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: network.IPAllocationMethod(ipAllocationMethod),
			PublicIPAddressVersion:   ipVersion,
			IdleTimeoutInMinutes:     utils.Int32(int32(idleTimeout)),
		},
		Tags: tags.Expand(t),
	}

	zones := zones.Expand(d.Get("zones").(*schema.Set).List())
	if len(zones) > 0 {
		publicIp.Zones = &zones
	}

	publicIpPrefixId, publicIpPrefixIdOk := d.GetOk("public_ip_prefix_id")

	if publicIpPrefixIdOk {
		publicIpPrefix := network.SubResource{}
		publicIpPrefix.ID = utils.String(publicIpPrefixId.(string))
		publicIp.PublicIPAddressPropertiesFormat.PublicIPPrefix = &publicIpPrefix
	}

	dnl, dnlOk := d.GetOk("domain_name_label")
	rfqdn, rfqdnOk := d.GetOk("reverse_fqdn")

	if dnlOk || rfqdnOk {
		dnsSettings := network.PublicIPAddressDNSSettings{}

		if rfqdnOk {
			dnsSettings.ReverseFqdn = utils.String(rfqdn.(string))
		}

		if dnlOk {
			dnsSettings.DomainNameLabel = utils.String(dnl.(string))
		}

		publicIp.PublicIPAddressPropertiesFormat.DNSSettings = &dnsSettings
	}

	if v, ok := d.GetOk("ip_tags"); ok {
		ipTags := v.(map[string]interface{})
		newIpTags := []network.IPTag{}

		for key, val := range ipTags {
			ipTag := network.IPTag{
				IPTagType: utils.String(key),
				Tag:       utils.String(val.(string)),
			}
			newIpTags = append(newIpTags, ipTag)
		}

		publicIp.PublicIPAddressPropertiesFormat.IPTags = &newIpTags
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, publicIp)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation/update of %s: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourcePublicIpRead(d, meta)
}

func resourcePublicIpRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.PublicIPsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.PublicIpAddressID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))
	d.Set("edge_zone", flattenEdgeZone(resp.ExtendedLocation))
	d.Set("zones", zones.Flatten(resp.Zones))

	if sku := resp.Sku; sku != nil {
		d.Set("sku", string(sku.Name))
		d.Set("sku_tier", string(sku.Tier))
	}

	if props := resp.PublicIPAddressPropertiesFormat; props != nil {
		d.Set("allocation_method", string(props.PublicIPAllocationMethod))
		d.Set("ip_version", string(props.PublicIPAddressVersion))

		if publicIpPrefix := props.PublicIPPrefix; publicIpPrefix != nil {
			d.Set("public_ip_prefix_id", publicIpPrefix.ID)
		}

		if settings := props.DNSSettings; settings != nil {
			d.Set("fqdn", settings.Fqdn)
			d.Set("reverse_fqdn", settings.ReverseFqdn)
			d.Set("domain_name_label", settings.DomainNameLabel)
		}

		d.Set("ip_tags", flattenPublicIpPropsIpTags(props.IPTags))

		d.Set("ip_address", props.IPAddress)
		d.Set("idle_timeout_in_minutes", props.IdleTimeoutInMinutes)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourcePublicIpDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.PublicIPsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.PublicIpAddressID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
	}

	return nil
}

func flattenPublicIpPropsIpTags(input *[]network.IPTag) map[string]interface{} {
	out := make(map[string]interface{})

	if input != nil {
		for _, tag := range *input {
			if tag.IPTagType != nil {
				out[*tag.IPTagType] = tag.Tag
			}
		}
	}

	return out
}
