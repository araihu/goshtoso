//go:build e2e && controls

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestStaticControlTransitionsHonorReducedMotion(t *testing.T) {
	cleanupServer := setupServer(t)
	defer cleanupServer()

	page := newIsolatedPage(t)
	require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{
		ReducedMotion: playwright.ReducedMotionReduce,
	}))

	cases := []struct {
		name     string
		route    string
		selector string
	}{
		{name: "button", route: "/components/button", selector: "#button-fragment button"},
		{name: "modal", route: "/components/modal", selector: "#modal-default button"},
		{name: "select", route: "/components/select", selector: "#os-trigger"},
		{name: "alert", route: "/components/alert", selector: "#alert-action button"},
		{name: "dropdown", route: "/components/dropdown", selector: "#dropdown-click button"},
		{name: "search", route: "/components/search", selector: "#component-search button[aria-haspopup='dialog']"},
		{name: "tagslist", route: "/components/tags-list", selector: "#tagsDemo [data-tagslist-add]"},
		{name: "textarea", route: "/components/textarea", selector: "button[aria-label='send']"},
		{name: "structured-input", route: "/components/structured-input", selector: "#labelsDemo [data-add-row]"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := page.Goto(baseURL+tc.route, playwright.PageGotoOptions{
				WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			})
			require.NoError(t, err)
			control := page.Locator(tc.selector).First()
			require.NoError(t, control.WaitFor(playwright.LocatorWaitForOptions{
				State: playwright.WaitForSelectorStateAttached,
			}))

			transitionProperty, err := control.Evaluate("element => getComputedStyle(element).transitionProperty", nil)
			require.NoError(t, err)
			require.Equal(t, "none", transitionProperty, "reduced-motion %s control should not transition", tc.name)
		})
	}
}
