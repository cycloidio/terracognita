package iottimeseriesinsights

import (
	"fmt"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/timeseriesinsights/mgmt/2020-05-15/timeseriesinsights"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	azValidate "github.com/hashicorp/terraform-provider-azurerm/helpers/validate"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/iottimeseriesinsights/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceIoTTimeSeriesInsightsGen2Environment() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceIoTTimeSeriesInsightsGen2EnvironmentCreateUpdate,
		Read:   resourceIoTTimeSeriesInsightsGen2EnvironmentRead,
		Update: resourceIoTTimeSeriesInsightsGen2EnvironmentCreateUpdate,
		Delete: resourceIoTTimeSeriesInsightsGen2EnvironmentDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.EnvironmentID(id)
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
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^[-\w\._\(\)]+$`),
					"IoT Time Series Insights Gen2 Environment name must contain only word characters, periods, underscores, and parentheses.",
				),
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"sku_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"L1",
				}, false),
			},

			"warm_store_data_retention_time": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: azValidate.ISO8601Duration,
			},
			"id_properties": {
				Type:     pluginsdk.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
			},
			"storage": {
				Type:     pluginsdk.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"key": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"data_access_fqdn": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceIoTTimeSeriesInsightsGen2EnvironmentCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).IoTTimeSeriesInsights.EnvironmentsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewEnvironmentID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	location := azure.NormalizeLocation(d.Get("location").(string))
	t := d.Get("tags").(map[string]interface{})
	sku, err := convertEnvironmentSkuName(d.Get("sku_name").(string))
	if err != nil {
		return fmt.Errorf("expanding sku: %+v", err)
	}

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %s", id, err)
			}
		}

		if existing.Value != nil {
			environment, ok := existing.Value.AsGen2EnvironmentResource()
			if !ok {
				return fmt.Errorf("exisiting resource was not %s", id)
			}

			if environment.ID != nil && *environment.ID != "" {
				return tf.ImportAsExistsError("azurerm_iot_time_series_insights_gen2_environment", *environment.ID)
			}
		}
	}

	environment := timeseriesinsights.Gen2EnvironmentCreateOrUpdateParameters{
		Location: &location,
		Tags:     tags.Expand(t),
		Sku:      sku,
		Gen2EnvironmentCreationProperties: &timeseriesinsights.Gen2EnvironmentCreationProperties{
			TimeSeriesIDProperties: expandIdProperties(d.Get("id_properties").(*pluginsdk.Set).List()),
			StorageConfiguration:   expandStorage(d.Get("storage").([]interface{})),
		},
	}

	if v, ok := d.GetOk("warm_store_data_retention_time"); ok {
		environment.WarmStoreConfiguration = &timeseriesinsights.WarmStoreConfigurationProperties{
			DataRetention: utils.String(v.(string)),
		}
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, environment)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for completion of %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceIoTTimeSeriesInsightsGen2EnvironmentRead(d, meta)
}

func resourceIoTTimeSeriesInsightsGen2EnvironmentRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).IoTTimeSeriesInsights.EnvironmentsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.EnvironmentID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil || resp.Value == nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving IoT Time Series Insights Standard Environment %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	environment, ok := resp.Value.AsGen2EnvironmentResource()
	if !ok {
		return fmt.Errorf("exisiting resource was not a standard IoT Time Series Insights Standard Environment %q (Resource Group %q)", id.Name, id.ResourceGroup)
	}

	d.Set("name", environment.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("sku_name", environment.Sku.Name)
	d.Set("location", location.NormalizeNilable(environment.Location))
	d.Set("data_access_fqdn", environment.DataAccessFqdn)
	if err := d.Set("id_properties", flattenIdProperties(environment.TimeSeriesIDProperties)); err != nil {
		return fmt.Errorf("setting `id_properties`: %+v", err)
	}
	if props := environment.WarmStoreConfiguration; props != nil {
		d.Set("warm_store_data_retention_time", props.DataRetention)
	}
	if err := d.Set("storage", flattenIoTTimeSeriesGen2EnvironmentStorage(environment.StorageConfiguration, d.Get("storage.0.key").(string))); err != nil {
		return fmt.Errorf("setting `storage`: %+v", err)
	}

	return tags.FlattenAndSet(d, environment.Tags)
}

func resourceIoTTimeSeriesInsightsGen2EnvironmentDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).IoTTimeSeriesInsights.EnvironmentsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.EnvironmentID(d.Id())
	if err != nil {
		return err
	}

	response, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if !utils.ResponseWasNotFound(response) {
			return fmt.Errorf("deleting IoT Time Series Insights Gen2 Environment %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}
	}

	return nil
}

func convertEnvironmentSkuName(skuName string) (*timeseriesinsights.Sku, error) {
	var name timeseriesinsights.SkuName
	switch skuName {
	case "L1":
		name = timeseriesinsights.L1
	default:
		return nil, fmt.Errorf("sku_name %s has unknown sku tier %s", skuName, skuName)
	}

	// Gen2 cannot set capacity manually but SDK requires capacity
	capacity := utils.Int32(1)

	return &timeseriesinsights.Sku{
		Name:     name,
		Capacity: capacity,
	}, nil
}

func expandStorage(input []interface{}) *timeseriesinsights.Gen2StorageConfigurationInput {
	if input == nil || input[0] == nil {
		return nil
	}
	storageMap := input[0].(map[string]interface{})
	accountName := storageMap["name"].(string)
	managementKey := storageMap["key"].(string)

	return &timeseriesinsights.Gen2StorageConfigurationInput{
		AccountName:   &accountName,
		ManagementKey: &managementKey,
	}
}

func expandIdProperties(input []interface{}) *[]timeseriesinsights.TimeSeriesIDProperty {
	if input == nil || input[0] == nil {
		return nil
	}
	result := make([]timeseriesinsights.TimeSeriesIDProperty, 0)
	for _, item := range input {
		result = append(result, timeseriesinsights.TimeSeriesIDProperty{
			Name: utils.String(item.(string)),
			Type: "String",
		})
	}
	return &result
}

func flattenIdProperties(input *[]timeseriesinsights.TimeSeriesIDProperty) []string {
	output := make([]string, 0)
	if input == nil {
		return output
	}

	for _, v := range *input {
		if v.Name != nil {
			output = append(output, *v.Name)
		}
	}

	return output
}

func flattenIoTTimeSeriesGen2EnvironmentStorage(input *timeseriesinsights.Gen2StorageConfigurationOutput, key string) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	attr := make(map[string]interface{})
	if input.AccountName != nil {
		attr["name"] = *input.AccountName
	}
	// Key is not returned by the api so we'll set it to the key from config to help with diffs
	attr["key"] = key

	return []interface{}{attr}
}
