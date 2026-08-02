package e2eimpact

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestButtonChangeSelectsButtonAndReverseConsumers(t *testing.T) {
	result := selectChanges(context.Background(), repositoryRoot(t), []Change{{
		Status: "M", OldPath: "components/button/button.templ", NewPath: "components/button/button.templ",
	}})
	require.Equal(t, "focused", result.Mode, result.Reasons)
	require.Contains(t, result.Tags, "button")
	require.Contains(t, result.Tags, "actiongroup")
	require.Contains(t, result.Tags, "example_chat")
	require.NotContains(t, result.Tags, "rating")
}

func TestRenameAndDeleteAlwaysSelectFull(t *testing.T) {
	for _, change := range []Change{
		{Status: "D", OldPath: "components/button/types.go", NewPath: "components/button/types.go"},
		{Status: "R100", OldPath: "components/button/types.go", NewPath: "components/button/config.go"},
	} {
		result := selectChanges(context.Background(), repositoryRoot(t), []Change{change})
		require.Equal(t, "full", result.Mode)
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "../.."))
}
