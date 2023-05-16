package maintenance

import (
	"github.com/hashicorp/terraform-provider-azurerm/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
)

type Registration struct{}

var _ sdk.UntypedServiceRegistrationWithAGitHubLabel = Registration{}

func (r Registration) AssociatedGitHubLabel() string {
	return "service/maintenance"
}

func (r Registration) Name() string {
	return "Maintenance"
}

func (r Registration) WebsiteCategories() []string {
	return []string{
		"Maintenance",
	}
}

func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_maintenance_configuration": dataSourceMaintenanceConfiguration(),
	}
}

func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_maintenance_assignment_dedicated_host":            resourceArmMaintenanceAssignmentDedicatedHost(),
		"azurerm_maintenance_assignment_virtual_machine":           resourceArmMaintenanceAssignmentVirtualMachine(),
		"azurerm_maintenance_assignment_virtual_machine_scale_set": resourceArmMaintenanceAssignmentVirtualMachineScaleSet(),
		"azurerm_maintenance_configuration":                        resourceArmMaintenanceConfiguration(),
	}
}
