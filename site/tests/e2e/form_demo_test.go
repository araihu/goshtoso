//go:build e2e

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestFormDemoExternalSubmitUsesModalConfirmation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL+"/components/form", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	_, err = page.Evaluate(`() => {
		window.externalFormSubmitCount = 0;
		const form = document.querySelector('#external-form');
		form.addEventListener('submit', () => { window.externalFormSubmitCount += 1; });
	}`)
	require.NoError(t, err)
	initialURL := page.URL()

	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Upgrade (External Button)",
	}).Click())

	count, err := page.Evaluate(`() => String(window.externalFormSubmitCount)`)
	require.NoError(t, err)
	require.Equal(t, "0", count, "opening the modal must not submit the form")

	dialog := page.Locator("[role='dialog'][aria-labelledby='externalSubmitConfirmTitle']")
	require.NoError(t, dialog.WaitFor())
	require.NoError(t, dialog.GetByRole("button", playwright.LocatorGetByRoleOptions{
		Name: "Confirm upgrade",
	}).Click())

	_, err = page.WaitForFunction(`() => window.externalFormSubmitCount === 1`, nil)
	require.NoError(t, err)

	toast := page.Locator("#toast-container [data-toast-id]", playwright.PageLocatorOptions{
		HasText: "Upgrade request submitted",
	})
	require.NoError(t, toast.WaitFor())
	text, err := toast.InnerText()
	require.NoError(t, err)
	require.Contains(t, text, "v1.31.5")
	require.Equal(t, initialURL, page.URL(), "HTMX submit should not reload or navigate the page")
}

func TestFormDemoCollapsibleComboboxEscapesAccordionClip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	_, err := page.Goto(baseURL+"/components/form", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("#advanced").ScrollIntoViewIfNeeded())
	require.NoError(t, page.Locator("#advanced > button").Click())
	require.NoError(t, page.Locator("#advanced-content").WaitFor())
	require.NoError(t, page.Locator(`#log-level [role="combobox"]`).Click())
	require.NoError(t, page.Locator(`#log-level [data-combobox-option][data-value="debug"]`).WaitFor())

	hitDebugOption, err := page.Evaluate(`() => {
		const option = document.querySelector('#log-level [data-combobox-option][data-value="debug"]');
		if (!option) return false;
		const rect = option.getBoundingClientRect();
		const target = document.elementFromPoint(rect.left + 12, rect.top + rect.height / 2);
		return Boolean(target && target.closest('#log-level [data-combobox-option][data-value="debug"]'));
	}`)
	require.NoError(t, err)
	require.Equal(t, true, hitDebugOption, "expanded accordion content must not clip the open combobox dropdown")
}

func TestFormDemoStickyFooterKeepsMobileActionsAndLastControlUsable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{Width: 390, Height: 844},
	})
	_, err := page.Goto(baseURL+"/components/form", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	scroller := page.Locator("#responsive-footer-preview")
	require.NoError(t, scroller.ScrollIntoViewIfNeeded())
	footer := page.Locator("#responsive-footer-form > div:last-child")
	require.NoError(t, footer.WaitFor())

	result, err := page.Evaluate(`() => {
		const scroller = document.querySelector('#responsive-footer-preview');
		const footer = document.querySelector('#responsive-footer-form > div:last-child');
		const cancel = footer?.querySelector('a');
		const submit = footer?.querySelector('button[type="submit"]');
		if (!scroller || !footer || !cancel || !submit) return null;

		const footerStyle = getComputedStyle(footer);
		const cancelRect = cancel.getBoundingClientRect();
		const submitRect = submit.getBoundingClientRect();
		const footerRect = footer.getBoundingClientRect();
		const scrollerRect = scroller.getBoundingClientRect();
		const background = footerStyle.backgroundColor.match(/[\d.]+/g)?.map(Number) ?? [];
		const alpha = background.length === 4 ? background[3] : 1;

		scroller.scrollTop = scroller.scrollHeight;
		const lastControlRect = document.querySelector('#responsive-footer-last-control-input')?.getBoundingClientRect();
		const settledFooterRect = footer.getBoundingClientRect();

		return {
			flexDirection: footerStyle.flexDirection,
			position: footerStyle.position,
			zIndex: Number(footerStyle.zIndex),
			paddingBottom: Number.parseFloat(footerStyle.paddingBottom),
			opaque: alpha === 1,
			stacked: submitRect.top >= cancelRect.bottom,
			fullWidth: Math.abs(cancelRect.width - footerRect.width) < 1 && Math.abs(submitRect.width - footerRect.width) < 1,
			scrollerOverflow: scroller.scrollWidth - scroller.clientWidth,
			actionRightOverflow: Math.max(cancelRect.right, submitRect.right) - scrollerRect.right,
			actionOverflow: Math.max(cancel.scrollWidth - cancel.clientWidth, submit.scrollWidth - submit.clientWidth),
			lastControlClearance: settledFooterRect.top - (lastControlRect?.bottom ?? Number.POSITIVE_INFINITY),
		};
	}`)
	require.NoError(t, err)
	require.NotNil(t, result, "responsive footer demo must render")

	metrics := result.(map[string]any)
	number := func(value any) float64 {
		switch value := value.(type) {
		case int:
			return float64(value)
		case float64:
			return value
		default:
			t.Fatalf("expected numeric footer metric, got %T (%v)", value, value)
			return 0
		}
	}
	require.Equal(t, "column", metrics["flexDirection"])
	require.Equal(t, "static", metrics["position"],
		"a stacked mobile action bar must remain in flow instead of covering the last controls")
	require.GreaterOrEqual(t, number(metrics["zIndex"]), float64(20))
	require.GreaterOrEqual(t, number(metrics["paddingBottom"]), float64(8))
	require.Equal(t, true, metrics["opaque"])
	require.Equal(t, true, metrics["stacked"])
	require.Equal(t, true, metrics["fullWidth"])
	require.LessOrEqual(t, number(metrics["scrollerOverflow"]), float64(0))
	require.LessOrEqual(t, number(metrics["actionRightOverflow"]), float64(0))
	require.LessOrEqual(t, number(metrics["actionOverflow"]), float64(0))
	require.GreaterOrEqual(t, number(metrics["lastControlClearance"]), float64(0),
		"the sticky footer's normal-flow footprint must let the last control scroll fully above it")

	themeResult, err := page.Evaluate(`() => {
		const root = document.documentElement;
		const footer = document.querySelector('#responsive-footer-form > div:last-child');
		const scroller = document.querySelector('#responsive-footer-preview');
		return ['goshtoso', 'minimal'].flatMap(theme => [false, true].map(dark => {
			root.dataset.theme = theme;
			root.classList.toggle('dark', dark);
			const style = getComputedStyle(footer);
			const background = style.backgroundColor.match(/[\d.]+/g)?.map(Number) ?? [];
			const alpha = background.length === 4 ? background[3] : 1;
			const right = Math.max(...[...footer.querySelectorAll('a, button')].map(action => action.getBoundingClientRect().right));
			return {
				label: theme + (dark ? '-dark' : '-light'),
				opaque: alpha === 1,
				rightOverflow: right - scroller.getBoundingClientRect().right,
			};
		}));
	}`)
	require.NoError(t, err)
	for _, raw := range themeResult.([]any) {
		variant := raw.(map[string]any)
		require.Equal(t, true, variant["opaque"], "%s footer must remain opaque", variant["label"])
		require.LessOrEqual(t, number(variant["rightOverflow"]), float64(0),
			"%s footer actions must remain inside their scroll container", variant["label"])
	}

	require.NoError(t, page.SetViewportSize(1440, 900))
	desktopPosition, err := footer.Evaluate("element => getComputedStyle(element).position", nil)
	require.NoError(t, err)
	require.Equal(t, "sticky", desktopPosition, "Sticky should apply once actions fit on one row")
	desktopDirection, err := footer.Evaluate("element => getComputedStyle(element).flexDirection", nil)
	require.NoError(t, err)
	require.Equal(t, "row", desktopDirection, "footer actions should return to an inline row at sm and above")
}
