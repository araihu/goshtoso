//go:build e2e

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckboxCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)

	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL+"/components/checkbox", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)
	require.NoError(t, page.Locator("#checkbox-fragment").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	mainText, err := page.Locator("main").TextContent()
	require.NoError(t, err)
	require.Contains(t, mainText, "Checkbox")

	for _, selector := range []string{
		"#checkbox-default",
		"#checkbox-colors",
		"#checkbox-icons",
		"#checkbox-animations",
		"#checkbox-description",
		"#checkbox-container",
		"div#checkbox-group",
		"#checkbox-disabled",
	} {
		require.NoError(t, page.Locator(selector).ScrollIntoViewIfNeeded())
		require.NoError(t, page.Locator(selector).WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		}), "%s should be visible", selector)
	}

	unchecked := page.Locator("#defaultUnchecked")
	require.NoError(t, unchecked.ScrollIntoViewIfNeeded())
	checked, err := unchecked.IsChecked()
	require.NoError(t, err)
	assert.False(t, checked)
	require.NoError(t, unchecked.Check())
	checked, err = unchecked.IsChecked()
	require.NoError(t, err)
	assert.True(t, checked)
	require.NoError(t, unchecked.Uncheck())
	checked, err = unchecked.IsChecked()
	require.NoError(t, err)
	assert.False(t, checked)

	variantClasses := map[string]string{
		"variantPrimary":   "checked:border-primary",
		"variantSecondary": "checked:border-secondary",
		"variantInfo":      "checked:border-info",
		"variantSuccess":   "checked:border-success",
		"variantWarning":   "checked:border-warning",
		"variantDanger":    "checked:border-danger",
	}
	for id, wantClass := range variantClasses {
		input := page.Locator("#" + id)
		className, err := input.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, className, wantClass, "%s should carry variant class", id)
		checked, err := input.IsChecked()
		require.NoError(t, err)
		assert.True(t, checked, "%s should start checked", id)
	}

	iconPaths := map[string]string{
		"iconXmark": "M6 18L18 6M6 6l12 12",
		"iconMinus": "M18 12H6",
		"iconPlus":  "M12 4.5v15m7.5-7.5h-15",
	}
	for id, wantPath := range iconPaths {
		path, err := page.Locator("label[for='" + id + "'] path").GetAttribute("d")
		require.NoError(t, err)
		assert.Equal(t, wantPath, path)
	}

	animSlideUpSVG := page.Locator("label[for='animSlideUp'] svg")
	slideUpClass, err := animSlideUpSVG.GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, slideUpClass, "peer-checked:-translate-y-1/2")

	animScaleUpInput := page.Locator("#animScaleUp")
	scaleClass, err := animScaleUpInput.GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, scaleClass, "checked:before:scale-125")

	animSlideDownInput := page.Locator("#animSlideDown")
	slideDownClass, err := animSlideDownInput.GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, slideDownClass, "checked:before:translate-y-0")

	description := page.Locator("#descriptionCheckbox")
	aria, err := description.GetAttribute("aria-describedby")
	require.NoError(t, err)
	assert.Equal(t, "checkboxDescription", aria)

	containerLabelClass, err := page.Locator("label[for='containerCheckbox']").GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, containerLabelClass, "justify-between")

	groupPush := page.Locator("#groupPush")
	checked, err = groupPush.IsChecked()
	require.NoError(t, err)
	assert.False(t, checked)
	require.NoError(t, groupPush.Check())
	checked, err = groupPush.IsChecked()
	require.NoError(t, err)
	assert.True(t, checked)

	disabled := page.Locator("#disabledUnchecked")
	isDisabled, err := disabled.IsDisabled()
	require.NoError(t, err)
	assert.True(t, isDisabled)
	checked, err = disabled.IsChecked()
	require.NoError(t, err)
	assert.False(t, checked)

	require.Empty(t, jsErrors, "no JS console/page errors on checkbox demo: %v", jsErrors)
}
