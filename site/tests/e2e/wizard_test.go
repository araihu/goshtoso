//go:build e2e && (full || example_wizard)

package e2e

import (
	"strconv"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// gotoWizard navigates to a fresh wizard page (step 1) and waits for Alpine.
// It pre-allows storage so the bottom-right cookie banner does not overlap the
// step controls on a fresh browser context (mirrors gotoTodo).
func gotoWizard(t *testing.T, page playwright.Page) {
	t.Helper()
	require.NoError(t, page.AddInitScript(playwright.Script{
		Content: new("try{document.cookie='gt_storage=allowed; Path=/; SameSite=Lax'}catch(e){}"),
	}))
	_, err := page.Goto(baseURL + "/examples/wizard")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
}

// fillAccountStep fills step 1 with the given values (no submit).
func fillAccountStep(t *testing.T, page playwright.Page, name, email, password string) {
	t.Helper()
	require.NoError(t, page.Locator("input[name='name']").Fill(name))
	require.NoError(t, page.Locator("input[name='email']").Fill(email))
	require.NoError(t, page.Locator("input[name='password']").Fill(password))
}

// clickContinue clicks the primary submit in the current step and waits until
// the given step number becomes current (aria-current="step").
func clickContinue(t *testing.T, page playwright.Page, wantStep int) {
	t.Helper()
	clickUntil(t, page,
		page.Locator("#wizard-app button[type='submit']"),
		stepCurrentJS(wantStep))
}

// stepCurrentJS returns a JS predicate that is true when step n is the current step.
func stepCurrentJS(n int) string {
	return "() => { const el = document.querySelector('#wizard-step-" + strconv.Itoa(n) + "'); return !!el && el.getAttribute('aria-current') === 'step'; }"
}

// selectCountry opens the custom Goshtoso select and picks a country by label.
func selectCountry(t *testing.T, page playwright.Page, label string) {
	t.Helper()
	require.NoError(t, page.Locator("#wizard-country-trigger").Click())
	require.NoError(t, page.Locator("li[role='option']", playwright.PageLocatorOptions{HasText: label}).Click())
}

// TestWizard_ForwardPath walks all four steps with valid input and confirms.
func TestWizard_ForwardPath(t *testing.T) {
	page := newIsolatedPage(t)
	gotoWizard(t, page)

	// Step 1 is current on load.
	_, err := page.WaitForFunction(stepCurrentJS(1), nil)
	require.NoError(t, err)

	fillAccountStep(t, page, "Ada Lovelace", "ada@example.com", "hunter2hunter")
	clickContinue(t, page, 2)

	// Step 2: address.
	require.NoError(t, page.Locator("input[name='line1']").Fill("1 Analytical Way"))
	require.NoError(t, page.Locator("input[name='city']").Fill("London"))
	require.NoError(t, page.Locator("input[name='postal']").Fill("EC1A"))
	selectCountry(t, page, "United Kingdom")
	clickContinue(t, page, 3)

	// Step 3: pick the Pro plan.
	require.NoError(t, page.Locator("input[name='plan'][value='pro']").Check())
	clickContinue(t, page, 4)

	// Step 4: review shows entered data.
	body, err := page.Locator("#wizard-app").TextContent()
	require.NoError(t, err)
	require.Contains(t, body, "ada@example.com")
	require.Contains(t, body, "London")

	// Confirm → success state.
	clickUntil(t, page,
		page.Locator("#wizard-app button[type='submit']"),
		"() => Array.from(document.querySelectorAll('#wizard-app *')).some(e => e.textContent && e.textContent.includes('onboarded'))")
}

// TestWizard_ValidationBlocks submits an invalid email on step 1 and asserts the
// error appears, a toast shows, and the wizard does not advance.
func TestWizard_ValidationBlocks(t *testing.T) {
	page := newIsolatedPage(t)
	gotoWizard(t, page)

	fillAccountStep(t, page, "Ada", "not-an-email", "hunter2hunter")
	clickUntil(t, page,
		page.Locator("#wizard-app button[type='submit']"),
		"() => Array.from(document.querySelectorAll('#toast-container *')).some(e => e.textContent && e.textContent.includes('Check your entries'))")

	// Still on step 1.
	stillStep1, err := page.Evaluate(stepCurrentJS(1))
	require.NoError(t, err)
	require.Equal(t, true, stillStep1, "invalid submit must not advance past step 1")

	// Inline email error is visible.
	body, err := page.Locator("#wizard-app").TextContent()
	require.NoError(t, err)
	require.Contains(t, body, "valid email")
}

// TestWizard_BackPreservesData fills step 1, advances, goes Back, and verifies
// the entered name is still present.
func TestWizard_BackPreservesData(t *testing.T) {
	page := newIsolatedPage(t)
	gotoWizard(t, page)

	fillAccountStep(t, page, "Grace Hopper", "grace@example.com", "compilers!")
	clickContinue(t, page, 2)

	// Click Back (secondary button, not the submit).
	clickUntil(t, page,
		page.Locator("#wizard-app button:has-text('Back')"),
		stepCurrentJS(1))

	value, err := page.Locator("input[name='name']").InputValue()
	require.NoError(t, err)
	require.Equal(t, "Grace Hopper", value, "Back must preserve entered data")
}

// TestWizard_SidebarNavNoErrors lands elsewhere, navigates to the wizard via the
// sidebar (htmx fragment swap), and verifies no console/page errors while
// completing a step through that path.
func TestWizard_SidebarNavNoErrors(t *testing.T) {
	page := newIsolatedPage(t)
	dismissCookieBanner(t, page)

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

	require.NoError(t, page.Locator("a[href='/examples/wizard']").First().Click())
	_, err = page.WaitForFunction(stepCurrentJS(1), nil)
	require.NoError(t, err)

	// Complete step 1 through the fragment-loaded page.
	fillAccountStep(t, page, "Ada Lovelace", "ada@example.com", "hunter2hunter")
	clickContinue(t, page, 2)

	require.Empty(t, jsErrors, "no JS console/page errors on fragment-nav wizard page: %v", jsErrors)
}

// TestWizard_SidebarPresent verifies the sidebar exposes the wizard link.
func TestWizard_SidebarPresent(t *testing.T) {
	page := newIsolatedPage(t)
	_, err := page.Goto(baseURL + "/examples/wizard")
	require.NoError(t, err)
	require.NoError(t, page.Locator("text=Examples").First().WaitFor())
	require.NoError(t, page.Locator("a[href='/examples/wizard']").First().WaitFor())
}
