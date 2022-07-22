# Contributing Guidelines

Cycloid Team is glad to see you contributing to this project ! In this document, we will provide you some guidelines in order to help get your contribution accepted.

## Reporting an issue

### Issues

When you find a bug in Terracognita, it should be reported using [GitHub issues](https://github.com/cycloidio/terracognita/issues). Please define key information like your Operating System (OS), Terracognita origin (docker or from source) and finally the version you are using.

### Issue Types

There are 6 types of labels, they can be used for issues or PRs:

- `enhancement`: These track specific feature requests and ideas until they are completed. They can evolve from a `specification` or they can be submitted individually depending on the size.
- `specification`: These track issues with a detailed description, this is like a proposal.
- `bug`: These track bugs with the code
- `docs`: These track problems with the documentation (i.e. missing or incomplete)
- `maintenance`: These tracks problems, update and migration for dependencies / third-party tools
- `refactoring`: These tracks internal improvement with no direct impact on the product
- `need review`: this status must be set when you feel confident with your submission
- `in progress`: some important change has been requested on your submission, so you can toggle from `need review` to `in progress`
- `under discussion`: it's time to take a break, think about this submission and try to figure out how we can implement this or this

## Submit a contribution

### Setup your git repository

If you want to contribute to an existing issue, you can start by _forking_ this repository, then clone your fork on your machine.

```shell
$ git clone https://github.com/<your-username>/terracognita.git
$ cd terracognita
```

In order to stay updated with the upstream, it's highly recommended to add `cycloidio/terracognita` as a remote upstream.

```shell
$ git remote add upstream https://github.com/cycloidio/terracognita.git
```

Do not forget to frequently update your fork with the upstream.

```shell
$ git fetch upstream --prune
$ git rebase upstream/master
```

### Play with the codebase

#### Build from sources

Since Terracognita is a Go project, Go must be installed and configured on your machine (really ?). We currently support Go1.12 and Go.13 and go `modules` as dependency manager. You can simply pull all necessaries dependencies by running an initial.

```shell
$ make build
```

This basically builds `terracognita` with the current sources.

You also need to install other code dependencies not mandatory in the runtime environment:
  * [enumer](https://github.com/dmarkham/enumer) is used to generate some code.
  * [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports) is used to format / organize code and imports. CI will perform a check on this, we highly recommend to run `$ goimports -w <your-modified-files>`.
  * [mockgen](https://github.com/golang/mock) is used to generates the mocks for testing

#### Add a new Provider

:warning: It has to support `github.com/hashicorp/terraform-plugin-sdk@v2.0.0` or higher, providers with `v1` will not work anymore :warning:

For this, please open an issue to describe the provider that you want to add. We will discuss about the best way to help you in the implementation.

All the provider names of packages and so one are the ones used by the provider itself, the one in which all the resources are prefixed with (ex: `aws_instance` is `aws`)

**Provider**

All providers have to follow the [provider.Provider](https://github.com/cycloidio/terracognita/blob/master/provider/provider.go#L14) interface in order to be able to be integrated to the codebase. The best way would be to follow any of the other providers and modify it accordingly to your needs. Also reading the documentation of the interface will help a lot in any questions regarding what means any of the functions.

New providers have to be on the root with the package name being the name of the provider on Terraform (ex: Amazon is `aws`)

The `NewProvider` will need all the information required to initialize the specific provider client(SDK), and the specific provider configuration so then Terraform can use it to read information remotely and load it to the Schemas. All the information has to be sent via parameters, no ENV variables.

**Resources**

**Note:** In the following paragraphs, we will use AWS as example as it is the most advanced of the project, but the logic could apply similarly across the other providers too.

Resources are each of one of the Terraform resources that this specific provider has and can import using TerraCognita. The main idea is that the `provider.Provider.Resources` returns all the instances of one resource. That information has to be fetched from the provider API using the SDK initialized before and for each instance we have to build the Terraform resource ID and return the resource. This then will be used by Terraform to read all the needed data from the provider to load the resource Schema (no need to do anything as this is abstracted into the caller of that function).

To know what's the ID expected by Terraform you can check on the Terraform documentation of the resource ([doc](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/instance#import), sometimes is more complex and it needs several elements to build an ID ([doc](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy#import) and [code](https://github.com/hashicorp/terraform-provider-aws/blob/main/aws/resource_aws_iam_role_policy.go#L171)) so we have to set it on the right format for then be read for Terraform.

There are some specific cases in which a resource is imported with another one, on the AWS case that would be `aws_security_groups`, so some times one resource can generate also other resources when then it's pulled by Terraform.

As providers have a lot of resources we have created custom SDKs on top of the provider SDKs that we use to simplify things. Those wrappers are autogenerated from a configuration that is custom for each Provider. We do this to make adding new resources easy; by automating the process and always returning all elements - thus avoiding repetitive code writing and pagination logic. For example in `aws` we have the [configuration](https://github.com/cycloidio/terracognita/blob/master/aws/cmd/functions.go) which has the [go:generate](https://github.com/cycloidio/terracognita/blob/master/aws/reader/connector.go#L56) that generates the end [reader](https://github.com/cycloidio/terracognita/blob/master/aws/reader/reader.go). We have similar logic for the other SDKs

The next point on the documentation explains how to add new resources to a provider which will help on doing it from zero too.

*Note:* If a provider does not have the `Import` function on that specific resource (meaning it cannot be imported by `terraform import ...` it'll not be imported. But we can dynamically add that function so it can be imported as in [this example](https://github.com/cycloidio/terracognita/blob/dae282137c905c90602ae7e0d82cc7601ff1fc85/aws/resources.go#L1629)

**CMD**

On the `cmd/` pkg you'll need to add the provider subcommand which you can mainly check how the other providers are done but it requires the initialization of the provider and the call to the main function `provider.Import()` with all the required parameters.

Aside from that you also need to add the `resources` subcommand which will list all the resources supported by the provider in TerraCognita

**Documentation**

You should also add documentation as the one how to add new resources to the provider (like the following documentation section) and to the main README with the current version used of the provider

#### Add a component to an existing provider (AWS, GCP or AzureRM)

We currently support three providers: Amazon Web Services (AWS), Google Cloud Platform (GCP) and Azure Resource Manager (AzureRM). If you want to play around with one of this provider, you can follow the following guidelines.

For both providers, you need to add your component to this places or equivalent:

* As a `const` value:
  * [here](https://github.com/cycloidio/terracognita/blob/21f3387820bae5577c277acb634ebde5607c23ec/aws/resources.go#L23) for AWS.
  * [here](https://github.com/cycloidio/terracognita/blob/21f3387820bae5577c277acb634ebde5607c23ec/azurerm/resources.go#L17) for AzureRM.
  * [here](https://github.com/cycloidio/terracognita/blob/21f3387820bae5577c277acb634ebde5607c23ec/google/resources.go#L19) for Google.
  * [here](https://github.com/cycloidio/terracognita/blob/21f3387820bae5577c277acb634ebde5607c23ec/vsphere/resources.go#L19) for vSphere.

:warning: your component name must exactly map the name as in Terraform documentation. Example: `ComputeInstance` will be later used as `google_compute_instance` :warning:.

```go
//go:generate enumer -type ResourceType -addprefix google_ -transform snake -linecomment
```

This means you will need to generate some code.

* As a `key/value`, in this:
  * [array](https://github.com/cycloidio/terracognita/blob/21f3387820bae5577c277acb634ebde5607c23ec/aws/resources.go#L111) for AWS.
  * [array](https://github.com/cycloidio/terracognita/blob/21f3387820bae5577c277acb634ebde5607c23ec/azurerm/resources.go#L30) for AzureRM.
  * [array](https://github.com/cycloidio/terracognita/blob/21f3387820bae5577c277acb634ebde5607c23ec/google/resources.go#L46) for Google.

This is where you map your component with a middleware function.

```go
var (
	resources = map[ResourceType]rtFn{
		...
		ComputeInstance: computeInstances,
	}
)
```

Now, you can write your `computeInstances` function. Check-out the other functions, you basically fetch resources from a middleware layer and you add them as Terraform resources.

##### AWS Middleware layer

We have an `aws/cmd` that generates the `aws/reader` interface, which is then used by each resource. To add a new call you have to add a new Function to the list in `aws/cmd/functions.go` and run `make generate`, you'll have the code fully generated for that function. If it has a specific implementation, which is too different from the others you can check the `ListBuckets` Function.

###### Example with aws_db_parameter_group

1. Add your function

Functions are based on `https://github.com/aws/aws-sdk-go/tree/master/service`.
For example with `aws_db_parameter_group` you should be able to find the Entity, Prefix, Service in https://github.com/aws/aws-sdk-go/blob/master/service/rds/rdsiface/interface.go#L343
In our case we want to read data, the dedicated function is `DescribeDBParameterGroups`.

  * Entity: use the function name without prefix "DBParameterGroups"
  * Prefix: use the function prefix (generally Describe, List or Get)
  * Service: SubDirectory from aws-sdk-go service `service/rds`

```shell
$ vim aws/cmd/functions.go
```
```go

	Function{
		Entity:  "DBParameterGroups",
		Prefix:  "Describe",
		Service: "rds",
		Documentation: `
		// GetDBParameterGroups returns all DB parameterGroups based on the input given.
		// Returned values are commented in the interface doc comment block.
		`,
	},

```

Functions are used to generate Get methods in `aws/reader/reader.go`. After a `make generate` execution, we should find a `GetDBParameterGroups` method at the end of the file corresponding to our previous example.

2. Add your resource type

The naming is based on terraform resources like `aws_db_parameter_group` but without the `aws` prefix and Snake case like `bla_foo` needs to be replaced by Camel case  `blaFoo`. The function should start with a `lowercase` and end with a `s`.
The end result would be `dbParameterGroups`.

```shell
$ vim aws/resources.go
```
```go
const (
	...
	DBParameterGroup
)

...

var (
	resources = map[ResourceType]rtFn{
		DBParameterGroup: dbParameterGroups,
        ...
	}
)
```

The const is used to generate `aws/resourcetype_enumer.go`.

3. Add the associated function to generate terraform codes.

```go
func dbParameterGroups(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	dbParameterGroups, err := a.awsr.GetDBParameterGroups(ctx, nil)

	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range dbParameterGroups.DBParameterGroups {

		r, err := initializeResource(a, *i.DBParameterGroupName, resourceType)
		if err != nil {
			return nil, err
		}

		resources = append(resources, r)
	}

	return resources, nil
}
```

4. Make generate

Last step is to re-generate the following files with enumer and build/install.

  * `aws/reader/reader.go`
  * `aws/resourcetype_enumer.go`

```shell
$ make generate
$ make test
```

5. Update CHANGELOG

Don't forget to update the `CHANGELOG.md`

###### How to update terraform-provider-aws

If aws provider need to updated

> :warning: **Only Cycloid can do it**: Since we use a fork of terraform-provider-aws on Cycloid repository.

Since terraform-provider-azurerm moved, source code under `internal`, Terracognita use a fork including a commit to create [aws/provider.go](https://github.com/cycloidio/terraform-provider-azurerm/commit/e3be7857050579e69bea178b83bcae25fb1d4e3f) file.
The process to update terraform-provider-azurerm is the following

```
PROVIDER=aws make update-terraform-provider
```

##### AzureRM Middleware layer

We have an `azurerm/cmd` that generates the `azurerm/reader_generated.go` methods used to get the resources from the Azure SDK. To add a new call you have to add a new Function to the list in `aws/cmd/generate.go` (and the corresponding AzureAPI if necessary) and run `make generate`, you'll have the code fully generated for that function.

###### Example with azurerm_virtual_machine

1. Add your function

Functions are based on `https://github.com/Azure/azure-sdk-for-go/tree/master/services`.
For example with `azurerm_virtual_machine` you should be able to find the required information in https://github.com/Azure/azure-sdk-for-go/blob/master/services/compute/mgmt/2019-12-01/compute/virtualmachines.go
The API we will use for this example is the `compute` service with the version `2019-12-01`.
In our case we want to list data from the ResourceGroup used when executing `terracognita`, the corresponding function is `List`, sometimes another List function is needed as it's depends if it based on a ResourceGroup, a Location or will only list all resources without those delimiters.
The functions are generated using templates that are defined at `azurerm/cmd/template.go`, they have different variables allowing to cover most of scenarios. For the current example:

  * API: SubDirectory azure-sdk-for-go services `compute`
  * APIVersion: SubDirectory azure-sdk-for-go mgmt `2019-12-01`
  * ResourceName: resource name in singular PascalCase `VirtualMachine`
  * ResourceGroup: `true` as the List function require the resource group name as a parameter

```shell
$ vim aws/cmd/generate.go
```
```go
var azureAPIs = []AzureAPI{
	...
	AzureAPI{API: "compute", APIVersion: "2019-07-01"},
}

var functions = []Function{
	...
	Function{Resource: "VirtualMachine", API: "compute", ResourceGroup: true},
}
```

Functions are used to generate List methods in `azurerm/reader_generated.go`. After a `make generate` execution, we should find a `ListVirtualMachines` method at the end of the file corresponding to our previous example.

**Tips!** Sometimes it could be easier to check directly on terraform-provider-azure, each resource is defined in a specific file at https://github.com/hashicorp/terraform-provider-azurerm/tree/main/internal/services/<API>. To check the client to use just go to the Read method, but note that sometimes some logic may be needed to retrieve the resource list.

2. Add your resource type

The naming is based on terraform resources like `azurerm_virtual_machine` but without the `azurerm` prefix and Snake case like `bla_foo` needs to be replaced by Camel case  `blaFoo`. The function should start with a `lowercase` and end with a `s`.
The end result would be `virtualMachines`.

```shell
$ vim azurerm/resources.go
```
```go
const (
	...
	VirtualMachine ResourceType = iota
)

...

var (
	resources = map[ResourceType]rtFn{
		...
		VirtualMachine: virtualMachines,
	}
)
```

The const is used to generate `azurerm/resourcetype_enumer.go`.

3. Add the associated function to generate terraform codes.

```go
func virtualMachines(ctx context.Context, a *azurerm, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	virtualMachines, err := a.azurer.ListVirtualMachines(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list virtual machines from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, virtualMachine := range virtualMachines {
		r := provider.NewResource(*virtualMachine.ID, resourceType, a)
		resources = append(resources, r)
	}
	return resources, nil
}
```
5. Optional: If the resource data is required to list other resources, you should add caching, e.g.: listing virtual machine extension requires the name of the virtual machines.

	1. To enable this you should first change the resource code at `azurerm/resources.go` by adding the setting the resource data name on the method , check other virtualnetwork method for an example of this.

	2. Then create the caching functions to use to retrieve the resources names at `azurerm/cache.go`, it should be straigthfroad and easy to do, just follow the example of the other functions.

6. Make generate

Last step is to re-generate the following files with enumer and build/install.

  * `azurerm/reader_generated.go`
  * `azurerm/resourcetype_enumer.go`

```
$ make generate
$ make test
```

6. Update CHANGELOG

Don't forget to update the `CHANGELOG.md`.

###### How to update terraform-provider-azurerm

If azurerm provider need to updated

> :warning: **Only Cycloid can do it**: Since we use a fork of terraform-provider-azurerm on Cycloid repository.

Since terraform-provider-azurerm moved, source code under `internal`, Terracognita use a fork including a commit to create [azurerm/provider.go](https://github.com/cycloidio/terraform-provider-azurerm/commit/e3be7857050579e69bea178b83bcae25fb1d4e3f) file.
The process to update terraform-provider-azurerm is the following

```
PROVIDER=azurerm make update-terraform-provider
```

##### GCP Middleware layer

In `reader.go`, you can add your middleware function `ListInstances`. You will need to be equipped with this [documentation](https://godoc.org/google.golang.org/api/compute/v1). Google SDK is pretty standard, APIs are most of the time used in a similar way.
You only need to find out if your component belongs to a `project` or a `project` and a `zone`. It's highly recommended to base your code on the other functions (a method to generate this function will be provided soon).

We have a `google/cmd` that generates the `google/reader_generated.go` methods used to get the resources from the GCP SDK. To add a new call you have to add a new Function to the list in `google/cmd/generate.go` (and the corresponding GCP API if necessary) and run `make generate`, you'll have the code fully generated for that function.

###### Example with google_compute_instance

1. Add your function

Functions are based on `https://github.com/googleapis/google-api-go-client`. Google SDK is a bit different than the other providers, there's not a file per object but a file per api so it makes it a bit more tricky to search for information. You need to find the corresponding API file,for example with `google_compute_instance` you should be able to find the required information in https://github.com/googleapis/google-api-go-client/blob/main/compute/v1/compute-gen.go. The API to use is the `compute` service with the version `v1`.
*TIP!* A quick way to find the API is to check the corresponding import in the [terraform-provider-google](https://github.com/hashicorp/terraform-provider-google) repository for the given resource.

Once you know the API file to use to retrieve the information you need to write your function is quite easy:

i. Check the `Services` struct where all the API services are specified, in the case of the example `InstancesService`,
ii. Using the service name search in the file for the `List` method that applies to this Service and returns a ListCall pointer for the object, in the case of the example `*InstancesListCall`
iii. Then using this return object search for the available methods to apply to this call, these are the methods you'll need to define the function
*Note!* you can also retrieve this information using the online docs for the example it would be here https://cloud.google.com/compute/docs/reference/rest/v1/instances/list

With this information you can now create your functions using the template parameters:
  * Resource: resource name in singular PascalCase
  * API: the api name by default compute
  * Region/Zone: if the list method requires one of these parameters
You can find the full list of available parameters at `google/cmd/template.go`

```shell
$ vim google/cmd/generate.go
```
```go
var functions = []Function{
	...
	Function{Resource: "Instance", Zone: true},
}
```

Functions are used to generate List methods in `google/reader_generated.go`. After a `make generate` execution, we should find a `ListInstances` method at the end of the file corresponding to our previous example.

*Note!* If you cannot use the template to create the List method, because the resource is too irregular, you can define it yourself in the file `google/reader_irregular_cases.go`

2. Add your resource type

The naming is based on terraform resources like `google_compute_instance` but without the `google` prefix and Snake case like `bla_foo` needs to be replaced by Camel case  `blaFoo`. The end result would be `ComputeInstance`.

```shell
$ vim azurerm/resources.go
```
```go
const (
	...
	ComputeInstance ResourceType = iota
)

...

var (
	resources = map[ResourceType]rtFn{
		...
		ComputeInstance: computeInstance,
	}
)
```

The const is used to generate `google/resourcetype_enumer.go`.

3. Add the associated function to generate terraform codes.

```go
func computeInstance(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	f := initializeFilter(filters)
	instancesList, err := g.gcpr.ListInstances(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list instances from reader")
	}
	resources := make([]provider.Resource, 0)
	for z, instances := range instancesList {
		for _, instance := range instances {
			r := provider.NewResource(fmt.Sprintf("%s/%s/%s", g.Project(), z, instance.Name), resourceType, g)
			resources = append(resources, r)
		}
	}
	return resources, nil
}
```

4. Make generate

Last step is to re-generate the following files with enumer and build/install.

  * `google/reader_generated.go`
  * `google/resourcetype_enumer.go`

```
$ make generate
$ make test
```

5. Update CHANGELOG

Don't forget to update the `CHANGELOG.md`.


#### Build and test your component

That's it ! You can now generate some code and build your binary:

```shell
$ make generate && make build
$ ./terracognita google resources
...
google_compute_instance
...
```
###### How to update terraform-provider-google

Only if google provider need to updated

```
PROVIDER=google make update-terraform-provider
```

##### Terraform Middleware layer

Terraform is used to generate HCL or/and Terraform State. At some point, the Terraform version might need to be updated.

###### How to update terraform

Only if terraform need to updated

```
PROVIDER=terraform make update-terraform-provider
```
