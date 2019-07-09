package state_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/mock"
	"github.com/cycloidio/terracognita/provider"
	"github.com/cycloidio/terracognita/state"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform/configs/hcl2shim"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/providers"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terraform-providers/terraform-provider-aws/aws"
)

func TestNewWriter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		sw := state.NewWriter(nil)

		assert.Equal(t, make(map[string]provider.Resource), sw.Config)
	})
}

func TestWrite(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			prv  = mock.NewProvider(ctrl)
			res  = mock.NewResource(ctrl)
			b    = &bytes.Buffer{}
			sw   = state.NewWriter(b)
			tp   = "aws_iam_user"
		)
		defer ctrl.Finish()

		s, err := hcl2shim.HCL2ValueFromFlatmap(map[string]string{"name": "Pepito"}, aws.Provider().(*schema.Provider).ResourcesMap[tp].CoreConfigSchema().ImpliedType())
		require.NoError(t, err)

		res.EXPECT().Type().Return(tp)
		res.EXPECT().Provider().Return(prv)
		res.EXPECT().TFResource().Return(aws.Provider().(*schema.Provider).ResourcesMap[tp])
		res.EXPECT().CoreConfigSchema().Return(aws.Provider().(*schema.Provider).ResourcesMap[tp].CoreConfigSchema())
		res.EXPECT().ResourceInstanceObject().Return(providers.ImportedResource{
			TypeName: tp,
			State:    s,
		}.AsInstanceObject())

		prv.EXPECT().String().Return("aws").AnyTimes()

		err = sw.Write("aws.name", res)
		require.NoError(t, err)

		assert.Equal(t, map[string]provider.Resource{
			"aws.name": res,
		}, sw.Config)
	})
	t.Run("ErrRequiredKey", func(t *testing.T) {
		sw := state.NewWriter(nil)

		err := sw.Write("", nil)
		assert.Equal(t, errcode.ErrWriterRequiredKey, errors.Cause(err))
	})
	t.Run("ErrRequiredValue", func(t *testing.T) {
		sw := state.NewWriter(nil)

		err := sw.Write("aws.key", nil)
		assert.Equal(t, errcode.ErrWriterRequiredValue, errors.Cause(err))
	})
	t.Run("ErrAlreadyExistsKey", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			prv  = mock.NewProvider(ctrl)
			res  = mock.NewResource(ctrl)
			b    = &bytes.Buffer{}
			sw   = state.NewWriter(b)
			tp   = "aws_iam_user"
		)
		defer ctrl.Finish()

		s, err := hcl2shim.HCL2ValueFromFlatmap(map[string]string{"name": "Pepito"}, aws.Provider().(*schema.Provider).ResourcesMap[tp].CoreConfigSchema().ImpliedType())
		require.NoError(t, err)

		res.EXPECT().Type().Return(tp)
		res.EXPECT().Provider().Return(prv)
		res.EXPECT().TFResource().Return(aws.Provider().(*schema.Provider).ResourcesMap[tp])
		res.EXPECT().CoreConfigSchema().Return(aws.Provider().(*schema.Provider).ResourcesMap[tp].CoreConfigSchema())
		res.EXPECT().ResourceInstanceObject().Return(providers.ImportedResource{
			TypeName: tp,
			State:    s,
		}.AsInstanceObject())

		prv.EXPECT().String().Return("aws")

		err = sw.Write("aws.name", res)
		require.NoError(t, err)

		err = sw.Write("aws.name", res)
		assert.Equal(t, errcode.ErrWriterAlreadyExistsKey, errors.Cause(err))
	})
	t.Run("ErrInvalidTypeValue", func(t *testing.T) {
		sw := state.NewWriter(nil)

		err := sw.Write("aws.key", 0)
		assert.Equal(t, errcode.ErrWriterInvalidTypeValue, errors.Cause(err))
	})
	t.Run("ErrInvalidKey", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			res  = mock.NewResource(ctrl)
		)
		defer ctrl.Finish()
		sw := state.NewWriter(nil)

		err := sw.Write("key", res)
		assert.Equal(t, errcode.ErrWriterInvalidKey, errors.Cause(err))

		err = sw.Write("key.a.b", res)
		assert.Equal(t, errcode.ErrWriterInvalidKey, errors.Cause(err))
	})
}

func TestSync(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			ctrl  = gomock.NewController(t)
			b     = &bytes.Buffer{}
			sw    = state.NewWriter(b)
			prv   = mock.NewProvider(ctrl)
			res   = mock.NewResource(ctrl)
			tp    = "aws_iam_user"
			state = `{
   "lineage":"lineage",
   "outputs":{},
   "resources":[
      {
         "instances":[
            {
               "attributes":{
                  "arn":null,
                  "force_destroy":null,
                  "id":null,
                  "name":"Pepito",
                  "path":null,
                  "permissions_boundary":null,
                  "tags":null,
                  "unique_id":null
               },
               "schema_version":0
            }
         ],
         "mode":"managed",
         "name":"name",
         "provider":"provider.aws",
         "type":"aws_iam_user"
      }
   ],
   "serial":0,
   "terraform_version":"0.12.7",
   "version":4
}`
		)

		defer ctrl.Finish()

		s, err := hcl2shim.HCL2ValueFromFlatmap(map[string]string{"name": "Pepito"}, aws.Provider().(*schema.Provider).ResourcesMap[tp].CoreConfigSchema().ImpliedType())
		require.NoError(t, err)

		res.EXPECT().Type().Return(tp)
		res.EXPECT().Provider().Return(prv)
		res.EXPECT().TFResource().Return(aws.Provider().(*schema.Provider).ResourcesMap[tp])
		res.EXPECT().CoreConfigSchema().Return(aws.Provider().(*schema.Provider).ResourcesMap[tp].CoreConfigSchema())
		res.EXPECT().ResourceInstanceObject().Return(providers.ImportedResource{
			TypeName: tp,
			State:    s,
		}.AsInstanceObject())

		prv.EXPECT().String().Return("aws")

		err = sw.Write("aws_iam_user.name", res)
		require.NoError(t, err)

		err = sw.Sync()
		require.NoError(t, err)

		var st map[string]interface{}
		err = json.Unmarshal(b.Bytes(), &st)
		require.NoError(t, err)

		st["lineage"] = "lineage"

		var est map[string]interface{}
		err = json.Unmarshal([]byte(state), &est)
		require.NoError(t, err)

		assert.Equal(t, est, st)
	})
}
