package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestActionGroupResponsiveTransformAndAccessibility(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser, playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{Width: 1440, Height: 1000},
	})
	failures := watchPageFailures(page)
	_, err := page.Goto(baseURL+"/components/action-group", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => {
		const root = document.querySelector("#action-group-stacked [data-goshtoso-action-group]");
		return root && root.dataset.actionGroupInitialized === "true";
	}`, nil)
	require.NoError(t, err)
	withoutStorage := page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Use without storage"})
	if visible, visibleErr := withoutStorage.IsVisible(); visibleErr == nil && visible {
		require.NoError(t, withoutStorage.Click())
	}

	stacked := page.Locator("#action-group-stacked")
	root := stacked.Locator("[data-goshtoso-action-group]")
	primary := stacked.Locator("[data-action-group-primary]")
	group := stacked.Locator("[data-action-group-secondary]").Nth(0)
	overflow := stacked.Locator("[data-action-group-overflow]")

	t.Run("wide grouped action uses one flat dropdown", func(t *testing.T) {
		_, err := stacked.Evaluate("el => { el.style.width = '760px' }", nil)
		require.NoError(t, err)
		waitForActionGroupLayout(t, page, "#action-group-stacked", false)

		hidden, err := page.Evaluate(`() => document.querySelector("#action-group-stacked [data-action-group-secondary]").hidden`, nil)
		require.NoError(t, err)
		require.Equal(t, false, hidden)
		overflowHidden, err := page.Evaluate(`() => document.querySelector("#action-group-stacked [data-action-group-overflow]").hidden`, nil)
		require.NoError(t, err)
		require.Equal(t, true, overflowHidden)

		trigger := group.Locator("button").First()
		require.NoError(t, trigger.Focus())
		require.NoError(t, page.Keyboard().Press("ArrowDown"))
		_, err = page.WaitForFunction(`() => document.querySelector("#action-group-export button").getAttribute("aria-expanded") === "true"`, nil)
		require.NoError(t, err)
		require.Equal(t, 0, actionGroupMustCount(t, group.Locator(`[role="menu"] [role="menu"]`)))

		_, err = page.WaitForFunction(`() => document.activeElement && document.activeElement.getAttribute("role") === "menuitem"`, nil)
		require.NoError(t, err)
		require.Equal(t, "PNG", actionGroupMustText(t, group.Locator(`[role="menuitem"]:focus`)))
		require.NoError(t, page.Keyboard().Press("ArrowDown"))
		_, err = page.WaitForFunction(`() => document.activeElement && document.activeElement.textContent.trim() === "CSV"`, nil)
		require.NoError(t, err)

		require.NoError(t, page.Keyboard().Press("Escape"))
		_, err = page.WaitForFunction(`() => document.querySelector("#action-group-export button").getAttribute("aria-expanded") === "false"`, nil)
		require.NoError(t, err)
		activeID, err := page.Evaluate(`() => document.activeElement && document.activeElement.closest("#action-group-export")?.id`, nil)
		require.NoError(t, err)
		require.Equal(t, "action-group-export", activeID)
	})

	t.Run("constrained layout flattens grouped children into shared overflow", func(t *testing.T) {
		groupTrigger := group.Locator("button").First()
		require.NoError(t, groupTrigger.Focus())
		_, err := stacked.Evaluate("el => { el.style.width = '220px' }", nil)
		require.NoError(t, err)
		waitForActionGroupLayout(t, page, "#action-group-stacked", true)

		primaryHidden, err := page.Evaluate(`() => document.querySelector("#action-group-stacked [data-action-group-primary]").hidden`, nil)
		require.NoError(t, err)
		require.Equal(t, false, primaryHidden)
		groupHidden, err := page.Evaluate(`() => document.querySelector("#action-group-stacked [data-action-group-secondary]").hidden`, nil)
		require.NoError(t, err)
		require.Equal(t, true, groupHidden)
		overflowHidden, err := page.Evaluate(`() => document.querySelector("#action-group-stacked [data-action-group-overflow]").hidden`, nil)
		require.NoError(t, err)
		require.Equal(t, false, overflowHidden)

		_, err = page.WaitForFunction(`() => document.activeElement && document.activeElement.getAttribute("aria-label") === "More actions"`, nil)
		require.NoError(t, err)

		trigger := overflow.Locator("button").First()
		require.Equal(t, "More actions", actionGroupMustAttribute(t, trigger, "aria-label"))
		require.Equal(t, "true", actionGroupMustAttribute(t, trigger, "aria-haspopup"))
		require.NoError(t, page.Keyboard().Press("ArrowDown"))
		_, err = page.WaitForFunction(`() => {
			const trigger = document.querySelector("#action-group-stacked [data-action-group-overflow] button");
			return trigger.getAttribute("aria-expanded") === "true";
		}`, nil)
		require.NoError(t, err)

		menu := overflow.Locator(`[role="menu"]`)
		require.Equal(t, 0, actionGroupMustCount(t, menu.Locator(`[role="menu"]`)))
		visibleItems := menu.Locator(`[role="menuitem"]:visible`)
		_, err = page.WaitForFunction(`() =>
			document.querySelectorAll("#action-group-stacked [data-action-group-overflow] [role=menuitem]:not([hidden])").length === 4
		`, nil)
		require.NoError(t, err)
		visibilityState, err := menu.Evaluate(`menu => ({
			menu: { hidden: menu.hidden, display: getComputedStyle(menu).display },
			sections: Array.from(menu.children).map(el => ({ hidden: el.hidden, inline: el.style.display, display: getComputedStyle(el).display })),
			items: Array.from(menu.querySelectorAll('[role="menuitem"]')).map(el => ({ label: el.textContent.trim(), hidden: el.hidden, inline: el.style.display, display: getComputedStyle(el).display })),
		})`, nil)
		require.NoError(t, err)
		require.Equalf(t, 4, actionGroupMustCount(t, visibleItems), "visibility state: %#v", visibilityState)
		require.Equal(t, "Export", actionGroupMustText(t, visibleItems.Nth(0)))
		require.Equal(t, "PNG", actionGroupMustText(t, visibleItems.Nth(1)))
		require.Equal(t, "CSV", actionGroupMustText(t, visibleItems.Nth(2)))
		require.Equal(t, "Duplicate", actionGroupMustText(t, visibleItems.Nth(3)))
		disabled, err := visibleItems.Nth(0).Evaluate("el => el.hasAttribute('disabled')", nil)
		require.NoError(t, err)
		require.Equal(t, true, disabled)

		require.NoError(t, visibleItems.Nth(2).Click())
		_, err = page.WaitForFunction(`() => document.querySelector("#action-group-stacked [x-text='lastAction']").textContent === "csv"`, nil)
		require.NoError(t, err)

		require.NoError(t, page.Locator("h1").Click())
		_, err = page.WaitForFunction(`() => {
			const trigger = document.querySelector("#action-group-stacked [data-action-group-overflow] button");
			return trigger.getAttribute("aria-expanded") === "false";
		}`, nil)
		require.NoError(t, err)
	})

	t.Run("source order remains priority order", func(t *testing.T) {
		priority := page.Locator("#action-group-priority")
		_, err := priority.Evaluate("el => { el.style.width = '180px' }", nil)
		require.NoError(t, err)
		_, err = page.WaitForFunction(`() => {
			const root = document.querySelector("#action-group-priority [data-goshtoso-action-group]");
			const overflow = root.querySelector("[data-action-group-overflow]");
			return !overflow.hidden;
		}`, nil)
		require.NoError(t, err)

		hidden, err := priority.Locator("[data-action-group-secondary]").EvaluateAll(`els => els.map(el => el.hidden)`)
		require.NoError(t, err)
		values := hidden.([]any)
		seenHidden := false
		for _, value := range values {
			current := value.(bool)
			if current {
				seenHidden = true
			}
			if seenHidden {
				require.True(t, current, "lower-priority actions must not remain inline after an earlier action collapsed")
			}
		}
		require.True(t, seenHidden)

		collapsedCount := 0
		for _, value := range values {
			if value.(bool) {
				collapsedCount++
			}
		}
		priorityOverflow := priority.Locator("[data-action-group-overflow]")
		require.NoError(t, priorityOverflow.Locator("button").First().Click())
		require.NoError(t, priorityOverflow.Locator(`[role="menu"]`).WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		}))
		_, err = page.WaitForFunction(`count =>
			document.querySelectorAll("#action-group-priority [data-action-group-overflow] [role=menuitem]:not([hidden])").length === count
		`, collapsedCount)
		require.NoError(t, err)
		require.Equal(t, collapsedCount, actionGroupMustCount(t, priorityOverflow.Locator(`[role="menuitem"]:visible`)))
		require.NoError(t, page.Keyboard().Press("Escape"))
	})

	t.Run("expanding restores focus from overflow to first inline secondary", func(t *testing.T) {
		_, err := stacked.Evaluate("el => { el.style.width = '220px' }", nil)
		require.NoError(t, err)
		waitForActionGroupLayout(t, page, "#action-group-stacked", true)
		require.NoError(t, overflow.Locator("button").First().Focus())

		_, err = stacked.Evaluate("el => { el.style.width = '760px' }", nil)
		require.NoError(t, err)
		waitForActionGroupLayout(t, page, "#action-group-stacked", false)
		restored, err := page.Evaluate(`() =>
			document.activeElement &&
			document.activeElement.closest("[data-action-group-secondary]") ===
				document.querySelector("#action-group-stacked [data-action-group-secondary]")
		`, nil)
		require.NoError(t, err)
		require.Equal(t, true, restored)
	})

	t.Run("viewport theme and color-mode matrix has no page overflow", func(t *testing.T) {
		_, err := page.Evaluate(`() => {
			document.querySelector("#action-group-stacked").style.removeProperty("width");
			document.querySelector("#action-group-priority").style.removeProperty("width");
		}`, nil)
		require.NoError(t, err)
		for _, theme := range []string{"goshtoso", "minimal"} {
			for _, dark := range []bool{false, true} {
				_, err := page.Evaluate(`([theme, dark]) => {
					document.documentElement.dataset.theme = theme;
					document.documentElement.classList.toggle("dark", dark);
				}`, []any{theme, dark})
				require.NoError(t, err)
				for _, width := range []int{320, 390, 768, 1440} {
					require.NoError(t, page.SetViewportSize(width, 1000))
					time.Sleep(100 * time.Millisecond)
					overflows, err := page.Evaluate(`() => document.documentElement.scrollWidth > document.documentElement.clientWidth`, nil)
					require.NoError(t, err)
					require.Equalf(t, false, overflows, "theme %s dark=%t width=%d", theme, dark, width)
					primaryVisible, err := primary.IsVisible()
					require.NoError(t, err)
					require.Truef(t, primaryVisible, "primary action theme %s dark=%t width=%d", theme, dark, width)
					if os.Getenv("GOSHTOSO_CAPTURE_ACTIONGROUP") == "1" {
						require.NoError(t, os.MkdirAll(screenshotDir, 0o755))
						for _, variant := range []string{"responsive", "stacked"} {
							_, err = page.Locator(
								"#action-group-" + variant + " [data-goshtoso-action-group]",
							).Screenshot(playwright.LocatorScreenshotOptions{
								Path: playwright.String(filepath.Join(
									screenshotDir,
									fmt.Sprintf("action-group-%s-%s-dark-%t-%d.png", variant, theme, dark, width),
								)),
							})
							require.NoError(t, err)
						}
					}
				}
			}
		}
	})

	require.Equal(t, "group", actionGroupMustAttribute(t, root, "role"))
	require.Equal(t, "Chart actions", actionGroupMustAttribute(t, root, "aria-label"))
	waitForPageSettled(t, page)
	failures.RequireEmpty(t)
}

func waitForActionGroupLayout(t *testing.T, page playwright.Page, selector string, overflowVisible bool) {
	t.Helper()
	_, err := page.WaitForFunction(`([selector, overflowVisible]) => {
		const root = document.querySelector(selector + " [data-goshtoso-action-group]");
		const overflow = root && root.querySelector("[data-action-group-overflow]");
		if (!root || !overflow) return false;
		const expectedWidth = overflowVisible ? 220 : 760;
		return Math.abs(root.getBoundingClientRect().width - expectedWidth) < 2 &&
			overflow.hidden === !overflowVisible;
	}`, []any{selector, overflowVisible})
	require.NoError(t, err)
}

func actionGroupMustCount(t *testing.T, locator playwright.Locator) int {
	t.Helper()
	count, err := locator.Count()
	require.NoError(t, err)
	return count
}

func actionGroupMustAttribute(t *testing.T, locator playwright.Locator, name string) string {
	t.Helper()
	value, err := locator.GetAttribute(name)
	require.NoError(t, err)
	return value
}

func actionGroupMustText(t *testing.T, locator playwright.Locator) string {
	t.Helper()
	value, err := locator.TextContent()
	require.NoError(t, err)
	return value
}
