package tabs

import (
	"bytes"
	"context"
	"html"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// render is a small local helper that renders a Config and returns the raw HTML.
func render(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Tabs(cfg).Render(context.Background(), &buf))
	return buf.String()
}

// TestCoverageDefaultContainerID verifies that an empty ID falls back to the
// "tabs" container prefix used to build panel IDs.
func TestCoverageDefaultContainerID(t *testing.T) {
	out := render(t, Config{
		Tabs: []Tab{
			{ID: "one", Label: "One", Content: templ.Raw("body-one")},
		},
	})
	browser := html.UnescapeString(out)

	assert.Contains(t, browser, `aria-controls="tabpaneltabsone"`)
	assert.Contains(t, browser, `id="tabpaneltabsone"`)
	assert.Contains(t, browser, `role="tablist"`)
	assert.Contains(t, browser, `aria-label="tab options"`)
	assert.Contains(t, browser, "body-one")
}

// TestCoverageRootClassApplied confirms RootClass is appended to the container.
func TestCoverageRootClassApplied(t *testing.T) {
	out := render(t, Config{
		RootClass: "my-extra-class",
		Tabs:      []Tab{{ID: "a", Label: "A"}},
	})
	assert.Contains(t, out, "w-full my-extra-class")
}

// TestCoverageIconAndBadgeBranches exercises the tabButton icon+badge layout
// branch plus BadgeActiveClasses / BadgeInactiveClasses.
func TestCoverageIconAndBadgeBranches(t *testing.T) {
	out := render(t, Config{
		ID: "acct",
		Tabs: []Tab{
			{
				ID:      "billing",
				Label:   "Billing",
				Icon:    templ.Raw(`<svg data-testid="icon"></svg>`),
				Badge:   "9",
				Content: templ.Raw("billing-body"),
			},
		},
	})
	browser := html.UnescapeString(out)

	// Icon rendered.
	assert.Contains(t, browser, `data-testid="icon"`)
	// Badge text rendered.
	assert.Contains(t, browser, ">9<")
	// Icon+badge layout class branch (flex variant).
	assert.Contains(t, browser, "flex h-min items-center gap-2 px-4 py-2 text-sm")
	// Badge active/inactive classes wired into x-bind:class.
	assert.Contains(t, browser, BadgeActiveClasses())
	assert.Contains(t, browser, BadgeInactiveClasses())
}

// TestCoverageLabelSlotOverridesVisibleLabel keeps Label available for ARIA
// while allowing callers to render richer visual labels such as status badges.
func TestCoverageLabelSlotOverridesVisibleLabel(t *testing.T) {
	out := render(t, Config{
		ID: "status",
		Tabs: []Tab{
			{
				ID:        "ok",
				Label:     "200",
				LabelSlot: templ.Raw(`<span data-testid="status-badge">200 OK</span>`),
				Content:   templ.Raw("ok-body"),
			},
		},
	})
	browser := html.UnescapeString(out)

	assert.Contains(t, browser, `data-testid="status-badge"`)
	assert.Contains(t, browser, `>200 OK<`)
	assert.Contains(t, browser, `aria-label="200"`)
	assert.Contains(t, browser, "flex h-min items-center gap-2 px-4 py-2 text-sm")
	assert.NotContains(t, browser, `<button>200</button>`)
}

// TestCoveragePlainButtonLayout verifies the non-icon/non-badge base class.
func TestCoveragePlainButtonLayout(t *testing.T) {
	out := render(t, Config{
		Tabs: []Tab{{ID: "plain", Label: "Plain"}},
	})
	assert.Contains(t, out, "h-min px-4 py-2 text-sm")
	assert.NotContains(t, out, "flex h-min items-center gap-2")
}

// TestCoverageActiveInactiveClassesWired confirms the active/inactive class
// helpers feed the tab button's x-bind:class expression.
func TestCoverageActiveInactiveClassesWired(t *testing.T) {
	out := render(t, Config{
		Tabs: []Tab{{ID: "a", Label: "A"}},
	})
	assert.Contains(t, out, ActiveClasses())
	assert.Contains(t, out, InactiveClasses())
}

// TestCoverageStaticPanelNilContent confirms a static panel renders even when
// Content is nil (the nil branch in staticTabPanel).
func TestCoverageStaticPanelNilContent(t *testing.T) {
	out := render(t, Config{
		ID:   "x",
		Tabs: []Tab{{ID: "empty", Label: "Empty"}},
	})
	browser := html.UnescapeString(out)
	assert.Contains(t, browser, `id="tabpanelxempty"`)
	assert.Contains(t, browser, `role="tabpanel"`)
	assert.Contains(t, browser, `aria-label="Empty"`)
}

// TestCoverageHTMXPanelDefaultSwap covers htmxTabPanel with default swap and no
// indicator.
func TestCoverageHTMXPanelDefaultSwap(t *testing.T) {
	out := render(t, Config{
		ID: "lazy",
		Tabs: []Tab{
			{
				ID:    "details",
				Label: "Details",
				HTMX:  &TabHTMX{Get: "/api/tabs/details"},
			},
		},
	})
	browser := html.UnescapeString(out)

	assert.Contains(t, browser, `hx-get="/api/tabs/details"`)
	assert.Contains(t, browser, `hx-trigger="intersect once"`)
	assert.Contains(t, browser, `hx-swap="innerHTML"`)
	assert.Contains(t, browser, "Loading...")
	// x-effect lazy-load expression targets the panel id and uses default swap.
	assert.Contains(t, browser, "htmx.ajax('GET', '/api/tabs/details'")
	assert.Contains(t, browser, "target: '#tabpanellazydetails'")
	assert.Contains(t, browser, "swap: 'innerHTML'")
	// No indicator attribute when unset.
	assert.NotContains(t, browser, "hx-indicator")
}

// TestCoverageHTMXPanelCustomSwapAndIndicator covers the custom swap value and
// the indicator attribute branch.
func TestCoverageHTMXPanelCustomSwapAndIndicator(t *testing.T) {
	out := render(t, Config{
		ID: "lazy2",
		Tabs: []Tab{
			{
				ID:    "report",
				Label: "Report",
				HTMX: &TabHTMX{
					Get:       "/api/tabs/report",
					Swap:      "outerHTML",
					Indicator: "#spin",
				},
			},
		},
	})
	browser := html.UnescapeString(out)

	assert.Contains(t, browser, `hx-swap="outerHTML"`)
	assert.Contains(t, browser, `hx-indicator="#spin"`)
	assert.Contains(t, browser, "swap: 'outerHTML'")
}

// TestCoverageMixedStaticAndHTMXPanels renders both panel kinds in one config so
// the tabPanel dispatch branch is exercised in both directions.
func TestCoverageMixedStaticAndHTMXPanels(t *testing.T) {
	out := render(t, Config{
		ID: "mix",
		Tabs: []Tab{
			{ID: "static", Label: "Static", Content: templ.Raw("static-body")},
			{ID: "lazy", Label: "Lazy", HTMX: &TabHTMX{Get: "/api/mix"}},
		},
	})
	browser := html.UnescapeString(out)
	assert.Contains(t, browser, "static-body")
	assert.Contains(t, browser, `hx-get="/api/mix"`)
	assert.Contains(t, browser, `id="tabpanelmixstatic"`)
	assert.Contains(t, browser, `id="tabpanelmixlazy"`)
}

// TestCoverageSyncHashInitExpression covers the SyncHash branch of tabsData,
// including the valid-ID array and the hash watcher.
func TestCoverageSyncHashInitExpression(t *testing.T) {
	out := render(t, Config{
		ID:       "hash",
		SyncHash: true,
		Tabs: []Tab{
			{ID: "first", Label: "First"},
			{ID: "second", Label: "Second"},
		},
	})
	browser := html.UnescapeString(out)

	assert.Contains(t, browser, "init()")
	assert.Contains(t, browser, "var v=['first','second'];")
	assert.Contains(t, browser, "history.replaceState(null,'','#'+t)")
	assert.Contains(t, browser, "selectedTab:'first'")
}

// TestCoverageDefaultTabSelection verifies DefaultTab seeds selectedTab when set
// explicitly rather than defaulting to the first tab.
func TestCoverageDefaultTabSelection(t *testing.T) {
	cfg := Config{
		DefaultTab: "second",
		Tabs: []Tab{
			{ID: "first", Label: "First"},
			{ID: "second", Label: "Second"},
		},
	}
	assert.Contains(t, tabsData(cfg), "selectedTab:'second'")
}

// TestCoverageNoSyncHashOmitsInit confirms the non-SyncHash branch produces the
// simpler x-data without an init() hook.
func TestCoverageNoSyncHashOmitsInit(t *testing.T) {
	data := tabsData(Config{Tabs: []Tab{{ID: "a", Label: "A"}}})
	assert.Contains(t, data, "selectedTab:'a'")
	assert.Contains(t, data, "moveFocus:function")
	assert.NotContains(t, data, "init()")
	assert.NotContains(t, data, "history.replaceState")
}

// TestCoverageEmptyConfigRenders ensures a zero-value Config renders the shell
// with no panic and an empty tab list.
func TestCoverageEmptyConfigRenders(t *testing.T) {
	out := render(t, Config{})
	assert.Contains(t, out, `role="tablist"`)
	assert.True(t, strings.Contains(out, "w-full"))
}
