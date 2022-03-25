package azurerm

import (
	"context"
	"fmt"

	autorestAzure "github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfazurerm "github.com/hashicorp/terraform-provider-azurerm/azurerm"
	"github.com/pkg/errors"

	"github.com/cycloidio/terracognita/cache"
	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/provider"
)

// skippableCodes is a list of codes
// which won't make Terracognita failed
// but they will be printed on the output
// they are based on the err.Code() content
// of the Azure error
var skippableCodes = map[string]struct{}{
	"ParentResourceNotFound": struct{}{},
}

type azurerm struct {
	tfAzureRMClient interface{}
	tfProvider      *schema.Provider
	azurer          *AzureReader

	configuraiton map[string]interface{}

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
		configuraiton: map[string]interface{}{
			"environment": environment,
		},
	}, nil
}

func (a *azurerm) HasResourceType(t string) bool {
	_, err := ResourceTypeString(t)
	return err == nil
}

func (a *azurerm) ResourceGroup() string                 { return a.azurer.GetResourceGroupName() }
func (a *azurerm) Region() string                        { return a.azurer.GetLocation() }
func (a *azurerm) String() string                        { return "azurerm" }
func (a *azurerm) TagKey() string                        { return "tags" }
func (a *azurerm) Source() string                        { return "hashicorp/azurerm" }
func (a *azurerm) Configuration() map[string]interface{} { return a.configuraiton }

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
		// we filter the error from Azure and return a custom error
		// type if it's an error that we want to skip
		// Remove all wrap layer to get the right type
		unwrapErr := err
		for errors.Unwrap(unwrapErr) != nil {
			unwrapErr = errors.Unwrap(unwrapErr)
		}
		if reqErr, ok := unwrapErr.(*autorestAzure.RequestError); ok {
			if _, ok := skippableCodes[reqErr.ServiceError.Code]; ok {
				return nil, fmt.Errorf("%w: %v", errcode.ErrProviderAPI, reqErr)
			}
		}

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
