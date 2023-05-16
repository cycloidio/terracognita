package iottimeseriesinsights

import (
	"fmt"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/timeseriesinsights/mgmt/2020-05-15/timeseriesinsights"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/iottimeseriesinsights/migration"
	"github.com/hashicorp/terraform-provider-azurerm/services/iottimeseriesinsights/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/iottimeseriesinsights/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceIoTTimeSeriesInsightsAccessPolicy() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceIoTTimeSeriesInsightsAccessPolicyCreateUpdate,
		Read:   resourceIoTTimeSeriesInsightsAccessPolicyRead,
		Update: resourceIoTTimeSeriesInsightsAccessPolicyCreateUpdate,
		Delete: resourceIoTTimeSeriesInsightsAccessPolicyDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.AccessPolicyID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		SchemaVersion: 1,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.StandardEnvironmentAccessPolicyV0ToV1{},
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^[-\w\._\(\)]+$`),
					"IoT Time Series Insights Access Policy name must contain only word characters, periods, underscores, hyphens, and parentheses.",
				),
			},

			"time_series_insights_environment_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.TimeSeriesInsightsEnvironmentID,
			},

			"principal_object_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"description": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"roles": {
				Type:     pluginsdk.TypeSet,
				Required: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						string(timeseriesinsights.Contributor),
						string(timeseriesinsights.Reader),
					}, false),
				},
			},
		},
	}
}

func resourceIoTTimeSeriesInsightsAccessPolicyCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).IoTTimeSeriesInsights.AccessPoliciesClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	environmentId, err := parse.EnvironmentID(d.Get("time_series_insights_environment_id").(string))
	if err != nil {
		return err
	}

	resourceId := parse.NewAccessPolicyID(subscriptionId, environmentId.ResourceGroup, environmentId.Name, name)
	if d.IsNewResource() {
		existing, err := client.Get(ctx, environmentId.ResourceGroup, environmentId.Name, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing IoT Time Series Insights Access Policy %q (Resource Group %q): %s", name, environmentId.ResourceGroup, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_iot_time_series_insights_access_policy", resourceId.ID())
		}
	}

	policy := timeseriesinsights.AccessPolicyCreateOrUpdateParameters{
		AccessPolicyResourceProperties: &timeseriesinsights.AccessPolicyResourceProperties{
			Description:       utils.String(d.Get("description").(string)),
			PrincipalObjectID: utils.String(d.Get("principal_object_id").(string)),
			Roles:             expandIoTTimeSeriesInsightsAccessPolicyRoles(d.Get("roles").(*pluginsdk.Set).List()),
		},
	}

	if _, err := client.CreateOrUpdate(ctx, environmentId.ResourceGroup, environmentId.Name, name, policy); err != nil {
		return fmt.Errorf("creating/updating IoT Time Series Insights Access Policy %q (Resource Group %q): %+v", name, environmentId.ResourceGroup, err)
	}

	d.SetId(resourceId.ID())
	return resourceIoTTimeSeriesInsightsAccessPolicyRead(d, meta)
}

func resourceIoTTimeSeriesInsightsAccessPolicyRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).IoTTimeSeriesInsights.AccessPoliciesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.AccessPolicyID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.EnvironmentName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving IoT Time Series Insights Access Policy %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	environmentId := parse.NewEnvironmentID(id.SubscriptionId, id.ResourceGroup, id.EnvironmentName).ID()

	d.Set("name", resp.Name)
	d.Set("time_series_insights_environment_id", environmentId)

	if props := resp.AccessPolicyResourceProperties; props != nil {
		d.Set("description", props.Description)
		d.Set("principal_object_id", props.PrincipalObjectID)
		d.Set("roles", flattenIoTTimeSeriesInsightsAccessPolicyRoles(resp.Roles))
	}

	return nil
}

func resourceIoTTimeSeriesInsightsAccessPolicyDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).IoTTimeSeriesInsights.AccessPoliciesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.AccessPolicyID(d.Id())
	if err != nil {
		return err
	}

	response, err := client.Delete(ctx, id.ResourceGroup, id.EnvironmentName, id.Name)
	if err != nil {
		if !utils.ResponseWasNotFound(response) {
			return fmt.Errorf("deleting IoT Time Series Insights Access Policy %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}
	}

	return nil
}

func expandIoTTimeSeriesInsightsAccessPolicyRoles(input []interface{}) *[]timeseriesinsights.AccessPolicyRole {
	roles := make([]timeseriesinsights.AccessPolicyRole, 0)

	for _, v := range input {
		if v == nil {
			continue
		}
		roles = append(roles, timeseriesinsights.AccessPolicyRole(v.(string)))
	}

	return &roles
}

func flattenIoTTimeSeriesInsightsAccessPolicyRoles(input *[]timeseriesinsights.AccessPolicyRole) []interface{} {
	result := make([]interface{}, 0)
	if input != nil {
		for _, item := range *input {
			result = append(result, string(item))
		}
	}
	return result
}
