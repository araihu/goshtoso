package e2e

import (
	"strings"
	"testing"

	"github.com/araihu/goshtoso/components/carousel"
	"github.com/araihu/goshtoso/components/head"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestCarouselRejectsExecutableCTAHref(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	dialogSeen := listenForDialogs(t, page)

	_, err := page.Goto(baseURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	require.NoError(t, err)

	html := renderInteractiveDocument(t, head.DependenciesMinimal(head.WithLocalRuntime()), carousel.Carousel(carousel.Config{
		ID: "security-carousel",
		Slides: []carousel.Slide{
			{
				ImgSrc:      "/assets/images/avatars/avatar-1.webp",
				ImgAlt:      "Profile",
				Title:       "Security slide",
				Description: "CTA should never execute script URLs.",
				CTAHref:     `javascript:alert('carousel-cta-xss')`,
				CTALabel:    "Open",
			},
		},
	}))
	// SetContent injects an already-complete document, where Chromium does not
	// preserve parser-deferred ordering. Make this synthetic fixture blocking so
	// the bundle retains its production order before Alpine's initial scan.
	html = strings.ReplaceAll(html, "<script defer ", "<script ")
	require.NoError(t, page.SetContent(html, playwright.PageSetContentOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	}))
	_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined'`, nil)
	require.NoError(t, err)

	link := page.GetByRole("link", playwright.PageGetByRoleOptions{
		Name: "Open",
	})
	require.NoError(t, link.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	href, err := link.GetAttribute("href")
	require.NoError(t, err)
	require.Equal(t, "about:invalid#TemplFailedSanitizationURL", href)
	_, err = link.Evaluate("element => element.click()", nil)
	require.NoError(t, err)
	requireNoDialog(t, dialogSeen, "carousel CTA executed javascript: href")
}
