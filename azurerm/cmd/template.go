package main

import (
	"io"
	"text/template"

	"github.com/pkg/errors"
)

const (
	// packageTmpl it's the package definition
	packageTmpl = `
	package azurerm
	// Code generated by 'go generate'; DO NOT EDIT
	import (
		"context"

		"github.com/pkg/errors"

		{{ range .AzureAPIs -}}
		"github.com/Azure/azure-sdk-for-go/services/{{ if .IsPreview }}preview/{{ end }}{{ .API }}/mgmt/{{ .APIVersion }}{{ if .IsPreview }}-preview{{ end }}/{{ .API }}"
		{{ end }}
	)
	`

	// functionTmpl it's the implementation of a reader function
	functionTmpl = `
	// List{{ .Name }} returns a list of {{ .Name }} within a subscription {{ if .Location }}and a location {{ end }}{{ if .ResourceGroup }}and a resource group {{ end }}
	func (ar *AzureReader) List{{ .Name }}(ctx context.Context{{ range .ExtraArgs }},{{ .Name }} {{ .Type }} {{ end }}) ([]{{ .API }}.{{ .Resource }}, error) {
		client := {{ .API }}.New{{ .Resource }}sClient(ar.config.SubscriptionID)
		client.Authorizer = ar.authorizer

		output, err := client.{{ .ListFunction }}(ctx{{ if .Location }}, ar.GetLocation(){{ end }}{{ if .ResourceGroup }}, ar.GetResourceGroupName(){{ end }}{{ range .ExtraArgs }},{{ .Name }}{{ end }})
		if err != nil {
			return nil, errors.Wrap(err, "unable to list {{ .API }}.{{ .Resource }} from Azure APIs")
		}
		resources := make([]{{ .API }}.{{ .Resource }}, 0)
		for output.NotDone() {
			{{ if .Iterator }}
			res = output.Value()
			resources = append(resources, res)
			{{ else }}
			for _, res := range output.Values() {
				resources = append(resources, res)
			}
			{{ end }}
			if err := output.NextWithContext(ctx); err != nil {
				break
			}
		}
		return resources, nil
	}
	`
)

var (
	fnTmpl  *template.Template
	pkgTmpl *template.Template
)

func init() {
	var err error
	fnTmpl, err = template.New("template").Parse(functionTmpl)
	if err != nil {
		panic(err)
	}
	pkgTmpl, err = template.New("template").Parse(packageTmpl)
	if err != nil {
		panic(err)
	}
}

// Arg can be used to define
// extra args to pass to the generated
// functions
type Arg struct {
	// Name of the arg
	Name string
	// Type of the arg
	Type string
}

// AzureAPI is the definition of one of the Azure APIs
type AzureAPI struct {
	// API is used to determine the
	// Azure API to use
	// ex: compute, network
	API string

	// APIVersion is used to determine the
	// Azure API Version to use
	// ex: 2019-07-01
	APIVersion string

	// IsPreview defines if the API is a preview
	// functionality
	// https://docs.microsoft.com/en-us/azure/search/search-api-preview
	IsPreview bool
}

// Function is the definition of one of the functions
type Function struct {
	// Resource is the Azure name of the entity, like
	// VirtualMachine, VirtualNetwork, etc.
	Resource string

	// Name is the function name to be generated
	// it can be useful if you `Resource` is `SslCertificate`, which is not `go`
	// compliant, `Name` will be `SSLCertificate`, your Function name will be
	// `ListSSLCertificates`
	Name string

	// API is used to determine the
	// Azure API to use
	// ex: compute, network
	API string

	// Location is used to determine whether the resource should be filtered by Azure locations or not
	Location bool

	// ResourceGroup is used to determine whether the resource should be filtered by Azure Resource Group or not
	ResourceGroup bool

	// ListFunction is the Azure SKP list function to use
	ListFunction string

	// Iterator should be true is the ListFunction returns an Iterator and not a Page
	Iterator bool

	// ExtraArgs should be specified if extra arguments are required for the list function
	ExtraArgs []Arg
}

// Execute uses the fnTmpl to interpolate f
// and write the result to w
func (f Function) Execute(w io.Writer) error {
	if len(f.Name) == 0 {
		f.Name = f.Resource + "s"
	}
	if len(f.ListFunction) == 0 {
		f.ListFunction = "List"
	}
	if err := fnTmpl.Execute(w, f); err != nil {
		return errors.Wrapf(err, "failed to Execute with Function %+v", f)
	}
	return nil
}
