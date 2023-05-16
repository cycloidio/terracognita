package loadtest

import (
	"github.com/hashicorp/terraform-provider-azurerm/sdk"
)

var _ sdk.TypedServiceRegistrationWithAGitHubLabel = Registration{}

type Registration struct{}

func (r Registration) AssociatedGitHubLabel() string {
	return "service/load-test"
}

func (r Registration) WebsiteCategories() []string {
	return []string{
		"Load Test",
	}
}

func (r Registration) Name() string {
	return "Load Test"
}

func (r Registration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{}
}

func (r Registration) Resources() []sdk.Resource {
	return []sdk.Resource{
		LoadTestResource{},
	}
}
