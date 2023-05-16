package securitycenter

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/security/mgmt/v3.0/security"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/securitycenter/azuresdkhacks"
	"github.com/hashicorp/terraform-provider-azurerm/services/securitycenter/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

// TODO: this resource should be split into data_export_setting and alert_sync_setting

func resourceSecurityCenterSetting() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceSecurityCenterSettingUpdate,
		Read:   resourceSecurityCenterSettingRead,
		Update: resourceSecurityCenterSettingUpdate,
		Delete: resourceSecurityCenterSettingDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.SettingID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(10 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(10 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"setting_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"MCAS",
					"WDATP",
				}, false),
			},
			"enabled": {
				Type:     pluginsdk.TypeBool,
				Required: true,
			},
		},
	}
}

func resourceSecurityCenterSettingUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).SecurityCenter.SettingClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewSettingID(subscriptionId, d.Get("setting_name").(string))

	if d.IsNewResource() {
		// TODO: switch back when Swagger/API bug has been fixed:
		// https://github.com/Azure/azure-sdk-for-go/issues/12724 (`Enabled` field missing)
		existing, err := azuresdkhacks.GetSecurityCenterSetting(ctx, client, id.Name)
		if err != nil {
			return fmt.Errorf("checking for presence of existing %s: %v", id, err)
		}

		if existing.DataExportSettingProperties != nil && existing.DataExportSettingProperties.Enabled != nil && *existing.DataExportSettingProperties.Enabled {
			return tf.ImportAsExistsError("azurerm_security_center_setting", id.ID())
		}
	}

	enabled := d.Get("enabled").(bool)
	setting := security.DataExportSettings{
		DataExportSettingProperties: &security.DataExportSettingProperties{
			Enabled: &enabled,
		},
		Kind: security.KindDataExportSettings,
	}

	if _, err := client.Update(ctx, id.Name, setting); err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceSecurityCenterSettingRead(d, meta)
}

func resourceSecurityCenterSettingRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).SecurityCenter.SettingClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SettingID(d.Id())
	if err != nil {
		return err
	}

	// TODO: switch to back when Swagger/API bug has been fixed:
	// https://github.com/Azure/azure-sdk-for-go/issues/12724 (`Enabled` field missing)
	resp, err := azuresdkhacks.GetSecurityCenterSetting(ctx, client, id.Name)
	if err != nil {
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	if properties := resp.DataExportSettingProperties; properties != nil {
		d.Set("enabled", properties.Enabled)
	}
	d.Set("setting_name", id.Name)

	return nil
}

func resourceSecurityCenterSettingDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).SecurityCenter.SettingClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SettingID(d.Id())
	if err != nil {
		return err
	}

	setting := security.DataExportSettings{
		DataExportSettingProperties: &security.DataExportSettingProperties{
			Enabled: utils.Bool(false),
		},
		Kind: security.KindDataExportSettings,
	}

	if _, err := client.Update(ctx, id.Name, setting); err != nil {
		return fmt.Errorf("disabling %s: %+v", id, err)
	}

	return nil
}
