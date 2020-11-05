package tag

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/chr4/pwgen"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// nameRegexp is the new regexp used to validate the names
// of the resources on TF (defined on configs/configschema/internal_validate.go)
var nameRegexp = regexp.MustCompile(`^[a-z0-9_]+$`)

// Tag it's an easy representation of
// a ec2.Filter for tags
type Tag struct {
	Name  string
	Value string
}

// ToEC2Filter transforms the Tag to a ec2.Filter
// to use on AWS filters
func (t Tag) ToEC2Filter() *ec2.Filter {
	return &ec2.Filter{
		Name:   aws.String(fmt.Sprintf("tag:%s", t.Name)),
		Values: []*string{aws.String(t.Value)},
	}
}

// GetNameFromTag returns the 'tags.Name' from the src or the fallback
// if it's not defined.
// Also validates that the 'tags.Name' and fallback are valid, if not it
// generates a random one
func GetNameFromTag(key string, srd *schema.ResourceData, fallback string) string {
	var n string
	if name, ok := srd.GetOk(fmt.Sprintf("%s.Name", key)); ok {
		n = name.(string)
	}

	if isValidResourceName(n) {
		return n
	} else if isValidResourceName(fallback) {
		return fallback
	} else {
		return pwgen.Alpha(5)
	}
}

// isValidResourceName checks with the TF regex
// for names to validate if it's valid
func isValidResourceName(name string) bool {
	return nameRegexp.MatchString(name)
}
