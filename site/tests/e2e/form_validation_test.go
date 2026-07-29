package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const formValidationPath = "/form"

func navigateToFormValidation(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.Goto(baseURL+formValidationPath, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)
}

func TestFormValidation_MergedIntoFormPage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	navigateToFormValidation(t, page)

	require.NoError(t, page.Locator("#form-fragment #demo-validation").WaitFor())
	count, err := page.Locator("aside a", playwright.PageLocatorOptions{
		HasText: "Form Validation",
	}).Count()
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Form Validation should not be a separate sidebar component")
}

// fillAndTriggerValidation sets a field value and triggers HTMX field-level
// validation via htmx.ajax(). Uses htmx.ajax() directly because the native
// change event -> hx-trigger pipeline produces empty XHR responses in
// headless Chromium (the outerHTML swap receives 0 bytes despite a 200 status).
func fillAndTriggerValidation(t *testing.T, page playwright.Page, fieldName, value string) {
	t.Helper()

	// Collect all current form values and override the target field
	fieldID := "goshtoso-field-" + fieldName
	js := fmt.Sprintf(`() => new Promise(resolve => {
		const form = document.querySelector('#demo-validation');
		const fd = new FormData(form);
		const vals = {};
		for (const [k, v] of fd.entries()) { vals[k] = v; }
		vals[%q] = %q;
		vals['X-Goshtoso-Validation'] = 'field';

		// Also set the input value in the DOM so the form state is consistent
		const input = document.querySelector('input[name=%q]');
		if (input) input.value = %q;

		const targetID = %q;
		const onAfterSettle = event => {
			if (event.detail?.target?.id === targetID) {
				document.body.removeEventListener('htmx:afterSettle', onAfterSettle);
				resolve();
			}
		};
		document.body.addEventListener('htmx:afterSettle', onAfterSettle);
		const el = document.getElementById(targetID);
		htmx.ajax('POST', '/api/components/form-validation', {
			source: el,
			target: el,
			swap: 'outerHTML',
			values: vals,
			headers: {'HX-Trigger-Name': %q}
		});
	})`, fieldName, value, fieldName, value, fieldID, fieldName)

	_, err := page.Evaluate(js)
	require.NoError(t, err)

}

// fillWithoutValidation sets input value directly without triggering events.
func fillWithoutValidation(t *testing.T, page playwright.Page, fieldName, value string) {
	t.Helper()
	input := page.Locator("input[name='" + fieldName + "']")
	_, err := input.Evaluate("(el, val) => { el.value = val; }", value)
	require.NoError(t, err)
}

func TestFormValidation_SubmitEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	navigateToFormValidation(t, page)

	// Bypass native constraint validation so this test exercises the server's
	// invalid-submit response rather than the browser's required-field popover.
	_, err := page.Locator("#demo-validation").Evaluate("form => { form.noValidate = true; }", nil)
	require.NoError(t, err)

	// Submit without filling any fields.
	require.NoError(t, page.Locator("#demo-validation button[type='submit']").Click())
	require.NoError(t, page.Locator("#goshtoso-field-name [id$='-errors'] .text-danger-text").First().WaitFor())

	// Check for error messages on all 3 required fields
	nameErrors := page.Locator("#goshtoso-field-name [id$='-errors'] .text-danger-text")
	nameErrCount, err := nameErrors.Count()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, nameErrCount, 1, "name field should have error messages")

	slugErrors := page.Locator("#goshtoso-field-slug [id$='-errors'] .text-danger-text")
	slugErrCount, err := slugErrors.Count()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, slugErrCount, 1, "slug field should have error messages")

	emailErrors := page.Locator("#goshtoso-field-email [id$='-errors'] .text-danger-text")
	emailErrCount, err := emailErrors.Count()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, emailErrCount, 1, "email field should have error messages")

	for _, fieldName := range []string{"name", "slug", "email"} {
		control := page.Locator("input[name='" + fieldName + "']")
		assert.Equal(t, "true", mustAttribute(t, control, "aria-invalid"))
		assert.Contains(t, mustAttribute(t, control, "aria-describedby"), "-errors")
	}
}

func TestFormValidation_SubmitValid(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	navigateToFormValidation(t, page)

	// Fill all fields without triggering field-level validation
	fillWithoutValidation(t, page, "name", "My Project")
	fillWithoutValidation(t, page, "slug", "my-project")
	fillWithoutValidation(t, page, "email", "test@example.com")

	// Submit the form
	require.NoError(t, page.Locator("#demo-validation button[type='submit']").Click())

	// Verify success message
	successMsg := page.Locator("#form-result")
	require.NoError(t, successMsg.WaitFor())
	text, err := successMsg.InnerText()
	require.NoError(t, err)
	assert.Contains(t, text, "Form submitted successfully!")
}

func TestFormValidation_FieldChange_NameTooShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	navigateToFormValidation(t, page)

	fillAndTriggerValidation(t, page, "name", "ab")

	// Check name input has error border
	nameInput := page.Locator("input[name='name']")
	classes, err := nameInput.GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, classes, "border-danger", "name input should have border-danger class")

	// Check error text
	nameField := page.Locator("#goshtoso-field-name")
	text, err := nameField.InnerText()
	require.NoError(t, err)
	assert.Contains(t, strings.ToLower(text), "at least 3 characters")
}

func TestFormValidation_FieldChange_NameValid(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	navigateToFormValidation(t, page)

	fillAndTriggerValidation(t, page, "name", "My Project")

	// Check name input has success border
	nameInput := page.Locator("input[name='name']")
	classes, err := nameInput.GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, classes, "border-success", "name input should have border-success class")

	// Check no error text in name field
	nameErrors := page.Locator("#goshtoso-field-name [id$='-errors'] .text-danger-text")
	count, err := nameErrors.Count()
	require.NoError(t, err)
	assert.Equal(t, 0, count, "name field should have no error messages")
}

func TestFormValidation_Dependency_SlugAutoUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	navigateToFormValidation(t, page)

	fillAndTriggerValidation(t, page, "name", "My Project")

	// Verify the slug field's input value was auto-populated via OOB swap
	slugInput := page.Locator("input[name='slug']")
	slugVal, err := slugInput.InputValue()
	require.NoError(t, err)
	assert.Equal(t, "my-project", slugVal, "slug should be auto-generated from name")
}

func TestFormValidation_SlugTaken(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	navigateToFormValidation(t, page)

	fillAndTriggerValidation(t, page, "slug", "admin")

	// Check error text
	slugField := page.Locator("#goshtoso-field-slug")
	text, err := slugField.InnerText()
	require.NoError(t, err)
	assert.Contains(t, strings.ToLower(text), "already taken")
}

func TestFormValidation_EmailInvalid(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	navigateToFormValidation(t, page)

	fillAndTriggerValidation(t, page, "email", "notanemail")

	// Check error text
	emailField := page.Locator("#goshtoso-field-email")
	text, err := emailField.InnerText()
	require.NoError(t, err)
	assert.Contains(t, strings.ToLower(text), "valid email")
}

func TestFormValidation_ValuePreservation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	navigateToFormValidation(t, page)

	// Fill name and trigger validation
	fillAndTriggerValidation(t, page, "name", "My Project")

	// Fill email and trigger validation
	fillAndTriggerValidation(t, page, "email", "test@example.com")

	// Verify name field still has its value
	nameInput := page.Locator("input[name='name']")
	nameVal, err := nameInput.InputValue()
	require.NoError(t, err)
	assert.Equal(t, "My Project", nameVal, "name field value should be preserved after email validation")
}

func TestFormValidation_ErrorClearing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	page := newPage(t, sharedBrowser)
	navigateToFormValidation(t, page)

	// Type too-short name to trigger error
	fillAndTriggerValidation(t, page, "name", "ab")

	// Verify error is present
	nameField := page.Locator("#goshtoso-field-name")
	text, err := nameField.InnerText()
	require.NoError(t, err)
	assert.Contains(t, strings.ToLower(text), "at least 3 characters")

	// Type valid name to clear error
	fillAndTriggerValidation(t, page, "name", "Good Name")

	// Verify error is gone and success state shows
	nameInput := page.Locator("input[name='name']")
	classes, err := nameInput.GetAttribute("class")
	require.NoError(t, err)
	assert.Contains(t, classes, "border-success", "name input should have border-success after correction")
	assert.NotContains(t, classes, "border-danger", "name input should not have border-danger after correction")

	// No error messages
	nameErrors := page.Locator("#goshtoso-field-name [id$='-errors'] .text-danger-text")
	count, err := nameErrors.Count()
	require.NoError(t, err)
	assert.Equal(t, 0, count, "name field should have no error messages after correction")
}
