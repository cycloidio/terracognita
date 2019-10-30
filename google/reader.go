package google

import (
	"context"
	"strings"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// GCPReader is the middleware between TC and GCP
type GCPReader struct {
	compute *compute.Service
	project string
	region  string
}

// NewGcpReader returns a GCPReader with a catalog of services
// ready to be used
func NewGcpReader(ctx context.Context, project, region, credentials string) (*GCPReader, error) {
	comp, err := compute.NewService(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, err
	}
	return &GCPReader{
		compute: comp,
		project: project,
		region:  region,
	}, nil
}

func (r *GCPReader) getZones() ([]string, error) {
	rs := compute.NewRegionsService(r.compute)
	region, err := rs.Get(r.project, r.region).Do()
	if err != nil {
		return nil, err
	}
	// zones are URL format, e.g:
	// https://www.googleapis.com/compute/v1/projects/my-project/zones/us-central1-c
	// Need to split them
	zones := make([]string, 0, len(region.Zones))
	for _, URL := range region.Zones {
		tmp := strings.Split(URL, "/")
		zones = append(zones, tmp[len(tmp)-1])
	}
	return zones, nil
}

// ListInstances return a list of instances in the project/region
func (r *GCPReader) ListInstances(ctx context.Context, filter string) (map[string][]compute.Instance, error) {
	is := compute.NewInstancesService(r.compute)
	instancesList := make(map[string][]compute.Instance)
	zones, err := r.getZones()
	if err != nil {
		return nil, err
	}
	for _, zone := range zones {
		// 500 is the current `maxResults`
		instances := make([]compute.Instance, 0, 500)
		if err := is.List(r.project, zone).
			Filter(filter).
			Pages(ctx, func(list *compute.InstanceList) error {
				for _, instance := range list.Items {
					instances = append(instances, *instance)
				}
				return nil
			}); err != nil {
			return nil, err
		}
		instancesList[zone] = instances
	}
	return instancesList, nil
}

// ListFirewalls return a list of firewalls in the project
func (r *GCPReader) ListFirewalls(ctx context.Context, filter string) ([]compute.Firewall, error) {
	is := compute.NewFirewallsService(r.compute)
	// 500 is the current `maxResults`
	firewalls := make([]compute.Firewall, 0, 500)
	if err := is.List(r.project).
		Filter(filter).
		Pages(ctx, func(list *compute.FirewallList) error {
			for _, firewall := range list.Items {
				firewalls = append(firewalls, *firewall)
			}
			return nil
		}); err != nil {
		return nil, err
	}
	return firewalls, nil
}
