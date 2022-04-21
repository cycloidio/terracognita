package google

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/provider"
	"github.com/pkg/errors"
)

// ResourceType is the type used to define all the Resources
// from the Provider
type ResourceType int

//go:generate enumer -type ResourceType -addprefix google_ -transform snake -linecomment
const (
	// compute engine
	ComputeInstance ResourceType = iota
	ComputeFirewall
	ComputeNetwork
	ComputeHealthCheck
	ComputeInstanceGroup
	ComputeInstanceIAMPolicy
	ComputeBackendBucket
	ComputeBackendService
	ComputeSSLCertificate
	ComputeTargetHTTPProxy
	ComputeTargetHTTPSProxy
	ComputeURLMap
	ComputeGlobalForwardingRule
	ComputeForwardingRule
	ComputeDisk
	ComputeAddress
	ComputeAttachedDisk
	ComputeAutoscaler
	ComputeGlobalAddress
	ComputeImage
	ComputeInstanceGroupManager
	ComputeInstanceTemplate
	ComputeManagedSSLCertificate
	ComputeNetworkEndpointGroup
	ComputeRoute
	ComputeSecurityPolicy
	ComputeServiceAttachment
	ComputeSnapshot
	ComputeSSLPolicy
	ComputeSubnetwork
	ComputeTargetGRPCProxy
	ComputeTargetInstance
	ComputeTargetPool
	ComputeTargetSSLProxy
	ComputeTargetTCPProxy
	// cloud dns
	DNSManagedZone
	DNSRecordSet
	DNSPolicy
	// cloud platform
	ProjectIAMCustomRole
	BillingSubaccount
	// cloud sql
	SQLDatabaseInstance
	SQLDatabase
	// cloud storage
	StorageBucket
	StorageBucketIAMPolicy
	// filestore
	FilestoreInstance
	// k8s container engine
	ContainerCluster
	ContainerNodePool
	// memorystore (redis)
	RedisInstance
	// cloud (Stackdriver) Logging
	LoggingMetric
	// cloud (Stackdriver) Monitoring
	MonitoringAlertPolicy
	MonitoringGroup
	MonitoringNotificationChannel
	MonitoringUptimeCheckConfig

	noFilter = ""
)

type rtFn func(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error)

var (
	resources = map[ResourceType]rtFn{
		// compute engine
		ComputeInstance:              computeInstance,
		ComputeFirewall:              computeFirewall,
		ComputeNetwork:               computeNetwork,
		ComputeHealthCheck:           computeHealthCheck,
		ComputeInstanceGroup:         computeInstanceGroup,
		ComputeInstanceIAMPolicy:     computeInstanceIAMPolicy,
		ComputeBackendService:        computeBackendService,
		ComputeBackendBucket:         computeBackendBucket,
		ComputeSSLCertificate:        computeSSLCertificate,
		ComputeTargetHTTPProxy:       computeTargetHTTPProxy,
		ComputeTargetHTTPSProxy:      computeTargetHTTPSProxy,
		ComputeURLMap:                computeURLMap,
		ComputeGlobalForwardingRule:  computeGlobalForwardingRule,
		ComputeForwardingRule:        computeForwardingRule,
		ComputeDisk:                  computeDisk,
		ComputeAddress:               computeAddress,
		ComputeAttachedDisk:          computeAttachedDisk,
		ComputeAutoscaler:            computeAutoscaler,
		ComputeGlobalAddress:         computeGlobalAddress,
		ComputeImage:                 computeImage,
		ComputeInstanceGroupManager:  computeInstanceGroupManager,
		ComputeInstanceTemplate:      computeInstanceTemplate,
		ComputeManagedSSLCertificate: computeManagedSSLCertificate,
		ComputeNetworkEndpointGroup:  computeNetworkEndpointGroup,
		ComputeRoute:                 computeRoute,
		ComputeSecurityPolicy:        computeSecurityPolicy,
		ComputeServiceAttachment:     computeServiceAttachment,
		ComputeSnapshot:              computeSnapshot,
		ComputeSSLPolicy:             computeSSLPolicy,
		ComputeSubnetwork:            computeSubnetwork,
		ComputeTargetGRPCProxy:       computeTargetGRPCProxy,
		ComputeTargetInstance:        computeTargetInstance,
		ComputeTargetPool:            computeTargetPool,
		ComputeTargetSSLProxy:        computeTargetSSLProxy,
		ComputeTargetTCPProxy:        computeTargetTCPProxy,

		// cloud dns
		DNSManagedZone: dnsManagedZone,
		DNSRecordSet:   dnsRecordSet,
		DNSPolicy:      dnsPolicy,
		// cloud platform
		ProjectIAMCustomRole: projectIAMCustomRole,
		BillingSubaccount:    billingSubaccount,
		// cloud sql
		SQLDatabaseInstance: sqlDatabaseInstance,
		SQLDatabase:         sqlDatabase,
		// cloud storage
		StorageBucket:          storageBucket,
		StorageBucketIAMPolicy: storageBucketIAMPolicy,
		// filestore
		FilestoreInstance: filestoreInstance,
		// k8s container engine
		ContainerCluster:  containerCluster,
		ContainerNodePool: containerNodePool,
		// memorystore (redis)
		RedisInstance: redisInstance,
		// cloud (Stackdriver) Logging
		LoggingMetric: loggingMetric,
		// cloud (Stackdriver) Monitoring
		MonitoringAlertPolicy:         monitoringAlertPolicy,
		MonitoringGroup:               monitoringGroup,
		MonitoringNotificationChannel: monitoringNotificationChannel,
		MonitoringUptimeCheckConfig:   monitoringUptimeCheckConfig,
	}
)

func initializeFilter(filters *filter.Filter) string {
	var b bytes.Buffer
	for _, t := range filters.Tags {
		// if multiple tags, we suppose it's a "AND" operation
		b.WriteString(fmt.Sprintf("(labels.%s=%s) ", t.Name, t.Value))
	}
	return b.String()
}

//compute

func computeInstance(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	f := initializeFilter(filters)
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

func computeFirewall(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	firewalls, err := g.gcpr.ListFirewalls(ctx, noFilter)
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

func computeNetwork(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	networks, err := g.gcpr.ListNetworks(ctx, noFilter)
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

func computeHealthCheck(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	checks, err := g.gcpr.ListHealthChecks(ctx, noFilter)
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

func computeInstanceGroup(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	instanceGroups, err := g.gcpr.ListInstanceGroups(ctx, noFilter)
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

func computeBackendService(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	backends, err := g.gcpr.ListBackendServices(ctx, noFilter)
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

func computeURLMap(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	maps, err := g.gcpr.ListURLMaps(ctx, noFilter)
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

func computeTargetHTTPProxy(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	targets, err := g.gcpr.ListTargetHTTPProxies(ctx, noFilter)
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

func computeTargetHTTPSProxy(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	targets, err := g.gcpr.ListTargetHTTPSProxies(ctx, noFilter)
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

func computeSSLCertificate(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	certs, err := g.gcpr.ListSSLCertificates(ctx, noFilter)
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

func computeGlobalForwardingRule(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	f := initializeFilter(filters)
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

func computeForwardingRule(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	f := initializeFilter(filters)
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

func computeDisk(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	f := initializeFilter(filters)
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

// computeDiskIAMPolicy will import the policies binded to a disk. We need to iterate over the
// disk list
func computeDiskIAMPolicy(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	disksList, err := g.gcpr.ListDisks(ctx, noFilter)
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

func computeBackendBucket(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	backends, err := g.gcpr.ListBackendBuckets(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list backend buckets from reader")
	}
	resources := make([]provider.Resource, 0, len(backends))
	for _, backend := range backends {
		r := provider.NewResource(backend.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

// computeInstanceIAMPolicy will import the policies binded to a compute instance. We need to iterate over the
// compute instance list
func computeInstanceIAMPolicy(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	f := initializeFilter(filters)
	list, err := g.gcpr.ListInstances(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list compute instances from reader")
	}
	resources := make([]provider.Resource, 0)
	for zone, instances := range list {
		for _, instance := range instances {
			r := provider.NewResource(fmt.Sprintf("projects/%s/zones/%s/instances/%s", g.Project(), zone, instance.Name), resourceType, g)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func computeAddress(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	addresses, err := g.gcpr.ListAddresses(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list adresses from reader")
	}
	resources := make([]provider.Resource, 0, len(addresses))
	for _, address := range addresses {
		r := provider.NewResource(address.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeAttachedDisk(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {

	// if google_compute_instance defined, attachment are done by attached_disk block.
	if filters.IsIncluded("google_compute_instance") && !filters.IsExcluded("google_compute_instance") {
		return nil, nil
	}

	instancesList, err := g.gcpr.ListInstances(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list attached disks from reader")
	}
	resources := make([]provider.Resource, 0)
	for zone, instances := range instancesList {
		for _, instance := range instances {
			for _, disk := range instance.Disks {
				r := provider.NewResource(fmt.Sprintf("%s/%s/%s/%s", g.Project(), zone, instance.Name, disk.DeviceName), resourceType, g)
				resources = append(resources, r)
			}
		}
	}
	return resources, nil
}

func computeAutoscaler(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	f := initializeFilter(filters)
	autoscalersList, err := g.gcpr.ListAutoscalers(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list autoscalers from reader")
	}
	resources := make([]provider.Resource, 0)
	for zone, autoscalers := range autoscalersList {
		for _, autoscaler := range autoscalers {
			r := provider.NewResource(fmt.Sprintf("%s/%s", zone, autoscaler.Name), resourceType, g)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func computeGlobalAddress(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	globalAddresses, err := g.gcpr.ListGlobalAddresses(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list global addresses from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, globalAddress := range globalAddresses {
		r := provider.NewResource(globalAddress.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeImage(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	images, err := g.gcpr.ListImages(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list compute images from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, image := range images {
		r := provider.NewResource(image.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeInstanceGroupManager(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	computeInstanceGroupManagersList, err := g.gcpr.ListInstanceGroupManagers(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list instance group manager from reader")
	}
	resources := make([]provider.Resource, 0)
	for zone, computeInstanceGroupManagers := range computeInstanceGroupManagersList {
		for _, computeInstanceGroupManager := range computeInstanceGroupManagers {
			computeInstanceGroupManager.Zone = zone
			r := provider.NewResource(fmt.Sprintf("%s/%s/%s", g.Project(), zone, computeInstanceGroupManager.Name), resourceType, g)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func computeInstanceTemplate(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	f := initializeFilter(filters)
	instanceTemplates, err := g.gcpr.ListInstanceTemplates(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list instance template from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, instanceTemplate := range instanceTemplates {
		r := provider.NewResource(instanceTemplate.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeManagedSSLCertificate(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	managedSSLCertificates, err := g.gcpr.ListManagedSslCertificates(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list managed ssl certificates from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, managedSSLCertificate := range managedSSLCertificates {
		r := provider.NewResource(managedSSLCertificate.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeNetworkEndpointGroup(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	networkEndpointGroupList, err := g.gcpr.ListNetworkEndpointGroups(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list network endpoint groups from reader")
	}
	resources := make([]provider.Resource, 0)
	for zone, networkEndpointGroups := range networkEndpointGroupList {
		for _, networkEndpointGroup := range networkEndpointGroups {
			r := provider.NewResource(fmt.Sprintf("%s/%s/%s", g.Project(), zone, networkEndpointGroup.Name), resourceType, g)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func computeRoute(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	f := initializeFilter(filters)
	computeRoutes, err := g.gcpr.ListRoutes(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list routes from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, computeRoute := range computeRoutes {
		r := provider.NewResource(computeRoute.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeSecurityPolicy(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	securityPolicies, err := g.gcpr.ListSecurityPolicies(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list security policies from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, securityPolicy := range securityPolicies {
		r := provider.NewResource(securityPolicy.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeServiceAttachment(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	serviceAttachments, err := g.gcpr.ListServiceAttachments(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list service attachments from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, serviceAttachment := range serviceAttachments {
		r := provider.NewResource(serviceAttachment.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeSnapshot(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	snapshots, err := g.gcpr.ListSnapshots(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list snapshots from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, snapshot := range snapshots {
		r := provider.NewResource(snapshot.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeSSLPolicy(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sslpolicies, err := g.gcpr.ListSslPolicies(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list ssl policies from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, sslpolicy := range sslpolicies {
		r := provider.NewResource(sslpolicy.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeSubnetwork(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	subnetworks, err := g.gcpr.ListSubnetworks(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list subnetworks from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, subnetwork := range subnetworks {
		r := provider.NewResource(subnetwork.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeTargetGRPCProxy(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	grpcProxies, err := g.gcpr.ListTargetGrpcProxies(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list target grpc proxies from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, grpcProxy := range grpcProxies {
		r := provider.NewResource(grpcProxy.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeTargetInstance(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	targetInstancesList, err := g.gcpr.ListTargetInstances(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list target instances from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, targetInstances := range targetInstancesList {
		for _, targetInstance := range targetInstances {
			r := provider.NewResource(targetInstance.Name, resourceType, g)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func computeTargetPool(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	targetPools, err := g.gcpr.ListTargetPools(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list target pools from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, targetPool := range targetPools {
		r := provider.NewResource(targetPool.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeTargetSSLProxy(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	targetSSLProxies, err := g.gcpr.ListTargetSslProxies(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list target ssl proxies from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, targetSSLProxy := range targetSSLProxies {
		r := provider.NewResource(targetSSLProxy.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func computeTargetTCPProxy(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	targetTCPProxies, err := g.gcpr.ListTargetTCPProxies(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list target tcp proxies from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, targetTCPProxy := range targetTCPProxies {
		r := provider.NewResource(targetTCPProxy.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

//cloud dns
func dnsManagedZone(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	zones, err := g.gcpr.ListDNSManagedZones(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list DNS managed zone from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, zone := range zones {
		r := provider.NewResource(zone.Name, resourceType, g)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", zone.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the managed zone '%s'", zone.Name)
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func dnsRecordSet(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	managedZones, err := getDNSManagedZones(ctx, g, DNSManagedZone.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list previously fetch managed zones")
	}
	rrsetsList, err := g.gcpr.ListDNSResourceRecordSets(ctx, managedZones)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list resources record se record set from reader")
	}
	resources := make([]provider.Resource, 0)
	for z, rrsets := range rrsetsList {
		for _, rrset := range rrsets {
			r := provider.NewResource(fmt.Sprintf("%s/%s/%s", z, rrset.Name, rrset.Type), resourceType, g)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func dnsPolicy(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	policies, err := g.gcpr.ListDNSPolicies(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list DNS policies from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, policy := range policies {
		r := provider.NewResource(policy.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

//cloud platform

func projectIAMCustomRole(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	roles, err := g.gcpr.ListProjectIAMCustomRoles(ctx, fmt.Sprintf("projects/%s", g.Project()))
	if err != nil {
		return nil, errors.Wrap(err, "unable to list project IAM custom roles from reader")
	}
	resources := make([]provider.Resource, 0, len(roles))
	for _, role := range roles {
		r := provider.NewResource(role.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func billingSubaccount(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	billingSubaccounts, err := g.gcpr.ListBillingSubaccounts(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list project IAM custom roles from reader")
	}
	resources := make([]provider.Resource, 0, len(billingSubaccounts))
	for _, billingSubaccount := range billingSubaccounts {
		r := provider.NewResource(billingSubaccount.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

//cloud sql

func sqlDatabaseInstance(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	instances, err := g.gcpr.ListSQLDatabaseInstances(ctx, noFilter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list sql storage instances rules from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, instance := range instances {
		r := provider.NewResource(instance.Name, resourceType, g)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		// On_premise instances not stored
		if instance.InstanceType != "ON_PREMISES_INSTANCE" {
			if err := r.Data().Set("name", instance.Name); err != nil {
				return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the sql database instance '%s'", instance.Name)
			}
		}
		resources = append(resources, r)
	}
	return resources, nil
}

//sqlDatabase
func sqlDatabase(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sqlDatabaseInstances, err := getSQLDatabaseInstances(ctx, g, SQLDatabaseInstance.String(), filters)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list previously fetch sql database instances")
	}
	sqlDatabasesList, err := g.gcpr.ListSQLDatabases(ctx, noFilter, sqlDatabaseInstances)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list resources sql databases from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, sqlDatabases := range sqlDatabasesList {
		for _, sqlDatabase := range sqlDatabases {
			r := provider.NewResource(fmt.Sprintf("projects/%s/instances/%s/databases/%s", g.Project(), sqlDatabase.Instance, sqlDatabase.Name), resourceType, g)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

// cloud storage

func storageBucket(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	buckets, err := g.gcpr.ListSTORAGEBuckets(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list storage buckets from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, bucket := range buckets {
		r := provider.NewResource(bucket.Name, resourceType, g)
		// we set the name prior of reading it from the state
		// as it is required to able to List resources depending on this one
		if err := r.Data().Set("name", bucket.Name); err != nil {
			return nil, errors.Wrapf(err, "unable to set name data on the provider.Resource for the bucket '%s'", bucket.Name)
		}
		resources = append(resources, r)
	}

	return resources, nil
}

// storageBucketIAMPolicy will import the policies binded to a bucket. We need to iterate over the
// bucket list
func storageBucketIAMPolicy(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	buckets, err := g.gcpr.ListSTORAGEBuckets(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list bucket policies custom roles from reader")
	}
	resources := make([]provider.Resource, 0, len(buckets))
	for _, bucket := range buckets {
		r := provider.NewResource(bucket.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

// filestore
func filestoreInstance(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	// To retrieve instance information for all locations, use "-" for the `{location}` value.
	instances, err := g.gcpr.ListFilestoreInstances(ctx, noFilter, fmt.Sprintf("projects/%s/locations/%s", g.Project(), g.Region()))
	if err != nil {
		return nil, errors.Wrap(err, "unable to list filestore instances from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, instance := range instances {
		r := provider.NewResource(fmt.Sprintf("%s/%s", g.Region(), instance.Name), resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

// k8s container engine
func containerCluster(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	f := initializeFilter(filters)
	// To retrieve instance information for all locations, use "-" for the `{location}` value.
	clusters, err := g.gcpr.ListCONTAINERClusters(ctx, f, fmt.Sprintf("projects/%s/locations/%s", g.Project(), g.Region()))
	if err != nil {
		return nil, errors.Wrap(err, "unable to list kubernetes clusters from reader")
	}

	resources := make([]provider.Resource, 0)
	for _, cluster := range clusters {
		r := provider.NewResource(fmt.Sprintf("%s/%s", g.Region(), cluster.Name), resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func containerNodePool(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {

	// if google_container_cluster defined, node pool are managed by google_container_cluster block.
	if filters.IsIncluded("google_container_cluster") && !filters.IsExcluded("google_container_cluster") {
		return nil, nil
	}

	f := initializeFilter(filters)
	// To retrieve instance information for all locations, use "-" for the `{location}` value.
	clusters, err := g.gcpr.ListCONTAINERClusters(ctx, f, fmt.Sprintf("projects/%s/locations/%s", g.Project(), g.Region()))
	if err != nil {
		return nil, errors.Wrap(err, "unable to list kubernetes clusters from reader")
	}

	resources := make([]provider.Resource, 0)
	for _, cluster := range clusters {
		for _, node := range cluster.NodePools {
			r := provider.NewResource(fmt.Sprintf("%s/%s/%s", cluster.Location, cluster.Name, node.Name), resourceType, g)
			resources = append(resources, r)
		}
	}
	return resources, nil
}

// memorystore (redis)

func redisInstance(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	instances, err := g.gcpr.ListRedisInstances(ctx, fmt.Sprintf("projects/%s/locations/%s", g.Project(), g.Region()))
	if err != nil {
		return nil, errors.Wrap(err, "unable to list redis instances from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, instance := range instances {
		r := provider.NewResource(instance.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

// cloud (Stackdriver) Logging
func loggingMetric(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	logMetrics, err := g.gcpr.ListLogMetrics(ctx, fmt.Sprintf("projects/%s", g.Project()))
	if err != nil {
		return nil, errors.Wrap(err, "unable to list logging metrics from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, logMetric := range logMetrics {
		r := provider.NewResource(logMetric.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

// cloud (Stackdriver) Monitoring
func monitoringAlertPolicy(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	alertPolicies, err := g.gcpr.ListMONITORINGAlertPolicies(ctx, noFilter, fmt.Sprintf("projects/%s", g.Project()))
	if err != nil {
		return nil, errors.Wrap(err, "unable to list monitoring alert policies from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, alertPolicy := range alertPolicies {
		r := provider.NewResource(alertPolicy.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func monitoringGroup(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	groups, err := g.gcpr.ListMONITORINGGroups(ctx, fmt.Sprintf("projects/%s", g.Project()))
	if err != nil {
		return nil, errors.Wrap(err, "unable to list monitoring groups from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, group := range groups {
		r := provider.NewResource(group.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func monitoringNotificationChannel(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	notifChannels, err := g.gcpr.ListMONITORINGNotificationChannels(ctx, fmt.Sprintf("projects/%s", g.Project()))
	if err != nil {
		return nil, errors.Wrap(err, "unable to list monitoring notification channels from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, notifChannel := range notifChannels {
		r := provider.NewResource(notifChannel.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}

func monitoringUptimeCheckConfig(ctx context.Context, g *google, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	uptimeConfigs, err := g.gcpr.ListMONITORINGUptimeCheckConfigs(ctx, fmt.Sprintf("projects/%s", g.Project()))
	if err != nil {
		return nil, errors.Wrap(err, "unable to list monitoring uptime check configs from reader")
	}
	resources := make([]provider.Resource, 0)
	for _, uptimeConfig := range uptimeConfigs {
		r := provider.NewResource(uptimeConfig.Name, resourceType, g)
		resources = append(resources, r)
	}
	return resources, nil
}
