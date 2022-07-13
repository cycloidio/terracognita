package provider

import (
	"context"

	"github.com/cycloidio/terracognita/filter"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

//go:generate mockgen -destination=../mock/provider.go -mock_names=Provider=Provider -package mock github.com/cycloidio/terracognita/provider Provider

// Provider is the general interface used to abstract
// a cloud provider from Terraform
type Provider interface {
	// Region returns the actual region in which the
	// provider is based
	Region() string

	// ResourceTypes returns all the resource types from
	// the Provider
	ResourceTypes() []string

	// HasResourceType validates if the string t is a valid
	// resource type for this provider
	HasResourceType(t string) bool

	// Resources returns all the Resources of the resourceType
	// on the cloud provider
	Resources(ctx context.Context, resourceType string, f *filter.Filter) ([]Resource, error)

	// TFClient returns the Terraform client which may change
	// on the provider
	TFClient() interface{}

	// TFProvider returns the Terraform provider
	TFProvider() *schema.Provider

	// String returns the string representation of the Provider
	// which is the shorted version (Amazon Web Services = aws)
	String() string

	// TagKey returns the different name used to identify
	// tags on the cloud provider
	TagKey() string

	// Source is the source of the Provider used
	// to declare on the HCL
	Source() string

	// Configuration returns the Provider configuration
	// that may be interpolated with HCL when declaring
	// the provider. The keys have to be the Provider
	// attributes as defined on the TF Schema
	Configuration() map[string]interface{}
}
