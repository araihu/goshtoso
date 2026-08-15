//go:build e2e && (full || actiongroup)

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestActionGroupResponsiveTransformAndAccessibility(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{
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
		setActionGroupWidthAndWait(t, page, "#action-group-stacked", 760, "expanded")

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
		_, err = page.WaitForFunction(`() =>
			document.activeElement === document.querySelector("#action-group-export button")
		`, nil)
		require.NoError(t, err)
		activeID, err := page.Evaluate(`() => document.activeElement && document.activeElement.closest("#action-group-export")?.id`, nil)
		require.NoError(t, err)
		require.Equal(t, "action-group-export", activeID)
	})

	t.Run("constrained layout flattens grouped children into shared overflow", func(t *testing.T) {
		groupTrigger := group.Locator("button").First()
		require.NoError(t, groupTrigger.Focus())
		setActionGroupWidthAndWait(t, page, "#action-group-stacked", 220, "collapsed")

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
		(() => {
			const trigger = document.querySelector("#action-group-stacked [data-action-group-overflow] button");
			const menu = document.querySelector("#action-group-stacked [data-action-group-overflow] [role=menu]");
			if (trigger?.getAttribute("aria-expanded") !== "true" || !menu || getComputedStyle(menu).display === "none") return false;
			const visibleItems = Array.from(menu.querySelectorAll("[role=menuitem]")).filter(item =>
				!item.hidden && getComputedStyle(item).display !== "none" &&
				getComputedStyle(item).visibility !== "hidden" && item.getClientRects().length > 0
			);
			const active = document.activeElement;
			return visibleItems.length === 4 && active && menu.contains(active) && active.getAttribute("role") === "menuitem";
		})()
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
		setActionGroupWidthAndWait(t, page, "#action-group-priority", 180, "partial")

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
		waitForActionGroupTriggerFocus(t, page)
	})

	t.Run("viewport theme and color-mode matrix has no page overflow", func(t *testing.T) {
		_, err := page.Evaluate(`() => {
			document.querySelector("#action-group-stacked").style.removeProperty("width");
			document.querySelector("#action-group-priority").style.removeProperty("width");
		}`, nil)
		require.NoError(t, err)
		for _, theme := range []string{"goshtoso", "minimal", "modern"} {
			for _, dark := range []bool{false, true} {
				_, err := page.Evaluate(`([theme, dark]) => {
					document.documentElement.dataset.theme = theme;
					const darkMode = Alpine.store('darkMode');
					if (darkMode.on !== dark) darkMode.toggle();
				}`, []any{theme, dark})
				require.NoError(t, err)
				for _, width := range []int{320, 390, 768, 1440} {
					beforeRevision := actionGroupLayoutRevision(t, page, "#action-group-stacked")
					beforeWidth := actionGroupLayoutWidth(t, page, "#action-group-stacked")
					require.NoError(t, page.SetViewportSize(width, 1000))
					waitForActionGroupViewportLayout(t, page, "#action-group-stacked", beforeRevision, beforeWidth, width)
					overflows, err := page.Evaluate(`() => document.documentElement.scrollWidth > document.documentElement.clientWidth`, nil)
					require.NoError(t, err)
					require.Equalf(t, false, overflows, "theme %s dark=%t width=%d", theme, dark, width)
					primaryVisible, err := primary.IsVisible()
					require.NoError(t, err)
					require.Truef(t, primaryVisible, "primary action theme %s dark=%t width=%d", theme, dark, width)
					if os.Getenv("GOSHTOSO_CAPTURE_ACTIONGROUP") == "1" {
						require.NoError(t, os.MkdirAll(screenshotDir, 0o755))
						variants := []struct {
							name     string
							selector string
						}{
							{name: "fragment", selector: "#action-group-fragment"},
							{name: "responsive", selector: "#action-group-responsive [data-goshtoso-action-group]"},
							{name: "stacked", selector: "#action-group-stacked [data-goshtoso-action-group]"},
							{name: "priority", selector: "#action-group-priority [data-goshtoso-action-group]"},
						}
						for _, variant := range variants {
							_, err = page.Locator(variant.selector).Screenshot(playwright.LocatorScreenshotOptions{
								Path: new(filepath.Join(
									screenshotDir,
									fmt.Sprintf("action-group-%s-%s-dark-%t-%d.png", variant.name, theme, dark, width),
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

func TestActionGroupPartialCollapseKeyboardNavigation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{Width: 1440, Height: 1000},
	})
	failures := watchPageFailures(page)
	_, err := page.Goto(baseURL+"/components/action-group", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() =>
		document.querySelector("#action-group-priority [data-goshtoso-action-group]")?.dataset.actionGroupInitialized === "true"
	`, nil)
	require.NoError(t, err)

	stacked := page.Locator("#action-group-stacked")
	stackedOverflow := stacked.Locator("[data-action-group-overflow]")
	collapsedRevision := setActionGroupWidthAndWait(t, page, "#action-group-stacked", 220, "collapsed")
	require.NoError(t, stackedOverflow.Locator("button").First().Focus())
	_, err = page.WaitForFunction(`() =>
		document.activeElement === document.querySelector("#action-group-stacked [data-action-group-overflow] button")
	`, nil)
	require.NoError(t, err)
	expandedRevision := setActionGroupWidthAndWait(t, page, "#action-group-stacked", 760, "expanded")
	require.Greater(t, expandedRevision, collapsedRevision)
	_, err = page.WaitForFunction(`() =>
		document.activeElement?.closest("[data-action-group-secondary]") ===
			document.querySelector("#action-group-stacked [data-action-group-secondary]")
	`, nil)
	require.NoError(t, err)

	priority := page.Locator("#action-group-priority")
	setActionGroupWidthAndWait(t, page, "#action-group-priority", 180, "partial")

	hidden, err := priority.Locator("[data-action-group-secondary]").EvaluateAll(`els => els.map(el => el.hidden)`)
	require.NoError(t, err)
	require.Equal(t, []any{false, true, true}, hidden.([]any), "fixture must keep Edit inline before Share and Delete collapse")

	overflow := priority.Locator("[data-action-group-overflow]")
	trigger := overflow.Locator("button").First()
	require.NoError(t, trigger.Focus())
	require.NoError(t, page.Keyboard().Press("ArrowDown"))
	_, err = page.WaitForFunction(`() => document.activeElement?.textContent.trim() === "Share"`, nil)
	require.NoError(t, err)

	require.NoError(t, page.Keyboard().Press("ArrowDown"))
	require.Equal(t, "Delete", actionGroupMustText(t, overflow.Locator(`[role="menuitem"]:focus`)))
	require.NoError(t, page.Keyboard().Press("ArrowDown"))
	require.Equal(t, "Share", actionGroupMustText(t, overflow.Locator(`[role="menuitem"]:focus`)))
	require.NoError(t, page.Keyboard().Press("ArrowUp"))
	require.Equal(t, "Delete", actionGroupMustText(t, overflow.Locator(`[role="menuitem"]:focus`)))

	require.NoError(t, page.Keyboard().Press("Escape"))
	waitForActionGroupDropdownExpanded(t, page, "#action-group-priority", false)
	waitForActionGroupTriggerFocus(t, page)
	require.NoError(t, page.Keyboard().Press("Enter"))
	waitForActionGroupDropdownExpanded(t, page, "#action-group-priority", true)
	_, err = page.WaitForFunction(`() => document.activeElement?.textContent.trim() === "Share"`, nil)
	require.NoError(t, err)
	require.NoError(t, page.Keyboard().Press("Escape"))
	waitForActionGroupDropdownExpanded(t, page, "#action-group-priority", false)
	waitForActionGroupTriggerFocus(t, page)

	waitForPageSettled(t, page)
	failures.RequireEmpty(t)
}

func TestActionGroupOverflowInheritsReducedMotionDropdownContract(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{Width: 1440, Height: 1000},
	})
	err := page.EmulateMedia(playwright.PageEmulateMediaOptions{
		ReducedMotion: playwright.ReducedMotionReduce,
	})
	require.NoError(t, err)
	_, err = page.Goto(baseURL+"/components/action-group", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() =>
		document.querySelector("#action-group-stacked [data-goshtoso-action-group]")?.dataset.actionGroupInitialized === "true"
	`, nil)
	require.NoError(t, err)

	withoutStorage := page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Use without storage"})
	if visible, visibleErr := withoutStorage.IsVisible(); visibleErr == nil && visible {
		require.NoError(t, withoutStorage.Click())
	}

	setActionGroupWidthAndWait(t, page, "#action-group-stacked", 220, "collapsed")
	trigger := page.Locator("#action-group-stacked [data-action-group-overflow] button").First()
	menu := page.Locator("#action-group-stacked [data-action-group-overflow] [role='menu']")
	require.NoError(t, trigger.Click())
	_, err = page.WaitForFunction(`() => {
		const menu = document.querySelector("#action-group-stacked [data-action-group-overflow] [role=menu]");
		if (!menu || getComputedStyle(menu).display === "none") return false;
		const style = getComputedStyle(menu);
		return style.transitionProperty === "none" &&
			style.opacity === "1" &&
			style.transform === "none";
	}`, nil)
	require.NoError(t, err)

	visible, err := menu.IsVisible()
	require.NoError(t, err)
	require.True(t, visible)
}

func waitForActionGroupTriggerFocus(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.WaitForFunction(`() =>
		document.activeElement === document.querySelector("#action-group-priority [data-action-group-overflow] button")
	`, nil)
	if err != nil {
		diagnostic, diagnosticErr := page.Evaluate(`() => {
			const owner = document.querySelector("#action-group-priority");
			const root = owner?.querySelector("[data-goshtoso-action-group]");
			const trigger = owner?.querySelector("[data-action-group-overflow] button");
			return {
				active: document.activeElement?.textContent?.trim() || document.activeElement?.tagName || "",
				expanded: trigger?.getAttribute("aria-expanded"),
				layoutState: root?.dataset.actionGroupLayoutState,
				layoutRevision: root?.dataset.actionGroupLayoutRevision,
			};
		}`, nil)
		require.NoError(t, diagnosticErr)
		t.Fatalf("overflow trigger focus failed: %v; diagnostic=%#v", err, diagnostic)
	}
}

func waitForActionGroupDropdownExpanded(t *testing.T, page playwright.Page, selector string, expanded bool) {
	t.Helper()
	_, err := page.WaitForFunction(`([selector, expanded]) => {
		const trigger = document.querySelector(selector + " [data-action-group-overflow] button");
		return trigger && trigger.getAttribute("aria-expanded") === String(expanded);
	}`, []any{selector, expanded})
	require.NoError(t, err)
}

func actionGroupLayoutRevision(t *testing.T, page playwright.Page, selector string) int {
	t.Helper()
	value, err := page.Evaluate(`selector => {
		const root = document.querySelector(selector + " [data-goshtoso-action-group]");
		return Number(root?.dataset.actionGroupLayoutRevision ?? -1);
	}`, selector)
	require.NoError(t, err)
	revision, ok := value.(int)
	if ok {
		return revision
	}
	return int(value.(float64))
}

func actionGroupLayoutWidth(t *testing.T, page playwright.Page, selector string) float64 {
	t.Helper()
	value, err := page.Evaluate(`selector => {
		const root = document.querySelector(selector + " [data-goshtoso-action-group]");
		return root?.getBoundingClientRect().width ?? -1;
	}`, selector)
	require.NoError(t, err)
	if width, ok := value.(int); ok {
		return float64(width)
	}
	return value.(float64)
}

func waitForActionGroupViewportLayout(
	t *testing.T,
	page playwright.Page,
	selector string,
	afterRevision int,
	beforeWidth float64,
	expectedViewportWidth int,
) int {
	t.Helper()
	_, err := page.WaitForFunction(`([selector, afterRevision, beforeWidth, expectedViewportWidth]) => {
		const root = document.querySelector(selector + " [data-goshtoso-action-group]");
		const overflow = root?.querySelector(":scope > [data-action-group-overflow]");
		const secondary = root ? Array.from(root.querySelectorAll(":scope > [data-action-group-secondary]")) : [];
		if (!root || !overflow || secondary.length === 0) return false;

		const revision = Number(root.dataset.actionGroupLayoutRevision);
		const state = root.dataset.actionGroupLayoutState;
		const collapsedCount = secondary.filter(wrapper => wrapper.hidden).length;
		const actualState = overflow.hidden
			? "expanded"
			: collapsedCount === secondary.length
				? "collapsed"
				: collapsedCount > 0
					? "partial"
					: "invalid";
		const widthChanged = Math.abs(root.getBoundingClientRect().width - beforeWidth) > 0.5;
		const revisionReady = widthChanged ? revision > afterRevision : revision >= afterRevision;

		return document.documentElement.clientWidth === expectedViewportWidth &&
			Number.isSafeInteger(revision) && revisionReady &&
			state !== "pending" && state === actualState;
	}`, []any{selector, afterRevision, beforeWidth, expectedViewportWidth})
	require.NoError(t, err)
	return actionGroupLayoutRevision(t, page, selector)
}

func setActionGroupWidthAndWait(t *testing.T, page playwright.Page, selector string, width int, state string) int {
	t.Helper()
	previousRevision := actionGroupLayoutRevision(t, page, selector)
	_, err := page.Evaluate(`([selector, width]) => {
		document.querySelector(selector).style.width = width + "px";
	}`, []any{selector, width})
	require.NoError(t, err)
	return waitForActionGroupLayout(t, page, selector, state, previousRevision)
}

func waitForActionGroupLayout(t *testing.T, page playwright.Page, selector, expectedState string, afterRevision int) int {
	t.Helper()
	_, err := page.WaitForFunction(`([selector, expectedState, afterRevision]) => {
		const root = document.querySelector(selector + " [data-goshtoso-action-group]");
		const overflow = root?.querySelector(":scope > [data-action-group-overflow]");
		const secondary = root ? Array.from(root.querySelectorAll(":scope > [data-action-group-secondary]")) : [];
		if (!root || !overflow || secondary.length === 0) return false;

		const revision = Number(root.dataset.actionGroupLayoutRevision);
		const state = root.dataset.actionGroupLayoutState;
		const collapsedCount = secondary.filter(wrapper => wrapper.hidden).length;
		const actualState = overflow.hidden
			? "expanded"
			: collapsedCount === secondary.length
				? "collapsed"
				: collapsedCount > 0
					? "partial"
					: "invalid";

		return Number.isSafeInteger(revision) && revision > afterRevision &&
			state !== "pending" && state === actualState &&
			(expectedState === "" || state === expectedState);
	}`, []any{selector, expectedState, afterRevision})
	require.NoError(t, err)
	return actionGroupLayoutRevision(t, page, selector)
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
	return strings.TrimSpace(value)
}
