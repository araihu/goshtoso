package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestDrawerCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	var consoleErrors []string
	page.On("console", func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			consoleErrors = append(consoleErrors, msg.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/drawer", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	require.NoError(t, page.Locator("#drawer-default").WaitFor())
	require.NoError(t, page.Locator("#drawer-left").WaitFor())
	require.NoError(t, page.GetByRole("heading", playwright.PageGetByRoleOptions{
		Name: "Drawer",
	}).WaitFor())

	projectDialog := page.GetByRole("dialog", playwright.PageGetByRoleOptions{
		Name: "Project details",
	})
	require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Open details",
	}).Click())
	require.NoError(t, projectDialog.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	require.NoError(t, projectDialog.GetByText("Deployment target").WaitFor())

	require.NoError(t, page.Keyboard().Press("Escape"))
	require.NoError(t, projectDialog.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateHidden,
	}))

	filtersButton := page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Open filters",
	})
	require.NoError(t, filtersButton.ScrollIntoViewIfNeeded())
	require.NoError(t, filtersButton.Click())

	filtersDialog := page.GetByRole("dialog", playwright.PageGetByRoleOptions{
		Name: "Filters",
	})
	require.NoError(t, filtersDialog.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	require.NoError(t, filtersDialog.GetByText("Only active records").WaitFor())
	require.NoError(t, filtersDialog.GetByText("Needs attention").WaitFor())

	className, err := filtersDialog.GetAttribute("class")
	require.NoError(t, err)
	for _, want := range []string{"left-0", "border-r", "max-w-[320px]"} {
		if !strings.Contains(className, want) {
			t.Fatalf("filters drawer class %q missing %q", className, want)
		}
	}

	require.NoError(t, filtersDialog.Locator("#filtersDrawer-body").WaitFor())

	overlay := page.Locator("#drawer-left div[aria-hidden='true']")
	require.NoError(t, overlay.Click())
	require.NoError(t, filtersDialog.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateHidden,
	}))

	if len(consoleErrors) > 0 {
		t.Fatalf("console errors on drawer demo: %s", strings.Join(consoleErrors, "\n"))
	}
}

func TestDrawerCoverageDispatchIgnoresMismatchedIDs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/components/drawer", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	for _, dialogName := range []string{"Project details", "Filters"} {
		dialog := page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: dialogName})
		require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateHidden,
		}))
	}

	_, err = page.Evaluate(`() => window.dispatchEvent(new CustomEvent('drawer:open', { detail: { id: 'not-a-real-drawer' } }))`)
	require.NoError(t, err)

	for _, dialogName := range []string{"Project details", "Filters"} {
		dialog := page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: dialogName})
		visible, err := dialog.IsVisible()
		require.NoError(t, err)
		if visible {
			t.Fatalf("%s drawer opened after mismatched event id", dialogName)
		}
	}

	_, err = page.Evaluate(fmt.Sprintf(`() => window.dispatchEvent(new CustomEvent('drawer:open', { detail: { id: %q } }))`, "filtersDrawer"))
	require.NoError(t, err)
	require.NoError(t, page.GetByRole("dialog", playwright.PageGetByRoleOptions{
		Name: "Filters",
	}).WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
}
