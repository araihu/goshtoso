package appshell

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components"
	"github.com/stretchr/testify/require"
)

func renderHTML(t *testing.T, component templ.Component) string {
	t.Helper()

	var output bytes.Buffer
	require.NoError(t, component.Render(context.Background(), &output))
	return output.String()
}

func TestAppShellRendersAccessibleSingleScrollLayoutByDefault(t *testing.T) {
	html := renderHTML(t, AppShell(Config{}))

	require.Contains(t, html, `href="#main-content"`)
	require.Contains(t, html, `>Skip to main content</a>`)
	require.Contains(t, html, "-translate-y-20")
	require.Contains(t, html, "focus:translate-y-0")
	require.NotContains(t, html, "transition-transform")
	require.NotContains(t, html, "duration-200")
	require.Contains(t, html, `<main id="main-content"`)
	require.Contains(t, html, `tabindex="-1"`)
	require.Contains(t, html, "min-h-screen")
	require.Contains(t, html, "relative")
	require.Contains(t, html, "overflow-hidden")
	require.Contains(t, html, "overflow-y-auto")
	require.Equal(t, components.KindAppShell, AppShell(Config{}).Kind())
}

func TestAppShellMainFocusTargetCanBeOverriddenWithoutMutatingAttrs(t *testing.T) {
	attrs := templ.Attributes{
		"data-region": "main",
		"tabindex":    "0",
	}

	html := renderHTML(t, AppShell(Config{MainAttrs: attrs}))

	require.Contains(t, html, `data-region="main"`)
	require.Contains(t, html, `tabindex="0"`)
	require.NotContains(t, html, `tabindex="-1"`)
	require.Equal(t, 1, strings.Count(html, "tabindex="))
	require.Equal(t, templ.Attributes{
		"data-region": "main",
		"tabindex":    "0",
	}, attrs)
}

func TestAppShellMainFocusTargetDefaultDoesNotMutateAttrs(t *testing.T) {
	attrs := templ.Attributes{"data-region": "main"}

	html := renderHTML(t, AppShell(Config{MainAttrs: attrs}))

	require.Contains(t, html, `tabindex="-1"`)
	require.NotContains(t, attrs, "tabindex")
}

func TestAppShellRendersSlotsAndTargetHooks(t *testing.T) {
	html := renderHTML(t, AppShell(Config{
		Header:        templ.Raw(`<nav data-slot="header">Header</nav>`),
		Sidebar:       templ.Raw(`<nav data-slot="sidebar">Sidebar</nav>`),
		Content:       templ.Raw(`<p data-slot="content">Content</p>`),
		MainID:        "workspace",
		SkipLinkLabel: "Skip to workspace",
		RootClass:     "shell-hook",
		RootAttrs:     templ.Attributes{"data-shell": "operations"},
		HeaderClass:   "header-hook",
		HeaderAttrs:   templ.Attributes{"data-region": "header"},
		SidebarClass:  "sidebar-hook",
		SidebarAttrs:  templ.Attributes{"data-region": "sidebar"},
		MainClass:     "main-hook",
		MainAttrs:     templ.Attributes{"data-region": "main"},
	}))

	require.Contains(t, html, `href="#workspace"`)
	require.Contains(t, html, "Skip to workspace")
	require.Contains(t, html, "shell-hook")
	require.Contains(t, html, `data-shell="operations"`)
	require.Contains(t, html, "header-hook")
	require.Contains(t, html, `data-region="header"`)
	require.Contains(t, html, "sidebar-hook")
	require.Contains(t, html, `data-region="sidebar"`)
	require.Contains(t, html, "main-hook")
	require.Contains(t, html, `data-region="main"`)

	headerIndex := strings.Index(html, `data-slot="header"`)
	sidebarIndex := strings.Index(html, `data-slot="sidebar"`)
	contentIndex := strings.Index(html, `data-slot="content"`)
	require.Greater(t, headerIndex, -1)
	require.Greater(t, sidebarIndex, headerIndex)
	require.Greater(t, contentIndex, sidebarIndex)
}

func TestAppShellUsesChildrenAsContentFallback(t *testing.T) {
	ctx := templ.WithChildren(context.Background(), templ.Raw(`<p data-slot="children">Children</p>`))
	var output bytes.Buffer

	require.NoError(t, AppShell(Config{}).Render(ctx, &output))
	require.Contains(t, output.String(), `<p data-slot="children">Children</p>`)
}

func TestAppShellContentConfigTakesPrecedenceOverChildren(t *testing.T) {
	ctx := templ.WithChildren(context.Background(), templ.Raw(`<p data-slot="children">Children</p>`))
	var output bytes.Buffer

	require.NoError(t, AppShell(Config{
		Content: templ.Raw(`<p data-slot="content">Content</p>`),
	}).Render(ctx, &output))
	require.Contains(t, output.String(), `<p data-slot="content">Content</p>`)
	require.NotContains(t, output.String(), `data-slot="children"`)
}
