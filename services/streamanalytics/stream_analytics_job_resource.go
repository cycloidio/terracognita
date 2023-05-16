package streamanalytics

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/streamanalytics/mgmt/2020-03-01/streamanalytics"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/identity"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/locks"
	"github.com/hashicorp/terraform-provider-azurerm/services/streamanalytics/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/streamanalytics/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceStreamAnalyticsJob() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceStreamAnalyticsJobCreateUpdate,
		Read:   resourceStreamAnalyticsJobRead,
		Update: resourceStreamAnalyticsJobCreateUpdate,
		Delete: resourceStreamAnalyticsJobDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.StreamingJobID(id)
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
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"stream_analytics_cluster_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"compatibility_level": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					// values found in the other API the portal uses
					string(streamanalytics.CompatibilityLevelOneFullStopZero),
					"1.1",
					"1.2",
				}, false),
			},

			"data_locale": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"events_late_arrival_max_delay_in_seconds": {
				// portal allows for up to 20d 23h 59m 59s
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(-1, 1814399),
				Default:      5,
			},

			"events_out_of_order_max_delay_in_seconds": {
				// portal allows for up to 9m 59s
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 599),
				Default:      0,
			},

			"events_out_of_order_policy": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(streamanalytics.EventsOutOfOrderPolicyAdjust),
					string(streamanalytics.EventsOutOfOrderPolicyDrop),
				}, false),
				Default: string(streamanalytics.EventsOutOfOrderPolicyAdjust),
			},

			"type": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(streamanalytics.JobTypeCloud),
					string(streamanalytics.JobTypeEdge),
				}, false),
				Default: string(streamanalytics.JobTypeCloud),
			},

			"output_error_policy": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(streamanalytics.OutputErrorPolicyDrop),
					string(streamanalytics.OutputErrorPolicyStop),
				}, false),
				Default: string(streamanalytics.OutputErrorPolicyDrop),
			},

			"streaming_units": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				ValidateFunc: validate.StreamAnalyticsJobStreamingUnits,
			},

			"transformation_query": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"identity": commonschema.SystemAssignedIdentityOptional(),

			"job_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceStreamAnalyticsJobCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).StreamAnalytics.JobsClient
	transformationsClient := meta.(*clients.Client).StreamAnalytics.TransformationsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Azure Stream Analytics Job creation.")

	id := parse.NewStreamingJobID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	locks.ByID(id.ID())
	defer locks.UnlockByID(id.ID())

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_stream_analytics_job", id.ID())
		}
	}

	compatibilityLevel := d.Get("compatibility_level").(string)
	eventsLateArrivalMaxDelayInSeconds := d.Get("events_late_arrival_max_delay_in_seconds").(int)
	eventsOutOfOrderMaxDelayInSeconds := d.Get("events_out_of_order_max_delay_in_seconds").(int)
	eventsOutOfOrderPolicy := d.Get("events_out_of_order_policy").(string)
	jobType := d.Get("type").(string)
	location := azure.NormalizeLocation(d.Get("location").(string))
	outputErrorPolicy := d.Get("output_error_policy").(string)
	transformationQuery := d.Get("transformation_query").(string)
	t := d.Get("tags").(map[string]interface{})

	// needs to be defined inline for a Create but via a separate API for Update
	transformation := streamanalytics.Transformation{
		Name: utils.String("main"),
		TransformationProperties: &streamanalytics.TransformationProperties{
			Query: utils.String(transformationQuery),
		},
	}

	if jobType == string(streamanalytics.JobTypeEdge) {
		if _, ok := d.GetOk("streaming_units"); ok {
			return fmt.Errorf("the job type `Edge` doesn't support `streaming_units`")
		}
	} else {
		if v, ok := d.GetOk("streaming_units"); ok {
			transformation.TransformationProperties.StreamingUnits = utils.Int32(int32(v.(int)))
		} else {
			return fmt.Errorf("`streaming_units` must be set when `type` is `Cloud`")
		}
	}

	expandedIdentity, err := expandStreamAnalyticsJobIdentity(d.Get("identity").([]interface{}))
	if err != nil {
		return fmt.Errorf("expanding `identity`: %+v", err)
	}

	props := streamanalytics.StreamingJob{
		Name:     utils.String(id.Name),
		Location: utils.String(location),
		StreamingJobProperties: &streamanalytics.StreamingJobProperties{
			Sku: &streamanalytics.Sku{
				Name: streamanalytics.SkuNameStandard,
			},
			CompatibilityLevel:                 streamanalytics.CompatibilityLevel(compatibilityLevel),
			EventsLateArrivalMaxDelayInSeconds: utils.Int32(int32(eventsLateArrivalMaxDelayInSeconds)),
			EventsOutOfOrderMaxDelayInSeconds:  utils.Int32(int32(eventsOutOfOrderMaxDelayInSeconds)),
			EventsOutOfOrderPolicy:             streamanalytics.EventsOutOfOrderPolicy(eventsOutOfOrderPolicy),
			OutputErrorPolicy:                  streamanalytics.OutputErrorPolicy(outputErrorPolicy),
			JobType:                            streamanalytics.JobType(jobType),
		},
		Identity: expandedIdentity,
		Tags:     tags.Expand(t),
	}

	if jobType == string(streamanalytics.JobTypeEdge) {
		if _, ok := d.GetOk("stream_analytics_cluster_id"); ok {
			return fmt.Errorf("the job type `Edge` doesn't support `stream_analytics_cluster_id`")
		}
	} else {
		if streamAnalyticsCluster := d.Get("stream_analytics_cluster_id"); streamAnalyticsCluster != "" {
			props.StreamingJobProperties.Cluster = &streamanalytics.ClusterInfo{
				ID: utils.String(streamAnalyticsCluster.(string)),
			}
		} else {
			props.StreamingJobProperties.Cluster = &streamanalytics.ClusterInfo{
				ID: nil,
			}
		}
	}

	if dataLocale, ok := d.GetOk("data_locale"); ok {
		props.StreamingJobProperties.DataLocale = utils.String(dataLocale.(string))
	}

	if d.IsNewResource() {
		props.StreamingJobProperties.Transformation = &transformation

		future, err := client.CreateOrReplace(ctx, props, id.ResourceGroup, id.Name, "", "")
		if err != nil {
			return fmt.Errorf("creating %s: %+v", id, err)
		}

		if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
			return fmt.Errorf("waiting for creation of %s: %+v", id, err)
		}

		d.SetId(id.ID())
	} else {
		if _, err := client.Update(ctx, props, id.ResourceGroup, id.Name, ""); err != nil {
			return fmt.Errorf("updating %s: %+v", id, err)
		}

		job, err := client.Get(ctx, id.ResourceGroup, id.Name, "transformation")
		if err != nil {
			return err
		}

		if readTransformation := job.Transformation; readTransformation != nil {
			if _, err := transformationsClient.Update(ctx, transformation, id.ResourceGroup, id.Name, *readTransformation.Name, ""); err != nil {
				return fmt.Errorf("updating transformation for %s: %+v", id, err)
			}
		}
	}

	return resourceStreamAnalyticsJobRead(d, meta)
}

func resourceStreamAnalyticsJobRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).StreamAnalytics.JobsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.StreamingJobID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name, "transformation")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] %s was not found - removing from state!", *id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)

	if resp.Location != nil {
		d.Set("location", azure.NormalizeLocation(*resp.Location))
	}

	if err := d.Set("identity", flattenJobIdentity(resp.Identity)); err != nil {
		return fmt.Errorf("setting `identity`: %v", err)
	}

	if props := resp.StreamingJobProperties; props != nil {
		d.Set("compatibility_level", string(props.CompatibilityLevel))
		d.Set("data_locale", props.DataLocale)
		if props.EventsLateArrivalMaxDelayInSeconds != nil {
			d.Set("events_late_arrival_max_delay_in_seconds", int(*props.EventsLateArrivalMaxDelayInSeconds))
		}
		if props.EventsOutOfOrderMaxDelayInSeconds != nil {
			d.Set("events_out_of_order_max_delay_in_seconds", int(*props.EventsOutOfOrderMaxDelayInSeconds))
		}
		if props.Cluster != nil {
			d.Set("stream_analytics_cluster_id", props.Cluster.ID)
		}
		d.Set("events_out_of_order_policy", string(props.EventsOutOfOrderPolicy))
		d.Set("output_error_policy", string(props.OutputErrorPolicy))
		d.Set("type", string(props.JobType))

		// Computed
		d.Set("job_id", props.JobID)

		if transformation := props.Transformation; transformation != nil {
			if units := transformation.StreamingUnits; units != nil {
				d.Set("streaming_units", int(*units))
			}
			d.Set("transformation_query", transformation.Query)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceStreamAnalyticsJobDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).StreamAnalytics.JobsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.StreamingJobID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
	}

	return nil
}

func expandStreamAnalyticsJobIdentity(input []interface{}) (*streamanalytics.Identity, error) {
	expanded, err := identity.ExpandSystemAssigned(input)
	if err != nil {
		return nil, err
	}

	// Otherwise we get:
	//   Code="BadRequest"
	//   Message="The JSON provided in the request body is invalid. Cannot convert value 'None' to
	//   type 'System.Nullable`1[Microsoft.Streaming.Service.Contracts.CSMResourceProvider.IdentityType]"
	// Upstream issue: https://github.com/Azure/azure-rest-api-specs/issues/17649
	if expanded.Type == identity.TypeNone {
		return nil, nil
	}

	return &streamanalytics.Identity{
		Type: utils.String(string(expanded.Type)),
	}, nil
}

func flattenJobIdentity(identity *streamanalytics.Identity) []interface{} {
	if identity == nil {
		return nil
	}

	var t string
	if identity.Type != nil {
		t = *identity.Type
	}

	var tenantId string
	if identity.TenantID != nil {
		tenantId = *identity.TenantID
	}

	var principalId string
	if identity.PrincipalID != nil {
		principalId = *identity.PrincipalID
	}

	return []interface{}{
		map[string]interface{}{
			"type":         t,
			"tenant_id":    tenantId,
			"principal_id": principalId,
		},
	}
}
