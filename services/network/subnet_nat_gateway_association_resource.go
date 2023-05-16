package network

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/locks"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceSubnetNatGatewayAssociation() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceSubnetNatGatewayAssociationCreate,
		Read:   resourceSubnetNatGatewayAssociationRead,
		Delete: resourceSubnetNatGatewayAssociationDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.SubnetID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"subnet_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"nat_gateway_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},
		},
	}
}

func resourceSubnetNatGatewayAssociationCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SubnetsClient
	vnetClient := meta.(*clients.Client).Network.VnetClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Subnet <-> NAT Gateway Association creation.")
	natGatewayId := d.Get("nat_gateway_id").(string)
	parsedSubnetId, err := parse.SubnetID(d.Get("subnet_id").(string))
	if err != nil {
		return err
	}

	parsedGatewayId, err := parse.NatGatewayID(d.Get("nat_gateway_id").(string))
	if err != nil {
		return fmt.Errorf("parsing NAT gateway id '%s': %+v", natGatewayId, err)
	}

	locks.ByName(parsedGatewayId.Name, natGatewayResourceName)
	defer locks.UnlockByName(parsedGatewayId.Name, natGatewayResourceName)
	locks.ByName(parsedSubnetId.VirtualNetworkName, VirtualNetworkResourceName)
	defer locks.UnlockByName(parsedSubnetId.VirtualNetworkName, VirtualNetworkResourceName)
	locks.ByName(parsedSubnetId.Name, SubnetResourceName)
	defer locks.UnlockByName(parsedSubnetId.Name, SubnetResourceName)

	subnet, err := client.Get(ctx, parsedSubnetId.ResourceGroup, parsedSubnetId.VirtualNetworkName, parsedSubnetId.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(subnet.Response) {
			return fmt.Errorf("%s was not found!", *parsedSubnetId)
		}
		return fmt.Errorf("retrieving %s: %+v", *parsedSubnetId, err)
	}

	if props := subnet.SubnetPropertiesFormat; props != nil {
		// check if the resources are imported
		if gateway := props.NatGateway; gateway != nil {
			if gateway.ID != nil && subnet.ID != nil {
				return tf.ImportAsExistsError("azurerm_subnet_nat_gateway_association", *subnet.ID)
			}
		}
		props.NatGateway = &network.SubResource{
			ID: utils.String(natGatewayId),
		}
	}

	future, err := client.CreateOrUpdate(ctx, parsedSubnetId.ResourceGroup, parsedSubnetId.VirtualNetworkName, parsedSubnetId.Name, subnet)
	if err != nil {
		return fmt.Errorf("updating NAT Gateway Association for %s: %+v", *parsedSubnetId, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for completion of NAT Gateway Association for %s: %+v", *parsedSubnetId, err)
	}

	timeout, _ := ctx.Deadline()

	stateConf := &pluginsdk.StateChangeConf{
		Pending:    []string{string(network.ProvisioningStateUpdating)},
		Target:     []string{string(network.ProvisioningStateSucceeded)},
		Refresh:    SubnetProvisioningStateRefreshFunc(ctx, client, *parsedSubnetId),
		MinTimeout: 1 * time.Minute,
		Timeout:    time.Until(timeout),
	}
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for provisioning state of subnet for NAT Gateway Association for %s: %+v", *parsedSubnetId, err)
	}

	vnetId := parse.NewVirtualNetworkID(parsedSubnetId.SubscriptionId, parsedSubnetId.ResourceGroup, parsedSubnetId.VirtualNetworkName)
	vnetStateConf := &pluginsdk.StateChangeConf{
		Pending:    []string{string(network.ProvisioningStateUpdating)},
		Target:     []string{string(network.ProvisioningStateSucceeded)},
		Refresh:    VirtualNetworkProvisioningStateRefreshFunc(ctx, vnetClient, vnetId),
		MinTimeout: 1 * time.Minute,
		Timeout:    time.Until(timeout),
	}
	if _, err = vnetStateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for provisioning state of virtual network for NAT Gateway Association for %s: %+v", *parsedSubnetId, err)
	}

	d.SetId(parsedSubnetId.ID())

	return resourceSubnetNatGatewayAssociationRead(d, meta)
}

func resourceSubnetNatGatewayAssociationRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SubnetsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SubnetID(d.Id())
	if err != nil {
		return err
	}

	subnet, err := client.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(subnet.Response) {
			log.Printf("[DEBUG] %s could not be found - removing from state!", *id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	props := subnet.SubnetPropertiesFormat
	if props == nil {
		return fmt.Errorf("Error: `properties` was nil for %s", *id)
	}
	natGateway := props.NatGateway
	if natGateway == nil {
		log.Printf("[DEBUG] %s doesn't have a NAT Gateway - removing from state!", *id)
		d.SetId("")
		return nil
	}

	d.Set("subnet_id", subnet.ID)
	d.Set("nat_gateway_id", natGateway.ID)

	return nil
}

func resourceSubnetNatGatewayAssociationDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SubnetsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SubnetID(d.Id())
	if err != nil {
		return err
	}

	subnet, err := client.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(subnet.Response) {
			log.Printf("[DEBUG] %s could not be found - removing from state!", *id)
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	props := subnet.SubnetPropertiesFormat
	if props == nil {
		return fmt.Errorf("`Properties` was nil for %s ", *id)
	}
	if props.NatGateway == nil || props.NatGateway.ID == nil {
		log.Printf("[DEBUG] %s has no NAT Gateway - removing from state!", *id)
		return nil
	}
	parsedGatewayId, err := parse.NatGatewayID(*props.NatGateway.ID)
	if err != nil {
		return err
	}

	locks.ByName(parsedGatewayId.Name, natGatewayResourceName)
	defer locks.UnlockByName(parsedGatewayId.Name, natGatewayResourceName)
	locks.ByName(id.VirtualNetworkName, VirtualNetworkResourceName)
	defer locks.UnlockByName(id.VirtualNetworkName, VirtualNetworkResourceName)

	// ensure we get the latest state
	subnet, err = client.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(subnet.Response) {
			log.Printf("[DEBUG] %s could not be found - removing from state!", *id)
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	subnet.SubnetPropertiesFormat.NatGateway = nil // remove the nat gateway from subnet

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name, subnet)
	if err != nil {
		return fmt.Errorf("removing %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for removal of %s: %+v", *id, err)
	}

	return nil
}
