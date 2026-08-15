//go:build e2e && (full || scrollregion)

package e2e

import (
	"fmt"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestScrollRegionDirectRouteStatesAndInputModes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	t.Run("public viewport is a named region", func(t *testing.T) {
		page, viewport, failures := newScrollRegionTestPage(t)
		actual, err := viewport.Evaluate(`el => ({
			role: el.getAttribute('role') || '',
			ariaLabel: el.getAttribute('aria-label') || '',
			ariaLabelledBy: el.getAttribute('aria-labelledby') || '',
		})`, nil)
		require.NoError(t, err)
		require.Equal(t, map[string]any{
			"role":           "region",
			"ariaLabel":      "Activity history",
			"ariaLabelledBy": "",
		}, actual, "the public focusable viewport must expose an explicit named region contract")
		requireScrollRegionPageHealthy(t, page, failures)
	})

	t.Run("public examples have distinct region names", func(t *testing.T) {
		page, _, failures := newScrollRegionTestPage(t)
		actual, err := page.Locator("#scroll-region-fragment [data-goshtoso-scroll-viewport][role='region']").EvaluateAll(`els => els.map(el => el.getAttribute('aria-label') || el.getAttribute('aria-labelledby') || '')`, nil)
		require.NoError(t, err)
		names, ok := actual.([]any)
		require.True(t, ok, "region names must be returned as an array")
		require.Len(t, names, 3)
		seen := make(map[string]struct{}, len(names))
		for _, value := range names {
			name, ok := value.(string)
			require.True(t, ok)
			require.NotEmpty(t, name)
			_, exists := seen[name]
			require.Falsef(t, exists, "Scroll Region demo landmark name %q must be unique", name)
			seen[name] = struct{}{}
		}
		requireScrollRegionPageHealthy(t, page, failures)
	})

	t.Run("default and no-overflow", func(t *testing.T) {
		page, viewport, failures := newScrollRegionTestPage(t)
		require.Equal(t, "0", scrollRegionAttribute(t, viewport, "tabindex"))
		_, err := page.WaitForFunction(`() => {
			const viewport = document.querySelector("#scroll-region-no-overflow [data-goshtoso-scroll-viewport]");
			const root = viewport && viewport.closest("[data-goshtoso-scroll-region]");
			const start = root && root.querySelector("[data-goshtoso-scroll-start-indicator]");
			const end = root && root.querySelector("[data-goshtoso-scroll-end-indicator]");
			return viewport && start && end && viewport.scrollHeight <= viewport.clientHeight + 1 &&
				start.hidden && end.hidden;
		}`, nil)
		require.NoError(t, err)
		requireScrollRegionPageHealthy(t, page, failures)
	})

	t.Run("focused keyboard start middle and end", func(t *testing.T) {
		t.Run("PageDown reaches the middle", func(t *testing.T) {
			page, viewport, failures := newScrollRegionTestPage(t)
			setScrollRegionPosition(t, page, "#scroll-region-default", "start")
			require.NoError(t, viewport.Focus())
			focused, err := viewport.Evaluate(`el => ({
				active: document.activeElement === el,
				outlineStyle: getComputedStyle(el).outlineStyle,
				outlineWidth: getComputedStyle(el).outlineWidth,
			})`, nil)
			require.NoError(t, err)
			focusState := focused.(map[string]any)
			require.Equal(t, true, focusState["active"])
			require.NotEqual(t, "none", focusState["outlineStyle"])
			require.NotEqual(t, "0px", focusState["outlineWidth"])

			require.NoError(t, viewport.Press("PageDown"))
			waitForScrollRegionPosition(t, page, "#scroll-region-default", "middle")
			requireScrollRegionPageHealthy(t, page, failures)
		})

		t.Run("End reaches the end", func(t *testing.T) {
			page, viewport, failures := newScrollRegionTestPage(t)
			setScrollRegionPosition(t, page, "#scroll-region-default", "start")
			require.NoError(t, viewport.Focus())
			require.NoError(t, viewport.Press("End"))
			waitForScrollRegionPosition(t, page, "#scroll-region-default", "end")
			requireScrollRegionPageHealthy(t, page, failures)
		})

		t.Run("Home returns to the start", func(t *testing.T) {
			page, viewport, failures := newScrollRegionTestPage(t)
			setScrollRegionPosition(t, page, "#scroll-region-default", "end")
			require.NoError(t, viewport.Focus())
			require.NoError(t, viewport.Press("Home"))
			waitForScrollRegionPosition(t, page, "#scroll-region-default", "start")
			requireScrollRegionPageHealthy(t, page, failures)
		})
	})

	t.Run("mouse wheel updates boundary state", func(t *testing.T) {
		page, viewport, failures := newScrollRegionTestPage(t)
		setScrollRegionPosition(t, page, "#scroll-region-default", "start")
		box, err := viewport.BoundingBox()
		require.NoError(t, err)
		require.NotNil(t, box)
		require.NoError(t, page.Mouse().Move(box.X+box.Width/2, box.Y+box.Height/2))
		require.NoError(t, page.Mouse().Wheel(0, 180))
		waitForScrollRegionPosition(t, page, "#scroll-region-default", "middle")
		requireScrollRegionPageHealthy(t, page, failures)
	})

	t.Run("touch scroll updates boundary state", func(t *testing.T) {
		page, viewport, failures := newScrollRegionTestPage(t)
		setScrollRegionPosition(t, page, "#scroll-region-default", "start")
		require.NoError(t, viewport.Focus())
		box, err := viewport.BoundingBox()
		require.NoError(t, err)
		require.NotNil(t, box)
		session, err := page.Context().NewCDPSession(page)
		require.NoError(t, err)
		t.Cleanup(func() { _ = session.Detach() })
		_, err = session.Send("Input.synthesizeScrollGesture", map[string]any{
			"x":                 int(box.X + box.Width/2),
			"y":                 int(box.Y + box.Height/2),
			"yDistance":         -160,
			"speed":             800,
			"gestureSourceType": "touch",
		})
		require.NoError(t, err)
		waitForScrollRegionPosition(t, page, "#scroll-region-default", "middle")
		requireScrollRegionPageHealthy(t, page, failures)
	})

	t.Run("narrow layout and catalog theme mode matrix", func(t *testing.T) {
		page, viewport, failures := newScrollRegionTestPage(t)
		setScrollRegionPosition(t, page, "#scroll-region-default", "start")
		require.NoError(t, viewport.Focus())
		require.NoError(t, viewport.Press("End"))
		waitForScrollRegionPosition(t, page, "#scroll-region-default", "end")

		require.NoError(t, page.SetViewportSize(320, 900))
		for _, theme := range []string{"araihu", "goshtoso", "minimal", "modern"} {
			for _, dark := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/dark=%t", theme, dark), func(t *testing.T) {
					_, err := page.Evaluate(`([theme, dark]) => {
						document.documentElement.dataset.theme = theme;
						document.documentElement.classList.toggle("dark", dark);
						const viewport = document.querySelector("#scroll-region-default [data-goshtoso-scroll-viewport]");
						viewport.scrollTop = 0;
					}`, []any{theme, dark})
					require.NoError(t, err)
					waitForScrollRegionPosition(t, page, "#scroll-region-default", "start")
					overflow, err := page.Evaluate(`() => document.documentElement.scrollWidth > document.documentElement.clientWidth`, nil)
					require.NoError(t, err)
					require.Equalf(t, false, overflow, "theme=%s dark=%t must not overflow at 320px", theme, dark)
				})
			}
		}
		requireScrollRegionPageHealthy(t, page, failures)
	})
}

func newScrollRegionTestPage(t *testing.T) (playwright.Page, playwright.Locator, *pageFailures) {
	t.Helper()
	page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{
		HasTouch: new(true),
		Viewport: &playwright.Size{Width: 720, Height: 900},
	})
	failures := watchPageFailures(page)
	response, err := page.Goto(baseURL+"/components/scroll-region", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.Equal(t, 200, response.Status(), "Scroll Region must have a direct documentation route")

	root := page.Locator("#scroll-region-default [data-goshtoso-scroll-region]")
	viewport := root.Locator("[data-goshtoso-scroll-viewport]")
	require.NoError(t, root.WaitFor())
	waitForScrollRegionPosition(t, page, "#scroll-region-default", "start")
	return page, viewport, failures
}

func setScrollRegionPosition(t *testing.T, page playwright.Page, containerID, position string) {
	t.Helper()
	_, err := page.Evaluate(`([containerID, position]) => {
		const viewport = document.querySelector(containerID + " [data-goshtoso-scroll-viewport]");
		const max = viewport.scrollHeight - viewport.clientHeight;
		viewport.scrollTop = position === "end" ? max : 0;
	}`, []any{containerID, position})
	require.NoError(t, err)
	waitForScrollRegionPosition(t, page, containerID, position)
}

func requireScrollRegionPageHealthy(t *testing.T, page playwright.Page, failures *pageFailures) {
	t.Helper()
	waitForPageSettled(t, page)
	failures.RequireEmpty(t)
}

func waitForScrollRegionPosition(t *testing.T, page playwright.Page, containerID, position string) {
	t.Helper()
	_, err := page.WaitForFunction(`([containerID, position]) => {
		const viewport = document.querySelector(containerID + " [data-goshtoso-scroll-viewport]");
		const root = viewport && viewport.closest("[data-goshtoso-scroll-region]");
		const start = root && root.querySelector("[data-goshtoso-scroll-start-indicator]");
		const end = root && root.querySelector("[data-goshtoso-scroll-end-indicator]");
		if (!viewport || !start || !end) return false;
		const max = viewport.scrollHeight - viewport.clientHeight;
		if (position === "start") return viewport.scrollTop <= 1 && start.hidden && !end.hidden;
		if (position === "middle") return viewport.scrollTop > 1 && viewport.scrollTop < max - 1 && !start.hidden && !end.hidden;
		if (position === "end") return max > 1 && viewport.scrollTop >= max - 1 && !start.hidden && end.hidden;
		return false;
	}`, []any{containerID, position})
	if err == nil {
		return
	}
	snapshot, snapshotErr := page.Evaluate(`containerID => {
		const viewport = document.querySelector(containerID + " [data-goshtoso-scroll-viewport]");
		const root = viewport && viewport.closest("[data-goshtoso-scroll-region]");
		const start = root && root.querySelector("[data-goshtoso-scroll-start-indicator]");
		const end = root && root.querySelector("[data-goshtoso-scroll-end-indicator]");
		return {
			viewport: !!viewport,
			scrollTop: viewport && viewport.scrollTop,
			clientHeight: viewport && viewport.clientHeight,
			scrollHeight: viewport && viewport.scrollHeight,
			startHidden: start && start.hidden,
			endHidden: end && end.hidden,
		};
	}`, containerID)
	require.NoError(t, snapshotErr)
	t.Fatalf("wait for Scroll Region %s state: %v; snapshot=%#v", position, err, snapshot)
}

func scrollRegionAttribute(t *testing.T, locator playwright.Locator, name string) string {
	t.Helper()
	value, err := locator.GetAttribute(name)
	require.NoError(t, err)
	return value
}
