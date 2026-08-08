//go:build e2e && (full || sidebar)

package e2e

import (
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/sidebar"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestSidebarCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)

	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/sidebar", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	require.NoError(t, page.Locator("#sidebar-fragment").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	mainText, err := page.Locator("main").TextContent()
	require.NoError(t, err)
	require.Contains(t, mainText, "Sidebar")

	t.Run("simple and section variants expose active states and badges", func(t *testing.T) {
		simple := page.Locator("#sidebar-simple")
		require.NoError(t, simple.ScrollIntoViewIfNeeded())
		require.NoError(t, simple.Locator("input[type='search'][placeholder='Search...']").WaitFor())
		require.NoError(t, simple.Locator("a.text-primary").Filter(playwright.LocatorFilterOptions{HasText: "Profile"}).WaitFor())
		require.NoError(t, simple.Locator("a").Filter(playwright.LocatorFilterOptions{HasText: "Inbox"}).GetByText("3", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())

		sections := page.Locator("#sidebar-sections")
		require.NoError(t, sections.ScrollIntoViewIfNeeded())
		require.NoError(t, sections.Locator("[data-sidebar-section='Components']").GetByText("Table", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())
		require.NoError(t, sections.Locator("a.pointer-events-none").Filter(playwright.LocatorFilterOptions{HasText: "Overview"}).WaitFor())
	})

	t.Run("nested section children and badges render", func(t *testing.T) {
		subItems := page.Locator("#sidebar-sub-items")
		require.NoError(t, subItems.ScrollIntoViewIfNeeded())

		require.False(t, sidebarLinkVisibleContaining(t, page, "#sidebar-sub-items", "Create User"))
		require.False(t, sidebarLinkVisibleContaining(t, page, "#sidebar-sub-items", "List Products"))

		users := subItems.Locator("a[aria-controls='ep-users-children']")
		products := subItems.Locator("a[aria-controls='ep-products-children']")
		require.NoError(t, users.Click())
		waitForSidebarLinkVisibleContaining(t, page, "#sidebar-sub-items", "Create User")

		require.NoError(t, products.Click())
		waitForSidebarLinkVisibleContaining(t, page, "#sidebar-sub-items", "Create User")
		waitForSidebarLinkVisibleContaining(t, page, "#sidebar-sub-items", "List Products")

		for _, badge := range []string{"POST", "GET", "PUT", "DEL"} {
			require.NoError(t, subItems.Locator("sup").Filter(playwright.LocatorFilterOptions{HasText: badge}).First().WaitFor())
		}
	})

	t.Run("collapsible demo updates live Alpine visibility", func(t *testing.T) {
		collapsible := page.Locator("#sidebar-collapsible")
		require.NoError(t, collapsible.ScrollIntoViewIfNeeded())

		require.True(t, sidebarVisible(t, page, "#sidebar-collapsible", "Dashboard"))
		require.False(t, sidebarVisible(t, page, "#sidebar-collapsible", "Articles"))

		contentButton := collapsible.Locator("button").Filter(playwright.LocatorFilterOptions{HasText: "Content"})
		require.NoError(t, contentButton.Click())
		_, err := page.WaitForFunction(
			`() => Array.from(document.querySelectorAll("#sidebar-collapsible a")).some((el) => el.textContent.trim() === "Articles" && el.offsetParent !== null)`,
			nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err)
	})

	t.Run("overlay opens and closes with Alpine state", func(t *testing.T) {
		overlay := page.Locator("#sidebar-overlay")
		require.NoError(t, overlay.ScrollIntoViewIfNeeded())

		panel := overlay.GetByText("Overlay", playwright.LocatorGetByTextOptions{Exact: new(true)})
		visible, err := panel.IsVisible()
		require.NoError(t, err)
		require.False(t, visible)

		require.NoError(t, overlay.Locator("button[aria-label='Open sidebar']").Click())
		require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(3000),
		}))
		require.NoError(t, overlay.Locator("input[type='search'][placeholder='Search...']").WaitFor())

		require.NoError(t, overlay.Locator("div[aria-hidden='true']").Click())
		require.NoError(t, panel.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateHidden,
			Timeout: playwright.Float(3000),
		}))
	})

	t.Run("mobile target preserves themes focus and desktop layout", func(t *testing.T) {
		verifySidebarOverlayMobileTargetAcrossAcceptanceThemes(t)
	})

	require.Empty(t, jsErrors, "no JS console/page errors on sidebar demo: %v", jsErrors)
}

func verifySidebarOverlayMobileTargetAcrossAcceptanceThemes(t *testing.T) {
	page := newPage(t, sharedBrowser)
	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(message playwright.ConsoleMessage) {
		if message.Type() == "error" {
			jsErrors = append(jsErrors, message.Text())
		}
	})

	_, err := page.Goto(baseURL, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateLoad})
	require.NoError(t, err)
	html := renderInteractiveDocument(
		t,
		head.Dependencies(head.WithLocalRuntime()),
		templ.Raw(`<main><div data-testid="desktop-sidebar" class="hidden w-72 lg:block">`+
			renderComponentFragment(t, sidebar.Sidebar(sidebar.Config{LogoText: "Desktop navigation"}))+
			`</div><div id="default-overlay">`+
			renderComponentFragment(t, sidebar.Overlay(sidebar.OverlayConfig{
				ID:           "acceptance-default",
				RootClass:    "lg:hidden",
				TriggerLabel: "Open navigation",
				Sidebar: sidebar.Config{
					LogoText: "Default overlay",
					Items:    []sidebar.Item{{Label: "Default panel link", Href: "#default-panel"}},
				},
			}))+
			`</div><div id="custom-overlay">`+
			renderComponentFragment(t, sidebar.Overlay(sidebar.OverlayConfig{
				ID:           "acceptance-custom",
				RootClass:    "lg:hidden",
				Trigger:      templ.Raw(`<svg data-testid="custom-sidebar-icon" viewBox="0 0 28 28" class="size-7" aria-hidden="true"><path d="M4 8h20M4 14h20M4 20h20" stroke="currentColor" stroke-width="2"/></svg>`),
				TriggerLabel: "Open the complete application navigation for this workspace",
				Sidebar: sidebar.Config{
					LogoText: "Custom overlay",
					Items:    []sidebar.Item{{Label: "Custom panel link", Href: "#custom-panel"}},
				},
			}))+
			`</div></main>`),
	)
	require.NoError(t, page.SetContent(html, playwright.PageSetContentOptions{WaitUntil: playwright.WaitUntilStateLoad}))
	require.NoError(t, waitForAlpine(page))
	require.NoError(t, page.SetViewportSize(390, 844))

	defaultTrigger := page.Locator("#default-overlay button[aria-label='Open navigation']")
	customTrigger := page.Locator("#custom-overlay button[aria-controls='acceptance-custom-panel']")
	require.NoError(t, defaultTrigger.WaitFor())
	require.NoError(t, customTrigger.WaitFor())

	for _, theme := range []string{"goshtoso", "minimal", "araihu"} {
		for _, dark := range []bool{false, true} {
			t.Run(theme+sidebarModeName(dark), func(t *testing.T) {
				setSidebarThemeMode(t, page, theme, dark)
				defaultBox, err := defaultTrigger.BoundingBox()
				require.NoError(t, err)
				require.NotNil(t, defaultBox)
				require.GreaterOrEqual(t, defaultBox.Width, 44.0)
				require.GreaterOrEqual(t, defaultBox.Height, 44.0)

				customBox, err := customTrigger.BoundingBox()
				require.NoError(t, err)
				require.NotNil(t, customBox)
				require.GreaterOrEqual(t, customBox.Width, 44.0)
				require.GreaterOrEqual(t, customBox.Height, 44.0)
				requireSidebarTriggerContainsContent(t, customTrigger)
				t.Logf(
					"theme=%s dark=%t viewport=390x844 default=%.0fx%.0f custom=%.0fx%.0f",
					theme,
					dark,
					defaultBox.Width,
					defaultBox.Height,
					customBox.Width,
					customBox.Height,
				)
			})
		}
	}

	require.Equal(t,
		"Open the complete application navigation for this workspace",
		mustAttribute(t, customTrigger, "aria-label"),
	)
	require.Equal(t, 1, mustLocatorCount(t, customTrigger.Locator("[data-testid='custom-sidebar-icon']")))

	_, err = page.Evaluate(`() => document.activeElement?.blur()`, nil)
	require.NoError(t, err)
	require.NoError(t, page.Keyboard().Press("Tab"))
	requireSidebarTriggerKeyboardFocus(t, defaultTrigger)
	require.NoError(t, page.Keyboard().Press("Enter"))
	waitForSidebarExpanded(t, page, "#default-overlay", true)
	waitForSidebarPanelVisible(t, page, "#acceptance-default-panel")
	tabIntoSidebarPanel(t, page, "#acceptance-default-panel")
	require.NoError(t, page.Keyboard().Press("Escape"))
	waitForSidebarExpanded(t, page, "#default-overlay", false)
	require.NoError(t, page.Locator("#acceptance-default-panel").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateHidden,
	}))
	focused, err := defaultTrigger.Evaluate("el => el === document.activeElement", nil)
	require.NoError(t, err)
	require.Equal(t, true, focused, "Escape should preserve focus on the trigger")
	require.NoError(t, page.Keyboard().Press("Tab"))
	requireSidebarTriggerKeyboardFocus(t, customTrigger)
	require.NoError(t, page.Keyboard().Press("Enter"))
	waitForSidebarExpanded(t, page, "#custom-overlay", true)
	waitForSidebarPanelVisible(t, page, "#acceptance-custom-panel")
	tabIntoSidebarPanel(t, page, "#acceptance-custom-panel")
	require.NoError(t, page.Keyboard().Press("Escape"))
	waitForSidebarExpanded(t, page, "#custom-overlay", false)
	require.NoError(t, page.Locator("#acceptance-custom-panel").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateHidden,
	}))
	focused, err = customTrigger.Evaluate("el => el === document.activeElement", nil)
	require.NoError(t, err)
	require.Equal(t, true, focused, "Escape should return focus to the custom trigger")

	require.NoError(t, page.SetViewportSize(1280, 900))
	hidden, err := defaultTrigger.IsHidden()
	require.NoError(t, err)
	require.True(t, hidden, "mobile overlay trigger should stay out of desktop layout")
	desktop := page.Locator("[data-testid='desktop-sidebar']")
	desktopBox, err := desktop.BoundingBox()
	require.NoError(t, err)
	require.NotNil(t, desktopBox)
	require.InDelta(t, 288.0, desktopBox.Width, 0.5, "desktop w-72 sidebar width should remain unchanged")

	require.Empty(t, jsErrors, "no JS console/page errors in sidebar target fixture: %v", jsErrors)
}

func sidebarModeName(dark bool) string {
	if dark {
		return "Dark"
	}
	return "Light"
}

func setSidebarThemeMode(t *testing.T, page playwright.Page, theme string, dark bool) {
	t.Helper()
	_, err := page.Evaluate(`([theme, dark]) => {
		const root = document.documentElement;
		root.setAttribute('data-theme', theme);
		root.classList.toggle('dark', dark);
	}`, []any{theme, dark})
	require.NoError(t, err)
}

func requireSidebarTriggerContainsContent(t *testing.T, trigger playwright.Locator) {
	t.Helper()
	contained, err := trigger.Evaluate(`el =>
		el.scrollWidth <= el.clientWidth && el.scrollHeight <= el.clientHeight &&
		Array.from(el.children).every(child => {
			const parent = el.getBoundingClientRect();
			const box = child.getBoundingClientRect();
			return box.left >= parent.left && box.right <= parent.right &&
				box.top >= parent.top && box.bottom <= parent.bottom;
		})`, nil)
	require.NoError(t, err)
	require.Equal(t, true, contained, "custom icon should stay inside trigger bounds")
}

func tabIntoSidebarPanel(t *testing.T, page playwright.Page, panel string) {
	t.Helper()
	for attempt := 0; attempt < 8; attempt++ {
		focused, err := page.Evaluate(`panel => {
			const node = document.querySelector(panel);
			return !!node && node.contains(document.activeElement);
		}`, panel)
		require.NoError(t, err)
		if focused == true {
			return
		}
		require.NoError(t, page.Keyboard().Press("Tab"))
	}
	t.Fatalf("Tab did not move focus into open sidebar panel %s", panel)
}

func waitForSidebarPanelVisible(t *testing.T, page playwright.Page, panel string) {
	t.Helper()
	require.NoError(t, page.Locator(panel).WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}))
}

func requireSidebarTriggerKeyboardFocus(t *testing.T, trigger playwright.Locator) {
	t.Helper()
	focus, err := trigger.Evaluate(`el => {
		const style = getComputedStyle(el);
		return {
			active: el === document.activeElement,
			outlineStyle: style.outlineStyle,
			outlineWidth: parseFloat(style.outlineWidth),
		};
	}`, nil)
	require.NoError(t, err)
	values, ok := focus.(map[string]any)
	require.True(t, ok, "unexpected focus result %T: %v", focus, focus)
	require.Equal(t, true, values["active"])
	require.NotEqual(t, "none", values["outlineStyle"])
	outlineWidth := 0.0
	switch value := values["outlineWidth"].(type) {
	case float64:
		outlineWidth = value
	case int:
		outlineWidth = float64(value)
	default:
		t.Fatalf("unexpected outline width %T: %v", value, value)
	}
	require.GreaterOrEqual(t, outlineWidth, 2.0)
}

func waitForSidebarExpanded(t *testing.T, page playwright.Page, scope string, expanded bool) {
	t.Helper()
	want := "false"
	if expanded {
		want = "true"
	}
	_, err := page.WaitForFunction(
		`([scope, want]) => document.querySelector(scope + ' button')?.getAttribute('aria-expanded') === want`,
		[]any{scope, want},
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	require.NoError(t, err)
}

func sidebarVisible(t *testing.T, page playwright.Page, scope string, text string) bool {
	t.Helper()

	result, err := page.Evaluate(`([scope, text]) => Array.from(document.querySelectorAll(scope + " a")).some((el) => el.textContent.trim() === text && el.offsetParent !== null)`, []string{scope, text})
	require.NoError(t, err)

	return result.(bool)
}

func sidebarLinkVisibleContaining(t *testing.T, page playwright.Page, scope string, text string) bool {
	t.Helper()

	result, err := page.Evaluate(`([scope, text]) => Array.from(document.querySelectorAll(scope + " a")).some((el) => el.textContent.includes(text) && el.offsetParent !== null)`, []string{scope, text})
	require.NoError(t, err)

	return result.(bool)
}

func waitForSidebarLinkVisibleContaining(t *testing.T, page playwright.Page, scope string, text string) {
	t.Helper()

	_, err := page.WaitForFunction(
		`([scope, text]) => Array.from(document.querySelectorAll(scope + " a")).some((el) => el.textContent.includes(text) && el.offsetParent !== null)`,
		[]string{scope, text},
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	require.NoError(t, err)
}
