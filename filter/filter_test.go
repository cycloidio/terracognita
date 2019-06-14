package filter_test

import (
	"testing"

	"github.com/cycloidio/terracognita/filter"
	"github.com/stretchr/testify/assert"
)

func TestIsExcluded(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		f := filter.Filter{Exclude: []string{"a", "b"}}
		assert.True(t, f.IsExcluded("a"))
	})
	t.Run("False", func(t *testing.T) {
		f := filter.Filter{Exclude: []string{"a", "b"}}
		assert.False(t, f.IsExcluded("c"))
	})
}
