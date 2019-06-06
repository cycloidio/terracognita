# TerraCognita (beta)

Imports your current Cloud infrastructure to an Infrastructure As Code [Terraform](https://www.terraform.io/) configuration (HCL) or/and to a Terraform State.

To do so, it relies on [raws](https://github.com/cycloidio/raws/) a self-generated AWS reader.
At this stage, raws supports [various technologies](https://github.com/cycloidio/raws/issues/8).

At [Cycloid](https://www.cycloid.io/), Infrastructure As Code is in the company DNA since the beginning. To help our new customers adopting this best practice, we decided to build Terracognita to convert an existing infrastructure on AWS into Terraform code in an automated way, relying on Terraform providers built by the community. We focused on AWS for a first release, but Azure, GCP, Alibaba, Vmware and Openstack will be the next to be integrated.

We decided to opensource this tool as we believe that it will help people to adopt IaC in an easy way. Cycloid provides this tool to let people import their infrastructure into [Cycloid's pipelines]([Cycloid.io](https://www.cycloid.io/devops-platform-with-ci-cd-container-pipeline), allow them to generate infrastructure diagram and manage all infra/application lifecycle from a single interface.

## Installation

It uses Go Modules, so GO 11 or higher is required.
If you wish to run it via docker then `docker` binary is required.
While if you wish to run it locally; you will need to have the extra `brz` lib installed.

```
$> go get -u github.com/cycloidio/terracognita
$> cd $GOPATH/src/github/cycloidio/terracognita
$> go install .
```

## Versions

Terracognita currently imports only AWS cloud provider as terraform resource/state.
Please see the following versions as follow:

Terraform: 0.11.14
Providers:
 * AWS: 2.12.0

## Usage

Using the `terracognita --help` you will know the basics.

```bash
$ make help
help: Makefile                   This help dialog
lint: $(GOLANGCI_LINT) $(GOLINT) Runs the linter
test:                            Runs the tests
ci: lint test                    Runs the linter and the tests
dbuild:                          Builds the docker image with same name as the binary
build:                           Bulids the binary
clean:                           Removes binary and/or docker image
```

### Docker

You can use directly [the image built](https://cloud.docker.com/u/cycloid/repository/docker/cycloid/terracognita/general), or you can build your own.
To build your Docker image just run:

```bash
$ make dbuild
```

And then depending on the image you want to use (`cycloid/terracognita` or your local build `terracognita`):

```bash
$ docker run cycloid/terracognita -h
```

Example:

```bash
$ export AWS_ACCESS_KEY_ID=XXXXXXXXXXXXXXXXXXXX
$ export AWS_SECRET_ACCESS_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
$ export AWS_DEFAULT_REGION=xx-yyyy-0
$ docker run \
		-v "${PWD}"/outputs:/app/outputs \
		cycloid/terracognita aws \
		--access-key="${AWS_ACCESS_KEY_ID}" \
		--secret-key="${AWS_SECRET_ACCESS_KEY}" \
		--region="${AWS_DEFAULT_REGION}" \
		--hcl app/outputs/resources.tf
```

### Local

The local version can be used the same way as docker. You simply need to be build it locally.

### To test

To speed up the testing, you can write a small `provider.tf`file within the same folder you imported your resources & tfstate:

```bash
terraform {
 backend "local" {
   path = "./$TFSTATE_PATH"
 }
}

provider "aws" {
 access_key = "${var.access_key}"
 secret_key = "${var.secret_key}"
 region     = "${var.region}"
 version    = "2.12.0"
}

variable "access_key" {}
variable "secret_key" {}
variable "region" {}
```

Then run the terraform init & plan commands:

```bash
$ terraform init
$ terraform plan -var access_key=$AWS_ACCESS_KEY_ID -var secret_key=$AWS_SECRET_ACCESS_KEY -var region=$AWS_DEFAULT_REGION
```

## License

Please see the [MIT LICENSE](https://github.com/cycloidio/raws/blob/master/LICENSE) file.

## About Cycloid
<p align="center">
  <img width="300" height="300" src="https://pbs.twimg.com/profile_images/786183670080086016/0O9JzolW_400x400.jpg">
</p>

[Cycloid](https://www.cycloid.io/our-culture) is a European fully-remote company, building a product to **simplify**, **accelerate** and **optimize your DevOps and Cloud adoption**.

We built [Cycloid, your DevOps platform](https://www.cycloid.io/devops-platform-with-ci-cd-container-pipeline) to encourage Developers and Ops to work together with the respect of best practices. We want to provide a tool that eliminates the silo effect in a company and allows to share the same level of informations within all professions.

[Cycloid](https://www.cycloid.io/devops-platform-with-ci-cd-container-pipeline) supports you to factorize your application in a reproducable way, to deploy a new environment in one click. This is what we call a stack.

A stack is composed of 3 pillars:

1. the pipeline ([Concourse](https://concourse-ci.org/))
2. infrastructure layer ([Terraform](https://www.terraform.io/))
3. applicative layer ([Ansible](https://www.ansible.com/))

Thanks to the flexible pipeline, all the steps and technologies are configurable.

To make it easier to create a stack, we build an Infrastructure designer named **StackCraft** that allows you to drag & drop Terraform ressources and generate your Terraform files for you.

Terracognita is a brick that will help us to import an existing infrastructure into a stack to easily adopt Cycloid product.

The product comes also with an Open Source service catalog ([all our public stacks are on Github](https://github.com/cycloid-community-catalog)) to deploy applications seamlessly.
To manage the whole lifecycle of an application, it also integrates the diagram of the infrastructure and the application, a cost management control to centralize Cloud billing, the monitoring, logs and events centralized with Prometheus, Grafana, ELK.

[Don't hesitate to contact us, we'll be happy to meet you !](https://www.cycloid.io/meet-us)
