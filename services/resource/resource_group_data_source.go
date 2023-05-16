package resource

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func dataSourceResourceGroup() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceResourceGroupRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name":     commonschema.ResourceGroupNameForDataSource(),
			"location": commonschema.LocationComputed(),
			"tags":     tags.SchemaDataSource(),
		},
	}
}

func dataSourceResourceGroupRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Resource.GroupsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resp, err := client.Get(ctx, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Error: Resource Group %q was not found", name)
		}
		return err
	}

	// @tombuildsstuff: intentionally leaving this for now, since this'll need
	// details in the upgrade notes given how the Resource Group ID is cased incorrectly
	// but needs to be fixed (resourcegroups -> resourceGroups)
	d.SetId(*resp.ID)

	d.Set("name", resp.Name)
	d.Set("location", location.NormalizeNilable(resp.Location))
	return tags.FlattenAndSet(d, resp.Tags)
}
