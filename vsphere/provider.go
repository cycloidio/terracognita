package vsphere

import (
	"context"

	"fmt"

	"github.com/cycloidio/terracognita/cache"
	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/provider"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfvsphere "github.com/hashicorp/terraform-provider-vsphere/vsphere"
	"github.com/pkg/errors"
)

// version of the Terraform provider, this is automatically changed with the 'make update-terraform-provider'
const version = "2.2.0"

type vsphere struct {
	tfVSphereClient interface{}
	tfProvider      *schema.Provider

	configuration map[string]interface{}

	cache  cache.Cache
	reader *reader
}

// NewProvider returns an vSphere Provider
func NewProvider(ctx context.Context, soapURL, user, password, vsphereserver string, insecure bool) (provider.Provider, error) {
	r, err := newVSphereReader(ctx, soapURL, optional{
		Username: user,
		Password: password,
		Insecure: insecure,
	})
	if err != nil {
		return nil, err
	}

	cfg := tfvsphere.Config{
		InsecureFlag:  insecure,
		User:          user,
		Password:      password,
		VSphereServer: vsphereserver,
	}

	log.Get().Log("func", "vsphere.Client", "msg", "configuring TF Client")
	client, err := cfg.Client()
	if err != nil {
		return nil, fmt.Errorf("could not initialize VSphere client: %w", err)
	}

	tfp := tfvsphere.Provider()
	tfp.SetMeta(client)

	// The variables names come from
	// https://registry.terraform.io/providers/hashicorp/vsphere/latest/docs#argument-reference
	rawCfg := terraform.NewResourceConfigRaw(map[string]interface{}{
		"user":                 user,
		"password":             password,
		"vsphere_server":       vsphereserver,
		"allow_unverified_ssl": insecure,
	})

	log.Get().Log("func", "vsphere.NewProvider", "msg", "loading TF client")
	if diags := tfp.Configure(ctx, rawCfg); diags.HasError() {
		return nil, fmt.Errorf("could not initialize 'terraform/vsphere.Provider.Configure()' because: %s", diags[0].Summary)
	}

	return &vsphere{
		tfVSphereClient: client,
		tfProvider:      tfp,
		cache:           cache.New(),
		reader:          r,
	}, nil
}

func (vs vsphere) TFClient() interface{} { return vs.tfVSphereClient }

func (vs vsphere) TFProvider() *schema.Provider { return vs.tfProvider }

func (vs vsphere) String() string { return "vsphere" }

func (vs vsphere) Region() string { return "" }

func (vs vsphere) TagKey() string { return "tags" }

func (vs vsphere) ResourceTypes() []string { return ResourceTypeStrings() }

func (vs vsphere) HasResourceType(t string) bool {
	_, err := ResourceTypeString(t)
	return err == nil
}

func (vs vsphere) Resources(ctx context.Context, resourceType string, f *filter.Filter) ([]provider.Resource, error) {
	rt, err := ResourceTypeString(resourceType)
	if err != nil {
		return nil, err
	}

	rfn, ok := resources[rt]
	if !ok {
		return nil, errors.Errorf("the resource %q it's not implemented", resourceType)
	}

	resources := make([]provider.Resource, 0, 0)
	nres, err := rfn(ctx, &vs, vs.reader, resourceType, f)
	if err != nil {
		return nil, errors.Wrapf(err, "error while reading from resource %q", resourceType)
	}
	resources = append(resources, nres...)

	return resources, nil
}

func (vs vsphere) Source() string  { return "hashicorp/vsphere" }
func (vs vsphere) Version() string { return version }

func (vs vsphere) Configuration() map[string]interface{}                { return vs.configuration }
func (vs vsphere) FixResource(t string, v cty.Value) (cty.Value, error) { return v, nil }
func (vs vsphere) FilterByTags(tags interface{}) error                  { return nil }
