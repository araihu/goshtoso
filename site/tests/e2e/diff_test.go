//go:build e2e && (full || diff)

package e2e

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/diff"
	"github.com/araihu/goshtoso/components/head"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestDiffThemesResponsiveAndVerbatimWithoutJavaScript(t *testing.T) {
	for _, width := range []int{390, 1440} {
		for _, theme := range []string{"goshtoso", "minimal"} {
			for _, dark := range []bool{false, true} {
				t.Run(fmt.Sprintf("%d/%s/dark=%t", width, theme, dark), func(t *testing.T) {
					page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: width, Height: 1000}, JavaScriptEnabled: playwright.Bool(false)})
					failures := watchPageFailures(page)
					_, err := page.Goto(baseURL + "/components/diff")
					require.NoError(t, err)
					setThemeMode(t, page, theme, dark)
					assertDiffLayout(t, page, width)
					assertDiffTheme(t, page, dark)
					code := page.Locator("#diff-verbatim code")
					for index, expected := range []string{"\t  <div data-version='old'>old</div>  ", "\t  <img src=x onerror=alert('new')>  ", "  Olá, 世界 👋  ", "  Olá, 世界 👋  ", strings.Repeat("original text · ", 24), strings.Repeat("revised text · ", 12)} {
						actual, err := code.Nth(index).TextContent()
						require.NoError(t, err)
						require.Equal(t, expected, actual)
					}
					count, err := page.Locator("#diff-verbatim script, #diff-verbatim img").Count()
					require.NoError(t, err)
					require.Zero(t, count)
					empty, err := page.Locator("#diff-empty .gs-diff-empty").TextContent()
					require.NoError(t, err)
					require.Equal(t, "No lines to compare.", empty)
					unchanged, err := page.Locator(`#diff-identical [data-diff-marker=" "]`).Count()
					require.NoError(t, err)
					require.Equal(t, 2, unchanged)
					failures.RequireEmpty(t)
				})
			}
		}
	}
}

func assertDiffLayout(t *testing.T, page playwright.Page, width int) {
	t.Helper()
	result, err := page.Evaluate(`width => {
        const viewers = [...document.querySelectorAll('#diff-fragment .gs-diff')];
        const mobile = width < 640;
        return viewers.length === 6 && viewers.every(viewer => {
            if (viewer.getBoundingClientRect().right > document.documentElement.clientWidth + 1) return false;
            const rows = [...viewer.querySelectorAll('.gs-diff-row')];
            return rows.every(row => {
                const cells = [...row.children];
                const a = cells[0].getBoundingClientRect(), b = cells[1].getBoundingClientRect();
                const codeFits = [...row.querySelectorAll('code')].every(code => {
                    const source = code.getBoundingClientRect(), cell = code.closest('.gs-diff-cell').getBoundingClientRect();
                    const number = code.closest('.gs-diff-source').querySelector('.gs-diff-number').getBoundingClientRect();
                    const marker = code.closest('.gs-diff-source').querySelector('.gs-diff-marker').getBoundingClientRect();
                    return source.right <= cell.right + 1 && number.right <= marker.left + 1 && marker.right <= source.left + 1;
                });
                if (!codeFits) return false;
                if (!mobile) return Math.abs(a.top - b.top) <= 1 && Math.abs(a.width - b.width) <= 1 && a.right <= b.left + 1;
                if (cells[0].classList.contains('gs-diff-padding')) return getComputedStyle(cells[0]).display === 'none';
                if (cells[1].classList.contains('gs-diff-padding')) return getComputedStyle(cells[1]).display === 'none';
                return a.bottom <= b.top + 1 && Math.abs(a.left - b.left) <= 1;
            });
        }) && [...document.querySelectorAll('#diff-unequal .gs-diff-label')].every(label => (getComputedStyle(label).display !== 'none') === mobile);
    }`, width)
	require.NoError(t, err)
	require.Equal(t, true, result, "source lines must fit consistent aligned columns or stacked pairs")
	longLineScrolls, err := page.Locator("#diff-verbatim").Evaluate("el => el.scrollWidth > el.clientWidth", nil)
	require.NoError(t, err)
	require.Equal(t, true, longLineScrolls)
	blankLines, err := page.Locator("#diff-unequal code").EvaluateAll(`nodes => nodes.filter(node => node.textContent === '').every(node => node.closest('.gs-diff-cell').getBoundingClientRect().height > 0) && nodes.filter(node => node.textContent === '').length === 2`)
	require.NoError(t, err)
	require.Equal(t, true, blankLines, "real blank lines must stay visible")
}

func assertDiffTheme(t *testing.T, page playwright.Page, dark bool) {
	t.Helper()
	result, err := page.Locator("#diff-text").Evaluate(`(viewer, dark) => {
        const suffix = dark ? '-dark' : '';
        const probe = document.createElement('span'); viewer.append(probe);
        const expected = token => { probe.style.color = 'var(--color-' + token + suffix + ')'; return getComputedStyle(probe).color; };
        const added = viewer.querySelector('[data-diff-marker="+"] .gs-diff-marker');
        const removed = viewer.querySelector('[data-diff-marker="−"] .gs-diff-marker');
        const gutters = [added, removed].every(marker => {
            const cell = marker.closest('.gs-diff-cell');
            const number = cell.querySelector('.gs-diff-number');
            return getComputedStyle(number).backgroundColor !== getComputedStyle(cell).backgroundColor &&
                getComputedStyle(number).color === expected('on-surface');
        });
        const valid = getComputedStyle(viewer).color === expected('on-surface') &&
            gutters &&
            getComputedStyle(added).color === expected('success-action') &&
            getComputedStyle(removed).color === expected('danger-action') &&
            added.textContent === '+' && removed.textContent === '−' &&
            added.getAttribute('aria-label') === 'Added' && removed.getAttribute('aria-label') === 'Removed' &&
            getComputedStyle(added.closest('.gs-diff-cell')).backgroundColor !== getComputedStyle(removed.closest('.gs-diff-cell')).backgroundColor;
        probe.remove(); return valid;
    }`, dark)
	require.NoError(t, err)
	require.Equal(t, true, result, "change markers use readable semantic tokens and accessible names")
}

func TestDiffNativeKeyboardScrolling(t *testing.T) {
	page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{JavaScriptEnabled: playwright.Bool(false)})
	_, err := page.Goto(baseURL + "/components/diff")
	require.NoError(t, err)
	defaultHeight, err := page.Locator("#diff-text").Evaluate("el => getComputedStyle(el).maxHeight", nil)
	require.NoError(t, err)
	require.Equal(t, "512px", defaultHeight)
	scroll := page.Locator("#diff-scroll")
	bounds, err := scroll.Evaluate("el => getComputedStyle(el).maxHeight === '180px' && el.getBoundingClientRect().height <= 181 && el.scrollHeight > el.clientHeight", nil)
	require.NoError(t, err)
	require.Equal(t, true, bounds)
	require.NoError(t, scroll.ScrollIntoViewIfNeeded())
	require.NoError(t, scroll.Focus())
	active, activeErr := scroll.Evaluate("el => document.activeElement === el", nil)
	require.NoError(t, activeErr)
	require.Equal(t, true, active)
	require.NoError(t, page.Keyboard().Press("PageDown"))
	require.Eventually(t, func() bool {
		scrolled, scrollErr := scroll.Evaluate("el => el.scrollTop > 0", nil)
		return scrollErr == nil && scrolled == true
	}, 5*time.Second, 50*time.Millisecond, "focused comparison must scroll vertically with PageDown")
	longLine := page.Locator("#diff-verbatim")
	require.NoError(t, longLine.ScrollIntoViewIfNeeded())
	require.NoError(t, longLine.Focus())
	require.NoError(t, page.Keyboard().Press("ArrowRight"))
	require.Eventually(t, func() bool {
		scrolled, scrollErr := longLine.Evaluate("el => el.scrollLeft > 0", nil)
		return scrollErr == nil && scrolled == true
	}, 5*time.Second, 50*time.Millisecond, "focused comparison must scroll horizontally with ArrowRight")
	focus, err := longLine.Evaluate("el => el.matches(':focus-visible') && getComputedStyle(el).outlineStyle !== 'none'", nil)
	require.NoError(t, err)
	require.Equal(t, true, focus)
}

func TestDiffOrdinaryHTMXReplacementAndSidebarNavigation(t *testing.T) {
	page := newPage(t, sharedBrowser)
	failures := watchPageFailures(page)
	_, err := page.Goto(baseURL + "/components/button")
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	_, err = page.ExpectResponse("**/components/diff", func() error {
		return page.Locator(`a[hx-get="/components/diff"]`).First().Click()
	})
	require.NoError(t, err)
	require.NoError(t, page.Locator("#diff-fragment").WaitFor())
	assertDiffLayout(t, page, 1280)
	failures.RequireEmpty(t)

	// A consumer-owned endpoint returns just the component, not a demo page.
	mux := http.NewServeMux()
	mux.Handle("/assets/", assets.Handler())
	mux.HandleFunc("GET /comparison", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = diff.Diff(diff.Config{ID: "comparison", BeforeLabel: "Earlier text", AfterLabel: "Later text", Rows: []diff.Row{{Operation: diff.OperationReplace, Before: &diff.Line{Text: "old", Number: 1}, After: &diff.Line{Text: "\t<updated>& safe", Number: 2}}}}).Render(r.Context(), w)
	})
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		var content bytes.Buffer
		content.WriteString(`<!doctype html><html><head><meta name="viewport" content="width=device-width, initial-scale=1"><title>Text comparison</title>`)
		_ = head.Dependencies(head.WithLocalRuntime()).Render(r.Context(), &content)
		content.WriteString(`</head><body><main><button hx-get="/comparison" hx-target="#comparison" hx-swap="outerHTML">Replace comparison</button>`)
		_ = diff.Diff(diff.Config{ID: "comparison"}).Render(r.Context(), &content)
		content.WriteString(`</main></body></html>`)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(content.Bytes())
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	consumer := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{Viewport: &playwright.Size{Width: 390, Height: 844}})
	consumerFailures := watchPageFailures(consumer)
	_, err = consumer.Goto(server.URL)
	require.NoError(t, err)
	_, err = consumer.WaitForFunction("() => typeof htmx !== 'undefined'", nil)
	require.NoError(t, err)
	require.NoError(t, consumer.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Replace comparison"}).Click())
	require.NoError(t, consumer.Locator("#comparison code").First().WaitFor())
	newText, err := consumer.Locator("#comparison code").Nth(1).TextContent()
	require.NoError(t, err)
	require.Equal(t, "\t<updated>& safe", newText)
	label, err := consumer.Locator("#comparison").GetAttribute("aria-label")
	require.NoError(t, err)
	require.Equal(t, "Earlier text / Later text", label)
	stacked, err := consumer.Locator("#comparison").Evaluate(`el => { const cells = el.querySelectorAll('.gs-diff-cell'); return cells[0].getBoundingClientRect().bottom <= cells[1].getBoundingClientRect().top; }`, nil)
	require.NoError(t, err)
	require.Equal(t, true, stacked)
	consumerFailures.RequireEmpty(t)
}
