//go:build e2e && full && goshtoso_current_source

package e2e

import (
	"bytes"
	"encoding/json"
	"image/png"
	"io"
	"net/http"
	"strings"
	"testing"

	demothemes "github.com/araihu/goshtoso/site/internal/themes"
	corethemes "github.com/araihu/goshtoso/themes"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

type renderedTheme struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

func TestThemePage_PublicCatalogInventoryAndSocialContract(t *testing.T) {
	publicByKey := make(map[string]string)
	for _, theme := range corethemes.BuiltIn() {
		key := string(theme.Key)
		require.NotContains(t, publicByKey, key)
		publicByKey[key] = theme.Label
	}

	siteCatalog := demothemes.All()
	expected := make([]renderedTheme, 0, len(siteCatalog))
	for _, theme := range siteCatalog {
		canonical, ok := publicByKey[theme.Key]
		require.True(t, ok, "site theme %q missing from public catalog", theme.Key)
		if override, overridden := demothemes.PresentationLabelOverride(theme.Key); overridden {
			require.Equal(t, override, theme.Label)
		} else {
			require.Equal(t, canonical, theme.Label, "canonical label drift for %q", theme.Key)
		}
		expected = append(expected, renderedTheme{Key: theme.Key, Label: theme.Label})
	}
	require.Len(t, siteCatalog, len(publicByKey))
	require.Equal(t, "Zombie", publicByKey["zombie"])
	require.Equal(t, "Halloween II", demothemes.ZombiePresentationLabelOverride)

	expectedKeys := make([]string, 0, len(expected))
	for _, theme := range expected {
		expectedKeys = append(expectedKeys, theme.Key)
	}

	page := newIsolatedPage(t)
	dismissCookieBanner(t, page)
	failures := watchPageFailures(page)
	_, err := page.Goto(baseURL+"/getting-started", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	themeLink := page.Locator("nav[aria-label='sidebar navigation'] a[href='/docs/theme']").First()
	_, err = page.ExpectResponse("**/docs/theme", func() error {
		return themeLink.Click()
	}, playwright.PageExpectResponseOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err)
	require.NoError(t, page.WaitForURL("**/docs/theme"))
	_, err = page.WaitForFunction(`() => {
		const root = document.querySelector('[x-data="themePage"]');
		const data = root && Alpine.$data(root);
		return root && root._x_dataStack && data && data.allThemes.length > 0;
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)

	for _, mode := range []struct {
		name  string
		theme string
		dark  bool
	}{
		{name: "GoshtosoLight", theme: "goshtoso"},
		{name: "GoshtosoDark", theme: "goshtoso", dark: true},
		{name: "MinimalLight", theme: "minimal"},
		{name: "MinimalDark", theme: "minimal", dark: true},
	} {
		t.Run(mode.name, func(t *testing.T) {
			setSidebarThemeMode(t, page, mode.theme, mode.dark)
			assertRenderedThemeInventory(t, page, expectedKeys, expected)
		})
	}
	waitForPageSettled(t, page)
	failures.RequireEmpty(t)

	response, err := http.Get(baseURL + "/docs/theme")
	require.NoError(t, err)
	t.Cleanup(func() { _ = response.Body.Close() })
	require.Equal(t, http.StatusOK, response.StatusCode)
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	html := string(body)
	title := "Themes - Goshtoso UI Library for Go"
	description := "Customize Goshtoso themes with Tailwind CSS tokens, dark mode, live previews, and server-rendered component examples."
	imageURL := "https://goshtoso.araihu.com/assets/images/goshtoso-social-card.png"
	alt := title + " — Goshtoso Go UI component library preview"
	for _, tag := range []string{
		"<title>" + title + "</title>",
		`<meta name="description" content="` + description + `">`,
		`<link rel="canonical" href="https://goshtoso.araihu.com/docs/theme">`,
		`<meta property="og:url" content="https://goshtoso.araihu.com/docs/theme">`,
		`<meta property="og:type" content="website">`,
		`<meta property="og:title" content="` + title + `">`,
		`<meta property="og:description" content="` + description + `">`,
		`<meta property="og:site_name" content="Goshtoso">`,
		`<meta property="og:image" content="` + imageURL + `">`,
		`<meta property="og:image:type" content="image/png">`,
		`<meta property="og:image:width" content="1200">`,
		`<meta property="og:image:height" content="630">`,
		`<meta property="og:image:alt" content="` + alt + `">`,
		`<meta name="twitter:card" content="summary_large_image">`,
		`<meta name="twitter:title" content="` + title + `">`,
		`<meta name="twitter:description" content="` + description + `">`,
		`<meta name="twitter:image" content="` + imageURL + `">`,
		`<meta name="twitter:image:alt" content="` + alt + `">`,
	} {
		require.Equal(t, 1, strings.Count(html, tag), "initial HTML must contain exactly one %s", tag)
	}

	imageResponse, err := http.Get(baseURL + "/assets/images/goshtoso-social-card.png")
	require.NoError(t, err)
	t.Cleanup(func() { _ = imageResponse.Body.Close() })
	require.Equal(t, http.StatusOK, imageResponse.StatusCode)
	require.Equal(t, "image/png", strings.Split(imageResponse.Header.Get("Content-Type"), ";")[0])
	imageBody, err := io.ReadAll(imageResponse.Body)
	require.NoError(t, err)
	require.Less(t, len(imageBody), 1<<20, "social preview image must stay below 1 MiB")
	config, err := png.DecodeConfig(bytes.NewReader(imageBody))
	require.NoError(t, err)
	require.Equal(t, 1200, config.Width)
	require.Equal(t, 630, config.Height)
}

func assertRenderedThemeInventory(t *testing.T, page playwright.Page, expectedKeys []string, expected []renderedTheme) {
	t.Helper()
	rawInventory, err := page.Evaluate(`() => {
		const themePage = document.querySelector('[x-data="themePage"]');
		const selectorKeys = Alpine.$data(themePage).allThemes;
		const cards = [...document.querySelectorAll('h2 + .grid button[data-theme-key]')].map(card => ({
			key: card.dataset.themeKey,
			label: card.querySelector('.text-xs.font-bold').textContent.trim(),
		}));
		return JSON.stringify({selectorKeys, cards});
	}`, nil)
	require.NoError(t, err)
	inventoryJSON, ok := rawInventory.(string)
	require.True(t, ok, "unexpected inventory type %T", rawInventory)
	var inventory struct {
		SelectorKeys []string        `json:"selectorKeys"`
		Cards        []renderedTheme `json:"cards"`
	}
	require.NoError(t, json.Unmarshal([]byte(inventoryJSON), &inventory))
	require.Equal(t, expectedKeys, inventory.SelectorKeys, "rendered theme selector state must follow site catalog")
	require.Equal(t, expected, inventory.Cards, "rendered theme cards must follow site presentation catalog")
}
