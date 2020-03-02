package azurerm

import (
	"context"

	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/provider"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
)

type azurerm struct {}

// NewProvider returns a Gooogle Provider
func NewProvider(ctx context.Context, maxResults uint64, project, region, credentials string) (provider.Provider, error) {
	return nil, nil
}

func (g *azurerm) HasResourceType(t string) bool {
	_, err := ResourceTypeString(t)
	return err == nil
}

func (g *azurerm) Region() string  { return "a-region" }
func (g *azurerm) Project() string { return "a-project" }
func (g *azurerm) String() string  { return "azurerm" }
func (g *azurerm) TagKey() string  { return "n/a" }

func (g *azurerm) ResourceTypes() []string {
	return ResourceTypeStrings()
}

func (g *azurerm) Resources(ctx context.Context, t string, f *filter.Filter) ([]provider.Resource, error) {
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

func (g *azurerm) TFClient() interface{} { return nil }

func (g *azurerm) TFProvider() *schema.Provider { return nil }
