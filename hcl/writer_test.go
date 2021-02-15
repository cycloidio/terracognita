package hcl_test

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/cycloidio/mxwriter"
	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/hcl"
	"github.com/cycloidio/terracognita/writer"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHCLWriter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		hw := hcl.NewWriter(nil, &writer.Options{Interpolate: true})

		assert.Equal(t, map[string]map[string]interface{}{}, hw.Config)
	})
}

func TestHCLWriter_Write(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			mw    = mxwriter.NewMux()
			hw    = hcl.NewWriter(mw, &writer.Options{Interpolate: true})
			value = map[string]interface{}{
				"key": "value",
			}
			key = "type.name"
		)

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
				mw = mxwriter.NewMux()
				hw = hcl.NewWriter(mw, &writer.Options{Interpolate: true, Module: "s"})

				ok, err = hw.Has("type.new")
				require.NoError(t, err)
				assert.False(t, ok)
			})
		})
	})
	t.Run("ErrRequiredKey", func(t *testing.T) {
		var (
			mw = mxwriter.NewMux()
			hw = hcl.NewWriter(mw, &writer.Options{Interpolate: true})
		)

		err := hw.Write("", nil)
		assert.Equal(t, errcode.ErrWriterRequiredKey, errors.Cause(err))
	})
	t.Run("ErrRequiredValue", func(t *testing.T) {
		var (
			mw = mxwriter.NewMux()
			hw = hcl.NewWriter(mw, &writer.Options{Interpolate: true})
		)

		err := hw.Write("type.name", nil)
		assert.Equal(t, errcode.ErrWriterRequiredValue, errors.Cause(err))
	})
	t.Run("ErrInvalidKey", func(t *testing.T) {
		var (
			mw = mxwriter.NewMux()
			hw = hcl.NewWriter(mw, &writer.Options{Interpolate: true})
		)

		err := hw.Write("type.name.name", "")
		assert.Equal(t, errcode.ErrWriterInvalidKey, errors.Cause(err))

		err = hw.Write("type", "")
		assert.Equal(t, errcode.ErrWriterInvalidKey, errors.Cause(err))

		err = hw.Write("type.", "")
		assert.Equal(t, errcode.ErrWriterInvalidKey, errors.Cause(err))
	})
	t.Run("ErrAlreadyExistsKey", func(t *testing.T) {
		var (
			mw = mxwriter.NewMux()
			hw = hcl.NewWriter(mw, &writer.Options{Interpolate: true})
		)

		err := hw.Write("type.name", map[string]interface{}{})
		require.NoError(t, err)

		err = hw.Write("type.name", map[string]interface{}{})
		assert.Equal(t, errcode.ErrWriterAlreadyExistsKey, errors.Cause(err))
	})
}

func TestHCLWriter_Sync(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			mx    = mxwriter.NewMux()
			hw    = hcl.NewWriter(mx, &writer.Options{Interpolate: true})
			value = map[string]interface{}{
				"key": "value",
			}
			hcl = `
resource "type" "name" {
  key = "value"
}
`
		)

		err := hw.Write("type.name", value)
		require.NoError(t, err)

		err = hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mx)
		require.NoError(t, err)

		assert.Equal(t, strings.Join(strings.Fields(hcl), " "), strings.Join(strings.Fields(string(b)), " "))
	})
	t.Run("Module", func(t *testing.T) {
		var (
			mx    = mxwriter.NewMux()
			hw    = hcl.NewWriter(mx, &writer.Options{Interpolate: true, Module: "test"})
			value = map[string]interface{}{
				"key": "value",
			}
			value2 = map[string]interface{}{
				"key":  "value",
				"key2": "value",
			}
			hcl = `
resource "type" "name" {
  key = var.type_name_key
}

resource "type" "name2" {
  key = var.type_name2_key
	key2 = var.type_name2_key2
}

module "test" {
	# type_name2_key = "value"
	# type_name2_key2 = "value"
	# type_name_key = "value"
  source = "module-test"
}

variable "type_name2_key" {
	default = "value"
}

variable "type_name2_key2" {
	default = "value"
}

variable "type_name_key" {
	default = "value"
}
`
		)

		err := hw.Write("type.name", value)
		require.NoError(t, err)

		err = hw.Write("type.name2", value2)
		require.NoError(t, err)

		err = hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mx)
		require.NoError(t, err)

		assert.Equal(t, strings.Join(strings.Fields(hcl), " "), strings.Join(strings.Fields(string(b)), " "))
	})
	t.Run("ModuleVariables", func(t *testing.T) {
		var (
			mx    = mxwriter.NewMux()
			hw    = hcl.NewWriter(mx, &writer.Options{Interpolate: true, Module: "test", ModuleVariables: map[string]struct{}{"type.key": struct{}{}}})
			value = map[string]interface{}{
				"key": "value",
			}
			value2 = map[string]interface{}{
				"key":  "value",
				"key2": "value",
			}
			hcl = `
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
  source = "module-test"
}

variable "type_name2_key" {
	default = "value"
}

variable "type_name_key" {
	default = "value"
}
`
		)

		err := hw.Write("type.name", value)
		require.NoError(t, err)

		err = hw.Write("type.name2", value2)
		require.NoError(t, err)

		err = hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mx)
		require.NoError(t, err)

		assert.Equal(t, strings.Join(strings.Fields(hcl), " "), strings.Join(strings.Fields(string(b)), " "))
	})
	t.Run("Slice", func(t *testing.T) {
		var (
			mw    = mxwriter.NewMux()
			hw    = hcl.NewWriter(mw, &writer.Options{Interpolate: true})
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
			hcl = `
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
`
		)

		err := hw.Write("type.name", value)
		require.NoError(t, err)

		err = hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mw)
		require.NoError(t, err)

		assert.Equal(t, strings.Join(strings.Fields(hcl), " "), strings.Join(strings.Fields(string(b)), " "))
	})
	t.Run("EmptySlice", func(t *testing.T) {
		var (
			mw    = mxwriter.NewMux()
			hw    = hcl.NewWriter(mw, &writer.Options{Interpolate: true})
			value = map[string]interface{}{
				"ingress": []map[string]interface{}{
					{
						"cidr_blocks": []string{},
						"from_port":   80,
					},
				},
			}
			hcl = `
resource "type" "name" {
  ingress {
		cidr_blocks = []
		from_port = 80
	}
}
`
		)

		err := hw.Write("type.name", value)
		require.NoError(t, err)

		err = hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mw)
		require.NoError(t, err)

		assert.Equal(t, strings.Join(strings.Fields(hcl), " "), strings.Join(strings.Fields(string(b)), " "))
	})
	t.Run("NestedMap", func(t *testing.T) {
		var (
			mw    = mxwriter.NewMux()
			hw    = hcl.NewWriter(mw, &writer.Options{Interpolate: true})
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
			hcl = `
resource "type" "name" {
  ingress {
		cidr_blocks = []
		from_port {
			in = "vin"
			out = "vout"
		}
	}
}
`
		)

		err := hw.Write("type.name", value)
		require.NoError(t, err)

		err = hw.Sync()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(mw)
		require.NoError(t, err)

		assert.Equal(t, strings.Join(strings.Fields(hcl), " "), strings.Join(strings.Fields(string(b)), " "))
	})
}

func TestHCLWriter_Interpolate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			mw    = mxwriter.NewMux()
			hw    = hcl.NewWriter(mw, &writer.Options{Interpolate: true})
			value = map[string]interface{}{
				"network": "to-be-interpolated",
			}
			network = map[string]interface{}{
				"id": "interpolated",
			}
			i = make(map[string]string)
		)
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
	t.Run("SuccessAvoidInterpolaception", func(t *testing.T) {
		var (
			mw    = mxwriter.NewMux()
			hw    = hcl.NewWriter(mw, &writer.Options{Interpolate: true})
			value = map[string]interface{}{
				"network": "to-be-interpolated",
			}
			network = map[string]interface{}{
				"id":   "interpolated",
				"name": "to-be-interpolated",
			}
			i = make(map[string]string)
		)
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
			hw       = hcl.NewWriter(mw, &writer.Options{Interpolate: true})
			instance = map[string]interface{}{
				"subnet_id": "1234",
			}
			subnet = map[string]interface{}{
				"id":                "subnet-1",
				"availability_zone": "a-zone",
			}
			i = make(map[string]string)
		)
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
			hw    = hcl.NewWriter(mw, &writer.Options{Interpolate: false})
			value = map[string]interface{}{
				"network": "should-not-be-interpolated",
			}
			network = map[string]interface{}{
				"id":   "interpolated",
				"name": "to-be-interpolated",
			}
			i = make(map[string]string)
		)
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
