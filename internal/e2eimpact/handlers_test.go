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
}
