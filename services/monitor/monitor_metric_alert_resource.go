package monitor

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2021-07-01-preview/insights"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/applicationinsights/validate"
	"github.com/hashicorp/terraform-provider-azurerm/services/monitor/migration"
	"github.com/hashicorp/terraform-provider-azurerm/services/monitor/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceMonitorMetricAlert() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceMonitorMetricAlertCreateUpdate,
		Read:   resourceMonitorMetricAlertRead,
		Update: resourceMonitorMetricAlertCreateUpdate,
		Delete: resourceMonitorMetricAlertDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.MetricAlertID(id)
			return err
		}),

		SchemaVersion: 1,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.MetricAlertUpgradeV0ToV1{},
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

			"scopes": {
				Type:     pluginsdk.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeString,
					ValidateFunc: azure.ValidateResourceID,
				},
				Set: pluginsdk.HashString,
			},

			"target_resource_type": {
				Type:        pluginsdk.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `The resource type (e.g. Microsoft.Compute/virtualMachines) of the target pluginsdk. Required when using subscription, resource group scope or multiple scopes.`,
			},

			"target_resource_location": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				Computed:         true,
				StateFunc:        location.StateFunc,
				DiffSuppressFunc: location.DiffSuppressFunc,
				Description:      `The location of the target pluginsdk. Required when using subscription, resource group scope or multiple scopes.`,
			},

			// static criteria
			"criteria": {
				Type:         pluginsdk.TypeSet,
				Optional:     true,
				MinItems:     1,
				ExactlyOneOf: []string{"criteria", "dynamic_criteria", "application_insights_web_test_location_availability_criteria"},
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"metric_namespace": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"metric_name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"aggregation": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Average",
								"Count",
								"Minimum",
								"Maximum",
								"Total",
							}, false),
						},
						"dimension": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"name": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"operator": {
										Type:     pluginsdk.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"Include",
											"Exclude",
											"StartsWith",
										}, false),
									},
									"values": {
										Type:     pluginsdk.TypeList,
										Required: true,
										MinItems: 1,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
										},
									},
								},
							},
						},
						"operator": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(insights.OperatorEquals),
								string(insights.OperatorGreaterThan),
								string(insights.OperatorGreaterThanOrEqual),
								string(insights.OperatorLessThan),
								string(insights.OperatorLessThanOrEqual),
								string(insights.OperatorNotEquals),
							}, false),
						},
						"threshold": {
							Type:     pluginsdk.TypeFloat,
							Required: true,
						},
						"skip_metric_validation": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			//lintignore: S018
			"dynamic_criteria": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				MinItems: 1,
				// Curently, it allows to define only one dynamic criteria in one metric alert.
				MaxItems:     1,
				ExactlyOneOf: []string{"criteria", "dynamic_criteria", "application_insights_web_test_location_availability_criteria"},
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"metric_namespace": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"metric_name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"aggregation": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Average",
								"Count",
								"Minimum",
								"Maximum",
								"Total",
							}, false),
						},
						"dimension": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"name": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"operator": {
										Type:     pluginsdk.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"Include",
											"Exclude",
											"StartsWith",
										}, false),
									},
									"values": {
										Type:     pluginsdk.TypeList,
										Required: true,
										MinItems: 1,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
										},
									},
								},
							},
						},
						"operator": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(insights.DynamicThresholdOperatorLessThan),
								string(insights.DynamicThresholdOperatorGreaterThan),
								string(insights.DynamicThresholdOperatorGreaterOrLessThan),
							}, false),
						},
						"alert_sensitivity": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(insights.DynamicThresholdSensitivityLow),
								string(insights.DynamicThresholdSensitivityMedium),
								string(insights.DynamicThresholdSensitivityHigh),
							}, false),
						},

						"evaluation_total_count": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
							Default:      4,
						},

						"evaluation_failure_count": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
							Default:      4,
						},

						"ignore_data_before": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsRFC3339Time,
						},
						"skip_metric_validation": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
						},
					},
				},
			},

			"application_insights_web_test_location_availability_criteria": {
				Type:         pluginsdk.TypeList,
				Optional:     true,
				MinItems:     1,
				MaxItems:     1,
				ExactlyOneOf: []string{"criteria", "dynamic_criteria", "application_insights_web_test_location_availability_criteria"},
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"web_test_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validate.WebTestID,
						},
						"component_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validate.ComponentID,
						},
						"failed_location_count": {
							Type:         pluginsdk.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},

			"action": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"action_group_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: azure.ValidateResourceID,
						},
						"webhook_properties": {
							Type:     pluginsdk.TypeMap,
							Optional: true,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
							},
						},
					},
				},
				Set: resourceMonitorMetricAlertActionHash,
			},

			"auto_mitigate": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"description": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"frequency": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  "PT1M",
				ValidateFunc: validation.StringInSlice([]string{
					"PT1M",
					"PT5M",
					"PT15M",
					"PT30M",
					"PT1H",
				}, false),
			},

			"severity": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Default:      3,
				ValidateFunc: validation.IntBetween(0, 4),
			},

			"window_size": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  "PT5M",
				ValidateFunc: validation.StringInSlice([]string{
					"PT1M",
					"PT5M",
					"PT15M",
					"PT30M",
					"PT1H",
					"PT6H",
					"PT12H",
					"P1D",
				}, false),
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceMonitorMetricAlertCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.MetricAlertsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewMetricAlertID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Monitor %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_monitor_metric_alert", id.ID())
		}
	}

	enabled := d.Get("enabled").(bool)
	autoMitigate := d.Get("auto_mitigate").(bool)
	description := d.Get("description").(string)
	scopesRaw := d.Get("scopes").(*pluginsdk.Set).List()
	severity := d.Get("severity").(int)
	frequency := d.Get("frequency").(string)
	windowSize := d.Get("window_size").(string)
	actionRaw := d.Get("action").(*pluginsdk.Set).List()
	targetResourceType := d.Get("target_resource_type").(string)
	targetResourceLocation := d.Get("target_resource_location").(string)

	t := d.Get("tags").(map[string]interface{})
	expandedTags := tags.Expand(t)

	// The criteria type of "old" resource is `MetricAlertSingleResourceMultipleMetricCriteria` (rather than `MetricAlertMultipleResourceMultipleMetricCriteria`).
	// We need to keep using that type in order to keep backward compatibility. Otherwise, changing the criteria type will cause error as reported in issue:
	// https://github.com/hashicorp/terraform-provider-azurerm/issues/7910
	var isLegacy bool
	if !d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
		if err != nil {
			return fmt.Errorf("retrieving Monitor %s: %+v", id, err)
		}
		if existing.MetricAlertProperties == nil || existing.MetricAlertProperties.Criteria == nil {
			return fmt.Errorf("unexpected nil properties of Monitor %s", id)
		}
		_, isLegacy = existing.MetricAlertProperties.Criteria.AsMetricAlertSingleResourceMultipleMetricCriteria()
	}

	criteria, err := expandMonitorMetricAlertCriteria(d, isLegacy)
	if err != nil {
		return fmt.Errorf(`Expanding criteria: %+v`, err)
	}

	parameters := insights.MetricAlertResource{
		Location: utils.String(azure.NormalizeLocation("Global")),
		MetricAlertProperties: &insights.MetricAlertProperties{
			Enabled:              utils.Bool(enabled),
			AutoMitigate:         utils.Bool(autoMitigate),
			Description:          utils.String(description),
			Severity:             utils.Int32(int32(severity)),
			EvaluationFrequency:  utils.String(frequency),
			WindowSize:           utils.String(windowSize),
			Scopes:               utils.ExpandStringSlice(scopesRaw),
			Criteria:             criteria,
			Actions:              expandMonitorMetricAlertAction(actionRaw),
			TargetResourceType:   utils.String(targetResourceType),
			TargetResourceRegion: utils.String(targetResourceLocation),
		},
		Tags: expandedTags,
	}

	if _, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, parameters); err != nil {
		return fmt.Errorf("creating or updating Monitor %s: %+v", id, err)
	}

	// Monitor Metric Alert API would return 404 while creating multiple Monitor Metric Alerts and get each resource immediately once it's created successfully in parallel.
	// Tracked by this issue: https://github.com/Azure/azure-rest-api-specs/issues/10973
	log.Printf("[DEBUG] Waiting for Monitor Metric Alert %q (Resource Group %q) to be created", id.Name, id.ResourceGroup)
	stateConf := &pluginsdk.StateChangeConf{
		Pending:                   []string{"404"},
		Target:                    []string{"200"},
		Refresh:                   monitorMetricAlertStateRefreshFunc(ctx, client, id.ResourceGroup, id.Name),
		MinTimeout:                15 * time.Second,
		ContinuousTargetOccurence: 10,
	}

	if d.IsNewResource() {
		stateConf.Timeout = d.Timeout(pluginsdk.TimeoutCreate)
	} else {
		stateConf.Timeout = d.Timeout(pluginsdk.TimeoutUpdate)
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for Monitor %s to finish provisioning: %s", id, err)
	}

	d.SetId(id.ID())

	return resourceMonitorMetricAlertRead(d, meta)
}

func resourceMonitorMetricAlertRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.MetricAlertsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.MetricAlertID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Metric Alert %q was not found in Resource Group %q - removing from state!", id.Name, id.ResourceGroup)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("getting Monitor %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	if alert := resp.MetricAlertProperties; alert != nil {
		d.Set("enabled", alert.Enabled)
		d.Set("auto_mitigate", alert.AutoMitigate)
		d.Set("description", alert.Description)
		d.Set("severity", alert.Severity)
		d.Set("frequency", alert.EvaluationFrequency)
		d.Set("window_size", alert.WindowSize)
		if err := d.Set("scopes", utils.FlattenStringSlice(alert.Scopes)); err != nil {
			return fmt.Errorf("setting `scopes`: %+v", err)
		}

		// Determine the correct criteria schema to set
		var criteriaSchema string
		switch c := alert.Criteria.(type) {
		case insights.MetricAlertSingleResourceMultipleMetricCriteria:
			criteriaSchema = "criteria"
		case insights.MetricAlertMultipleResourceMultipleMetricCriteria:
			if c.AllOf == nil || len(*c.AllOf) == 0 {
				return fmt.Errorf("nil or empty contained criteria of MultipleResourceMultipleMetricCriteria")
			}
			// `MinItems` defined in schema guaranteed there is at least one element.
			switch (*c.AllOf)[0].(type) {
			case insights.DynamicMetricCriteria:
				criteriaSchema = "dynamic_criteria"
			case insights.MetricCriteria:
				criteriaSchema = "criteria"
			}
		case insights.WebtestLocationAvailabilityCriteria:
			criteriaSchema = "application_insights_web_test_location_availability_criteria"
		default:
			return fmt.Errorf("Unknown criteria type")
		}

		monitorMetricAlertCriteria := flattenMonitorMetricAlertCriteria(alert.Criteria)
		//lintignore:R001
		if err := d.Set(criteriaSchema, monitorMetricAlertCriteria); err != nil {
			return fmt.Errorf("failed setting `%s`: %+v", criteriaSchema, err)
		}

		if err := d.Set("action", flattenMonitorMetricAlertAction(alert.Actions)); err != nil {
			return fmt.Errorf("setting `action`: %+v", err)
		}
		d.Set("target_resource_type", alert.TargetResourceType)
		d.Set("target_resource_location", alert.TargetResourceRegion)
	}
	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceMonitorMetricAlertDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.MetricAlertsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.MetricAlertID(d.Id())
	if err != nil {
		return err
	}

	if resp, err := client.Delete(ctx, id.ResourceGroup, id.Name); err != nil {
		if !response.WasNotFound(resp.Response) {
			return fmt.Errorf("deleting Monitor %s: %+v", *id, err)
		}
	}

	return nil
}

func expandMonitorMetricAlertCriteria(d *pluginsdk.ResourceData, isLegacy bool) (insights.BasicMetricAlertCriteria, error) {
	switch {
	case d.Get("criteria").(*pluginsdk.Set).Len() != 0:
		if isLegacy {
			return expandMonitorMetricAlertSingleResourceMultiMetricCriteria(d.Get("criteria").(*pluginsdk.Set).List()), nil
		}
		return expandMonitorMetricAlertMultiResourceMultiMetricForStaticMetricCriteria(d.Get("criteria").(*pluginsdk.Set).List()), nil
	case d.Get("dynamic_criteria").(*pluginsdk.Set).Len() != 0:
		return expandMonitorMetricAlertMultiResourceMultiMetricForDynamicMetricCriteria(d.Get("dynamic_criteria").(*pluginsdk.Set).List()), nil
	case len(d.Get("application_insights_web_test_location_availability_criteria").([]interface{})) != 0:
		return expandMonitorMetricAlertWebtestLocAvailCriteria(d.Get("application_insights_web_test_location_availability_criteria").([]interface{})), nil
	default:
		// Guaranteed by schema `AtLeastOne` constraint
		return nil, fmt.Errorf("unknown criteria type")
	}
}

func expandMonitorMetricAlertSingleResourceMultiMetricCriteria(input []interface{}) insights.BasicMetricAlertCriteria {
	criteria := make([]insights.MetricCriteria, 0)
	for i, item := range input {
		v := item.(map[string]interface{})
		dimensions := expandMonitorMetricDimension(v["dimension"].([]interface{}))
		criteria = append(criteria, insights.MetricCriteria{
			Name:                 utils.String(fmt.Sprintf("Metric%d", i+1)),
			MetricNamespace:      utils.String(v["metric_namespace"].(string)),
			MetricName:           utils.String(v["metric_name"].(string)),
			TimeAggregation:      insights.AggregationTypeEnum(v["aggregation"].(string)),
			Dimensions:           &dimensions,
			Operator:             insights.Operator(v["operator"].(string)),
			Threshold:            utils.Float(v["threshold"].(float64)),
			SkipMetricValidation: utils.Bool(v["skip_metric_validation"].(bool)),
		})
	}
	return &insights.MetricAlertSingleResourceMultipleMetricCriteria{
		AllOf:     &criteria,
		OdataType: insights.OdataTypeBasicMetricAlertCriteriaOdataTypeMicrosoftAzureMonitorSingleResourceMultipleMetricCriteria,
	}
}

func expandMonitorMetricAlertMultiResourceMultiMetricForStaticMetricCriteria(input []interface{}) insights.BasicMetricAlertCriteria {
	criteria := make([]insights.BasicMultiMetricCriteria, 0)
	for i, item := range input {
		v := item.(map[string]interface{})
		dimensions := expandMonitorMetricDimension(v["dimension"].([]interface{}))
		criteria = append(criteria, insights.MetricCriteria{
			Name:                 utils.String(fmt.Sprintf("Metric%d", i+1)),
			MetricNamespace:      utils.String(v["metric_namespace"].(string)),
			MetricName:           utils.String(v["metric_name"].(string)),
			TimeAggregation:      insights.AggregationTypeEnum(v["aggregation"].(string)),
			Dimensions:           &dimensions,
			Operator:             insights.Operator(v["operator"].(string)),
			Threshold:            utils.Float(v["threshold"].(float64)),
			SkipMetricValidation: utils.Bool(v["skip_metric_validation"].(bool)),
		})
	}
	return &insights.MetricAlertMultipleResourceMultipleMetricCriteria{
		AllOf:     &criteria,
		OdataType: insights.OdataTypeBasicMetricAlertCriteriaOdataTypeMicrosoftAzureMonitorMultipleResourceMultipleMetricCriteria,
	}
}

func expandMonitorMetricAlertMultiResourceMultiMetricForDynamicMetricCriteria(input []interface{}) insights.BasicMetricAlertCriteria {
	criteria := make([]insights.BasicMultiMetricCriteria, 0)
	for i, item := range input {
		v := item.(map[string]interface{})
		dimensions := expandMonitorMetricDimension(v["dimension"].([]interface{}))
		var ignoreDataBefore *date.Time
		if v := v["ignore_data_before"].(string); v != "" {
			// Guaranteed in schema validation func.
			t, _ := time.Parse(time.RFC3339, v)
			ignoreDataBefore = &date.Time{Time: t}
		}
		criteria = append(criteria, insights.DynamicMetricCriteria{
			Name:             utils.String(fmt.Sprintf("Metric%d", i+1)),
			MetricNamespace:  utils.String(v["metric_namespace"].(string)),
			MetricName:       utils.String(v["metric_name"].(string)),
			TimeAggregation:  insights.AggregationTypeEnum(v["aggregation"].(string)),
			Dimensions:       &dimensions,
			Operator:         insights.DynamicThresholdOperator(v["operator"].(string)),
			AlertSensitivity: insights.DynamicThresholdSensitivity(v["alert_sensitivity"].(string)),
			FailingPeriods: &insights.DynamicThresholdFailingPeriods{
				NumberOfEvaluationPeriods: utils.Float(float64(v["evaluation_total_count"].(int))),
				MinFailingPeriodsToAlert:  utils.Float(float64(v["evaluation_failure_count"].(int))),
			},
			IgnoreDataBefore:     ignoreDataBefore,
			SkipMetricValidation: utils.Bool(v["skip_metric_validation"].(bool)),
		})
	}
	return &insights.MetricAlertMultipleResourceMultipleMetricCriteria{
		AllOf:     &criteria,
		OdataType: insights.OdataTypeBasicMetricAlertCriteriaOdataTypeMicrosoftAzureMonitorMultipleResourceMultipleMetricCriteria,
	}
}

func expandMonitorMetricAlertWebtestLocAvailCriteria(input []interface{}) insights.BasicMetricAlertCriteria {
	if len(input) == 0 {
		return nil
	}
	v := input[0].(map[string]interface{})
	return &insights.WebtestLocationAvailabilityCriteria{
		WebTestID:           utils.String(v["web_test_id"].(string)),
		ComponentID:         utils.String(v["component_id"].(string)),
		FailedLocationCount: utils.Float(float64(v["failed_location_count"].(int))),
		OdataType:           insights.OdataTypeBasicMetricAlertCriteriaOdataTypeMicrosoftAzureMonitorWebtestLocationAvailabilityCriteria,
	}
}

func expandMonitorMetricDimension(input []interface{}) []insights.MetricDimension {
	result := make([]insights.MetricDimension, 0)
	for _, dimension := range input {
		dVal := dimension.(map[string]interface{})
		result = append(result, insights.MetricDimension{
			Name:     utils.String(dVal["name"].(string)),
			Operator: utils.String(dVal["operator"].(string)),
			Values:   utils.ExpandStringSlice(dVal["values"].([]interface{})),
		})
	}
	return result
}

func expandMonitorMetricAlertAction(input []interface{}) *[]insights.MetricAlertAction {
	actions := make([]insights.MetricAlertAction, 0)
	for _, item := range input {
		v := item.(map[string]interface{})
		if agID := v["action_group_id"].(string); agID != "" {
			props := make(map[string]*string)
			if pVal, ok := v["webhook_properties"]; ok {
				for pk, pv := range pVal.(map[string]interface{}) {
					props[pk] = utils.String(pv.(string))
				}
			}

			actions = append(actions, insights.MetricAlertAction{
				ActionGroupID:     utils.String(agID),
				WebHookProperties: props,
			})
		}
	}
	return &actions
}

func flattenMonitorMetricAlertCriteria(input insights.BasicMetricAlertCriteria) []interface{} {
	switch criteria := input.(type) {
	case insights.MetricAlertSingleResourceMultipleMetricCriteria:
		return flattenMonitorMetricAlertSingleResourceMultiMetricCriteria(criteria.AllOf)
	case insights.MetricAlertMultipleResourceMultipleMetricCriteria:
		return flattenMonitorMetricAlertMultiResourceMultiMetricCriteria(criteria.AllOf)
	case insights.WebtestLocationAvailabilityCriteria:
		return flattenMonitorMetricAlertWebtestLocAvailCriteria(&criteria)
	default:
		return nil
	}
}

func flattenMonitorMetricAlertSingleResourceMultiMetricCriteria(input *[]insights.MetricCriteria) []interface{} {
	if input == nil || len(*input) == 0 {
		return nil
	}
	criteria := (*input)[0]

	metricName := ""
	if criteria.MetricName != nil {
		metricName = *criteria.MetricName
	}

	metricNamespace := ""
	if criteria.MetricNamespace != nil {
		metricNamespace = *criteria.MetricNamespace
	}

	timeAggregation := criteria.TimeAggregation

	dimResult := make([]map[string]interface{}, 0)
	if criteria.Dimensions != nil {
		for _, dimension := range *criteria.Dimensions {
			dVal := make(map[string]interface{})
			if dimension.Name != nil {
				dVal["name"] = *dimension.Name
			}
			if dimension.Operator != nil {
				dVal["operator"] = *dimension.Operator
			}
			dVal["values"] = utils.FlattenStringSlice(dimension.Values)
			dimResult = append(dimResult, dVal)
		}
	}

	operator := string(criteria.Operator)

	threshold := 0.0
	if criteria.Threshold != nil {
		threshold = *criteria.Threshold
	}

	var skipMetricValidation bool
	if criteria.SkipMetricValidation != nil {
		skipMetricValidation = *criteria.SkipMetricValidation
	}

	return []interface{}{
		map[string]interface{}{
			"metric_namespace":       metricNamespace,
			"metric_name":            metricName,
			"aggregation":            timeAggregation,
			"dimension":              dimResult,
			"operator":               operator,
			"threshold":              threshold,
			"skip_metric_validation": skipMetricValidation,
		},
	}
}

func flattenMonitorMetricAlertMultiResourceMultiMetricCriteria(input *[]insights.BasicMultiMetricCriteria) []interface{} {
	if input == nil {
		return nil
	}
	result := make([]interface{}, 0)

	for _, criteria := range *input {
		v := make(map[string]interface{})
		var (
			metricName           string
			metricNamespace      string
			timeAggregation      interface{}
			dimensions           []insights.MetricDimension
			skipMetricValidation bool
		)

		switch criteria := criteria.(type) {
		case insights.MetricCriteria:
			if criteria.MetricName != nil {
				metricName = *criteria.MetricName
			}
			if criteria.MetricNamespace != nil {
				metricNamespace = *criteria.MetricNamespace
			}
			timeAggregation = criteria.TimeAggregation
			if criteria.Dimensions != nil {
				dimensions = *criteria.Dimensions
			}

			// MetricCriteria specific properties
			v["operator"] = string(criteria.Operator)

			threshold := 0.0
			if criteria.Threshold != nil {
				threshold = *criteria.Threshold
			}
			v["threshold"] = threshold
			if criteria.SkipMetricValidation != nil {
				skipMetricValidation = *criteria.SkipMetricValidation
			}
		case insights.DynamicMetricCriteria:
			if criteria.MetricName != nil {
				metricName = *criteria.MetricName
			}
			if criteria.MetricNamespace != nil {
				metricNamespace = *criteria.MetricNamespace
			}
			timeAggregation = criteria.TimeAggregation
			if criteria.Dimensions != nil {
				dimensions = *criteria.Dimensions
			}
			if criteria.SkipMetricValidation != nil {
				skipMetricValidation = *criteria.SkipMetricValidation
			}
			// DynamicMetricCriteria specific properties
			v["operator"] = string(criteria.Operator)
			v["alert_sensitivity"] = string(criteria.AlertSensitivity)
			var (
				nEvl     = 1
				nFailEvl = 1
			)
			if period := criteria.FailingPeriods; period != nil {
				if period.NumberOfEvaluationPeriods != nil {
					nEvl = int(*period.NumberOfEvaluationPeriods)
				}
				if period.MinFailingPeriodsToAlert != nil {
					nFailEvl = int(*period.MinFailingPeriodsToAlert)
				}
			}
			v["evaluation_total_count"] = nEvl
			v["evaluation_failure_count"] = nFailEvl

			ignoreDataBefore := ""
			if criteria.IgnoreDataBefore != nil {
				ignoreDataBefore = criteria.IgnoreDataBefore.Format(time.RFC3339)
			}
			v["ignore_data_before"] = ignoreDataBefore
		}

		// Common properties
		v["metric_name"] = metricName
		v["metric_namespace"] = metricNamespace
		v["aggregation"] = timeAggregation
		v["skip_metric_validation"] = skipMetricValidation
		if dimensions != nil {
			dimResult := make([]map[string]interface{}, 0)
			for _, dimension := range dimensions {
				dVal := make(map[string]interface{})
				if dimension.Name != nil {
					dVal["name"] = *dimension.Name
				}
				if dimension.Operator != nil {
					dVal["operator"] = *dimension.Operator
				}
				dVal["values"] = utils.FlattenStringSlice(dimension.Values)
				dimResult = append(dimResult, dVal)
			}
			v["dimension"] = dimResult
		}

		result = append(result, v)
	}
	return result
}

func flattenMonitorMetricAlertWebtestLocAvailCriteria(input *insights.WebtestLocationAvailabilityCriteria) []interface{} {
	if input == nil {
		return nil
	}
	webtestID := ""
	if input.WebTestID != nil {
		webtestID = *input.WebTestID
	}

	componentID := ""
	if input.ComponentID != nil {
		componentID = *input.ComponentID
	}

	failedLocationCount := 0
	if input.FailedLocationCount != nil {
		failedLocationCount = int(*input.FailedLocationCount)
	}

	return []interface{}{
		map[string]interface{}{
			"web_test_id":           webtestID,
			"component_id":          componentID,
			"failed_location_count": failedLocationCount,
		},
	}
}

func flattenMonitorMetricAlertAction(input *[]insights.MetricAlertAction) (result []interface{}) {
	result = make([]interface{}, 0)
	if input == nil {
		return
	}
	for _, action := range *input {
		v := make(map[string]interface{})

		if action.ActionGroupID != nil {
			v["action_group_id"] = *action.ActionGroupID
		}

		props := make(map[string]string)
		for pk, pv := range action.WebHookProperties {
			if pv != nil {
				props[pk] = *pv
			}
		}
		v["webhook_properties"] = props

		result = append(result, v)
	}

	return result
}

func resourceMonitorMetricAlertActionHash(input interface{}) int {
	var buf bytes.Buffer
	if v, ok := input.(map[string]interface{}); ok {
		buf.WriteString(fmt.Sprintf("%s-", v["action_group_id"].(string)))
	}
	return pluginsdk.HashString(buf.String())
}

func monitorMetricAlertStateRefreshFunc(ctx context.Context, client *insights.MetricAlertsClient, resourceGroupName string, name string) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		res, err := client.Get(ctx, resourceGroupName, name)
		if err != nil {
			if utils.ResponseWasNotFound(res.Response) {
				return nil, "404", nil
			}

			return nil, "", fmt.Errorf("retrieving Monitor Metric Alert %q (Resource Group %q): %s", name, resourceGroupName, err)
		}

		return res, strconv.Itoa(res.StatusCode), nil
	}
}
