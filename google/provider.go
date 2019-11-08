package google

import (
	"context"
	"fmt"

	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/provider"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	tfgoogle "github.com/terraform-providers/terraform-provider-google/google"
)

type google struct {
	tfGoogleClient interface{}
	tfProvider     *schema.Provider
	gcpr           *GCPReader
}

// NewProvider returns a Gooogle Provider
func NewProvider(ctx context.Context, project, region, credentials string) (provider.Provider, error) {
	cfg := tfgoogle.Config{
		Credentials: credentials,
		Project:     project,
		Region:      region,
	}

	tfgoogle.ConfigureBasePaths(&cfg)
	log.Get().Log("func", "google.NewProvider", "msg", "loading TF client")
	if err := cfg.LoadAndValidate(); err != nil {
		return nil, fmt.Errorf("could not initialize 'terraform/google.Config.LoadAndValidate()' because: %s", err)
	}

	tfp := tfgoogle.Provider().(*schema.Provider)
	tfp.SetMeta(cfg)

	log.Get().Log("func", "google.NewProvider", "msg", "loading GCP client")
	reader, err := NewGcpReader(ctx, project, region, credentials)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize GCPReader: %v", err)
	}

	return &google{
		tfGoogleClient: &cfg,
		tfProvider:     tfp,
		gcpr:           reader,
	}, nil
}

func (g *google) HasResourceType(t string) bool {
	_, err := ResourceTypeString(t)
	return err == nil
}

func (g *google) Region() string  { return g.tfGoogleClient.(*tfgoogle.Config).Region }
func (g *google) Project() string { return g.tfGoogleClient.(*tfgoogle.Config).Project }
func (g *google) String() string  { return "google" }
func (g *google) TagKey() string  { return "labels" }

func (g *google) ResourceTypes() []string {
	return ResourceTypeStrings()
}

func (g *google) Resources(ctx context.Context, t string, f *filter.Filter) ([]provider.Resource, error) {
	rt, err := ResourceTypeString(t)
	if err != nil {
		return nil, err
	}

	rfn, ok := resources[rt]
	if !ok {
		return nil, errors.Errorf("the resource %q it's not implemented", t)
	}

	resources, err := rfn(ctx, g, t, f.Tags)
	if err != nil {
		return nil, errors.Wrapf(err, "error while reading from resource %q", t)
	}

	return resources, nil
}

func (g *google) TFClient() interface{} {
	return g.tfGoogleClient
}

func (g *google) TFProvider() *schema.Provider {
	return g.tfProvider
}
