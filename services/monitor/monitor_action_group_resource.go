package monitor

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2021-09-01-preview/insights"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	eventHubParser "github.com/hashicorp/terraform-provider-azurerm/services/eventhub/parse"
	eventHubValidation "github.com/hashicorp/terraform-provider-azurerm/services/eventhub/validate"
	"github.com/hashicorp/terraform-provider-azurerm/services/monitor/migration"
	"github.com/hashicorp/terraform-provider-azurerm/services/monitor/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/monitor/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceMonitorActionGroup() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceMonitorActionGroupCreateUpdate,
		Read:   resourceMonitorActionGroupRead,
		Update: resourceMonitorActionGroupCreateUpdate,
		Delete: resourceMonitorActionGroupDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ActionGroupID(id)
			return err
		}),

		SchemaVersion: 1,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.ActionGroupUpgradeV0ToV1{},
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

			"short_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 12),
			},

			"enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"email_receiver": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"email_address": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"use_common_alert_schema": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
						},
					},
				},
			},

			"itsm_receiver": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"workspace_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validate.WorkspaceID,
						},
						"connection_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.IsUUID,
						},
						"ticket_configuration": {
							Type:             pluginsdk.TypeString,
							Required:         true,
							ValidateFunc:     validation.StringIsJSON,
							DiffSuppressFunc: pluginsdk.SuppressJsonDiff,
						},
						"region": {
							Type:             pluginsdk.TypeString,
							Required:         true,
							ValidateFunc:     validation.StringIsNotEmpty,
							DiffSuppressFunc: location.DiffSuppressFunc,
						},
					},
				},
			},

			"azure_app_push_receiver": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"email_address": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"sms_receiver": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"country_code": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"phone_number": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"webhook_receiver": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"service_uri": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.IsURLWithScheme([]string{"http", "https"}),
						},
						"use_common_alert_schema": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
						},

						"aad_auth": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"object_id": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.IsUUID,
									},

									"identifier_uri": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IsURLWithScheme([]string{"api"}),
									},

									"tenant_id": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IsUUID,
									},
								},
							},
						},
					},
				},
			},

			"automation_runbook_receiver": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"automation_account_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: azure.ValidateResourceID,
						},
						"runbook_name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"webhook_resource_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"is_global_runbook": {
							Type:     pluginsdk.TypeBool,
							Required: true,
						},
						"service_uri": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.IsURLWithScheme([]string{"http", "https"}),
						},
						"use_common_alert_schema": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			"voice_receiver": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"country_code": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"phone_number": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"logic_app_receiver": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"resource_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"callback_url": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.IsURLWithScheme([]string{"http", "https"}),
						},
						"use_common_alert_schema": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
						},
					},
				},
			},

			"azure_function_receiver": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"function_app_resource_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: azure.ValidateResourceID,
						},
						"function_name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"http_trigger_url": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.IsURLWithScheme([]string{"http", "https"}),
						},
						"use_common_alert_schema": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
						},
					},
				},
			},

			"arm_role_receiver": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"role_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.IsUUID,
						},
						"use_common_alert_schema": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
						},
					},
				},
			},

			"event_hub_receiver": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"event_hub_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: eventHubValidation.EventhubID,
						},
						"tenant_id": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsUUID,
						},
						"use_common_alert_schema": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"tags": tags.Schema(),
		},
	}
}

func resourceMonitorActionGroupCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.ActionGroupsClient
	tenantId := meta.(*clients.Client).Account.TenantId
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewActionGroupID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Monitor %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_monitor_action_group", id.ID())
		}
	}

	shortName := d.Get("short_name").(string)
	enabled := d.Get("enabled").(bool)

	emailReceiversRaw := d.Get("email_receiver").([]interface{})
	itsmReceiversRaw := d.Get("itsm_receiver").([]interface{})
	azureAppPushReceiversRaw := d.Get("azure_app_push_receiver").([]interface{})
	smsReceiversRaw := d.Get("sms_receiver").([]interface{})
	webhookReceiversRaw := d.Get("webhook_receiver").([]interface{})
	automationRunbookReceiversRaw := d.Get("automation_runbook_receiver").([]interface{})
	voiceReceiversRaw := d.Get("voice_receiver").([]interface{})
	logicAppReceiversRaw := d.Get("logic_app_receiver").([]interface{})
	azureFunctionReceiversRaw := d.Get("azure_function_receiver").([]interface{})
	armRoleReceiversRaw := d.Get("arm_role_receiver").([]interface{})
	eventHubReceiversRaw := d.Get("event_hub_receiver").([]interface{})

	expandedEventHubReceiver, err := expandMonitorActionGroupEventHubReceiver(tenantId, eventHubReceiversRaw)
	if err != nil {
		return err
	}

	t := d.Get("tags").(map[string]interface{})
	expandedTags := tags.Expand(t)

	parameters := insights.ActionGroupResource{
		Location: utils.String(azure.NormalizeLocation("Global")),
		ActionGroup: &insights.ActionGroup{
			GroupShortName:             utils.String(shortName),
			Enabled:                    utils.Bool(enabled),
			EmailReceivers:             expandMonitorActionGroupEmailReceiver(emailReceiversRaw),
			AzureAppPushReceivers:      expandMonitorActionGroupAzureAppPushReceiver(azureAppPushReceiversRaw),
			ItsmReceivers:              expandMonitorActionGroupItsmReceiver(itsmReceiversRaw),
			SmsReceivers:               expandMonitorActionGroupSmsReceiver(smsReceiversRaw),
			WebhookReceivers:           expandMonitorActionGroupWebHookReceiver(tenantId, webhookReceiversRaw),
			AutomationRunbookReceivers: expandMonitorActionGroupAutomationRunbookReceiver(automationRunbookReceiversRaw),
			VoiceReceivers:             expandMonitorActionGroupVoiceReceiver(voiceReceiversRaw),
			LogicAppReceivers:          expandMonitorActionGroupLogicAppReceiver(logicAppReceiversRaw),
			AzureFunctionReceivers:     expandMonitorActionGroupAzureFunctionReceiver(azureFunctionReceiversRaw),
			ArmRoleReceivers:           expandMonitorActionGroupRoleReceiver(armRoleReceiversRaw),
			EventHubReceivers:          expandedEventHubReceiver,
		},
		Tags: expandedTags,
	}

	if _, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, parameters); err != nil {
		return fmt.Errorf("creating or updating %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceMonitorActionGroupRead(d, meta)
}

func resourceMonitorActionGroupRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.ActionGroupsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ActionGroupID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if response.WasNotFound(resp.Response.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)

	if group := resp.ActionGroup; group != nil {
		d.Set("short_name", group.GroupShortName)
		d.Set("enabled", group.Enabled)

		if err = d.Set("email_receiver", flattenMonitorActionGroupEmailReceiver(group.EmailReceivers)); err != nil {
			return fmt.Errorf("setting `email_receiver`: %+v", err)
		}

		if err = d.Set("itsm_receiver", flattenMonitorActionGroupItsmReceiver(group.ItsmReceivers)); err != nil {
			return fmt.Errorf("setting `itsm_receiver`: %+v", err)
		}

		if err = d.Set("azure_app_push_receiver", flattenMonitorActionGroupAzureAppPushReceiver(group.AzureAppPushReceivers)); err != nil {
			return fmt.Errorf("setting `azure_app_push_receiver`: %+v", err)
		}

		if err = d.Set("sms_receiver", flattenMonitorActionGroupSmsReceiver(group.SmsReceivers)); err != nil {
			return fmt.Errorf("setting `sms_receiver`: %+v", err)
		}

		if err = d.Set("webhook_receiver", flattenMonitorActionGroupWebHookReceiver(group.WebhookReceivers)); err != nil {
			return fmt.Errorf("setting `webhook_receiver`: %+v", err)
		}

		if err = d.Set("automation_runbook_receiver", flattenMonitorActionGroupAutomationRunbookReceiver(group.AutomationRunbookReceivers)); err != nil {
			return fmt.Errorf("setting `automation_runbook_receiver`: %+v", err)
		}

		if err = d.Set("voice_receiver", flattenMonitorActionGroupVoiceReceiver(group.VoiceReceivers)); err != nil {
			return fmt.Errorf("setting `voice_receiver`: %+v", err)
		}

		if err = d.Set("logic_app_receiver", flattenMonitorActionGroupLogicAppReceiver(group.LogicAppReceivers)); err != nil {
			return fmt.Errorf("setting `logic_app_receiver`: %+v", err)
		}

		if err = d.Set("azure_function_receiver", flattenMonitorActionGroupAzureFunctionReceiver(group.AzureFunctionReceivers)); err != nil {
			return fmt.Errorf("setting `azure_function_receiver`: %+v", err)
		}
		if err = d.Set("arm_role_receiver", flattenMonitorActionGroupRoleReceiver(group.ArmRoleReceivers)); err != nil {
			return fmt.Errorf("setting `arm_role_receiver`: %+v", err)
		}
		if err = d.Set("event_hub_receiver", flattenMonitorActionGroupEventHubReceiver(id.ResourceGroup, group.EventHubReceivers)); err != nil {
			return fmt.Errorf("setting `event_hub_receiver`: %+v", err)
		}
	}
	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceMonitorActionGroupDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.ActionGroupsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ActionGroupID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if !response.WasNotFound(resp.Response) {
			return fmt.Errorf("deleting %s: %+v", *id, err)
		}
	}

	return nil
}

func expandMonitorActionGroupEmailReceiver(v []interface{}) *[]insights.EmailReceiver {
	receivers := make([]insights.EmailReceiver, 0)
	for _, receiverValue := range v {
		val := receiverValue.(map[string]interface{})
		receiver := insights.EmailReceiver{
			Name:                 utils.String(val["name"].(string)),
			EmailAddress:         utils.String(val["email_address"].(string)),
			UseCommonAlertSchema: utils.Bool(val["use_common_alert_schema"].(bool)),
		}
		receivers = append(receivers, receiver)
	}
	return &receivers
}

func expandMonitorActionGroupItsmReceiver(v []interface{}) *[]insights.ItsmReceiver {
	receivers := make([]insights.ItsmReceiver, 0)
	for _, receiverValue := range v {
		val := receiverValue.(map[string]interface{})
		receiver := insights.ItsmReceiver{
			Name:                utils.String(val["name"].(string)),
			WorkspaceID:         utils.String(val["workspace_id"].(string)),
			ConnectionID:        utils.String(val["connection_id"].(string)),
			TicketConfiguration: utils.String(val["ticket_configuration"].(string)),
			Region:              utils.String(azure.NormalizeLocation(val["region"].(string))),
		}
		receivers = append(receivers, receiver)
	}
	return &receivers
}

func expandMonitorActionGroupAzureAppPushReceiver(v []interface{}) *[]insights.AzureAppPushReceiver {
	receivers := make([]insights.AzureAppPushReceiver, 0)
	for _, receiverValue := range v {
		val := receiverValue.(map[string]interface{})
		receiver := insights.AzureAppPushReceiver{
			Name:         utils.String(val["name"].(string)),
			EmailAddress: utils.String(val["email_address"].(string)),
		}
		receivers = append(receivers, receiver)
	}
	return &receivers
}

func expandMonitorActionGroupSmsReceiver(v []interface{}) *[]insights.SmsReceiver {
	receivers := make([]insights.SmsReceiver, 0)
	for _, receiverValue := range v {
		val := receiverValue.(map[string]interface{})
		receiver := insights.SmsReceiver{
			Name:        utils.String(val["name"].(string)),
			CountryCode: utils.String(val["country_code"].(string)),
			PhoneNumber: utils.String(val["phone_number"].(string)),
		}
		receivers = append(receivers, receiver)
	}
	return &receivers
}

func expandMonitorActionGroupWebHookReceiver(tenantId string, v []interface{}) *[]insights.WebhookReceiver {
	receivers := make([]insights.WebhookReceiver, 0)
	for _, receiverValue := range v {
		val := receiverValue.(map[string]interface{})
		receiver := insights.WebhookReceiver{
			Name:                 utils.String(val["name"].(string)),
			ServiceURI:           utils.String(val["service_uri"].(string)),
			UseCommonAlertSchema: utils.Bool(val["use_common_alert_schema"].(bool)),
		}
		if v, ok := val["aad_auth"].([]interface{}); ok && len(v) > 0 {
			secureWebhook := v[0].(map[string]interface{})
			receiver.UseAadAuth = utils.Bool(true)
			receiver.ObjectID = utils.String(secureWebhook["object_id"].(string))
			receiver.IdentifierURI = utils.String(secureWebhook["identifier_uri"].(string))
			if v := secureWebhook["tenant_id"].(string); v != "" {
				receiver.TenantID = utils.String(v)
			} else {
				receiver.TenantID = utils.String(tenantId)
			}
		}
		receivers = append(receivers, receiver)
	}
	return &receivers
}

func expandMonitorActionGroupAutomationRunbookReceiver(v []interface{}) *[]insights.AutomationRunbookReceiver {
	receivers := make([]insights.AutomationRunbookReceiver, 0)
	for _, receiverValue := range v {
		val := receiverValue.(map[string]interface{})
		receiver := insights.AutomationRunbookReceiver{
			Name:                 utils.String(val["name"].(string)),
			AutomationAccountID:  utils.String(val["automation_account_id"].(string)),
			RunbookName:          utils.String(val["runbook_name"].(string)),
			WebhookResourceID:    utils.String(val["webhook_resource_id"].(string)),
			IsGlobalRunbook:      utils.Bool(val["is_global_runbook"].(bool)),
			ServiceURI:           utils.String(val["service_uri"].(string)),
			UseCommonAlertSchema: utils.Bool(val["use_common_alert_schema"].(bool)),
		}
		receivers = append(receivers, receiver)
	}
	return &receivers
}

func expandMonitorActionGroupVoiceReceiver(v []interface{}) *[]insights.VoiceReceiver {
	receivers := make([]insights.VoiceReceiver, 0)
	for _, receiverValue := range v {
		val := receiverValue.(map[string]interface{})
		receiver := insights.VoiceReceiver{
			Name:        utils.String(val["name"].(string)),
			CountryCode: utils.String(val["country_code"].(string)),
			PhoneNumber: utils.String(val["phone_number"].(string)),
		}
		receivers = append(receivers, receiver)
	}
	return &receivers
}

func expandMonitorActionGroupLogicAppReceiver(v []interface{}) *[]insights.LogicAppReceiver {
	receivers := make([]insights.LogicAppReceiver, 0)
	for _, receiverValue := range v {
		val := receiverValue.(map[string]interface{})
		receiver := insights.LogicAppReceiver{
			Name:                 utils.String(val["name"].(string)),
			ResourceID:           utils.String(val["resource_id"].(string)),
			CallbackURL:          utils.String(val["callback_url"].(string)),
			UseCommonAlertSchema: utils.Bool(val["use_common_alert_schema"].(bool)),
		}
		receivers = append(receivers, receiver)
	}
	return &receivers
}

func expandMonitorActionGroupAzureFunctionReceiver(v []interface{}) *[]insights.AzureFunctionReceiver {
	receivers := make([]insights.AzureFunctionReceiver, 0)
	for _, receiverValue := range v {
		val := receiverValue.(map[string]interface{})
		receiver := insights.AzureFunctionReceiver{
			Name:                  utils.String(val["name"].(string)),
			FunctionAppResourceID: utils.String(val["function_app_resource_id"].(string)),
			FunctionName:          utils.String(val["function_name"].(string)),
			HTTPTriggerURL:        utils.String(val["http_trigger_url"].(string)),
			UseCommonAlertSchema:  utils.Bool(val["use_common_alert_schema"].(bool)),
		}
		receivers = append(receivers, receiver)
	}
	return &receivers
}

func expandMonitorActionGroupRoleReceiver(v []interface{}) *[]insights.ArmRoleReceiver {
	receivers := make([]insights.ArmRoleReceiver, 0)
	for _, receiverValue := range v {
		val := receiverValue.(map[string]interface{})
		receiver := insights.ArmRoleReceiver{
			Name:                 utils.String(val["name"].(string)),
			RoleID:               utils.String(val["role_id"].(string)),
			UseCommonAlertSchema: utils.Bool(val["use_common_alert_schema"].(bool)),
		}
		receivers = append(receivers, receiver)
	}
	return &receivers
}

func expandMonitorActionGroupEventHubReceiver(tenantId string, v []interface{}) (*[]insights.EventHubReceiver, error) {
	receivers := make([]insights.EventHubReceiver, 0)
	for _, receiverValue := range v {
		val := receiverValue.(map[string]interface{})

		eventHubId, err := eventHubParser.EventhubID(*utils.String(val["event_hub_id"].(string)))
		if err != nil {
			return nil, err
		}

		receiver := insights.EventHubReceiver{
			Name:                 utils.String(val["name"].(string)),
			EventHubNameSpace:    &eventHubId.NamespaceName,
			EventHubName:         &eventHubId.Name,
			UseCommonAlertSchema: utils.Bool(val["use_common_alert_schema"].(bool)),
		}
		if v := val["tenant_id"].(string); v != "" {
			receiver.TenantID = utils.String(v)
		} else {
			receiver.TenantID = utils.String(tenantId)
		}
		receiver.SubscriptionID = &eventHubId.SubscriptionId
		receivers = append(receivers, receiver)
	}
	return &receivers, nil
}

func flattenMonitorActionGroupEmailReceiver(receivers *[]insights.EmailReceiver) []interface{} {
	result := make([]interface{}, 0)
	if receivers != nil {
		for _, receiver := range *receivers {
			val := make(map[string]interface{})
			if receiver.Name != nil {
				val["name"] = *receiver.Name
			}
			if receiver.EmailAddress != nil {
				val["email_address"] = *receiver.EmailAddress
			}
			if receiver.UseCommonAlertSchema != nil {
				val["use_common_alert_schema"] = *receiver.UseCommonAlertSchema
			}
			result = append(result, val)
		}
	}
	return result
}

func flattenMonitorActionGroupItsmReceiver(receivers *[]insights.ItsmReceiver) []interface{} {
	result := make([]interface{}, 0)
	if receivers != nil {
		for _, receiver := range *receivers {
			val := make(map[string]interface{})
			if receiver.Name != nil {
				val["name"] = *receiver.Name
			}
			if receiver.WorkspaceID != nil {
				val["workspace_id"] = *receiver.WorkspaceID
			}
			if receiver.ConnectionID != nil {
				val["connection_id"] = *receiver.ConnectionID
			}
			if receiver.TicketConfiguration != nil {
				val["ticket_configuration"] = *receiver.TicketConfiguration
			}
			if receiver.Region != nil {
				val["region"] = azure.NormalizeLocation(*receiver.Region)
			}
			result = append(result, val)
		}
	}
	return result
}

func flattenMonitorActionGroupAzureAppPushReceiver(receivers *[]insights.AzureAppPushReceiver) []interface{} {
	result := make([]interface{}, 0)
	if receivers != nil {
		for _, receiver := range *receivers {
			val := make(map[string]interface{})
			if receiver.Name != nil {
				val["name"] = *receiver.Name
			}
			if receiver.EmailAddress != nil {
				val["email_address"] = *receiver.EmailAddress
			}
			result = append(result, val)
		}
	}
	return result
}

func flattenMonitorActionGroupSmsReceiver(receivers *[]insights.SmsReceiver) []interface{} {
	result := make([]interface{}, 0)
	if receivers != nil {
		for _, receiver := range *receivers {
			val := make(map[string]interface{})
			if receiver.Name != nil {
				val["name"] = *receiver.Name
			}
			if receiver.CountryCode != nil {
				val["country_code"] = *receiver.CountryCode
			}
			if receiver.PhoneNumber != nil {
				val["phone_number"] = *receiver.PhoneNumber
			}

			result = append(result, val)
		}
	}
	return result
}

func flattenMonitorActionGroupWebHookReceiver(receivers *[]insights.WebhookReceiver) []interface{} {
	result := make([]interface{}, 0)
	if receivers != nil {
		for _, receiver := range *receivers {
			var useCommonAlert bool
			var name, serviceUri string
			if receiver.Name != nil {
				name = *receiver.Name
			}
			if receiver.ServiceURI != nil {
				serviceUri = *receiver.ServiceURI
			}
			if receiver.UseCommonAlertSchema != nil {
				useCommonAlert = *receiver.UseCommonAlertSchema
			}

			result = append(result, map[string]interface{}{
				"name":                    name,
				"service_uri":             serviceUri,
				"use_common_alert_schema": useCommonAlert,
				"aad_auth":                flattenMonitorActionGroupSecureWebHookReceiver(receiver),
			})
		}
	}
	return result
}

func flattenMonitorActionGroupSecureWebHookReceiver(receiver insights.WebhookReceiver) []interface{} {
	if receiver.UseAadAuth == nil || !*receiver.UseAadAuth {
		return []interface{}{}
	}

	var objectId, identifierUri, tenantId string

	if v := receiver.ObjectID; v != nil {
		objectId = *v
	}
	if v := receiver.IdentifierURI; v != nil {
		identifierUri = *v
	}
	if v := receiver.TenantID; v != nil {
		tenantId = *v
	}
	return []interface{}{
		map[string]interface{}{
			"object_id":      objectId,
			"identifier_uri": identifierUri,
			"tenant_id":      tenantId,
		},
	}
}

func flattenMonitorActionGroupAutomationRunbookReceiver(receivers *[]insights.AutomationRunbookReceiver) []interface{} {
	result := make([]interface{}, 0)
	if receivers != nil {
		for _, receiver := range *receivers {
			val := make(map[string]interface{})
			if receiver.Name != nil {
				val["name"] = *receiver.Name
			}
			if receiver.AutomationAccountID != nil {
				val["automation_account_id"] = *receiver.AutomationAccountID
			}
			if receiver.RunbookName != nil {
				val["runbook_name"] = *receiver.RunbookName
			}
			if receiver.WebhookResourceID != nil {
				val["webhook_resource_id"] = *receiver.WebhookResourceID
			}
			if receiver.IsGlobalRunbook != nil {
				val["is_global_runbook"] = *receiver.IsGlobalRunbook
			}
			if receiver.ServiceURI != nil {
				val["service_uri"] = *receiver.ServiceURI
			}
			if receiver.UseCommonAlertSchema != nil {
				val["use_common_alert_schema"] = *receiver.UseCommonAlertSchema
			}
			result = append(result, val)
		}
	}
	return result
}

func flattenMonitorActionGroupVoiceReceiver(receivers *[]insights.VoiceReceiver) []interface{} {
	result := make([]interface{}, 0)
	if receivers != nil {
		for _, receiver := range *receivers {
			val := make(map[string]interface{})
			if receiver.Name != nil {
				val["name"] = *receiver.Name
			}
			if receiver.CountryCode != nil {
				val["country_code"] = *receiver.CountryCode
			}
			if receiver.PhoneNumber != nil {
				val["phone_number"] = *receiver.PhoneNumber
			}
			result = append(result, val)
		}
	}
	return result
}

func flattenMonitorActionGroupLogicAppReceiver(receivers *[]insights.LogicAppReceiver) []interface{} {
	result := make([]interface{}, 0)
	if receivers != nil {
		for _, receiver := range *receivers {
			val := make(map[string]interface{})
			if receiver.Name != nil {
				val["name"] = *receiver.Name
			}
			if receiver.ResourceID != nil {
				val["resource_id"] = *receiver.ResourceID
			}
			if receiver.CallbackURL != nil {
				val["callback_url"] = *receiver.CallbackURL
			}
			if receiver.UseCommonAlertSchema != nil {
				val["use_common_alert_schema"] = *receiver.UseCommonAlertSchema
			}
			result = append(result, val)
		}
	}
	return result
}

func flattenMonitorActionGroupAzureFunctionReceiver(receivers *[]insights.AzureFunctionReceiver) []interface{} {
	result := make([]interface{}, 0)
	if receivers != nil {
		for _, receiver := range *receivers {
			val := make(map[string]interface{})
			if receiver.Name != nil {
				val["name"] = *receiver.Name
			}
			if receiver.FunctionAppResourceID != nil {
				val["function_app_resource_id"] = *receiver.FunctionAppResourceID
			}
			if receiver.FunctionName != nil {
				val["function_name"] = *receiver.FunctionName
			}
			if receiver.HTTPTriggerURL != nil {
				val["http_trigger_url"] = *receiver.HTTPTriggerURL
			}
			if receiver.UseCommonAlertSchema != nil {
				val["use_common_alert_schema"] = *receiver.UseCommonAlertSchema
			}
			result = append(result, val)
		}
	}
	return result
}

func flattenMonitorActionGroupRoleReceiver(receivers *[]insights.ArmRoleReceiver) []interface{} {
	result := make([]interface{}, 0)
	if receivers != nil {
		for _, receiver := range *receivers {
			val := make(map[string]interface{})
			if receiver.Name != nil {
				val["name"] = *receiver.Name
			}
			if receiver.RoleID != nil {
				val["role_id"] = *receiver.RoleID
			}
			if receiver.UseCommonAlertSchema != nil {
				val["use_common_alert_schema"] = *receiver.UseCommonAlertSchema
			}
			result = append(result, val)
		}
	}
	return result
}

func flattenMonitorActionGroupEventHubReceiver(resourceGroup string, receivers *[]insights.EventHubReceiver) []interface{} {
	result := make([]interface{}, 0)
	if receivers != nil {
		for _, receiver := range *receivers {
			val := make(map[string]interface{})
			if receiver.Name != nil {
				val["name"] = *receiver.Name
			}
			if receiver.EventHubNameSpace != nil && receiver.EventHubName != nil && receiver.SubscriptionID != nil {
				event_hub_namespace := *receiver.EventHubNameSpace
				event_hub_name := *receiver.EventHubName
				subscription_id := *receiver.SubscriptionID

				val["event_hub_id"] = eventHubParser.NewEventhubID(subscription_id, resourceGroup, event_hub_namespace, event_hub_name).ID()
			}
			if receiver.UseCommonAlertSchema != nil {
				val["use_common_alert_schema"] = *receiver.UseCommonAlertSchema
			}
			if receiver.TenantID != nil {
				val["tenant_id"] = *receiver.TenantID
			}
			result = append(result, val)
		}
	}
	return result
}
