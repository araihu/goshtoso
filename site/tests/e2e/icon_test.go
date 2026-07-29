package e2e

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIconCatalogSpriteWorkbench(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	failures := watchPageFailures(page)
	stubIconClipboard(t, page)

	_, err := page.Goto(baseURL+"/components/icon", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	t.Run("route renders every bundled sprite card", func(t *testing.T) {
		require.NoError(t, page.Locator("#icon-fragment").WaitFor())
		require.NoError(t, page.GetByRole("heading", playwright.PageGetByRoleOptions{
			Name: "Heroicons, ready for Go",
		}).WaitFor())

		cards := page.Locator("[data-icon-card]")
		count, err := cards.Count()
		require.NoError(t, err)
		assert.Equal(t, 67, count)

		uses := cards.Locator("svg use")
		useCount, err := uses.Count()
		require.NoError(t, err)
		assert.Equal(t, 67, useCount)

		for _, index := range []int{0, 41, 66} {
			href, err := uses.Nth(index).GetAttribute("href")
			require.NoError(t, err)
			assert.Truef(t, strings.HasPrefix(href, "/assets/icons/heroicons.svg#hi-"), "card %d href = %q", index, href)
		}
	})

	t.Run("grid uses one three and six columns at responsive breakpoints", func(t *testing.T) {
		grid := page.Locator("#icon-catalog-title").Locator("xpath=following-sibling::div[1]")
		for _, viewport := range []struct {
			width   int
			columns int
		}{
			{width: 390, columns: 1},
			{width: 800, columns: 3},
			{width: 1440, columns: 6},
		} {
			require.NoError(t, page.SetViewportSize(viewport.width, 900))
			actual, err := grid.Evaluate(`el => getComputedStyle(el).gridTemplateColumns.trim().split(/\s+/).filter(Boolean).length`, nil)
			require.NoError(t, err)
			assert.Equal(t, float64(viewport.columns), actual, "viewport width %d", viewport.width)
		}
	})

	t.Run("showcase demonstrates accessibility and inherited currentColor", func(t *testing.T) {
		accessible := page.Locator("[data-variant='accessible'] svg")
		require.NoError(t, accessible.WaitFor())
		role, err := accessible.GetAttribute("role")
		require.NoError(t, err)
		assert.Equal(t, "img", role)
		label, err := accessible.GetAttribute("aria-label")
		require.NoError(t, err)
		assert.Equal(t, "Approved", label)

		decorative := page.Locator("[data-variant='decorative'] svg")
		hidden, err := decorative.GetAttribute("aria-hidden")
		require.NoError(t, err)
		assert.Equal(t, "true", hidden)
		decorativeRole, err := decorative.GetAttribute("role")
		require.NoError(t, err)
		assert.Empty(t, decorativeRole)

		currentColor := page.Locator("[data-variant='current-color'] svg")
		colors, err := currentColor.Evaluate(`el => ({ icon: getComputedStyle(el).color, parent: getComputedStyle(el.parentElement).color, fill: el.getAttribute('fill'), stroke: el.getAttribute('stroke') })`, nil)
		require.NoError(t, err)
		colorMap, ok := colors.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, colorMap["parent"], colorMap["icon"])
		assert.Nil(t, colorMap["fill"])
		assert.Nil(t, colorMap["stroke"])
	})

	card := page.Locator("[data-icon-card='Icon16SolidMagnifyingGlass']")
	dialog := page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: "Icon16SolidMagnifyingGlass"})

	t.Run("selected card opens focus-trapped modal and returns focus on escape", func(t *testing.T) {
		require.NoError(t, card.Click())
		require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		}))
		_, err := page.WaitForFunction(`() => document.activeElement?.id === "icon-size"`, nil)
		require.NoError(t, err)

		require.NoError(t, page.Keyboard().Press("Escape"))
		require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateHidden,
		}))
		focused, err := card.Evaluate("el => el === document.activeElement", nil)
		require.NoError(t, err)
		assert.Equal(t, true, focused)
	})

	t.Run("options update live preview and copied standalone Go", func(t *testing.T) {
		require.NoError(t, card.Click())
		require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		}))

		_, err = page.Locator("#icon-size").SelectOption(playwright.SelectOptionValues{Values: &[]string{"lg"}})
		require.NoError(t, err)
		require.NoError(t, page.Locator("#icon-label").Fill("Search"))
		isDecorative, err := page.Locator("#icon-fragment input[type='checkbox']").IsChecked()
		require.NoError(t, err)
		assert.False(t, isDecorative)
		require.NoError(t, page.Locator("#icon-root-class").Fill("text-accent"))

		preview := dialog.Locator("svg[role='img'][aria-label='Search']")
		require.NoError(t, preview.WaitFor())
		previewClass, err := preview.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, previewClass, "size-6")
		assert.Contains(t, previewClass, "text-accent")
		href, err := preview.Locator("use").GetAttribute("href")
		require.NoError(t, err)
		assert.Equal(t, "/assets/icons/heroicons.svg#hi-16-solid-magnifying-glass", href)

		copyButton := dialog.GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Copy compilable Go"})
		require.NoError(t, copyButton.Click())
		require.NoError(t, dialog.GetByText("Copied Go source", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())
		require.NoError(t, dialog.GetByText("Compilable Go source copied to clipboard.", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())

		copied, err := page.Evaluate("() => window.__iconClipboard", nil)
		require.NoError(t, err)
		copiedSource, ok := copied.(string)
		require.True(t, ok)
		for _, want := range []string{
			"package main",
			`"github.com/araihu/goshtoso/components/icon"`,
			`"github.com/araihu/goshtoso/components/icon/heroicons"`,
			"SpriteURL: heroicons.SpriteURL",
			"Symbol:    heroicons.Icon16SolidMagnifyingGlass",
			"Size:      icon.SizeLG",
			`Label:     "Search"`,
			`RootClass: "text-accent"`,
			".Render(context.Background(), os.Stdout)",
		} {
			assert.Contains(t, copiedSource, want)
		}
		assert.NotContains(t, copiedSource, "Decorative:")
		compileCopiedIconSource(t, copiedSource)
	})

	t.Run("decorative and blank labels hide preview from assistive technology", func(t *testing.T) {
		decorativeControl := page.Locator("#icon-fragment input[type='checkbox']")
		require.NoError(t, decorativeControl.Check())
		decorativePreview := dialog.Locator("svg[aria-hidden='true']")
		require.NoError(t, decorativePreview.WaitFor())
		decorativeLabel, err := decorativePreview.GetAttribute("aria-label")
		require.NoError(t, err)
		assert.Empty(t, decorativeLabel)
		decorativeRole, err := decorativePreview.GetAttribute("role")
		require.NoError(t, err)
		assert.Empty(t, decorativeRole)

		require.NoError(t, decorativeControl.Uncheck())
		require.NoError(t, page.Locator("#icon-label").Fill(""))
		require.NoError(t, dialog.Locator("svg[aria-hidden='true']").WaitFor())
		blankPreviewRole, err := dialog.Locator("svg[aria-hidden='true']").GetAttribute("role")
		require.NoError(t, err)
		assert.Empty(t, blankPreviewRole)
	})

	t.Run("bundled sprite and license are served", func(t *testing.T) {
		for _, path := range []string{"/assets/icons/heroicons.svg", "/assets/icons/HEROICONS_LICENSE.txt"} {
			response, err := http.Get(baseURL + path)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode, path)
			require.NoError(t, response.Body.Close())
		}
	})

	waitForPageSettled(t, page)
	failures.RequireEmpty(t)
}

func stubIconClipboard(t *testing.T, page playwright.Page) {
	t.Helper()
	require.NoError(t, page.AddInitScript(playwright.Script{Content: new(`(() => {
		Object.defineProperty(navigator, "clipboard", {
			configurable: true,
			value: { writeText: (value) => { window.__iconClipboard = value; return Promise.resolve(); } },
		});
	})();`)}))
}

func compileCopiedIconSource(t *testing.T, source string) {
	t.Helper()
	root := iconE2ERepoRoot(t)
	directory, err := os.MkdirTemp(root, "icon-e2e-example-*")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.RemoveAll(directory)) })
	require.NoError(t, os.WriteFile(filepath.Join(directory, "main.go"), []byte(source), 0o600))

	command := exec.Command("go", "run", "./"+filepath.Base(directory))
	command.Dir = root
	output, err := command.CombinedOutput()
	require.NoErrorf(t, err, "copied icon source did not compile: %s\n%s", err, output)
	require.Contains(t, string(output), `href="/assets/icons/heroicons.svg#hi-16-solid-magnifying-glass"`)
}

func iconE2ERepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "../../.."))
}
