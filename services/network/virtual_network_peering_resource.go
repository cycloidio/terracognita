package network

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

// peerMutex is used to prevent multiple Peering resources being created, updated
// or deleted at the same time
var peerMutex = &sync.Mutex{}

func resourceVirtualNetworkPeering() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceVirtualNetworkPeeringCreateUpdate,
		Read:   resourceVirtualNetworkPeeringRead,
		Update: resourceVirtualNetworkPeeringCreateUpdate,
		Delete: resourceVirtualNetworkPeeringDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.VirtualNetworkPeeringID(id)
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

			"virtual_network_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"remote_virtual_network_id": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"allow_virtual_network_access": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"allow_forwarded_traffic": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Computed: true,
			},

			"allow_gateway_transit": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Computed: true,
			},

			"use_remote_gateways": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceVirtualNetworkPeeringCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VnetPeeringsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Azure ARM virtual network peering creation.")

	id := parse.NewVirtualNetworkPeeringID(subscriptionId, d.Get("resource_group_name").(string), d.Get("virtual_network_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %s", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_virtual_network_peering", id.ID())
		}
	}

	peer := network.VirtualNetworkPeering{
		Name:                                  &id.Name,
		VirtualNetworkPeeringPropertiesFormat: getVirtualNetworkPeeringProperties(d),
	}

	peerMutex.Lock()
	defer peerMutex.Unlock()

	if err := pluginsdk.Retry(300*time.Second, retryVnetPeeringsClientCreateUpdate(d, id.ResourceGroup, id.VirtualNetworkName, id.Name, peer, meta)); err != nil {
		return err
	}

	d.SetId(id.ID())

	return resourceVirtualNetworkPeeringRead(d, meta)
}

func resourceVirtualNetworkPeeringRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VnetPeeringsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VirtualNetworkPeeringID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("making Read request on %s: %+v", *id, err)
	}

	// update appropriate values
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("name", id.Name)
	d.Set("virtual_network_name", id.VirtualNetworkName)

	if peer := resp.VirtualNetworkPeeringPropertiesFormat; peer != nil {
		d.Set("allow_virtual_network_access", peer.AllowVirtualNetworkAccess)
		d.Set("allow_forwarded_traffic", peer.AllowForwardedTraffic)
		d.Set("allow_gateway_transit", peer.AllowGatewayTransit)
		d.Set("use_remote_gateways", peer.UseRemoteGateways)
		if network := peer.RemoteVirtualNetwork; network != nil {
			d.Set("remote_virtual_network_id", network.ID)
		}
	}

	return nil
}

func resourceVirtualNetworkPeeringDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VnetPeeringsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VirtualNetworkPeeringID(d.Id())
	if err != nil {
		return err
	}

	peerMutex.Lock()
	defer peerMutex.Unlock()

	future, err := client.Delete(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
	}

	return err
}

func getVirtualNetworkPeeringProperties(d *pluginsdk.ResourceData) *network.VirtualNetworkPeeringPropertiesFormat {
	allowVirtualNetworkAccess := d.Get("allow_virtual_network_access").(bool)
	allowForwardedTraffic := d.Get("allow_forwarded_traffic").(bool)
	allowGatewayTransit := d.Get("allow_gateway_transit").(bool)
	useRemoteGateways := d.Get("use_remote_gateways").(bool)
	remoteVirtualNetworkID := d.Get("remote_virtual_network_id").(string)

	return &network.VirtualNetworkPeeringPropertiesFormat{
		AllowVirtualNetworkAccess: &allowVirtualNetworkAccess,
		AllowForwardedTraffic:     &allowForwardedTraffic,
		AllowGatewayTransit:       &allowGatewayTransit,
		UseRemoteGateways:         &useRemoteGateways,
		RemoteVirtualNetwork: &network.SubResource{
			ID: &remoteVirtualNetworkID,
		},
	}
}

func retryVnetPeeringsClientCreateUpdate(d *pluginsdk.ResourceData, resGroup string, vnetName string, name string, peer network.VirtualNetworkPeering, meta interface{}) func() *pluginsdk.RetryError {
	return func() *pluginsdk.RetryError {
		vnetPeeringsClient := meta.(*clients.Client).Network.VnetPeeringsClient
		ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
		defer cancel()

		future, err := vnetPeeringsClient.CreateOrUpdate(ctx, resGroup, vnetName, name, peer, network.SyncRemoteAddressSpaceTrue)
		if err != nil {
			if utils.ResponseErrorIsRetryable(err) {
				return pluginsdk.RetryableError(err)
			} else if future.Response().StatusCode == 400 && strings.Contains(err.Error(), "ReferencedResourceNotProvisioned") {
				// Resource is not yet ready, this may be the case if the Vnet was just created or another peering was just initiated.
				return pluginsdk.RetryableError(err)
			}

			return pluginsdk.NonRetryableError(err)
		}

		if err = future.WaitForCompletionRef(ctx, vnetPeeringsClient.Client); err != nil {
			return pluginsdk.NonRetryableError(err)
		}

		return nil
	}
}
