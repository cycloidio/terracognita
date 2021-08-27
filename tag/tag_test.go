package tag_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/tag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToEC2Filer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tt := tag.Tag{Name: "tag-name", Value: "tag-value"}
		assert.Equal(t, &ec2.Filter{
			Name:   aws.String("tag:tag-name"),
			Values: []*string{aws.String("tag-value")},
		}, tt.ToEC2Filter())
	})
}

func TestToNeptuneFiler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tt := tag.Tag{Name: "tag-name", Value: "tag-value"}
		assert.Equal(t, &neptune.Filter{
			Name:   aws.String("tag:tag-name"),
			Values: []*string{aws.String("tag-value")},
		}, tt.ToNeptuneFilter())
	})
}

func TestToRDSFiler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tt := tag.Tag{Name: "tag-name", Value: "tag-value"}
		assert.Equal(t, &rds.Filter{
			Name:   aws.String("tag:tag-name"),
			Values: []*string{aws.String("tag-value")},
		}, tt.ToRDSFilter())
	})
}

func TestNew(t *testing.T) {
	tests := []struct {
		Name  string
		STag  string
		ETag  tag.Tag
		Error bool
	}{
		{
			Name: "Success",
			STag: "key:val",
			ETag: tag.Tag{Name: "key", Value: "val"},
		},
		{
			Name:  "ErrorEmpty",
			STag:  "",
			Error: true,
		},
		{
			Name:  "ErrorNoSeparator",
			STag:  "key",
			Error: true,
		},
		{
			Name:  "ErrorMoreThanOneSeparator",
			STag:  "key:value:what",
			Error: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tg, err := tag.New(tt.STag)
			if tt.Error {
				assert.EqualError(t, err, errcode.ErrTagInvalidForamt.Error())
			}
			assert.Equal(t, tt.ETag, tg)
		})
	}
}

func TestGetNameFromTag(t *testing.T) {
	var tagKey = "Name"
	tests := []struct {
		Name     string
		Key      string
		SRD      *schema.ResourceData
		Fallback string
		Result   string
	}{
		{
			Name:     "WithTags",
			Key:      "tags",
			SRD:      createSRD(t, "tags", tagKey, "res"),
			Fallback: "fallback",
			Result:   "res",
		},
		{
			Name:     "WithTagsButInvalidNameAndEmptyForced",
			Key:      "tags",
			SRD:      createSRD(t, "tags", tagKey, "res.res.res"),
			Fallback: "fallback",
			Result:   "res_res_res",
		},
		{
			Name:     "WithTagsButInvalidNameAndEmptyForced)",
			Key:      "tags",
			SRD:      createSRD(t, "tags", tagKey, "..."),
			Fallback: "fallback",
			Result:   "fallback",
		},
		{
			Name:     "WithTagsButNo'Name'",
			Key:      "tags",
			SRD:      createSRD(t, "tags", "notName", "res"),
			Fallback: "fallback",
			Result:   "fallback",
		},
		{
			Name:     "WithTagsButNo'Name'AndInvalidFallbackForced",
			Key:      "tags",
			SRD:      createSRD(t, "tags", "notName", "res"),
			Fallback: "fal.lback",
			Result:   "fal_lback",
		},
		{
			Name:     "WithTagsButNo'Name'AndInvalidFallbackForcedEmpty",
			Key:      "tags",
			SRD:      createSRD(t, "tags", "notName", "res"),
			Fallback: "...",
			Result:   "",
		},
		{
			Name:     "WithNoTags",
			Key:      "tags",
			SRD:      createSRD(t, "noTags", tagKey, "res"),
			Fallback: "fallback",
			Result:   "fallback",
		},
		{
			Name:     "WithNoTagsAndInvalidFallback",
			Key:      "tags",
			SRD:      createSRD(t, "noTags", tagKey, "res"),
			Fallback: "...",
			Result:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			name := tag.GetNameFromTag(tt.Key, tt.SRD, tt.Fallback)
			if tt.Result == "" {
				assert.Len(t, name, 5)
			} else {
				assert.Equal(t, tt.Result, name)
			}
		})
	}
}

// createSRD creates a schema.ResourceData with a
// 'schemaKey' of TypeMap with a 'tagKey' with 'tagValue'
func createSRD(t *testing.T, schemaKey, tagKey, tagValue string) *schema.ResourceData {
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{
			schemaKey: &schema.Schema{
				Type:     schema.TypeMap,
				Optional: false,
			},
		},
	}

	rd := r.Data(nil)

	err := rd.Set(schemaKey, map[string]interface{}{
		tagKey: tagValue,
	})
	require.NoError(t, err)

	return rd
}

// createSRDOtherTags creates a schema.ResourceData with a
// 'schemaKey' of Set with a 'Key=tagKey' with 'Value=tagValue'
// https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/autoscaling_group#tag-and-tags
// https://github.com/hashicorp/terraform-plugin-sk/blob/main/helper/schema/set.go#L50
func createSRDOtherTags(t *testing.T, schemaKey, tagKey, tagValue string) *schema.ResourceData {
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{
			schemaKey: &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"value": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"propagate_at_launch": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}

	rd := r.Data(nil)

	listTag := []interface{}{
		map[string]interface{}{
			"key":                 tagKey,
			"value":               tagValue,
			"propagate_at_launch": true,
		},
		map[string]interface{}{
			"key":                 "fakeOtkerKey",
			"value":               "fakeOtkerValue",
			"propagate_at_launch": true,
		},
	}

	err := rd.Set(schemaKey, listTag)
	require.NoError(t, err)

	return rd
}

func TestGetOtherTags(t *testing.T) {
	var filterTagKey = "TagName"
	var filterTagValue = "TagValue"
	tests := []struct {
		Name      string
		Provider  string
		FilterTag tag.Tag
		SRD       *schema.ResourceData
		Match     bool
		Result    string
	}{
		{
			Name:      "WithoutTagButTags",
			Provider:  "aws",
			FilterTag: tag.Tag{Name: filterTagKey, Value: filterTagValue},
			SRD:       createSRD(t, "tags", filterTagKey, filterTagValue),
			Match:     false,
			Result:    "",
		},
		{
			Name:      "WithTagNoMatch",
			Provider:  "aws",
			FilterTag: tag.Tag{Name: filterTagKey, Value: filterTagValue},
			SRD:       createSRDOtherTags(t, "tag", "TagNameNoMatch", filterTagValue),
			Match:     false,
			Result:    "",
		},
		{
			Name:      "WithTagAndMatch",
			Provider:  "aws",
			FilterTag: tag.Tag{Name: filterTagKey, Value: filterTagValue},
			SRD:       createSRDOtherTags(t, "tag", filterTagKey, filterTagValue),
			Match:     true,
			Result:    filterTagValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			value, match := tag.GetOtherTags(tt.Provider, tt.SRD, tt.FilterTag)
			assert.Equal(t, tt.Match, match)
			assert.Equal(t, tt.Result, value)
		})
	}
}
