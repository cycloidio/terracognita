# Terraforming

Imports the current cloud infrastructure to a Terraform configuration (HCL) or/and to a Terraform State

## Installation

```
$> go get -u github.com/cycloidio/terraforming
$> cd $GOPATH/src/github/cycloidio/terraforming
$> go install .
```

## Usage

Using the `terraforming --help` you will know the basics.

## Versions

Right now it uses `terraform 0.12` and for the Providers:

* AWS: 2.9.0

## Docker

To build the Docker image just run

```
$> make dbuild
```

And then:

```
$> docker run terraforming -h
```

Building it manually may cause an error on `exec: "bzr": executable file not found in $PATH` you need to install `bzr` lib

## Contribute

It uses Go Modules, so GO 11 or higher is required
