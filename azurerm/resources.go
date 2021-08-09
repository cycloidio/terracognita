package azurerm

import (
	"context"

	"github.com/pkg/errors"

	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/provider"
)

// ResourceType is the type used to define all the Resources
// from the Provider
type ResourceType int

//go:generate enumer -type ResourceType -addprefix azurerm_ -transform snake -linecomment
const (
	ResourceGroup ResourceType = iota
	//Compute Resources
	VirtualMachine
	VirtualMachineExtension
	VirtualMachineScaleSet
	VirtualNetwork
	AvailabilitySet
	Image
	//Network Resources
	Subnet
	NetworkInterface
	NetworkSecurityGroup
	ApplicationGateway
	ApplicationSecurityGroup
	DdosProtectionPlan
	AzureFirewall
	LocalNetworkGateway
	NatGateway
	Profile
	SecurityRule
	PublicIPAddress
	PublicIPPrefix
	Route
	RouteTable
	VirtualNetworkGateway
	VirtualNetworkGatewayConnection
	VirtualNetworkPeering
	WebApplicationFirewallPolicy
	//Desktop Resources
	VirtualDesktopHostPool
	VirtualDesktopApplicationGroup
	//Logic Resources
	LogicAppWorkflow
	LogicAppTriggerCustom
	LogicAppActionCustom
	//Container Registry Resources
	ContainerRegistry
	ContainerRegistryWebhook
	//Storage Resources
	StorageAccount
	StorageQueue
	StorageFileShare
	StorageTable
)

type rtFn func(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error)

var (
	resources = map[ResourceType]rtFn{
		//Compute Resources
		VirtualMachine:          virtualMachines,
		VirtualMachineExtension: virtualMachineExtensions,
		VirtualNetwork:          cacheVirtualNetworks,
		VirtualMachineScaleSet:  virtualMachineScaleSets,
		AvailabilitySet:         availabilitySets,
		Image:                   images,
		//Network Resources
		Subnet:                          subnets,
		NetworkInterface:                networkInterfaces,
		NetworkSecurityGroup:            networkSecurityGroups,
		ApplicationGateway:              applicationGateways,
		ApplicationSecurityGroup:        applicationSecurityGroups,
		DdosProtectionPlan:              ddosProtectionPlans,
		AzureFirewall:                   azureFirewalls,
		LocalNetworkGateway:             localNetworkGateways,
		NatGateway:                      natGateways,
		Profile:                         profiles,
		SecurityRule:                    securityRules,
		PublicIPAddress:                 publicIPAddresses,
		PublicIPPrefix:                  publicIPPrefixes,
		Route:                           routes,
		RouteTable:                      routeTables,
		VirtualNetworkGateway:           virtualNetworkGateways,
		VirtualNetworkGatewayConnection: virtualNetworkGatewayConnections,
		VirtualNetworkPeering:           virtualNetworkPeerings,
		WebApplicationFirewallPolicy:    webApplicationFirewallPolicies,
		//Desktop Resources
		VirtualDesktopApplicationGroup: virtualApplicationGroups,
		VirtualDesktopHostPool:         virtualDesktopHostPools,
		//Logic Resources
		LogicAppActionCustom:  logicAppActionCustoms,
		LogicAppWorkflow:      logicAppWorkflows,
		LogicAppTriggerCustom: logicAppTriggerCustoms,
		//Container Registry Resources
		ContainerRegistry:        containerRegistries,
		ContainerRegistryWebhook: containerRegistryWebhooks,
	}
)

func resourceGroup(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	resourceGroup := a.azurer.GetResourceGroup()
	r := provider.NewResource(*resourceGroup.ID, resourceType, a)
	resources := []provider.Resource{r}
	return resources, nil
}

//Compute Resources
func virtualMachines(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualMachines, err := a.azurer.ListVirtualMachines(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list virtual machines from reader")
	}
	resources := make([]provider.Resource, 0, len(virtualMachines))
	for _, virtualMachine := range virtualMachines {
		r := provider.NewResource(*virtualMachine.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *virtualMachine.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the virtual machine '%s'", *virtualMachine.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func virtualMachineScaleSets(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualMachineScaleSets, err := a.azurer.ListVirtualMachineScaleSets(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list virtual machines scale sets from reader")
	}
	resources := make([]provider.Resource, 0, len(virtualMachineScaleSets))
	for _, virtualMachineScaleSet := range virtualMachineScaleSets {
		r := provider.NewResource(*virtualMachineScaleSet.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func virtualMachineExtensions(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualMachineNames, err := getVirtualMachineNames(ctx, a, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list virtual machines from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, virtualMachineName := range virtualMachineNames {
		extensions, err := a.azurer.ListVirtualMachineExtensions(ctx, virtualMachineName, "")
		if err != nil {
			return nil, errors.Wrap(err, "unable to list virtual machine extensions from reader")
		}
		for _, extension := range extensions {
			r := provider.NewResource(*extension.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func availabilitySets(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	availabilitySets, err := a.azurer.ListAvailabilitySets(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list availability sets from reader")
	}
	resources := make([]provider.Resource, 0, len(availabilitySets))
	for _, availabilitySet := range availabilitySets {
		r := provider.NewResource(*availabilitySet.ID, resourceType, a)
		if err := r.Data().Set("name", *availabilitySet.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the availability set '%s'", *availabilitySet.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func images(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	images, err := a.azurer.ListImages(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list availability sets from reader")
	}
	resources := make([]provider.Resource, 0, len(images))
	for _, image := range images {
		r := provider.NewResource(*image.ID, resourceType, a)
		if err := r.Data().Set("name", *image.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the IMAGE '%s'", *image.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

//Network Resources
func virtualNetworks(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualNetworks, err := a.azurer.ListVirtualNetworks(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list virtual networks from reader")
	}
	resources := make([]provider.Resource, 0, len(virtualNetworks))
	for _, virtualNetwork := range virtualNetworks {
		r := provider.NewResource(*virtualNetwork.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *virtualNetwork.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the virtual network '%s'", *virtualNetwork.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func subnets(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualNetworkNames, err := getVirtualNetworkNames(ctx, a, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list virtual networks from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, virtualNetworkName := range virtualNetworkNames {
		subnets, err := a.azurer.ListSubnets(ctx, virtualNetworkName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list subnets from reader")
		}
		for _, subnet := range subnets {
			r := provider.NewResource(*subnet.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func networkInterfaces(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	networkInterfaces, err := a.azurer.ListInterfaces(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list network interfaces from reader")
	}
	resources := make([]provider.Resource, 0, len(networkInterfaces))
	for _, networkInterface := range networkInterfaces {
		r := provider.NewResource(*networkInterface.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func networkSecurityGroups(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	securityGroups, err := a.azurer.ListSecurityGroups(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list network security groups from reader")
	}
	resources := make([]provider.Resource, 0, len(securityGroups))
	for _, securityGroup := range securityGroups {
		r := provider.NewResource(*securityGroup.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func applicationGateways(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	applicationGateways, err := a.azurer.ListApplicationGateways(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list network application gateways from reader")
	}
	resources := make([]provider.Resource, 0, len(applicationGateways))
	for _, applicationGateway := range applicationGateways {
		r := provider.NewResource(*applicationGateway.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func applicationSecurityGroups(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	applicationSecurityGroups, err := a.azurer.ListApplicationSecurityGroups(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list network application security groups from reader")
	}
	resources := make([]provider.Resource, 0, len(applicationSecurityGroups))
	for _, applicationSecurityGroup := range applicationSecurityGroups {
		r := provider.NewResource(*applicationSecurityGroup.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func ddosProtectionPlans(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	ddosProtectionPlans, err := a.azurer.ListDdosProtectionPlans(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list network ddos protection plans from reader")
	}
	resources := make([]provider.Resource, 0, len(ddosProtectionPlans))
	for _, ddosProtectionPlan := range ddosProtectionPlans {
		r := provider.NewResource(*ddosProtectionPlan.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func azureFirewalls(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	azureFirewalls, err := a.azurer.ListAzureFirewalls(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list azure network firewall from reader")
	}
	resources := make([]provider.Resource, 0, len(azureFirewalls))
	for _, azureFirewall := range azureFirewalls {
		r := provider.NewResource(*azureFirewall.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func localNetworkGateways(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	localNetworkGateways, err := a.azurer.ListLocalNetworkGateways(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list local network gateways from reader")
	}
	resources := make([]provider.Resource, 0, len(localNetworkGateways))
	for _, localNetworkGateway := range localNetworkGateways {
		r := provider.NewResource(*localNetworkGateway.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func natGateways(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	natGateways, err := a.azurer.ListNatGateways(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list local network gateways from reader")
	}
	resources := make([]provider.Resource, 0, len(natGateways))
	for _, natGateway := range natGateways {
		r := provider.NewResource(*natGateway.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func profiles(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	profiles, err := a.azurer.ListProfiles(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list profiles from reader")
	}
	resources := make([]provider.Resource, 0, len(profiles))
	for _, profile := range profiles {
		r := provider.NewResource(*profile.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func securityRules(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	securityGroupNames, err := getSecurityGroups(ctx, a, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list security Groups from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, securityGroupName := range securityGroupNames {
		securityRule, err := a.azurer.ListSecurityRules(ctx, securityGroupName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list security rules from reader")
		}
		for _, securityRule := range securityRule {
			r := provider.NewResource(*securityRule.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func publicIPAddresses(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	publicIpAddresses, err := a.azurer.ListPublicIPAddresses(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list public IP addresses from reader")
	}
	resources := make([]provider.Resource, 0, len(publicIpAddresses))
	for _, publicIpAddress := range publicIpAddresses {
		r := provider.NewResource(*publicIpAddress.ID, resourceType, a)
		if err := r.Data().Set("name", *publicIpAddress.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the public Ip adress '%s'", *publicIpAddress.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func publicIPPrefixes(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	publicIpPrefixes, err := a.azurer.ListPublicIPPrefixes(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list public IP addresses from reader")
	}
	resources := make([]provider.Resource, 0, len(publicIpPrefixes))
	for _, publicIpPrefix := range publicIpPrefixes {
		r := provider.NewResource(*publicIpPrefix.ID, resourceType, a)
		if err := r.Data().Set("name", *publicIpPrefix.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the public IP prefix '%s'", *publicIpPrefix.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func routeTables(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	routeTables, err := a.azurer.ListRouteTables(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list route tables from reader")
	}
	resources := make([]provider.Resource, 0, len(routeTables))
	for _, routeTable := range routeTables {
		r := provider.NewResource(*routeTable.ID, resourceType, a)
		if err := r.Data().Set("name", *routeTable.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the routeTable '%s'", *routeTable.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func routes(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	routeTablesNames, err := getRouteTables(ctx, a, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list route Tables from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, routeTableName := range routeTablesNames {
		routes, err := a.azurer.ListRoutes(ctx, routeTableName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list routes from reader")
		}
		for _, route := range routes {
			r := provider.NewResource(*route.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func virtualNetworkGateways(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualNetworkGateways, err := a.azurer.ListVirtualNetworkGateways(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Virtual Network Gateways from reader")
	}
	resources := make([]provider.Resource, 0, len(virtualNetworkGateways))
	for _, virtualNetworkGateway := range virtualNetworkGateways {
		r := provider.NewResource(*virtualNetworkGateway.ID, resourceType, a)
		if err := r.Data().Set("name", *virtualNetworkGateway.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the virtual Network Gateway '%s'", *virtualNetworkGateway.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func virtualNetworkGatewayConnections(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualNetworkGatewayConnections, err := a.azurer.ListVirtualNetworkGatewayConnections(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list virtual network gateway connections from reader")
	}
	resources := make([]provider.Resource, 0, len(virtualNetworkGatewayConnections))
	for _, virtualNetworkGatewayConnection := range virtualNetworkGatewayConnections {
		r := provider.NewResource(*virtualNetworkGatewayConnection.ID, resourceType, a)
		if err := r.Data().Set("name", *virtualNetworkGatewayConnection.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the virtual Network Gateway connection '%s'", *virtualNetworkGatewayConnection.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func virtualNetworkPeerings(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualNetworkNames, err := getVirtualNetworkNames(ctx, a, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list virtual network names from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, virtualNetworkName := range virtualNetworkNames {
		virtualNetworkPeerings, err := a.azurer.ListVirtualNetworkPeerings(ctx, virtualNetworkName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list virtual network peerings from reader")
		}
		for _, virtualNetworkPeering := range virtualNetworkPeerings {
			r := provider.NewResource(*virtualNetworkPeering.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func webApplicationFirewallPolicies(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	webApplicationFirewallPolicies, err := a.azurer.ListWebApplicationFirewallPolicies(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list web application firewall policies from reader")
	}
	resources := make([]provider.Resource, 0, len(webApplicationFirewallPolicies))
	for _, webApplicationFirewallPolicy := range webApplicationFirewallPolicies {
		r := provider.NewResource(*webApplicationFirewallPolicy.ID, resourceType, a)
		if err := r.Data().Set("name", *webApplicationFirewallPolicy.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the web application firewall policy '%s'", *webApplicationFirewallPolicy.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

//Desktop Resources
func virtualDesktopHostPools(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	pools, err := a.azurer.ListHostPools(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list host pools from reader")
	}
	resources := make([]provider.Resource, 0, len(pools))
	for _, hostPool := range pools {
		r := provider.NewResource(*hostPool.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func virtualApplicationGroups(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	// the second argument; "filter" is set to "" because "Valid properties for filtering are applicationGroupType."
	// https://godoc.org/github.com/Azure/azure-sdk-for-go/services/preview/desktopvirtualization/mgmt/2019-12-10-preview/desktopvirtualization#ApplicationGroupsClient.ListByResourceGroup
	applicationGroups, err := a.azurer.ListApplicationGroups(ctx, "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list application groups from reader")
	}
	resources := make([]provider.Resource, 0, len(applicationGroups))
	for _, applicationGroup := range applicationGroups {
		r := provider.NewResource(*applicationGroup.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

//Logic Resources
func logicAppWorkflows(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	appWorkflows, err := a.azurer.ListWorkflows(ctx, nil, "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list logic app workflows from reader")
	}
	resources := make([]provider.Resource, 0, len(appWorkflows))
	for _, appWorkflow := range appWorkflows {
		r := provider.NewResource(*appWorkflow.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *appWorkflow.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the app workflow '%s'", *appWorkflow.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func logicAppTriggerCustoms(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	appWorkflowNames, err := getWorkflowNames(ctx, a, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list logic app workflows from reader")
	}

	resources := make([]provider.Resource, 0)
	for _, appWorkflowName := range appWorkflowNames {
		triggers, err := a.azurer.ListWorkflowTriggers(ctx, appWorkflowName, nil, "")
		if err != nil {
			return nil, errors.Wrap(err, "unable to list logic app trigger HTTP requests from reader")
		}
		for _, trigger := range triggers {
			r := provider.NewResource(*trigger.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func logicAppActionCustoms(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	appWorkflowNames, err := getWorkflowNames(ctx, a, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list logic app workflows from reader")
	}

	resources := make([]provider.Resource, 0)
	for _, appWorkflowName := range appWorkflowNames {
		runs, err := a.azurer.ListWorkflowRuns(ctx, appWorkflowName, nil, "")
		if err != nil {
			return nil, errors.Wrap(err, "unable to list workflow runs from reader")
		}

		for _, run := range runs {
			actions, err := a.azurer.ListWorkflowRunActions(ctx, appWorkflowName, *run.Name, nil, "")
			if err != nil {
				return nil, errors.Wrap(err, "unable to list workflow run actions from reader")
			}
			for _, action := range actions {
				r := provider.NewResource(*action.ID, resourceType, a)
				resources = append(resources, r)
			}
		}
	}
	return resources, nil
}

//Container Registry Resources
func containerRegistries(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	containerRegistries, err := a.azurer.ListContainerRegistries(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list logic app workflows from reader")
	}
	resources := make([]provider.Resource, 0, len(containerRegistries))
	for _, containerRegistry := range containerRegistries {
		r := provider.NewResource(*containerRegistry.ID, resourceType, a)
		if err := r.Data().Set("name", *containerRegistry.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the container Registry'%s'", *containerRegistry.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func containerRegistryWebhooks(ctx context.Context, a *azurerm, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	containerRegistriesNames, err := getContainerRegistries(ctx, a, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list container Registries from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, containerRegistryName := range containerRegistriesNames {
		containerRegistryWebhooks, err := a.azurer.ListContainerRegistryWebhooks(ctx, containerRegistryName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list container registry webhooks from reader")
		}
		for _, containerRegistryWebhook := range containerRegistryWebhooks {
			r := provider.NewResource(*containerRegistryWebhook.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}
