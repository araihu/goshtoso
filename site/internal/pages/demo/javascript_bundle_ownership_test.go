package demo

import (
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/head"
	siteassets "github.com/araihu/goshtoso/site/assets"
	"github.com/stretchr/testify/require"
)

func TestComponentDocsLayoutOwnsDemoBundleOutsideHeadDependencies(t *testing.T) {
	t.Parallel()

	cfg := componentDocsConfig(false)
	require.Equal(t, []string{
		siteassets.DemoBundleURL,
		assets.HTMXExtWSURL,
		assets.HTMXExtSSEURL,
	}, cfg.Interactions.RuntimeScripts)

	var page strings.Builder
	require.NoError(t, ComponentDocsLayout(
		DefaultMeta("Bundle ownership"),
		"",
		templ.Raw("<p>content</p>"),
		false,
	).Render(context.Background(), &page))
	require.Equal(t, 1, strings.Count(page.String(), siteassets.DemoBundleURL))
	require.Contains(t, page.String(), `<script src="`+siteassets.DemoBundleURL+`"></script>`)
	require.Contains(t, page.String(), `defer src="`+assets.FirstPartyBundleURL+`"`)
	require.Contains(t, page.String(), `defer src="`+assets.AlpineJSURL+`"`)

	var publicHead strings.Builder
	require.NoError(t, head.Dependencies(head.WithLocalRuntime()).Render(context.Background(), &publicHead))
	require.NotContains(t, publicHead.String(), siteassets.DemoBundleURL,
		"public head dependencies must never load demo-site JavaScript")
}
