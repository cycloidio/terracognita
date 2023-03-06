package interpolator_test

import (
	"testing"

	"github.com/cycloidio/terracognita/interpolator"
	"github.com/stretchr/testify/assert"
)

func TestInterpolate(t *testing.T) {
	i := interpolator.New("azurerm")
	i.AddResourceAttributes("azurerm_virtual_machine.front", map[string]string{
		"id":    "secretid",
		"other": "ovalue",
	})
	i.AddResourceAttributes("azurerm_virtual_something.front", map[string]string{
		"id": "secretid",
	})

	s, ok := i.Interpolate("virtual_machine_id", "secretid")
	assert.Equal(t, s, "${azurerm_virtual_machine.front.id}")
	assert.True(t, ok)

	s, ok = i.Interpolate("virtual_machine_potato", "ovalue")
	assert.Equal(t, s, "${azurerm_virtual_machine.front.other}")
	assert.True(t, ok)

	s, ok = i.Interpolate("totally_random", "secretid")
	assert.Equal(t, s, "${azurerm_virtual_something.front.id}")
	assert.True(t, ok)

	s, ok = i.Interpolate("virtual_id", "secretid")
	assert.Equal(t, s, "${azurerm_virtual_machine.front.id}")
	assert.True(t, ok)

	s, ok = i.Interpolate("virtual_machine_potato", "none")
	assert.Equal(t, s, "")
	assert.False(t, ok)
}
