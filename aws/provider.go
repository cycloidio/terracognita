package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cycloidio/terracognita/aws/reader"
	"github.com/cycloidio/terracognita/cache"
	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/provider"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
	tfaws "github.com/terraform-providers/terraform-provider-aws/aws"
)

// skippableErrors is a list of errors
// which won't make Terracognita failed
// but they will be printed on the output
// they are based on the err.Message() content
// of the AWS error
var skippableErrors = map[string]struct{}{
	"Unavailable Operation": struct{}{},
}

type aws struct {
	awsr reader.Reader

	tfAWSClient interface{}
	tfProvider  *schema.Provider

	cache cache.Cache
}

// NewProvider returns an AWS Provider
func NewProvider(ctx context.Context, accessKey, secretKey, region string) (provider.Provider, error) {
	log.Get().Log("func", "reader.New", "msg", "configuring aws Reader")
	awsr, err := reader.New(ctx, accessKey, secretKey, region, nil)
	if err != nil {
		return nil, fmt.Errorf("could not initialize 'reader' because: %s", err)
	}

	cfg := tfaws.Config{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    region,
	}

	log.Get().Log("func", "aws.NewProvider", "msg", "configuring TF Client")
	awsClient, err := cfg.Client()
	if err != nil {
		return nil, fmt.Errorf("could not initialize 'terraform/aws.Config.Client()' because: %s", err)
	}

	tfp := tfaws.Provider().(*schema.Provider)
	tfp.SetMeta(awsClient)

	return &aws{
		awsr:        awsr,
		tfAWSClient: awsClient,
		tfProvider:  tfp,
		cache:       cache.New(),
	}, nil
}

func (a *aws) ResourceTypes() []string {
	return ResourceTypeStrings()
}

func (a *aws) Resources(ctx context.Context, t string, f *filter.Filter) ([]provider.Resource, error) {
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
		// we filter the error from AWS and return a custom error
		// type if it's an error that we want to skip
		if reqErr, ok := err.(awserr.RequestFailure); ok {
			if _, ok := skippableErrors[reqErr.Message()]; ok {
				return nil, fmt.Errorf("%w: %v", errcode.ErrProviderAPI, reqErr)
			}
		}
		return nil, errors.Wrapf(err, "error while reading from resource %q", t)
	}

	return resources, nil
}

func (a *aws) TFClient() interface{} {
	return a.tfAWSClient
}

func (a *aws) TFProvider() *schema.Provider {
	return a.tfProvider
}

func (a *aws) String() string { return "aws" }

func (a *aws) Region() string { return a.awsr.GetRegion() }
func (a *aws) TagKey() string { return "tags" }
func (a *aws) HasResourceType(t string) bool {
	_, err := ResourceTypeString(t)
	return err == nil
}
