//go:build e2e && (full || modules)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestAppShellLivePreviews(t *testing.T) {
	for _, spec := range []struct{ slug, family string }{
		{"component-docs-shell", "componentdocshell"}, {"console-shell", "consoleshell"}, {"landing-shell", "landingshell"},
	} {
		t.Run(spec.family, func(t *testing.T) {
			page := newPage(t, sharedBrowser)
			failures := watchPageFailures(page)
			require.NoError(t, page.SetViewportSize(1162, 1240))
			_, err := page.Goto(baseURL + "/modules/app-shells/shells/" + spec.slug)
			require.NoError(t, err)
			frame := page.FrameLocator("iframe")
			require.NoError(t, page.Locator("iframe").ScrollIntoViewIfNeeded())
			require.NoError(t, frame.Locator("h1").WaitFor())
			_, err = page.WaitForFunction(`() => document.querySelector('iframe').contentWindow.innerWidth===1280`, nil)
			require.NoError(t, err)
			_, err = page.Evaluate(`() => { window.previewParentMarker=true; document.querySelector('iframe').contentWindow.previewMarker=true; }`)
			require.NoError(t, err)
			require.NoError(t, frame.Locator(`a[href$="/activity"]`).First().Click())
			_, err = page.WaitForFunction(`() => document.querySelector('iframe').contentWindow.location.pathname.endsWith('/activity')`, nil)
			require.NoError(t, err)
			retained, err := page.Evaluate(`() => window.previewParentMarker`)
			require.NoError(t, err)
			require.Equal(t, true, retained, "preview navigation must not replace the documentation")
			if spec.family != "landingshell" {
				retained, err = page.Evaluate(`() => document.querySelector('iframe').contentWindow.previewMarker`)
				require.NoError(t, err)
				require.Equal(t, true, retained, "docs and console links use fragment navigation")
			}
			require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Mobile", Exact: new(true)}).Click())
			_, err = page.WaitForFunction(`() => document.querySelector('iframe').contentWindow.innerWidth===390`, nil)
			require.NoError(t, err)
			fits, err := page.Locator("iframe").Evaluate(`e => { const frame=e.getBoundingClientRect(), box=e.parentElement.getBoundingClientRect(); return frame.left>=box.left-1 && frame.right<=box.right+1; }`, nil)
			require.NoError(t, err)
			require.Equal(t, true, fits)
			menu := `button[aria-label="Open navigation"]`
			links := `.sidebar-scroll a[href$="/overview"]`
			if spec.family == "landingshell" {
				menu = `.landing-shell__mobile-enhanced button[aria-haspopup="dialog"]`
				links = `.landing-shell__mobile-links a[href$="/overview"]`
			}
			require.NoError(t, page.Locator("iframe").ScrollIntoViewIfNeeded())
			require.NoError(t, frame.Locator(menu).Click())
			if spec.family == "landingshell" {
				// Wait for the drawer focus trap to activate before following a link.
				// Alpine schedules activation after the opening event.
				_, err = page.WaitForFunction(`() => {
					const doc = document.querySelector('iframe').contentDocument;
					return !!doc.activeElement?.closest('[role="dialog"]');
				}`, nil)
				require.NoError(t, err)
			}
			require.NoError(t, frame.Locator(links).First().Click())
			_, err = page.WaitForFunction(`() => document.querySelector('iframe').contentWindow.location.pathname.endsWith('/overview')`, nil)
			require.NoError(t, err)
			failures.RequireEmpty(t)
		})
	}
}
