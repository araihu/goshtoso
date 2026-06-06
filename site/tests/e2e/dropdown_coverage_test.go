package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDropdownCoverageDemo(t *testing.T) {
	page := newPage(t, sharedBrowser)

	var jsErrors []string
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})
	page.On("pageerror", func(err error) {
		jsErrors = append(jsErrors, err.Error())
	})

	_, err := page.Goto(baseURL+"/components/dropdown", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	require.NoError(t, page.Locator("#dropdown-click").ScrollIntoViewIfNeeded())
	require.NoError(t, page.Locator("main").GetByText("Dropdown").First().WaitFor())

	hoverTrigger := page.Locator("#dropdown-hover button").First()
	require.NoError(t, hoverTrigger.Hover())
	require.NoError(t, page.Locator("#dropdown-hover [role='menu']").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	hoverExpanded, err := hoverTrigger.Evaluate("el => el.getAttribute('aria-expanded')", nil)
	require.NoError(t, err)
	assert.Equal(t, "true", hoverExpanded)

	contextTrigger := page.Locator("#dropdown-context button").First()
	require.NoError(t, contextTrigger.Click())
	require.NoError(t, page.Locator("#dropdown-context ul[role='none']").First().WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	contextExpanded, err := contextTrigger.Evaluate("el => el.getAttribute('aria-expanded')", nil)
	require.NoError(t, err)
	assert.Equal(t, "true", contextExpanded)
	contextItemTag, err := page.Locator("#dropdown-context [role='menuitem']").First().Evaluate("el => el.tagName.toLowerCase()", nil)
	require.NoError(t, err)
	assert.Equal(t, "li", contextItemTag)
	require.NoError(t, page.Keyboard().Press("Escape"))
	_, err = page.WaitForFunction(`() => document.querySelector("#dropdown-context button").getAttribute("aria-expanded") === "false"`, nil)
	require.NoError(t, err)

	actionTrigger := page.Locator("#dropdown-actions button").First()
	require.NoError(t, actionTrigger.ScrollIntoViewIfNeeded())
	ariaLabel, err := actionTrigger.GetAttribute("aria-label")
	require.NoError(t, err)
	assert.Equal(t, "cluster actions", ariaLabel)
	svgCount, err := actionTrigger.Locator("svg").Count()
	require.NoError(t, err)
	assert.Equal(t, 1, svgCount)

	require.NoError(t, actionTrigger.Click())
	_, err = page.WaitForFunction(`() => document.querySelector("#dropdown-actions button").getAttribute("aria-expanded") === "true"`, nil)
	require.NoError(t, err)

	editItem := page.Locator("#dropdown-actions-edit")
	editTag, err := editItem.Evaluate("el => el.tagName.toLowerCase()", nil)
	require.NoError(t, err)
	assert.Equal(t, "button", editTag)
	require.NoError(t, editItem.Click())
	_, err = page.WaitForFunction(`() => {
		const el = document.querySelector('[x-data*="editOpen"]');
		return el && el._x_dataStack[0].editOpen === true && el._x_dataStack[0].editCount === 1;
	}`, nil)
	require.NoError(t, err)

	archiveItem := page.Locator("#dropdown-actions-archive")
	disabled, err := archiveItem.Evaluate("el => el.hasAttribute('disabled')", nil)
	require.NoError(t, err)
	assert.Equal(t, true, disabled)
	archiveClass, err := archiveItem.GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, archiveClass, "pointer-events-none")

	deleteClass, err := page.Locator("#dropdown-actions-delete").GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, deleteClass, "text-danger")

	require.Empty(t, jsErrors, "no JS console/page errors on dropdown demo: %v", jsErrors)
}
