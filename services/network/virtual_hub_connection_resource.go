package network

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
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

func resourceVirtualHubConnection() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceVirtualHubConnectionCreateOrUpdate,
		Read:   resourceVirtualHubConnectionRead,
		Update: resourceVirtualHubConnectionCreateOrUpdate,
		Delete: resourceVirtualHubConnectionDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(60 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(60 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(60 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.HubVirtualNetworkConnectionID(id)
			return err
		}),

		Schema: resourceVirtualHubConnectionSchema(),
	}
}

func resourceVirtualHubConnectionSchema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.VirtualHubConnectionName,
		},

		"virtual_hub_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.VirtualHubID,
		},

		"remote_virtual_network_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.VirtualNetworkID,
		},

		"internet_security_enabled": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Default:  false,
		},

		"routing": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"associated_route_table_id": {
						Type:         pluginsdk.TypeString,
						Optional:     true,
						Computed:     true,
						ValidateFunc: validate.HubRouteTableID,
						AtLeastOneOf: []string{"routing.0.associated_route_table_id", "routing.0.propagated_route_table", "routing.0.static_vnet_route"},
					},

					"propagated_route_table": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						Computed: true,
						MaxItems: 1,
						Elem: &pluginsdk.Resource{
							Schema: map[string]*pluginsdk.Schema{
								"labels": {
									Type:     pluginsdk.TypeSet,
									Optional: true,
									Computed: true,
									Elem: &pluginsdk.Schema{
										Type:         pluginsdk.TypeString,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									AtLeastOneOf: []string{"routing.0.propagated_route_table.0.labels", "routing.0.propagated_route_table.0.route_table_ids"},
								},

								"route_table_ids": {
									Type:     pluginsdk.TypeList,
									Optional: true,
									Computed: true,
									Elem: &pluginsdk.Schema{
										Type:         pluginsdk.TypeString,
										ValidateFunc: validate.HubRouteTableID,
									},
									AtLeastOneOf: []string{"routing.0.propagated_route_table.0.labels", "routing.0.propagated_route_table.0.route_table_ids"},
								},
							},
						},
						AtLeastOneOf: []string{"routing.0.associated_route_table_id", "routing.0.propagated_route_table", "routing.0.static_vnet_route"},
					},

					//lintignore:XS003
					"static_vnet_route": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						Elem: &pluginsdk.Resource{
							Schema: map[string]*pluginsdk.Schema{
								"name": {
									Type:         pluginsdk.TypeString,
									Optional:     true,
									ValidateFunc: validation.StringIsNotEmpty,
								},

								"address_prefixes": {
									Type:     pluginsdk.TypeSet,
									Optional: true,
									Elem: &pluginsdk.Schema{
										Type:         pluginsdk.TypeString,
										ValidateFunc: validation.IsCIDR,
									},
								},

								"next_hop_ip_address": {
									Type:         pluginsdk.TypeString,
									Optional:     true,
									ValidateFunc: validation.IsIPv4Address,
								},
							},
						},
						AtLeastOneOf: []string{"routing.0.associated_route_table_id", "routing.0.propagated_route_table", "routing.0.static_vnet_route"},
					},
				},
			},
		},
	}
}

func resourceVirtualHubConnectionCreateOrUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.HubVirtualNetworkConnectionClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	virtualHubId, err := parse.VirtualHubID(d.Get("virtual_hub_id").(string))
	if err != nil {
		return err
	}

	id := parse.NewHubVirtualNetworkConnectionID(virtualHubId.SubscriptionId, virtualHubId.ResourceGroup, virtualHubId.Name, d.Get("name").(string))

	locks.ByName(virtualHubId.Name, virtualHubResourceName)
	defer locks.UnlockByName(virtualHubId.Name, virtualHubResourceName)

	remoteVirtualNetworkId, err := parse.VirtualNetworkID(d.Get("remote_virtual_network_id").(string))
	if err != nil {
		return err
	}

	locks.ByName(remoteVirtualNetworkId.Name, VirtualNetworkResourceName)
	defer locks.UnlockByName(remoteVirtualNetworkId.Name, VirtualNetworkResourceName)

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.VirtualHubName, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of %s: %+v", id, err)
			}
		}
		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_virtual_hub_connection", id.ID())
		}
	}

	connection := network.HubVirtualNetworkConnection{
		Name: utils.String(id.Name),
		HubVirtualNetworkConnectionProperties: &network.HubVirtualNetworkConnectionProperties{
			RemoteVirtualNetwork: &network.SubResource{
				ID: utils.String(remoteVirtualNetworkId.ID()),
			},
			EnableInternetSecurity: utils.Bool(d.Get("internet_security_enabled").(bool)),
		},
	}

	if v, ok := d.GetOk("routing"); ok {
		connection.HubVirtualNetworkConnectionProperties.RoutingConfiguration = expandVirtualHubConnectionRouting(v.([]interface{}))
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.VirtualHubName, id.Name, connection)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of %s: %+v", id, err)
	}

	timeout, _ := ctx.Deadline()

	vnetStateConf := &pluginsdk.StateChangeConf{
		Pending:    []string{string(network.ProvisioningStateUpdating)},
		Target:     []string{string(network.ProvisioningStateSucceeded)},
		Refresh:    virtualHubConnectionProvisioningStateRefreshFunc(ctx, client, id),
		MinTimeout: 1 * time.Minute,
		Timeout:    time.Until(timeout),
	}
	if _, err = vnetStateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for provisioning state of %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceVirtualHubConnectionRead(d, meta)
}

func resourceVirtualHubConnectionRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.HubVirtualNetworkConnectionClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.HubVirtualNetworkConnectionID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.VirtualHubName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] %s does not exist - removing from state", *id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("reading %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("virtual_hub_id", parse.NewVirtualHubID(id.SubscriptionId, id.ResourceGroup, id.VirtualHubName).ID())

	if props := resp.HubVirtualNetworkConnectionProperties; props != nil {
		d.Set("internet_security_enabled", props.EnableInternetSecurity)
		remoteVirtualNetworkId := ""
		if props.RemoteVirtualNetwork != nil && props.RemoteVirtualNetwork.ID != nil {
			remoteVirtualNetworkId = *props.RemoteVirtualNetwork.ID
		}
		d.Set("remote_virtual_network_id", remoteVirtualNetworkId)

		if err := d.Set("routing", flattenVirtualHubConnectionRouting(props.RoutingConfiguration)); err != nil {
			return fmt.Errorf("setting `routing`: %+v", err)
		}
	}

	return nil
}

func resourceVirtualHubConnectionDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.HubVirtualNetworkConnectionClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.HubVirtualNetworkConnectionID(d.Id())
	if err != nil {
		return err
	}

	locks.ByName(id.VirtualHubName, virtualHubResourceName)
	defer locks.UnlockByName(id.VirtualHubName, virtualHubResourceName)

	future, err := client.Delete(ctx, id.ResourceGroup, id.VirtualHubName, id.Name)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
	}

	return nil
}

func expandVirtualHubConnectionRouting(input []interface{}) *network.RoutingConfiguration {
	if len(input) == 0 {
		return &network.RoutingConfiguration{}
	}

	v := input[0].(map[string]interface{})
	result := network.RoutingConfiguration{}

	if associatedRouteTableId := v["associated_route_table_id"].(string); associatedRouteTableId != "" {
		result.AssociatedRouteTable = &network.SubResource{
			ID: utils.String(associatedRouteTableId),
		}
	}

	if vnetStaticRoute := v["static_vnet_route"].([]interface{}); len(vnetStaticRoute) != 0 {
		result.VnetRoutes = expandVirtualHubConnectionVnetStaticRoute(vnetStaticRoute)
	}

	if propagatedRouteTable := v["propagated_route_table"].([]interface{}); len(propagatedRouteTable) != 0 {
		result.PropagatedRouteTables = expandVirtualHubConnectionPropagatedRouteTable(propagatedRouteTable)
	}

	return &result
}

func expandVirtualHubConnectionPropagatedRouteTable(input []interface{}) *network.PropagatedRouteTable {
	if len(input) == 0 {
		return &network.PropagatedRouteTable{}
	}

	v := input[0].(map[string]interface{})

	result := network.PropagatedRouteTable{}

	if labels := v["labels"].(*pluginsdk.Set).List(); len(labels) != 0 {
		result.Labels = utils.ExpandStringSlice(labels)
	}

	if routeTableIds := v["route_table_ids"].([]interface{}); len(routeTableIds) != 0 {
		result.Ids = expandIDsToSubResources(routeTableIds)
	}

	return &result
}

func expandVirtualHubConnectionVnetStaticRoute(input []interface{}) *network.VnetRoute {
	if len(input) == 0 {
		return &network.VnetRoute{}
	}

	results := make([]network.StaticRoute, 0)

	for _, item := range input {
		if item == nil {
			continue
		}

		v := item.(map[string]interface{})

		result := network.StaticRoute{}

		if name := v["name"].(string); name != "" {
			result.Name = utils.String(name)
		}

		if addressPrefixes := v["address_prefixes"].(*pluginsdk.Set).List(); len(addressPrefixes) != 0 {
			result.AddressPrefixes = utils.ExpandStringSlice(addressPrefixes)
		}

		if nextHopIPAddress := v["next_hop_ip_address"].(string); nextHopIPAddress != "" {
			result.NextHopIPAddress = utils.String(nextHopIPAddress)
		}

		results = append(results, result)
	}

	return &network.VnetRoute{
		StaticRoutes: &results,
	}
}

func expandIDsToSubResources(input []interface{}) *[]network.SubResource {
	ids := make([]network.SubResource, 0)

	for _, v := range input {
		ids = append(ids, network.SubResource{
			ID: utils.String(v.(string)),
		})
	}

	return &ids
}

func flattenVirtualHubConnectionRouting(input *network.RoutingConfiguration) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	associatedRouteTableId := ""
	if input.AssociatedRouteTable != nil && input.AssociatedRouteTable.ID != nil {
		associatedRouteTableId = *input.AssociatedRouteTable.ID
	}

	return []interface{}{
		map[string]interface{}{
			"associated_route_table_id": associatedRouteTableId,
			"propagated_route_table":    flattenVirtualHubConnectionPropagatedRouteTable(input.PropagatedRouteTables),
			"static_vnet_route":         flattenVirtualHubConnectionVnetStaticRoute(input.VnetRoutes),
		},
	}
}

func flattenVirtualHubConnectionPropagatedRouteTable(input *network.PropagatedRouteTable) []interface{} {
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

func flattenVirtualHubConnectionVnetStaticRoute(input *network.VnetRoute) []interface{} {
	results := make([]interface{}, 0)
	if input == nil || input.StaticRoutes == nil {
		return results
	}

	for _, item := range *input.StaticRoutes {
		var name string
		if item.Name != nil {
			name = *item.Name
		}

		var nextHopIpAddress string
		if item.NextHopIPAddress != nil {
			nextHopIpAddress = *item.NextHopIPAddress
		}

		addressPrefixes := make([]interface{}, 0)
		if item.AddressPrefixes != nil {
			addressPrefixes = utils.FlattenStringSlice(item.AddressPrefixes)
		}

		v := map[string]interface{}{
			"name":                name,
			"address_prefixes":    addressPrefixes,
			"next_hop_ip_address": nextHopIpAddress,
		}

		results = append(results, v)
	}

	return results
}

func flattenSubResourcesToIDs(input *[]network.SubResource) []interface{} {
	ids := make([]interface{}, 0)
	if input == nil {
		return ids
	}

	for _, v := range *input {
		if v.ID == nil {
			continue
		}

		ids = append(ids, *v.ID)
	}

	return ids
}

func virtualHubConnectionProvisioningStateRefreshFunc(ctx context.Context, client *network.HubVirtualNetworkConnectionsClient, id parse.HubVirtualNetworkConnectionId) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		res, err := client.Get(ctx, id.ResourceGroup, id.VirtualHubName, id.Name)
		if err != nil {
			return nil, "", fmt.Errorf("polling for %s: %+v", id, err)
		}

		return res, string(res.ProvisioningState), nil
	}
}
