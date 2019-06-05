package writer_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/writer"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTFStateWriter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		sw := writer.NewTFStateWriter(nil)

		assert.Equal(t, make(map[string]*terraform.ResourceState), sw.Config)
	})
}

func TestTFStateWriter_Write(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			b     = &bytes.Buffer{}
			sw    = writer.NewTFStateWriter(b)
			value = &terraform.ResourceState{
				Type: "my_type",
			}
		)

		err := sw.Write("type.name", value)
		require.NoError(t, err)

		assert.Equal(t, map[string]*terraform.ResourceState{
			"type.name": &terraform.ResourceState{Type: "my_type"},
		}, sw.Config)
	})
	t.Run("ErrRequiredKey", func(t *testing.T) {
		sw := writer.NewTFStateWriter(nil)

		err := sw.Write("", nil)
		assert.Equal(t, errcode.ErrWriterRequiredKey, errors.Cause(err))
	})
	t.Run("ErrRequiredValue", func(t *testing.T) {
		sw := writer.NewTFStateWriter(nil)

		err := sw.Write("key", nil)
		assert.Equal(t, errcode.ErrWriterRequiredValue, errors.Cause(err))
	})
	t.Run("ErrAlreadyExistsKey", func(t *testing.T) {
		var (
			sw    = writer.NewTFStateWriter(nil)
			value = &terraform.ResourceState{
				Type: "my_type",
			}
		)
		err := sw.Write("key", value)
		require.NoError(t, err)

		err = sw.Write("key", value)
		assert.Equal(t, errcode.ErrWriterAlreadyExistsKey, errors.Cause(err))
	})
	t.Run("ErrInvalidTypeValue", func(t *testing.T) {
		sw := writer.NewTFStateWriter(nil)

		err := sw.Write("key", 0)
		assert.Equal(t, errcode.ErrWriterInvalidTypeValue, errors.Cause(err))
	})
}

func TestTFStateWriter_Sync(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			b     = &bytes.Buffer{}
			sw    = writer.NewTFStateWriter(b)
			value = &terraform.ResourceState{
				Type: "my_type",
			}
			state = `{
  "version": 3,
  "serial": 0,
  "lineage": "lineage",
  "modules": [
    {
      "path": [
        "root"
      ],
      "outputs": {},
      "resources": {
        "type.name": {
          "type": "my_type",
          "depends_on": [],
          "primary": {
            "id": "",
            "attributes": {},
            "meta": {},
            "tainted": false
          },
          "deposed": [],
          "provider": ""
        }
      },
      "depends_on": []
    }
  ]
}`
		)

		err := sw.Write("type.name", value)
		require.NoError(t, err)

		err = sw.Sync()
		require.NoError(t, err)

		var st map[string]interface{}
		err = json.Unmarshal(b.Bytes(), &st)
		require.NoError(t, err)

		st["lineage"] = "lineage"

		var est map[string]interface{}
		err = json.Unmarshal([]byte(state), &est)

		assert.Equal(t, est, st)
	})
}
