package aws

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cycloidio/terracognita/aws/reader"
	"github.com/cycloidio/terracognita/cache"
	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/provider"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-cty/cty/gocty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/conns"
	tfaws "github.com/hashicorp/terraform-provider-aws/provider"
	"github.com/pkg/errors"
)

// version of the Terraform provider, this is automatically changed with the 'make update-terraform-provider'
const version = "4.9.0"

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

	cfg := conns.Config{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    region,
		Token:     sessionToken,
	}

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
func (a *aws) Version() string                       { return version }
func (a *aws) Configuration() map[string]interface{} { return a.configuration }
func (a *aws) FixResource(t string, v cty.Value) (cty.Value, error) {
	var err error
	switch t {
	case "aws_db_subnet_group":
		err = cty.Walk(v, func(path cty.Path, val cty.Value) (bool, error) {
			if len(path) > 0 {
				if gas, ok := path[0].(cty.GetAttrStep); ok {
					switch gas.Name {
					case "name":
						var sd string
						err := gocty.FromCtyValue(val, &sd)
						if err != nil {
							return false, errors.Wrapf(err, "failed to convert CTY value to GO type")
						}
						if sd == "default" {
							return false, fmt.Errorf("ignoring 'aws_db_subnet_group' with 'default' name as it's managed for AWS")
						}
					}
				}
			}
			return true, nil
		})
		if err != nil {
			return v, errors.Wrapf(err, "failed to fix resources")
		}
	case "aws_alb_listener_rule", "aws_lb_listener_rule":
		err = cty.Walk(v, func(path cty.Path, val cty.Value) (bool, error) {
			if len(path) > 0 {
				if gas, ok := path[0].(cty.GetAttrStep); ok {
					switch gas.Name {
					case "priority":
						var sp string
						err := gocty.FromCtyValue(val, &sp)
						if err != nil {
							return false, errors.Wrapf(err, "failed to convert CTY value to GO type")
						}
						if sp == "99999" {
							return false, fmt.Errorf("ignoring 'aws_alb_listener_rule' or 'aws_lb_listener_rule' with 'priority: 99999' name as it's managed for AWS")
						}
					}
				}
			}
			return true, nil
		})
		if err != nil {
			return v, errors.Wrapf(err, "failed to fix resources")
		}
	}
	return v, nil
}

var (
	autogeneratedAWSResourcesRe = regexp.MustCompile(`^aws:(?:autoscaling|cloudformation)`)
)

func (a *aws) FilterByTags(tags interface{}) error {
	ts, ok := tags.(map[string]interface{})
	if !ok {
		return nil
	}
	for k := range ts {
		if autogeneratedAWSResourcesRe.MatchString(k) {
			return errors.WithStack(errcode.ErrProviderResourceAutogenerated)
		}
	}
	return nil
}
