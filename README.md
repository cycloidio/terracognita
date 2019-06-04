# Terracognita

Imports your current Cloud infrastructure to an Infrastructure As Code [Terraform](https://www.terraform.io/) configuration (HCL) or/and to a Terraform State.

To do so, it relies on [raws](https://github.com/cycloidio/raws/) a self-generated AWS reader.
At this stage, raws supports various technologies. https://github.com/cycloidio/raws/issues/8

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
 * AWS: 2.9.0

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
$ export AWS_REGION=xx-yyyy-0
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

## License

Please see the [MIT LICENSE](https://github.com/cycloidio/raws/blob/master/LICENSE) file.

## About Cycloid

![Cycloid logo](https://pbs.twimg.com/profile_images/786183670080086016/0O9JzolW_400x400.jpg)

This code is maintened by Cycloid full-remote European team https://www.cycloid.io/, firstly for his own usage to accelerate Cloud migration and keeping the respect of the best practices then for the Open Source world. Respecting the best practices is not an option and a DevOps culture doesn't mean open bar for everyone. We know what does it mean to maintain infrastructure on production. Azure, GCP, Alibaba, Vmware and OpenStack are in progressed to be released soon. 

This tools comes as a piece of Cycloid, your DevOps platform to help Dev & Ops working together with a respect of the best practices. Our goal are simple: simplify, accelerate and optimize your DevOps and Cloud adoption.

Cycloid includes a Cloud designer with StackCraft that allows you to drag & drop Terraform ressources and generate your Terraform files for you. You will also find a CI/CD pipeline with Concourse as a service, an Open Source service catalogue, the presentation of the infrastructure and the application, the cost management control to centralize Cloud billing, the monitoring, logs and events centralized with Prometheus, Grafana, ELK. You will also find various automation to simplify the DevOps adoption.

