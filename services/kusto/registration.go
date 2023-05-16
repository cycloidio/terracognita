package kusto

import (
	"github.com/hashicorp/terraform-provider-azurerm/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
)

type Registration struct{}

var _ sdk.UntypedServiceRegistrationWithAGitHubLabel = Registration{}

func (r Registration) AssociatedGitHubLabel() string {
	return "service/kusto"
}

// Name is the name of this Service
func (r Registration) Name() string {
	return "Kusto"
}

// WebsiteCategories returns a list of categories which can be used for the sidebar
func (r Registration) WebsiteCategories() []string {
	return []string{
		"Data Explorer",
	}
}

// SupportedDataSources returns the supported Data Sources supported by this Service
func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_kusto_cluster":  dataSourceKustoCluster(),
		"azurerm_kusto_database": dataSourceKustoDatabase(),
	}
}

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_kusto_cluster":                         resourceKustoCluster(),
		"azurerm_kusto_cluster_customer_managed_key":    resourceKustoClusterCustomerManagedKey(),
		"azurerm_kusto_cluster_principal_assignment":    resourceKustoClusterPrincipalAssignment(),
		"azurerm_kusto_database":                        resourceKustoDatabase(),
		"azurerm_kusto_database_principal_assignment":   resourceKustoDatabasePrincipalAssignment(),
		"azurerm_kusto_eventgrid_data_connection":       resourceKustoEventGridDataConnection(),
		"azurerm_kusto_eventhub_data_connection":        resourceKustoEventHubDataConnection(),
		"azurerm_kusto_iothub_data_connection":          resourceKustoIotHubDataConnection(),
		"azurerm_kusto_attached_database_configuration": resourceKustoAttachedDatabaseConfiguration(),
		"azurerm_kusto_script":                          resourceKustoDatabaseScript(),
	}
}
