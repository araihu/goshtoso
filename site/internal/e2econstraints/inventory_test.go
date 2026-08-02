package e2econstraints

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestE2ESharedDeclarationsLiveInSupportFiles(t *testing.T) {
	findings, err := FindCrossFileDeclarations(siteRoot(t))
	require.NoError(t, err)
	require.Empty(t, findings, "move cross-file declarations into *_support_test.go files")
}

func siteRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "../.."))
}
