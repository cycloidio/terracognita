package streamanalytics

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/streamanalytics/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func dataSourceStreamAnalyticsJob() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceStreamAnalyticsJobRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"resource_group_name": commonschema.ResourceGroupNameForDataSource(),

			"location": commonschema.LocationComputed(),

			"compatibility_level": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"data_locale": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"events_late_arrival_max_delay_in_seconds": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"events_out_of_order_max_delay_in_seconds": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"events_out_of_order_policy": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"job_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"identity": commonschema.SystemAssignedIdentityComputed(),

			"last_output_time": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"output_error_policy": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"start_mode": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"start_time": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"streaming_units": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"transformation_query": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceStreamAnalyticsJobRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).StreamAnalytics.JobsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewStreamingJobID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	resp, err := client.Get(ctx, id.ResourceGroup, id.Name, "transformation")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("%s was not found", id)
		}

		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.SetId(id.ID())

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	if err := d.Set("identity", flattenJobIdentity(resp.Identity)); err != nil {
		return fmt.Errorf("setting `identity`: %v", err)
	}

	if props := resp.StreamingJobProperties; props != nil {
		d.Set("compatibility_level", string(props.CompatibilityLevel))
		d.Set("data_locale", props.DataLocale)
		if v := props.EventsLateArrivalMaxDelayInSeconds; v != nil {
			d.Set("events_late_arrival_max_delay_in_seconds", int(*v))
		}
		if v := props.EventsOutOfOrderMaxDelayInSeconds; v != nil {
			d.Set("events_out_of_order_max_delay_in_seconds", int(*v))
		}
		d.Set("events_out_of_order_policy", string(props.EventsOutOfOrderPolicy))
		d.Set("job_id", props.JobID)
		d.Set("output_error_policy", string(props.OutputErrorPolicy))

		if v := props.LastOutputEventTime; v != nil {
			d.Set("last_output_time", v.String())
		}

		if v := props.OutputStartTime; v != nil {
			d.Set("start_time", v.String())
		}

		d.Set("start_mode", props.OutputStartMode)

		if props.Transformation != nil && props.Transformation.TransformationProperties != nil {
			d.Set("streaming_units", props.Transformation.TransformationProperties.StreamingUnits)
			d.Set("transformation_query", props.Transformation.TransformationProperties.Query)
		}
	}

	return nil
}
