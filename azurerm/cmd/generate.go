package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

var azureAPIs = []AzureAPI{
	{API: "compute", APIVersion: "2019-07-01"},
	{API: "network", APIVersion: "2019-06-01"},
	{API: "desktopvirtualization", APIVersion: "2019-12-10", IsPreview: true},
	{API: "logic", APIVersion: "2019-05-01"},
	{API: "containerregistry", APIVersion: "2019-05-01"},
	{API: "storage", APIVersion: "2021-02-01", AddAPISufix: true},
	{API: "mariadb", APIVersion: "2020-01-01", AddAPISufix: true},
	{API: "mysql", APIVersion: "2020-01-01", AddAPISufix: true},
	{API: "postgresql", APIVersion: "2020-01-01", AddAPISufix: true},
	{API: "sql", APIVersion: "2014-04-01", AddAPISufix: true},
}

var functions = []Function{
	// Compute API Resources
	{ResourceName: "VirtualMachine", API: "compute", ResourceGroup: true},
	{ResourceName: "VirtualMachineScaleSet", API: "compute", ResourceGroup: true},
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
			Type: "storage.ListSharesExpand",
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
	// sql
	{ResourceName: "Database", API: "sql", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ReturnsList: true, ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
		{
			Name: "expand",
			Type: "string",
		},
		{
			Name: "filter",
			Type: "string",
		},
	}},
	{ResourceName: "ElasticPool", API: "sql", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ReturnsList: true, ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
	}},
	{ResourceName: "FirewallRule", API: "sql", ResourceGroup: true, AzureSDKListFunction: "ListByServer", ReturnsList: true, ExtraArgs: []Arg{
		{
			Name: "serverName",
			Type: "string",
		},
	}},
	{ResourceName: "Server", API: "sql", ReturnsList: true, ResourceGroup: false},
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
