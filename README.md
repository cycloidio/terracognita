# Terracognita

Imports your current cloud infrastructure to a [Terraform](https://www.terraform.io/) configuration (HCL) or/and to a Terraform State.

To do so it relies on [raws](https://github.com/cycloidio/raws/) a self-generated AWS reader.
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

Please see the [LICENSE](https://github.com/cycloidio/raws/blob/master/LICENSE) file.
