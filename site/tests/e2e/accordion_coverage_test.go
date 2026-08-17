//go:build e2e && (full || accordion)

package e2e

import (
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestAccordionCoverageDemo(t *testing.T) {
	page := newPage(t, sharedBrowser)
	var consoleErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			consoleErrors = append(consoleErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/accordion", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	require.NoError(t, page.Locator("#accordion-fragment").WaitFor())

	t.Run("single open variant closes the previously expanded item", func(t *testing.T) {
		accordion := page.Locator("#accordion-default")
		first := accordion.Locator("#controls-default-1")
		second := accordion.Locator("#controls-default-2")

		require.NoError(t, first.Click())
		_, err := page.WaitForFunction(
			`() => document.querySelector('#controls-default-1')?.getAttribute('aria-expanded') === 'true'`,
			nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)},
		)
		require.NoError(t, err)

		require.NoError(t, second.Click())
		_, err = page.WaitForFunction(
			`() => document.querySelector('#controls-default-1')?.getAttribute('aria-expanded') === 'false' &&
				document.querySelector('#controls-default-2')?.getAttribute('aria-expanded') === 'true'`,
			nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)},
		)
		require.NoError(t, err)
	})

	t.Run("allow multiple variant keeps both selected items expanded", func(t *testing.T) {
		accordion := page.Locator("#accordion-multi")
		require.NoError(t, accordion.ScrollIntoViewIfNeeded())

		first := accordion.Locator("#controls-multi-1")
		second := accordion.Locator("#controls-multi-2")
		require.NoError(t, first.Click())
		require.NoError(t, second.Click())

		_, err := page.WaitForFunction(
			`() => document.querySelector('#controls-multi-1')?.getAttribute('aria-expanded') === 'true' &&
				document.querySelector('#controls-multi-2')?.getAttribute('aria-expanded') === 'true'`,
			nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)},
		)
		require.NoError(t, err)
	})

	t.Run("split variant renders each item as a separate bordered card", func(t *testing.T) {
		accordion := page.Locator("#accordion-split")
		require.NoError(t, accordion.ScrollIntoViewIfNeeded())

		classAttr, err := accordion.GetAttribute("class")
		require.NoError(t, err)
		require.Contains(t, classAttr, "flex")
		require.Contains(t, classAttr, "gap-4")

		cardClass, err := accordion.Locator("> div").First().GetAttribute("class")
		require.NoError(t, err)
		require.Contains(t, cardClass, "rounded-radius")
		require.Contains(t, cardClass, "border")
	})

	t.Run("lazy HTMX content loads after the item is revealed", func(t *testing.T) {
		lazy := page.Locator("#accordion-lazy")
		require.NoError(t, lazy.ScrollIntoViewIfNeeded())
		require.NoError(t, lazy.Locator("#controls-lazy-1").Click())
		require.NoError(t, lazy.Locator("#lazy-content-a").ScrollIntoViewIfNeeded())

		_, err := page.WaitForFunction(
			`() => document.querySelector('#lazy-content-a')?.textContent.includes('Server Response A')`,
			nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(4000)},
		)
		require.NoError(t, err)
		require.NoError(t, page.Locator("#lazy-content-a").GetByText("Loaded via HTMX").WaitFor())
	})

	for _, msg := range consoleErrors {
		require.False(t, strings.Contains(msg, "Alpine Expression Error"), "unexpected console error: %s", msg)
		require.False(t, strings.Contains(msg, "htmx:swapError"), "unexpected console error: %s", msg)
	}
}

func TestAccordionReducedMotionStopsCollapseChevronAndLazySpinnerMotion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newIsolatedPage(t)
	require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{
		ReducedMotion: playwright.ReducedMotionReduce,
	}))
	_, err := page.Goto(baseURL+"/components/accordion", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	_, err = page.WaitForFunction(`() => {
		const accordion = document.querySelector("#accordion-default");
		const chevron = accordion && accordion.querySelector("button svg");
		const region = accordion && accordion.querySelector("[x-collapse]");
		const spinner = document.querySelector("#lazy-content-a svg.animate-spin");
		return chevron && region && spinner &&
			getComputedStyle(chevron).transitionProperty === "none" &&
			getComputedStyle(region).transitionProperty === "none" &&
			getComputedStyle(spinner).animationName === "none";
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "Accordion motion should be disabled under reduced-motion")

	button := page.Locator("#accordion-default button").First()
	require.NoError(t, button.Click())
	_, err = page.WaitForFunction(`() => {
		const button = document.querySelector("#accordion-default button");
		const region = document.querySelector("#accordion-default [x-collapse]");
		return button && button.getAttribute("aria-expanded") === "true" && region &&
			getComputedStyle(region).transitionProperty === "none";
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err, "Accordion should still expand while reduced-motion removes the transition")
}
