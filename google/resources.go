package google

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cycloidio/terracognita/provider"
	"github.com/cycloidio/terracognita/tag"
)

// ResourceType is the type used to define all the Resources
// from the Provider
type ResourceType int

//go:generate enumer -type ResourceType -addprefix google_ -transform snake -linecomment
const (
	ComputeInstance ResourceType = iota
)

type rtFn func(ctx context.Context, g *google, resourceType string, tags []tag.Tag) ([]provider.Resource, error)

var (
	resources = map[ResourceType]rtFn{
		ComputeInstance: computeInstance,
	}
)

func initializeFilter(tags []tag.Tag) string {
	var b bytes.Buffer
	for _, t := range tags {
		// if multiple tags, we suppose it's a "AND" operation
		b.WriteString(fmt.Sprintf("(labels.%s=%s) ", t.Name, t.Value))
	}
	return b.String()
}

func computeInstance(ctx context.Context, g *google, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	f := initializeFilter(tags)
	instancesList, err := g.gcpr.ListInstances(ctx, f)
	if err != nil {
		return nil, err
	}
	resources := make([]provider.Resource, 0)
	for z, instances := range instancesList {
		for _, instance := range instances {
			r := provider.NewResource(fmt.Sprintf("%s/%s/%s", g.Project(), z, instance.Name), resourceType, g)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}
	return resources, nil
}
