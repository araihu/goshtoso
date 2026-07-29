package e2e

import (
	"strings"
	"sync"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestBannerCoverageDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, _ := setupPlaywright(t)
	page := newPage(t, browser)
	dismissCookieBanner(t, page)

	var (
		mu            sync.Mutex
		consoleErrors []string
	)
	page.On("console", func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			mu.Lock()
			consoleErrors = append(consoleErrors, msg.Text())
			mu.Unlock()
		}
	})

	_, err := page.Goto(baseURL+"/components/banner", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForAlpine(page))

	require.NoError(t, page.Locator("h1#banner").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	for _, id := range []string{
		"#banner-simple",
		"#banner-persistent",
		"#banner-cta",
		"#banner-variants",
		"#banner-cookie",
	} {
		require.NoError(t, page.Locator(id).WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateAttached,
		}))
	}

	cookieDialog := page.Locator("#banner-cookie [role='dialog']").First()
	require.NoError(t, cookieDialog.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))
	require.NoError(t, cookieDialog.Locator("button").Filter(playwright.LocatorFilterOptions{
		HasText: "Sounds Good!",
	}).Click())
	require.NoError(t, cookieDialog.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateHidden,
	}))

	ctaButton := page.Locator("#banner-cta button", playwright.PageLocatorOptions{
		HasText: "Start free trial",
	})
	require.NoError(t, ctaButton.ScrollIntoViewIfNeeded())
	require.NoError(t, ctaButton.Click())
	_, err = page.WaitForFunction(`() => document.querySelector("#banner-cta-result")?.textContent.includes("Banner action received")`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)})
	require.NoError(t, err)

	_, err = page.Evaluate(`() => {
		document.documentElement.setAttribute('data-theme', 'minimal')
		document.documentElement.classList.add('dark')
	}`, nil)
	require.NoError(t, err)
	require.NoError(t, page.Locator("#banner-variants [role='banner']").First().WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	mu.Lock()
	defer mu.Unlock()
	require.Empty(t, consoleErrors, "unexpected console errors: %s", strings.Join(consoleErrors, "; "))
}
