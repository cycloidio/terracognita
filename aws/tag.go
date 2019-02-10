package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Tag it's an easy representation of
// a ec2.Filter for tags
type Tag struct {
	Name  string
	Value string
}

func (t Tag) toEC2Filter() *ec2.Filter {
	return &ec2.Filter{
		Name:   aws.String(fmt.Sprintf("tag:%s", t.Name)),
		Values: []*string{aws.String(t.Value)},
	}
}

func toEC2Filters(tags []Tag) []*ec2.Filter {
	filters := make([]*ec2.Filter, 0, len(tags))

	for _, t := range tags {
		filters = append(filters, t.toEC2Filter())
	}

	return filters
}
