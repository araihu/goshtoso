package e2econstraints

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestActiveE2ECommandsPassExplicitSuiteTags(t *testing.T) {
	findings, err := FindBareE2ECommands(filepath.Clean(filepath.Join(siteRoot(t), "..")))
	require.NoError(t, err)
	require.Empty(t, findings, "active E2E commands must pass -tags=e2e,full or a focused identity set")
}

func TestActiveCommandFileIncludesLowercaseJustfile(t *testing.T) {
	require.True(t, activeCommandFile("justfile"))
}
