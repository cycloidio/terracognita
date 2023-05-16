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
	networkValidate "github.com/hashicorp/terraform-provider-azurerm/services/network/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceVirtualHubRouteTable() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceVirtualHubRouteTableCreateUpdate,
		Read:   resourceVirtualHubRouteTableRead,
		Update: resourceVirtualHubRouteTableCreateUpdate,
		Delete: resourceVirtualHubRouteTableDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.HubRouteTableID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: networkValidate.HubRouteTableName,
			},

			"virtual_hub_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: networkValidate.VirtualHubID,
			},

			"labels": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"route": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},

						"destinations": {
							Type:     pluginsdk.TypeSet,
							Required: true,
							Elem: &pluginsdk.Schema{
								Type:         pluginsdk.TypeString,
								ValidateFunc: validation.StringIsNotEmpty,
							},
						},

						"destinations_type": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"CIDR",
								"ResourceId",
								"Service",
							}, false),
						},

						"next_hop": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: azure.ValidateResourceID,
						},

						"next_hop_type": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							Default:  "ResourceId",
							ValidateFunc: validation.StringInSlice([]string{
								"ResourceId",
							}, false),
						},
					},
				},
			},
		},
	}
}

func resourceVirtualHubRouteTableCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.HubRouteTableClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	virtHubId, err := parse.VirtualHubID(d.Get("virtual_hub_id").(string))
	if err != nil {
		return err
	}

	locks.ByName(virtHubId.Name, virtualHubResourceName)
	defer locks.UnlockByName(virtHubId.Name, virtualHubResourceName)

	id := parse.NewHubRouteTableID(virtHubId.SubscriptionId, virtHubId.ResourceGroup, virtHubId.Name, d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.VirtualHubName, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_virtual_hub_route_table", id.ID())
		}
	}

	parameters := network.HubRouteTable{
		Name: utils.String(d.Get("name").(string)),
		HubRouteTableProperties: &network.HubRouteTableProperties{
			Labels: utils.ExpandStringSlice(d.Get("labels").(*pluginsdk.Set).List()),
			Routes: expandVirtualHubRouteTableHubRoutes(d.Get("route").(*pluginsdk.Set).List()),
		},
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.VirtualHubName, id.Name, parameters)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %+v", id, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting on creating/updating future for %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceVirtualHubRouteTableRead(d, meta)
}

func resourceVirtualHubRouteTableRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.HubRouteTableClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.HubRouteTableID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.VirtualHubName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Virtual Hub Route Table %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving HubRouteTable %q (Resource Group %q / Virtual Hub %q): %+v", id.Name, id.ResourceGroup, id.VirtualHubName, err)
	}

	d.Set("name", id.Name)
	d.Set("virtual_hub_id", parse.NewVirtualHubID(id.SubscriptionId, id.ResourceGroup, id.VirtualHubName).ID())

	if props := resp.HubRouteTableProperties; props != nil {
		d.Set("labels", utils.FlattenStringSlice(props.Labels))

		if err := d.Set("route", flattenVirtualHubRouteTableHubRoutes(props.Routes)); err != nil {
			return fmt.Errorf("setting `route`: %+v", err)
		}
	}
	return nil
}

func resourceVirtualHubRouteTableDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.HubRouteTableClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.HubRouteTableID(d.Id())
	if err != nil {
		return err
	}

	locks.ByName(id.VirtualHubName, virtualHubResourceName)
	defer locks.UnlockByName(id.VirtualHubName, virtualHubResourceName)

	future, err := client.Delete(ctx, id.ResourceGroup, id.VirtualHubName, id.Name)
	if err != nil {
		return fmt.Errorf("deleting HubRouteTable %q (Resource Group %q / Virtual Hub %q): %+v", id.Name, id.ResourceGroup, id.VirtualHubName, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting on deleting future for HubRouteTable %q (Resource Group %q / Virtual Hub %q): %+v", id.Name, id.ResourceGroup, id.VirtualHubName, err)
	}

	return nil
}

func expandVirtualHubRouteTableHubRoutes(input []interface{}) *[]network.HubRoute {
	results := make([]network.HubRoute, 0)

	for _, item := range input {
		v := item.(map[string]interface{})

		result := network.HubRoute{
			Name:            utils.String(v["name"].(string)),
			DestinationType: utils.String(v["destinations_type"].(string)),
			Destinations:    utils.ExpandStringSlice(v["destinations"].(*pluginsdk.Set).List()),
			NextHopType:     utils.String(v["next_hop_type"].(string)),
			NextHop:         utils.String(v["next_hop"].(string)),
		}

		results = append(results, result)
	}

	return &results
}

func flattenVirtualHubRouteTableHubRoutes(input *[]network.HubRoute) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	for _, item := range *input {
		var name string
		if item.Name != nil {
			name = *item.Name
		}

		var destinationType string
		if item.DestinationType != nil {
			destinationType = *item.DestinationType
		}

		var nextHop string
		if item.NextHop != nil {
			nextHop = *item.NextHop
		}

		var nextHopType string
		if item.NextHopType != nil {
			nextHopType = *item.NextHopType
		}

		v := map[string]interface{}{
			"name":              name,
			"destinations":      utils.FlattenStringSlice(item.Destinations),
			"destinations_type": destinationType,
			"next_hop":          nextHop,
			"next_hop_type":     nextHopType,
		}

		results = append(results, v)
	}

	return results
}
