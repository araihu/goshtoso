//go:build e2e && full

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/alert"
	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/link"
	"github.com/araihu/goshtoso/components/navbar"
	"github.com/araihu/goshtoso/components/search"
	"github.com/araihu/goshtoso/components/sidebar"
	"github.com/araihu/goshtoso/components/table"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func securityFixtureServer(t *testing.T, body templ.Component) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.Handle("/assets/", assets.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, `<!doctype html><html><head><meta charset="utf-8">`)
		require.NoError(t, head.Dependencies().Render(r.Context(), w))
		_, _ = fmt.Fprint(w, `</head><body>`)
		require.NoError(t, body.Render(r.Context(), w))
		_, _ = fmt.Fprint(w, `</body></html>`)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func renderSecurityComponent(t *testing.T, c templ.Component) string {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, c.Render(context.Background(), &buf))
	return buf.String()
}

func renderSecurityComponentWithChildren(t *testing.T, c templ.Component, children templ.Component) string {
	t.Helper()

	var buf bytes.Buffer
	ctx := templ.WithChildren(context.Background(), children)
	require.NoError(t, c.Render(ctx, &buf))
	return buf.String()
}

func TestSecurityAttackSurfaceUnsafeHrefSchemesDoNotExecute(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	page.SetDefaultTimeout(5000)

	jsHref := `javascript:void(window.__goshtosoHrefPwned='href')`
	body := templ.Raw(strings.Join([]string{
		renderSecurityComponentWithChildren(t, link.Link(
			jsHref,
			link.WithAttrs(templ.Attributes{"data-attack-surface": "link"}),
		), templ.Raw("Link")),
		renderSecurityComponent(t, navbar.Navbar(navbar.Config{
			Brand:     templ.Raw("Brand"),
			BrandHref: jsHref,
			Links: []navbar.NavLink{{
				Label:     "Navbar link",
				Href:      jsHref,
				LinkAttrs: templ.Attributes{"data-attack-surface": "navbar-link"},
			}},
			User: &navbar.UserProfile{Name: "User"},
			UserMenu: []navbar.UserMenuItem{{
				Label:     "Navbar menu",
				Href:      jsHref,
				LinkAttrs: templ.Attributes{"data-attack-surface": "navbar-menu"},
			}},
		})),
		renderSecurityComponent(t, sidebar.Sidebar(sidebar.Config{
			LogoText: "Logo",
			LogoHref: jsHref,
			Items: []sidebar.Item{{
				Label:     "Sidebar item",
				Href:      jsHref,
				LinkAttrs: templ.Attributes{"data-attack-surface": "sidebar-item"},
			}},
			Sections: []sidebar.Section{{
				Title: "Section",
				Items: []sidebar.Item{{
					Label:     "Sidebar section",
					Href:      jsHref,
					LinkAttrs: templ.Attributes{"data-attack-surface": "sidebar-section"},
				}},
			}},
		})),
		`<div id="alert-fixture">` + renderSecurityComponent(t, alert.Alert(alert.Config{
			Title:       "Alert",
			Description: "Unsafe href fixture",
			Link: &alert.LinkConfig{
				Label: "Alert link",
				Href:  jsHref,
			},
		})) + `</div>`,
	}, "\n"))

	srv := securityFixtureServer(t, body)
	_, err := page.Goto(srv.URL, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	require.NoError(t, err)

	locators := []string{
		`a[data-attack-surface="link"]`,
		`a[data-attack-surface="navbar-link"]`,
		`a[data-attack-surface="sidebar-item"]`,
		`a[data-attack-surface="sidebar-section"]`,
		`#alert-fixture a`,
	}
	for _, selector := range locators {
		_, err = page.Evaluate(`selector => {
			const anchor = document.querySelector(selector);
			if (!anchor) throw new Error('missing anchor: ' + selector);
			const href = anchor.getAttribute('href') || '';
			if (href.trim().toLowerCase().startsWith('javascript:')) {
				window.location.href = href;
			}
		}`, selector)
		require.NoError(t, err)
	}

	got, err := page.Evaluate(`() => window.__goshtosoHrefPwned || ''`, nil)
	require.NoError(t, err)
	require.Empty(t, got, "unsafe href schemes must not execute script")
}

func TestSecurityAttackSurfaceSearchAttrsNavigationIsRevalidated(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	page.SetDefaultTimeout(5000)

	jsHref := `javascript:void(window.__goshtosoSearchPwned='attrs')`
	srv := securityFixtureServer(t, search.Search(search.Config{
		ID: "security-search",
		Items: []search.Item{
			{
				Title: "Unsafe attrs result",
				Attrs: templ.Attributes{"data-search-href": jsHref},
			},
			{
				Title: "HTMX GET result",
				Attrs: templ.Attributes{
					"data-search-href": "/must-not-navigate-get",
					"hx-get":           "/htmx-get",
					"hx-swap":          "none",
				},
			},
			{
				Title: "HTMX POST result",
				Attrs: templ.Attributes{
					"data-search-href": "/must-not-navigate-post",
					"hx-post":          "/htmx-post",
					"hx-swap":          "none",
				},
			},
			{
				Title: "Allowed relative result",
				Href:  "/safe-target",
			},
		},
	}))

	_, err := page.Goto(srv.URL, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined'`, nil)
	require.NoError(t, err)
	_, err = page.Evaluate(`() => {
		const modal = document.querySelector('#security-search-dialog');
		const state = modal && modal._x_dataStack && modal._x_dataStack[0];
		if (!state) throw new Error('search Alpine state is unavailable');
		const safeHref = state.safeHref.bind(state);
		window.__goshtosoSearchSafeHrefCalls = [];
		state.safeHref = function (value) {
			window.__goshtosoSearchSafeHrefCalls.push(value);
			return safeHref(value);
		};
	}`, nil)
	require.NoError(t, err)

	trigger := page.Locator(`#security-search button[aria-haspopup="dialog"]`)
	require.NoError(t, trigger.Click())
	input := page.Locator("#security-search-input")
	require.NoError(t, input.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	require.NoError(t, input.Fill("unsafe attrs"))
	unsafeResult := page.Locator(`#security-search-dialog [data-search-result]:visible`)
	require.NoError(t, unsafeResult.Click())

	got, err := page.Evaluate(`() => window.__goshtosoSearchPwned || ''`, nil)
	require.NoError(t, err)
	require.Empty(t, got, "Attrs-provided data-search-href must be revalidated before navigation")
	revalidated, err := page.Evaluate(
		`expected => window.__goshtosoSearchSafeHrefCalls.length === 1 && window.__goshtosoSearchSafeHrefCalls[0] === expected`,
		jsHref,
	)
	require.NoError(t, err)
	require.Equal(t, true, revalidated, "selection must pass the live DOM data-search-href through safeHref")
	require.Equal(t, srv.URL+"/", page.URL(), "unsafe Attrs navigation must leave the current page unchanged")

	for _, tc := range []struct {
		query  string
		path   string
		method string
	}{
		{query: "htmx get", path: "/htmx-get", method: "GET"},
		{query: "htmx post", path: "/htmx-post", method: "POST"},
	} {
		require.NoError(t, trigger.Click())
		require.NoError(t, input.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		}))
		require.NoError(t, input.Fill(tc.query))
		htmxResult := page.Locator(`#security-search-dialog [data-search-result]:visible`)
		request, requestErr := page.ExpectRequest("**"+tc.path, func() error {
			return htmxResult.Click()
		})
		require.NoError(t, requestErr)
		require.Equal(t, tc.method, request.Method())
		require.Equal(t, srv.URL+"/", page.URL(), "HTMX result must suppress plain data-search-href navigation")
	}

	require.NoError(t, trigger.Click())
	require.NoError(t, input.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	require.NoError(t, input.Fill("allowed relative"))
	allowedResult := page.Locator(`#security-search-dialog [data-search-result]:visible`)
	require.NoError(t, allowedResult.Click())
	require.NoError(t, page.WaitForURL("**/safe-target"))
}

func TestSecurityAttackSurfaceTableFilterTargetDoesNotExecuteScript(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	page.SetDefaultTimeout(5000)

	attackTarget := `#rows', pwned: (window.__goshtosoTablePwned = 'target'), unused: '`
	srv := securityFixtureServer(t, table.Table(table.Config{
		ID:   "security-filtered-table",
		HTMX: &table.HTMXConfig{Endpoint: "/rows"},
		Columns: []table.Column{
			{Key: "name", Label: "Name"},
		},
		Rows: []table.Row{{
			ID:    "1",
			Cells: map[string]table.Cell{"name": {Text: "Ada"}},
		}},
		Filters: &table.FilterConfig{
			InitiallyExpanded: true,
			HTMX:              &table.FilterHTMXConfig{Target: attackTarget},
			Filters: []table.Filter{{
				Key:         "search",
				Label:       "Search",
				Type:        table.FilterSearch,
				Placeholder: "Search",
			}},
		},
	}))

	_, err := page.Goto(srv.URL, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined'`, nil)
	require.NoError(t, err)

	input := page.Locator(`#security-filtered-table-filters input[type="search"]`)
	require.NoError(t, input.Fill("a"))
	_, err = input.Evaluate(`el => el.dispatchEvent(new Event('input', { bubbles: true }))`, nil)
	require.NoError(t, err)
	_, _ = page.WaitForFunction(`() => window.__goshtosoTablePwned === 'target'`, nil, playwright.PageWaitForFunctionOptions{
		Timeout: playwright.Float(1000),
	})

	got, err := page.Evaluate(`() => window.__goshtosoTablePwned || ''`, nil)
	require.NoError(t, err)
	require.Empty(t, got, "table filter HTMX target must remain inert data")
}

func TestSecurityAttackSurfaceTableFilterKeyDoesNotExecuteScript(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	page.SetDefaultTimeout(5000)

	attackKey := `search; window.__goshtosoTablePwned = 'filter-key'; search`
	srv := securityFixtureServer(t, table.Table(table.Config{
		ID:   "security-filter-key-table",
		HTMX: &table.HTMXConfig{Endpoint: "/rows"},
		Columns: []table.Column{
			{Key: "name", Label: "Name"},
		},
		Rows: []table.Row{{
			ID:    "1",
			Cells: map[string]table.Cell{"name": {Text: "Ada"}},
		}},
		Filters: &table.FilterConfig{
			InitiallyExpanded: true,
			Filters: []table.Filter{{
				Key:         attackKey,
				Label:       "Search",
				Type:        table.FilterSearch,
				Placeholder: "Search",
			}},
		},
	}))

	_, err := page.Goto(srv.URL, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined'`, nil)
	require.NoError(t, err)

	got, err := page.Evaluate(`() => window.__goshtosoTablePwned || ''`, nil)
	require.NoError(t, err)
	require.Empty(t, got, "table filter keys must not become executable Alpine expressions")
}
