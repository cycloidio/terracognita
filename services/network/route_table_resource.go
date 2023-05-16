package network

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/go-azure-helpers/lang/response"
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

var routeTableResourceName = "azurerm_route_table"

func resourceRouteTable() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceRouteTableCreateUpdate,
		Read:   resourceRouteTableRead,
		Update: resourceRouteTableCreateUpdate,
		Delete: resourceRouteTableDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.RouteTableID(id)
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
				ValidateFunc: validate.RouteTableName,
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"route": {
				Type:       pluginsdk.TypeSet,
				ConfigMode: pluginsdk.SchemaConfigModeAttr,
				Optional:   true,
				Computed:   true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validate.RouteName,
						},

						"address_prefix": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},

						"next_hop_type": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(network.RouteNextHopTypeVirtualNetworkGateway),
								string(network.RouteNextHopTypeVnetLocal),
								string(network.RouteNextHopTypeInternet),
								string(network.RouteNextHopTypeVirtualAppliance),
								string(network.RouteNextHopTypeNone),
							}, false),
						},

						"next_hop_in_ip_address": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"disable_bgp_route_propagation": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			"subnets": {
				Type:     pluginsdk.TypeSet,
				Computed: true,
				Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
				Set:      pluginsdk.HashString,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceRouteTableCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.RouteTablesClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for AzureRM Route Table creation.")

	id := parse.NewRouteTableID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	location := azure.NormalizeLocation(d.Get("location").(string))
	t := d.Get("tags").(map[string]interface{})

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_route_table", id.ID())
		}
	}

	routeSet := network.RouteTable{
		Name:     &id.Name,
		Location: &location,
		RouteTablePropertiesFormat: &network.RouteTablePropertiesFormat{
			Routes:                     expandRouteTableRoutes(d),
			DisableBgpRoutePropagation: utils.Bool(d.Get("disable_bgp_route_propagation").(bool)),
		},
		Tags: tags.Expand(t),
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, routeSet)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for completion of %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceRouteTableRead(d, meta)
}

func resourceRouteTableRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.RouteTablesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.RouteTableID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Route Table %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if props := resp.RouteTablePropertiesFormat; props != nil {
		d.Set("disable_bgp_route_propagation", props.DisableBgpRoutePropagation)
		if err := d.Set("route", flattenRouteTableRoutes(props.Routes)); err != nil {
			return err
		}

		if err := d.Set("subnets", flattenRouteTableSubnets(props.Subnets)); err != nil {
			return err
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceRouteTableDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.RouteTablesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.RouteTableID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("deleting Route Table %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of Route Table %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	return nil
}

func expandRouteTableRoutes(d *pluginsdk.ResourceData) *[]network.Route {
	configs := d.Get("route").(*pluginsdk.Set).List()
	routes := make([]network.Route, 0, len(configs))

	for _, configRaw := range configs {
		data := configRaw.(map[string]interface{})

		route := network.Route{
			Name: utils.String(data["name"].(string)),
			RoutePropertiesFormat: &network.RoutePropertiesFormat{
				AddressPrefix: utils.String(data["address_prefix"].(string)),
				NextHopType:   network.RouteNextHopType(data["next_hop_type"].(string)),
			},
		}

		if v := data["next_hop_in_ip_address"].(string); v != "" {
			route.RoutePropertiesFormat.NextHopIPAddress = &v
		}

		routes = append(routes, route)
	}

	return &routes
}

func flattenRouteTableRoutes(input *[]network.Route) []interface{} {
	results := make([]interface{}, 0)

	if routes := input; routes != nil {
		for _, route := range *routes {
			r := make(map[string]interface{})

			r["name"] = *route.Name

			if props := route.RoutePropertiesFormat; props != nil {
				r["address_prefix"] = *props.AddressPrefix
				r["next_hop_type"] = string(props.NextHopType)
				if ip := props.NextHopIPAddress; ip != nil {
					r["next_hop_in_ip_address"] = *ip
				}
			}

			results = append(results, r)
		}
	}

	return results
}

func flattenRouteTableSubnets(subnets *[]network.Subnet) []string {
	output := make([]string, 0)

	if subnets != nil {
		for _, subnet := range *subnets {
			output = append(output, *subnet.ID)
		}
	}

	return output
}
