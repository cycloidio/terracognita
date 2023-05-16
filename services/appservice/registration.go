package appservice

import (
	"github.com/hashicorp/terraform-provider-azurerm/sdk"
)

var _ sdk.TypedServiceRegistrationWithAGitHubLabel = Registration{}

type Registration struct{}

func (r Registration) AssociatedGitHubLabel() string {
	return "service/app-service"
}

func (r Registration) WebsiteCategories() []string {
	return nil
}

func (r Registration) Name() string {
	return "AppService"
}

func (r Registration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{
		AppServiceSourceControlTokenDataSource{},
		LinuxFunctionAppDataSource{},
		LinuxWebAppDataSource{},
		ServicePlanDataSource{},
		WindowsFunctionAppDataSource{},
		WindowsWebAppDataSource{},
	}
}

func (r Registration) Resources() []sdk.Resource {
	return []sdk.Resource{
		AppServiceSourceControlTokenResource{},
		FunctionAppActiveSlotResource{},
		FunctionAppFunctionResource{},
		FunctionAppHybridConnectionResource{},
		LinuxFunctionAppResource{},
		LinuxFunctionAppSlotResource{},
		LinuxWebAppResource{},
		LinuxWebAppSlotResource{},
		ServicePlanResource{},
		SourceControlResource{},
		SourceControlSlotResource{},
		WebAppActiveSlotResource{},
		WebAppHybridConnectionResource{},
		WindowsFunctionAppResource{},
		WindowsFunctionAppSlotResource{},
		WindowsWebAppResource{},
		WindowsWebAppSlotResource{},
	}
}
