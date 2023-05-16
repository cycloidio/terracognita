package monitor

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2021-07-01-preview/insights"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/monitor/migration"
	"github.com/hashicorp/terraform-provider-azurerm/services/monitor/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/monitor/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceMonitorScheduledQueryRulesAlert() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceMonitorScheduledQueryRulesAlertCreateUpdate,
		Read:   resourceMonitorScheduledQueryRulesAlertRead,
		Update: resourceMonitorScheduledQueryRulesAlertCreateUpdate,
		Delete: resourceMonitorScheduledQueryRulesAlertDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ScheduledQueryRulesID(id)
			return err
		}),

		SchemaVersion: 1,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.ScheduledQueryRulesAlertUpgradeV0ToV1{},
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
				ValidateFunc: validation.StringDoesNotContainAny("<>*%&:\\?+/"),
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"authorized_resource_ids": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				MaxItems: 100,
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeString,
					ValidateFunc: azure.ValidateResourceID,
				},
			},
			"action": {
				Type:     pluginsdk.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"action_group": {
							Type:     pluginsdk.TypeSet,
							Required: true,
							Elem: &pluginsdk.Schema{
								Type:         pluginsdk.TypeString,
								ValidateFunc: azure.ValidateResourceID,
							},
						},
						"custom_webhook_payload": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsJSON,
						},
						"email_subject": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},
			"data_source_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: azure.ValidateResourceID,
			},
			"auto_mitigation_enabled": {
				Type:          pluginsdk.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"throttling"},
			},
			"description": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 4096),
			},
			"enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},
			"frequency": {
				Type:         pluginsdk.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(5, 1440),
			},
			"query": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"query_type": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  "ResultCount",
				ValidateFunc: validation.StringInSlice([]string{
					"ResultCount",
				}, false),
			},
			"severity": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 4),
			},
			"throttling": {
				Type:          pluginsdk.TypeInt,
				Optional:      true,
				ValidateFunc:  validation.IntBetween(0, 10000),
				ConflictsWith: []string{"auto_mitigation_enabled"},
			},
			"time_window": {
				Type:         pluginsdk.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(5, 2880),
			},
			"trigger": {
				Type:     pluginsdk.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"metric_trigger": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"metric_column": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"metric_trigger_type": {
										Type:     pluginsdk.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"Consecutive",
											"Total",
										}, false),
									},
									"operator": {
										Type:     pluginsdk.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"GreaterThan",
											"GreaterThanOrEqual",
											"LessThan",
											"LessThanOrEqual",
											"Equal",
										}, false),
									},
									"threshold": {
										Type:         pluginsdk.TypeFloat,
										Required:     true,
										ValidateFunc: validate.ScheduledQueryRulesAlertThreshold,
									},
								},
							},
						},
						"operator": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"GreaterThan",
								"GreaterThanOrEqual",
								"LessThan",
								"LessThanOrEqual",
								"Equal",
							}, false),
						},
						"threshold": {
							Type:         pluginsdk.TypeFloat,
							Required:     true,
							ValidateFunc: validate.ScheduledQueryRulesAlertThreshold,
						},
					},
				},
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceMonitorScheduledQueryRulesAlertCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	action := expandMonitorScheduledQueryRulesAlertingAction(d)
	schedule := expandMonitorScheduledQueryRulesAlertSchedule(d)
	client := meta.(*clients.Client).Monitor.ScheduledQueryRulesClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewScheduledQueryRulesID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	frequency := d.Get("frequency").(int)
	timeWindow := d.Get("time_window").(int)
	if timeWindow < frequency {
		return fmt.Errorf("in parameter values for %s: time_window must be greater than or equal to frequency", id)
	}

	query := d.Get("query").(string)
	_, ok := d.GetOk("metric_trigger")
	if ok {
		if !(strings.Contains(query, "summarize") &&
			strings.Contains(query, "AggregatedValue") &&
			strings.Contains(query, "bin")) {
			return fmt.Errorf("in parameter values for %s: query must contain summarize, AggregatedValue, and bin when metric_trigger is specified", id)
		}
	}

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.ScheduledQueryRuleName)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Monitor %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_monitor_scheduled_query_rules_alert", id.ID())
		}
	}

	autoMitigate := d.Get("auto_mitigation_enabled").(bool)
	description := d.Get("description").(string)
	enabledRaw := d.Get("enabled").(bool)

	enabled := insights.EnabledTrue
	if !enabledRaw {
		enabled = insights.EnabledFalse
	}

	location := azure.NormalizeLocation(d.Get("location"))

	source := expandMonitorScheduledQueryRulesCommonSource(d)

	t := d.Get("tags").(map[string]interface{})
	expandedTags := tags.Expand(t)

	parameters := insights.LogSearchRuleResource{
		Location: utils.String(location),
		LogSearchRule: &insights.LogSearchRule{
			Description:  utils.String(description),
			Enabled:      enabled,
			Source:       source,
			Schedule:     schedule,
			Action:       action,
			AutoMitigate: utils.Bool(autoMitigate),
		},
		Tags: expandedTags,
	}

	if _, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.ScheduledQueryRuleName, parameters); err != nil {
		return fmt.Errorf("creating or updating Monitor %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceMonitorScheduledQueryRulesAlertRead(d, meta)
}

func resourceMonitorScheduledQueryRulesAlertRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.ScheduledQueryRulesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ScheduledQueryRulesID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.ScheduledQueryRuleName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Scheduled Query Rule %q was not found in Resource Group %q", id.ScheduledQueryRuleName, id.ResourceGroup)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("getting Monitor %s: %+v", *id, err)
	}

	d.Set("name", id.ScheduledQueryRuleName)
	d.Set("resource_group_name", id.ResourceGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	d.Set("auto_mitigation_enabled", resp.AutoMitigate)
	d.Set("description", resp.Description)
	if resp.Enabled == insights.EnabledTrue {
		d.Set("enabled", true)
	} else {
		d.Set("enabled", false)
	}

	action, ok := resp.Action.(insights.AlertingAction)
	if !ok {
		return fmt.Errorf("wrong action type in %s: %T", *id, resp.Action)
	}
	if err = d.Set("action", flattenAzureRmScheduledQueryRulesAlertAction(action.AznsAction)); err != nil {
		return fmt.Errorf("setting `action`: %+v", err)
	}
	severity, err := strconv.Atoi(string(action.Severity))
	if err != nil {
		return fmt.Errorf("converting action.Severity %q to int in %s: %+v", action.Severity, *id, err)
	}
	d.Set("severity", severity)
	d.Set("throttling", action.ThrottlingInMin)
	if err = d.Set("trigger", flattenAzureRmScheduledQueryRulesAlertTrigger(action.Trigger)); err != nil {
		return fmt.Errorf("setting `trigger`: %+v", err)
	}

	if schedule := resp.Schedule; schedule != nil {
		if schedule.FrequencyInMinutes != nil {
			d.Set("frequency", schedule.FrequencyInMinutes)
		}
		if schedule.TimeWindowInMinutes != nil {
			d.Set("time_window", schedule.TimeWindowInMinutes)
		}
	}

	if source := resp.Source; source != nil {
		if source.AuthorizedResources != nil {
			d.Set("authorized_resource_ids", utils.FlattenStringSlice(source.AuthorizedResources))
		}
		if source.DataSourceID != nil {
			d.Set("data_source_id", source.DataSourceID)
		}
		if source.Query != nil {
			d.Set("query", source.Query)
		}
		d.Set("query_type", string(source.QueryType))
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceMonitorScheduledQueryRulesAlertDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.ScheduledQueryRulesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ScheduledQueryRulesID(d.Id())
	if err != nil {
		return err
	}

	if resp, err := client.Delete(ctx, id.ResourceGroup, id.ScheduledQueryRuleName); err != nil {
		if !response.WasNotFound(resp.Response) {
			return fmt.Errorf("deleting Monitor %s: %+v", *id, err)
		}
	}

	return nil
}

func expandMonitorScheduledQueryRulesAlertingAction(d *pluginsdk.ResourceData) *insights.AlertingAction {
	alertActionRaw := d.Get("action").([]interface{})
	alertAction := expandMonitorScheduledQueryRulesAlertAction(alertActionRaw)
	severityRaw := d.Get("severity").(int)
	severity := strconv.Itoa(severityRaw)

	triggerRaw := d.Get("trigger").([]interface{})
	trigger := expandMonitorScheduledQueryRulesAlertTrigger(triggerRaw)

	action := insights.AlertingAction{
		AznsAction: alertAction,
		Severity:   insights.AlertSeverity(severity),
		Trigger:    trigger,
		OdataType:  insights.OdataTypeBasicActionOdataTypeMicrosoftWindowsAzureManagementMonitoringAlertsModelsMicrosoftAppInsightsNexusDataContractsResourcesScheduledQueryRulesAlertingAction,
	}

	if throttling, ok := d.Get("throttling").(int); ok && throttling != 0 {
		action.ThrottlingInMin = utils.Int32(int32(throttling))
	}

	return &action
}

func expandMonitorScheduledQueryRulesAlertAction(input []interface{}) *insights.AzNsActionGroup {
	result := insights.AzNsActionGroup{}

	if len(input) == 0 {
		return &result
	}
	for _, item := range input {
		if item == nil {
			continue
		}

		v, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		actionGroups := v["action_group"].(*pluginsdk.Set).List()
		result.ActionGroup = utils.ExpandStringSlice(actionGroups)
		result.EmailSubject = utils.String(v["email_subject"].(string))
		if v := v["custom_webhook_payload"].(string); v != "" {
			result.CustomWebhookPayload = utils.String(v)
		}
	}

	return &result
}

func expandMonitorScheduledQueryRulesAlertMetricTrigger(input []interface{}) *insights.LogMetricTrigger {
	if len(input) == 0 {
		return nil
	}

	result := insights.LogMetricTrigger{}
	for _, item := range input {
		if item == nil {
			continue
		}
		v, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		result.ThresholdOperator = insights.ConditionalOperator(v["operator"].(string))
		result.Threshold = utils.Float(v["threshold"].(float64))
		result.MetricTriggerType = insights.MetricTriggerType(v["metric_trigger_type"].(string))
		result.MetricColumn = utils.String(v["metric_column"].(string))
	}

	return &result
}

func expandMonitorScheduledQueryRulesAlertSchedule(d *pluginsdk.ResourceData) *insights.Schedule {
	frequency := d.Get("frequency").(int)
	timeWindow := d.Get("time_window").(int)

	schedule := insights.Schedule{
		FrequencyInMinutes:  utils.Int32(int32(frequency)),
		TimeWindowInMinutes: utils.Int32(int32(timeWindow)),
	}

	return &schedule
}

func expandMonitorScheduledQueryRulesAlertTrigger(input []interface{}) *insights.TriggerCondition {
	result := insights.TriggerCondition{}
	if len(input) == 0 {
		return &result
	}

	for _, item := range input {
		if item == nil {
			continue
		}
		v, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		metricTriggerRaw := v["metric_trigger"].([]interface{})

		result.ThresholdOperator = insights.ConditionalOperator(v["operator"].(string))
		result.Threshold = utils.Float(v["threshold"].(float64))
		result.MetricTrigger = expandMonitorScheduledQueryRulesAlertMetricTrigger(metricTriggerRaw)
	}

	return &result
}
