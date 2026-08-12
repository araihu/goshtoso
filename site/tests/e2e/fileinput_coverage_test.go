//go:build e2e && (full || fileinput)

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestFileinputCoverageDemo exercises the file-input demo page beyond the
// variant snapshots in fileinput_test.go: it drives the drop zone's Alpine
// drag state machine (dragenter/dragleave toggling the highlight border) and
// asserts the page renders without JS console or page errors.
func TestFileinputCoverageDemo(t *testing.T) {
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

	_, err := page.Goto(baseURL+"/components/fileinput", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	require.NoError(t, page.Locator("main").First().WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}))

	mainText, err := page.Locator("main").First().TextContent()
	require.NoError(t, err)
	require.Contains(t, mainText, "File Input")

	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	// All five demo variants should render their drop zone / upload markers.
	dropZones := page.Locator("[data-fileinput-variant='dropzone']")
	dzCount, err := dropZones.Count()
	require.NoError(t, err)
	require.GreaterOrEqual(t, dzCount, 4, "default, required, disabled, nolabel drop zones")

	uploads := page.Locator("[data-fileinput-variant='upload']")
	upCount, err := uploads.Count()
	require.NoError(t, err)
	require.Equal(t, 1, upCount)

	// Drive the drop zone's Alpine drag state: dragenter highlights the border,
	// dragleave clears it. This exercises the @dragenter/@dragleave handlers.
	zone := page.Locator("#fileinput-default [data-fileinput-variant='dropzone']")
	require.NoError(t, zone.ScrollIntoViewIfNeeded())

	require.NoError(t, zone.DispatchEvent("dragenter", nil))
	_, err = page.WaitForFunction(
		"() => document.querySelector('#fileinput-default [data-fileinput-variant=\"dropzone\"]').className.includes('border-primary')",
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	require.NoError(t, err, "dragenter should toggle the primary highlight border")

	require.NoError(t, zone.DispatchEvent("dragleave", nil))
	_, err = page.WaitForFunction(
		`() => {
			const zone = document.querySelector('#fileinput-default [data-fileinput-variant="dropzone"]')
			return zone &&
				Alpine.$data(zone).dragging === false &&
				zone.classList.contains('border-control-outline') &&
				!zone.classList.contains('border-primary')
		}`,
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	require.NoError(t, err, "dragleave should clear drag state and restore the default border")
	dropZoneClass, err := zone.GetAttribute("class")
	require.NoError(t, err)
	require.Contains(t, dropZoneClass, "border-control-outline")
	require.Contains(t, dropZoneClass, "dark:border-control-outline-dark")
	require.NotContains(t, dropZoneClass, "border-primary")

	// Selecting a file on the drop zone input keeps the page error-free and the
	// hidden input retains the file (drop zone has no x-text display).
	require.NoError(t, page.Locator("#demoCoverPicture").SetInputFiles(playwright.InputFile{
		Name:     "cover.png",
		MimeType: "image/png",
		Buffer:   []byte("\x89PNG\r\n"),
	}))
	fileCount, err := page.Locator("#demoCoverPicture").Evaluate("el => el.files.length", nil)
	require.NoError(t, err)
	require.EqualValues(t, 1, fileCount)

	require.Empty(t, jsErrors, "no JS console/page errors on file input demo: %v", jsErrors)
}
