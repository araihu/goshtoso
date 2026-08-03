package e2eimpact

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEveryProductionServerFileHasCheckedOwnership(t *testing.T) {
	manifest, err := loadIdentityManifest(repositoryRoot(t))
	require.NoError(t, err)
	require.NoError(t, validateHandlerOwnership(repositoryRoot(t), manifest.known))
}

func TestHandlerMappingsSelectKnownFocusedIdentities(t *testing.T) {
	identities, ok := identitiesForHandler("site/internal/server/form_validation_handler.go")
	require.True(t, ok)
	require.Equal(t, []string{"form", "textinput"}, identities)
	_, ok = identitiesForHandler("site/internal/server/server.go")
	require.False(t, ok)

	for _, path := range []string{
		"site/internal/server/table_fragments.templ",
		"site/internal/server/table_fragments_templ.go",
	} {
		identities, ok = identitiesForHandler(path)
		require.True(t, ok)
		require.Equal(t, []string{"table"}, identities)
	}
}
