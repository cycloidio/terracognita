package hcl_test

import (
	"bytes"
	"strings"
	"testing"

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

		assert.Equal(t, map[string]interface{}{
			"resource": make(map[string]map[string]interface{}),
		}, hw.Config)
	})
}

func TestHCLWriter_Write(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			b     = &bytes.Buffer{}
			hw    = hcl.NewWriter(b, &writer.Options{Interpolate: true})
			value = map[string]interface{}{
				"key": "value",
			}
			key = "type.name"
		)

		err := hw.Write(key, value)
		require.NoError(t, err)

		assert.Equal(t, map[string]interface{}{
			"resource": map[string]map[string]interface{}{
				"type": map[string]interface{}{
					"name": map[string]interface{}{
						"key": "value",
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
		})
	})
	t.Run("ErrRequiredKey", func(t *testing.T) {
		var (
			b  = &bytes.Buffer{}
			hw = hcl.NewWriter(b, &writer.Options{Interpolate: true})
		)

		err := hw.Write("", nil)
		assert.Equal(t, errcode.ErrWriterRequiredKey, errors.Cause(err))
	})
	t.Run("ErrRequiredValue", func(t *testing.T) {
		var (
			b  = &bytes.Buffer{}
			hw = hcl.NewWriter(b, &writer.Options{Interpolate: true})
		)

		err := hw.Write("type.name", nil)
		assert.Equal(t, errcode.ErrWriterRequiredValue, errors.Cause(err))
	})
	t.Run("ErrInvalidKey", func(t *testing.T) {
		var (
			b  = &bytes.Buffer{}
			hw = hcl.NewWriter(b, &writer.Options{Interpolate: true})
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
			b  = &bytes.Buffer{}
			hw = hcl.NewWriter(b, &writer.Options{Interpolate: true})
		)

		err := hw.Write("type.name", "")
		require.NoError(t, err)

		err = hw.Write("type.name", "")
		assert.Equal(t, errcode.ErrWriterAlreadyExistsKey, errors.Cause(err))
	})
}

func TestHCLWriter_Sync(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			b     = &bytes.Buffer{}
			hw    = hcl.NewWriter(b, &writer.Options{Interpolate: true})
			value = map[string]interface{}{
				"key": "value",
			}
			hcl = `resource "type" "name" {
  key = "value"
}
`
		)

		err := hw.Write("type.name", value)
		require.NoError(t, err)

		err = hw.Sync()
		require.NoError(t, err)

		assert.Equal(t, hcl, b.String())
	})
}

func TestHCLWriter_Interpolate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			b     = &bytes.Buffer{}
			hw    = hcl.NewWriter(b, &writer.Options{Interpolate: true})
			value = map[string]interface{}{
				"network": "to-be-interpolated",
			}
			network = map[string]interface{}{
				"id": "interpolated",
			}
			i   = make(map[string]string)
			hcl = `resource "aType" "aName" {
  id = "interpolated"
}

resource "type" "name" {
  network = "${aType.aName.id}"
}
`
		)
		i["to-be-interpolated"] = "${aType.aName.id}"
		hw.Write("type.name", value)
		hw.Write("aType.aName", network)

		hw.Interpolate(i)
		hw.Sync()

		assert.Equal(t, hcl, b.String())
	})
	t.Run("SuccessAvoidInterpolaception", func(t *testing.T) {
		var (
			b     = &bytes.Buffer{}
			hw    = hcl.NewWriter(b, &writer.Options{Interpolate: true})
			value = map[string]interface{}{
				"network": "to-be-interpolated",
			}
			network = map[string]interface{}{
				"id":   "interpolated",
				"name": "to-be-interpolated",
			}
			i   = make(map[string]string)
			hcl = `resource "aType" "aName" {
  id   = "interpolated"
  name = "to-be-interpolated"
}

resource "type" "name" {
  network = "${aType.aName.id}"
}
`
		)
		i["to-be-interpolated"] = "${aType.aName.id}"
		hw.Write("type.name", value)
		hw.Write("aType.aName", network)

		hw.Interpolate(i)
		hw.Sync()

		assert.Equal(t, hcl, b.String())
	})
	t.Run("SuccessMutualInterpolation", func(t *testing.T) {
		var (
			b        = &bytes.Buffer{}
			hw       = hcl.NewWriter(b, &writer.Options{Interpolate: true})
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
		hw.Sync()
		// the config relies on map, since the order is pseudo random
		// we just assert that we only have one "$" which means: there is
		// only one interpolation
		assert.Equal(t, 1, strings.Count(b.String(), "$"))
	})
}
