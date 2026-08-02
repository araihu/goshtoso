package startpages

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGettingStartedRepositoryLinkIncludesGitHubIcon(t *testing.T) {
	t.Parallel()

	var page strings.Builder
	require.NoError(t, gettingStartedContent().Render(context.Background(), &page))

	html := page.String()
	require.Contains(t, html, `href="https://github.com/araihu/goshtoso-getting-started"`)
	require.Contains(t, html, `class="size-5 shrink-0"`)
	require.Contains(t, html, `M12 0a12 12 0 0 0-3.79 23.39`)
}
