package util

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

func (t Tag) ToEC2Filter() *ec2.Filter {
	return &ec2.Filter{
		Name:   aws.String(fmt.Sprintf("tag:%s", t.Name)),
		Values: []*string{aws.String(t.Value)},
	}
}
