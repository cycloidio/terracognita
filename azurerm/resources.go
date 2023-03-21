package azurerm

import (
	"context"
	"fmt"
	"strings"

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
	AvailabilitySet
	Image
	ManagedDisk
	VirtualMachine
	VirtualMachineDataDiskAttachment
	VirtualMachineExtension
	VirtualMachineScaleSetExtension
	VirtualNetwork
	LinuxVirtualMachine
	LinuxVirtualMachineScaleSet
	WindowsVirtualMachine
	WindowsVirtualMachineScaleSet
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
	VirtualHub
	VirtualHubBgpConnection
	VirtualHubConnection
	VirtualHubIP
	VirtualHubRouteTable
	VirtualHubSecurityPartnerProvider
	// Load Balancer
	Lb
	LbBackendAddressPool
	LbRule
	LbOutboundRule
	LbNatRule
	LbNatPool
	LbProbe
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
	// Container Service Resources - k8s services
	KubernetesCluster
	KubernetesClusterNodePool
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
	// Database Resources- mssql
	MssqlElasticpool
	MssqlDatabase
	MssqlFirewallRule
	MssqlServer
	MssqlServerSecurityAlertPolicy
	MssqlServerVulnerabilityAssessment
	MssqlVirtualMachine
	MssqlVirtualNetworkRule
	// Redis
	RedisCache
	RedisFirewallRule
	// DNS
	DNSZone
	DNSARecord //dns_a_record
	DNSAaaaRecord
	DNSCaaRecord
	DNSCnameRecord
	DNSMxRecord
	DNSNsRecord
	DNSPtrRecord
	DNSSrvRecord
	DNSTxtRecord
	// Private DNS
	PrivateDNSZone
	PrivateDNSARecord //private_dns_a_record
	PrivateDNSAaaaRecord
	PrivateDNSCnameRecord
	PrivateDNSMxRecord
	PrivateDNSPtrRecord
	PrivateDNSSrvRecord
	PrivateDNSTxtRecord
	PrivateDNSZoneVirtualNetworkLink
	// Policy
	PolicyDefinition
	PolicyRemediation
	PolicySetDefinition
	// Vault
	KeyVault
	KeyVaultAccessPolicy
	// Application Insigths
	ApplicationInsights
	ApplicationInsightsAPIKey
	ApplicationInsightsAnalyticsItem
	//ApplicationInsightsWebTest
	// Log Analytics
	LogAnalyticsWorkspace
	LogAnalyticsLinkedService
	LogAnalyticsDatasourceWindowsPerformanceCounter
	LogAnalyticsDatasourceWindowsEvent
	// Monitor
	MonitorActionGroup
	MonitorActivityLogAlert
	MonitorAutoscaleSetting
	MonitorLogProfile
	MonitorMetricAlert
	// App service
	WindowsWebApp
	LinuxWebApp
	LinuxWebAppSlot
	WindowsWebAppSlot
	WebAppActiveSlot
	ServicePlan
	SourceControlToken
	StaticSite
	StaticSiteCustomDomain
	WebAppHybridConnection
	// dataprotection
	DataProtectionBackupVault
)

type rtFn func(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error)

var (
	resources = map[ResourceType]rtFn{
		ResourceGroup: resourceGroup,
		// Compute Resources
		VirtualMachine:                   virtualMachines,
		WindowsVirtualMachine:            virtualMachines,
		LinuxVirtualMachine:              virtualMachines,
		VirtualMachineExtension:          virtualMachineExtensions,
		VirtualNetwork:                   virtualNetworks,
		WindowsVirtualMachineScaleSet:    virtualMachineScaleSets,
		LinuxVirtualMachineScaleSet:      virtualMachineScaleSets,
		VirtualMachineScaleSetExtension:  virtualMachineScaleSetExtensions,
		AvailabilitySet:                  availabilitySets,
		ManagedDisk:                      disks,
		VirtualMachineDataDiskAttachment: virtualMachineDataDiskAttachments,
		Image:                            images,
		// Network Resources
		Subnet:                            subnets,
		NetworkInterface:                  networkInterfaces,
		NetworkSecurityGroup:              networkSecurityGroups,
		ApplicationGateway:                applicationGateways,
		ApplicationSecurityGroup:          applicationSecurityGroups,
		NetworkDdosProtectionPlan:         networkddosProtectionPlans,
		Firewall:                          firewalls,
		LocalNetworkGateway:               localNetworkGateways,
		NatGateway:                        natGateways,
		NetworkProfile:                    networkProfiles,
		NetworkSecurityRule:               networkSecurityRules,
		PublicIP:                          publicIP,
		PublicIPPrefix:                    publicIPPrefixes,
		Route:                             routes,
		RouteTable:                        routeTables,
		VirtualNetworkGateway:             virtualNetworkGateways,
		VirtualNetworkGatewayConnection:   virtualNetworkGatewayConnections,
		VirtualNetworkPeering:             virtualNetworkPeerings,
		WebApplicationFirewallPolicy:      webApplicationFirewallPolicies,
		VirtualHub:                        virtualHubs,
		VirtualHubBgpConnection:           virtualHubBgpConnection,
		VirtualHubConnection:              virtualHubConnection,
		VirtualHubIP:                      virtualHubIP,
		VirtualHubRouteTable:              virtualHubRouteTable,
		VirtualHubSecurityPartnerProvider: virtualHubSecurityPartnerProvider,
		// Load Balancer
		Lb:                   lbs,
		LbBackendAddressPool: lbBackendAddressPools,
		LbRule:               lbProperties,
		LbOutboundRule:       lbProperties,
		LbNatRule:            lbProperties,
		LbNatPool:            lbProperties,
		LbProbe:              lbProperties,
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
		// Container Service Resources
		KubernetesCluster:         kubernetesClusters,
		KubernetesClusterNodePool: kubernetesClustersNodePools,
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
		// Database Resources- mssql
		MssqlElasticpool:                   mssqlElasticPools,
		MssqlDatabase:                      mssqlDatabases,
		MssqlFirewallRule:                  mssqlFirewallRules,
		MssqlServer:                        mssqlServers,
		MssqlServerSecurityAlertPolicy:     mssqlServerSecurityAlertPolicies,
		MssqlServerVulnerabilityAssessment: mssqlServerVulnerabilityAssessments,
		MssqlVirtualMachine:                mssqlVirtualMachines,
		MssqlVirtualNetworkRule:            mssqlVirtualNetworkRules,
		// Redis
		RedisCache:        redisCaches,
		RedisFirewallRule: redisFirewallRules,
		// 	Dns
		DNSZone:        dnsZones,
		DNSARecord:     dnsRecordSets,
		DNSAaaaRecord:  dnsRecordSets,
		DNSCaaRecord:   dnsRecordSets,
		DNSCnameRecord: dnsRecordSets,
		DNSMxRecord:    dnsRecordSets,
		DNSNsRecord:    dnsRecordSets,
		DNSPtrRecord:   dnsRecordSets,
		DNSSrvRecord:   dnsRecordSets,
		DNSTxtRecord:   dnsRecordSets,
		// Private DNS
		PrivateDNSZone:                   privateDNSZones,
		PrivateDNSARecord:                privateDNSRecordSets,
		PrivateDNSAaaaRecord:             privateDNSRecordSets,
		PrivateDNSCnameRecord:            privateDNSRecordSets,
		PrivateDNSMxRecord:               privateDNSRecordSets,
		PrivateDNSPtrRecord:              privateDNSRecordSets,
		PrivateDNSSrvRecord:              privateDNSRecordSets,
		PrivateDNSTxtRecord:              privateDNSRecordSets,
		PrivateDNSZoneVirtualNetworkLink: privateDNSVirtualNetworkLinks,
		// Policy
		PolicyDefinition:    policyDefinitions,
		PolicyRemediation:   policyRemediations,
		PolicySetDefinition: policySetDefinitions,
		// Vault
		KeyVault:             keyVaults,
		KeyVaultAccessPolicy: keyVaultProperties,
		// Application Insigths
		ApplicationInsights:              applicationInsights,
		ApplicationInsightsAPIKey:        applicationInsightsAPIKeys,
		ApplicationInsightsAnalyticsItem: applicationInsightsAnalyticsItems,
		//ApplicationInsightsWebTest:       applicationInsightsWebTests,
		// Log Analytics
		LogAnalyticsWorkspace:                           logAnalyticsWorkspaces,
		LogAnalyticsLinkedService:                       logAnalyticsLinkedServices,
		LogAnalyticsDatasourceWindowsPerformanceCounter: logAnalyticsDatasources,
		LogAnalyticsDatasourceWindowsEvent:              logAnalyticsDatasources,
		// Monitor
		MonitorActionGroup:      monitorActionGroups,
		MonitorActivityLogAlert: monitorActivityLogAlerts,
		MonitorAutoscaleSetting: monitorAutoscaleSettings,
		MonitorLogProfile:       monitorLogProfiles,
		MonitorMetricAlert:      monitorMetricAlerts,
		// App service
		WindowsWebApp:          webApps,
		LinuxWebApp:            webApps,
		LinuxWebAppSlot:        linuxWebAppSlots,
		WindowsWebAppSlot:      windowsWebAppSlots,
		WebAppActiveSlot:       webAppActiveSlots,
		ServicePlan:            servicePlans,
		SourceControlToken:     sourceControlTokens,
		StaticSite:             staticSites,
		StaticSiteCustomDomain: staticSiteCustomDomains,
		WebAppHybridConnection: webAppHybridConnections,
		// dataprotection
		DataProtectionBackupVault: dataProtectionBackupVaults,
	}
)

func resourceGroup(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	resourceGroup := ar.GetResourceGroup()
	r := provider.NewResource(*resourceGroup.ID, resourceType, a)
	resources := []provider.Resource{r}
	return resources, nil
}

func filterByTags(f *filter.Filter, tags map[string]*string) bool {
	if len(f.Tags) == 0 {
		return true
	}
	for _, t := range f.Tags {
		if v, ok := tags[t.Name]; ok {
			if v == nil {
				continue
			}
			if *v == t.Value {
				return true
			}
		}
	}
	return false
}

// Compute Resources

func virtualMachines(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualMachines, err := ar.ListVirtualMachines(ctx, "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list virtual machines from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, virtualMachine := range virtualMachines {
		if !filterByTags(filters, virtualMachine.Tags) {
			continue
		}

		// To avoid having the same vm for different resources (azurerm_virtual_machine and azurerm_windows_virtual_machine or azurerm_linux_virtual_machine)
		// Check VM OS (based on the criteria to create specific os type vm resources)
		// then based on the result we check if the vm should be added to the list for the specific resource_type or not
		// checks the storageProfile OS disk, check thats unmanaged disks are not used (mandatory for the resource) and also that the osProfile is attached with the OS configuration
		// for more info about the implentation: https://github.com/hashicorp/terraform-provider-azurerm/blob/main/internal/services/compute/virtual_machine_import.go
		vmOS := ""

		if storageProfile := virtualMachine.VirtualMachineProperties.StorageProfile; storageProfile != nil && storageProfile.OsDisk.Vhd == nil {
			// type windows
			if storageProfile.OsDisk.OsType == "Windows" {
				if osProfile := virtualMachine.VirtualMachineProperties.OsProfile; osProfile != nil && osProfile.WindowsConfiguration != nil {
					vmOS = "windows"
				}

				//type linux
			} else if storageProfile.OsDisk.OsType == "Linux" {
				if osProfile := virtualMachine.VirtualMachineProperties.OsProfile; osProfile != nil && osProfile.LinuxConfiguration != nil {
					vmOS = "linux"
				}
			}
		}

		// if resource_type is azurerm_virtual_machine
		// and vmOS was retrived (not null)
		// and the corresponding os vm resource is included
		// then dont import vm
		if resourceType == "azurerm_virtual_machine" && vmOS != "" && filters.IsIncluded("azurerm_virtual_machine", "azurerm_"+vmOS+"_virtual_machine") {
			continue

			// if resource_type is azurerm_linux|windows_virtual_machine
			// and a vmOS was retrieved (not null)
			// and the resource_type doesn't contain contain the vmOS
			// then don't import vm
		} else if (resourceType == "azurerm_linux_virtual_machine" || resourceType == "azurerm_windows_virtual_machine") && vmOS != "" && !strings.Contains(resourceType, vmOS) {
			continue
		}

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
	resources := make([]provider.Resource, 0)
	for _, virtualMachineScaleSet := range virtualMachineScaleSets {
		if !filterByTags(filters, virtualMachineScaleSet.Tags) {
			continue
		}

		// if resource_type is one of the elements of vm and not a caching method
		if resourceType == "azurerm_linux_virtual_machine_scale_set" || resourceType == "azurerm_windows_virtual_machine_scale_set" {

			//check scale set VM OS (based on the criteria to create specific os type vm resources)
			// then based on the result we check if the vm should be added to the list for the specific resource_type or not
			// to avoid having the same vm for different resources (azurerm_virtual_machine_scale_set and azurerm_windows_virtual_machine_scale_set or azurerm_linux_virtual_machine_scale_set)
			// for more info about the implentation: https://github.com/hashicorp/terraform-provider-azurerm/blob/main/internal/services/compute/virtual_machine_scale_set_import.go

			vmOS := ""

			if osProfile := virtualMachineScaleSet.VirtualMachineScaleSetProperties.VirtualMachineProfile.OsProfile; osProfile != nil {
				// type windows
				if osProfile.WindowsConfiguration != nil {
					vmOS = "windows"

					//type linux
				} else if osProfile.LinuxConfiguration != nil {
					vmOS = "linux"
				}
			}

			// ifthe resource_type doesn't contain the vmOS
			// then don't import vm
			if !strings.Contains(resourceType, vmOS) {
				continue
			}
		}

		r := provider.NewResource(*virtualMachineScaleSet.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *virtualMachineScaleSet.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the virtual machine '%s'", *virtualMachineScaleSet.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func virtualMachineScaleSetExtensions(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	scaleSetNames, err := getVirtualMachineScaleSetNames(ctx, a, ar, []string{WindowsVirtualMachineScaleSet.String(), LinuxVirtualMachineScaleSet.String()}, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list virtual machines scale sets from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, scaleSetNames := range scaleSetNames {
		extensions, err := ar.ListVirtualMachineScaleSetExtensions(ctx, scaleSetNames)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list list virtual machines scale set extensions from reader")
		}
		for _, extension := range extensions {
			r := provider.NewResource(*extension.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func disks(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	disks, err := ar.ListDisks(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list disks from reader")
	}
	resources := make([]provider.Resource, 0, len(disks))
	for _, disk := range disks {
		if !filterByTags(filters, disk.Tags) {
			continue
		}
		// If disk is used as Operating System, the disk is managed by the virtual_machine resource, not a dedicated disk
		if disk.DiskProperties.OsType != "" {
			continue
		}

		// When using azurerm_virtual_machine resource, extra attached disk are managed via storage_data_disk
		// CreateOption == Empty : fully managed by
		if (disk.DiskProperties.DiskState == "Attached" || disk.DiskProperties.DiskState == "Reserved") && filters.IsIncluded("azurerm_virtual_machine") {
			continue
		}
		r := provider.NewResource(*disk.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func virtualMachineDataDiskAttachments(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	// only Managed Disks are supported via this separate resource,
	// Unmanaged Disks can be attached using the storage_data_disk block in the azurerm_virtual_machine resource.
	// So if using azurerm_virtual_machine, do not define azurerm_virtual_machine_data_disk_attachment.
	if filters.IsIncluded("azurerm_virtual_machine") {
		return nil, nil
	}

	// Get the list of disks
	disks, err := ar.ListDisks(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list disks attachments from reader")
	}

	// Get the vms list to check if disk attached
	virtualMachines, err := ar.ListVirtualMachines(ctx, "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list virtual machines from reader")
	}

	resources := make([]provider.Resource, 0)
	for _, disk := range disks {
		if !filterByTags(filters, disk.Tags) {
			continue
		}
		if disk.DiskProperties.DiskState == "Attached" || disk.DiskProperties.DiskState == "Reserved" {
			// check on wich VM the disk is attached
			for _, virtualMachine := range virtualMachines {
				if profile := virtualMachine.StorageProfile; profile != nil {
					if dataDisks := profile.DataDisks; dataDisks != nil {
						for _, dataDisk := range *dataDisks {
							if *dataDisk.Name == *disk.Name {
								r := provider.NewResource(fmt.Sprintf("%s/dataDisks/%s", *virtualMachine.ID, *disk.Name), resourceType, a)
								resources = append(resources, r)
								break
							}
						}
					}
				}
			}
		}
	}
	return resources, nil
}

func virtualMachineExtensions(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualMachineNames, err := getVirtualMachineNames(ctx, a, ar, []string{VirtualMachine.String(), WindowsVirtualMachine.String(), LinuxVirtualMachine.String()}, filters)
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
			if !filterByTags(filters, extension.Tags) {
				continue
			}
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
		if !filterByTags(filters, availabilitySet.Tags) {
			continue
		}
		r := provider.NewResource(*availabilitySet.ID, resourceType, a)
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
		if !filterByTags(filters, image.Tags) {
			continue
		}
		r := provider.NewResource(*image.ID, resourceType, a)
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
		if !filterByTags(filters, virtualNetwork.Tags) {
			continue
		}
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
	virtualNetworkNames, err := getVirtualNetworkNames(ctx, a, ar, VirtualNetwork.String(), filters)
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
		if !filterByTags(filters, networkInterface.Tags) {
			continue
		}
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
		if !filterByTags(filters, securityGroup.Tags) {
			continue
		}
		r := provider.NewResource(*securityGroup.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *securityGroup.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the virtual machine '%s'", *securityGroup.Name)
		}
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
		if !filterByTags(filters, applicationSecurityGroup.Tags) {
			continue
		}
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
		if !filterByTags(filters, ddosProtectionPlan.Tags) {
			continue
		}
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
		if !filterByTags(filters, azureFirewall.Tags) {
			continue
		}
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
		if !filterByTags(filters, localNetworkGateway.Tags) {
			continue
		}
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
		if !filterByTags(filters, natGateway.Tags) {
			continue
		}
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
		if !filterByTags(filters, profile.Tags) {
			continue
		}
		r := provider.NewResource(*profile.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func networkSecurityRules(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	securityGroupNames, err := getSecurityGroups(ctx, a, ar, NetworkSecurityGroup.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list security Groups from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, securityGroupName := range securityGroupNames {
		securityRules, err := ar.ListSecurityRules(ctx, securityGroupName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list security rules from reader")
		}
		for _, securityRule := range securityRules {
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
		if !filterByTags(filters, publicIPAddress.Tags) {
			continue
		}
		r := provider.NewResource(*publicIPAddress.ID, resourceType, a)
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
		if !filterByTags(filters, publicIPPrefix.Tags) {
			continue
		}
		r := provider.NewResource(*publicIPPrefix.ID, resourceType, a)
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
		if !filterByTags(filters, routeTable.Tags) {
			continue
		}
		r := provider.NewResource(*routeTable.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *routeTable.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the routeTable '%s'", *routeTable.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func routes(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	routeTablesNames, err := getRouteTables(ctx, a, ar, RouteTable.String(), filters)
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
		if !filterByTags(filters, virtualNetworkGateway.Tags) {
			continue
		}
		r := provider.NewResource(*virtualNetworkGateway.ID, resourceType, a)
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
		if !filterByTags(filters, virtualNetworkGatewayConnection.Tags) {
			continue
		}
		r := provider.NewResource(*virtualNetworkGatewayConnection.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func virtualNetworkPeerings(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualNetworkNames, err := getVirtualNetworkNames(ctx, a, ar, VirtualNetwork.String(), filters)
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
		if !filterByTags(filters, webApplicationFirewallPolicy.Tags) {
			continue
		}
		r := provider.NewResource(*webApplicationFirewallPolicy.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func virtualHubs(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualHubs, err := ar.ListVirtualHubs(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Virtual Hubs from reader")
	}
	resources := make([]provider.Resource, 0, len(virtualHubs))
	for _, virtualHub := range virtualHubs {
		if !filterByTags(filters, virtualHub.Tags) {
			continue
		}
		r := provider.NewResource(*virtualHub.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *virtualHub.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the virtual Hub '%s'", *virtualHub.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func virtualHubBgpConnection(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualHubNames, err := getVirtualHub(ctx, a, ar, VirtualHub.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list virtual hubs from reader")
	}

	resources := make([]provider.Resource, 0)
	for _, virtualHubName := range virtualHubNames {
		virtualHubBgpConnections, err := ar.ListVirtualHubBgpConnections(ctx, virtualHubName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list virtual hub BGP connections from reader")
		}
		for _, virtualHubBgpConnection := range virtualHubBgpConnections {
			r := provider.NewResource(*virtualHubBgpConnection.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func virtualHubConnection(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualHubNames, err := getVirtualHub(ctx, a, ar, VirtualHub.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list virtual hubs from reader")
	}

	resources := make([]provider.Resource, 0)
	for _, virtualHubName := range virtualHubNames {
		virtualHubConnections, err := ar.ListHubVirtualNetworkConnections(ctx, virtualHubName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list virtual hub connections from reader")
		}
		for _, virtualHubConnection := range virtualHubConnections {
			r := provider.NewResource(*virtualHubConnection.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func virtualHubIP(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualHubNames, err := getVirtualHub(ctx, a, ar, VirtualHub.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list virtual hubs from reader")
	}

	resources := make([]provider.Resource, 0)
	for _, virtualHubName := range virtualHubNames {
		virtualHubIPs, err := ar.ListVirtualHubIPConfiguration(ctx, virtualHubName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list virtual hub IPs from reader")
		}
		for _, virtualHubIP := range virtualHubIPs {
			r := provider.NewResource(*virtualHubIP.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func virtualHubRouteTable(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualHubNames, err := getVirtualHub(ctx, a, ar, VirtualHub.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list virtual hubs from reader")
	}

	resources := make([]provider.Resource, 0)
	for _, virtualHubName := range virtualHubNames {
		virtualHubRouteTables, err := ar.ListHubRouteTables(ctx, virtualHubName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list virtual hub route tables from reader")
		}
		for _, virtualHubRouteTable := range virtualHubRouteTables {
			r := provider.NewResource(*virtualHubRouteTable.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func virtualHubSecurityPartnerProvider(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	virtualHubSecurityPartnerProviders, err := ar.ListSecurityPartnerProviders(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Virtual Hubs security partner provider from reader")
	}
	resources := make([]provider.Resource, 0, len(virtualHubSecurityPartnerProviders))
	for _, virtualHubSecurityPartnerProvider := range virtualHubSecurityPartnerProviders {
		if !filterByTags(filters, virtualHubSecurityPartnerProvider.Tags) {
			continue
		}
		r := provider.NewResource(*virtualHubSecurityPartnerProvider.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

// Load Balancer
func lbs(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	lbs, err := ar.ListLoadBalancers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Load Balancer from reader")
	}
	resources := make([]provider.Resource, 0, len(lbs))
	for _, lb := range lbs {
		if !filterByTags(filters, lb.Tags) {
			continue
		}
		r := provider.NewResource(*lb.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *lb.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the load balancer '%s'", *lb.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func lbBackendAddressPools(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	lbNames, err := getLbs(ctx, a, ar, Lb.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list load balancers from reader")
	}

	resources := make([]provider.Resource, 0)
	for _, lbName := range lbNames {
		lbBackendAddressPools, err := ar.ListLoadBalancerBackendAddressPools(ctx, lbName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list lb backend address pools from reader")
		}
		for _, lbBackendAddressPool := range lbBackendAddressPools {
			r := provider.NewResource(*lbBackendAddressPool.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func lbProperties(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	lbs, err := ar.ListLoadBalancers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Load Balancer from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, lb := range lbs {
		if !filterByTags(filters, lb.Tags) {
			continue
		}
		if lbProperties := lb.LoadBalancerPropertiesFormat; lbProperties != nil {
			if resourceType == "azurerm_lb_rule" && lbProperties.LoadBalancingRules != nil {
				for _, lbRule := range *lbProperties.LoadBalancingRules {
					r := provider.NewResource(*lbRule.ID, resourceType, a)
					resources = append(resources, r)
				}
			} else if resourceType == "azurerm_lb_outbound_rule" && lbProperties.OutboundRules != nil {
				for _, outboundRule := range *lbProperties.OutboundRules {
					r := provider.NewResource(*outboundRule.ID, resourceType, a)
					resources = append(resources, r)
				}
			} else if resourceType == "azurerm_lb_nat_rule" && lbProperties.InboundNatRules != nil {
				for _, natRule := range *lbProperties.InboundNatRules {
					r := provider.NewResource(*natRule.ID, resourceType, a)
					resources = append(resources, r)
				}
			} else if resourceType == "azurerm_lb_nat_pool" && lbProperties.InboundNatPools != nil {
				for _, natPool := range *lbProperties.InboundNatPools {
					r := provider.NewResource(*natPool.ID, resourceType, a)
					resources = append(resources, r)
				}
			} else if resourceType == "azurerm_lb_probe" && lbProperties.Probes != nil {
				for _, probe := range *lbProperties.Probes {
					r := provider.NewResource(*probe.ID, resourceType, a)
					resources = append(resources, r)
				}
			}
		}
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
		if !filterByTags(filters, hostPool.Tags) {
			continue
		}
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
		if !filterByTags(filters, applicationGroup.Tags) {
			continue
		}
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
		if !filterByTags(filters, appWorkflow.Tags) {
			continue
		}
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
	appWorkflowNames, err := getWorkflowNames(ctx, a, ar, LogicAppWorkflow.String(), filters)
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
	appWorkflowNames, err := getWorkflowNames(ctx, a, ar, LogicAppWorkflow.String(), filters)
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
		return nil, errors.Wrap(err, "unable to list container registries from reader")
	}
	resources := make([]provider.Resource, 0, len(containerRegistries))
	for _, containerRegistry := range containerRegistries {
		if !filterByTags(filters, containerRegistry.Tags) {
			continue
		}
		r := provider.NewResource(*containerRegistry.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *containerRegistry.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the container Registry'%s'", *containerRegistry.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func containerRegistryWebhooks(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	containerRegistriesNames, err := getContainerRegistries(ctx, a, ar, ContainerRegistry.String(), filters)
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
			if !filterByTags(filters, containerRegistryWebhook.Tags) {
				continue
			}
			r := provider.NewResource(*containerRegistryWebhook.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

// Container Service Resources

func kubernetesClusters(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	kubernetesClusters, err := ar.ListKubernetesClusters(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list kubernetes clusters from reader")
	}
	resources := make([]provider.Resource, 0, len(kubernetesClusters))
	for _, kubernetesCluster := range kubernetesClusters {
		if !filterByTags(filters, kubernetesCluster.Tags) {
			continue
		}
		r := provider.NewResource(*kubernetesCluster.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *kubernetesCluster.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the kubernetes cluster'%s'", *kubernetesCluster.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func kubernetesClustersNodePools(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	kubernetesClusters, err := getKubernetesClusters(ctx, a, ar, KubernetesCluster.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list kubernetes clusters from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, kubernetesCluster := range kubernetesClusters {
		kubernetesClustersNodePools, err := ar.ListKubernetesClusterNodes(ctx, kubernetesCluster)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list kubernetes clusters node pools from reader")
		}
		for _, kubernetesClustersNodePool := range kubernetesClustersNodePools {
			if !filterByTags(filters, kubernetesClustersNodePool.Tags) {
				continue
			}
			r := provider.NewResource(*kubernetesClustersNodePool.ID, resourceType, a)
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
		if !filterByTags(filters, storageAccount.Tags) {
			continue
		}
		r := provider.NewResource(*storageAccount.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *storageAccount.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the storage accounts '%s'", *storageAccount.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func storageQueues(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	storageAccountNames, err := getStorageAccounts(ctx, a, ar, StorageAccount.String(), filters)
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
	storageAccountNames, err := getStorageAccounts(ctx, a, ar, StorageAccount.String(), filters)
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
	storageAccountNames, err := getStorageAccounts(ctx, a, ar, StorageAccount.String(), filters)
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
	storageAccountNames, err := getStorageAccounts(ctx, a, ar, StorageAccount.String(), filters)
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
		if !filterByTags(filters, mariadbServer.Tags) {
			continue
		}
		r := provider.NewResource(*mariadbServer.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *mariadbServer.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the MariaDB Server '%s'", *mariadbServer.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func mariadbConfigurations(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	mariadbServerNames, err := getMariadbServers(ctx, a, ar, MariadbServer.String(), filters)
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
	mariadbServerNames, err := getMariadbServers(ctx, a, ar, MariadbServer.String(), filters)
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
	mariadbServerNames, err := getMariadbServers(ctx, a, ar, MariadbServer.String(), filters)
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
	mariadbServerNames, err := getMariadbServers(ctx, a, ar, MariadbServer.String(), filters)
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
		if !filterByTags(filters, mysqlServer.Tags) {
			continue
		}
		r := provider.NewResource(*mysqlServer.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *mysqlServer.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the MySQL Server '%s'", *mysqlServer.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func mysqlConfigurations(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	mysqlServerNames, err := getMysqlServers(ctx, a, ar, MysqlServer.String(), filters)
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
	mysqlServerNames, err := getMysqlServers(ctx, a, ar, MysqlServer.String(), filters)
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
	mysqlServerNames, err := getMysqlServers(ctx, a, ar, MysqlServer.String(), filters)
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
	mysqlServerNames, err := getMysqlServers(ctx, a, ar, MysqlServer.String(), filters)
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
		if !filterByTags(filters, postgresqlServer.Tags) {
			continue
		}
		r := provider.NewResource(*postgresqlServer.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *postgresqlServer.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the PostgreSQL Server '%s'", *postgresqlServer.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func postgresqlConfigurations(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	postgresqlServerNames, err := getPostgresqlServers(ctx, a, ar, PostgresqlServer.String(), filters)
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
	postgresqlServerNames, err := getPostgresqlServers(ctx, a, ar, PostgresqlServer.String(), filters)
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
	postgresqlServerNames, err := getPostgresqlServers(ctx, a, ar, PostgresqlServer.String(), filters)
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
	postgresqlServerNames, err := getPostgresqlServers(ctx, a, ar, PostgresqlServer.String(), filters)
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

// Database Resources- mssql

func mssqlServers(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sqlServers, err := ar.ListSQLServers(ctx, "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list SQL Servers from reader")
	}
	resources := make([]provider.Resource, 0, len(sqlServers))
	for _, sqlServer := range sqlServers {
		if !filterByTags(filters, sqlServer.Tags) {
			continue
		}
		r := provider.NewResource(*sqlServer.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *sqlServer.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the SQL Server '%s'", *sqlServer.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func mssqlElasticPools(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sqlServerNames, err := getMsSQLServers(ctx, a, ar, MssqlServer.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list SQL Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, sqlServerName := range sqlServerNames {
		sqlElasticPools, err := ar.ListSQLElasticPools(ctx, sqlServerName, nil)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list SQL Elastic Pools from reader")
		}
		for _, sqlElasticPool := range sqlElasticPools {
			if !filterByTags(filters, sqlElasticPool.Tags) {
				continue
			}
			r := provider.NewResource(*sqlElasticPool.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func mssqlDatabases(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sqlServerNames, err := getMsSQLServers(ctx, a, ar, MssqlServer.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list SQL Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, sqlServerName := range sqlServerNames {
		// last 2 args of list function "" because they're not required
		// https://github.com/Azure/azure-sdk-for-go/blob/main/services/sql/mgmt/2014-04-01/sql/databases.go#:~:text=func%20(client%20DatabasesClient)-,ListByServer,-(ctx%20context.Context
		// expand - expand - a comma separated list of child objects to expand in the response.
		// filter - an OData filter expression that describes a subset of databases to return.
		sqlDatabases, err := ar.ListSQLDatabases(ctx, sqlServerName, "")
		if err != nil {
			return nil, errors.Wrap(err, "unable to list SQL databases from reader")
		}
		for _, sqlDatabase := range sqlDatabases {
			if !filterByTags(filters, sqlDatabase.Tags) {
				continue
			}
			r := provider.NewResource(*sqlDatabase.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func mssqlFirewallRules(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sqlServerNames, err := getMsSQLServers(ctx, a, ar, MssqlServer.String(), filters)
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

func mssqlServerSecurityAlertPolicies(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sqlServerNames, err := getMsSQLServers(ctx, a, ar, MssqlServer.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list SQL Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, sqlServerName := range sqlServerNames {
		sqlServerSecurityAlertPolicies, err := ar.ListSQLServerSecurityAlertPolicies(ctx, sqlServerName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list SQL server security alert policies from reader")
		}
		for _, sqlServerSecurityAlertPolicy := range sqlServerSecurityAlertPolicies {
			r := provider.NewResource(*sqlServerSecurityAlertPolicy.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func mssqlServerVulnerabilityAssessments(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sqlServerNames, err := getMsSQLServers(ctx, a, ar, MssqlServer.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list SQL Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, sqlServerName := range sqlServerNames {
		sqlServerVulnerabilityAssessments, err := ar.ListSQLServerVulnerabilityAssessments(ctx, sqlServerName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list SQL server vulnerability assessments from reader")
		}
		for _, sqlServerVulnerabilityAssessment := range sqlServerVulnerabilityAssessments {
			r := provider.NewResource(*sqlServerVulnerabilityAssessment.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func mssqlVirtualMachines(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sqlVirtualMachines, err := ar.ListSQLVirtualMachines(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list SQL Virtual Machines from reader")
	}
	resources := make([]provider.Resource, 0, len(sqlVirtualMachines))
	for _, sqlVirtualMachine := range sqlVirtualMachines {
		if !filterByTags(filters, sqlVirtualMachine.Tags) {
			continue
		}
		r := provider.NewResource(*sqlVirtualMachine.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func mssqlVirtualNetworkRules(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sqlServerNames, err := getMsSQLServers(ctx, a, ar, MssqlServer.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list SQL Servers from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, sqlServerName := range sqlServerNames {
		sqlVirtualNetworkRules, err := ar.ListSQLVirtualNetworkRules(ctx, sqlServerName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list SQL virtual network rules from reader")
		}
		for _, sqlVirtualNetworkRule := range sqlVirtualNetworkRules {
			r := provider.NewResource(*sqlVirtualNetworkRule.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

// Redis

func redisCaches(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	redisCaches, err := ar.ListRedisCaches(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Redis Caches from reader")
	}
	resources := make([]provider.Resource, 0, len(redisCaches))
	for _, redisCache := range redisCaches {
		if !filterByTags(filters, redisCache.Tags) {
			continue
		}
		r := provider.NewResource(*redisCache.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *redisCache.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the Redis Cache '%s'", *redisCache.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func redisFirewallRules(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	redisCachesNames, err := getRedisCaches(ctx, a, ar, RedisCache.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Redis Caches from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, redisCachesName := range redisCachesNames {
		redisFirewallRules, err := ar.ListREDISFirewallRules(ctx, redisCachesName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list redis firewall rules from reader")
		}
		for _, redisFirewallRule := range redisFirewallRules {
			r := provider.NewResource(*redisFirewallRule.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

// DNS
func dnsZones(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	// by default maximum number of DNS zones to return is 100 zones
	dnsZones, err := ar.ListDNSZones(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list DNS Zones from reader")
	}
	resources := make([]provider.Resource, 0, len(dnsZones))
	for _, dnsZone := range dnsZones {
		if !filterByTags(filters, dnsZone.Tags) {
			continue
		}
		r := provider.NewResource(*dnsZone.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *dnsZone.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the DNS Zone '%s'", *dnsZone.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func dnsRecordSets(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	dnsZones, err := getDNSZones(ctx, a, ar, DNSZone.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list DNS Zones from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, dnsZone := range dnsZones {
		// by default maximum number of DNS records to return is 100
		// recordSetNameSuffix corresponds to the suffix label of record set name
		dnsRecordSets, err := ar.ListDNSRecordSets(ctx, dnsZone, nil, "")
		if err != nil {
			return nil, errors.Wrap(err, "unable to list DNS Record set from reader")
		}
		for _, dnsRecordSet := range dnsRecordSets {
			//adds record if the correspondent properties are set
			if resourceType == "azurerm_dns_a_record" && dnsRecordSet.RecordSetProperties.ARecords == nil {
				continue
			} else if resourceType == "azurerm_dns_aaaa_record" && dnsRecordSet.RecordSetProperties.AaaaRecords == nil {
				continue
			} else if resourceType == "azurerm_dns_caa_record" && dnsRecordSet.RecordSetProperties.CaaRecords == nil {
				continue
			} else if resourceType == "azurerm_dns_cname_record" && dnsRecordSet.RecordSetProperties.CnameRecord == nil {
				continue
			} else if resourceType == "azurerm_dns_mx_record" && dnsRecordSet.RecordSetProperties.MxRecords == nil {
				continue
			} else if resourceType == "azurerm_dns_ns_record" && dnsRecordSet.RecordSetProperties.NsRecords == nil {
				continue
			} else if resourceType == "azurerm_dns_ptr_record" && dnsRecordSet.RecordSetProperties.PtrRecords == nil {
				continue
			} else if resourceType == "azurerm_dns_srv_record" && dnsRecordSet.RecordSetProperties.SrvRecords == nil {
				continue
			} else if resourceType == "azurerm_dns_txt_record" && dnsRecordSet.RecordSetProperties.TxtRecords == nil {
				continue
			}

			r := provider.NewResource(*dnsRecordSet.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

// Private DNS
func privateDNSZones(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	// by default maximum number of DNS zones to return is 100 zones
	privateDNSZones, err := ar.ListPRIVATEDNSPrivateZones(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Private DNS Zones from reader")
	}
	resources := make([]provider.Resource, 0, len(privateDNSZones))
	for _, privateDNSZone := range privateDNSZones {
		if !filterByTags(filters, privateDNSZone.Tags) {
			continue
		}
		r := provider.NewResource(*privateDNSZone.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *privateDNSZone.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the Private DNS Zone '%s'", *privateDNSZone.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func privateDNSRecordSets(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	privateDNSZones, err := getPrivateDNSZones(ctx, a, ar, PrivateDNSZone.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Private DNS Zones from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, privateDNSZone := range privateDNSZones {
		// by default maximum number of DNS records to return is 100
		// recordSetNameSuffix corresponds to the suffix label of record set name
		privateDNSRecordSets, err := ar.ListPRIVATEDNSRecordSets(ctx, privateDNSZone, nil, "")
		if err != nil {
			return nil, errors.Wrap(err, "unable to list Private DNS Record set from reader")
		}
		for _, privateDNSRecordSet := range privateDNSRecordSets {
			//adds record if the correspondent properties are set
			if resourceType == "azurerm_dns_a_record" && privateDNSRecordSet.RecordSetProperties.ARecords == nil {
				continue
			} else if resourceType == "azurerm_dns_aaaa_record" && privateDNSRecordSet.RecordSetProperties.AaaaRecords == nil {
				continue
			} else if resourceType == "azurerm_dns_cname_record" && privateDNSRecordSet.RecordSetProperties.CnameRecord == nil {
				continue
			} else if resourceType == "azurerm_dns_mx_record" && privateDNSRecordSet.RecordSetProperties.MxRecords == nil {
				continue
			} else if resourceType == "azurerm_dns_ptr_record" && privateDNSRecordSet.RecordSetProperties.PtrRecords == nil {
				continue
			} else if resourceType == "azurerm_dns_srv_record" && privateDNSRecordSet.RecordSetProperties.SrvRecords == nil {
				continue
			} else if resourceType == "azurerm_dns_txt_record" && privateDNSRecordSet.RecordSetProperties.TxtRecords == nil {
				continue
			}

			r := provider.NewResource(*privateDNSRecordSet.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func privateDNSVirtualNetworkLinks(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	privateDNSZones, err := getPrivateDNSZones(ctx, a, ar, PrivateDNSZone.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Private DNS Zones from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, privateDNSZone := range privateDNSZones {
		// by default maximum number of DNS records to return is 100
		privateDNSVirtualNetworkLinks, err := ar.ListPRIVATEDNSVirtualNetworkLinks(ctx, privateDNSZone, nil)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list Private DNS Record set from reader")
		}
		for _, privateDNSVirtualNetworkLink := range privateDNSVirtualNetworkLinks {
			r := provider.NewResource(*privateDNSVirtualNetworkLink.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

// Policy

func policyDefinitions(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	policyDefinitions, err := ar.ListPOLICYDefinitions(ctx, "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Policy Definitions from reader")
	}
	resources := make([]provider.Resource, 0, len(policyDefinitions))
	for _, policyDefinition := range policyDefinitions {
		r := provider.NewResource(*policyDefinition.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func policyRemediations(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	policyRemediations, err := ar.ListPOLICYINSIGHTSRemediations(ctx, nil, "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Policy Remediations from reader")
	}
	resources := make([]provider.Resource, 0, len(policyRemediations))
	for _, policyRemediation := range policyRemediations {
		r := provider.NewResource(*policyRemediation.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func policySetDefinitions(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	policySetDefinitions, err := ar.ListPOLICYSetDefinitions(ctx, "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Policy Set Definitions from reader")
	}
	resources := make([]provider.Resource, 0, len(policySetDefinitions))
	for _, policySetDefinition := range policySetDefinitions {
		r := provider.NewResource(*policySetDefinition.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

// KeyVault
func keyVaults(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	keyVaults, err := ar.ListKeyVaults(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list key vault from reader")
	}
	resources := make([]provider.Resource, 0, len(keyVaults))
	for _, keyVault := range keyVaults {
		if !filterByTags(filters, keyVault.Tags) {
			continue
		}
		r := provider.NewResource(*keyVault.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func keyVaultProperties(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	keyVaults, err := ar.ListKeyVaults(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list key vault from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, keyVault := range keyVaults {
		if !filterByTags(filters, keyVault.Tags) {
			continue
		}
		if vaultProps := keyVault.Properties; vaultProps == nil {
			if resourceType == "azurerm_key_vault_access_policy" && vaultProps.AccessPolicies != nil {
				for _, vaultAcessPolicy := range *vaultProps.AccessPolicies {
					r := provider.NewResource(*vaultAcessPolicy.ObjectID, resourceType, a)
					resources = append(resources, r)
				}
			}
		}
	}
	return resources, nil
}

// Application Insigths
func applicationInsights(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	applicationInsights, err := ar.ListINSIGHTSComponents(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list application insights components from reader")
	}
	resources := make([]provider.Resource, 0, len(applicationInsights))
	for _, applicationInsight := range applicationInsights {
		if !filterByTags(filters, applicationInsight.Tags) {
			continue
		}
		r := provider.NewResource(*applicationInsight.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *applicationInsight.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the Application Insights components '%s'", *applicationInsight.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func applicationInsightsAPIKeys(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	applicationInsightsNames, err := getApplicationInsightsComponents(ctx, a, ar, ApplicationInsights.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Application Insights components from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, applicationInsightsName := range applicationInsightsNames {
		applicationInsightsAPIKeys, err := ar.ListINSIGHTSAPIKeys(ctx, applicationInsightsName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list application insigths api keys set from reader")
		}
		for _, applicationInsightsAPIKey := range applicationInsightsAPIKeys {
			r := provider.NewResource(*applicationInsightsAPIKey.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func applicationInsightsAnalyticsItems(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	applicationInsightsNames, err := getApplicationInsightsComponents(ctx, a, ar, ApplicationInsights.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Application Insights components from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, applicationInsightsName := range applicationInsightsNames {
		applicationInsightsAnalyticsItems, err := ar.ListINSIGHTSAnalyticsItems(ctx, applicationInsightsName, "", "", "", nil)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list application insigths api keys set from reader")
		}
		for _, applicationInsightsAnalyticsItem := range applicationInsightsAnalyticsItems {
			r := provider.NewResource(*applicationInsightsAnalyticsItem.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

// issue import Error = 'json: cannot unmarshal array into Go value of type insights.WebTestListResult' JSON
// follow-up at https://github.com/Azure/azure-rest-api-specs/issues/9463
func applicationInsightsWebTests(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	insightsWebTests, err := ar.ListINSIGHTSWebTests(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list application insights web tests from reader")
	}
	resources := make([]provider.Resource, 0, len(insightsWebTests))
	for _, insightsWebTest := range insightsWebTests {
		if !filterByTags(filters, insightsWebTest.Tags) {
			continue
		}
		r := provider.NewResource(*insightsWebTest.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

// Log Analytics
func logAnalyticsWorkspaces(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	workspaces, err := ar.ListLogAnalyticsWorkspaces(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list log analytics workspaces from reader")
	}
	resources := make([]provider.Resource, 0, len(workspaces))
	for _, workspace := range workspaces {
		if !filterByTags(filters, workspace.Tags) {
			continue
		}
		r := provider.NewResource(*workspace.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *workspace.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the log analytics workspace '%s'", *workspace.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func logAnalyticsLinkedServices(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	workspaceNames, err := getLogAnalyticsWorkspaces(ctx, a, ar, LogAnalyticsWorkspace.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Application Insights components from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, workspaceName := range workspaceNames {
		linkedServices, err := ar.ListLogAnalyticsLinkedService(ctx, workspaceName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list log analytics linked services set from reader")
		}
		for _, linkedService := range linkedServices {
			if !filterByTags(filters, linkedService.Tags) {
				continue
			}
			r := provider.NewResource(*linkedService.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func logAnalyticsDatasources(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	workspaceNames, err := getLogAnalyticsWorkspaces(ctx, a, ar, LogAnalyticsWorkspace.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Application Insights components from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, workspaceName := range workspaceNames {
		datasources, err := ar.ListLogAnalyticsDatasource(ctx, workspaceName, "", "")
		if err != nil {
			return nil, errors.Wrap(err, "unable to list log analytics datasources set from reader")
		}
		for _, datasource := range datasources {
			if !filterByTags(filters, datasource.Tags) {
				continue
			}
			if resourceType == "azurerm_log_analytics_datasource_windows_performance_counter" && datasource.Kind != "WindowsPerformanceCounter" {
				continue
			} else if resourceType == "azurerm_log_analytics_datasource_windows_event" && datasource.Kind != "WindowsEvent" {
				continue
			}
			r := provider.NewResource(*datasource.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

// Monitor
func monitorActionGroups(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	actionGroups, err := ar.ListMonitorActionsGroup(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list monitor action groups from reader")
	}
	resources := make([]provider.Resource, 0, len(actionGroups))
	for _, actionGroup := range actionGroups {
		if !filterByTags(filters, actionGroup.Tags) {
			continue
		}
		r := provider.NewResource(*actionGroup.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func monitorActivityLogAlerts(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	activityLogAlerts, err := ar.ListMonitorActivityLogAlert(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list monitor activity log alert from reader")
	}
	resources := make([]provider.Resource, 0, len(activityLogAlerts))
	for _, activityLogAlert := range activityLogAlerts {
		if !filterByTags(filters, activityLogAlert.Tags) {
			continue
		}
		r := provider.NewResource(*activityLogAlert.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func monitorAutoscaleSettings(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	autoscaleSettings, err := ar.ListMonitorAutoScaleSettings(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list monitor autoscale settings from reader")
	}
	resources := make([]provider.Resource, 0, len(autoscaleSettings))
	for _, autoscaleSetting := range autoscaleSettings {
		if !filterByTags(filters, autoscaleSetting.Tags) {
			continue
		}
		r := provider.NewResource(*autoscaleSetting.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func monitorLogProfiles(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	logProfiles, err := ar.ListMonitorLogProfiles(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list monitor log profile from reader")
	}
	resources := make([]provider.Resource, 0, len(logProfiles))
	for _, logProfile := range logProfiles {
		if !filterByTags(filters, logProfile.Tags) {
			continue
		}
		r := provider.NewResource(*logProfile.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

func monitorMetricAlerts(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	metricsAlerts, err := ar.ListMonitorMetricsAlerts(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list monitor metrics alerts from reader")
	}
	resources := make([]provider.Resource, 0, len(metricsAlerts))
	for _, metricsAlert := range metricsAlerts {
		if !filterByTags(filters, metricsAlert.Tags) {
			continue
		}
		r := provider.NewResource(*metricsAlert.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}

// App service
func webApps(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	webApps, err := ar.ListWebApps(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list app service web apps from reader")
	}
	resources := make([]provider.Resource, 0, len(webApps))
	for _, webApp := range webApps {
		if !filterByTags(filters, webApp.Tags) {
			continue
		}

		// https://azure.github.io/AppService/2021/08/31/Kind-property-overview.html
		// is linux if reserved is set
		if resourceType == "azurerm_windows_web_app" && *webApp.SiteProperties.Reserved == false || resourceType == "azurerm_linux_web_app" && *webApp.SiteProperties.Reserved == true {
			r := provider.NewResource(*webApp.ID, resourceType, a)
			// we set the name prior of reading it from the state
			// as it is required to able to List resources depending on this one
			if err := r.Data().Set("name", *webApp.Name); err != nil {
				return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the app service web app '%s'", *webApp.Name)
			}
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func servicePlans(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	// specify detailed = true to return all App Service plan properties, defaults to false, which returns a subset of the properties.
	// Note! Retrieval of all properties may increase the API latency.
	detailed := false
	servicePlans, err := ar.ListAppServicePlans(ctx, &detailed)

	if err != nil {
		return nil, errors.Wrap(err, "unable to list app service service plans from reader")
	}
	resources := make([]provider.Resource, 0, len(servicePlans))
	for _, servicePlan := range servicePlans {
		if !filterByTags(filters, servicePlan.Tags) {
			continue
		}

		r := provider.NewResource(*servicePlan.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *servicePlan.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the app service plan '%s'", *servicePlan.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func sourceControlTokens(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sourceControls, err := ar.ListSourceControls(ctx)

	if err != nil {
		return nil, errors.Wrap(err, "unable to list app service source controls from reader")
	}
	resources := make([]provider.Resource, 0, len(sourceControls))
	for _, sourceControl := range sourceControls {
		if sourceControl.SourceControlProperties != nil && sourceControl.ID != nil {
			//fmt.Sprintf("/providers/Microsoft.Web/sourcecontrols/%s", *sourceControl.Type)
			r := provider.NewResource(*sourceControl.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func staticSites(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	staticSites, err := ar.ListStaticSites(ctx)

	if err != nil {
		return nil, errors.Wrap(err, "unable to list app service static sites from reader")
	}
	resources := make([]provider.Resource, 0, len(staticSites))
	for _, staticSite := range staticSites {
		if !filterByTags(filters, staticSite.Tags) {
			continue
		}
		r := provider.NewResource(*staticSite.ID, resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *staticSite.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the app service static site '%s'", *staticSite.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func staticSiteCustomDomains(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	staticSitesNames, err := getStaticSites(ctx, a, ar, StaticSite.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list app service static sites from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, staticSitesName := range staticSitesNames {
		staticSiteCustomDomains, err := ar.ListStaticSitesCustomDomain(ctx, staticSitesName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list app service static sites custom domains set from reader")
		}
		for _, staticSiteCustomDomain := range staticSiteCustomDomains {
			r := provider.NewResource(*staticSiteCustomDomain.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func webAppHybridConnections(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	servicePlansNames, err := getServicePlans(ctx, a, ar, ServicePlan.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list app service plan from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, servicePlanName := range servicePlansNames {
		hybridConnections, err := ar.ListHybridConnections(ctx, servicePlanName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list app service web app hybrid connections set from reader")
		}
		for _, hybridConnection := range hybridConnections {
			r := provider.NewResource(*hybridConnection.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func linuxWebAppSlots(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	webAppsNames, err := getWebApps(ctx, a, ar, []string{LinuxWebApp.String()}, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list app service linux web apps from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, webAppsName := range webAppsNames {
		deploymentSlots, err := ar.ListDeploymentSlots(ctx, webAppsName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list app service linux web app deployment slots set from reader")
		}
		for _, deploymentSlot := range deploymentSlots {
			if !filterByTags(filters, deploymentSlot.Tags) {
				continue
			}
			r := provider.NewResource(*deploymentSlot.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func windowsWebAppSlots(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	webAppsNames, err := getWebApps(ctx, a, ar, []string{WindowsWebApp.String()}, filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list app service windows web apps from cache")
	}
	resources := make([]provider.Resource, 0)
	for _, webAppsName := range webAppsNames {
		deploymentSlots, err := ar.ListDeploymentSlots(ctx, webAppsName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to list app service web windows app deployment slots set from reader")
		}
		for _, deploymentSlot := range deploymentSlots {
			if !filterByTags(filters, deploymentSlot.Tags) {
				continue
			}
			r := provider.NewResource(*deploymentSlot.ID, resourceType, a)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func webAppActiveSlots(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	webApps, err := ar.ListWebApps(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list app service web apps from reader")
	}
	resources := make([]provider.Resource, 0, len(webApps))
	for _, webApp := range webApps {
		if !filterByTags(filters, webApp.Tags) {
			continue
		}

		if webApp.SiteProperties != nil && webApp.SiteProperties.SlotSwapStatus != nil {

			if webApp.SiteProperties.SlotSwapStatus.SourceSlotName != nil {
				r := provider.NewResource(*webApp.ID, resourceType, a)

				resources = append(resources, r)
			}
		}
	}
	return resources, nil
}

// dataprotection
func dataProtectionBackupVaults(ctx context.Context, a *azurerm, ar *AzureReader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	backupVaults, err := ar.ListBackupVaultResources(ctx)

	if err != nil {
		return nil, errors.Wrap(err, "unable to list backup vaults from reader")
	}
	resources := make([]provider.Resource, 0, len(backupVaults))
	for _, backupVault := range backupVaults {
		if !filterByTags(filters, backupVault.Tags) {
			continue
		}

		// TODO: recheck with upgrade SDK if still need to change the string to avoid error on import
		r := provider.NewResource(strings.ReplaceAll(*backupVault.ID, "BackupVault", "backupVault"), resourceType, a)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", *backupVault.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the app service static site '%s'", *backupVault.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}
