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
	// * frontend configuration: target_http(s)_proxy + global_forwarding_rule
	ComputeHealthCheck
	ComputeInstanceGroup
	ComputeBackendService
	ComputeSSLCertificate
	ComputeTargetHTTPProxy
	ComputeTargetHTTPSProxy
	ComputeURLMap
	ComputeGlobalForwardingRule
	ComputeForwardingRule
	ComputeDisk
	StorageBucket
	SQLDatabaseInstance
)

type rtFn func(ctx context.Context, g *google, resourceType string, tags []tag.Tag) ([]provider.Resource, error)

var (
	resources = map[ResourceType]rtFn{
		ComputeInstance:             computeInstance,
		ComputeFirewall:             computeFirewall,
		ComputeNetwork:              computeNetwork,
		ComputeHealthCheck:          computeHealthCheck,
		ComputeInstanceGroup:        computeInstanceGroup,
		ComputeBackendService:       computeBackendService,
		ComputeSSLCertificate:       computeSSLCertificate,
		ComputeTargetHTTPProxy:      computeTargetHTTPProxy,
		ComputeTargetHTTPSProxy:     computeTargetHTTPSProxy,
		ComputeURLMap:               computeURLMap,
		ComputeGlobalForwardingRule: computeGlobalForwardingRule,
		ComputeForwardingRule:       computeForwardingRule,
		ComputeDisk:                 computeDisk,
		StorageBucket:               storageBucket,
		SQLDatabaseInstance:         sqlDatabaseInstance,
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
	checks, err := g.gcpr.ListHealthChecks(ctx, f)
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
	instanceGroups, err := g.gcpr.ListInstanceGroups(ctx, f)
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
	backends, err := g.gcpr.ListBackendServices(ctx, f)
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

func computeURLMap(ctx context.Context, g *google, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	f := initializeFilter(tags)
	maps, err := g.gcpr.ListURLMaps(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list URL maps from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, urlMap := range maps {
		r := provider.NewResource(urlMap.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeTargetHTTPProxy(ctx context.Context, g *google, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	f := initializeFilter(tags)
	targets, err := g.gcpr.ListTargetHTTPProxies(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list target http proxies from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, target := range targets {
		r := provider.NewResource(target.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeTargetHTTPSProxy(ctx context.Context, g *google, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	f := initializeFilter(tags)
	targets, err := g.gcpr.ListTargetHTTPSProxies(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list target https proxies from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, target := range targets {
		r := provider.NewResource(target.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeSSLCertificate(ctx context.Context, g *google, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	f := initializeFilter(tags)
	certs, err := g.gcpr.ListSSLCertificates(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list SSL certificates from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, cert := range certs {
		r := provider.NewResource(cert.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeGlobalForwardingRule(ctx context.Context, g *google, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	f := initializeFilter(tags)
	rules, err := g.gcpr.ListGlobalForwardingRules(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list global forwarding rules from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, rule := range rules {
		r := provider.NewResource(rule.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeForwardingRule(ctx context.Context, g *google, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	f := initializeFilter(tags)
	rules, err := g.gcpr.ListForwardingRules(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list global forwarding rules from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, rule := range rules {
		r := provider.NewResource(rule.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeDisk(ctx context.Context, g *google, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	f := initializeFilter(tags)
	disksList, err := g.gcpr.ListDisks(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list disks from reader")
	}
	resources := make([]provider.Resource, 0)
	for z, disks := range disksList {
		for _, disk := range disks {
			r := provider.NewResource(fmt.Sprintf("%s/%s", z, disk.Name), resourceType, g)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func storageBucket(ctx context.Context, g *google, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	rules, err := g.gcpr.ListBuckets(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list global forwarding rules from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, rule := range rules {
		r := provider.NewResource(rule.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func sqlDatabaseInstance(ctx context.Context, g *google, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	f := initializeFilter(tags)
	instances, err := g.gcpr.ListStorageInstances(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list sql storage instances rules from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, instance := range instances {
		r := provider.NewResource(instance.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}
