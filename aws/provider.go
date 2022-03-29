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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfaws "github.com/hashicorp/terraform-provider-aws/aws"
	"github.com/pkg/errors"
)

// skippableCodes is a list of codes
// which won't make Terracognita failed
// but they will be printed on the output
// they are based on the err.Code() content
// of the AWS error
var skippableCodes = map[string]struct{}{
	"InvalidAction":         struct{}{},
	"AccessDeniedException": struct{}{},
	"RequestError":          struct{}{},
}

type aws struct {
	awsr reader.Reader

	tfAWSClient interface{}
	tfProvider  *schema.Provider

	configuration map[string]interface{}

	cache cache.Cache
}

// NewProvider returns an AWS Provider
func NewProvider(ctx context.Context, accessKey, secretKey, region, sessionToken string) (provider.Provider, error) {
	log.Get().Log("func", "reader.New", "msg", "configuring aws Reader")
	awsr, err := reader.New(ctx, accessKey, secretKey, region, sessionToken, nil)
	if err != nil {
		return nil, fmt.Errorf("could not initialize 'reader' because: %s", err)
	}

	cfg := tfaws.Config(accessKey, secretKey, region, sessionToken)

	log.Get().Log("func", "aws.NewProvider", "msg", "configuring TF Client")
	awsClient, diags := cfg.Client(ctx)
	if diags.HasError() {
		var errdiags string
		for i := range diags {
			errdiags += fmt.Sprintf("%s: %s", diags[i].Summary, diags[i].Detail)
		}
		return nil, fmt.Errorf("could not initialize 'terraform/aws.Config.Client()' because: %s", errdiags)
	}

	tfp := tfaws.Provider()
	tfp.SetMeta(awsClient)

	return &aws{
		awsr:        awsr,
		tfAWSClient: awsClient,
		tfProvider:  tfp,
		cache:       cache.New(),
		configuration: map[string]interface{}{
			"region": region,
		},
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
		if reqErr, ok := err.(awserr.Error); ok {
			if _, ok := skippableCodes[reqErr.Code()]; ok {
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
func (a *aws) Source() string                        { return "hashicorp/aws" }
func (a *aws) Configuration() map[string]interface{} { return a.configuration }
