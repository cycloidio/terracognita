<p align="center">
  <img src="docs/terracognita-icon_color.png" width="150">
</p>

# TerraCognita

[![GoDoc](https://godoc.org/github.com/cycloidio/terracognita?status.svg)](https://godoc.org/github.com/cycloidio/terracognita)
[![Gitter chat](https://badges.gitter.im/gitterHQ/gitter.png)](https://gitter.im/cycloidio/terracognita)
[![AUR package](https://repology.org/badge/version-for-repo/aur/terracognita.svg)](https://repology.org/project/terracognita/versions)
[![Homebrew](https://img.shields.io/badge/dynamic/json.svg?url=https://formulae.brew.sh/api/formula/terracognita.json&query=$.versions.stable&label=homebrew)](https://formulae.brew.sh/formula/terracognita)

Imports your current Cloud infrastructure to an Infrastructure As Code [Terraform](https://www.terraform.io/) configuration (HCL) or/and to a Terraform State.

At [Cycloid](https://www.cycloid.io/), Infrastructure As Code is in the company DNA since the beginning. To help our new customers adopting this best practice, we decided to build Terracognita to convert an existing infrastructure on Cloud Infrastructure into Terraform code in an automated way, relying on Terraform providers built by the community. We focused on AWS, GCP and Azure but Alibaba, Vmware and Openstack will be the next to be integrated.

We decided to Open Source this tool as we believe that it will help people to adopt IaC in an easy way. Cycloid provides this tool to let people import their infrastructure into [Cycloid's pipelines](https://www.cycloid.io/devops-framework), allow them to generate infrastructure diagram and manage all infra/application life cycle from a single interface.

If you are interested in contributing to Terracognita or simply curious about what's next, take a look at the public [roadmap](https://github.com/cycloidio/terracognita/issues/118). For a high level overview, check out the [What is Terracognita?](https://blog.cycloid.io/what-is-terracognita) blogpost or watch this [video](https://www.youtube.com/watch?v=PtJgf7AWBVA).

## Cloud providers

Terracognita currently imports AWS, GCP, AzureRM and VMware vSphere cloud providers as Terraform (v1.1.9) resource/state.
Please see the following versions as follow:

Providers:
 * AWS: v4.9.0
 * AzureRM: v3.6.0
 * Google: v4.9.0
 * vSphere: v2.2.0

## Installation

### Binary

Visit the [releases](https://github.com/cycloidio/terracognita/releases) page to select your system, architecture and version you need. To pull the latest release:

```shell
curl -L https://github.com/cycloidio/terracognita/releases/latest/download/terracognita-linux-amd64.tar.gz -o terracognita-linux-amd64.tar.gz
tar -xf terracognita-linux-amd64.tar.gz
chmod u+x terracognita-linux-amd64
sudo mv terracognita-linux-amd64 /usr/local/bin/terracognita
```

### Development

You can build and install with the latest sources, you will enjoy the new features and bug fixes. It uses Go Modules, so GO 1.17+ is required.

```shell
git clone https://github.com/cycloidio/terracognita
cd terracognita
make install
```

### Arch Linux

There are two entries in the AUR: [terracognita-git](https://aur.archlinux.org/packages/terracognita-git/) (targets the latest git commit) and [terracognita](https://aur.archlinux.org/packages/terracognita) (targets the latest stable release).

```shell
yay -Ss terracognita
  aur/terracognita 1:0.3.0-1 (+0 0.00%)
      Reads from existing Cloud Providers (reverse Terraform) and generates your infrastructure as code on Terraform configuration
  aur/terracognita-git 1:v0.3.0.r27.gdfc5a99-1 (+0 0.00%)
      Reads from existing Cloud Providers (reverse Terraform) and generates your infrastructure as code on Terraform configuration
```

### Install via brew

If you're macOS user and using [Homebrew](https://brew.sh/), you can install via brew command:

```sh
brew install terracognita
```

## Usage

The main usage of Terracognita is:

```bash
terracognita [TERRAFORM_PROVIDER] [--flags]
```

You replace the `TERRAFORM_PROVIDER` with the Provider you want to use (for example `aws`) and then add the other required flags.
Each Provider has different flags and different required flags.

The more general ones are the `--hcl` or `--module` and `--tfstate` which indicates the output file for the HCL (or module)
and the TFState that will be generated.

You can also `--include` or `--exclude` multiple resources by using the Terraform name it has like `aws_instance`.

For more options you can always use `terracognita --help` and `terracognita [TERRAFORM_PROVIDER] --help` for the
specific documentation of the Provider.

We also have `make help` that provide some helpers on using the actual codebase of Terracognita

[![asciicast](https://asciinema.org/a/330055.svg)](https://asciinema.org/a/330055)

### Modules

Terracognita can generate Terraform Modules directly when importing. To enable this feature you'll need to use the `--module {module/path/name}` and then on that specific path is where the module will be generated. The path has to be directory or a none existent path (it'll be created), the content of the path will be deleted (after user confirmation) so we can have a clean import.

The output structure will look like (having `--module test`) this where each file aggregates the resources from the same "category":

```
test/
├── module-test
│   ├── autoscaling.tf
│   ├── cloud_front.tf
│   ├── cloud_watch.tf
│   ├── ec2.tf
│   ├── elastic_load_balancing_v2_alb_nlb.tf
│   ├── iam.tf
│   ├── rds.tf
│   ├── route53_resolver.tf
│   ├── route53.tf
│   ├── s3.tf
│   ├── ses.tf
│   └── variables.tf
└── module.tf
```

By default all the attributes will be changed for variables, those variables will then be on the `module-{name}/variables.tf` and exposed on the `module.tf` like so:

```hcl
module "test" {
  # aws_instance_front_instance_type = "t2.small"
  [...]
  source = "module-test"
}
```

If you want to change this behavior, as for big infrastructures this will create a lot of variables, you can use the `--module-variables path/to/file` and the file will have the list of attributes that you want to actually be used as variables, it can be in JSON or YAML:

```json
{
  "aws_instance": [
    "instance_type",
    "cpu_threads_per_core",
    "cpu_core_count"
  ]
}
```

```yaml
aws_instance:
  - instance_type
  - cpu_threads_per_core
  - cpu_core_count
```

### Docker

You can use directly [the image built](https://hub.docker.com/r/cycloid/terracognita), or you can build your own.
To build your Docker image just run:

```bash
make dbuild
```

And then depending on the image you want to use (`cycloid/terracognita` or your local build `terracognita`):

```bash
docker run cycloid/terracognita -h
```

Example:

```bash
export AWS_ACCESS_KEY_ID=XXXXXXXXXXXXXXXXXXXX
export AWS_SECRET_ACCESS_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
export AWS_DEFAULT_REGION=xx-yyyy-0
docker run \
		-v "${PWD}"/outputs:/app/outputs \
		cycloid/terracognita aws \
		--hcl app/outputs/resources.tf
```

### Local

The local version can be used the same way as docker. You simply need to be build it locally.

### To test

On the same folder you imported you can run the terraform init & plan commands:

```bash
terraform init
terraform plan -var access_key=$AWS_ACCESS_KEY_ID -var secret_key=$AWS_SECRET_ACCESS_KEY -var region=$AWS_DEFAULT_REGION
```

## License

Please see the [MIT LICENSE](https://github.com/cycloidio/terracognita/blob/master/LICENSE) file.

## Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Meet Cycloid

[Cycloid](https://www.cycloid.io/) is a hybrid cloud DevOps collaboration platform providing end-to-end frameworks to accelerate and industrialize software delivery.

As of now, we have three open-source tools:

* [TerraCognita](https://github.com/cycloidio/terracognita): Read from your existing cloud providers and generate IaC in Terraform 
* [InfraMap](https://github.com/cycloidio/inframap): Reads .tfstate or HCL to generate a graph specific for each provider
* [TerraCost](https://github.com/cycloidio/terracost): Cloud cost estimation for Terraform in the CLI

...and the functionality of each is also embedded in our DevOps solution, which you can find out more about [here](https://www.cycloid.io/hybrid-cloud-devops-platform).
