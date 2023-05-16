package mixedreality

import (
	"github.com/hashicorp/terraform-provider-azurerm/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
)

type Registration struct{}

var _ sdk.UntypedServiceRegistrationWithAGitHubLabel = Registration{}

func (r Registration) AssociatedGitHubLabel() string {
	return "service/mixed-reality"
}

// Name is the name of this Service
func (r Registration) Name() string {
	return "Mixed Reality"
}

// WebsiteCategories returns a list of categories which can be used for the sidebar
func (r Registration) WebsiteCategories() []string {
	return []string{
		"Mixed Reality",
	}
}

// SupportedDataSources returns the supported Data Sources supported by this Service
func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_spatial_anchors_account": dataSourceSpatialAnchorsAccount(),
	}
}

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_spatial_anchors_account": resourceSpatialAnchorsAccount(),
	}
}
