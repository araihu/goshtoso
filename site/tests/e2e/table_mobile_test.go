//go:build e2e && (full || table)

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/table"
	"github.com/araihu/goshtoso/components/tooltip"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func mobileTableFixture(expandable bool) table.Config {
	return table.Config{
		ID: "mobile-fixture", Caption: "Reported hosts",
		Mobile:  &table.MobileConfig{PrimaryColumn: "name", Expandable: expandable},
		HTMX:    &table.HTMXConfig{Endpoint: "/mobile-rows"},
		Columns: []table.Column{{Key: "pin", Label: "Pinned", HeaderSuffix: tooltip.Help("mobile-pin-help", "How pinning works", "Check to pin a host to your homepage.")}, {Key: "name", Label: "Host", Sortable: true}, {Key: "cpu", Label: "CPU"}, {Key: "arch", Label: "Architecture", MobileSecondary: true}},
		Rows: []table.Row{{ID: "__proto__", Link: "/row-destination", LinkMode: table.LinkFull,
			Expandable: true, Detail: templ.Raw(`<p>Full row details</p>`), Actions: templ.Raw(`<button id="nested-action" type="button">Action</button>`), Cells: map[string]table.Cell{
				"name": {Text: "worker-with-a-long-but-complete-identity.example.internal", Description: "Reported host"}, "cpu": {Text: "Unavailable"},
				"arch": {Text: "arm64"}, "pin": {Component: templ.Raw(`<input id="mobile-pin" type="checkbox" aria-label="Pin worker">`)},
			}}},
	}
}

func TestTableMobileStackedExpansion(t *testing.T) {
	for _, theme := range []struct {
		name string
		dark bool
	}{{"goshtoso", false}, {"goshtoso", true}, {"minimal", false}, {"minimal", true}} {
		t.Run(fmt.Sprintf("%s/dark=%t", theme.name, theme.dark), func(t *testing.T) {
			page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: 390, Height: 844}})
			var pageErrors []string
			page.OnPageError(func(err error) { pageErrors = append(pageErrors, err.Error()) })
			cfg := mobileTableFixture(true)
			var fragment bytes.Buffer
			require.NoError(t, table.TableRows(cfg).Render(context.Background(), &fragment))
			sorted := cfg
			sorted.SortBy, sorted.SortDir = "name", table.SortAsc
			fmt.Fprintf(&fragment, `<hx-partial hx-target="#%s" hx-swap="innerHTML">`, cfg.TheadID())
			require.NoError(t, table.TableHeadContent(sorted).Render(context.Background(), &fragment))
			fragment.WriteString(`</hx-partial>`)
			require.NoError(t, page.Route("**/mobile-rows*", func(route playwright.Route) {
				require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{ContentType: playwright.String("text/html"), Body: fragment.String()}))
			}))
			document := renderInteractiveDocument(t, head.DependenciesMinimal(head.WithLocalRuntime()), table.Table(cfg))
			require.NoError(t, page.Route("**/mobile-fixture", func(route playwright.Route) {
				require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{ContentType: playwright.String("text/html"), Body: document}))
			}))
			_, err := page.Goto(baseURL + "/mobile-fixture")
			require.NoError(t, err)
			_, err = page.WaitForFunction(`() => document.querySelector('[data-table-mobile-expanded="false"]') !== null`, nil)
			require.NoError(t, err)
			setThemeMode(t, page, theme.name, theme.dark)
			identity := page.Locator(`tbody a[href="/row-destination"]`)
			identityText, err := identity.InnerText()
			require.NoError(t, err)
			require.Equal(t, "worker-with-a-long-but-complete-identity.example.internal", identityText)
			role, err := page.Locator("tbody tr").First().GetAttribute("role")
			require.NoError(t, err)
			require.Equal(t, "row", role)
			rowDetails := page.Locator(".goshtoso-table-row-detail-toggle")
			require.NoError(t, rowDetails.Press("Enter"))
			_, err = page.WaitForFunction(`() => document.querySelector('.goshtoso-table-row-detail-toggle').getAttribute('aria-expanded') === 'true'`, nil)
			require.NoError(t, err)
			require.NoError(t, page.Locator("#nested-action").Click())
			expanded, err := rowDetails.GetAttribute("aria-expanded")
			require.NoError(t, err)
			require.Equal(t, "true", expanded, "nested actions must not toggle row details")
			require.NoError(t, rowDetails.Press("Space"))
			secondary := page.Locator("[data-table-mobile-secondary]")
			toggle := page.Locator(".goshtoso-table-mobile-toggle button")
			for _, width := range []int{320, 390} {
				require.NoError(t, page.SetViewportSize(width, 844))
				visible, err := secondary.IsVisible()
				require.NoError(t, err)
				require.False(t, visible)
				overflow, err := page.Evaluate(`() => document.documentElement.scrollWidth > innerWidth || document.querySelector('.goshtoso-table-responsive').scrollWidth > document.querySelector('.goshtoso-table-responsive').clientWidth`)
				require.NoError(t, err)
				require.Equal(t, false, overflow)
			}
			takeScreenshot(t, page, fmt.Sprintf("table-mobile-%s-dark-%t", theme.name, theme.dark))
			require.NoError(t, toggle.Press("Enter"))
			_, err = page.WaitForFunction(`() => document.querySelector('[data-table-mobile-secondary]').getBoundingClientRect().height > 0`, nil)
			require.NoError(t, err)
			require.NoError(t, page.Locator("#mobile-pin").Check())
			require.NotContains(t, page.URL(), "row-destination")
			require.NoError(t, page.Locator("thead [data-table-sort-request] button").Click())
			_, err = page.WaitForFunction(`() => document.querySelector('th[aria-sort="ascending"]') !== null`, nil)
			require.NoError(t, err)
			focused, err := page.Evaluate(`() => document.activeElement.id === 'mobile-fixture-column-name-sort'`)
			require.NoError(t, err)
			require.Equal(t, true, focused, "sorting must retain focus after header replacement")
			helpVisible, err := page.Locator("thead th:not([data-table-sort-request]) button").IsVisible()
			require.NoError(t, err)
			require.False(t, helpVisible, "header replacement must not expose informational headings")
			_, err = page.WaitForFunction(`() => document.querySelector('.goshtoso-table-mobile-toggle button').getAttribute('aria-expanded') === 'true'`, nil)
			require.NoError(t, err)
			visible, err := secondary.IsVisible()
			require.NoError(t, err)
			require.True(t, visible, "row replacement must retain expanded state")
			require.NoError(t, toggle.Press("Space"))
			require.NoError(t, page.SetViewportSize(1440, 1000))
			visible, err = secondary.IsVisible()
			require.NoError(t, err)
			require.True(t, visible, "desktop must show secondary columns even when mobile is collapsed")
			visible, err = toggle.IsVisible()
			require.NoError(t, err)
			require.False(t, visible)
			takeScreenshot(t, page, fmt.Sprintf("table-desktop-%s-dark-%t", theme.name, theme.dark))
			// A narrow dashboard panel must adapt even on a desktop viewport.
			_, err = page.Evaluate(`() => document.querySelector('.goshtoso-table-responsive').style.maxWidth = '390px'`)
			require.NoError(t, err)
			require.NoError(t, toggle.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
			count, err := page.Locator("#mobile-pin").Count()
			require.NoError(t, err)
			require.Equal(t, 1, count)
			require.Empty(t, pageErrors)
		})
	}
}

func TestTableMobileAllFieldsWithoutExpansionOrJavaScript(t *testing.T) {
	for _, javascript := range []bool{true, false} {
		t.Run(fmt.Sprintf("javascript=%t", javascript), func(t *testing.T) {
			page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: 390, Height: 844}, JavaScriptEnabled: playwright.Bool(javascript)})
			cfg := mobileTableFixture(!javascript)
			document := renderInteractiveDocument(t, head.DependenciesMinimal(head.WithLocalRuntime()), table.Table(cfg))
			require.NoError(t, page.Route("**/mobile-fallback", func(route playwright.Route) {
				require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{ContentType: playwright.String("text/html"), Body: document}))
			}))
			_, err := page.Goto(baseURL + "/mobile-fallback")
			require.NoError(t, err)
			visible, err := page.Locator("[data-table-mobile-secondary]").IsVisible()
			require.NoError(t, err)
			require.True(t, visible)
			visible, err = page.Locator(".goshtoso-table-mobile-toggle button").IsVisible()
			require.NoError(t, err)
			require.False(t, visible)
		})
	}
}

func TestTableMobileHeaderKeepsOnlyTableActions(t *testing.T) {
	for _, theme := range []struct {
		name string
		dark bool
	}{{"goshtoso", false}, {"goshtoso", true}, {"minimal", false}, {"minimal", true}} {
		for _, action := range []string{"none", "sort", "select"} {
			t.Run(fmt.Sprintf("%s/dark=%t/%s", theme.name, theme.dark, action), func(t *testing.T) {
				page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: 509, Height: 844}})
				cfg := mobileTableFixture(false)
				cfg.Columns[1].Sortable = action == "sort"
				if action == "select" {
					cfg.ShowCheckbox = true
					cfg.Columns = cfg.Columns[1:]
				}
				cfg.Columns[0].HeaderSuffix = tooltip.Help("mobile-header-help", "How pinning works", "Check to pin a record to your homepage.")
				document := renderInteractiveDocument(t, head.DependenciesMinimal(head.WithLocalRuntime()), table.Table(cfg))
				require.NoError(t, page.Route("**/mobile-header", func(route playwright.Route) {
					require.NoError(t, route.Fulfill(playwright.RouteFulfillOptions{ContentType: playwright.String("text/html"), Body: document}))
				}))
				_, err := page.Goto(baseURL + "/mobile-header")
				require.NoError(t, err)
				setThemeMode(t, page, theme.name, theme.dark)
				help := page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "How pinning works", IncludeHidden: playwright.Bool(true)})
				for _, width := range []int{320, 509} {
					require.NoError(t, page.SetViewportSize(width, 844))
					visible, err := help.IsVisible()
					require.NoError(t, err)
					require.False(t, visible, "static column help must not keep a heading visible")
					height, err := page.Locator("thead").Evaluate(`el => el.getBoundingClientRect().height`, nil)
					require.NoError(t, err)
					if action == "none" {
						require.EqualValues(t, 0, height, "an informational header must occupy no space")
					} else {
						require.NotEqualValues(t, 0, height)
					}
				}
				if action == "sort" {
					visible, err := page.Locator("thead [data-table-sort-request] button").IsVisible()
					require.NoError(t, err)
					require.True(t, visible)
				}
				if action == "none" {
					require.NoError(t, page.Keyboard().Press("Tab"))
					focused, err := page.Evaluate(`() => document.activeElement.id`)
					require.NoError(t, err)
					require.Equal(t, "mobile-pin", focused, "hidden header help must not retain a tab stop")
				}
				if action == "select" {
					selectAll := page.Locator(`thead input[type="checkbox"]`)
					require.NoError(t, selectAll.Check())
					checked, err := page.Locator(`tbody input[type="checkbox"]`).IsChecked()
					require.NoError(t, err)
					require.True(t, checked)
				}
				associations, err := page.Evaluate(`() => [...document.querySelectorAll('td[headers]')].every(cell => document.getElementById(cell.getAttribute('headers'))?.tagName === 'TH')`)
				require.NoError(t, err)
				require.Equal(t, true, associations)
				require.NoError(t, page.SetViewportSize(1440, 1000))
				visible, err := help.IsVisible()
				require.NoError(t, err)
				require.True(t, visible, "desktop must retain column help")
				require.NoError(t, help.Click())
				popover := page.Locator("#mobile-header-help")
				require.NoError(t, popover.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
				require.NoError(t, page.SetViewportSize(509, 844))
				require.NoError(t, popover.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}), "open help must not remain detached when its heading disappears")
			})
		}
	}
}
