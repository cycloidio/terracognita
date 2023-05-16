package network

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/locks"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceVirtualNetworkDnsServers() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceVirtualNetworkDnsServersCreateUpdate,
		Read:   resourceVirtualNetworkDnsServersRead,
		Update: resourceVirtualNetworkDnsServersCreateUpdate,
		Delete: resourceVirtualNetworkDnsServersDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.VirtualNetworkDnsServersID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"virtual_network_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.VirtualNetworkID,
			},

			"dns_servers": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
			},
		},
	}
}

func resourceVirtualNetworkDnsServersCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VnetClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	vnetId, err := parse.VirtualNetworkID(d.Get("virtual_network_id").(string))
	if err != nil {
		return err
	}

	// This is a virtual resource so the last segment is hardcoded
	id := parse.NewVirtualNetworkDnsServersID(vnetId.SubscriptionId, vnetId.ResourceGroup, vnetId.Name, "default")

	vnet, err := client.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, "")
	if err != nil {
		if utils.ResponseWasNotFound(vnet.Response) {
			return fmt.Errorf("%s could not be found: %s", vnetId, err)
		}
		return fmt.Errorf("reading %s: %s", vnetId, err)
	}

	locks.ByName(id.VirtualNetworkName, VirtualNetworkResourceName)
	defer locks.UnlockByName(id.VirtualNetworkName, VirtualNetworkResourceName)

	if vnet.VirtualNetworkPropertiesFormat == nil {
		return fmt.Errorf("%s was returned without any properties", vnetId)
	}
	if vnet.VirtualNetworkPropertiesFormat.DhcpOptions == nil {
		vnet.VirtualNetworkPropertiesFormat.DhcpOptions = &network.DhcpOptions{}
	}

	vnet.VirtualNetworkPropertiesFormat.DhcpOptions.DNSServers = utils.ExpandStringSlice(d.Get("dns_servers").([]interface{}))

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.VirtualNetworkName, vnet)
	if err != nil {
		return fmt.Errorf("updating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for update of %s: %+v", id, err)
	}

	timeout, _ := ctx.Deadline()

	vnetStateConf := &pluginsdk.StateChangeConf{
		Pending:    []string{string(network.ProvisioningStateUpdating)},
		Target:     []string{string(network.ProvisioningStateSucceeded)},
		Refresh:    VirtualNetworkProvisioningStateRefreshFunc(ctx, client, *vnetId),
		MinTimeout: 1 * time.Minute,
		Timeout:    time.Until(timeout),
	}
	if _, err = vnetStateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for provisioning state of virtual network for %s: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceVirtualNetworkDnsServersRead(d, meta)
}

func resourceVirtualNetworkDnsServersRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VnetClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VirtualNetworkDnsServersID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	vnetId := parse.NewVirtualNetworkID(id.SubscriptionId, id.ResourceGroup, id.VirtualNetworkName)
	d.Set("virtual_network_id", vnetId.ID())

	if props := resp.VirtualNetworkPropertiesFormat; props != nil {
		if err := d.Set("dns_servers", flattenVirtualNetworkDNSServers(props.DhcpOptions)); err != nil {
			return fmt.Errorf("setting `dns_servers`: %+v", err)
		}
	}

	return nil
}

func resourceVirtualNetworkDnsServersDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VnetClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VirtualNetworkDnsServersID(d.Id())
	if err != nil {
		return err
	}

	vnetId := parse.NewVirtualNetworkID(id.SubscriptionId, id.ResourceGroup, id.VirtualNetworkName)

	vnet, err := client.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, "")
	if err != nil {
		if utils.ResponseWasNotFound(vnet.Response) {
			log.Printf("[INFO] Virtual Network %q does not exist - removing %s from state", vnetId.ID(), id)
			return nil
		}
		return fmt.Errorf("reading %s: %s", vnetId, err)
	}

	locks.ByName(id.VirtualNetworkName, VirtualNetworkResourceName)
	defer locks.UnlockByName(id.VirtualNetworkName, VirtualNetworkResourceName)

	if vnet.VirtualNetworkPropertiesFormat == nil {
		return fmt.Errorf("%s was returned without any properties", vnetId)
	}
	if vnet.VirtualNetworkPropertiesFormat.DhcpOptions == nil {
		log.Printf("[INFO] dhcpOptions for %s was nil, dnsServers already deleted - removing %s from state", vnetId.ID(), id)
		return nil
	}

	vnet.VirtualNetworkPropertiesFormat.DhcpOptions.DNSServers = utils.ExpandStringSlice(make([]interface{}, 0))

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.VirtualNetworkName, vnet)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting to delete %s: %+v", id, err)
	}

	return nil
}
