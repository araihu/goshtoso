//go:build e2e && (full || tooltip)

package e2e

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/tooltip"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type persistentTooltipMode struct {
	name  string
	theme string
	dark  bool
}

type persistentTooltipState struct {
	expanded    string
	controls    string
	describedBy string
	hasPopup    string
	role        string
	visible     bool
	focused     bool
}

func TestTooltipPersistentSemanticsAcrossFoundationThemes(t *testing.T) {
	server := httptest.NewServer(persistentTooltipHandler(t))
	t.Cleanup(server.Close)

	var modes []persistentTooltipMode
	for _, theme := range []string{"araihu", "modern", "goshtoso"} {
		for _, dark := range []bool{false, true} {
			modes = append(modes, persistentTooltipMode{
				name:  fmt.Sprintf("%s-dark=%t", theme, dark),
				theme: theme,
				dark:  dark,
			})
		}
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			page := newPage(t, sharedBrowser, playwright.BrowserNewPageOptions{
				HasTouch: playwright.Bool(true),
				Viewport: &playwright.Size{Width: 1280, Height: 900},
			})
			_, err := page.Goto(
				fmt.Sprintf("%s/?theme=%s&dark=%t", server.URL, mode.theme, mode.dark),
				playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateLoad},
			)
			require.NoError(t, err)
			_, err = page.WaitForFunction(`() => {
				return typeof Alpine !== 'undefined' &&
					document.querySelector('#hover-custom-trigger')?.getAttribute('aria-describedby') === 'hover-custom' &&
					document.querySelector('#persistent-custom-trigger')?.getAttribute('aria-describedby') === 'persistent-custom';
			}`, nil)
			require.NoError(t, err)

			assertOrdinaryTooltipSemantics(t, page, "[aria-describedby='hover-default']", "hover-default")
			assertOrdinaryTooltipSemantics(t, page, "#hover-custom-trigger", "hover-custom")

			triggers := []struct {
				name     string
				selector string
				tipID    string
			}{
				{name: "default", selector: "[aria-describedby='persistent-default']", tipID: "persistent-default"},
				{name: "custom-native", selector: "#persistent-custom-trigger", tipID: "persistent-custom"},
				{name: "custom-role", selector: "#persistent-role-trigger", tipID: "persistent-role"},
				{name: "custom-fallback", selector: "[data-tooltip-content-id='persistent-fallback']", tipID: "persistent-fallback"},
			}
			for _, trigger := range triggers {
				t.Run(trigger.name, func(t *testing.T) {
					exercisePersistentTooltip(t, page, trigger.selector, trigger.tipID)
				})
			}
		})
	}
}

func assertOrdinaryTooltipSemantics(t *testing.T, page playwright.Page, triggerSelector, tipID string) {
	t.Helper()
	state := readPersistentTooltipState(t, page, triggerSelector, tipID)
	assert.Contains(t, strings.Fields(state.describedBy), tipID, "ordinary Tooltip description")
	assert.Empty(t, state.expanded, "ordinary Tooltip must not expose persistent expanded state")
	assert.Empty(t, state.controls, "ordinary Tooltip needs only its descriptive relationship")
	assert.Empty(t, state.hasPopup, "role=tooltip is not an aria-haspopup popup")
	assert.Equal(t, "tooltip", state.role)
}

func exercisePersistentTooltip(t *testing.T, page playwright.Page, triggerSelector, tipID string) {
	t.Helper()
	trigger := page.Locator(triggerSelector)
	outside := page.Locator("#outside-target")
	require.NoError(t, trigger.ScrollIntoViewIfNeeded())

	assertPersistentTooltipState(t, page, triggerSelector, tipID, false, boolExpectation(false), "initial")

	require.NoError(t, trigger.Click())
	assertPersistentTooltipState(t, page, triggerSelector, tipID, true, boolExpectation(true), "after click")
	require.NoError(t, outside.Click())
	assertPersistentTooltipState(t, page, triggerSelector, tipID, false, boolExpectation(false), "after outside click")

	require.NoError(t, trigger.Focus())
	assertPersistentTooltipState(t, page, triggerSelector, tipID, false, boolExpectation(true), "before Enter")
	require.NoError(t, trigger.Press("Enter"))
	assertPersistentTooltipState(t, page, triggerSelector, tipID, true, boolExpectation(true), "after Enter")
	require.NoError(t, outside.Click())
	assertPersistentTooltipState(t, page, triggerSelector, tipID, false, boolExpectation(false), "after Enter reset")

	require.NoError(t, trigger.Focus())
	assertPersistentTooltipState(t, page, triggerSelector, tipID, false, boolExpectation(true), "before Space")
	require.NoError(t, trigger.Press("Space"))
	assertPersistentTooltipState(t, page, triggerSelector, tipID, true, boolExpectation(true), "after Space")
	require.NoError(t, outside.Click())
	assertPersistentTooltipState(t, page, triggerSelector, tipID, false, boolExpectation(false), "after Space reset")

	require.NoError(t, trigger.Click())
	require.NoError(t, trigger.Focus())
	assertPersistentTooltipState(t, page, triggerSelector, tipID, true, boolExpectation(true), "before Escape")
	require.NoError(t, trigger.Press("Escape"))
	assertPersistentTooltipState(t, page, triggerSelector, tipID, false, boolExpectation(true), "after Escape")
	if state := readPersistentTooltipState(t, page, triggerSelector, tipID); state.visible {
		require.NoError(t, outside.Click())
	}

	require.NoError(t, outside.Focus())
	assertPersistentTooltipState(t, page, triggerSelector, tipID, false, boolExpectation(false), "before touch tap")
	require.NoError(t, trigger.Tap())
	assertPersistentTooltipState(t, page, triggerSelector, tipID, true, nil, "after touch tap")
	require.NoError(t, outside.Tap())
	assertPersistentTooltipState(t, page, triggerSelector, tipID, false, nil, "after outside touch tap")
}

func boolExpectation(value bool) *bool {
	return &value
}

func assertPersistentTooltipState(t *testing.T, page playwright.Page, triggerSelector, tipID string, expanded bool, focused *bool, action string) {
	t.Helper()
	wantDisplay := "none"
	if expanded {
		wantDisplay = "visible"
	}
	_, waitErr := page.WaitForFunction(`args => {
		const tip = document.getElementById(args.tipID);
		if (!tip) return false;
		const visible = getComputedStyle(tip).display !== 'none';
		return visible === args.expanded;
	}`, map[string]any{"tipID": tipID, "expanded": expanded}, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(800)})
	assert.NoErrorf(t, waitErr, "%s: Tooltip display did not become %s", action, wantDisplay)

	state := readPersistentTooltipState(t, page, triggerSelector, tipID)
	assert.Equalf(t, fmt.Sprint(expanded), state.expanded, "%s: reflected aria-expanded", action)
	assert.Equalf(t, tipID, state.controls, "%s: aria-controls", action)
	assert.Containsf(t, strings.Fields(state.describedBy), tipID, "%s: aria-describedby", action)
	assert.Emptyf(t, state.hasPopup, "%s: tooltip is not an aria-haspopup popup", action)
	assert.Equalf(t, "tooltip", state.role, "%s: content role", action)
	assert.Equalf(t, expanded, state.visible, "%s: visual visibility", action)
	if focused != nil {
		assert.Equalf(t, *focused, state.focused, "%s: trigger focus", action)
	}
}

func readPersistentTooltipState(t *testing.T, page playwright.Page, triggerSelector, tipID string) persistentTooltipState {
	t.Helper()
	value, err := page.Evaluate(`args => {
		const trigger = document.querySelector(args.triggerSelector);
		const tip = document.getElementById(args.tipID);
		return {
			expanded: trigger?.getAttribute('aria-expanded') || '',
			controls: trigger?.getAttribute('aria-controls') || '',
			describedBy: trigger?.getAttribute('aria-describedby') || '',
			hasPopup: trigger?.getAttribute('aria-haspopup') || '',
			role: tip?.getAttribute('role') || '',
			visible: !!tip && getComputedStyle(tip).display !== 'none',
			focused: document.activeElement === trigger,
		};
	}`, map[string]any{"triggerSelector": triggerSelector, "tipID": tipID})
	require.NoError(t, err)
	result := value.(map[string]any)
	return persistentTooltipState{
		expanded:    result["expanded"].(string),
		controls:    result["controls"].(string),
		describedBy: result["describedBy"].(string),
		hasPopup:    result["hasPopup"].(string),
		role:        result["role"].(string),
		visible:     result["visible"].(bool),
		focused:     result["focused"].(bool),
	}
}

func persistentTooltipHandler(t *testing.T) http.Handler {
	t.Helper()
	body := persistentTooltipFixture(t)
	dependencies := renderPersistentTooltip(t, head.Dependencies(head.WithLocalRuntime()))
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", assets.Handler())
	mux.HandleFunc("GET /", func(writer http.ResponseWriter, request *http.Request) {
		darkClass := ""
		if request.URL.Query().Get("dark") == "true" {
			darkClass = " dark"
		}
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(writer, `<!doctype html><html data-theme=%q class=%q><head><meta charset="utf-8">%s</head><body><main class="space-y-8 p-8">%s<button id="outside-target" type="button">Outside</button></main></body></html>`, request.URL.Query().Get("theme"), darkClass, dependencies, body)
	})
	return mux
}

func persistentTooltipFixture(t *testing.T) string {
	t.Helper()
	components := []templ.Component{
		tooltip.Tooltip("hover-default", "Hover default", tooltip.WithTriggerLabel("Hover default")),
		tooltip.Tooltip("hover-custom", "Hover custom", tooltip.WithTrigger(persistentTooltipMarkup(`<button id="hover-custom-trigger" type="button">Hover custom</button>`))),
		tooltip.Tooltip("persistent-default", "Persistent default", tooltip.WithActivation(tooltip.ActivationClick), tooltip.WithTriggerLabel("Persistent default")),
		tooltip.Tooltip("persistent-custom", "Persistent custom", tooltip.WithActivation(tooltip.ActivationClick), tooltip.WithTrigger(persistentTooltipMarkup(`<button id="persistent-custom-trigger" type="button">Persistent custom</button>`))),
		tooltip.Tooltip("persistent-role", "Persistent role", tooltip.WithActivation(tooltip.ActivationClick), tooltip.WithTrigger(persistentTooltipMarkup(`<span id="persistent-role-trigger" role="button" tabindex="0">Persistent role</span>`))),
		tooltip.Tooltip("persistent-fallback", "Persistent fallback", tooltip.WithActivation(tooltip.ActivationClick), tooltip.WithTrigger(persistentTooltipMarkup(`<span>Persistent fallback</span>`))),
	}
	var output strings.Builder
	for _, component := range components {
		output.WriteString(renderPersistentTooltip(t, component))
	}
	return output.String()
}

func persistentTooltipMarkup(markup string) templ.Component {
	return templ.ComponentFunc(func(_ context.Context, writer io.Writer) error {
		_, err := io.WriteString(writer, markup)
		return err
	})
}

func renderPersistentTooltip(t *testing.T, component templ.Component) string {
	t.Helper()
	var output strings.Builder
	require.NoError(t, component.Render(context.Background(), &output))
	return output.String()
}
