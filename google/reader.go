package google

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/storage/v1"
)

//go:generate go run ./cmd

// GCPReader is the middleware between TC and GCP
type GCPReader struct {
	compute    *compute.Service
	storage    *storage.Service
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
	storage, err := storage.NewService(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create storage service")
	}
	return &GCPReader{
		compute:    comp,
		storage:    storage,
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
