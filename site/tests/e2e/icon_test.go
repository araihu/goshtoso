package e2e

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/araihu/goshtoso/components/icon/heroicons"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const iconCopiedMagnifyingGlassSource = `package main

import (
    "context"
    "log"
    "os"

    "github.com/araihu/goshtoso/components/icon"
    "github.com/araihu/goshtoso/components/icon/heroicons"
)

func main() {
    if err := icon.Icon(icon.Config{
        SpriteURL: heroicons.SpriteURL,
        Symbol:    heroicons.Icon16SolidMagnifyingGlass,
        Size:      icon.SizeLG,
        Label:     "Search",
        RootClass: "text-accent",
    }).Render(context.Background(), os.Stdout); err != nil {
        log.Fatal(err)
    }
}
`

func TestIconCatalogSpriteWorkbench(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	t.Run("route renders every generated glyph and symbol reference", func(t *testing.T) {
		page := openIconCatalogPage(t)
		require.NoError(t, page.GetByRole("heading", playwright.PageGetByRoleOptions{
			Name: "Heroicons, ready for Go",
		}).WaitFor())

		cards := page.Locator("[data-icon-card]")
		count, err := cards.Count()
		require.NoError(t, err)
		assert.Equal(t, len(heroicons.Glyphs), count)
		assert.Equal(t, 67, count)

		for _, glyph := range heroicons.Glyphs {
			card := page.Locator("[data-icon-card='" + glyph.GoName + "']")
			require.NoError(t, card.WaitFor())
			symbol, err := card.GetAttribute("data-icon-symbol")
			require.NoError(t, err)
			assert.Equal(t, string(glyph.Symbol), symbol)
			href, err := card.Locator("svg use").GetAttribute("href")
			require.NoError(t, err)
			assert.Equal(t, heroicons.SpriteURL+"#"+string(glyph.Symbol), href)
		}
	})

	t.Run("grid uses one three and six columns at responsive breakpoints", func(t *testing.T) {
		page := openIconCatalogPage(t)
		grid := page.Locator("[data-testid='icon-catalog-grid']")
		require.NoError(t, grid.WaitFor())

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
			assert.Equal(t, viewport.columns, actual, "viewport width %d", viewport.width)
		}
	})

	t.Run("showcase exposes accessible decorative and painted currentColor SVGs", func(t *testing.T) {
		page := openIconCatalogPage(t)

		accessible := page.Locator("[data-testid='icon-variant-accessible'] svg")
		require.NoError(t, accessible.WaitFor())
		assertSVGAttributes(t, accessible, "img", "Approved", "")

		decorative := page.Locator("[data-testid='icon-variant-decorative'] svg")
		require.NoError(t, decorative.WaitFor())
		assertSVGAttributes(t, decorative, "", "", "true")

		currentColor := page.Locator("[data-testid='icon-variant-current-color'] svg")
		require.NoError(t, currentColor.WaitFor())
		assertSVGGeometryAndCurrentColor(t, currentColor)
	})

	t.Run("magnifying glass modal traps focus and returns it on escape", func(t *testing.T) {
		page, card, dialog := openMagnifyingGlassModal(t)
		_, err := page.WaitForFunction(`() => document.activeElement?.id === "icon-size"`, nil)
		require.NoError(t, err)

		first := dialog.Locator("[data-testid='icon-modal-close']")
		last := dialog.Locator("#icon-root-class")
		require.NoError(t, last.Focus())
		require.NoError(t, page.Keyboard().Press("Tab"))
		_, err = page.WaitForFunction(`() => document.activeElement?.dataset?.testid === "icon-modal-close"`, nil)
		require.NoError(t, err)

		require.NoError(t, first.Focus())
		require.NoError(t, page.Keyboard().Press("Shift+Tab"))
		_, err = page.WaitForFunction(`() => document.activeElement?.id === "icon-root-class"`, nil)
		require.NoError(t, err)

		require.NoError(t, page.Keyboard().Press("Escape"))
		require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateHidden,
		}))
		focused, err := card.Evaluate("el => el === document.activeElement", nil)
		require.NoError(t, err)
		assert.Equal(t, true, focused)
	})

	t.Run("options update painted preview and copy exact standalone Go source", func(t *testing.T) {
		page, _, dialog := openMagnifyingGlassModal(t)
		_, err := page.Locator("#icon-size").SelectOption(playwright.SelectOptionValues{Values: &[]string{"lg"}})
		require.NoError(t, err)
		require.NoError(t, page.Locator("#icon-label").Fill("Search"))
		decorativeControl := page.Locator("#icon-fragment input[type='checkbox']")
		isDecorative, err := decorativeControl.IsChecked()
		require.NoError(t, err)
		assert.False(t, isDecorative)
		require.NoError(t, page.Locator("#icon-root-class").Fill("text-accent"))

		preview := dialog.Locator("[data-testid='icon-live-preview']")
		require.NoError(t, preview.WaitFor())
		assertSVGAttributes(t, preview, "img", "Search", "")
		assertSVGGeometryAndCurrentColor(t, preview)
		href, err := preview.Locator("use").GetAttribute("href")
		require.NoError(t, err)
		assert.Equal(t, heroicons.SpriteURL+"#"+string(heroicons.Icon16SolidMagnifyingGlass), href)
		previewClass, err := preview.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, previewClass, "size-6")
		assert.Contains(t, previewClass, "text-accent")

		copyButton := dialog.GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Copy compilable Go"})
		require.NoError(t, copyButton.Click())
		require.NoError(t, dialog.GetByText("Copied Go source", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())
		require.NoError(t, dialog.GetByText("Compilable Go source copied to clipboard.", playwright.LocatorGetByTextOptions{Exact: new(true)}).WaitFor())

		copied, err := page.Evaluate("() => window.__iconClipboard", nil)
		require.NoError(t, err)
		assert.Equal(t, iconCopiedMagnifyingGlassSource, copied)
		compileCopiedIconSource(t, iconCopiedMagnifyingGlassSource)
	})

	t.Run("decorative and blank labels hide the live preview", func(t *testing.T) {
		page, _, dialog := openMagnifyingGlassModal(t)
		decorativeControl := page.Locator("#icon-fragment input[type='checkbox']")
		require.NoError(t, decorativeControl.Check())
		preview := dialog.Locator("[data-testid='icon-live-preview']")
		require.NoError(t, preview.WaitFor())
		assertSVGAttributes(t, preview, "", "", "true")

		require.NoError(t, decorativeControl.Uncheck())
		require.NoError(t, page.Locator("#icon-label").Fill(""))
		assertSVGAttributes(t, preview, "", "", "true")
	})

	t.Run("bundled sprite and license return success", func(t *testing.T) {
		_ = openIconCatalogPage(t)
		for _, path := range []string{"/assets/icons/heroicons.svg", "/assets/icons/HEROICONS_LICENSE.txt"} {
			response, err := http.Get(baseURL + path)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode, path)
			require.NoError(t, response.Body.Close())
		}
	})
}

func openIconCatalogPage(t *testing.T) playwright.Page {
	t.Helper()
	page := newPage(t, sharedBrowser)
	failures := watchPageFailures(page)
	stubIconClipboard(t, page)

	_, err := page.Goto(baseURL+"/components/icon", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))
	require.NoError(t, page.Locator("#icon-fragment").WaitFor())
	t.Cleanup(func() {
		waitForPageSettled(t, page)
		failures.RequireEmpty(t)
	})
	return page
}

func openMagnifyingGlassModal(t *testing.T) (playwright.Page, playwright.Locator, playwright.Locator) {
	t.Helper()
	page := openIconCatalogPage(t)
	card := page.Locator("[data-icon-card='Icon16SolidMagnifyingGlass']")
	dialog := page.Locator("[data-testid='icon-picker-dialog']")
	require.NoError(t, card.Click())
	require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	return page, card, dialog
}

func assertSVGAttributes(t *testing.T, svg playwright.Locator, role, label, hidden string) {
	t.Helper()
	actualRole, err := svg.GetAttribute("role")
	require.NoError(t, err)
	assert.Equal(t, role, actualRole)
	actualLabel, err := svg.GetAttribute("aria-label")
	require.NoError(t, err)
	assert.Equal(t, label, actualLabel)
	actualHidden, err := svg.GetAttribute("aria-hidden")
	require.NoError(t, err)
	assert.Equal(t, hidden, actualHidden)
}

func assertSVGGeometryAndCurrentColor(t *testing.T, svg playwright.Locator) {
	t.Helper()
	actual, err := svg.Evaluate(`el => {
		const use = el.querySelector("use");
		const glyph = use.getBBox();
		const bounds = el.getBoundingClientRect();
		return {
			glyphWidth: glyph.width,
			glyphHeight: glyph.height,
			width: bounds.width,
			height: bounds.height,
			color: getComputedStyle(el).color,
			parentColor: getComputedStyle(el.parentElement).color,
		};
	}`, nil)
	require.NoError(t, err)
	geometry, ok := actual.(map[string]interface{})
	require.True(t, ok)
	assert.Greater(t, svgMeasurement(t, geometry, "glyphWidth"), float64(0))
	assert.Greater(t, svgMeasurement(t, geometry, "glyphHeight"), float64(0))
	assert.Greater(t, svgMeasurement(t, geometry, "width"), float64(0))
	assert.Greater(t, svgMeasurement(t, geometry, "height"), float64(0))
	assert.Equal(t, geometry["parentColor"], geometry["color"])
}

func svgMeasurement(t *testing.T, geometry map[string]interface{}, name string) float64 {
	t.Helper()
	switch value := geometry[name].(type) {
	case float64:
		return value
	case int:
		return float64(value)
	default:
		t.Fatalf("%s measurement has unexpected type %T", name, geometry[name])
		return 0
	}
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
	directory, err := os.MkdirTemp(root, "_icon-e2e-example-*")
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
