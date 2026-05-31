package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToggle_PageLoads tests that the toggle demo page loads successfully
func TestToggle_PageLoads(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/toggle", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Page_Title_Contains_Toggle", func(t *testing.T) {
		title, err := page.Title()
		require.NoError(t, err)
		assert.Contains(t, title, "Toggle")
		t.Log("✓ Toggle page loads successfully")
	})

	t.Run("Has_Toggle_Heading", func(t *testing.T) {
		heading := page.Locator("h1")
		text, err := heading.TextContent()
		require.NoError(t, err)
		assert.Contains(t, text, "Toggle")
		t.Log("✓ Toggle heading present")
	})
}

// TestToggle_DefaultToggle tests the default toggle renders correctly
func TestToggle_DefaultToggle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/toggle", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Default_Toggle_Has_Correct_Structure", func(t *testing.T) {
		toggle := page.Locator("#demoDefault")
		visible, err := toggle.IsVisible()
		require.NoError(t, err)
		assert.True(t, visible, "default toggle should exist")

		// Check role attribute
		role, err := toggle.GetAttribute("role")
		require.NoError(t, err)
		assert.Equal(t, "switch", role)

		// Check it's checked
		checked, err := toggle.IsChecked()
		require.NoError(t, err)
		assert.True(t, checked, "default toggle should be checked")

		t.Log("✓ Default toggle has correct structure")
	})

	t.Run("Toggle_Has_Label", func(t *testing.T) {
		label := page.Locator("label[for='demoDefault'] span")
		text, err := label.First().TextContent()
		require.NoError(t, err)
		assert.Contains(t, text, "Toggle")
		t.Log("✓ Toggle has label text")
	})

	t.Run("Toggle_Has_Track_Div", func(t *testing.T) {
		track := page.Locator("label[for='demoDefault'] div[aria-hidden='true']")
		count, err := track.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "should have one track div")

		classAttr, err := track.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, classAttr, "rounded-full")
		assert.Contains(t, classAttr, "h-6")
		assert.Contains(t, classAttr, "w-11")
		t.Log("✓ Toggle has track div with correct classes")
	})
}

// TestToggle_ColorVariants tests that all color variants render
func TestToggle_ColorVariants(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/toggle", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	variants := []struct {
		id      string
		label   string
		bgClass string
	}{
		{"demoPrimary", "primary", "peer-checked:bg-primary"},
		{"demoSecondary", "secondary", "peer-checked:bg-secondary"},
		{"demoSuccess", "success", "peer-checked:bg-success"},
		{"demoWarning", "warning", "peer-checked:bg-warning"},
		{"demoDanger", "danger", "peer-checked:bg-danger"},
		{"demoInfo", "info", "peer-checked:bg-info"},
	}

	for _, v := range variants {
		t.Run("Variant_"+v.label, func(t *testing.T) {
			toggle := page.Locator("#" + v.id)
			visible, err := toggle.IsVisible()
			require.NoError(t, err)
			assert.True(t, visible, "%s toggle should be visible", v.label)

			checked, err := toggle.IsChecked()
			require.NoError(t, err)
			assert.True(t, checked, "%s toggle should be checked", v.label)

			// Check track has correct variant class
			track := page.Locator("label[for='" + v.id + "'] div[aria-hidden='true']")
			classAttr, err := track.GetAttribute("class")
			require.NoError(t, err)
			assert.Contains(t, classAttr, v.bgClass, "%s toggle should have %s class", v.label, v.bgClass)

			t.Logf("✓ %s variant renders correctly", v.label)
		})
	}
}

// TestToggle_ContainerStyle tests the container style toggle
func TestToggle_ContainerStyle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/toggle", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Container_Toggle_Has_Border", func(t *testing.T) {
		label := page.Locator("label[for='demoContainer']")
		classAttr, err := label.GetAttribute("class")
		require.NoError(t, err)
		assert.Contains(t, classAttr, "border")
		assert.Contains(t, classAttr, "rounded-radius")
		assert.Contains(t, classAttr, "min-w-52")
		t.Log("✓ Container toggle has border and rounded styling")
	})
}

// TestToggle_DisabledStates tests disabled toggle states
func TestToggle_DisabledStates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/toggle", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Disabled_Off_Toggle", func(t *testing.T) {
		toggle := page.Locator("#demoDisabledOff")
		disabled, err := toggle.IsDisabled()
		require.NoError(t, err)
		assert.True(t, disabled, "should be disabled")

		checked, err := toggle.IsChecked()
		require.NoError(t, err)
		assert.False(t, checked, "disabled off toggle should not be checked")

		t.Log("✓ Disabled off toggle is correctly disabled and unchecked")
	})

	t.Run("Disabled_On_Toggle", func(t *testing.T) {
		toggle := page.Locator("#demoDisabledOn")
		disabled, err := toggle.IsDisabled()
		require.NoError(t, err)
		assert.True(t, disabled, "should be disabled")

		checked, err := toggle.IsChecked()
		require.NoError(t, err)
		assert.True(t, checked, "disabled on toggle should be checked")

		t.Log("✓ Disabled on toggle is correctly disabled and checked")
	})
}

// TestToggle_Accessibility tests accessibility attributes
func TestToggle_Accessibility(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	cleanupServer := setupServer(t)
	defer cleanupServer()

	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	_, err := page.Goto(baseURL+"/components/toggle", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	t.Run("Toggle_Has_Role_Switch", func(t *testing.T) {
		toggles := page.Locator("input[role='switch']")
		count, err := toggles.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "should have at least one toggle with role=switch")
		t.Log("✓ Toggles have role='switch'")
	})

	t.Run("Toggle_Track_Is_AriaHidden", func(t *testing.T) {
		tracks := page.Locator("div[aria-hidden='true']")
		count, err := tracks.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "should have aria-hidden track divs")
		t.Log("✓ Toggle tracks are aria-hidden")
	})

	t.Run("Label_For_Matches_Input_ID", func(t *testing.T) {
		label := page.Locator("label[for='demoDefault']")
		count, err := label.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "should have label with matching for attribute")

		input := page.Locator("#demoDefault")
		inputCount, err := input.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, inputCount, "should have input with matching ID")

		t.Log("✓ Label for attribute matches input ID")
	})

	t.Run("Toggle_Is_Keyboard_Accessible", func(t *testing.T) {
		toggle := page.Locator("#demoDefault")

		// Focus the toggle
		err := toggle.Focus()
		require.NoError(t, err)

		// Check it's initially checked
		checked, err := toggle.IsChecked()
		require.NoError(t, err)
		assert.True(t, checked, "should start checked")

		// Press space to toggle
		err = page.Keyboard().Press("Space")
		require.NoError(t, err)

		// Should now be unchecked
		checked, err = toggle.IsChecked()
		require.NoError(t, err)
		assert.False(t, checked, "should be unchecked after space press")

		t.Log("✓ Toggle responds to keyboard space press")
	})
}

// TestToggleAttrsPassthrough confirms the Attrs passthrough refactor did not
// break rendering: the toggle demo page must contain at least one
// input[role='switch'] after the 4-branch input block was collapsed to a
// single element using checked?= / disabled?= / { cfg.Attrs... }.
func TestToggleAttrsPassthrough(t *testing.T) {
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)

	if _, err := page.Goto(baseURL + "/components/toggle"); err != nil {
		t.Fatalf("goto: %v", err)
	}
	_, err := page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	if err != nil {
		t.Fatalf("wait for Alpine: %v", err)
	}

	count, err := page.Locator("input[role='switch']").Count()
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count == 0 {
		t.Fatal("expected at least one toggle input on the toggle demo page")
	}
}
