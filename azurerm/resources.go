package azurerm

import (
	"context"

	"github.com/cycloidio/terracognita/provider"
	"github.com/cycloidio/terracognita/tag"
)

// ResourceType is the type used to define all the Resources
// from the Provider
type ResourceType int

//go:generate enumer -type ResourceType -addprefix azurerm_ -transform snake -linecomment
const (
	DummyResource ResourceType = iota
)

type rtFn func(ctx context.Context, a *azurerm, resourceType string, tags []tag.Tag) ([]provider.Resource, error)

var (
	resources = map[ResourceType]rtFn{
		DummyResource: dummyResource,
	}
)

func dummyResource(ctx context.Context, a *azurerm, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	return nil, nil
}
