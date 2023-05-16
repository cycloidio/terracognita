package iothub

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/iothub/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/iothub/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func dataSourceIotHubDPS() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceIotHubDPSRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.IoTHubName,
			},

			"resource_group_name": commonschema.ResourceGroupNameForDataSource(),

			"location": commonschema.LocationComputed(),

			"allocation_policy": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"device_provisioning_host_name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"id_scope": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"service_operations_host_name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"tags": tags.Schema(),
		},
	}
}

func dataSourceIotHubDPSRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).IoTHub.DPSResourceClient
	subscriptionId := meta.(*clients.Client).IoTHub.DPSResourceClient.SubscriptionID
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewIotHubDpsID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	resp, err := client.Get(ctx, id.ProvisioningServiceName, id.ResourceGroup)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Error: IoT Device Provisioning Service %s was not found", id)
		}

		return fmt.Errorf("retrieving IoT Device Provisioning Service %s: %+v", id, err)
	}

	d.Set("name", id.ProvisioningServiceName)
	d.Set("resource_group_name", id.ResourceGroup)
	d.SetId(id.ID())

	d.Set("location", location.NormalizeNilable(resp.Location))

	if props := resp.Properties; props != nil {
		d.Set("service_operations_host_name", props.ServiceOperationsHostName)
		d.Set("device_provisioning_host_name", props.DeviceProvisioningHostName)
		d.Set("id_scope", props.IDScope)
		d.Set("allocation_policy", props.AllocationPolicy)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}
