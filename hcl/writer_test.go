package hcl_test

import (
	"bytes"
	"testing"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/hcl"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHCLWriter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		hw := hcl.NewWriter(nil)

		assert.Equal(t, map[string]interface{}{
			"resource": make(map[string]map[string]interface{}),
		}, hw.Config)
	})
}

func TestHCLWriter_Write(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			b     = &bytes.Buffer{}
			hw    = hcl.NewWriter(b)
			value = map[string]interface{}{
				"key": "value",
			}
		)

		err := hw.Write("type.name", value)
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
	})
	t.Run("ErrRequiredKey", func(t *testing.T) {
		var (
			b  = &bytes.Buffer{}
			hw = hcl.NewWriter(b)
		)

		err := hw.Write("", nil)
		assert.Equal(t, errcode.ErrWriterRequiredKey, errors.Cause(err))
	})
	t.Run("ErrRequiredValue", func(t *testing.T) {
		var (
			b  = &bytes.Buffer{}
			hw = hcl.NewWriter(b)
		)

		err := hw.Write("type.name", nil)
		assert.Equal(t, errcode.ErrWriterRequiredValue, errors.Cause(err))
	})
	t.Run("ErrInvalidKey", func(t *testing.T) {
		var (
			b  = &bytes.Buffer{}
			hw = hcl.NewWriter(b)
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
			hw = hcl.NewWriter(b)
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
			hw    = hcl.NewWriter(b)
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
			hw    = hcl.NewWriter(b)
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
			hw    = hcl.NewWriter(b)
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
}
