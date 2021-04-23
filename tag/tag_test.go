package tag_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/tag"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
			Name:     "WithTagsButInvalidName",
			Key:      "tags",
			SRD:      createSRD(t, "tags", tagKey, "res.res"),
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
			Name:     "WithTagsButNo'Name'AndInvalidFallback",
			Key:      "tags",
			SRD:      createSRD(t, "tags", "notName", "res"),
			Fallback: "fal.lback",
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
			Fallback: "fall.back",
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
