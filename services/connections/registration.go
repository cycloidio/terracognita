package connections

import (
	"github.com/hashicorp/terraform-provider-azurerm/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
)

var _ sdk.UntypedServiceRegistration = Registration{}

type Registration struct{}

func (r Registration) Name() string {
	return "Connections"
}

func (r Registration) WebsiteCategories() []string {
	return []string{
		"Connections",
	}
}

func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_managed_api": dataSourceManagedApi(),
	}
}

func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_api_connection": resourceConnection(),
	}
}
