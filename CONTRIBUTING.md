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
- `refactoring`: These tracks internal improvment with no direct impact on the product
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

In order to stay updated with the upstream, it's highly recommended to add cycloidio/terracognita as a remote upstream.

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

Since Terracognita is a Go project, Go must be installed and configured on your machine (really ?). We currently support Go1.12 and Go.13 and go `modules` as dependency manager. You can simply pull all necessaries dependencies by running an initial

```shell
$ make build
```

This basically builds `terracognita` with the current sources.

You also need to install other code dependencies not mandatory in the runtime environment:
  * [enumer](https://github.com/dmarkham/enumer) is used to play around AWS or GCP provider, it is used to generate some code
  * [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports) is used to format / organize code and imports. CI will perform a check on this, we highly recommend to run `$ goimports -w <your-modified-files>`

#### Add a component to an existing provider (AWS or GCP)

We currently support two providers: Amazon Web Services (AWS) and Google Cloud Platform (GCP). If you want to play around with one of this provider, you can follow the following guidelines.

For both providers, you need to add your component to this places or equivalent:

* As a `const` value [here](https://github.com/cycloidio/terracognita/blob/000789d3bd61b81cf10695d414f4f45346ccc25f/google/resources.go#L17) for Google and [here](https://github.com/cycloidio/terracognita/blob/000789d3bd61b81cf10695d414f4f45346ccc25f/aws/resources.go#L22) for AWS. :warning: your component name must exactly maps the name as in Terraform documentation. Example: `ComputeInstance` will be later used as `google_compute_instance` :warning:.

```go
//go:generate enumer -type ResourceType -addprefix google_ -transform snake -linecomment
```

This means you will need to generate some code.

* As a `key/value`, in this [array](https://github.com/cycloidio/terracognita/blob/000789d3bd61b81cf10695d414f4f45346ccc25f/google/resources.go#L25)([here](https://github.com/cycloidio/terracognita/blob/000789d3bd61b81cf10695d414f4f45346ccc25f/aws/resources.go#L96) for AWS). This is where you map your component with a middleware function.

```go
...
ComputeInstance: computeInstance,
...
```

Now, you can write your `computeInstance` function. Check-out the other functions, you basically fetch resources from a middleware layer and you add them as Terraform resources.

##### AWS Middleware layer

We are using [RAWS](https://github.com/cycloidio/raws) in order to add abstraction between AWS APIs and Terracognita. This is the [list](https://github.com/cycloidio/raws/issues/8) of all the methods implemented by RAWS. If you do not find the one you are looking for: feel free to submit a PR or to add a comment in the issue.

##### GCP Middleware layer

In `reader.go`, you can add your middleware function `ListInstances`. You will need to be equiped with this [documentation](https://godoc.org/google.golang.org/api/compute/v1). Google SDK is pretty standard, APIs are most of the time used in a similar way.
You only need to find out if your component belongs to a `project` or a `project` and a `zone`. It's highly recommended to base your code on the other functions (a method to generate this function will be provided soon).

#### Build and test your component

That's it ! You can now generate some code and build your binary: 

```shell
$ make generate && make build
$ ./terracognita google resources
...
google_compute_instance
...
```

#### Add a new provider

For this, please open an issue to describe the provider that you want to add. We will discuss about the best way to help you in the implementation.
