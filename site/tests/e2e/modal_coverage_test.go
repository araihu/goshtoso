//go:build e2e && (full || modal)

package e2e

import (
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModalCoverageDemo exercises the /components/modal demo: default open/close
// (button, ESC), the four alert-dialog tones, the HTMX confirm action, and the
// JavaScript OnClick action. It mirrors the low-coverage branches in
// components/modal so the browser path is exercised end to end.
func TestModalCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	// Collect console errors to assert the demo stays clean across interactions.
	var consoleErrors []string
	page.On("console", func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			consoleErrors = append(consoleErrors, msg.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/modal", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("#modal-fragment").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateAttached,
	}))

	t.Run("Page_Loads", func(t *testing.T) {
		title, err := page.Title()
		require.NoError(t, err)
		assert.Contains(t, title, "Modal")
	})

	t.Run("Default_Open_Close_Button", func(t *testing.T) {
		container := page.Locator("#modal-default")
		dialog := container.Locator("[role='dialog']")

		// Dialog hidden before opening (x-cloak/x-show false).
		require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateHidden,
		}))

		require.NoError(t, container.GetByText("Open Modal").Click())
		require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		}))

		ariaModal, err := dialog.GetAttribute("aria-modal")
		require.NoError(t, err)
		assert.Equal(t, "true", ariaModal)

		// Close via the header close button.
		require.NoError(t, dialog.Locator("button[aria-label='close modal']").Click())
		require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateHidden,
		}))
	})

	t.Run("Default_Close_With_Escape", func(t *testing.T) {
		container := page.Locator("#modal-default")
		dialog := container.Locator("[role='dialog']")

		require.NoError(t, container.GetByText("Open Modal").Click())
		require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		}))

		require.NoError(t, page.Keyboard().Press("Escape"))
		require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateHidden,
		}))
	})

	t.Run("Alert_Dialog_Tones_Open", func(t *testing.T) {
		container := page.Locator("#modal-alert")
		for _, label := range []string{"Success Modal", "Info Modal", "Warning Modal", "Danger Modal"} {
			// Each alert trigger lives in its own x-data root; open then close it.
			root := container.Locator("div[x-data]").Filter(playwright.LocatorFilterOptions{
				HasText: label,
			}).First()
			dialog := root.Locator("[role='alertdialog']")

			require.NoError(t, root.GetByText(label).Click())
			require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
				State: playwright.WaitForSelectorStateVisible,
			}), "alert dialog %q should open", label)

			// Alert dialog renders a single full-width CTA button.
			cta := dialog.Locator("button.w-full")
			require.NoError(t, cta.WaitFor(playwright.LocatorWaitForOptions{
				State: playwright.WaitForSelectorStateVisible,
			}), "alert %q should render full-width CTA", label)

			require.NoError(t, dialog.Locator("button[aria-label='close modal']").Click())
			require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
				State: playwright.WaitForSelectorStateHidden,
			}))
		}
	})

	t.Run("HTMX_Action_Swaps_Result", func(t *testing.T) {
		container := page.Locator("#modal-htmx")
		dialog := container.Locator("[role='dialog']")

		require.NoError(t, container.GetByText("Open HTMX Modal").Click())
		require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		}))

		// Confirm fires the POST and closes the modal.
		require.NoError(t, dialog.GetByRole("button", playwright.LocatorGetByRoleOptions{
			Name: "Confirm", Exact: new(true),
		}).Click())

		result := page.Locator("#modal-htmx-result")
		require.NoError(t, result.GetByText("Hello from HTMX!").WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		}))
		text, err := result.TextContent()
		require.NoError(t, err)
		assert.Contains(t, text, "POST")
	})

	t.Run("JS_Action_Fires_Dialog", func(t *testing.T) {
		container := page.Locator("#modal-js")
		dialog := container.Locator("[role='dialog']")

		// Auto-accept the alert() raised by the OnClick handler.
		var dialogMessage string
		page.Once("dialog", func(d playwright.Dialog) {
			dialogMessage = d.Message()
			_ = d.Accept()
		})

		require.NoError(t, container.GetByRole("button", playwright.LocatorGetByRoleOptions{
			Name: "Open JS Modal", Exact: new(true),
		}).Click())
		require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		}))

		require.NoError(t, dialog.GetByRole("button", playwright.LocatorGetByRoleOptions{
			Name: "Delete", Exact: new(true),
		}).Click())
		require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateHidden,
		}))
		assert.Contains(t, dialogMessage, "deleted")
	})

	t.Run("No_Console_Errors", func(t *testing.T) {
		assert.Empty(t, consoleErrors, "demo should not log console errors: %s", strings.Join(consoleErrors, "; "))
	})
}
