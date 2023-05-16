package databasemigration

import (
	"github.com/hashicorp/terraform-provider-azurerm/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
)

type Registration struct{}

var _ sdk.UntypedServiceRegistrationWithAGitHubLabel = Registration{}

func (r Registration) AssociatedGitHubLabel() string {
	return "service/database-migration"
}

// Name is the name of this Service
func (r Registration) Name() string {
	return "Database Migration"
}

// WebsiteCategories returns a list of categories which can be used for the sidebar
func (r Registration) WebsiteCategories() []string {
	return []string{
		"Database Migration",
	}
}

// SupportedDataSources returns the supported Data Sources supported by this Service
func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_database_migration_service": dataSourceDatabaseMigrationService(),
		"azurerm_database_migration_project": dataSourceDatabaseMigrationProject(),
	}
}

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	resources := map[string]*pluginsdk.Resource{
		"azurerm_database_migration_service": resourceDatabaseMigrationService(),
		"azurerm_database_migration_project": resourceDatabaseMigrationProject(),
	}

	return resources
}
