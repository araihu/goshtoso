package e2e

import (
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

	html := renderInteractiveDocument(t, head.DependenciesMinimal(), carousel.Carousel(carousel.Config{
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
	require.NoError(t, page.SetContent(html, playwright.PageSetContentOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	}))
	_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined'`, nil)
	require.NoError(t, err)

	require.NoError(t, page.GetByRole("link", playwright.PageGetByRoleOptions{
		Name: "Open",
	}).Click())
	requireNoDialog(t, dialogSeen, "carousel CTA executed javascript: href")
}
