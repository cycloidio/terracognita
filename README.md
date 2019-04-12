# Terraforming

Imports cloud infrastructure to a Terraform file

## Installation

```
$> go get -u github.com/cycloidio/terraforming
```

## Usage

Using the `terraforming --help` you will know the basics

## Docker

To build the Docker image just run

```
$> make dbuild
```

And then:

```
$> docker run terraforming -h
```

## Contribute

It uses Go Modules, so GO 11 >= is needed
