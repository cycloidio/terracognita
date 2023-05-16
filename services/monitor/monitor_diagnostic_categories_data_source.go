package monitor

import (
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2021-07-01-preview/insights"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
)

func dataSourceMonitorDiagnosticCategories() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceMonitorDiagnosticCategoriesRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"resource_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"logs": {
				Type:     pluginsdk.TypeSet,
				Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
				Set:      pluginsdk.HashString,
				Computed: true,
			},

			"metrics": {
				Type:     pluginsdk.TypeSet,
				Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
				Set:      pluginsdk.HashString,
				Computed: true,
			},
		},
	}
}

func dataSourceMonitorDiagnosticCategoriesRead(d *pluginsdk.ResourceData, meta interface{}) error {
	categoriesClient := meta.(*clients.Client).Monitor.DiagnosticSettingsCategoryClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	actualResourceId := d.Get("resource_id").(string)
	// trim off the leading `/` since the CheckExistenceByID / List methods don't expect it
	resourceId := strings.TrimPrefix(actualResourceId, "/")

	// then retrieve the possible Diagnostics Categories for this Resource
	categories, err := categoriesClient.List(ctx, resourceId)
	if err != nil {
		return fmt.Errorf("retrieving Diagnostics Categories for Resource %q: %+v", actualResourceId, err)
	}

	if categories.Value == nil {
		return fmt.Errorf("retrieving Diagnostics Categories for Resource %q: `categories.Value` was nil", actualResourceId)
	}

	d.SetId(actualResourceId)
	val := *categories.Value

	metrics := make([]string, 0)
	logs := make([]string, 0)

	for _, v := range val {
		if v.Name == nil {
			continue
		}

		if category := v.DiagnosticSettingsCategory; category != nil {
			switch category.CategoryType {
			case insights.CategoryTypeLogs:
				logs = append(logs, *v.Name)
			case insights.CategoryTypeMetrics:
				metrics = append(metrics, *v.Name)
			default:
				return fmt.Errorf("Unsupported category type %q", string(category.CategoryType))
			}
		}
	}

	if err := d.Set("logs", logs); err != nil {
		return fmt.Errorf("setting `logs`: %+v", err)
	}

	if err := d.Set("metrics", metrics); err != nil {
		return fmt.Errorf("setting `metrics`: %+v", err)
	}

	return nil
}
