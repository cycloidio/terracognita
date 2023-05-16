package monitor

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/monitor/mgmt/2020-10-01/insights"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/monitor/migration"
	"github.com/hashicorp/terraform-provider-azurerm/services/monitor/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceMonitorActivityLogAlert() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceMonitorActivityLogAlertCreateUpdate,
		Read:   resourceMonitorActivityLogAlertRead,
		Update: resourceMonitorActivityLogAlertCreateUpdate,
		Delete: resourceMonitorActivityLogAlertDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ActivityLogAlertID(id)
			return err
		}),

		SchemaVersion: 1,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.ActivityLogAlertUpgradeV0ToV1{},
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
					ValidateFunc: validation.StringIsNotEmpty,
				},
				Set: pluginsdk.HashString,
			},

			"criteria": {
				Type:     pluginsdk.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"category": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Administrative",
								"Autoscale",
								"Policy",
								"Recommendation",
								"ResourceHealth",
								"Security",
								"ServiceHealth",
							}, false),
						},
						"operation_name": {
							Type:     pluginsdk.TypeString,
							Optional: true,
						},
						"caller": {
							Type:     pluginsdk.TypeString,
							Optional: true,
						},
						"level": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Verbose",
								"Informational",
								"Warning",
								"Error",
								"Critical",
							}, false),
						},
						"resource_provider": {
							Type:     pluginsdk.TypeString,
							Optional: true,
						},
						"resource_type": {
							Type:     pluginsdk.TypeString,
							Optional: true,
						},
						"resource_group": {
							Type:     pluginsdk.TypeString,
							Optional: true,
						},
						"resource_id": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: azure.ValidateResourceID,
						},
						"status": {
							Type:     pluginsdk.TypeString,
							Optional: true,
						},
						"sub_status": {
							Type:     pluginsdk.TypeString,
							Optional: true,
						},
						"recommendation_category": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Cost",
								"Reliability",
								"OperationalExcellence",
								"Performance",
							},
								false,
							),
							ConflictsWith: []string{"criteria.0.recommendation_type"},
						},
						"recommendation_impact": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"High",
								"Medium",
								"Low",
							},
								false,
							),
							ConflictsWith: []string{"criteria.0.recommendation_type"},
						},
						"recommendation_type": {
							Type:          pluginsdk.TypeString,
							Optional:      true,
							ConflictsWith: []string{"criteria.0.recommendation_category", "criteria.0.recommendation_impact"},
						},
						//lintignore:XS003
						"resource_health": {
							Type:     pluginsdk.TypeList,
							Computed: true,
							Optional: true,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"current": {
										Type:     pluginsdk.TypeSet,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
											ValidateFunc: validation.StringInSlice([]string{
												"Available",
												"Degraded",
												"Unavailable",
												"Unknown",
											},
												false,
											),
										},
										Set: pluginsdk.HashString,
									},
									"previous": {
										Type:     pluginsdk.TypeSet,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
											ValidateFunc: validation.StringInSlice([]string{
												"Available",
												"Degraded",
												"Unavailable",
												"Unknown",
											},
												false,
											),
										},
										Set: pluginsdk.HashString,
									},
									"reason": {
										Type:     pluginsdk.TypeSet,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
											ValidateFunc: validation.StringInSlice([]string{
												"PlatformInitiated",
												"UserInitiated",
												"Unknown",
											},
												false,
											),
										},
										Set: pluginsdk.HashString,
									},
								},
							},
							ConflictsWith: []string{"criteria.0.recommendation_category", "criteria.0.recommendation_impact", "criteria.0.status", "criteria.0.sub_status", "criteria.0.recommendation_impact", "criteria.0.resource_provider", "criteria.0.resource_type", "criteria.0.operation_name", "criteria.0.caller", "criteria.0.operation_name", "criteria.0.service_health"},
						},
						//lintignore:XS003
						"service_health": {
							Type:     pluginsdk.TypeList,
							Computed: true,
							Optional: true,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"events": {
										Type:     pluginsdk.TypeSet,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
											ValidateFunc: validation.StringInSlice([]string{
												"Incident",
												"Maintenance",
												"Informational",
												"ActionRequired",
												"Security",
											},
												false,
											),
										},
										Set: pluginsdk.HashString,
									},
									"locations": {
										Type:     pluginsdk.TypeSet,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type:         pluginsdk.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
										Set: pluginsdk.HashString,
									},
									"services": {
										Type:     pluginsdk.TypeSet,
										Optional: true,
										Elem: &pluginsdk.Schema{
											Type:         pluginsdk.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
										Set: pluginsdk.HashString,
									},
								},
							},
							ConflictsWith: []string{"criteria.0.recommendation_category", "criteria.0.recommendation_impact", "criteria.0.status", "criteria.0.sub_status", "criteria.0.recommendation_impact", "criteria.0.resource_provider", "criteria.0.resource_type", "criteria.0.operation_name", "criteria.0.caller", "criteria.0.operation_name", "criteria.0.resource_health"},
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
				Set: resourceMonitorActivityLogAlertActionHash,
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

			"tags": tags.Schema(),
		},
	}
}

func resourceMonitorActivityLogAlertCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.ActivityLogAlertsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewActivityLogAlertID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Monitor %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_monitor_activity_log_alert", id.ID())
		}
	}

	enabled := d.Get("enabled").(bool)
	description := d.Get("description").(string)
	scopesRaw := d.Get("scopes").(*pluginsdk.Set).List()
	criteriaRaw := d.Get("criteria").([]interface{})
	actionRaw := d.Get("action").(*pluginsdk.Set).List()

	t := d.Get("tags").(map[string]interface{})
	expandedTags := tags.Expand(t)

	parameters := insights.ActivityLogAlertResource{
		Location: utils.String(azure.NormalizeLocation("Global")),
		AlertRuleProperties: &insights.AlertRuleProperties{
			Enabled:     utils.Bool(enabled),
			Description: utils.String(description),
			Scopes:      utils.ExpandStringSlice(scopesRaw),
			Condition:   expandMonitorActivityLogAlertCriteria(criteriaRaw),
			Actions:     expandMonitorActivityLogAlertAction(actionRaw),
		},
		Tags: expandedTags,
	}

	if _, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, parameters); err != nil {
		return fmt.Errorf("creating or updating Monitor %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceMonitorActivityLogAlertRead(d, meta)
}

func resourceMonitorActivityLogAlertRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.ActivityLogAlertsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ActivityLogAlertID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Activity Log Alert %q was not found in Resource Group %q - removing from state!", id.Name, id.ResourceGroup)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("getting Monitor %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	if alert := resp.AlertRuleProperties; alert != nil {
		d.Set("enabled", alert.Enabled)
		d.Set("description", alert.Description)
		if err := d.Set("scopes", utils.FlattenStringSlice(alert.Scopes)); err != nil {
			return fmt.Errorf("setting `scopes`: %+v", err)
		}
		if err := d.Set("criteria", flattenMonitorActivityLogAlertCriteria(alert.Condition)); err != nil {
			return fmt.Errorf("setting `criteria`: %+v", err)
		}
		if err := d.Set("action", flattenMonitorActivityLogAlertAction(alert.Actions)); err != nil {
			return fmt.Errorf("setting `action`: %+v", err)
		}
	}
	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceMonitorActivityLogAlertDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.ActivityLogAlertsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ActivityLogAlertID(d.Id())
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

func expandMonitorActivityLogAlertCriteria(input []interface{}) *insights.AlertRuleAllOfCondition {
	conditions := make([]insights.AlertRuleAnyOfOrLeafCondition, 0)
	v := input[0].(map[string]interface{})

	if category := v["category"].(string); category != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("category"),
			Equals: utils.String(category),
		})
	}
	if op := v["operation_name"].(string); op != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("operationName"),
			Equals: utils.String(op),
		})
	}
	if caller := v["caller"].(string); caller != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("caller"),
			Equals: utils.String(caller),
		})
	}
	if level := v["level"].(string); level != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("level"),
			Equals: utils.String(level),
		})
	}
	if resourceProvider := v["resource_provider"].(string); resourceProvider != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("resourceProvider"),
			Equals: utils.String(resourceProvider),
		})
	}
	if resourceType := v["resource_type"].(string); resourceType != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("resourceType"),
			Equals: utils.String(resourceType),
		})
	}
	if resourceGroup := v["resource_group"].(string); resourceGroup != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("resourceGroup"),
			Equals: utils.String(resourceGroup),
		})
	}
	if id := v["resource_id"].(string); id != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("resourceId"),
			Equals: utils.String(id),
		})
	}
	if status := v["status"].(string); status != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("status"),
			Equals: utils.String(status),
		})
	}
	if subStatus := v["sub_status"].(string); subStatus != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("subStatus"),
			Equals: utils.String(subStatus),
		})
	}
	if recommendationType := v["recommendation_type"].(string); recommendationType != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("properties.recommendationType"),
			Equals: utils.String(recommendationType),
		})
	}

	if recommendationCategory := v["recommendation_category"].(string); recommendationCategory != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("properties.recommendationCategory"),
			Equals: utils.String(recommendationCategory),
		})
	}

	if recommendationImpact := v["recommendation_impact"].(string); recommendationImpact != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("properties.recommendationImpact"),
			Equals: utils.String(recommendationImpact),
		})
	}

	if resourceHealth := v["resource_health"].([]interface{}); len(resourceHealth) > 0 {
		conditions = expandResourceHealth(resourceHealth, conditions)
	}

	if serviceHealth := v["service_health"].([]interface{}); len(serviceHealth) > 0 {
		conditions = expandServiceHealth(serviceHealth, conditions)
	}

	return &insights.AlertRuleAllOfCondition{
		AllOf: &conditions,
	}
}

func expandResourceHealth(resourceHealth []interface{}, conditions []insights.AlertRuleAnyOfOrLeafCondition) []insights.AlertRuleAnyOfOrLeafCondition {
	for _, serviceItem := range resourceHealth {
		if serviceItem == nil {
			continue
		}
		vs := serviceItem.(map[string]interface{})

		cv := vs["current"].(*pluginsdk.Set)
		if len(cv.List()) > 0 {
			ruleLeafCondition := make([]insights.AlertRuleLeafCondition, 0)
			for _, e := range cv.List() {
				event := e.(string)
				ruleLeafCondition = append(ruleLeafCondition, insights.AlertRuleLeafCondition{
					Field:  utils.String("properties.currentHealthStatus"),
					Equals: utils.String(event),
				})
			}
			conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
				AnyOf: &ruleLeafCondition,
			})
		}

		pv := vs["previous"].(*pluginsdk.Set)
		if len(pv.List()) > 0 {
			ruleLeafCondition := make([]insights.AlertRuleLeafCondition, 0)
			for _, e := range pv.List() {
				event := e.(string)
				ruleLeafCondition = append(ruleLeafCondition, insights.AlertRuleLeafCondition{
					Field:  utils.String("properties.previousHealthStatus"),
					Equals: utils.String(event),
				})
			}
			conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
				AnyOf: &ruleLeafCondition,
			})
		}

		rv := vs["reason"].(*pluginsdk.Set)
		if len(rv.List()) > 0 {
			ruleLeafCondition := make([]insights.AlertRuleLeafCondition, 0)
			for _, e := range rv.List() {
				event := e.(string)
				ruleLeafCondition = append(ruleLeafCondition, insights.AlertRuleLeafCondition{
					Field:  utils.String("properties.cause"),
					Equals: utils.String(event),
				})
			}
			conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
				AnyOf: &ruleLeafCondition,
			})
		}
	}
	return conditions
}

func expandServiceHealth(serviceHealth []interface{}, conditions []insights.AlertRuleAnyOfOrLeafCondition) []insights.AlertRuleAnyOfOrLeafCondition {
	for _, serviceItem := range serviceHealth {
		if serviceItem == nil {
			continue
		}
		vs := serviceItem.(map[string]interface{})
		rv := vs["locations"].(*pluginsdk.Set)
		if len(rv.List()) > 0 {
			conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
				Field:       utils.String("properties.impactedServices[*].ImpactedRegions[*].RegionName"),
				ContainsAny: utils.ExpandStringSlice(rv.List()),
			})
		}

		ev := vs["events"].(*pluginsdk.Set)
		if len(ev.List()) > 0 {
			ruleLeafCondition := make([]insights.AlertRuleLeafCondition, 0)
			for _, e := range ev.List() {
				event := e.(string)
				ruleLeafCondition = append(ruleLeafCondition, insights.AlertRuleLeafCondition{
					Field:  utils.String("properties.incidentType"),
					Equals: utils.String(event),
				})
			}
			conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
				AnyOf: &ruleLeafCondition,
			})
		}

		sv := vs["services"].(*pluginsdk.Set)
		if len(sv.List()) > 0 {
			conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
				Field:       utils.String("properties.impactedServices[*].ServiceName"),
				ContainsAny: utils.ExpandStringSlice(sv.List()),
			})
		}
	}
	return conditions
}

func expandMonitorActivityLogAlertAction(input []interface{}) *insights.ActionList {
	actions := make([]insights.ActionGroup, 0)
	for _, item := range input {
		v := item.(map[string]interface{})
		if agID := v["action_group_id"].(string); agID != "" {
			props := make(map[string]*string)
			if pVal, ok := v["webhook_properties"]; ok {
				for pk, pv := range pVal.(map[string]interface{}) {
					props[pk] = utils.String(pv.(string))
				}
			}

			actions = append(actions, insights.ActionGroup{
				ActionGroupID:     utils.String(agID),
				WebhookProperties: props,
			})
		}
	}
	return &insights.ActionList{
		ActionGroups: &actions,
	}
}

func flattenMonitorActivityLogAlertCriteria(input *insights.AlertRuleAllOfCondition) []interface{} {
	result := make(map[string]interface{})
	if input == nil || input.AllOf == nil {
		return []interface{}{result}
	}
	for _, condition := range *input.AllOf {
		if condition.Field != nil && condition.Equals != nil {
			switch strings.ToLower(*condition.Field) {
			case "operationname":
				result["operation_name"] = *condition.Equals
			case "resourceprovider":
				result["resource_provider"] = *condition.Equals
			case "resourcetype":
				result["resource_type"] = *condition.Equals
			case "resourcegroup":
				result["resource_group"] = *condition.Equals
			case "resourceid":
				result["resource_id"] = *condition.Equals
			case "substatus":
				result["sub_status"] = *condition.Equals
			case "properties.recommendationtype":
				result["recommendation_type"] = *condition.Equals
			case "properties.recommendationcategory":
				result["recommendation_category"] = *condition.Equals
			case "properties.recommendationimpact":
				result["recommendation_impact"] = *condition.Equals
			case "caller", "category", "level", "status":
				result[*condition.Field] = *condition.Equals
			}
		}
	}

	if result["category"] == "ResourceHealth" {
		flattenMonitorActivityLogAlertResourceHealth(input, result)
	}

	if result["category"] == "ServiceHealth" {
		flattenMonitorActivityLogAlertServiceHealth(input, result)
	}

	return []interface{}{result}
}

func flattenMonitorActivityLogAlertResourceHealth(input *insights.AlertRuleAllOfCondition, result map[string]interface{}) {
	rhResult := make(map[string]interface{})
	for _, condition := range *input.AllOf {
		if condition.Field == nil && len(*condition.AnyOf) > 0 {
			values := []string{}
			for _, cond := range *condition.AnyOf {
				if cond.Field != nil && cond.Equals != nil {
					values = append(values, *cond.Equals)
				}
				switch strings.ToLower(*cond.Field) {
				case "properties.currenthealthstatus":
					rhResult["current"] = values
				case "properties.previoushealthstatus":
					rhResult["previous"] = values
				case "properties.cause":
					rhResult["reason"] = values
				}
			}
		}
	}

	result["resource_health"] = []interface{}{rhResult}
}

func flattenMonitorActivityLogAlertServiceHealth(input *insights.AlertRuleAllOfCondition, result map[string]interface{}) {
	shResult := make(map[string]interface{})
	for _, condition := range *input.AllOf {
		if condition.Field != nil && condition.ContainsAny != nil && len(*condition.ContainsAny) > 0 {
			switch strings.ToLower(*condition.Field) {
			case "properties.impactedservices[*].impactedregions[*].regionname":
				shResult["locations"] = *condition.ContainsAny
			case "properties.impactedservices[*].servicename":
				shResult["services"] = *condition.ContainsAny
			}
		}
		if condition.Field == nil && len(*condition.AnyOf) > 0 {
			events := []string{}
			for _, evCond := range *condition.AnyOf {
				if evCond.Field != nil && evCond.Equals != nil {
					events = append(events, *evCond.Equals)
				}
			}
			shResult["events"] = events
		}
	}

	result["service_health"] = []interface{}{shResult}
}

func flattenMonitorActivityLogAlertAction(input *insights.ActionList) (result []interface{}) {
	result = make([]interface{}, 0)
	if input == nil || input.ActionGroups == nil {
		return
	}
	for _, action := range *input.ActionGroups {
		v := make(map[string]interface{})

		if action.ActionGroupID != nil {
			v["action_group_id"] = *action.ActionGroupID
		}

		props := make(map[string]interface{})
		for pk, pv := range action.WebhookProperties {
			if pv != nil {
				props[pk] = *pv
			}
		}
		v["webhook_properties"] = props

		result = append(result, v)
	}
	return result
}

func resourceMonitorActivityLogAlertActionHash(input interface{}) int {
	var buf bytes.Buffer
	if v, ok := input.(map[string]interface{}); ok {
		buf.WriteString(fmt.Sprintf("%s-", v["action_group_id"].(string)))
	}
	return pluginsdk.HashString(buf.String())
}
