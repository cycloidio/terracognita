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
	// Compute Resources
	VirtualMachine
	VirtualMachineExtension
	VirtualMachineScaleSet
	VirtualNetwork
	AvailabilitySet
	Image
	// Network Resources
	Subnet
	NetworkInterface
	NetworkSecurityGroup
	ApplicationGateway
	ApplicationSecurityGroup
	NetworkDdosProtectionPlan
	Firewall
	LocalNetworkGateway
	NatGateway
	NetworkProfile
	NetworkSecurityRule
	PublicIP
	PublicIPPrefix
	Route
	RouteTable
	VirtualNetworkGateway
	VirtualNetworkGatewayConnection
	VirtualNetworkPeering
	WebApplicationFirewallPolicy
	// Desktop Resources
	VirtualDesktopHostPool
	VirtualDesktopApplicationGroup
	// Logic Resources
	LogicAppWorkflow
	LogicAppTriggerCustom
	LogicAppActionCustom
	// Container Registry Resources
	ContainerRegistry
	ContainerRegistryWebhook
	// Storage Resources
	StorageAccount
	StorageQueue
	StorageShare
	StorageTable
	StorageBlob
	// Database Resources- mariadb
	MariadbConfiguration
	MariadbDatabase
	MariadbFirewallRule
	MariadbServer
	MariadbVirtualNetworkRule
	// Database Resources - mysql
	MysqlConfiguration
	MysqlDatabase
	MysqlFirewallRule
	MysqlServer
	MysqlVirtualNetworkRule
	// Database Resources - postgresql
	PostgresqlConfiguration
	PostgresqlDatabase
	PostgresqlFirewallRule
	PostgresqlServer
	PostgresqlVirtualNetworkRule
	// Database Resources- sql
	SQLElasticPool // sql_elasticpool
	SQLDatabase
	SQLFirewallRule
	SQLServer
)

type rtFn func(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error)

var (
	resources = map[ResourceType]rtFn{
		ResourceGroup: resourceGroup,
		// Compute Resources
		VirtualMachine:          virtualMachines,
		VirtualMachineExtension: virtualMachineExtensions,
		VirtualNetwork:          cacheVirtualNetworks,
		VirtualMachineScaleSet:  virtualMachineScaleSets,
		AvailabilitySet:         availabilitySets,
		Image:                   images,
		// Network Resources
		Subnet:                          subnets,
		NetworkInterface:                networkInterfaces,
		NetworkSecurityGroup:            networkSecurityGroups,
		ApplicationGateway:              applicationGateways,
		ApplicationSecurityGroup:        applicationSecurityGroups,
		NetworkDdosProtectionPlan:       networkddosProtectionPlans,
		Firewall:                        firewalls,
		LocalNetworkGateway:             localNetworkGateways,
		NatGateway:                      natGateways,
		NetworkProfile:                  networkProfiles,
		NetworkSecurityRule:             networkSecurityRules,
		PublicIP:                        publicIP,
		PublicIPPrefix:                  publicIPPrefixes,
		Route:                           routes,
		RouteTable:                      routeTables,
		VirtualNetworkGateway:           virtualNetworkGateways,
		VirtualNetworkGatewayConnection: virtualNetworkGatewayConnections,
		VirtualNetworkPeering:           virtualNetworkPeerings,
		WebApplicationFirewallPolicy:    webApplicationFirewallPolicies,
		// Desktop Resources
		VirtualDesktopApplicationGroup: virtualApplicationGroups,
		VirtualDesktopHostPool:         virtualDesktopHostPools,
		// Logic Resources
		LogicAppActionCustom:  logicAppActionCustoms,
		LogicAppWorkflow:      logicAppWorkflows,
		LogicAppTriggerCustom: logicAppTriggerCustoms,
		// Container Registry Resources
		ContainerRegistry:        containerRegistries,
		ContainerRegistryWebhook: containerRegistryWebhooks,
		// Storage Resources
		StorageAccount: storageAccounts,
		StorageQueue:   storageQueues,
		StorageShare:   storageShares,
		StorageTable:   storageTables,
		StorageBlob:    storageBlobs,
		// Database Resources- mariadb
		MariadbConfiguration:      mariadbConfigurations,
		MariadbDatabase:           mariadbDatabases,
		MariadbFirewallRule:       mariadbFirewallRules,
		MariadbServer:             mariadbServers,
		MariadbVirtualNetworkRule: mariadbVirtualNetworkRules,
		// Database Resources - mysql
		MysqlConfiguration:      mysqlConfigurations,
		MysqlDatabase:           mysqlDatabases,
		MysqlFirewallRule:       mysqlFirewallRules,
		MysqlServer:             mysqlServers,
		MysqlVirtualNetworkRule: mysqlVirtualNetworkRules,
		// Database Resources - postgresql
		PostgresqlConfiguration:      postgresqlConfigurations,
		PostgresqlDatabase:           postgresqlDatabases,
		PostgresqlFirewallRule:       postgresqlFirewallRules,
		PostgresqlServer:             postgresqlServers,
		PostgresqlVirtualNetworkRule: postgresqlVirtualNetworkRules,
		// Database Resources- sql
		SQLElasticPool:  sqlElasticPools,
		SQLDatabase:     sqlDatabases,
		SQLFirewallRule: sqlFirewallRules,
		SQLServer:       sqlServers,
	}
)

func resourceGroup(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	resourceGroup := ar.GetResourceGroup()
	r := provider.NewResource(*resourceGroup.ID, resourceType, a)
	resources := []provider.Resource{r}
	return resources, nil
}

// Compute Resources

func virtualMachines(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualMachines, err := ar.ListVirtualMachines(ctx)
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

func virtualMachineScaleSets(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualMachineScaleSets, err := ar.ListVirtualMachineScaleSets(ctx)
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

func virtualMachineExtensions(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualMachineNames, err := getVirtualMachineNames(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list virtual machines from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, virtualMachineName := range virtualMachineNames {
		extensions, err := ar.ListVirtualMachineExtensions(ctx, virtualMachineName, "")
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

func availabilitySets(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	availabilitySets, err := ar.ListAvailabilitySets(ctx)
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

func images(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	images, err := ar.ListImages(ctx)
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

// Network Resources

func virtualNetworks(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualNetworks, err := ar.ListVirtualNetworks(ctx)
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

func subnets(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualNetworkNames, err := getVirtualNetworkNames(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list virtual networks from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, virtualNetworkName := range virtualNetworkNames {
		subnets, err := ar.ListSubnets(ctx, virtualNetworkName)
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

func networkInterfaces(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	networkInterfaces, err := ar.ListInterfaces(ctx)
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

func networkSecurityGroups(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	securityGroups, err := ar.ListSecurityGroups(ctx)
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

func applicationGateways(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	applicationGateways, err := ar.ListApplicationGateways(ctx)
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

func applicationSecurityGroups(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	applicationSecurityGroups, err := ar.ListApplicationSecurityGroups(ctx)
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

func networkddosProtectionPlans(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	ddosProtectionPlans, err := ar.ListDdosProtectionPlans(ctx)
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

func firewalls(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	azureFirewalls, err := ar.ListAzureFirewalls(ctx)
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

func localNetworkGateways(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	localNetworkGateways, err := ar.ListLocalNetworkGateways(ctx)
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

func natGateways(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	natGateways, err := ar.ListNatGateways(ctx)
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

func networkProfiles(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	profiles, err := ar.ListProfiles(ctx)
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

func networkSecurityRules(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	securityGroupNames, err := getSecurityGroups(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list security Groups from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, securityGroupName := range securityGroupNames {
		securityRule, err := ar.ListSecurityRules(ctx, securityGroupName)
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

func publicIP(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	publicIPAddresses, err := ar.ListPublicIPAddresses(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list public IP addresses from reader")
	}
	resources := make([]provider.Resource, 0, len(publicIPAddresses))
	for _, publicIPAddress := range publicIPAddresses {
		r := provider.NewResource(*publicIPAddress.ID, resourceType, a)
		if err := r.Data().Set("name", *publicIPAddress.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the public IP address '%s'", *publicIPAddress.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func publicIPPrefixes(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	publicIPPrefixes, err := ar.ListPublicIPPrefixes(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list public IP addresses from reader")
	}
	resources := make([]provider.Resource, 0, len(publicIPPrefixes))
	for _, publicIPPrefix := range publicIPPrefixes {
		r := provider.NewResource(*publicIPPrefix.ID, resourceType, a)
		if err := r.Data().Set("name", *publicIPPrefix.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the public IP prefix '%s'", *publicIPPrefix.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func routeTables(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	routeTables, err := ar.ListRouteTables(ctx)
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

func routes(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	routeTablesNames, err := getRouteTables(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list route Tables from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, routeTableName := range routeTablesNames {
		routes, err := ar.ListRoutes(ctx, routeTableName)
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

func virtualNetworkGateways(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualNetworkGateways, err := ar.ListVirtualNetworkGateways(ctx)
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

func virtualNetworkGatewayConnections(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualNetworkGatewayConnections, err := ar.ListVirtualNetworkGatewayConnections(ctx)
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

func virtualNetworkPeerings(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualNetworkNames, err := getVirtualNetworkNames(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list virtual network names from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, virtualNetworkName := range virtualNetworkNames {
		virtualNetworkPeerings, err := ar.ListVirtualNetworkPeerings(ctx, virtualNetworkName)
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

func webApplicationFirewallPolicies(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	webApplicationFirewallPolicies, err := ar.ListWebApplicationFirewallPolicies(ctx)
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

// Desktop Resources

func virtualDesktopHostPools(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	pools, err := ar.ListHostPools(ctx)
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

func virtualApplicationGroups(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	// the second argument; "filter" is set to "" because "Valid properties for filtering are applicationGroupType."
	// https://godoc.org/github.com/Azure/azure-sdk-for-go/services/preview/desktopvirtualization/mgmt/2019-12-10-preview/desktopvirtualization#ApplicationGroupsClient.ListByResourceGroup
	applicationGroups, err := ar.ListApplicationGroups(ctx, "")
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

// Logic Resources

func logicAppWorkflows(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	appWorkflows, err := ar.ListWorkflows(ctx, nil, "")
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

func logicAppTriggerCustoms(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	appWorkflowNames, err := getWorkflowNames(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list logic app workflows from reader")
	}

	resources := make([]provider.Resource, 0)
	for _, appWorkflowName := range appWorkflowNames {
		triggers, err := ar.ListWorkflowTriggers(ctx, appWorkflowName, nil, "")
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

func logicAppActionCustoms(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	appWorkflowNames, err := getWorkflowNames(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list logic app workflows from reader")
	}

	resources := make([]provider.Resource, 0)
	for _, appWorkflowName := range appWorkflowNames {
		runs, err := ar.ListWorkflowRuns(ctx, appWorkflowName, nil, "")
		if err != nil {
			return nil, errors.Wrap(err, "unable to list workflow runs from reader")
		}

		for _, run := range runs {
			actions, err := ar.ListWorkflowRunActions(ctx, appWorkflowName, *run.Name, nil, "")
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

// Container Registry Resources

func containerRegistries(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	containerRegistries, err := ar.ListContainerRegistries(ctx)
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

func containerRegistryWebhooks(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	containerRegistriesNames, err := getContainerRegistries(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list container Registries from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, containerRegistryName := range containerRegistriesNames {
		containerRegistryWebhooks, err := ar.ListContainerRegistryWebhooks(ctx, containerRegistryName)
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

// Storage Resources

func storageAccounts(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	storageAccounts, err := ar.ListSTORAGEAccounts(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list storage accounts from reader")
	}
	resources := make([]provider.Resource, 0, len(storageAccounts))
	for _, storageAccount := range storageAccounts {
		r := provider.NewResource(*storageAccount.ID, resourceType, a)
		if err := r.Data().Set("name", *storageAccount.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the storage accounts '%s'", *storageAccount.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func storageQueues(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	storageAccountNames, err := getStorageAccounts(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list storage Accounts from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, storageAccountName := range storageAccountNames {
		// last 2 args of list function "" because they're optional
		// https://github.com/Azure/azure-sdk-for-go/blob/main/services/storage/mgmt/2021-04-01/storage/queue.go#:~:text=//-,List,-gets%20a%20list
		// maxpagesize - optional, a maximum number of queues that should be included in a list queue response
		// filter - optional, When specified, only the queues with a name starting with the given filter will be
		storageQueues, err := ar.ListSTORAGEQueue(ctx, storageAccountName, "", "")
		if err != nil {
			return nil, errors.Wrap(err, "unable to list storage queues from reader")
		}
		for _, storageQueue := range storageQueues {
			r := provider.NewResource(*storageQueue.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func storageShares(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	storageAccountNames, err := getStorageAccounts(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list storage Accounts from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, storageAccountName := range storageAccountNames {
		// last 3 args of list function "" because they're optional
		// https://github.com/Azure/azure-sdk-for-go/blob/main/services/storage/mgmt/2021-04-01/storage/fileshares.go#:~:text=//-,List,-lists%20all%20shares
		// maxpagesize - optional, a maximum number of queues that should be included in a list queue response
		// filter - optional, When specified, only the queues with a name starting with the given filter will be
		// expand - optional, used to expand the properties within share's properties.
		storageFileShares, err := ar.ListSTORAGEFileShares(ctx, storageAccountName, "", "", "")
		if err != nil {
			return nil, errors.Wrap(err, "unable to list storage fileshares from reader")
		}
		for _, storageFileShare := range storageFileShares {
			r := provider.NewResource(*storageFileShare.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func storageBlobs(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	storageAccountNames, err := getStorageAccounts(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list storage Accounts from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, storageAccountName := range storageAccountNames {
		// last 3 args of list function "" because they're optional
		// https://github.com/Azure/azure-sdk-for-go/blob/main/services/storage/mgmt/2021-04-01/storage/blobcontainers.go#:~:text=//%20List-,lists,-all%20containers%20and
		// maxpagesize - optional, a maximum number of queues that should be included in a list queue response
		// filter - optional, When specified, only the queues with a name starting with the given filter will be
		// expand - optional, used to expand the properties within share's properties.
		storageBlobs, err := ar.ListSTORAGEBlobContainers(ctx, storageAccountName, "", "", "")
		if err != nil {
			return nil, errors.Wrap(err, "unable to list storage blobs from reader")
		}
		for _, storageBlob := range storageBlobs {
			r := provider.NewResource(*storageBlob.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func storageTables(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	storageAccountNames, err := getStorageAccounts(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list storage Accounts from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, storageAccountName := range storageAccountNames {
		storageTables, err := ar.ListSTORAGETable(ctx, storageAccountName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list storage table from reader")
		}
		for _, storageTable := range storageTables {
			r := provider.NewResource(*storageTable.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

// Database Resources- mariadb

func mariadbServers(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	mariadbServers, err := ar.ListMARIADBServers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list MariaDB Servers from reader")
	}
	resources := make([]provider.Resource, 0, len(mariadbServers))
	for _, mariadbServer := range mariadbServers {
		r := provider.NewResource(*mariadbServer.ID, resourceType, a)
		if err := r.Data().Set("name", *mariadbServer.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the MariaDB Server '%s'", *mariadbServer.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func mariadbConfigurations(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	mariadbServerNames, err := getMariadbServers(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Mariadb Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, mariadbServerName := range mariadbServerNames {
		mariadbConfigurations, err := ar.ListMARIADBConfigurations(ctx, mariadbServerName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list mariadb configurations from reader")
		}
		for _, mariadbConfiguration := range mariadbConfigurations {
			r := provider.NewResource(*mariadbConfiguration.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func mariadbDatabases(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	mariadbServerNames, err := getMariadbServers(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Mariadb Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, mariadbServerName := range mariadbServerNames {
		mariadbDatabases, err := ar.ListMARIADBDatabases(ctx, mariadbServerName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list mariadb databases from reader")
		}
		for _, mariadbDatabase := range mariadbDatabases {
			r := provider.NewResource(*mariadbDatabase.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func mariadbFirewallRules(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	mariadbServerNames, err := getMariadbServers(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Mariadb Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, mariadbServerName := range mariadbServerNames {
		mariadbFirewallRules, err := ar.ListMARIADBFirewallRules(ctx, mariadbServerName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list mariadb firewall rules from reader")
		}
		for _, mariadbFirewallRule := range mariadbFirewallRules {
			r := provider.NewResource(*mariadbFirewallRule.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func mariadbVirtualNetworkRules(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	mariadbServerNames, err := getMariadbServers(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Mariadb Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, mariadbServerName := range mariadbServerNames {
		mariadbVirtualNetworkRules, err := ar.ListMARIADBVirtualNetworkRules(ctx, mariadbServerName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list mariadb firewall rules from reader")
		}
		for _, mariadbVirtualNetworkRule := range mariadbVirtualNetworkRules {
			r := provider.NewResource(*mariadbVirtualNetworkRule.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

// Database Resources- mysql

func mysqlServers(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	mysqlServers, err := ar.ListMYSQLServers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list MySQL Servers from reader")
	}
	resources := make([]provider.Resource, 0, len(mysqlServers))
	for _, mysqlServer := range mysqlServers {
		r := provider.NewResource(*mysqlServer.ID, resourceType, a)
		if err := r.Data().Set("name", *mysqlServer.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the MySQL Server '%s'", *mysqlServer.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func mysqlConfigurations(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	mysqlServerNames, err := getMysqlServers(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list MySQL Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, mysqlServerName := range mysqlServerNames {
		mysqlConfigurations, err := ar.ListMYSQLConfigurations(ctx, mysqlServerName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list MySQL configurations from reader")
		}
		for _, mysqlConfiguration := range mysqlConfigurations {
			r := provider.NewResource(*mysqlConfiguration.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func mysqlDatabases(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	mysqlServerNames, err := getMysqlServers(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list MySQL Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, mysqlServerName := range mysqlServerNames {
		mysqlDatabases, err := ar.ListMYSQLDatabases(ctx, mysqlServerName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list MySQL databases from reader")
		}
		for _, mysqlDatabase := range mysqlDatabases {
			r := provider.NewResource(*mysqlDatabase.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func mysqlFirewallRules(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	mysqlServerNames, err := getMysqlServers(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list MySQL Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, mysqlServerName := range mysqlServerNames {
		mysqlFirewallRules, err := ar.ListMYSQLFirewallRules(ctx, mysqlServerName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list MySQL firewall rules from reader")
		}
		for _, mysqlFirewallRule := range mysqlFirewallRules {
			r := provider.NewResource(*mysqlFirewallRule.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func mysqlVirtualNetworkRules(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	mysqlServerNames, err := getMysqlServers(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list MySQL Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, mysqlServerName := range mysqlServerNames {
		mysqlVirtualNetworkRules, err := ar.ListMYSQLVirtualNetworkRules(ctx, mysqlServerName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list MySQL firewall rules from reader")
		}
		for _, mysqlVirtualNetworkRule := range mysqlVirtualNetworkRules {
			r := provider.NewResource(*mysqlVirtualNetworkRule.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

// Database Resources- PostgreSQL

func postgresqlServers(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	postgresqlServers, err := ar.ListPOSTGRESQLServers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list PostgreSQL Servers from reader")
	}
	resources := make([]provider.Resource, 0, len(postgresqlServers))
	for _, postgresqlServer := range postgresqlServers {
		r := provider.NewResource(*postgresqlServer.ID, resourceType, a)
		if err := r.Data().Set("name", *postgresqlServer.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the PostgreSQL Server '%s'", *postgresqlServer.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func postgresqlConfigurations(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	postgresqlServerNames, err := getPostgresqlServers(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list PostgreSQL Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, postgresqlServerName := range postgresqlServerNames {
		postgresqlConfigurations, err := ar.ListPOSTGRESQLConfigurations(ctx, postgresqlServerName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list PostgreSQL configurations from reader")
		}
		for _, postgresqlConfiguration := range postgresqlConfigurations {
			r := provider.NewResource(*postgresqlConfiguration.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func postgresqlDatabases(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	postgresqlServerNames, err := getPostgresqlServers(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list PostgreSQL Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, postgresqlServerName := range postgresqlServerNames {
		postgresqlDatabases, err := ar.ListPOSTGRESQLDatabases(ctx, postgresqlServerName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list PostgreSQL databases from reader")
		}
		for _, postgresqlDatabase := range postgresqlDatabases {
			r := provider.NewResource(*postgresqlDatabase.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func postgresqlFirewallRules(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	postgresqlServerNames, err := getPostgresqlServers(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list PostgreSQL Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, postgresqlServerName := range postgresqlServerNames {
		postgresqlFirewallRules, err := ar.ListPOSTGRESQLFirewallRules(ctx, postgresqlServerName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list PostgreSQL firewall rules from reader")
		}
		for _, postgresqlFirewallRule := range postgresqlFirewallRules {
			r := provider.NewResource(*postgresqlFirewallRule.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func postgresqlVirtualNetworkRules(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	postgresqlServerNames, err := getPostgresqlServers(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list PostgreSQL Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, postgresqlServerName := range postgresqlServerNames {
		postgresqlVirtualNetworkRules, err := ar.ListPOSTGRESQLVirtualNetworkRules(ctx, postgresqlServerName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list PostgreSQL firewall rules from reader")
		}
		for _, postgresqlVirtualNetworkRule := range postgresqlVirtualNetworkRules {
			r := provider.NewResource(*postgresqlVirtualNetworkRule.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

// Database Resources- SQL

func sqlServers(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sqlServers, err := ar.ListSQLServers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list SQL Servers from reader")
	}
	resources := make([]provider.Resource, 0, len(sqlServers))
	for _, sqlServer := range sqlServers {
		r := provider.NewResource(*sqlServer.ID, resourceType, a)
		if err := r.Data().Set("name", *sqlServer.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the SQL Server '%s'", *sqlServer.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func sqlElasticPools(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sqlServerNames, err := getSQLServers(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list SQL Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, sqlServerName := range sqlServerNames {
		sqlElasticPools, err := ar.ListSQLElasticPools(ctx, sqlServerName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list SQL Elastic Pools from reader")
		}
		for _, sqlElasticPool := range sqlElasticPools {
			r := provider.NewResource(*sqlElasticPool.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func sqlDatabases(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sqlServerNames, err := getSQLServers(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list SQL Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, sqlServerName := range sqlServerNames {
		// last 2 args of list function "" because they're not required
		// https://github.com/Azure/azure-sdk-for-go/blob/main/services/sql/mgmt/2014-04-01/sql/databases.go#:~:text=func%20(client%20DatabasesClient)-,ListByServer,-(ctx%20context.Context
		// expand - expand - a comma separated list of child objects to expand in the response.
		// filter - an OData filter expression that describes a subset of databases to return.
		sqlDatabases, err := ar.ListSQLDatabases(ctx, sqlServerName, "", "")
		if err != nil {
			return nil, errors.Wrap(err, "unable to list SQL databases from reader")
		}
		for _, sqlDatabase := range sqlDatabases {
			r := provider.NewResource(*sqlDatabase.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func sqlFirewallRules(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sqlServerNames, err := getSQLServers(ctx, a, ar, resourceType, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list SQL Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, sqlServerName := range sqlServerNames {
		sqlFirewallRules, err := ar.ListSQLFirewallRules(ctx, sqlServerName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list SQL firewall rules from reader")
		}
		for _, sqlFirewallRule := range sqlFirewallRules {
			r := provider.NewResource(*sqlFirewallRule.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}
