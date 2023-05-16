package costmanagement

import (
	"github.com/hashicorp/terraform-provider-azurerm/sdk"
)

type Registration struct{}

var _ sdk.TypedServiceRegistrationWithAGitHubLabel = Registration{}

func (r Registration) AssociatedGitHubLabel() string {
	return "service/cost-management"
}

func (r Registration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{}
}

func (r Registration) Resources() []sdk.Resource {
	return []sdk.Resource{
		ResourceGroupCostManagementExportResource{},
		SubscriptionCostManagementExportResource{},
	}
}

// Name is the name of this Service
func (r Registration) Name() string {
	return "Cost Management"
}

// WebsiteCategories returns a list of categories which can be used for the sidebar
func (r Registration) WebsiteCategories() []string {
	return []string{
		"Cost Management",
	}
}
