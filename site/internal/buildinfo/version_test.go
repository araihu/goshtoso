package buildinfo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGoDocsVersionDefaultsToDevelopment(t *testing.T) {
	require.Equal(t, "development", GoDocsVersion())
}
