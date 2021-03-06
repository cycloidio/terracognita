package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

var azureAPIs = []AzureAPI{
	AzureAPI{API: "compute", APIVersion: "2019-07-01"},
	AzureAPI{API: "network", APIVersion: "2019-06-01"},
	AzureAPI{API: "desktopvirtualization", APIVersion: "2019-12-10", IsPreview: true},
	AzureAPI{API: "logic", APIVersion: "2019-05-01"},
}

var functions = []Function{
	Function{Resource: "VirtualMachine", API: "compute", ResourceGroup: true},
	Function{Resource: "VirtualNetwork", API: "network", ResourceGroup: true},
	Function{Resource: "Subnet", API: "network", ResourceGroup: true, ExtraArgs: []Arg{
		Arg{
			Name: "virtualNetworkName",
			Type: "string",
		},
	}},
	Function{Resource: "Interface", API: "network", ResourceGroup: true},
	Function{Resource: "SecurityGroup", API: "network", ResourceGroup: true},
	Function{Resource: "VirtualMachineScaleSet", API: "compute", ResourceGroup: true},
	Function{Resource: "HostPool", ListFunction: "ListByResourceGroup", API: "desktopvirtualization", ResourceGroup: true},
	Function{Resource: "ApplicationGroup", ListFunction: "ListByResourceGroup", API: "desktopvirtualization", ResourceGroup: true, ExtraArgs: []Arg{
		Arg{
			Name: "filter",
			Type: "string",
		},
	}},
	Function{Resource: "Workflow", ListFunction: "ListByResourceGroup", API: "logic", ResourceGroup: true, ExtraArgs: []Arg{
		Arg{
			Name: "top",
			Type: "*int32",
		},
		Arg{
			Name: "filter",
			Type: "string",
		},
	}},
	Function{Resource: "WorkflowTrigger", API: "logic", ResourceGroup: true, ExtraArgs: []Arg{
		Arg{
			Name: "workflowName",
			Type: "string",
		},
		Arg{
			Name: "top",
			Type: "*int32",
		},
		Arg{
			Name: "filter",
			Type: "string",
		},
	}},
	Function{Resource: "WorkflowRun", API: "logic", ResourceGroup: true, ExtraArgs: []Arg{
		Arg{
			Name: "workflowName",
			Type: "string",
		},
		Arg{
			Name: "top",
			Type: "*int32",
		},
		Arg{
			Name: "filter",
			Type: "string",
		},
	}},
	Function{Resource: "WorkflowRunAction", API: "logic", ResourceGroup: true, ExtraArgs: []Arg{
		Arg{
			Name: "workflowName",
			Type: "string",
		},
		Arg{
			Name: "runName",
			Type: "string",
		},
		Arg{
			Name: "top",
			Type: "*int32",
		},
		Arg{
			Name: "filter",
			Type: "string",
		},
	}},
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
			return errors.Wrapf(err, "unable to execute function template for: %s", function.Resource)
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
