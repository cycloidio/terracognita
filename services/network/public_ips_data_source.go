package network

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
)

func dataSourcePublicIPs() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourcePublicIPsRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: dataSourcePublicIPSchema(),
	}
}

func dataSourcePublicIPSchema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"resource_group_name": commonschema.ResourceGroupNameForDataSource(),

		"name_prefix": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},

		"attachment_status": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			ValidateFunc: validation.StringInSlice([]string{
				"Attached",
				"Unattached",
			}, false),
		},

		"allocation_type": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			ValidateFunc: validation.StringInSlice([]string{
				string(network.IPAllocationMethodDynamic),
				string(network.IPAllocationMethodStatic),
			}, false),
		},

		"public_ips": {
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
					"fqdn": {
						Type:     pluginsdk.TypeString,
						Computed: true,
					},
					"domain_name_label": {
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
	}
}

func dataSourcePublicIPsRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.PublicIPsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceGroup := d.Get("resource_group_name").(string)

	log.Printf("[DEBUG] Reading Public IP's in Resource Group %q", resourceGroup)
	resp, err := client.List(ctx, resourceGroup)
	if err != nil {
		return fmt.Errorf("listing Public IP Addresses in the Resource Group %q: %v", resourceGroup, err)
	}

	filteredIPAddresses := make([]network.PublicIPAddress, 0)
	for _, element := range resp.Values() {
		nicIsAttached := element.IPConfiguration != nil || element.NatGateway != nil

		if prefix := d.Get("name_prefix").(string); prefix != "" {
			if !strings.HasPrefix(*element.Name, prefix) {
				continue
			}
		}

		attachmentStatus, attachmentStatusOk := d.GetOk("attachment_status")
		if attachmentStatusOk && attachmentStatus.(string) == "Attached" && !nicIsAttached {
			continue
		}
		if attachmentStatusOk && attachmentStatus.(string) == "Unattached" && nicIsAttached {
			continue
		}

		if allocationType := d.Get("allocation_type").(string); allocationType != "" {
			allocation := network.IPAllocationMethod(allocationType)
			if element.PublicIPAllocationMethod != allocation {
				continue
			}
		}

		filteredIPAddresses = append(filteredIPAddresses, element)
	}

	d.SetId(time.Now().UTC().String())

	results := flattenDataSourcePublicIPs(filteredIPAddresses)
	if err := d.Set("public_ips", results); err != nil {
		return fmt.Errorf("setting `public_ips`: %+v", err)
	}

	return nil
}

func flattenDataSourcePublicIPs(input []network.PublicIPAddress) []interface{} {
	results := make([]interface{}, 0)

	for _, element := range input {
		flattenedIPAddress := flattenDataSourcePublicIP(element)
		results = append(results, flattenedIPAddress)
	}

	return results
}

func flattenDataSourcePublicIP(input network.PublicIPAddress) map[string]string {
	id := ""
	if input.ID != nil {
		id = *input.ID
	}

	name := ""
	if input.Name != nil {
		name = *input.Name
	}

	domainNameLabel := ""
	fqdn := ""
	ipAddress := ""
	if props := input.PublicIPAddressPropertiesFormat; props != nil {
		if dns := props.DNSSettings; dns != nil {
			if dns.Fqdn != nil {
				fqdn = *dns.Fqdn
			}

			if dns.DomainNameLabel != nil {
				domainNameLabel = *dns.DomainNameLabel
			}
		}

		if props.IPAddress != nil {
			ipAddress = *props.IPAddress
		}
	}

	return map[string]string{
		"id":                id,
		"name":              name,
		"domain_name_label": domainNameLabel,
		"fqdn":              fqdn,
		"ip_address":        ipAddress,
	}
}
