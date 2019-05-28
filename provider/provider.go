package provider

import (
	"context"

	"github.com/cycloidio/terraforming/filter"
)

type Provider interface {
	Region() string
	ResourceTypes() []string
	Resources(ctx context.Context, resourceType string, f filter.Filter) ([]*Resource, error)
	TFClient() interface{}
	String() string
	TagKey() string
}
