package cache_test

import (
	"testing"

	"github.com/cycloidio/terracognita/cache"
	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/provider"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetGet(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := cache.New()
		r := provider.NewResource("id", "", nil, nil, nil)
		err := c.Set("k", []provider.Resource{r})
		require.NoError(t, err)

		rs, err := c.Get("k")
		require.NoError(t, err)
		assert.Equal(t, []provider.Resource{r}, rs)
	})

	t.Run("ErrCacheKeyNotFound", func(t *testing.T) {
		c := cache.New()

		rs, err := c.Get("k")
		require.Nil(t, rs)
		assert.Equal(t, errcode.ErrCacheKeyNotFound, errors.Cause(err))
	})

	t.Run("ErrCacheKeyAlreadyExisting", func(t *testing.T) {
		c := cache.New()

		err := c.Set("k", nil)
		require.Nil(t, err)
		err = c.Set("k", nil)
		assert.Equal(t, errcode.ErrCacheKeyAlreadyExisting, errors.Cause(err))
	})
}
