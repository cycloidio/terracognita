package google

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// GCPReader is the middleware between TC and GCP
type GCPReader struct {
	compute    *compute.Service
	project    string
	region     string
	zones      []string
	maxResults uint64
}

// NewGcpReader returns a GCPReader with a catalog of services
// ready to be used
func NewGcpReader(ctx context.Context, maxResults uint64, project, region, credentials string) (*GCPReader, error) {
	if maxResults > 500 {
		return nil, errors.New("max-results must be between 0 and 500, inclusive")
	}
	comp, err := compute.NewService(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create compute service")
	}
	return &GCPReader{
		compute:    comp,
		project:    project,
		region:     region,
		zones:      []string{},
		maxResults: maxResults,
	}, nil
}

func (r *GCPReader) getZones() ([]string, error) {
	if len(r.zones) > 0 {
		return r.zones, nil
	}
	rs := compute.NewRegionsService(r.compute)
	region, err := rs.Get(r.project, r.region).Do()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to fetch information for region %s", r.region)
	}
	// zones are URL format, e.g:
	// https://www.googleapis.com/compute/v1/projects/my-project/zones/us-central1-c
	// Need to split them
	zones := make([]string, 0, len(region.Zones))
	for _, URL := range region.Zones {
		tmp := strings.Split(URL, "/")
		zones = append(zones, tmp[len(tmp)-1])
	}
	r.zones = zones
	return zones, nil
}

// ListInstances return a list of instances in the project/region
func (r *GCPReader) ListInstances(ctx context.Context, filter string) (map[string][]compute.Instance, error) {
	is := compute.NewInstancesService(r.compute)
	instancesList := make(map[string][]compute.Instance)
	zones, err := r.getZones()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get zones in region")
	}
	for _, zone := range zones {
		instances := make([]compute.Instance, 0)
		if err := is.List(r.project, zone).
			Filter(filter).
			MaxResults(int64(r.maxResults)).
			Pages(ctx, func(list *compute.InstanceList) error {
				for _, instance := range list.Items {
					instances = append(instances, *instance)
				}
				return nil
			}); err != nil {
			return nil, errors.Wrap(err, "unable to list compute instance from google APIs")
		}
		instancesList[zone] = instances
	}
	return instancesList, nil
}

// ListFirewalls return a list of firewalls in the project
func (r *GCPReader) ListFirewalls(ctx context.Context, filter string) ([]compute.Firewall, error) {
	is := compute.NewFirewallsService(r.compute)
	firewalls := make([]compute.Firewall, 0)
	if err := is.List(r.project).
		Filter(filter).
		MaxResults(int64(r.maxResults)).
		Pages(ctx, func(list *compute.FirewallList) error {
			for _, firewall := range list.Items {
				firewalls = append(firewalls, *firewall)
			}
			return nil
		}); err != nil {
		return nil, errors.Wrap(err, "unable to list firewalls from google APIs")
	}
	return firewalls, nil
}

// ListNetworks return a list of networks in the project
func (r *GCPReader) ListNetworks(ctx context.Context, filter string) ([]compute.Network, error) {
	ns := compute.NewNetworksService(r.compute)
	networks := make([]compute.Network, 0)
	if err := ns.List(r.project).
		Filter(filter).
		MaxResults(int64(r.maxResults)).
		Pages(ctx, func(list *compute.NetworkList) error {
			for _, network := range list.Items {
				networks = append(networks, *network)
			}
			return nil
		}); err != nil {
		return nil, errors.Wrap(err, "unable to list networks from google APIs")
	}
	return networks, nil
}

// ListHealthCheck returns a list of health checks in the project
func (r *GCPReader) ListHealthCheck(ctx context.Context, filter string) ([]compute.HealthCheck, error) {
	ns := compute.NewHealthChecksService(r.compute)
	checks := make([]compute.HealthCheck, 0)
	if err := ns.List(r.project).
		Filter(filter).
		MaxResults(int64(r.maxResults)).
		Pages(ctx, func(list *compute.HealthCheckList) error {
			for _, check := range list.Items {
				checks = append(checks, *check)
			}
			return nil
		}); err != nil {
		return nil, errors.Wrap(err, "unable to list health checks from google APIs")
	}
	return checks, nil
}

// ListInstanceGroup returns a list of instance groups in the project within a zone
func (r *GCPReader) ListInstanceGroup(ctx context.Context, filter string) (map[string][]compute.InstanceGroup, error) {
	igs := compute.NewInstanceGroupsService(r.compute)
	instanceGroupList := make(map[string][]compute.InstanceGroup)
	zones, err := r.getZones()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get zones in region")
	}
	for _, zone := range zones {
		groups := make([]compute.InstanceGroup, 0)
		if err := igs.List(r.project, zone).
			Filter(filter).
			MaxResults(int64(r.maxResults)).
			Pages(ctx, func(list *compute.InstanceGroupList) error {
				for _, group := range list.Items {
					groups = append(groups, *group)
				}
				return nil
			}); err != nil {
			return nil, errors.Wrap(err, "unable to list instance groups from google APIs")
		}
		instanceGroupList[zone] = groups
	}
	return instanceGroupList, nil
}

// ListBackendService returns a list of backend service in the project
func (r *GCPReader) ListBackendService(ctx context.Context, filter string) ([]compute.BackendService, error) {
	bs := compute.NewBackendServicesService(r.compute)
	backends := make([]compute.BackendService, 0)
	if err := bs.List(r.project).
		Filter(filter).
		MaxResults(int64(r.maxResults)).
		Pages(ctx, func(list *compute.BackendServiceList) error {
			for _, backend := range list.Items {
				backends = append(backends, *backend)
			}
			return nil
		}); err != nil {
		return nil, errors.Wrap(err, "unable to list backend services from google APIs")
	}
	return backends, nil
}
