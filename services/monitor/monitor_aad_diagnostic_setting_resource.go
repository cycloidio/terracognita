package monitor

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/aad/mgmt/2017-04-01/aad"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	authRuleParse "github.com/hashicorp/terraform-provider-azurerm/services/eventhub/sdk/2017-04-01/authorizationrulesnamespaces"
	logAnalyticsParse "github.com/hashicorp/terraform-provider-azurerm/services/loganalytics/parse"
	logAnalyticsValidate "github.com/hashicorp/terraform-provider-azurerm/services/loganalytics/validate"
	"github.com/hashicorp/terraform-provider-azurerm/services/monitor/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/monitor/validate"
	storageParse "github.com/hashicorp/terraform-provider-azurerm/services/storage/parse"
	storageValidate "github.com/hashicorp/terraform-provider-azurerm/services/storage/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceMonitorAADDiagnosticSetting() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceMonitorAADDiagnosticSettingCreateUpdate,
		Read:   resourceMonitorAADDiagnosticSettingRead,
		Update: resourceMonitorAADDiagnosticSettingCreateUpdate,
		Delete: resourceMonitorAADDiagnosticSettingDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.MonitorAADDiagnosticSettingID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(5 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(5 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.MonitorDiagnosticSettingName,
			},

			// When absent, will use the default eventhub, whilst the Diagnostic Setting API will return this property as an empty string. Therefore, it is useless to make this property as Computed.
			"eventhub_name": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[a-zA-Z0-9]([-._a-zA-Z0-9]{0,48}[a-zA-Z0-9])?$"),
					"The event hub name can contain only letters, numbers, periods (.), hyphens (-),and underscores (_), up to 50 characters, and it must begin and end with a letter or number.",
				),
			},

			"eventhub_authorization_rule_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: authRuleParse.ValidateAuthorizationRuleID,
				AtLeastOneOf: []string{"eventhub_authorization_rule_id", "log_analytics_workspace_id", "storage_account_id"},
			},

			"log_analytics_workspace_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: logAnalyticsValidate.LogAnalyticsWorkspaceID,
				AtLeastOneOf: []string{"eventhub_authorization_rule_id", "log_analytics_workspace_id", "storage_account_id"},
			},

			"storage_account_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: storageValidate.StorageAccountID,
				AtLeastOneOf: []string{"eventhub_authorization_rule_id", "log_analytics_workspace_id", "storage_account_id"},
			},

			"log": {
				Type:     pluginsdk.TypeSet,
				Required: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"category": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},

						"enabled": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
							Default:  true,
						},

						"retention_policy": {
							Type:     pluginsdk.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"enabled": {
										Type:     pluginsdk.TypeBool,
										Optional: true,
										Default:  false,
									},

									"days": {
										Type:         pluginsdk.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
										Default:      0,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceMonitorAADDiagnosticSettingCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.AADDiagnosticSettingsClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	log.Printf("[INFO] preparing arguments for Azure ARM AAD Diagnostic Setting.")

	id := parse.NewMonitorAADDiagnosticSettingID(d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %s", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_monitor_aad_diagnostic_setting", id.ID())
		}
	}

	logs := expandMonitorAADDiagnosticsSettingsLogs(d.Get("log").(*pluginsdk.Set).List())

	// If there is no `enabled` log entry, the PUT will succeed while the next GET will return a 404.
	// Therefore, ensure users has at least one enabled log entry.
	valid := false
	for _, log := range logs {
		if log.Enabled != nil && *log.Enabled {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("At least one of the `log` of the %s should be enabled", id)
	}

	properties := aad.DiagnosticSettingsResource{
		DiagnosticSettings: &aad.DiagnosticSettings{
			Logs: &logs,
		},
	}

	eventHubAuthorizationRuleId := d.Get("eventhub_authorization_rule_id").(string)
	eventHubName := d.Get("eventhub_name").(string)
	if eventHubAuthorizationRuleId != "" {
		properties.DiagnosticSettings.EventHubAuthorizationRuleID = utils.String(eventHubAuthorizationRuleId)
		properties.DiagnosticSettings.EventHubName = utils.String(eventHubName)
	}

	workspaceId := d.Get("log_analytics_workspace_id").(string)
	if workspaceId != "" {
		properties.DiagnosticSettings.WorkspaceID = utils.String(workspaceId)
	}

	storageAccountId := d.Get("storage_account_id").(string)
	if storageAccountId != "" {
		properties.DiagnosticSettings.StorageAccountID = utils.String(storageAccountId)
	}

	if _, err := client.CreateOrUpdate(ctx, properties, id.Name); err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceMonitorAADDiagnosticSettingRead(d, meta)
}

func resourceMonitorAADDiagnosticSettingRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.AADDiagnosticSettingsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.MonitorAADDiagnosticSettingID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[WARN] %s was not found - removing from state!", id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.Set("name", id.Name)

	d.Set("eventhub_name", resp.EventHubName)
	eventhubAuthorizationRuleId := ""
	if resp.EventHubAuthorizationRuleID != nil && *resp.EventHubAuthorizationRuleID != "" {
		parsedId, err := authRuleParse.ParseAuthorizationRuleID(*resp.EventHubAuthorizationRuleID)
		if err != nil {
			return err
		}

		eventhubAuthorizationRuleId = parsedId.ID()
	}
	d.Set("eventhub_authorization_rule_id", eventhubAuthorizationRuleId)

	workspaceId := ""
	if resp.WorkspaceID != nil && *resp.WorkspaceID != "" {
		parsedId, err := logAnalyticsParse.LogAnalyticsWorkspaceID(*resp.WorkspaceID)
		if err != nil {
			return err
		}

		workspaceId = parsedId.ID()
	}
	d.Set("log_analytics_workspace_id", workspaceId)

	storageAccountId := ""
	if resp.StorageAccountID != nil && *resp.StorageAccountID != "" {
		parsedId, err := storageParse.StorageAccountID(*resp.StorageAccountID)
		if err != nil {
			return err
		}

		storageAccountId = parsedId.ID()
	}
	d.Set("storage_account_id", storageAccountId)

	if err := d.Set("log", flattenMonitorAADDiagnosticLogs(resp.Logs)); err != nil {
		return fmt.Errorf("setting `log`: %+v", err)
	}

	return nil
}

func resourceMonitorAADDiagnosticSettingDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.AADDiagnosticSettingsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.MonitorAADDiagnosticSettingID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Delete(ctx, id.Name)
	if err != nil {
		if !response.WasNotFound(resp.Response) {
			return fmt.Errorf("deleting %s: %+v", id, err)
		}
	}

	// API appears to be eventually consistent (identified during tainting this resource)
	log.Printf("[DEBUG] Waiting for %s to disappear", id)
	timeout, _ := ctx.Deadline()
	stateConf := &pluginsdk.StateChangeConf{
		Pending:                   []string{"Exists"},
		Target:                    []string{"NotFound"},
		Refresh:                   monitorAADDiagnosticSettingDeletedRefreshFunc(ctx, client, id.Name),
		MinTimeout:                15 * time.Second,
		ContinuousTargetOccurence: 5,
		Timeout:                   time.Until(timeout),
	}

	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for %s to become available: %s", id, err)
	}

	return nil
}

func monitorAADDiagnosticSettingDeletedRefreshFunc(ctx context.Context, client *aad.DiagnosticSettingsClient, name string) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		res, err := client.Get(ctx, name)
		if err != nil {
			if utils.ResponseWasNotFound(res.Response) {
				return "NotFound", "NotFound", nil
			}
			return nil, "", fmt.Errorf("issuing read request in monitorAADDiagnosticSettingDeletedRefreshFunc: %s", err)
		}

		return res, "Exists", nil
	}
}

func expandMonitorAADDiagnosticsSettingsLogs(input []interface{}) []aad.LogSettings {
	results := make([]aad.LogSettings, 0)

	for _, raw := range input {
		v := raw.(map[string]interface{})

		category := v["category"].(string)
		enabled := v["enabled"].(bool)

		policyRaw := v["retention_policy"].([]interface{})[0].(map[string]interface{})
		retentionDays := policyRaw["days"].(int)
		retentionEnabled := policyRaw["enabled"].(bool)

		output := aad.LogSettings{
			Category: aad.Category(category),
			Enabled:  utils.Bool(enabled),
			RetentionPolicy: &aad.RetentionPolicy{
				Days:    utils.Int32(int32(retentionDays)),
				Enabled: utils.Bool(retentionEnabled),
			},
		}

		results = append(results, output)
	}

	return results
}

func flattenMonitorAADDiagnosticLogs(input *[]aad.LogSettings) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	for _, v := range *input {
		category := string(v.Category)

		enabled := false
		if v.Enabled != nil {
			enabled = *v.Enabled
		}

		policies := make([]interface{}, 0)
		if inputPolicy := v.RetentionPolicy; inputPolicy != nil {
			days := 0
			if inputPolicy.Days != nil {
				days = int(*inputPolicy.Days)
			}

			enabled := false
			if inputPolicy.Enabled != nil {
				enabled = *inputPolicy.Enabled
			}

			policies = append(policies, map[string]interface{}{
				"days":    days,
				"enabled": enabled,
			})
		}

		results = append(results, map[string]interface{}{
			"category":         category,
			"enabled":          enabled,
			"retention_policy": policies,
		})
	}

	return results
}
