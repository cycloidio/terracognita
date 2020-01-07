package google

// Code generated by 'go generate'; DO NOT EDIT
import (
	"context"

	"github.com/pkg/errors"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/dns/v1"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"google.golang.org/api/storage/v1"
)

// ListInstances returns a list of Instances within a project and a zone
func (r *GCPReader) ListInstances(ctx context.Context, filter string) (map[string][]compute.Instance, error) {
	service := compute.NewInstancesService(r.compute)

	list := make(map[string][]compute.Instance)
	zones, err := r.getZones()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get zones in region")
	}
	for _, zone := range zones {

		resources := make([]compute.Instance, 0)

		if err := service.List(r.project, zone).
			Filter(filter).
			MaxResults(int64(r.maxResults)).
			Pages(ctx, func(list *compute.InstanceList) error {
				for _, res := range list.Items {
					resources = append(resources, *res)
				}
				return nil
			}); err != nil {
			return nil, errors.Wrap(err, "unable to list compute Instance from google APIs")
		}

		list[zone] = resources
	}
	return list, nil

}

// ListFirewalls returns a list of Firewalls within a project
func (r *GCPReader) ListFirewalls(ctx context.Context, filter string) ([]compute.Firewall, error) {
	service := compute.NewFirewallsService(r.compute)

	resources := make([]compute.Firewall, 0)

	if err := service.List(r.project).
		Filter(filter).
		MaxResults(int64(r.maxResults)).
		Pages(ctx, func(list *compute.FirewallList) error {
			for _, res := range list.Items {
				resources = append(resources, *res)
			}
			return nil
		}); err != nil {
		return nil, errors.Wrap(err, "unable to list compute Firewall from google APIs")
	}

	return resources, nil

}

// ListNetworks returns a list of Networks within a project
func (r *GCPReader) ListNetworks(ctx context.Context, filter string) ([]compute.Network, error) {
	service := compute.NewNetworksService(r.compute)

	resources := make([]compute.Network, 0)

	if err := service.List(r.project).
		Filter(filter).
		MaxResults(int64(r.maxResults)).
		Pages(ctx, func(list *compute.NetworkList) error {
			for _, res := range list.Items {
				resources = append(resources, *res)
			}
			return nil
		}); err != nil {
		return nil, errors.Wrap(err, "unable to list compute Network from google APIs")
	}

	return resources, nil

}

// ListInstanceGroups returns a list of InstanceGroups within a project and a zone
func (r *GCPReader) ListInstanceGroups(ctx context.Context, filter string) (map[string][]compute.InstanceGroup, error) {
	service := compute.NewInstanceGroupsService(r.compute)

	list := make(map[string][]compute.InstanceGroup)
	zones, err := r.getZones()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get zones in region")
	}
	for _, zone := range zones {

		resources := make([]compute.InstanceGroup, 0)

		if err := service.List(r.project, zone).
			Filter(filter).
			MaxResults(int64(r.maxResults)).
			Pages(ctx, func(list *compute.InstanceGroupList) error {
				for _, res := range list.Items {
					resources = append(resources, *res)
				}
				return nil
			}); err != nil {
			return nil, errors.Wrap(err, "unable to list compute InstanceGroup from google APIs")
		}

		list[zone] = resources
	}
	return list, nil

}

// ListBackendServices returns a list of BackendServices within a project
func (r *GCPReader) ListBackendServices(ctx context.Context, filter string) ([]compute.BackendService, error) {
	service := compute.NewBackendServicesService(r.compute)

	resources := make([]compute.BackendService, 0)

	if err := service.List(r.project).
		Filter(filter).
		MaxResults(int64(r.maxResults)).
		Pages(ctx, func(list *compute.BackendServiceList) error {
			for _, res := range list.Items {
				resources = append(resources, *res)
			}
			return nil
		}); err != nil {
		return nil, errors.Wrap(err, "unable to list compute BackendService from google APIs")
	}

	return resources, nil

}

// ListHealthChecks returns a list of HealthChecks within a project
func (r *GCPReader) ListHealthChecks(ctx context.Context, filter string) ([]compute.HealthCheck, error) {
	service := compute.NewHealthChecksService(r.compute)

	resources := make([]compute.HealthCheck, 0)

	if err := service.List(r.project).
		Filter(filter).
		MaxResults(int64(r.maxResults)).
		Pages(ctx, func(list *compute.HealthCheckList) error {
			for _, res := range list.Items {
				resources = append(resources, *res)
			}
			return nil
		}); err != nil {
		return nil, errors.Wrap(err, "unable to list compute HealthCheck from google APIs")
	}

	return resources, nil

}

// ListURLMaps returns a list of URLMaps within a project
func (r *GCPReader) ListURLMaps(ctx context.Context, filter string) ([]compute.UrlMap, error) {
	service := compute.NewUrlMapsService(r.compute)

	resources := make([]compute.UrlMap, 0)

	if err := service.List(r.project).
		Filter(filter).
		MaxResults(int64(r.maxResults)).
		Pages(ctx, func(list *compute.UrlMapList) error {
			for _, res := range list.Items {
				resources = append(resources, *res)
			}
			return nil
		}); err != nil {
		return nil, errors.Wrap(err, "unable to list compute UrlMap from google APIs")
	}

	return resources, nil

}

// ListTargetHTTPProxies returns a list of TargetHTTPProxies within a project
func (r *GCPReader) ListTargetHTTPProxies(ctx context.Context, filter string) ([]compute.TargetHttpProxy, error) {
	service := compute.NewTargetHttpProxiesService(r.compute)

	resources := make([]compute.TargetHttpProxy, 0)

	if err := service.List(r.project).
		Filter(filter).
		MaxResults(int64(r.maxResults)).
		Pages(ctx, func(list *compute.TargetHttpProxyList) error {
			for _, res := range list.Items {
				resources = append(resources, *res)
			}
			return nil
		}); err != nil {
		return nil, errors.Wrap(err, "unable to list compute TargetHttpProxy from google APIs")
	}

	return resources, nil

}

// ListTargetHTTPSProxies returns a list of TargetHTTPSProxies within a project
func (r *GCPReader) ListTargetHTTPSProxies(ctx context.Context, filter string) ([]compute.TargetHttpsProxy, error) {
	service := compute.NewTargetHttpsProxiesService(r.compute)

	resources := make([]compute.TargetHttpsProxy, 0)

	if err := service.List(r.project).
		Filter(filter).
		MaxResults(int64(r.maxResults)).
		Pages(ctx, func(list *compute.TargetHttpsProxyList) error {
			for _, res := range list.Items {
				resources = append(resources, *res)
			}
			return nil
		}); err != nil {
		return nil, errors.Wrap(err, "unable to list compute TargetHttpsProxy from google APIs")
	}

	return resources, nil

}

// ListSSLCertificates returns a list of SSLCertificates within a project
func (r *GCPReader) ListSSLCertificates(ctx context.Context, filter string) ([]compute.SslCertificate, error) {
	service := compute.NewSslCertificatesService(r.compute)

	resources := make([]compute.SslCertificate, 0)

	if err := service.List(r.project).
		Filter(filter).
		MaxResults(int64(r.maxResults)).
		Pages(ctx, func(list *compute.SslCertificateList) error {
			for _, res := range list.Items {
				resources = append(resources, *res)
			}
			return nil
		}); err != nil {
		return nil, errors.Wrap(err, "unable to list compute SslCertificate from google APIs")
	}

	return resources, nil

}

// ListGlobalForwardingRules returns a list of GlobalForwardingRules within a project
func (r *GCPReader) ListGlobalForwardingRules(ctx context.Context, filter string) ([]compute.ForwardingRule, error) {
	service := compute.NewGlobalForwardingRulesService(r.compute)

	resources := make([]compute.ForwardingRule, 0)

	if err := service.List(r.project).
		Filter(filter).
		MaxResults(int64(r.maxResults)).
		Pages(ctx, func(list *compute.ForwardingRuleList) error {
			for _, res := range list.Items {
				resources = append(resources, *res)
			}
			return nil
		}); err != nil {
		return nil, errors.Wrap(err, "unable to list compute ForwardingRule from google APIs")
	}

	return resources, nil

}

// ListForwardingRules returns a list of ForwardingRules within a project
func (r *GCPReader) ListForwardingRules(ctx context.Context, filter string) ([]compute.ForwardingRule, error) {
	service := compute.NewForwardingRulesService(r.compute)

	resources := make([]compute.ForwardingRule, 0)

	if err := service.List(r.project, r.region).
		Filter(filter).
		MaxResults(int64(r.maxResults)).
		Pages(ctx, func(list *compute.ForwardingRuleList) error {
			for _, res := range list.Items {
				resources = append(resources, *res)
			}
			return nil
		}); err != nil {
		return nil, errors.Wrap(err, "unable to list compute ForwardingRule from google APIs")
	}

	return resources, nil

}

// ListDisks returns a list of Disks within a project and a zone
func (r *GCPReader) ListDisks(ctx context.Context, filter string) (map[string][]compute.Disk, error) {
	service := compute.NewDisksService(r.compute)

	list := make(map[string][]compute.Disk)
	zones, err := r.getZones()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get zones in region")
	}
	for _, zone := range zones {

		resources := make([]compute.Disk, 0)

		if err := service.List(r.project, zone).
			Filter(filter).
			MaxResults(int64(r.maxResults)).
			Pages(ctx, func(list *compute.DiskList) error {
				for _, res := range list.Items {
					resources = append(resources, *res)
				}
				return nil
			}); err != nil {
			return nil, errors.Wrap(err, "unable to list compute Disk from google APIs")
		}

		list[zone] = resources
	}
	return list, nil

}

// ListBuckets returns a list of Buckets within a project
func (r *GCPReader) ListBuckets(ctx context.Context) ([]storage.Bucket, error) {
	service := storage.NewBucketsService(r.storage)

	resources := make([]storage.Bucket, 0)

	if err := service.List(r.project).
		MaxResults(int64(r.maxResults)).
		Pages(ctx, func(list *storage.Buckets) error {
			for _, res := range list.Items {
				resources = append(resources, *res)
			}
			return nil
		}); err != nil {
		return nil, errors.Wrap(err, "unable to list storage Bucket from google APIs")
	}

	return resources, nil

}

// ListStorageInstances returns a list of StorageInstances within a project
func (r *GCPReader) ListStorageInstances(ctx context.Context, filter string) ([]sqladmin.DatabaseInstance, error) {
	service := sqladmin.NewInstancesService(r.sqladmin)

	resources := make([]sqladmin.DatabaseInstance, 0)

	if err := service.List(r.project).
		Filter(filter).
		MaxResults(int64(r.maxResults)).
		Pages(ctx, func(list *sqladmin.InstancesListResponse) error {
			for _, res := range list.Items {
				resources = append(resources, *res)
			}
			return nil
		}); err != nil {
		return nil, errors.Wrap(err, "unable to list sqladmin DatabaseInstance from google APIs")
	}

	return resources, nil

}

// ListManagedZones returns a list of ManagedZones within a project
func (r *GCPReader) ListManagedZones(ctx context.Context) ([]dns.ManagedZone, error) {
	service := dns.NewManagedZonesService(r.dns)

	resources := make([]dns.ManagedZone, 0)

	if err := service.List(r.project).
		MaxResults(int64(r.maxResults)).
		Pages(ctx, func(list *dns.ManagedZonesListResponse) error {
			for _, res := range list.ManagedZones {
				resources = append(resources, *res)
			}
			return nil
		}); err != nil {
		return nil, errors.Wrap(err, "unable to list dns ManagedZone from google APIs")
	}

	return resources, nil

}
