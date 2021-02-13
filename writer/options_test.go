package writer_test

import (
	"testing"

	"github.com/cycloidio/terracognita/writer"
	"github.com/stretchr/testify/assert"
)

func TestOptionsHasModule(t *testing.T) {
	opt := writer.Options{Module: "a"}
	assert.True(t, opt.HasModule())

	opt.Module = ""
	assert.False(t, opt.HasModule())
}
