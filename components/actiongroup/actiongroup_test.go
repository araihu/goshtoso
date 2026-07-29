package actiongroup

import (
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components"
	"github.com/araihu/goshtoso/components/dropdown"
	"github.com/stretchr/testify/require"
)

func renderHTML(t *testing.T, component templ.Component) string {
	t.Helper()
	var output strings.Builder
	require.NoError(t, component.Render(context.Background(), &output))
	return output.String()
}

func TestActionGroupHTMXActionsRenderDirectGroupedAndOverflowCopies(t *testing.T) {
	htmx := &dropdown.HTMXConfig{
		Post:    "/actions/archive",
		Target:  "#action-result",
		Swap:    "innerHTML",
		Trigger: "click",
		Vals:    `{"source":"action-group"}`,
		Confirm: "Archive this item?",
	}
	html := renderHTML(t, ActionGroup(Config{
		Primary: Action{Label: "Publish", HTMX: htmx, ID: "publish"},
		Secondary: []Action{{
			Label: "Export",
			Items: []Action{{Label: "Archive", HTMX: htmx, ID: "archive"}},
		}},
	}))

	for _, want := range []string{
		`hx-post="/actions/archive"`,
		`hx-target="#action-result"`,
		`hx-swap="innerHTML"`,
		`hx-trigger="click"`,
		`hx-vals="{&#34;source&#34;:&#34;action-group&#34;}"`,
		`hx-confirm="Archive this item?"`,
		`id="archive"`,
		`id="archive-overflow"`,
	} {
		require.Contains(t, html, want)
	}
	require.Equal(t, 3, strings.Count(html, `hx-post="/actions/archive"`))
}

func TestActionGroupHTMXPostTakesPrecedenceOverGet(t *testing.T) {
	html := renderHTML(t, ActionGroup(Config{
		Primary: Action{Label: "Save", HTMX: &dropdown.HTMXConfig{Get: "/preview", Post: "/save"}},
	}))

	require.Contains(t, html, `hx-post="/save"`)
	require.NotContains(t, html, `hx-get="/preview"`)
}

func TestActionGroupIdentity(t *testing.T) {
	require.Equal(t, components.KindActionGroup, ActionGroup(Config{}).Kind())
}

func TestActionGroupRendersSmallPrimaryAndProgressiveEnhancementFallback(t *testing.T) {
	html := renderHTML(t, ActionGroup(Config{
		Primary: Action{Label: "Publish", ID: "publish"},
		Secondary: []Action{
			{Label: "Preview", Href: "/preview", ID: "preview"},
			{Label: "Archive", Disabled: true, Tooltip: "Locked", ID: "archive"},
		},
	}))

	require.Contains(t, html, `data-goshtoso-action-group`)
	require.Contains(t, html, `role="group"`)
	require.Contains(t, html, `aria-label="Actions"`)
	require.Contains(t, html, `id="publish"`)
	require.Contains(t, html, `text-xs`)
	require.Contains(t, html, `href="/preview"`)
	require.Contains(t, html, `id="archive"`)
	require.Contains(t, html, `disabled`)
	require.Contains(t, html, `title="Locked"`)
	require.Equal(t, 2, strings.Count(html, `data-action-group-secondary`))
	require.Contains(t, html, `data-action-group-overflow`)
	require.Contains(t, html, `data-action-group-overflow class="shrink-0" hidden`)
	require.Contains(t, html, `aria-label="More actions"`)
}

func TestActionGroupWideGroupUsesFlatDropdownAndOverflowFlattensChildren(t *testing.T) {
	html := renderHTML(t, ActionGroup(Config{
		Label:         "Chart actions",
		OverflowLabel: "Chart options",
		Primary:       Action{Label: "Refresh"},
		Secondary: []Action{
			{
				Label: "Export",
				ID:    "export-group",
				Items: []Action{
					{Label: "PNG", Href: "/export/png", ID: "export-png"},
					{Label: "CSV", OnClick: "downloadCSV()", ID: "export-csv"},
				},
			},
			{Label: "Duplicate", OnClick: "duplicate()", ID: "duplicate"},
		},
	}))

	require.Contains(t, html, `aria-label="Chart actions"`)
	require.Contains(t, html, `id="export-group"`)
	require.Contains(t, html, `>Export<`)
	require.Contains(t, html, `href="/export/png"`)
	require.Contains(t, html, `x-on:click="downloadCSV()"`)
	require.Contains(t, html, `aria-label="Chart options"`)
	require.Contains(t, html, `data-action-group-overflow-counts="3,1"`)
	require.GreaterOrEqual(t, strings.Count(html, `role="menu"`), 2)
	require.NotContains(t, html, `aria-haspopup="menu"`)
}

func TestOverflowSectionsFlattenDeeperGroupsWithoutNestedMenus(t *testing.T) {
	sections, counts := overflowSections([]Action{
		{
			Label: "Export",
			Items: []Action{
				{Label: "PNG"},
				{Label: "Advanced", Items: []Action{{Label: "Raw"}}},
				{Label: "CSV"},
			},
		},
		{Label: "Share"},
	})

	require.Equal(t, []int{5, 1}, counts)
	require.Len(t, sections, 2)
	require.Equal(t, "Export", sections[0].Items[0].Label)
	require.True(t, sections[0].Items[0].Disabled)
	require.Equal(t, "PNG", sections[0].Items[1].Label)
	require.Equal(t, "Advanced", sections[0].Items[2].Label)
	require.True(t, sections[0].Items[2].Disabled)
	require.Equal(t, "Raw", sections[0].Items[3].Label)
	require.Equal(t, "CSV", sections[0].Items[4].Label)
	require.Equal(t, "Share", sections[1].Items[0].Label)
}
