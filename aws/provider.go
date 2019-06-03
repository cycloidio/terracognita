package aws

import (
	"context"
	"fmt"

	"github.com/cycloidio/raws"
	"github.com/cycloidio/terracognita/cache"
	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/provider"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	tfaws "github.com/terraform-providers/terraform-provider-aws/aws"
)

type aws struct {
	awsr raws.AWSReader

	tfAWSClient interface{}
	tfProvider  *schema.Provider

	cache cache.Cache
}

// NewProvider returns an AWS Provider
func NewProvider(ctx context.Context, accessKey, secretKey, region string) (provider.Provider, error) {
	awsr, err := raws.NewAWSReader(ctx, accessKey, secretKey, []string{region}, nil)
	if err != nil {
		return nil, fmt.Errorf("could not initialize 'raws' because: %s", err)
	}

	cfg := tfaws.Config{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    region,
	}

	awsClient, err := cfg.Client()
	if err != nil {
		return nil, fmt.Errorf("could not initialize 'terraform/aws.Config.Client()' because: %s", err)
	}

	return &aws{
		awsr:        awsr,
		tfAWSClient: awsClient,
		tfProvider:  tfaws.Provider().(*schema.Provider),
		cache:       cache.New(),
	}, nil
}

func (a *aws) ResourceTypes() []string {
	return ResourceTypeStrings()
}

func (a *aws) Resources(ctx context.Context, t string, f filter.Filter) ([]*provider.Resource, error) {
	rt, err := ResourceTypeString(t)
	if err != nil {
		return nil, err
	}

	rfn, ok := resources[rt]
	if !ok {
		return nil, errors.Errorf("the resource %q it's not implemented", t)
	}

	resources, err := rfn(ctx, a, t, f.Tags)
	if err != nil {
		return nil, errors.Wrapf(err, "error while reading from resource %q", t)
	}

	return resources, nil
}

func (a *aws) TFClient() interface{} {
	return a.tfAWSClient
}

func (a *aws) String() string { return "aws" }

func (a *aws) Region() string { return a.awsr.GetRegions()[0] }
func (a *aws) TagKey() string { return "tags" }
