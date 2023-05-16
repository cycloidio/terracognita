package recoveryservices

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/recoveryservices/mgmt/2021-12-01/backup"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/recoveryservices/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/recoveryservices/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/set"
	"github.com/hashicorp/terraform-provider-azurerm/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceBackupProtectionPolicyFileShare() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceBackupProtectionPolicyFileShareCreateUpdate,
		Read:   resourceBackupProtectionPolicyFileShareRead,
		Update: resourceBackupProtectionPolicyFileShareCreateUpdate,
		Delete: resourceBackupProtectionPolicyFileShareDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.BackupPolicyID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: resourceBackupProtectionPolicyFileShareSchema(),

		// if daily, we need daily retention
		// if weekly daily cannot be set, and we need weekly
		CustomizeDiff: func(ctx context.Context, diff *pluginsdk.ResourceDiff, v interface{}) error {
			_, hasDaily := diff.GetOk("retention_daily")
			_, hasWeekly := diff.GetOk("retention_weekly")

			frequencyI, _ := diff.GetOk("backup.0.frequency")
			switch strings.ToLower(frequencyI.(string)) {
			case "daily":
				if !hasDaily {
					return fmt.Errorf("`retention_daily` must be set when backup.0.frequency is daily")
				}

				if _, ok := diff.GetOk("backup.0.weekdays"); ok {
					return fmt.Errorf("`backup.0.weekdays` should be not set when backup.0.frequency is daily")
				}
			case "weekly":
				if hasDaily {
					return fmt.Errorf("`retention_daily` must be not set when backup.0.frequency is weekly")
				}
				if !hasWeekly {
					return fmt.Errorf("`retention_weekly` must be set when backup.0.frequency is weekly")
				}
			default:
				return fmt.Errorf("Unrecognized value for backup.0.frequency")
			}
			return nil
		},
	}
}

func resourceBackupProtectionPolicyFileShareCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).RecoveryServices.ProtectionPoliciesClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	policyName := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	vaultName := d.Get("recovery_vault_name").(string)

	log.Printf("[DEBUG] Creating/updating Recovery Service Protection Policy %s (resource group %q)", policyName, resourceGroup)

	// getting this ready now because its shared between *everything*, time is... complicated for this resource
	timeOfDay := d.Get("backup.0.time").(string)
	dateOfDay, err := time.Parse(time.RFC3339, fmt.Sprintf("2018-07-30T%s:00Z", timeOfDay))
	if err != nil {
		return fmt.Errorf("generating time from %q for policy %q (Resource Group %q): %+v", timeOfDay, policyName, resourceGroup, err)
	}
	times := append(make([]date.Time, 0), date.Time{Time: dateOfDay})

	if d.IsNewResource() {
		existing, err2 := client.Get(ctx, vaultName, resourceGroup, policyName)
		if err2 != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Recovery Service Protection Policy %q (Resource Group %q): %+v", policyName, resourceGroup, err2)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_backup_policy_file_share", *existing.ID)
		}
	}

	AzureFileShareProtectionPolicyProperties := &backup.AzureFileShareProtectionPolicy{
		TimeZone:             utils.String(d.Get("timezone").(string)),
		BackupManagementType: backup.ManagementTypeBasicProtectionPolicyBackupManagementTypeAzureStorage,
		WorkLoadType:         backup.WorkloadTypeAzureFileShare,
		SchedulePolicy:       expandBackupProtectionPolicyFileShareSchedule(d, times),
		RetentionPolicy: &backup.LongTermRetentionPolicy{ // SimpleRetentionPolicy only has duration property ¯\_(ツ)_/¯
			RetentionPolicyType: backup.RetentionPolicyTypeLongTermRetentionPolicy,
			DailySchedule:       expandBackupProtectionPolicyFileShareRetentionDaily(d, times),
			WeeklySchedule:      expandBackupProtectionPolicyFileShareRetentionWeekly(d, times),
			MonthlySchedule:     expandBackupProtectionPolicyFileShareRetentionMonthly(d, times),
			YearlySchedule:      expandBackupProtectionPolicyFileShareRetentionYearly(d, times),
		},
	}

	policy := backup.ProtectionPolicyResource{
		Properties: AzureFileShareProtectionPolicyProperties,
	}

	if _, err = client.CreateOrUpdate(ctx, vaultName, resourceGroup, policyName, policy); err != nil {
		return fmt.Errorf("creating/updating Recovery Service Protection Policy %q (Resource Group %q): %+v", policyName, resourceGroup, err)
	}

	resp, err := resourceBackupProtectionPolicyFileShareWaitForUpdate(ctx, client, vaultName, resourceGroup, policyName, d)
	if err != nil {
		return err
	}

	id := strings.Replace(*resp.ID, "Subscriptions", "subscriptions", 1)
	d.SetId(id)

	return resourceBackupProtectionPolicyFileShareRead(d, meta)
}

func resourceBackupProtectionPolicyFileShareRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).RecoveryServices.ProtectionPoliciesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.BackupPolicyID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Reading Recovery Service Protection Policy %q (resource group %q)", id.Name, id.ResourceGroup)

	resp, err := client.Get(ctx, id.VaultName, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("making Read request on Recovery Service Protection Policy %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("recovery_vault_name", id.VaultName)

	if properties, ok := resp.Properties.AsAzureFileShareProtectionPolicy(); ok && properties != nil {
		d.Set("timezone", properties.TimeZone)

		if schedule, ok := properties.SchedulePolicy.AsSimpleSchedulePolicy(); ok && schedule != nil {
			if err := d.Set("backup", flattenBackupProtectionPolicyFileShareSchedule(schedule)); err != nil {
				return fmt.Errorf("setting `backup`: %+v", err)
			}
		}

		if retention, ok := properties.RetentionPolicy.AsLongTermRetentionPolicy(); ok && retention != nil {
			if s := retention.DailySchedule; s != nil {
				if err := d.Set("retention_daily", flattenBackupProtectionPolicyFileShareRetentionDaily(s)); err != nil {
					return fmt.Errorf("setting `retention_daily`: %+v", err)
				}
			} else {
				d.Set("retention_daily", nil)
			}

			if s := retention.WeeklySchedule; s != nil {
				if err := d.Set("retention_weekly", flattenBackupProtectionPolicyFileShareRetentionWeekly(s)); err != nil {
					return fmt.Errorf("setting `retention_weekly`: %+v", err)
				}
			} else {
				d.Set("retention_weekly", nil)
			}

			if s := retention.MonthlySchedule; s != nil {
				if err := d.Set("retention_monthly", flattenBackupProtectionPolicyFileShareRetentionMonthly(s)); err != nil {
					return fmt.Errorf("setting `retention_monthly`: %+v", err)
				}
			} else {
				d.Set("retention_monthly", nil)
			}

			if s := retention.YearlySchedule; s != nil {
				if err := d.Set("retention_yearly", flattenBackupProtectionPolicyFileShareRetentionYearly(s)); err != nil {
					return fmt.Errorf("setting `retention_yearly`: %+v", err)
				}
			} else {
				d.Set("retention_yearly", nil)
			}
		}
	}

	return nil
}

func resourceBackupProtectionPolicyFileShareDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).RecoveryServices.ProtectionPoliciesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.BackupPolicyID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting Recovery Service Protection Policy %q (resource group %q)", id.Name, id.ResourceGroup)

	future, err := client.Delete(ctx, id.VaultName, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
	}

	resp, err := future.Result(*client)
	if err != nil {
		if !utils.ResponseWasNotFound(resp) {
			return fmt.Errorf("issuing delete request for Recovery Service Protection Policy %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}
	}

	if _, err := resourceBackupProtectionPolicyFileShareWaitForDeletion(ctx, client, id.VaultName, id.ResourceGroup, id.Name, d); err != nil {
		return err
	}

	return nil
}

func expandBackupProtectionPolicyFileShareSchedule(d *pluginsdk.ResourceData, times []date.Time) *backup.SimpleSchedulePolicy {
	if bb, ok := d.Get("backup").([]interface{}); ok && len(bb) > 0 {
		block := bb[0].(map[string]interface{})

		schedule := backup.SimpleSchedulePolicy{ // LongTermSchedulePolicy has no properties
			SchedulePolicyType: backup.SchedulePolicyTypeSimpleSchedulePolicy,
			ScheduleRunTimes:   &times,
		}

		if v, ok := block["frequency"].(string); ok {
			schedule.ScheduleRunFrequency = backup.ScheduleRunType(v)
		}

		return &schedule
	}

	return nil
}

func expandBackupProtectionPolicyFileShareRetentionDaily(d *pluginsdk.ResourceData, times []date.Time) *backup.DailyRetentionSchedule {
	if rb, ok := d.Get("retention_daily").([]interface{}); ok && len(rb) > 0 {
		block := rb[0].(map[string]interface{})

		return &backup.DailyRetentionSchedule{
			RetentionTimes: &times,
			RetentionDuration: &backup.RetentionDuration{
				Count:        utils.Int32(int32(block["count"].(int))),
				DurationType: backup.RetentionDurationTypeDays,
			},
		}
	}

	return nil
}

func expandBackupProtectionPolicyFileShareRetentionWeekly(d *pluginsdk.ResourceData, times []date.Time) *backup.WeeklyRetentionSchedule {
	if rb, ok := d.Get("retention_weekly").([]interface{}); ok && len(rb) > 0 {
		block := rb[0].(map[string]interface{})

		retention := backup.WeeklyRetentionSchedule{
			RetentionTimes: &times,
			RetentionDuration: &backup.RetentionDuration{
				Count:        utils.Int32(int32(block["count"].(int))),
				DurationType: backup.RetentionDurationTypeWeeks,
			},
		}

		if v, ok := block["weekdays"].(*pluginsdk.Set); ok {
			days := make([]backup.DayOfWeek, 0)
			for _, day := range v.List() {
				days = append(days, backup.DayOfWeek(day.(string)))
			}
			retention.DaysOfTheWeek = &days
		}

		return &retention
	}

	return nil
}

func expandBackupProtectionPolicyFileShareRetentionMonthly(d *pluginsdk.ResourceData, times []date.Time) *backup.MonthlyRetentionSchedule {
	if rb, ok := d.Get("retention_monthly").([]interface{}); ok && len(rb) > 0 {
		block := rb[0].(map[string]interface{})

		retention := backup.MonthlyRetentionSchedule{
			RetentionScheduleFormatType: backup.RetentionScheduleFormatWeekly, // this is always weekly ¯\_(ツ)_/¯
			RetentionScheduleDaily:      nil,                                  // and this is always nil..
			RetentionScheduleWeekly:     expandBackupProtectionPolicyFileShareRetentionWeeklyFormat(block),
			RetentionTimes:              &times,
			RetentionDuration: &backup.RetentionDuration{
				Count:        utils.Int32(int32(block["count"].(int))),
				DurationType: backup.RetentionDurationTypeMonths,
			},
		}

		return &retention
	}

	return nil
}

func expandBackupProtectionPolicyFileShareRetentionYearly(d *pluginsdk.ResourceData, times []date.Time) *backup.YearlyRetentionSchedule {
	if rb, ok := d.Get("retention_yearly").([]interface{}); ok && len(rb) > 0 {
		block := rb[0].(map[string]interface{})

		retention := backup.YearlyRetentionSchedule{
			RetentionScheduleFormatType: backup.RetentionScheduleFormatWeekly, // this is always weekly ¯\_(ツ)_/¯
			RetentionScheduleDaily:      nil,                                  // and this is always nil..
			RetentionScheduleWeekly:     expandBackupProtectionPolicyFileShareRetentionWeeklyFormat(block),
			RetentionTimes:              &times,
			RetentionDuration: &backup.RetentionDuration{
				Count:        utils.Int32(int32(block["count"].(int))),
				DurationType: backup.RetentionDurationTypeYears,
			},
		}

		if v, ok := block["months"].(*pluginsdk.Set); ok {
			months := make([]backup.MonthOfYear, 0)
			for _, month := range v.List() {
				months = append(months, backup.MonthOfYear(month.(string)))
			}
			retention.MonthsOfYear = &months
		}

		return &retention
	}

	return nil
}

func expandBackupProtectionPolicyFileShareRetentionWeeklyFormat(block map[string]interface{}) *backup.WeeklyRetentionFormat {
	weekly := backup.WeeklyRetentionFormat{}

	if v, ok := block["weekdays"].(*pluginsdk.Set); ok {
		days := make([]backup.DayOfWeek, 0)
		for _, day := range v.List() {
			days = append(days, backup.DayOfWeek(day.(string)))
		}
		weekly.DaysOfTheWeek = &days
	}

	if v, ok := block["weeks"].(*pluginsdk.Set); ok {
		weeks := make([]backup.WeekOfMonth, 0)
		for _, week := range v.List() {
			weeks = append(weeks, backup.WeekOfMonth(week.(string)))
		}
		weekly.WeeksOfTheMonth = &weeks
	}

	return &weekly
}

func flattenBackupProtectionPolicyFileShareSchedule(schedule *backup.SimpleSchedulePolicy) []interface{} {
	block := map[string]interface{}{}

	block["frequency"] = string(schedule.ScheduleRunFrequency)

	if times := schedule.ScheduleRunTimes; times != nil && len(*times) > 0 {
		block["time"] = (*times)[0].Format("15:04")
	}

	return []interface{}{block}
}

func flattenBackupProtectionPolicyFileShareRetentionDaily(daily *backup.DailyRetentionSchedule) []interface{} {
	block := map[string]interface{}{}

	if duration := daily.RetentionDuration; duration != nil {
		if v := duration.Count; v != nil {
			block["count"] = *v
		}
	}

	return []interface{}{block}
}

func flattenBackupProtectionPolicyFileShareRetentionWeekly(weekly *backup.WeeklyRetentionSchedule) []interface{} {
	block := map[string]interface{}{}

	if duration := weekly.RetentionDuration; duration != nil {
		if v := duration.Count; v != nil {
			block["count"] = *v
		}
	}

	if days := weekly.DaysOfTheWeek; days != nil {
		weekdays := make([]interface{}, 0)
		for _, d := range *days {
			weekdays = append(weekdays, string(d))
		}
		block["weekdays"] = pluginsdk.NewSet(pluginsdk.HashString, weekdays)
	}

	return []interface{}{block}
}

func flattenBackupProtectionPolicyFileShareRetentionMonthly(monthly *backup.MonthlyRetentionSchedule) []interface{} {
	block := map[string]interface{}{}

	if duration := monthly.RetentionDuration; duration != nil {
		if v := duration.Count; v != nil {
			block["count"] = *v
		}
	}

	if weekly := monthly.RetentionScheduleWeekly; weekly != nil {
		block["weekdays"], block["weeks"] = flattenBackupProtectionPolicyFileShareRetentionWeeklyFormat(weekly)
	}

	return []interface{}{block}
}

func flattenBackupProtectionPolicyFileShareRetentionYearly(yearly *backup.YearlyRetentionSchedule) []interface{} {
	block := map[string]interface{}{}

	if duration := yearly.RetentionDuration; duration != nil {
		if v := duration.Count; v != nil {
			block["count"] = *v
		}
	}

	if weekly := yearly.RetentionScheduleWeekly; weekly != nil {
		block["weekdays"], block["weeks"] = flattenBackupProtectionPolicyFileShareRetentionWeeklyFormat(weekly)
	}

	if months := yearly.MonthsOfYear; months != nil {
		slice := make([]interface{}, 0)
		for _, d := range *months {
			slice = append(slice, string(d))
		}
		block["months"] = pluginsdk.NewSet(pluginsdk.HashString, slice)
	}

	return []interface{}{block}
}

func flattenBackupProtectionPolicyFileShareRetentionWeeklyFormat(retention *backup.WeeklyRetentionFormat) (weekdays, weeks *pluginsdk.Set) {
	if days := retention.DaysOfTheWeek; days != nil {
		slice := make([]interface{}, 0)
		for _, d := range *days {
			slice = append(slice, string(d))
		}
		weekdays = pluginsdk.NewSet(pluginsdk.HashString, slice)
	}

	if days := retention.WeeksOfTheMonth; days != nil {
		slice := make([]interface{}, 0)
		for _, d := range *days {
			slice = append(slice, string(d))
		}
		weeks = pluginsdk.NewSet(pluginsdk.HashString, slice)
	}

	return weekdays, weeks
}

func resourceBackupProtectionPolicyFileShareWaitForUpdate(ctx context.Context, client *backup.ProtectionPoliciesClient, vaultName, resourceGroup, policyName string, d *pluginsdk.ResourceData) (backup.ProtectionPolicyResource, error) {
	state := &pluginsdk.StateChangeConf{
		MinTimeout: 30 * time.Second,
		Delay:      10 * time.Second,
		Pending:    []string{"NotFound"},
		Target:     []string{"Found"},
		Refresh:    resourceBackupProtectionPolicyFileShareRefreshFunc(ctx, client, vaultName, resourceGroup, policyName),
	}

	if d.IsNewResource() {
		state.Timeout = d.Timeout(pluginsdk.TimeoutCreate)
	} else {
		state.Timeout = d.Timeout(pluginsdk.TimeoutUpdate)
	}

	resp, err := state.WaitForStateContext(ctx)
	if err != nil {
		return resp.(backup.ProtectionPolicyResource), fmt.Errorf("waiting for the Recovery Service Protection Policy %q to update (Resource Group %q): %+v", policyName, resourceGroup, err)
	}

	return resp.(backup.ProtectionPolicyResource), nil
}

func resourceBackupProtectionPolicyFileShareWaitForDeletion(ctx context.Context, client *backup.ProtectionPoliciesClient, vaultName, resourceGroup, policyName string, d *pluginsdk.ResourceData) (backup.ProtectionPolicyResource, error) {
	state := &pluginsdk.StateChangeConf{
		MinTimeout: 30 * time.Second,
		Delay:      10 * time.Second,
		Pending:    []string{"Found"},
		Target:     []string{"NotFound"},
		Refresh:    resourceBackupProtectionPolicyFileShareRefreshFunc(ctx, client, vaultName, resourceGroup, policyName),
		Timeout:    d.Timeout(pluginsdk.TimeoutDelete),
	}

	resp, err := state.WaitForStateContext(ctx)
	if err != nil {
		return resp.(backup.ProtectionPolicyResource), fmt.Errorf("waiting for the Recovery Service Protection Policy %q to be missing (Resource Group %q): %+v", policyName, resourceGroup, err)
	}

	return resp.(backup.ProtectionPolicyResource), nil
}

func resourceBackupProtectionPolicyFileShareRefreshFunc(ctx context.Context, client *backup.ProtectionPoliciesClient, vaultName, resourceGroup, policyName string) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := client.Get(ctx, vaultName, resourceGroup, policyName)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return resp, "NotFound", nil
			}

			return resp, "Error", fmt.Errorf("making Read request on Recovery Service Protection Policy %q (Resource Group %q): %+v", policyName, resourceGroup, err)
		}

		return resp, "Found", nil
	}
}

func resourceBackupProtectionPolicyFileShareSchema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
			ValidateFunc: validation.StringMatch(
				regexp.MustCompile("^[a-zA-Z][-_!a-zA-Z0-9]{2,149}$"),
				"Backup Policy name must be 3 - 150 characters long, start with a letter, contain only letters and numbers.",
			),
		},

		"resource_group_name": azure.SchemaResourceGroupName(),

		"recovery_vault_name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.RecoveryServicesVaultName,
		},

		"timezone": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			Default:  "UTC",
		},

		"backup": {
			Type:     pluginsdk.TypeList,
			MaxItems: 1,
			Required: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"frequency": {
						Type:             pluginsdk.TypeString,
						Required:         true,
						DiffSuppressFunc: suppress.CaseDifference,
						ValidateFunc: validation.StringInSlice([]string{
							string(backup.ScheduleRunTypeDaily),
						}, false),
					},

					"time": {
						Type:     pluginsdk.TypeString,
						Required: true,
						ValidateFunc: validation.StringMatch(
							regexp.MustCompile("^([01][0-9]|[2][0-3]):([03][0])$"), // time must be on the hour or half past
							"Time of day must match the format HH:mm where HH is 00-23 and mm is 00 or 30",
						),
					},
				},
			},
		},

		"retention_daily": {
			Type:     pluginsdk.TypeList,
			MaxItems: 1,
			Required: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"count": {
						Type:         pluginsdk.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntBetween(1, 200),
					},
				},
			},
		},

		"retention_weekly": {
			Type:     pluginsdk.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"count": {
						Type:         pluginsdk.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntBetween(1, 200),
					},

					"weekdays": {
						Type:     pluginsdk.TypeSet,
						Required: true,
						Set:      set.HashStringIgnoreCase,
						Elem: &pluginsdk.Schema{
							Type:             pluginsdk.TypeString,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc:     validation.IsDayOfTheWeek(true),
						},
					},
				},
			},
		},

		"retention_monthly": {
			Type:     pluginsdk.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"count": {
						Type:         pluginsdk.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntBetween(1, 120),
					},

					"weeks": {
						Type:     pluginsdk.TypeSet,
						Required: true,
						Set:      set.HashStringIgnoreCase,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
							ValidateFunc: validation.StringInSlice([]string{
								string(backup.WeekOfMonthFirst),
								string(backup.WeekOfMonthSecond),
								string(backup.WeekOfMonthThird),
								string(backup.WeekOfMonthFourth),
								string(backup.WeekOfMonthLast),
							}, false),
						},
					},

					"weekdays": {
						Type:     pluginsdk.TypeSet,
						Required: true,
						Set:      set.HashStringIgnoreCase,
						Elem: &pluginsdk.Schema{
							Type:             pluginsdk.TypeString,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc:     validation.IsDayOfTheWeek(true),
						},
					},
				},
			},
		},

		"retention_yearly": {
			Type:     pluginsdk.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"count": {
						Type:         pluginsdk.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntBetween(1, 10),
					},

					"months": {
						Type:     pluginsdk.TypeSet,
						Required: true,
						Set:      set.HashStringIgnoreCase,
						Elem: &pluginsdk.Schema{
							Type:             pluginsdk.TypeString,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc:     validation.IsMonth(true),
						},
					},

					"weeks": {
						Type:     pluginsdk.TypeSet,
						Required: true,
						Set:      set.HashStringIgnoreCase,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
							ValidateFunc: validation.StringInSlice([]string{
								string(backup.WeekOfMonthFirst),
								string(backup.WeekOfMonthSecond),
								string(backup.WeekOfMonthThird),
								string(backup.WeekOfMonthFourth),
								string(backup.WeekOfMonthLast),
							}, false),
						},
					},

					"weekdays": {
						Type:     pluginsdk.TypeSet,
						Required: true,
						Set:      set.HashStringIgnoreCase,
						Elem: &pluginsdk.Schema{
							Type:             pluginsdk.TypeString,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc:     validation.IsDayOfTheWeek(true),
						},
					},
				},
			},
		},
	}
}
