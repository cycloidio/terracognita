package state_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/mock"
	"github.com/cycloidio/terracognita/provider"
	"github.com/cycloidio/terracognita/state"
	"github.com/cycloidio/terracognita/util"
	"github.com/cycloidio/terracognita/writer"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws"
	"github.com/hashicorp/terraform/configs/hcl2shim"
	"github.com/hashicorp/terraform/providers"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWriter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		sw := state.NewWriter(nil, nil)

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
			sw   = state.NewWriter(b, &writer.Options{Interpolate: true})
			tp   = "aws_iam_user"
			key  = "aws.name"
		)
		defer ctrl.Finish()

		tpt, err := util.HashicorpToZclonfType(aws.Provider().ResourcesMap[tp].CoreConfigSchema().ImpliedType())
		require.NoError(t, err)

		s, err := hcl2shim.HCL2ValueFromFlatmap(map[string]string{"name": "Pepito"}, tpt)
		require.NoError(t, err)

		res.EXPECT().Type().Return(tp)
		res.EXPECT().Provider().Return(prv)
		res.EXPECT().TFResource().Return(aws.Provider().ResourcesMap[tp])
		res.EXPECT().ImpliedType().Return(aws.Provider().ResourcesMap[tp].CoreConfigSchema().ImpliedType())
		res.EXPECT().ResourceInstanceObject().Return(providers.ImportedResource{
			TypeName: tp,
			State:    s,
		}.AsInstanceObject())

		prv.EXPECT().String().Return("aws").AnyTimes()

		err = sw.Write(key, res)
		require.NoError(t, err)

		assert.Equal(t, map[string]provider.Resource{
			key: res,
		}, sw.Config)
		t.Run("Has", func(t *testing.T) {
			ok, err := sw.Has(key)
			require.NoError(t, err)
			assert.True(t, ok)

			ok, err = sw.Has("aws.new")
			require.NoError(t, err)
			assert.False(t, ok)
		})
	})
	t.Run("ErrRequiredKey", func(t *testing.T) {
		sw := state.NewWriter(nil, &writer.Options{Interpolate: true})

		err := sw.Write("", nil)
		assert.Equal(t, errcode.ErrWriterRequiredKey, errors.Cause(err))
	})
	t.Run("ErrRequiredValue", func(t *testing.T) {
		sw := state.NewWriter(nil, &writer.Options{Interpolate: true})

		err := sw.Write("aws.key", nil)
		assert.Equal(t, errcode.ErrWriterRequiredValue, errors.Cause(err))
	})
	t.Run("ErrAlreadyExistsKey", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			prv  = mock.NewProvider(ctrl)
			res  = mock.NewResource(ctrl)
			b    = &bytes.Buffer{}
			sw   = state.NewWriter(b, &writer.Options{Interpolate: true})
			tp   = "aws_iam_user"
		)
		defer ctrl.Finish()

		tpt, err := util.HashicorpToZclonfType(aws.Provider().ResourcesMap[tp].CoreConfigSchema().ImpliedType())
		require.NoError(t, err)

		s, err := hcl2shim.HCL2ValueFromFlatmap(map[string]string{"name": "Pepito"}, tpt)
		require.NoError(t, err)

		res.EXPECT().Type().Return(tp)
		res.EXPECT().Provider().Return(prv)
		res.EXPECT().TFResource().Return(aws.Provider().ResourcesMap[tp])
		res.EXPECT().ImpliedType().Return(aws.Provider().ResourcesMap[tp].CoreConfigSchema().ImpliedType())
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
		sw := state.NewWriter(nil, &writer.Options{Interpolate: true})

		err := sw.Write("aws.key", 0)
		assert.Equal(t, errcode.ErrWriterInvalidTypeValue, errors.Cause(err))
	})
	t.Run("ErrInvalidKey", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			res  = mock.NewResource(ctrl)
		)
		defer ctrl.Finish()
		sw := state.NewWriter(nil, &writer.Options{Interpolate: true})

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
			sw    = state.NewWriter(b, &writer.Options{Interpolate: true})
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
                  "tags_all":null,
                  "unique_id":null
               },
               "schema_version":0
            }
         ],
         "mode":"managed",
         "name":"name",
         "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
         "type":"aws_iam_user"
      }
   ],
   "serial":0,
   "terraform_version":"0.13.5",
   "version":4
}`
		)

		defer ctrl.Finish()

		tpt, err := util.HashicorpToZclonfType(aws.Provider().ResourcesMap[tp].CoreConfigSchema().ImpliedType())
		require.NoError(t, err)

		s, err := hcl2shim.HCL2ValueFromFlatmap(map[string]string{"name": "Pepito"}, tpt)
		require.NoError(t, err)

		res.EXPECT().Type().Return(tp)
		res.EXPECT().Provider().Return(prv)
		res.EXPECT().TFResource().Return(aws.Provider().ResourcesMap[tp])
		res.EXPECT().ImpliedType().Return(aws.Provider().ResourcesMap[tp].CoreConfigSchema().ImpliedType())
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
	t.Run("SuccessWithModule", func(t *testing.T) {
		var (
			ctrl  = gomock.NewController(t)
			b     = &bytes.Buffer{}
			sw    = state.NewWriter(b, &writer.Options{Interpolate: true, Module: "cycloid"})
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
                  "tags_all":null,
                  "unique_id":null
               },
               "schema_version":0
            }
         ],
         "mode":"managed",
         "name":"name",
         "module": "module.cycloid",
         "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
         "type":"aws_iam_user"
      }
   ],
   "serial":0,
   "terraform_version":"0.13.5",
   "version":4
}`
		)

		defer ctrl.Finish()

		tpt, err := util.HashicorpToZclonfType(aws.Provider().ResourcesMap[tp].CoreConfigSchema().ImpliedType())
		require.NoError(t, err)

		s, err := hcl2shim.HCL2ValueFromFlatmap(map[string]string{"name": "Pepito"}, tpt)
		require.NoError(t, err)

		res.EXPECT().Type().Return(tp)
		res.EXPECT().Provider().Return(prv)
		res.EXPECT().TFResource().Return(aws.Provider().ResourcesMap[tp])
		res.EXPECT().ImpliedType().Return(aws.Provider().ResourcesMap[tp].CoreConfigSchema().ImpliedType())
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

func TestDependencies(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			i      = make(map[string]string)
			ctrl   = gomock.NewController(t)
			b      = &bytes.Buffer{}
			sw     = state.NewWriter(b, &writer.Options{Interpolate: true})
			prv    = mock.NewProvider(ctrl)
			resSG  = mock.NewResource(ctrl)
			resSGR = mock.NewResource(ctrl)
			sg     = "aws_security_group"
			sgr    = "aws_security_group_rule"
			state  = `{
  "version": 4,
  "terraform_version": "0.13.5",
  "serial": 0,
  "lineage": "lineage",
  "outputs": {},
  "resources": [
    {
      "mode": "managed",
      "type": "aws_security_group",
      "name": "sg",
      "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
      "instances": [
        {
          "schema_version": 1,
          "attributes": {
            "id": "sg-1234",
            "name": "sg",
            "description": null,
            "egress": null,
            "ingress": null,
            "name_prefix": null,
	    "arn": null,
            "owner_id": null,
            "revoke_rules_on_delete": null,
            "timeouts": {
              "create": null,
              "delete": null
            },
						"tags": null,
						"tags_all": null,
            "vpc_id": null
          }
        }
      ]
    },
    {
      "mode": "managed",
      "type": "aws_security_group_rule",
      "name": "sgrule",
      "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
      "instances": [
        {
          "schema_version": 2,
          "attributes": {
            "id": "sgrule-1234",
            "security_group_id": "sg-1234",
            "cidr_blocks": null,
            "description": null,
            "from_port": null,
            "ipv6_cidr_blocks": null,
            "prefix_list_ids": null,
            "protocol": null,
            "self": null,
            "source_security_group_id": null,
            "to_port": null,
            "type": null
          },
          "dependencies": [
            "aws_security_group.sg"
          ]
        }
      ]
    }
  ]
}`
		)

		defer ctrl.Finish()

		sgt, err := util.HashicorpToZclonfType(aws.Provider().ResourcesMap[sg].CoreConfigSchema().ImpliedType())
		require.NoError(t, err)
		sgrt, err := util.HashicorpToZclonfType(aws.Provider().ResourcesMap[sgr].CoreConfigSchema().ImpliedType())
		require.NoError(t, err)

		stateSG, err := hcl2shim.HCL2ValueFromFlatmap(map[string]string{"id": "sg-1234", "name": "sg"}, sgt)
		stateSGR, err := hcl2shim.HCL2ValueFromFlatmap(map[string]string{"security_group_id": "sg-1234", "id": "sgrule-1234"}, sgrt)

		require.NoError(t, err)

		resSG.EXPECT().Type().Return(sg)
		resSG.EXPECT().Provider().Return(prv)
		resSG.EXPECT().TFResource().Return(aws.Provider().ResourcesMap[sg])
		resSG.EXPECT().ImpliedType().Return(aws.Provider().ResourcesMap[sg].CoreConfigSchema().ImpliedType())
		resSG.EXPECT().ResourceInstanceObject().Return(providers.ImportedResource{
			TypeName: sg,
			State:    stateSG,
		}.AsInstanceObject())
		attrsSG := make(map[string]string)
		resSG.EXPECT().InstanceState().Return(&terraform.InstanceState{
			Attributes: attrsSG,
		})

		prv.EXPECT().String().Return("aws").AnyTimes()

		resSGR.EXPECT().Type().Return(sgr).AnyTimes()
		resSGR.EXPECT().Provider().Return(prv)
		resSGR.EXPECT().TFResource().Return(aws.Provider().ResourcesMap[sgr])
		resSGR.EXPECT().ImpliedType().Return(aws.Provider().ResourcesMap[sgr].CoreConfigSchema().ImpliedType())
		resSGR.EXPECT().ResourceInstanceObject().Return(providers.ImportedResource{
			TypeName: sgr,
			State:    stateSGR,
		}.AsInstanceObject())
		attrsSGR := make(map[string]string)
		attrsSGR["security-group-id"] = "sg-1234"
		resSGR.EXPECT().InstanceState().Return(&terraform.InstanceState{
			Attributes: attrsSGR,
		})

		err = sw.Write("aws_security_group.sg", resSG)
		require.NoError(t, err)
		err = sw.Write("aws_security_group_rule.sgrule", resSGR)
		require.NoError(t, err)

		i["sg-1234"] = "${aws_security_group.sg.id}"
		sw.Interpolate(i)

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
	t.Run("SuccessWitmModules", func(t *testing.T) {
		var (
			i      = make(map[string]string)
			ctrl   = gomock.NewController(t)
			b      = &bytes.Buffer{}
			sw     = state.NewWriter(b, &writer.Options{Interpolate: true, Module: "cycloid"})
			prv    = mock.NewProvider(ctrl)
			resSG  = mock.NewResource(ctrl)
			resSGR = mock.NewResource(ctrl)
			sg     = "aws_security_group"
			sgr    = "aws_security_group_rule"
			state  = `{
  "version": 4,
  "terraform_version": "0.13.5",
  "serial": 0,
  "lineage": "lineage",
  "outputs": {},
  "resources": [
    {
      "mode": "managed",
      "module": "module.cycloid",
      "type": "aws_security_group",
      "name": "sg",
      "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
      "instances": [
        {
          "schema_version": 1,
          "attributes": {
            "id": "sg-1234",
            "name": "sg",
            "description": null,
            "egress": null,
            "ingress": null,
            "name_prefix": null,
            "arn": null,
            "owner_id": null,
            "revoke_rules_on_delete": null,
            "timeouts": {
              "create": null,
              "delete": null
            },
            "tags": null,
            "tags_all": null,
            "vpc_id": null
          }
        }
      ]
    },
    {
      "mode": "managed",
      "module": "module.cycloid",
      "type": "aws_security_group_rule",
      "name": "sgrule",
      "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
      "instances": [
        {
          "schema_version": 2,
          "attributes": {
            "id": "sgrule-1234",
            "security_group_id": "sg-1234",
            "cidr_blocks": null,
            "description": null,
            "from_port": null,
            "ipv6_cidr_blocks": null,
            "prefix_list_ids": null,
            "protocol": null,
            "self": null,
            "source_security_group_id": null,
            "to_port": null,
            "type": null
          },
          "dependencies": [
            "module.cycloid.aws_security_group.sg"
          ]
        }
      ]
    }
  ]
}`
		)

		defer ctrl.Finish()

		sgt, err := util.HashicorpToZclonfType(aws.Provider().ResourcesMap[sg].CoreConfigSchema().ImpliedType())
		require.NoError(t, err)
		sgrt, err := util.HashicorpToZclonfType(aws.Provider().ResourcesMap[sgr].CoreConfigSchema().ImpliedType())
		require.NoError(t, err)

		stateSG, err := hcl2shim.HCL2ValueFromFlatmap(map[string]string{"id": "sg-1234", "name": "sg"}, sgt)
		stateSGR, err := hcl2shim.HCL2ValueFromFlatmap(map[string]string{"security_group_id": "sg-1234", "id": "sgrule-1234"}, sgrt)

		require.NoError(t, err)

		resSG.EXPECT().Type().Return(sg)
		resSG.EXPECT().Provider().Return(prv)
		resSG.EXPECT().TFResource().Return(aws.Provider().ResourcesMap[sg])
		resSG.EXPECT().ImpliedType().Return(aws.Provider().ResourcesMap[sg].CoreConfigSchema().ImpliedType())
		resSG.EXPECT().ResourceInstanceObject().Return(providers.ImportedResource{
			TypeName: sg,
			State:    stateSG,
		}.AsInstanceObject())
		attrsSG := make(map[string]string)
		resSG.EXPECT().InstanceState().Return(&terraform.InstanceState{
			Attributes: attrsSG,
		})

		prv.EXPECT().String().Return("aws").AnyTimes()

		resSGR.EXPECT().Type().Return(sgr).AnyTimes()
		resSGR.EXPECT().Provider().Return(prv)
		resSGR.EXPECT().TFResource().Return(aws.Provider().ResourcesMap[sgr])
		resSGR.EXPECT().ImpliedType().Return(aws.Provider().ResourcesMap[sgr].CoreConfigSchema().ImpliedType())
		resSGR.EXPECT().ResourceInstanceObject().Return(providers.ImportedResource{
			TypeName: sgr,
			State:    stateSGR,
		}.AsInstanceObject())
		attrsSGR := make(map[string]string)
		attrsSGR["security-group-id"] = "sg-1234"
		resSGR.EXPECT().InstanceState().Return(&terraform.InstanceState{
			Attributes: attrsSGR,
		})

		err = sw.Write("aws_security_group.sg", resSG)
		require.NoError(t, err)
		err = sw.Write("aws_security_group_rule.sgrule", resSGR)
		require.NoError(t, err)

		i["sg-1234"] = "${aws_security_group.sg.id}"
		sw.Interpolate(i)

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
	t.Run("SuccessNoInterpolation", func(t *testing.T) {
		var (
			i      = make(map[string]string)
			ctrl   = gomock.NewController(t)
			b      = &bytes.Buffer{}
			sw     = state.NewWriter(b, &writer.Options{Interpolate: false})
			prv    = mock.NewProvider(ctrl)
			resSG  = mock.NewResource(ctrl)
			resSGR = mock.NewResource(ctrl)
			sg     = "aws_security_group"
			sgr    = "aws_security_group_rule"
			state  = `{
  "version": 4,
  "terraform_version": "0.13.5",
  "serial": 0,
  "lineage": "lineage",
  "outputs": {},
  "resources": [
    {
      "mode": "managed",
      "type": "aws_security_group",
      "name": "sg",
      "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
      "instances": [
        {
          "schema_version": 1,
          "attributes": {
            "id": "sg-1234",
            "name": "sg",
            "description": null,
            "egress": null,
            "ingress": null,
            "name_prefix": null,
	    "arn": null,
            "owner_id": null,
            "revoke_rules_on_delete": null,
            "timeouts": {
              "create": null,
              "delete": null
            },
	    			"tags": null,
						"tags_all": null,
            "vpc_id": null
          }
        }
      ]
    },
    {
      "mode": "managed",
      "type": "aws_security_group_rule",
      "name": "sgrule",
      "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
      "instances": [
        {
          "schema_version": 2,
          "attributes": {
            "id": "sgrule-1234",
            "security_group_id": "sg-1234",
            "cidr_blocks": null,
            "description": null,
            "from_port": null,
            "ipv6_cidr_blocks": null,
            "prefix_list_ids": null,
            "protocol": null,
            "self": null,
            "source_security_group_id": null,
            "to_port": null,
            "type": null
          }
        }
      ]
    }
  ]
}`
		)

		defer ctrl.Finish()

		sgt, err := util.HashicorpToZclonfType(aws.Provider().ResourcesMap[sg].CoreConfigSchema().ImpliedType())
		require.NoError(t, err)
		sgrt, err := util.HashicorpToZclonfType(aws.Provider().ResourcesMap[sgr].CoreConfigSchema().ImpliedType())
		require.NoError(t, err)

		stateSG, err := hcl2shim.HCL2ValueFromFlatmap(map[string]string{"id": "sg-1234", "name": "sg"}, sgt)
		stateSGR, err := hcl2shim.HCL2ValueFromFlatmap(map[string]string{"security_group_id": "sg-1234", "id": "sgrule-1234"}, sgrt)

		require.NoError(t, err)

		resSG.EXPECT().Type().Return(sg)
		resSG.EXPECT().Provider().Return(prv)
		resSG.EXPECT().TFResource().Return(aws.Provider().ResourcesMap[sg])
		resSG.EXPECT().ImpliedType().Return(aws.Provider().ResourcesMap[sg].CoreConfigSchema().ImpliedType())
		resSG.EXPECT().ResourceInstanceObject().Return(providers.ImportedResource{
			TypeName: sg,
			State:    stateSG,
		}.AsInstanceObject())

		prv.EXPECT().String().Return("aws").AnyTimes()

		resSGR.EXPECT().Type().Return(sgr)
		resSGR.EXPECT().Provider().Return(prv)
		resSGR.EXPECT().TFResource().Return(aws.Provider().ResourcesMap[sgr])
		resSGR.EXPECT().ImpliedType().Return(aws.Provider().ResourcesMap[sgr].CoreConfigSchema().ImpliedType())
		resSGR.EXPECT().ResourceInstanceObject().Return(providers.ImportedResource{
			TypeName: sgr,
			State:    stateSGR,
		}.AsInstanceObject())

		err = sw.Write("aws_security_group.sg", resSG)
		require.NoError(t, err)
		err = sw.Write("aws_security_group_rule.sgrule", resSGR)
		require.NoError(t, err)

		i["sg-1234"] = "${aws_security_group.sg.id}"
		sw.Interpolate(i)

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
