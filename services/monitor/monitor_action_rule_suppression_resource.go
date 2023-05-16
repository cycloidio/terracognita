package monitor

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/alertsmanagement/mgmt/2019-06-01-preview/alertsmanagement"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/monitor/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/monitor/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceMonitorActionRuleSuppression() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceMonitorActionRuleSuppressionCreateUpdate,
		Read:   resourceMonitorActionRuleSuppressionRead,
		Update: resourceMonitorActionRuleSuppressionCreateUpdate,
		Delete: resourceMonitorActionRuleSuppressionDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceIdThen(func(id string) error {
			_, err := parse.ActionRuleID(id)
			return err
		}, importMonitorActionRule(alertsmanagement.TypeSuppression)),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ActionRuleName,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"suppression": {
				Type:     pluginsdk.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"recurrence_type": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(alertsmanagement.Always),
								string(alertsmanagement.Once),
								string(alertsmanagement.Daily),
								string(alertsmanagement.Weekly),
								string(alertsmanagement.Monthly),
							}, false),
						},

						"schedule": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"start_date_utc": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.IsRFC3339Time,
									},

									"end_date_utc": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.IsRFC3339Time,
									},

									"recurrence_weekly": {
										Type:          pluginsdk.TypeSet,
										Optional:      true,
										MinItems:      1,
										ConflictsWith: []string{"suppression.0.schedule.0.recurrence_monthly"},
										Elem: &pluginsdk.Schema{
											Type:         pluginsdk.TypeString,
											ValidateFunc: validation.IsDayOfTheWeek(false),
										},
									},

									"recurrence_monthly": {
										Type:          pluginsdk.TypeSet,
										Optional:      true,
										MinItems:      1,
										ConflictsWith: []string{"suppression.0.schedule.0.recurrence_weekly"},
										Elem: &pluginsdk.Schema{
											Type:         pluginsdk.TypeInt,
											ValidateFunc: validation.IntBetween(1, 31),
										},
									},
								},
							},
						},
					},
				},
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

			"condition": schemaActionRuleConditions(),

			"scope": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"type": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(alertsmanagement.ScopeTypeResourceGroup),
								string(alertsmanagement.ScopeTypeResource),
							}, false),
						},

						"resource_ids": {
							Type:     pluginsdk.TypeSet,
							Required: true,
							Elem: &pluginsdk.Schema{
								Type:         pluginsdk.TypeString,
								ValidateFunc: azure.ValidateResourceID,
							},
						},
					},
				},
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceMonitorActionRuleSuppressionCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.ActionRulesClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewActionRuleID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.GetByName(ctx, id.ResourceGroup, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Monitor %s: %+v", id, err)
			}
		}
		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_monitor_action_rule_suppression", id.ID())
		}
	}

	actionRuleStatus := alertsmanagement.Enabled
	if !d.Get("enabled").(bool) {
		actionRuleStatus = alertsmanagement.Disabled
	}

	suppressionConfig, err := expandActionRuleSuppressionConfig(d.Get("suppression").([]interface{}))
	if err != nil {
		return err
	}

	actionRule := alertsmanagement.ActionRule{
		// the location is always global from the portal
		Location: utils.String(location.Normalize("Global")),
		Properties: &alertsmanagement.Suppression{
			SuppressionConfig: suppressionConfig,
			Scope:             expandActionRuleScope(d.Get("scope").([]interface{})),
			Conditions:        expandActionRuleConditions(d.Get("condition").([]interface{})),
			Description:       utils.String(d.Get("description").(string)),
			Status:            actionRuleStatus,
			Type:              alertsmanagement.TypeSuppression,
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	if _, err := client.CreateUpdate(ctx, id.ResourceGroup, id.Name, actionRule); err != nil {
		return fmt.Errorf("creating/updating Monitor %s: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceMonitorActionRuleSuppressionRead(d, meta)
}

func resourceMonitorActionRuleSuppressionRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.ActionRulesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ActionRuleID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.GetByName(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Action Rule %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Monitor %s: %+v", *id, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	if resp.Properties != nil {
		props, _ := resp.Properties.AsSuppression()
		d.Set("description", props.Description)
		d.Set("enabled", props.Status == alertsmanagement.Enabled)
		if err := d.Set("suppression", flattenActionRuleSuppression(props.SuppressionConfig)); err != nil {
			return fmt.Errorf("setting suppression: %+v", err)
		}
		if err := d.Set("scope", flattenActionRuleScope(props.Scope)); err != nil {
			return fmt.Errorf("setting scope: %+v", err)
		}
		if err := d.Set("condition", flattenActionRuleConditions(props.Conditions)); err != nil {
			return fmt.Errorf("setting condition: %+v", err)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceMonitorActionRuleSuppressionDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.ActionRulesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ActionRuleID(d.Id())
	if err != nil {
		return err
	}

	if _, err := client.Delete(ctx, id.ResourceGroup, id.Name); err != nil {
		return fmt.Errorf("deleting Monitor %s: %+v", *id, err)
	}
	return nil
}

func expandActionRuleSuppressionConfig(input []interface{}) (*alertsmanagement.SuppressionConfig, error) {
	if len(input) == 0 {
		return nil, nil
	}
	v := input[0].(map[string]interface{})
	recurrenceType := alertsmanagement.SuppressionType(v["recurrence_type"].(string))
	schedule, err := expandActionRuleSuppressionSchedule(v["schedule"].([]interface{}), recurrenceType)
	if err != nil {
		return nil, err
	}
	if recurrenceType != alertsmanagement.Always && schedule == nil {
		return nil, fmt.Errorf("`schedule` block must be set when `recurrence_type` is Once, Daily, Weekly or Monthly.")
	}
	return &alertsmanagement.SuppressionConfig{
		RecurrenceType: recurrenceType,
		Schedule:       schedule,
	}, nil
}

func expandActionRuleSuppressionSchedule(input []interface{}, suppressionType alertsmanagement.SuppressionType) (*alertsmanagement.SuppressionSchedule, error) {
	if len(input) == 0 {
		return nil, nil
	}
	v := input[0].(map[string]interface{})

	var recurrence []interface{}
	switch suppressionType {
	case alertsmanagement.Weekly:
		if recurrenceWeekly, ok := v["recurrence_weekly"]; ok {
			recurrence = expandActionRuleSuppressionScheduleRecurrenceWeekly(recurrenceWeekly.(*pluginsdk.Set).List())
		}
		if len(recurrence) == 0 {
			return nil, fmt.Errorf("`recurrence_weekly` must be set and should have at least one element when `recurrence_type` is Weekly.")
		}
	case alertsmanagement.Monthly:
		if recurrenceMonthly, ok := v["recurrence_monthly"]; ok {
			recurrence = recurrenceMonthly.(*pluginsdk.Set).List()
		}
		if len(recurrence) == 0 {
			return nil, fmt.Errorf("`recurrence_monthly` must be set and should have at least one element when `recurrence_type` is Monthly.")
		}
	}

	startDateUTC, _ := time.Parse(time.RFC3339, v["start_date_utc"].(string))
	endDateUTC, _ := time.Parse(time.RFC3339, v["end_date_utc"].(string))
	return &alertsmanagement.SuppressionSchedule{
		StartDate:        utils.String(startDateUTC.Format(scheduleDateLayout)),
		EndDate:          utils.String(endDateUTC.Format(scheduleDateLayout)),
		StartTime:        utils.String(startDateUTC.Format(scheduleTimeLayout)),
		EndTime:          utils.String(endDateUTC.Format(scheduleTimeLayout)),
		RecurrenceValues: utils.ExpandInt32Slice(recurrence),
	}, nil
}

func expandActionRuleSuppressionScheduleRecurrenceWeekly(input []interface{}) []interface{} {
	result := make([]interface{}, 0, len(input))
	for _, v := range input {
		result = append(result, weekDayMap[v.(string)])
	}
	return result
}

func flattenActionRuleSuppression(input *alertsmanagement.SuppressionConfig) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var recurrenceType alertsmanagement.SuppressionType
	if input.RecurrenceType != "" {
		recurrenceType = input.RecurrenceType
	}
	return []interface{}{
		map[string]interface{}{
			"recurrence_type": string(recurrenceType),
			"schedule":        flattenActionRuleSuppressionSchedule(input.Schedule, recurrenceType),
		},
	}
}

func flattenActionRuleSuppressionSchedule(input *alertsmanagement.SuppressionSchedule, recurrenceType alertsmanagement.SuppressionType) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	startDateUTCStr := ""
	endDateUTCStr := ""
	recurrenceWeekly := []interface{}{}
	recurrenceMonthly := []interface{}{}

	if input.StartDate != nil && input.StartTime != nil {
		date, _ := time.ParseInLocation(scheduleDateTimeLayout, fmt.Sprintf("%s %s", *input.StartDate, *input.StartTime), time.UTC)
		startDateUTCStr = date.Format(time.RFC3339)
	}
	if input.EndDate != nil && input.EndTime != nil {
		date, _ := time.ParseInLocation(scheduleDateTimeLayout, fmt.Sprintf("%s %s", *input.EndDate, *input.EndTime), time.UTC)
		endDateUTCStr = date.Format(time.RFC3339)
	}

	if recurrenceType == alertsmanagement.Weekly {
		recurrenceWeekly = flattenActionRuleSuppressionScheduleRecurrenceWeekly(input.RecurrenceValues)
	}
	if recurrenceType == alertsmanagement.Monthly {
		recurrenceMonthly = utils.FlattenInt32Slice(input.RecurrenceValues)
	}
	return []interface{}{
		map[string]interface{}{
			"start_date_utc":     startDateUTCStr,
			"end_date_utc":       endDateUTCStr,
			"recurrence_weekly":  recurrenceWeekly,
			"recurrence_monthly": recurrenceMonthly,
		},
	}
}

func flattenActionRuleSuppressionScheduleRecurrenceWeekly(input *[]int32) []interface{} {
	result := make([]interface{}, 0)
	if input != nil {
		for _, item := range *input {
			result = append(result, weekDays[int(item)])
		}
	}
	return result
}
