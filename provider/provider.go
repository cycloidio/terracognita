package provider

import (
	"context"

	"github.com/cycloidio/terracognita/filter"
)

// Provider is the ggeneral interface used to abstract
// a cloud provider from Terraform
type Provider interface {
	// Region returns the actual region in which the
	// provider is based
	Region() string

	// ResourceTypes returns all the resource types from
	// the Provider
	ResourceTypes() []string

	// Resources returns all the Resources of the resourceType
	// on the cloud provider
	Resources(ctx context.Context, resourceType string, f filter.Filter) ([]Resource, error)

	// TFClient returns the Terraform client which may change
	// on the provider
	TFClient() interface{}

	// String returns the string representation of the Provider
	// which is the shorted version (Amazon Web Services = aws)
	String() string

	// TagKey returns the different name used to identify
	// tags on the cloud provider
	TagKey() string
}
