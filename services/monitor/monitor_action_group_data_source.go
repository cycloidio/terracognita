package monitor

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	eventHubValidation "github.com/hashicorp/terraform-provider-azurerm/services/eventhub/validate"
	"github.com/hashicorp/terraform-provider-azurerm/services/monitor/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func dataSourceMonitorActionGroup() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceMonitorActionGroupRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"resource_group_name": commonschema.ResourceGroupNameForDataSource(),

			"short_name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"enabled": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"email_receiver": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"email_address": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"use_common_alert_schema": {
							Type:     pluginsdk.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"itsm_receiver": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"workspace_id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"connection_id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"ticket_configuration": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"region": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"azure_app_push_receiver": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"email_address": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"sms_receiver": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"country_code": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"phone_number": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"webhook_receiver": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"service_uri": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"use_common_alert_schema": {
							Type:     pluginsdk.TypeBool,
							Computed: true,
						},
						"aad_auth": {
							Type:     pluginsdk.TypeList,
							Computed: true,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"object_id": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},

									"identifier_uri": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},

									"tenant_id": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},

			"automation_runbook_receiver": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"automation_account_id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"runbook_name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"webhook_resource_id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"is_global_runbook": {
							Type:     pluginsdk.TypeBool,
							Computed: true,
						},
						"service_uri": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"use_common_alert_schema": {
							Type:     pluginsdk.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"voice_receiver": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"country_code": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"phone_number": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"logic_app_receiver": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"resource_id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"callback_url": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"use_common_alert_schema": {
							Type:     pluginsdk.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"azure_function_receiver": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"function_app_resource_id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"function_name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"http_trigger_url": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"use_common_alert_schema": {
							Type:     pluginsdk.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"arm_role_receiver": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"role_id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"use_common_alert_schema": {
							Type:     pluginsdk.TypeBool,
							Computed: true,
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
		},
	}
}

func dataSourceMonitorActionGroupRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.ActionGroupsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceGroup := d.Get("resource_group_name").(string)

	id := parse.NewActionGroupID(subscriptionId, resourceGroup, d.Get("name").(string))

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("%s was not found", id)
		}
		return fmt.Errorf("making Read request on %s: %+v", id, err)
	}
	d.SetId(id.ID())

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
		if err = d.Set("event_hub_receiver", flattenMonitorActionGroupEventHubReceiver(resourceGroup, group.EventHubReceivers)); err != nil {
			return fmt.Errorf("setting `event_hub_receiver`: %+v", err)
		}
	}

	return nil
}
