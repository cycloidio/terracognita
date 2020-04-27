# TerraCognita

[![GoDoc](https://godoc.org/github.com/cycloidio/terracognita?status.svg)](https://godoc.org/github.com/cycloidio/terracognita)
[![Gitter chat](https://badges.gitter.im/gitterHQ/gitter.png)](https://gitter.im/cycloid-community/terracognita)
[![AUR package](https://repology.org/badge/version-for-repo/aur/terracognita.svg)](https://repology.org/project/terracognita/versions)

Imports your current Cloud infrastructure to an Infrastructure As Code [Terraform](https://www.terraform.io/) configuration (HCL) or/and to a Terraform State.

At [Cycloid](https://www.cycloid.io/), Infrastructure As Code is in the company DNA since the beginning. To help our new customers adopting this best practice, we decided to build Terracognita to convert an existing infrastructure on Cloud Infrastructure into Terraform code in an automated way, relying on Terraform providers built by the community. We focused on AWS, GCP and Azure but Alibaba, Vmware and Openstack will be the next to be integrated.

We decided to opensource this tool as we believe that it will help people to adopt IaC in an easy way. Cycloid provides this tool to let people import their infrastructure into [Cycloid's pipelines](https://www.cycloid.io/devops-framework), allow them to generate infrastructure diagram and manage all infra/application lifecycle from a single interface.

## Cloud providers

Terracognita currently imports AWS, GCP and AzureRM cloud provider as terraform (0.12.8) resource/state.
Please see the following versions as follow:

Providers:
 * AWS: 2.31.0
 * AzureRM: 1.35.0
 * GCP: 2.16.0

## Installation

### Binary

Visit the [releases](https://github.com/cycloidio/terracognita/releases) page to select your system, architecture and version you need. Then:

```shell
$ curl -L https://github.com/cycloidio/terracognita/releases/download/v0.3.0/terracognita-linux-amd64.tar.gz -o terracognita-linux-amd64.tar.gz
$ tar -xf terracognita-linux-amd64.tar.gz
$ chmod u+x terracognita-linux-amd64
$ sudo mv terracognita-linux-amd64 /usr/local/bin/terracognita
```

### Development

You can build and install with the latest sources, you will enjoy the new features and bugfixes. It uses Go Modules, so GO 1.11+ is required.

```shell
$ git clone https://github.com/cycloidio/terracognita
$ cd terracognita
$ make install
```

### Arch Linux

There are two entries in the AUR: [terracognita-git](https://aur.archlinux.org/packages/terracognita-git/) (targets the latest git commit) and [terracognita](https://aur.archlinux.org/packages/terracognita) (targets the latest stable release).

```shell
$ yay -Ss terracognita
aur/terracognita 1:0.3.0-1 (+0 0.00%) 
    Reads from existing Cloud Providers (reverse Terraform) and generates your infrastructure as code on Terraform configuration
aur/terracognita-git 1:v0.3.0.r27.gdfc5a99-1 (+0 0.00%) 
    Reads from existing Cloud Providers (reverse Terraform) and generates your infrastructure as code on Terraform configuration
```

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

[![asciicast](https://asciinema.org/a/252604.svg)](https://asciinema.org/a/252604)

### Docker

You can use directly [the image built](https://hub.docker.com/r/cycloid/terracognita), or you can build your own.
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

Please see the [MIT LICENSE](https://github.com/cycloidio/terracognita/blob/master/LICENSE) file.

## Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## About Cycloid
<p align="center">
  <img src="https://user-images.githubusercontent.com/393324/65147266-0b010100-da1e-11e9-9a49-d27e5035c4c4.png">
</p>

[Cycloid](https://www.cycloid.io/our-culture) is a European fully-remote company, building a product to **simplify**, **accelerate** and **optimize your DevOps and Cloud adoption**.

We built [Cycloid, your DevOps framework](https://www.cycloid.io/devops-framework) to encourage Developers and Ops to work together with the respect of best practices. We want to provide a tool that eliminates the silo effect in a company and allows to share the same level of informations within all professions.

[Cycloid](https://www.cycloid.io/devops-framework) supports you to factorize your application in a reproducable way, to deploy a new environment in one click. This is what we call a stack.

A stack is composed of 3 pillars:

1. the pipeline ([Concourse](https://concourse-ci.org/))
2. infrastructure layer ([Terraform](https://www.terraform.io/))
3. applicative layer ([Ansible](https://www.ansible.com/))

Thanks to the flexible pipeline, all the steps and technologies are configurable.

To make it easier to create a stack, we build an Infrastructure designer named **StackCraft** that allows you to drag & drop Terraform ressources and generate your Terraform files for you.

Terracognita is a brick that will help us to import an existing infrastructure into a stack to easily adopt Cycloid product.

The product comes also with an Open Source service catalog ([all our public stacks are on Github](https://github.com/cycloid-community-catalog)) to deploy applications seamlessly.
To manage the whole lifecycle of an application, it also integrates the diagram of the infrastructure and the application, a cost management control to centralize Cloud billing, the monitoring, logs and events centralized with Prometheus, Grafana, ELK.

[Don't hesitate to contact us, we'll be happy to meet you !](https://www.cycloid.io/contact-us)
