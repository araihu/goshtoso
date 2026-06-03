package combobox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistry_DuplicateIDPanics(t *testing.T) {
	resetRegistry()
	registerID("status")
	assert.Panics(t, func() { registerID("status") }, "re-registering same ID must panic")
}

func TestRegistry_DistinctIDsAllowed(t *testing.T) {
	resetRegistry()
	registerID("status")
	registerID("provider")
	// no panic
}
