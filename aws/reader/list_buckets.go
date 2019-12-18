package reader

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func (c *connector) ListBuckets(ctx context.Context, input *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	var errs []error
	var ropt = &s3.ListBucketsOutput{}

	if c.svc.s3 == nil {
		c.svc.s3 = s3.New(c.svc.session)
	}

	opt, err := c.svc.s3.ListBucketsWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	newOpt := &s3.ListBucketsOutput{
		Owner:   opt.Owner,
		Buckets: make([]*s3.Bucket, 0),
	}
	for _, bucket := range opt.Buckets {
		inputLocation := &s3.GetBucketLocationInput{
			Bucket: bucket.Name,
		}
		result, err := c.svc.s3.GetBucketLocation(inputLocation)
		if err != nil {
			errs = append(errs, err)
		}
		if s3.NormalizeBucketLocation(aws.StringValue(result.LocationConstraint)) == c.svc.region {
			newOpt.Buckets = append(newOpt.Buckets, bucket)
		}
	}
	ropt = newOpt

	if len(errs) != 0 {
		serrs := make([]string, 0, len(errs))
		for _, e := range errs {
			serrs = append(serrs, e.Error())
		}
		return nil, errors.New(strings.Join(serrs, ","))
	}

	return ropt, nil
}
