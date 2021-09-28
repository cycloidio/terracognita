package hcl_test

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/cycloidio/mxwriter"
	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/hcl"
	"github.com/cycloidio/terracognita/mock"
	"github.com/cycloidio/terracognita/writer"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terraform-providers/terraform-provider-aws/aws"
)

func TestNewHCLWriter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			p    = mock.NewProvider(ctrl)
		)
		p.EXPECT().String().Return("aws").Times(3)
		p.EXPECT().Source().Return("hashicorp/aws")
		p.EXPECT().TFProvider().Return(aws.Provider())
		p.EXPECT().Configuration().Return(map[string]interface{}{
			"region": "eu-west-1",
		})

		hw := hcl.NewWriter(nil, p, &writer.Options{HCLProviderBlock: true})
		assert.Equal(t, map[string]map[string]interface{}{
			"hcl": map[string]interface{}{
				"provider": map[string]interface{}{
					"aws": map[string]interface{}{
						"region": "${var.region}",
					},
				},
				"resource": map[string]map[string]interface{}{},
				"terraform": map[string]interface{}{
					"required_providers": map[string]interface{}{
						"=tc=aws": map[string]interface{}{
							"source": "hashicorp/aws",
						},
					},
					"required_version": ">= 1.0",
				},
				"variable": map[string]interface{}{
					"region": map[string]interface{}{
						"default": "eu-west-1",
					},
				},
			},
		}, hw.Config)
	})
	t.Run("SuccessWithoutProviderBLock", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			p    = mock.NewProvider(ctrl)
		)
		p.EXPECT().String().Return("aws")
		p.EXPECT().Source().Return("hashicorp/aws")

		hw := hcl.NewWriter(nil, p, &writer.Options{})
		assert.Equal(t, map[string]map[string]interface{}{
			"hcl": map[string]interface{}{
				"resource": map[string]map[string]interface{}{},
				"terraform": map[string]interface{}{
					"required_providers": map[string]interface{}{
						"=tc=aws": map[string]interface{}{
							"source": "hashicorp/aws",
						},
					},
					"required_version": ">= 1.0",
				},
			},
		}, hw.Config)
	})
	t.Run("SuccessWithModule", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			p    = mock.NewProvider(ctrl)
		)
		p.EXPECT().String().Return("aws")
		p.EXPECT().Source().Return("hashicorp/aws")

		hw := hcl.NewWriter(nil, p, &writer.Options{Module: "my-module"})
		assert.Equal(t, map[string]map[string]interface{}{
			"tc_module": map[string]interface{}{
				"module": map[string]interface{}{
					"my-module": map[string]interface{}{
						"source": "./module-my-module",
					},
				},
				"terraform": map[string]interface{}{
					"required_providers": map[string]interface{}{
						"=tc=aws": map[string]interface{}{
							"source": "hashicorp/aws",
						},
					},
					"required_version": ">= 1.0",
				},
			},
		}, hw.Config)
	})
}

func TestHCLWriter_Write(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			mw    = mxwriter.NewMux()
			value = map[string]interface{}{
				"key": "value",
			}
			key  = "type.name"
			ctrl = gomock.NewController(t)
			p    = mock.NewProvider(ctrl)
		)
		defer ctrl.Finish()

		p.EXPECT().String().Return("aws")
		p.EXPECT().Source().Return("hashicorp/aws")

		hw := hcl.NewWriter(mw, p, &writer.Options{Interpolate: true})

		err := hw.Write(key, value)
		require.NoError(t, err)

		assert.Equal(t, map[string]map[string]interface{}{
			"hcl": map[string]interface{}{
				"resource": map[string]map[string]interface{}{
					"type": map[string]interface{}{
						"name": map[string]interface{}{
							"key": "value",
						},
					},
				},
				"terraform": map[string]interface{}{
					"required_providers": map[string]interface{}{
						"=tc=aws": map[string]interface{}{
							"source": "hashicorp/aws",
						},
					},
					"required_version": ">= 1.0",
				},
			},
		}, hw.Config)
		t.Run("Has", func(t *testing.T) {
			ok, err := hw.Has(key)
			require.NoError(t, err)
			assert.True(t, ok)

			ok, err = hw.Has("type.new")
			require.NoError(t, err)
			assert.False(t, ok)

			t.Run("Empty", func(t *testing.T) {
				var (
					ctrl = gomock.NewController(t)
					p    = mock.NewProvider(ctrl)
				)

				p.EXPECT().String().Return("aws")
				p.EXPECT().Source().Return("hashicorp/aws")

				mw = mxwriter.NewMux()
				hw := hcl.NewWriter(mw, p, &writer.Options{Interpolate: true, Module: "s"})

				ok, err = hw.Has("type.new")
				require.NoError(t, err)
				assert.False(t, ok)
			})
		})
	})
	t.Run("ErrRequiredKey", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			p    = mock.NewProvider(ctrl)
			mw   = mxwriter.NewMux()
		)
		p.EXPECT().String().Return("aws")
		p.EXPECT().Source().Return("hashicorp/aws")

		hw := hcl.NewWriter(mw, p, &writer.Options{Interpolate: true})

		err := hw.Write("", nil)
		assert.Equal(t, errcode.ErrWriterRequiredKey, errors.Cause(err))
	})
	t.Run("ErrRequiredValue", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			p    = mock.NewProvider(ctrl)
			mw   = mxwriter.NewMux()
		)
		p.EXPECT().String().Return("aws")
		p.EXPECT().Source().Return("hashicorp/aws")

		hw := hcl.NewWriter(mw, p, &writer.Options{Interpolate: true})

		err := hw.Write("type.name", nil)
		assert.Equal(t, errcode.ErrWriterRequiredValue, errors.Cause(err))
	})
	t.Run("ErrInvalidKey", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			p    = mock.NewProvider(ctrl)
			mw   = mxwriter.NewMux()
		)
		p.EXPECT().String().Return("aws")
		p.EXPECT().Source().Return("hashicorp/aws")

		hw := hcl.NewWriter(mw, p, &writer.Options{Interpolate: true})

		err := hw.Write("type.name.name", "")
		assert.Equal(t, errcode.ErrWriterInvalidKey, errors.Cause(err))

		err = hw.Write("type", "")
		assert.Equal(t, errcode.ErrWriterInvalidKey, errors.Cause(err))

		err = hw.Write("type.", "")
		assert.Equal(t, errcode.ErrWriterInvalidKey, errors.Cause(err))
	})
	t.Run("ErrAlreadyExistsKey", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			p    = mock.NewProvider(ctrl)
			mw   = mxwriter.NewMux()
		)
		p.EXPECT().String().Return("aws")
		p.EXPECT().Source().Return("hashicorp/aws")

		hw := hcl.NewWriter(mw, p, &writer.Options{Interpolate: true})

		err := hw.Write("type.name", map[string]interface{}{})
		require.NoError(t, err)

		err = hw.Write("type.name", map[string]interface{}{})
		assert.Equal(t, errcode.ErrWriterAlreadyExistsKey, errors.Cause(err))
	})
}

func TestHCLWriter_Sync(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			ctrl  = gomock.NewController(t)
			p     = mock.NewProvider(ctrl)
			mx    = mxwriter.NewMux()
			value = map[string]interface{}{
				"key":         "value",
				"tc_category": "some-category",
			}
			ehcl = `
provider "aws" {
	region = var.region
}

terraform {
	required_providers {
		aws = {
			source = "hashicorp/aws"
		}
	}
	required_version = ">= 1.0"
}

variable "region" {
	default = "eu-west-1"
}

resource "type" "name" {
  key = "value"
}

`
		)

		p.EXPECT().String().Return("aws").Times(3)
		p.EXPECT().Source().Return("hashicorp/aws")
		p.EXPECT().TFProvider().Return(aws.Provider())
		p.EXPECT().Configuration().Return(map[string]interface{}{
			"region": "eu-west-1",
		})

		hw := hcl.NewWriter(mx, p, &writer.Options{HCLProviderBlock: true, Interpolate: true})

		err := hw.Write("type.name", value)
		require.NoError(t, err)

		err = hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mx)
		require.NoError(t, err)

		assert.Equal(t, strings.Join(strings.Fields(ehcl), " "), strings.Join(strings.Fields(string(b)), " "))
	})
	t.Run("SuccessWithoutProviderBlock", func(t *testing.T) {
		var (
			ctrl  = gomock.NewController(t)
			p     = mock.NewProvider(ctrl)
			mx    = mxwriter.NewMux()
			value = map[string]interface{}{
				"key":         "value",
				"tc_category": "some-category",
			}
			ehcl = `
terraform {
	required_providers {
		aws = {
			source = "hashicorp/aws"
		}
	}
	required_version = ">= 1.0"
}

resource "type" "name" {
  key = "value"
}

`
		)

		p.EXPECT().String().Return("aws")
		p.EXPECT().Source().Return("hashicorp/aws")

		hw := hcl.NewWriter(mx, p, &writer.Options{Interpolate: true})

		err := hw.Write("type.name", value)
		require.NoError(t, err)

		err = hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mx)
		require.NoError(t, err)

		assert.Equal(t, strings.Join(strings.Fields(ehcl), " "), strings.Join(strings.Fields(string(b)), " "))
	})
	t.Run("Module", func(t *testing.T) {
		var (
			ctrl  = gomock.NewController(t)
			p     = mock.NewProvider(ctrl)
			mx    = mxwriter.NewMux()
			value = map[string]interface{}{
				"key": "value",
			}
			value2 = map[string]interface{}{
				"key":  "value",
				"key2": "value",
				"key3": []interface{}{},
				"key4": map[string]interface{}{
					"nested:key:4": "value4",
				},
			}
			ehcl = `
resource "type" "name" {
  key = var.type_name_key
}

resource "type" "name2" {
  key = var.type_name2_key
	key2 = var.type_name2_key2
	key3 = var.type_name2_key3
	key4 {
		"nested:key:4" = var.type_name2_key4_nested_key_4
	}
}

module "test" {
	# type_name2_key = "value"
	# type_name2_key2 = "value"
	# type_name2_key3 = []
	# type_name2_key4_nested_key_4 = "value4"
	# type_name_key = "value"
  source = "./module-test"
}

provider "aws" {
	region = var.region
}

terraform {
	required_providers {
		aws = {
			source = "hashicorp/aws"
		}
	}
	required_version = ">= 1.0"
}

variable "region" {
	default = "eu-west-1"
}

variable "type_name2_key" {
	default = "value"
}

variable "type_name2_key2" {
	default = "value"
}

variable "type_name2_key3" {
	default = []
}

variable "type_name2_key4_nested_key_4" {
	default = "value4"
}

variable "type_name_key" {
	default = "value"
}
`
		)
		p.EXPECT().String().Return("aws").Times(3)
		p.EXPECT().Source().Return("hashicorp/aws")
		p.EXPECT().TFProvider().Return(aws.Provider())
		p.EXPECT().Configuration().Return(map[string]interface{}{
			"region": "eu-west-1",
		})

		hw := hcl.NewWriter(mx, p, &writer.Options{Interpolate: true, HCLProviderBlock: true, Module: "test"})

		err := hw.Write("type.name", value)
		require.NoError(t, err)

		err = hw.Write("type.name2", value2)
		require.NoError(t, err)

		err = hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mx)
		require.NoError(t, err)

		assert.Equal(t, strings.Join(strings.Fields(ehcl), " "), strings.Join(strings.Fields(string(b)), " "))
	})
	t.Run("ModuleVariables", func(t *testing.T) {
		var (
			ctrl  = gomock.NewController(t)
			p     = mock.NewProvider(ctrl)
			mx    = mxwriter.NewMux()
			value = map[string]interface{}{
				"key": "value",
			}
			value2 = map[string]interface{}{
				"key":  "value",
				"key2": "value",
			}
			ehcl = `
resource "type" "name" {
  key = var.type_name_key
}

resource "type" "name2" {
  key = var.type_name2_key
	key2 = "value"
}

module "test" {
	# type_name2_key = "value"
	# type_name_key = "value"
  source = "./module-test"
}

provider "aws" {
	region = var.region
}

terraform {
	required_providers {
		aws = {
			source = "hashicorp/aws"
		}
	}
	required_version = ">= 1.0"
}

variable "region" {
	default = "eu-west-1"
}

variable "type_name2_key" {
	default = "value"
}

variable "type_name_key" {
	default = "value"
}
`
		)
		p.EXPECT().String().Return("aws").Times(3)
		p.EXPECT().Source().Return("hashicorp/aws")
		p.EXPECT().TFProvider().Return(aws.Provider())
		p.EXPECT().Configuration().Return(map[string]interface{}{
			"region": "eu-west-1",
		})

		hw := hcl.NewWriter(mx, p, &writer.Options{Interpolate: true, HCLProviderBlock: true, Module: "test", ModuleVariables: map[string]struct{}{"type.key": struct{}{}}})

		err := hw.Write("type.name", value)
		require.NoError(t, err)

		err = hw.Write("type.name2", value2)
		require.NoError(t, err)

		err = hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mx)
		require.NoError(t, err)

		assert.Equal(t, strings.Join(strings.Fields(ehcl), " "), strings.Join(strings.Fields(string(b)), " "))
	})
	t.Run("Slice", func(t *testing.T) {
		var (
			ctrl  = gomock.NewController(t)
			p     = mock.NewProvider(ctrl)
			mw    = mxwriter.NewMux()
			value = map[string]interface{}{
				"ingress": []map[string]interface{}{
					{
						"cidr_blocks": []string{"0.0.0.0/0"},
						"from_port":   80,
					},
					{
						"cidr_blocks": []string{"0.0.0.0/1"},
						"from_port":   81,
					},
				},
			}
			ehcl = `
resource "type" "name" {
  ingress {
		cidr_blocks = ["0.0.0.0/0"]
		from_port = 80
	}

  ingress {
		cidr_blocks = ["0.0.0.0/1"]
		from_port = 81
	}
}

terraform {
	required_providers {
		aws = {
			source = "hashicorp/aws"
		}
	}
	required_version = ">= 1.0"
}
`
		)
		p.EXPECT().String().Return("aws")
		p.EXPECT().Source().Return("hashicorp/aws")

		hw := hcl.NewWriter(mw, p, &writer.Options{Interpolate: true})

		err := hw.Write("type.name", value)
		require.NoError(t, err)

		err = hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mw)
		require.NoError(t, err)

		assert.Equal(t, strings.Join(strings.Fields(ehcl), " "), strings.Join(strings.Fields(string(b)), " "))
	})
	t.Run("EmptySlice", func(t *testing.T) {
		var (
			ctrl  = gomock.NewController(t)
			p     = mock.NewProvider(ctrl)
			mw    = mxwriter.NewMux()
			value = map[string]interface{}{
				"ingress": []map[string]interface{}{
					{
						"cidr_blocks": []string{},
						"from_port":   80,
					},
				},
			}
			ehcl = `
resource "type" "name" {
  ingress {
		cidr_blocks = []
		from_port = 80
	}
}

terraform {
	required_providers {
		aws = {
			source = "hashicorp/aws"
		}
	}
	required_version = ">= 1.0"
}
`
		)
		p.EXPECT().String().Return("aws")
		p.EXPECT().Source().Return("hashicorp/aws")

		hw := hcl.NewWriter(mw, p, &writer.Options{Interpolate: true})

		err := hw.Write("type.name", value)
		require.NoError(t, err)

		err = hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mw)
		require.NoError(t, err)

		assert.Equal(t, strings.Join(strings.Fields(ehcl), " "), strings.Join(strings.Fields(string(b)), " "))
	})
	t.Run("NestedMap", func(t *testing.T) {
		var (
			ctrl  = gomock.NewController(t)
			p     = mock.NewProvider(ctrl)
			mw    = mxwriter.NewMux()
			value = map[string]interface{}{
				"ingress": []map[string]interface{}{
					{
						"cidr_blocks": []string{},
						"from_port": map[string]interface{}{
							"in":  "vin",
							"out": "vout",
						},
					},
				},
			}
			ehcl = `
resource "type" "name" {
  ingress {
		cidr_blocks = []
		from_port {
			in = "vin"
			out = "vout"
		}
	}
}

terraform {
	required_providers {
		aws = {
			source = "hashicorp/aws"
		}
	}
	required_version = ">= 1.0"
}
`
		)
		p.EXPECT().String().Return("aws")
		p.EXPECT().Source().Return("hashicorp/aws")

		hw := hcl.NewWriter(mw, p, &writer.Options{Interpolate: true})

		err := hw.Write("type.name", value)
		require.NoError(t, err)

		err = hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mw)
		require.NoError(t, err)

		assert.Equal(t, strings.Join(strings.Fields(ehcl), " "), strings.Join(strings.Fields(string(b)), " "))
	})
}

func TestHCLWriter_Interpolate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			mw    = mxwriter.NewMux()
			ctrl  = gomock.NewController(t)
			p     = mock.NewProvider(ctrl)
			value = map[string]interface{}{
				"network": "to-be-interpolated",
			}
			network = map[string]interface{}{
				"id": "interpolated",
			}
			i = make(map[string]string)
		)
		p.EXPECT().String().Return("aws")
		p.EXPECT().Source().Return("hashicorp/aws")

		hw := hcl.NewWriter(mw, p, &writer.Options{Interpolate: true})
		i["to-be-interpolated"] = "${aType.aName.id}"
		hw.Write("type.name", value)
		hw.Write("aType.aName", network)

		hw.Interpolate(i)

		err := hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mw)
		require.NoError(t, err)

		assert.Contains(t, string(b), "network = aType.aName.id")
	})
	t.Run("SuccessWithModule", func(t *testing.T) {
		var (
			mw    = mxwriter.NewMux()
			ctrl  = gomock.NewController(t)
			p     = mock.NewProvider(ctrl)
			value = map[string]interface{}{
				"network": "to-be-interpolated",
			}
			network = map[string]interface{}{
				"id": "interpolated",
			}
			i = make(map[string]string)
		)
		p.EXPECT().String().Return("aws")
		p.EXPECT().Source().Return("hashicorp/aws")

		hw := hcl.NewWriter(mw, p, &writer.Options{Module: "test", Interpolate: true})
		i["to-be-interpolated"] = "${aType.aName.id}"
		hw.Write("type.name", value)
		hw.Write("aType.aName", network)

		hw.Interpolate(i)

		err := hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mw)
		require.NoError(t, err)

		assert.NotContains(t, string(b), "network = aType.aName.id")
	})
	t.Run("SuccessWithModuleVariablesNotSelected", func(t *testing.T) {
		var (
			mw    = mxwriter.NewMux()
			ctrl  = gomock.NewController(t)
			p     = mock.NewProvider(ctrl)
			value = map[string]interface{}{
				"network": "to-be-interpolated",
			}
			network = map[string]interface{}{
				"id": "interpolated",
			}
			i = make(map[string]string)
		)
		p.EXPECT().String().Return("aws")
		p.EXPECT().Source().Return("hashicorp/aws")

		hw := hcl.NewWriter(mw, p, &writer.Options{Module: "test", ModuleVariables: map[string]struct{}{"type.name": struct{}{}}, Interpolate: true})
		i["to-be-interpolated"] = "${aType.aName.id}"
		hw.Write("type.name", value)
		hw.Write("aType.aName", network)

		hw.Interpolate(i)

		err := hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mw)
		require.NoError(t, err)

		assert.Contains(t, string(b), "network = aType.aName.id")
	})
	t.Run("SuccessWithModuleVariablesSelected", func(t *testing.T) {
		var (
			mw    = mxwriter.NewMux()
			ctrl  = gomock.NewController(t)
			p     = mock.NewProvider(ctrl)
			value = map[string]interface{}{
				"network": "to-be-interpolated",
			}
			network = map[string]interface{}{
				"id": "interpolated",
			}
			i = make(map[string]string)
		)
		p.EXPECT().String().Return("aws")
		p.EXPECT().Source().Return("hashicorp/aws")

		hw := hcl.NewWriter(mw, p, &writer.Options{Module: "test", ModuleVariables: map[string]struct{}{"type.network": struct{}{}}, Interpolate: true})
		i["to-be-interpolated"] = "${aType.aName.id}"
		hw.Write("type.name", value)
		hw.Write("aType.aName", network)

		hw.Interpolate(i)

		err := hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mw)
		require.NoError(t, err)

		assert.NotContains(t, string(b), "network = aType.aName.id")
	})
	t.Run("SuccessAvoidInterpolaception", func(t *testing.T) {
		var (
			mw    = mxwriter.NewMux()
			ctrl  = gomock.NewController(t)
			p     = mock.NewProvider(ctrl)
			value = map[string]interface{}{
				"network": "to-be-interpolated",
			}
			network = map[string]interface{}{
				"id":   "interpolated",
				"name": "to-be-interpolated",
			}
			i = make(map[string]string)
		)
		p.EXPECT().String().Return("aws")
		p.EXPECT().Source().Return("hashicorp/aws")

		hw := hcl.NewWriter(mw, p, &writer.Options{Interpolate: true})
		i["to-be-interpolated"] = "${aType.aName.id}"
		hw.Write("type.name", value)
		hw.Write("aType.aName", network)

		hw.Interpolate(i)

		err := hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mw)
		require.NoError(t, err)

		assert.Contains(t, string(b), "to-be-interpolated")
	})
	t.Run("SuccessMutualInterpolation", func(t *testing.T) {
		var (
			mw       = mxwriter.NewMux()
			ctrl     = gomock.NewController(t)
			p        = mock.NewProvider(ctrl)
			instance = map[string]interface{}{
				"subnet_id": "1234",
			}
			subnet = map[string]interface{}{
				"id":                "subnet-1",
				"availability_zone": "a-zone",
			}
			i = make(map[string]string)
		)
		p.EXPECT().String().Return("aws")
		p.EXPECT().Source().Return("hashicorp/aws")

		hw := hcl.NewWriter(mw, p, &writer.Options{Interpolate: true})
		i["a-zone"] = "${aws_instance.instance.availability_zone}"
		i["1234"] = "${aws_subnet.subnet.id}"
		hw.Write("aws_subnet.subnet", subnet)
		hw.Write("aws_instance.instance", instance)

		hw.Interpolate(i)

		err := hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mw)
		require.NoError(t, err)

		// the only way to assert that there is one interpolation is to
		// check if we have exactly one value starting by `aws_`
		assert.Equal(t, 1, strings.Count(string(b), "= aws_"))
	})
	t.Run("SuccessNoInterpolation", func(t *testing.T) {
		var (
			mw    = mxwriter.NewMux()
			ctrl  = gomock.NewController(t)
			p     = mock.NewProvider(ctrl)
			value = map[string]interface{}{
				"network": "should-not-be-interpolated",
			}
			network = map[string]interface{}{
				"id":   "interpolated",
				"name": "to-be-interpolated",
			}
			i = make(map[string]string)
		)
		p.EXPECT().String().Return("aws")
		p.EXPECT().Source().Return("hashicorp/aws")

		hw := hcl.NewWriter(mw, p, &writer.Options{Interpolate: false})
		i["should-not-be-interpolated"] = "${aType.aName.id}"
		hw.Write("type.name", value)
		hw.Write("aType.aName", network)

		hw.Interpolate(i)

		err := hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mw)
		require.NoError(t, err)

		assert.Contains(t, string(b), "network = \"should-not-be-interpolated\"")
	})
}
