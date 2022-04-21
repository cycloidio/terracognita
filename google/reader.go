package google

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/file/v1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/logging/v2"
	"google.golang.org/api/monitoring/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/redis/v1"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"google.golang.org/api/storage/v1"
)

//go:generate go run ./cmd

// GCPReader is the middleware between TC and GCP
type GCPReader struct {
	compute      *compute.Service
	storage      *storage.Service
	sqladmin     *sqladmin.Service
	dns          *dns.Service
	iam          *iam.Service
	cloudbilling *cloudbilling.APIService
	file         *file.Service
	container    *container.Service
	redis        *redis.Service
	logging      *logging.Service
	monitoring   *monitoring.Service
	project      string
	region       string
	zones        []string
	maxResults   uint64
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
	sql, err := sqladmin.NewService(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create sqladmin service")
	}
	d, err := dns.NewService(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create sqladmin service")
	}
	i, err := iam.NewService(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create iam service")
	}
	bill, err := cloudbilling.NewService(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create cloud billing service")
	}
	file, err := file.NewService(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create filestore service")
	}
	container, err := container.NewService(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create container service")
	}
	redis, err := redis.NewService(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create redis service")
	}
	logging, err := logging.NewService(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create logging service")
	}
	monitoring, err := monitoring.NewService(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create logging service")
	}
	return &GCPReader{
		compute:      comp,
		storage:      storage,
		sqladmin:     sql,
		project:      project,
		region:       region,
		dns:          d,
		iam:          i,
		cloudbilling: bill,
		file:         file,
		container:    container,
		redis:        redis,
		logging:      logging,
		monitoring:   monitoring,
		zones:        []string{},
		maxResults:   maxResults,
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
