//go:build e2e

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestFileInputVariants(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/fileinput", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("DropZoneDefaultStillRenders", func(t *testing.T) {
		dropZone := page.Locator("#fileinput-default [data-fileinput-variant='dropzone']")
		count, err := dropZone.Count()
		require.NoError(t, err)
		require.Equal(t, 1, count)

		classAttr, err := dropZone.GetAttribute("class")
		require.NoError(t, err)
		require.Contains(t, classAttr, "border-dashed")
		require.Contains(t, classAttr, "p-8")

		input := page.Locator("#demoCoverPicture")
		accept, err := input.GetAttribute("accept")
		require.NoError(t, err)
		require.Equal(t, "image/*", accept)

		describedBy, err := input.GetAttribute("aria-describedby")
		require.NoError(t, err)
		require.Equal(t, "demoCoverPicture-helper", describedBy)

		helperText, err := page.Locator("#demoCoverPicture-helper").TextContent()
		require.NoError(t, err)
		require.Equal(t, "PNG, JPG, WebP - Max 5MB", helperText)
	})

	t.Run("UploadVariantUsesTextInputStyle", func(t *testing.T) {
		upload := page.Locator("#fileinput-upload [data-fileinput-variant='upload']")
		count, err := upload.Count()
		require.NoError(t, err)
		require.Equal(t, 1, count)

		classAttr, err := upload.GetAttribute("class")
		require.NoError(t, err)
		require.Contains(t, classAttr, "rounded-radius")
		require.Contains(t, classAttr, "border-outline")
		require.Contains(t, classAttr, "bg-surface-alt")
		require.NotContains(t, classAttr, "border-dashed")
		require.NotContains(t, classAttr, "p-8")

		input := page.Locator("#demoUpload")
		inputClass, err := input.GetAttribute("class")
		require.NoError(t, err)
		require.Contains(t, inputClass, "sr-only")

		inputType, err := input.GetAttribute("type")
		require.NoError(t, err)
		require.Equal(t, "file", inputType)

		name, err := input.GetAttribute("name")
		require.NoError(t, err)
		require.Equal(t, "resume", name)

		accept, err := input.GetAttribute("accept")
		require.NoError(t, err)
		require.Equal(t, ".pdf,.docx", accept)

		labelFor, err := page.Locator("#fileinput-upload [data-fileinput-variant='upload']").GetAttribute("for")
		require.NoError(t, err)
		require.Equal(t, "demoUpload", labelFor)

		helper := page.Locator("#demoUpload-helper")
		helperText, err := helper.TextContent()
		require.NoError(t, err)
		require.Equal(t, "PDF or DOCX - Max 10MB", helperText)

		describedBy, err := input.GetAttribute("aria-describedby")
		require.NoError(t, err)
		require.Equal(t, "demoUpload-helper", describedBy)
	})

	t.Run("UploadVariantShowsSelectedFileName", func(t *testing.T) {
		_, err := page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
		require.NoError(t, err)

		input := page.Locator("#demoUpload")
		require.NoError(t, input.SetInputFiles(playwright.InputFile{
			Name:     "portfolio.pdf",
			MimeType: "application/pdf",
			Buffer:   []byte("%PDF-1.4\n"),
		}))

		_, err = page.WaitForFunction(
			"() => document.querySelector('#fileinput-upload [x-text]')?.textContent === 'portfolio.pdf'",
			nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
		)
		require.NoError(t, err)
	})

	t.Run("DropZoneStatesKeepNativeAttributes", func(t *testing.T) {
		required := page.Locator("#demoDocument")
		requiredAttr, err := required.GetAttribute("required")
		require.NoError(t, err)
		require.Equal(t, "", requiredAttr)

		requiredDropZoneClass, err := page.Locator("#fileinput-required [data-fileinput-variant='dropzone']").GetAttribute("class")
		require.NoError(t, err)
		require.Contains(t, requiredDropZoneClass, "border-dashed")

		disabled := page.Locator("#demoDisabled")
		disabledAttr, err := disabled.GetAttribute("disabled")
		require.NoError(t, err)
		require.Equal(t, "", disabledAttr)

		disabledDropZoneClass, err := page.Locator("#fileinput-disabled [data-fileinput-variant='dropzone']").GetAttribute("class")
		require.NoError(t, err)
		require.Contains(t, disabledDropZoneClass, "opacity-50")
		require.Contains(t, disabledDropZoneClass, "cursor-not-allowed")
	})

	t.Run("NoLabelVariantOmitsVisibleLabel", func(t *testing.T) {
		labels := page.Locator("#fileinput-nolabel span.w-fit")
		count, err := labels.Count()
		require.NoError(t, err)
		require.Equal(t, 0, count)

		input := page.Locator("#demoNoLabel")
		accept, err := input.GetAttribute("accept")
		require.NoError(t, err)
		require.Equal(t, "image/*,.pdf", accept)
	})

}

func TestFileInputFragmentNavNoErrors(t *testing.T) {
	page := newIsolatedPage(t)

	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL + "/getting-started")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	require.NoError(t, page.Locator("a[href='/components/fileinput']").Click())
	_, err = page.WaitForFunction(
		"() => document.querySelectorAll('[data-fileinput-variant=\"dropzone\"]').length >= 1 && document.querySelectorAll('[data-fileinput-variant=\"upload\"]').length === 1",
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	require.NoError(t, err)

	require.NoError(t, page.Locator("#demoUpload").SetInputFiles(playwright.InputFile{
		Name:     "fragment-resume.pdf",
		MimeType: "application/pdf",
		Buffer:   []byte("%PDF-1.4\n"),
	}))
	_, err = page.WaitForFunction(
		"() => document.querySelector('#fileinput-upload [x-text]')?.textContent === 'fragment-resume.pdf'",
		nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(3000)},
	)
	require.NoError(t, err)
	require.Empty(t, jsErrors, "no JS console/page errors on fragment-nav file input page: %v", jsErrors)
}
