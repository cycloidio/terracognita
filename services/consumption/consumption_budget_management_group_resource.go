package consumption

import (
	"github.com/Azure/azure-sdk-for-go/services/consumption/mgmt/2019-10-01/consumption"
	"github.com/hashicorp/terraform-provider-azurerm/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/services/consumption/validate"
	validateManagementGroup "github.com/hashicorp/terraform-provider-azurerm/services/managementgroup/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
)

type ManagementGroupConsumptionBudget struct {
	base consumptionBudgetBaseResource
}

var (
	_ sdk.Resource                   = ManagementGroupConsumptionBudget{}
	_ sdk.ResourceWithCustomImporter = ManagementGroupConsumptionBudget{}
)

func (r ManagementGroupConsumptionBudget) Arguments() map[string]*pluginsdk.Schema {
	schema := map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsNotWhiteSpace,
		},
		"management_group_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validateManagementGroup.ManagementGroupID,
		},

		// Consumption Budgets for Management Groups have a different notification schema,
		// here we override the notification schema in the base resource
		"notification": {
			Type:     pluginsdk.TypeSet,
			Required: true,
			MinItems: 1,
			MaxItems: 5,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"enabled": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  true,
					},
					"threshold": {
						Type:         pluginsdk.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntBetween(0, 1000),
					},
					// Issue: https://github.com/Azure/azure-rest-api-specs/issues/16240
					// Toggling between these two values doesn't work at the moment and also doesn't throw an error
					// but it seems unlikely that a user would switch the threshold_type of their budgets frequently
					"threshold_type": {
						Type:     pluginsdk.TypeString,
						Optional: true,
						Default:  string(consumption.ThresholdTypeActual),
						ForceNew: true, // TODO: remove this when the above issue is fixed
						ValidateFunc: validation.StringInSlice([]string{
							string(consumption.ThresholdTypeActual),
							"Forecasted",
						}, false),
					},
					"operator": {
						Type:     pluginsdk.TypeString,
						Required: true,
						ValidateFunc: validation.StringInSlice([]string{
							string(consumption.OperatorTypeEqualTo),
							string(consumption.OperatorTypeGreaterThan),
							string(consumption.OperatorTypeGreaterThanOrEqualTo),
						}, false),
					},

					"contact_emails": {
						Type:     pluginsdk.TypeList,
						Required: true,
						MinItems: 1,
						Elem: &pluginsdk.Schema{
							Type:         pluginsdk.TypeString,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},
		},
	}
	return r.base.arguments(schema)
}

func (r ManagementGroupConsumptionBudget) Attributes() map[string]*pluginsdk.Schema {
	return r.base.attributes()
}

func (r ManagementGroupConsumptionBudget) ModelObject() interface{} {
	return nil
}

func (r ManagementGroupConsumptionBudget) ResourceType() string {
	return "azurerm_consumption_budget_management_group"
}

func (r ManagementGroupConsumptionBudget) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return validate.ConsumptionBudgetManagementGroupID
}

func (r ManagementGroupConsumptionBudget) Create() sdk.ResourceFunc {
	return r.base.createFunc(r.ResourceType(), "management_group_id")
}

func (r ManagementGroupConsumptionBudget) Read() sdk.ResourceFunc {
	return r.base.readFunc("management_group_id")
}

func (r ManagementGroupConsumptionBudget) Delete() sdk.ResourceFunc {
	return r.base.deleteFunc()
}

func (r ManagementGroupConsumptionBudget) Update() sdk.ResourceFunc {
	return r.base.updateFunc()
}

func (r ManagementGroupConsumptionBudget) CustomImporter() sdk.ResourceRunFunc {
	return r.base.importerFunc("management_group")
}
