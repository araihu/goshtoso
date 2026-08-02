//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

const (
	radioDemoURL  = "/components/radio"
	ratingDemoURL = "/components/rating"
)

func newIsolatedPage(t *testing.T) playwright.Page {
	t.Helper()
	ctx, err := sharedBrowser.NewContext()
	require.NoError(t, err)
	t.Cleanup(func() { _ = ctx.Close() })

	page, err := ctx.NewPage()
	require.NoError(t, err)
	page.SetDefaultTimeout(3000)
	page.SetDefaultNavigationTimeout(5000)
	t.Cleanup(func() { _ = page.Close() })
	return page
}

func waitForAlpine(page playwright.Page) error {
	_, err := page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil, playwright.PageWaitForFunctionOptions{
		Timeout: playwright.Float(3000),
	})
	return err
}

func dismissCookieBanner(t *testing.T, page playwright.Page) {
	t.Helper()
	require.NoError(t, page.AddInitScript(playwright.Script{
		Content: new("try{document.cookie='gt_storage=allowed; Path=/; SameSite=Lax'}catch(e){}"),
	}))
}

func navigateToRadioDemo(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.Goto(baseURL+radioDemoURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(5000),
	})
	require.NoError(t, err)
	require.NoError(t, page.Locator("[data-testid='radio-default-group']").WaitFor(
		playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateAttached,
			Timeout: playwright.Float(3000),
		},
	))
	_, err = page.WaitForFunction(
		`() => typeof window.Alpine !== 'undefined'`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	require.NoError(t, err)
}

func navigateToRatingDemo(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.Goto(baseURL+ratingDemoURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(5000),
	})
	require.NoError(t, err)
	require.NoError(t, page.Locator("#rating-fragment").WaitFor(
		playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateAttached,
			Timeout: playwright.Float(3000),
		},
	))
	_, err = page.WaitForFunction(
		`() => typeof window.Alpine !== 'undefined'`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	require.NoError(t, err)
}

func gotoTagsList(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.Goto(baseURL+"/components/tags-list", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
}

func fillAndDispatchInput(t *testing.T, locator playwright.Locator, value string) {
	t.Helper()
	require.NoError(t, locator.Fill(value))
	_, err := locator.Evaluate(`(el) => el.dispatchEvent(new Event('input', {bubbles: true}))`, nil)
	require.NoError(t, err)
}

func waitForCarouselIndex(page playwright.Page, selector string, index int) error {
	_, err := page.WaitForFunction(
		`([selector, index]) => {
			const el = document.querySelector(selector);
			return !!el && Alpine.$data(el).currentSlideIndex === index;
		}`,
		[]any{selector, index},
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	return err
}

func waitForTabSelected(page playwright.Page, selector string, selected bool) error {
	expected := "false"
	if selected {
		expected = "true"
	}
	_, err := page.WaitForFunction(
		`([selector, expected]) => document.querySelector(selector)?.getAttribute('aria-selected') === expected`,
		[]any{selector, expected},
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	return err
}

func ratingChecked(t *testing.T, page playwright.Page, selector string) bool {
	t.Helper()
	value, err := page.Locator(selector).Evaluate("el => el.checked", nil)
	require.NoError(t, err)
	checked, ok := value.(bool)
	require.True(t, ok)
	return checked
}

func renderComponentFragment(t *testing.T, component templ.Component) string {
	t.Helper()
	var buffer bytes.Buffer
	require.NoError(t, component.Render(context.Background(), &buffer))
	return buffer.String()
}

func renderInteractiveDocument(t *testing.T, components ...templ.Component) string {
	t.Helper()
	var rendered []string
	for _, component := range components {
		rendered = append(rendered, renderComponentFragment(t, component))
	}

	var headHTML []string
	var bodyHTML []string
	for _, fragment := range rendered {
		switch {
		case strings.Contains(fragment, "<link") || strings.Contains(fragment, "<script"):
			headHTML = append(headHTML, fragment)
		default:
			bodyHTML = append(bodyHTML, fragment)
		}
	}
	return "<!doctype html><html><head>" + strings.Join(headHTML, "") + "</head><body>" + strings.Join(bodyHTML, "") + "</body></html>"
}

func listenForDialogs(t *testing.T, page playwright.Page) <-chan string {
	t.Helper()
	dialogSeen := make(chan string, 1)
	page.On("dialog", func(dialog playwright.Dialog) {
		select {
		case dialogSeen <- dialog.Message():
		default:
		}
		require.NoError(t, dialog.Accept())
	})
	return dialogSeen
}

func requireNoDialog(t *testing.T, dialogSeen <-chan string, context string) {
	t.Helper()
	select {
	case message := <-dialogSeen:
		t.Fatalf("%s: %s", context, message)
	case <-time.After(300 * time.Millisecond):
	}
}

func requireComponentGoAPILink(t *testing.T, page playwright.Page, entry catalog.Entry) {
	t.Helper()
	reference := page.Locator("[data-go-api-reference]")
	version, err := reference.GetAttribute("data-go-api-version")
	require.NoError(t, err)
	require.Equal(t, goshtosoDocsVersion, version)

	link := reference.Locator("[data-go-api-link]")
	require.Equal(t, 1, mustLocatorCount(t, link))
	href, err := link.GetAttribute("href")
	require.NoError(t, err)
	require.Equal(t, entry.GoDocsURL(goshtosoDocsVersion), href)
	require.Equal(t, "_blank", mustAttribute(t, link, "target"))
	require.Equal(t, "noopener noreferrer", mustAttribute(t, link, "rel"))
}

func mustLocatorCount(t *testing.T, locator playwright.Locator) int {
	t.Helper()
	count, err := locator.Count()
	require.NoError(t, err)
	return count
}

func mustAttribute(t *testing.T, locator playwright.Locator, name string) string {
	t.Helper()
	value, err := locator.GetAttribute(name)
	require.NoError(t, err)
	return value
}

func mustText(t *testing.T, locator playwright.Locator) string {
	t.Helper()
	text, err := locator.TextContent()
	require.NoError(t, err)
	return text
}

func mustCount(t *testing.T, locator playwright.Locator) int {
	t.Helper()
	count, err := locator.Count()
	require.NoError(t, err)
	return count
}

func mustBeVisible(t *testing.T, locator playwright.Locator) bool {
	t.Helper()
	visible, err := locator.IsVisible()
	require.NoError(t, err)
	return visible
}

func mustContainText(t *testing.T, locator playwright.Locator, expected string) bool {
	t.Helper()
	text, err := locator.TextContent()
	require.NoError(t, err)
	return strings.Contains(text, expected)
}

var onePxPNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
	0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
	0x42, 0x60, 0x82,
}
