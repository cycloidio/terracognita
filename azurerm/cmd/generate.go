package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

var azureAPIs = []AzureAPI{
	{API: "compute", APIVersion: "2021-12-01"},
	{API: "network", APIVersion: "2021-05-01"},
	{API: "desktopvirtualization", APIVersion: "2021-09-03-preview", IsPreview: true},
	{API: "logic", APIVersion: "2019-05-01"},
	{API: "containerregistry", APIVersion: "2019-05-01"},
	{API: "containerservice", APIVersion: "2022-01-01"},
	{API: "storage", APIVersion: "2021-08-01", AddAPISufix: true},
	{API: "mariadb", APIVersion: "2020-01-01", AddAPISufix: true},
	{API: "mysql", APIVersion: "2020-01-01", AddAPISufix: true},
	{API: "postgresql", APIVersion: "2020-01-01", AddAPISufix: true},
	{API: "sql", APIVersion: "v5.0", AddAPISufix: true, IsPreview: true},          // used for mssql resources
	{API: "sqlvirtualmachine", APIVersion: "2017-03-01-preview", IsPreview: true}, // used for mssql resources
	{API: "redis", APIVersion: "2020-12-01", AddAPISufix: true},
	{API: "dns", APIVersion: "2018-05-01", AddAPISufix: true},
	{API: "privatedns", APIVersion: "2018-09-01", AddAPISufix: true},
	{API: "policy", OtherPath: "resources/mgmt", APIVersion: "2021-06-01-preview", AddAPISufix: true, IsPreview: true},
	{API: "policyinsights", APIVersion: "2020-07-01-preview", AddAPISufix: true, IsPreview: true},
	{API: "keyvault", APIVersion: "2020-04-01-preview", IsPreview: true},                                                                       // used for keyvault resources
	{API: "insights", OtherPath: "appinsights/mgmt", APIVersion: "2020-02-02", AddAPISufix: true},                                              // used for  app insights resources
	{API: "operationalinsights", APIVersion: "2020-08-01"},                                                                                     // used for log analytics resources
	{PackageIdentifier: "newActionGroupClient", API: "insights", OtherPath: "monitor/mgmt", APIVersion: "2021-09-01-preview", IsPreview: true}, // used for monitor resources
	{PackageIdentifier: "newActivityLogAlertsClient", API: "insights", OtherPath: "monitor/mgmt", APIVersion: "2020-10-01"},                    // used for monitor resources
	{PackageIdentifier: "monitor", API: "insights", OtherPath: "monitor/mgmt", APIVersion: "2021-07-01-preview", IsPreview: true},              // used for monitor resources
}

var functions = []Function{
	// Compute API Resources
	{ResourceName: "VirtualMachine", API: "compute", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "filter",
			Type: "string",
		},
	}},
	{ResourceName: "VirtualMachineScaleSet", API: "compute", ResourceGroup: true},
	{ResourceName: "VirtualMachineScaleSetExtension", API: "compute", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "VMScaleSetName",
			Type: "string",
		},
	}},
	{ResourceName: "VirtualMachineExtension", API: "compute", ResourceGroup: true, ReturnsList: true, ExtraArgs: []Arg{
		{
			Name: "VMName",
			Type: "string",
		},
		{
			Name: "expand",
			Type: "string",
		},
	}},
	{ResourceName: "AvailabilitySet", API: "compute", ResourceGroup: true},
	{ResourceName: "Image", API: "compute", ResourceGroup: false},
	{ResourceName: "Disk", AzureSDKListFunction: "ListByResourceGroup", API: "compute", ResourceGroup: true},
	// Network API Resources
	{ResourceName: "VirtualNetwork", API: "network", ResourceGroup: true},
	{ResourceName: "Subnet", API: "network", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "virtualNetworkName",
			Type: "string",
		},
	}},
	{ResourceName: "Interface", API: "network", ResourceGroup: true},
	{ResourceName: "SecurityGroup", API: "network", ResourceGroup: true},
	{ResourceName: "ApplicationGateway", API: "network", ResourceGroup: true},
	{ResourceName: "ApplicationSecurityGroup", API: "network", ResourceGroup: true},
	{ResourceName: "DdosProtectionPlan", API: "network", ResourceGroup: false},
	{ResourceName: "AzureFirewall", API: "network", ResourceGroup: true},
	{ResourceName: "LocalNetworkGateway", API: "network", ResourceGroup: true},
	{ResourceName: "NatGateway", API: "network", ResourceGroup: true},
	{ResourceName: "Profile", API: "network", ResourceGroup: true},
	{ResourceName: "SecurityRule", API: "network", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "networkSecurityGroupName",
			Type: "string",
		},
	}},
	{ResourceName: "PublicIPAddress", API: "network", ResourceGroup: true},
	{ResourceName: "PublicIPPrefix", API: "network", ResourceGroup: true},
	{ResourceName: "Route", API: "network", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "routeTableName",
			Type: "string",
		},
	}},
	{ResourceName: "RouteTable", API: "network", ResourceGroup: true},
	{ResourceName: "VirtualNetworkGateway", API: "network", ResourceGroup: true},
	{ResourceName: "VirtualNetworkGatewayConnection", API: "network", ResourceGroup: true},
	{ResourceName: "VirtualNetworkPeering", API: "network", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "virtualNetworkName",
			Type: "string",
		},
	}},
	{ResourceName: "WebApplicationFirewallPolicy", API: "network", ResourceGroup: true},
	{ResourceName: "VirtualHub", API: "network", ResourceGroup: false},
	{ResourceName: "BgpConnection", PluralName: "VirtualHubBgpConnections", API: "network", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "virtualHubName",
			Type: "string",
		},
	}},
	{ResourceName: "HubVirtualNetworkConnection", API: "network", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "virtualHubName",
			Type: "string",
		},
	}},
	{ResourceName: "HubIPConfiguration", PluralName: "VirtualHubIPConfiguration", API: "network", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "virtualHubName",
			Type: "string",
		},
	}},
	{ResourceName: "HubRouteTable", API: "network", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "virtualHubName",
			Type: "string",
		},
	}},
	{ResourceName: "SecurityPartnerProvider", API: "network", ResourceGroup: false},
	{ResourceName: "LoadBalancer", API: "network", ResourceGroup: true},
	{ResourceName: "BackendAddressPool", PluralName: "LoadBalancerBackendAddressPools", API: "network", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "loadBalancerName",
			Type: "string",
		},
	}},
	// Desktop API Resources
	{ResourceName: "HostPool", AzureSDKListFunction: "ListByResourceGroup", API: "desktopvirtualization", ResourceGroup: true},
	{ResourceName: "ApplicationGroup", AzureSDKListFunction: "ListByResourceGroup", API: "desktopvirtualization", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "filter",
			Type: "string",
		},
	}},
	// Logic API Resources
	{ResourceName: "Workflow", AzureSDKListFunction: "ListByResourceGroup", API: "logic", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "top",
			Type: "*int32",
		},
		{
			Name: "filter",
			Type: "string",
		},
	}},
	{ResourceName: "WorkflowTrigger", API: "logic", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "workflowName",
			Type: "string",
		},
		{
			Name: "top",
			Type: "*int32",
		},
		{
			Name: "filter",
			Type: "string",
		},
	}},
	{ResourceName: "WorkflowRun", API: "logic", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "workflowName",
			Type: "string",
		},
		{
			Name: "top",
			Type: "*int32",
		},
		{
			Name: "filter",
			Type: "string",
		},
	}},
	{ResourceName: "WorkflowRunAction", API: "logic", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "workflowName",
			Type: "string",
		},
		{
			Name: "runName",
			Type: "string",
		},
		{
			Name: "top",
			Type: "*int32",
		},
		{
			Name: "filter",
			Type: "string",
		},
	}},
	// Container Registry Resources
	{ResourceName: "Registry", API: "containerregistry", FunctionName: "ListContainerRegistries", ResourceGroup: false},
	{ResourceName: "Webhook", API: "containerregistry", FunctionName: "ListContainerRegistryWebhooks", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "registryName",
			Type: "string",
		},
	}},
	// Container Service Resources - k8s services
	{ResourceName: "ManagedCluster", API: "containerservice", FunctionName: "ListKubernetesClusters", ResourceGroup: false},
	{ResourceName: "AgentPool", API: "containerservice", FunctionName: "ListKubernetesClusterNodes", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "managedClusterName",
			Type: "string",
		},
	}},

	// Storage Resources
	{ResourceName: "Account", API: "storage", ResourceGroup: false},
	{ResourceName: "ListContainerItem", PluralName: "BlobContainers", API: "storage", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "accountName",
			Type: "string",
		},
		{
			Name: "maxpagesize",
			Type: "string",
		},
		{
			Name: "filter",
			Type: "string",
		},
		{
			Name: "include",
			Type: "storage.ListContainersInclude",
		},
	}},
	{ResourceName: "ListQueue", API: "storage", PluralName: "Queue", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "accountName",
			Type: "string",
		},
		{
			Name: "maxpagesize",
			Type: "string",
		},
		{
			Name: "filter",
			Type: "string",
		},
	}},
	{ResourceName: "FileShareItem", PluralName: "FileShares", API: "storage", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "accountName",
			Type: "string",
		},
		{
			Name: "maxpagesize",
			Type: "string",
		},
		{
			Name: "filter",
			Type: "string",
		},
		{
			Name: "expand",
			Type: "string",
		},
	}},
	{ResourceName: "Table", API: "storage", PluralName: "Table", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "accountName",
			Type: "string",
		},
	}},
	// Database Resources
	// mariadb
	{ResourceName: "Configuration", API: "mariadb", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ReturnsList: true, ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
	}},
	{ResourceName: "Database", API: "mariadb", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ReturnsList: true, ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
	}},
	{ResourceName: "FirewallRule", API: "mariadb", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ReturnsList: true, ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
	}},
	{ResourceName: "Server", API: "mariadb", ReturnsList: true, ResourceGroup: false},
	{ResourceName: "VirtualNetworkRule", API: "mariadb", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
	}},
	// mysql
	{ResourceName: "Configuration", API: "mysql", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ReturnsList: true, ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
	}},
	{ResourceName: "Database", API: "mysql", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ReturnsList: true, ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
	}},
	{ResourceName: "FirewallRule", API: "mysql", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ReturnsList: true, ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
	}},
	{ResourceName: "Server", API: "mysql", ReturnsList: true, ResourceGroup: false},
	{ResourceName: "VirtualNetworkRule", API: "mysql", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
	}},
	// postgresql
	{ResourceName: "Configuration", API: "postgresql", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ReturnsList: true, ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
	}},
	{ResourceName: "Database", API: "postgresql", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ReturnsList: true, ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
	}},
	{ResourceName: "FirewallRule", API: "postgresql", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ReturnsList: true, ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
	}},
	{ResourceName: "Server", API: "postgresql", ReturnsList: true, ResourceGroup: false},
	{ResourceName: "VirtualNetworkRule", API: "postgresql", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
	}},
	// mssql
	{ResourceName: "Database", API: "sql", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
		{
			Name: "skipToken",
			Type: "string",
		},
	}},
	{ResourceName: "ElasticPool", API: "sql", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
		{
			Name: "skip",
			Type: "*int32",
		},
	}},
	{ResourceName: "FirewallRule", API: "sql", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
	}},
	{ResourceName: "Server", API: "sql", ResourceGroup: false, ExtraArgs: []Arg{
		{
			Name: "expand",
			Type: "string",
		},
	}},
	{ResourceName: "ServerSecurityAlertPolicy", API: "sql", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
	}},
	{ResourceName: "ServerVulnerabilityAssessment", API: "sql", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
	}},
	{ResourceName: "VirtualNetworkRule", API: "sql", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
	}},
	{ResourceName: "SQLVirtualMachine", API: "sqlvirtualmachine"},
	//redis
	//Corresponds to redis cache resource
	{ResourceName: "ResourceType", API: "redis", FunctionName: "ListRedisCaches", PluralName: "RedisCaches", IrregularClientName: "NewClient", ResourceGroup: true, AzureSDKListFunction: "ListByResourceGroup"},
	{ResourceName: "FirewallRule", API: "redis", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "cacheName",
			Type: "string",
		},
	}},
	// dns
	{ResourceName: "Zone", API: "dns", ResourceGroup: false, ExtraArgs: []Arg{
		{
			Name: "top",
			Type: "*int32",
		},
	}},

	{ResourceName: "RecordSet", API: "dns", AzureSDKListFunction: "ListAllByDNSZone", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "zoneName",
			Type: "string",
		},
		{
			Name: "top",
			Type: "*int32",
		},
		{
			Name: "recordSetNameSuffix",
			Type: "string",
		},
	}},
	// privatedns
	{ResourceName: "PrivateZone", API: "privatedns", ResourceGroup: false, ExtraArgs: []Arg{
		{
			Name: "top",
			Type: "*int32",
		},
	}},
	{ResourceName: "RecordSet", API: "privatedns", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "zoneName",
			Type: "string",
		},
		{
			Name: "top",
			Type: "*int32",
		},
		{
			Name: "recordSetNameSuffix",
			Type: "string",
		},
	}},
	{ResourceName: "VirtualNetworkLink", API: "privatedns", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "privateZoneName",
			Type: "string",
		},
		{
			Name: "top",
			Type: "*int32",
		},
	}},
	// Policy
	{ResourceName: "Definition", API: "policy", ResourceGroup: false, ExtraArgs: []Arg{
		{
			Name: "filter",
			Type: "string",
		},
		{
			Name: "top",
			Type: "*int32",
		},
	}},
	{ResourceName: "SetDefinition", API: "policy", ResourceGroup: false, ExtraArgs: []Arg{
		{
			Name: "filter",
			Type: "string",
		},
		{
			Name: "top",
			Type: "*int32",
		},
	}},
	{ResourceName: "Remediation", API: "policyinsights", AzureSDKListFunction: "ListForResourceGroup", ResourceGroup: true, Subscription: true, ExtraArgs: []Arg{
		{
			Name: "top",
			Type: "*int32",
		},
		{
			Name: "filter",
			Type: "string",
		},
	}},
	// key vault
	{ResourceName: "Vault", FunctionName: "ListKeyVaults", ResourceGroup: true, AzureSDKListFunction: "ListByResourceGroup", API: "keyvault", ExtraArgs: []Arg{
		{
			Name: "top",
			Type: "*int32",
		},
	}},
	// app insights
	{ResourceName: "ApplicationInsightsComponent", PluralName: "Components", API: "insights", ResourceGroup: false},
	{ResourceName: "ApplicationInsightsComponentAPIKey", PluralName: "APIKeys", API: "insights", ResourceGroup: true, ReturnsList: true, ExtraArgs: []Arg{
		{
			Name: "ApplicationInsightsComponent",
			Type: "string",
		},
	}},
	{ResourceName: "ApplicationInsightsComponentAnalyticsItem", PluralName: "AnalyticsItems", API: "insights", ResourceGroup: true, ReturnsList: true, ExtraArgs: []Arg{
		{
			Name: "ApplicationInsightsComponent",
			Type: "string",
		},
		{
			Name: "scopePath",
			Type: "insights.ItemScopePath",
		},
		{
			Name: "scope",
			Type: "insights.ItemScope",
		},
		{
			Name: "typeParameter",
			Type: "insights.ItemTypeParameter",
		},
		{
			Name: "includeContent",
			Type: "*bool",
		},
	}},
	//issue at https://github.com/Azure/azure-rest-api-specs/issues/9463
	{ResourceName: "WebTest", API: "insights", ResourceGroup: false},
	// log analytics
	{ResourceName: "Workspace", API: "operationalinsights", FunctionName: "ListLogAnalyticsWorkspaces", ReturnsList: true, ResourceGroup: false},
	{ResourceName: "LinkedService", API: "operationalinsights", FunctionName: "ListLogAnalyticsLinkedService", AzureSDKListFunction: "ListByWorkspace", ReturnsList: true, ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "workspaceName",
			Type: "string",
		},
	}},
	{ResourceName: "DataSource", API: "operationalinsights", FunctionName: "ListLogAnalyticsDatasource", AzureSDKListFunction: "ListByWorkspace", ResourceGroup: true, ExtraArgs: []Arg{
		{
			Name: "workspaceName",
			Type: "string",
		},
		{
			Name: "filter",
			Type: "string",
		},
		{
			Name: "skiptoken",
			Type: "string",
		},
	}},
	//monitor
	{ResourceName: "ActionGroupResource", API: "newActionGroupClient", IrregularClientName: "NewActionGroupsClient", FunctionName: "ListMonitorActionsGroup", AzureSDKListFunction: "ListByResourceGroup", ReturnsList: true, ResourceGroup: true},
	{ResourceName: "ActivityLogAlertResource", API: "newActivityLogAlertsClient", IrregularClientName: "NewActivityLogAlertsClient", FunctionName: "ListMonitorActivityLogAlert", AzureSDKListFunction: "ListByResourceGroup", ResourceGroup: true},
	{ResourceName: "AutoscaleSettingResource", API: "monitor", IrregularClientName: "NewAutoscaleSettingsClient", FunctionName: "ListMonitorAutoScaleSettings", AzureSDKListFunction: "ListByResourceGroup", ResourceGroup: true},
	{ResourceName: "LogProfileResource", API: "monitor", IrregularClientName: "NewLogProfilesClient", FunctionName: "ListMonitorLogProfiles", ReturnsList: true},
	{ResourceName: "MetricAlertResource", API: "monitor", IrregularClientName: "NewMetricAlertsClient", FunctionName: "ListMonitorMetricsAlerts", ReturnsList: true, AzureSDKListFunction: "ListByResourceGroup", ResourceGroup: true},
}

func main() {
	f, err := os.OpenFile("./reader_generated.go", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := generate(f, azureAPIs, functions); err != nil {
		panic(err)
	}
}

func generate(opt io.Writer, azureAPIs []AzureAPI, fns []Function) error {
	var fnBuff = bytes.Buffer{}

	if err := pkgTmpl.Execute(&fnBuff, struct{ AzureAPIs []AzureAPI }{AzureAPIs: azureAPIs}); err != nil {
		return errors.Wrap(err, "unable to execute package template")
	}

	for _, function := range fns {
		if err := function.Execute(&fnBuff); err != nil {
			return errors.Wrapf(err, "unable to execute function template for: %s", function.ResourceName)
		}
	}

	// format
	cmd := exec.Command("goimports")
	cmd.Stdin = &fnBuff
	cmd.Stdout = opt
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "unable to run goimports command")
	}
	return nil
}
