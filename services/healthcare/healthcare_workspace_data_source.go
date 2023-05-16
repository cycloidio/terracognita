package healthcare

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/healthcare/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/healthcare/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func dataSourceHealthcareWorkspace() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceHealthcareWorkspaceRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.WorkspaceName,
			},

			"resource_group_name": commonschema.ResourceGroupName(),

			"location": commonschema.LocationComputed(),

			"tags": tags.SchemaDataSource(),
		},
	}
}

func dataSourceHealthcareWorkspaceRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).HealthCare.HealthcareWorkspaceClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewWorkspaceID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("%s was not found", id)
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.SetId(id.ID())

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)

	if locations := resp.Location; locations != nil {
		d.Set("location", location.Normalize(*locations))
	}

	return tags.FlattenAndSet(d, resp.Tags)
}
