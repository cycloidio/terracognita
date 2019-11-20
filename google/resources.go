package google

import (
	"bytes"
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/cycloidio/terracognita/provider"
	"github.com/cycloidio/terracognita/tag"
)

// ResourceType is the type used to define all the Resources
// from the Provider
type ResourceType int

//go:generate enumer -type ResourceType -addprefix google_ -transform snake -linecomment
const (
	ComputeInstance ResourceType = iota
	ComputeFirewall
	ComputeNetwork
	// With Google, an HTTP(S) load balancer has 3 parts:
	// * backend configuration: instance_group, backend_service and health_check
	// * host and path rules: url_map
	// * frontend configuration: target_http(s)_proxy
	ComputeHealthCheck
	ComputeInstanceGroup
	ComputeBackendService
)

type rtFn func(ctx context.Context, g *google, resourceType string, tags []tag.Tag) ([]provider.Resource, error)

var (
	resources = map[ResourceType]rtFn{
		ComputeInstance:       computeInstance,
		ComputeFirewall:       computeFirewall,
		ComputeNetwork:        computeNetwork,
		ComputeHealthCheck:    computeHealthCheck,
		ComputeInstanceGroup:  computeInstanceGroup,
		ComputeBackendService: computeBackendService,
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
		return nil, errors.Wrap(err, "unable to list instances from reader")
	}
	resources := make([]provider.Resource, 0)
	for z, instances := range instancesList {
		for _, instance := range instances {
			r := provider.NewResource(fmt.Sprintf("%s/%s/%s", g.Project(), z, instance.Name), resourceType, g)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func computeFirewall(ctx context.Context, g *google, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	f := initializeFilter(tags)
	firewalls, err := g.gcpr.ListFirewalls(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list firewalls from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, firewall := range firewalls {
		r := provider.NewResource(firewall.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeNetwork(ctx context.Context, g *google, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	f := initializeFilter(tags)
	networks, err := g.gcpr.ListNetworks(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list networks from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, network := range networks {
		r := provider.NewResource(network.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeHealthCheck(ctx context.Context, g *google, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	f := initializeFilter(tags)
	checks, err := g.gcpr.ListHealthCheck(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list health checks from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, check := range checks {
		r := provider.NewResource(check.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeInstanceGroup(ctx context.Context, g *google, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	f := initializeFilter(tags)
	instanceGroups, err := g.gcpr.ListInstanceGroup(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list instance groups from reader")
	}
	resources := make([]provider.Resource, 0)
	for z, groups := range instanceGroups {
		for _, group := range groups {
			r := provider.NewResource(fmt.Sprintf("%s/%s/%s", g.Project(), z, group.Name), resourceType, g)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func computeBackendService(ctx context.Context, g *google, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	f := initializeFilter(tags)
	backends, err := g.gcpr.ListBackendService(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list backend services from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, backend := range backends {
		r := provider.NewResource(backend.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}
