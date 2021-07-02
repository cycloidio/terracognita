package azurerm

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
	tfazurerm "github.com/terraform-providers/terraform-provider-azurerm/azurerm"

	"github.com/cycloidio/terracognita/cache"
	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/provider"
)

type azurerm struct {
	tfAzureRMClient interface{}
	tfProvider      *schema.Provider
	azurer          *AzureReader

	cache cache.Cache
}

// NewProvider returns a AzureRM Provider
func NewProvider(ctx context.Context, clientID, clientSecret, environment, resourceGroupName, subscriptionID, tenantID string) (provider.Provider, error) {
	log.Get().Log("func", "azurerm.NewProvider", "msg", "loading Azure reader")
	reader, err := NewAzureReader(ctx, clientID, clientSecret, environment, resourceGroupName, subscriptionID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("could not initialize AzureReader: %s", err)
	}

	log.Get().Log("func", "azurerm.NewProvider", "msg", "loading TF provider")
	tfp := tfazurerm.Provider()

	rawCfg := terraform.NewResourceConfigRaw(map[string]interface{}{
		"client_id":       clientID,
		"client_secret":   clientSecret,
		"environment":     environment,
		"subscription_id": subscriptionID,
		"tenant_id":       tenantID,
	})

	log.Get().Log("func", "azurerm.NewProvider", "msg", "loading TF client")
	if diags := tfp.Configure(ctx, rawCfg); diags.HasError() {
		return nil, fmt.Errorf("could not initialize 'terraform/azurerm.Provider.Configure()' because: %s", diags[0].Summary)
	}

	return &azurerm{
		tfAzureRMClient: tfp.Meta(),
		tfProvider:      tfp,
		azurer:          reader,
		cache:           cache.New(),
	}, nil
}

func (a *azurerm) HasResourceType(t string) bool {
	_, err := ResourceTypeString(t)
	return err == nil
}

func (a *azurerm) ResourceGroup() string { return a.azurer.GetResourceGroupName() }
func (a *azurerm) Region() string        { return a.azurer.GetLocation() }
func (a *azurerm) String() string        { return "azurerm" }
func (a *azurerm) TagKey() string        { return "tags" }

func (a *azurerm) ResourceTypes() []string {
	return ResourceTypeStrings()
}

func (a *azurerm) Resources(ctx context.Context, t string, f *filter.Filter) ([]provider.Resource, error) {
	rt, err := ResourceTypeString(t)
	if err != nil {
		return nil, err
	}

	rfn, ok := resources[rt]
	if !ok {
		return nil, errors.Errorf("the resource %q it's not implemented", t)
	}

	resources, err := rfn(ctx, a, t, f)
	if err != nil {
		return nil, errors.Wrapf(err, "error while reading from resource %q", t)
	}

	return resources, nil
}

func (a *azurerm) TFClient() interface{} {
	return a.tfAzureRMClient
}

func (a *azurerm) TFProvider() *schema.Provider {
	return a.tfProvider
}
