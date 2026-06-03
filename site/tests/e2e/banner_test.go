package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestBannerCookiePreviewStaysInsideDemoBox(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/components/banner", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	stage := page.Locator("#banner-cookie > div").First()
	dialog := page.Locator("#banner-cookie [role='dialog']").First()
	require.NoError(t, stage.WaitFor())
	require.NoError(t, dialog.WaitFor())

	classAttr, err := dialog.GetAttribute("class")
	require.NoError(t, err)
	require.Contains(t, classAttr, "absolute bottom-4")
	require.False(t, strings.Contains(classAttr, "fixed bottom-4"), "demo cookie banner should not be viewport fixed")

	stageBox, err := stage.BoundingBox()
	require.NoError(t, err)
	dialogBox, err := dialog.BoundingBox()
	require.NoError(t, err)

	require.GreaterOrEqual(t, dialogBox.X, stageBox.X)
	require.GreaterOrEqual(t, dialogBox.Y, stageBox.Y)
	require.LessOrEqual(t, dialogBox.X+dialogBox.Width, stageBox.X+stageBox.Width)
	require.LessOrEqual(t, dialogBox.Y+dialogBox.Height, stageBox.Y+stageBox.Height)
}
