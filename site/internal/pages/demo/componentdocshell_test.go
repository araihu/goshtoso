package demo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComponentDocsBuildBadgeMapsDevelopmentBuild(t *testing.T) {
	t.Parallel()
	badge := componentDocsBuildBadge("development")
	require.NotNil(t, badge)
	require.Equal(t, "dev", badge.Label)
	require.Equal(t, "Development build", badge.AriaLabel)
	require.Empty(t, badge.Href)
}

func TestComponentDocsBuildBadgeMapsReleaseBuild(t *testing.T) {
	t.Parallel()
	badge := componentDocsBuildBadge("v1.2.3")
	require.NotNil(t, badge)
	require.Equal(t, "v1.2.3", badge.Label)
	require.Equal(t, "Goshtoso release v1.2.3", badge.AriaLabel)
	require.Equal(t, "https://github.com/araihu/goshtoso/releases/tag/v1.2.3", badge.Href)
}

func TestComponentDocsBuildBadgeRejectsUnknownVersion(t *testing.T) {
	t.Parallel()
	require.Nil(t, componentDocsBuildBadge("main"))
}
