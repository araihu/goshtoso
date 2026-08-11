//go:build e2e && (full || statustext)

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/banner"
	"github.com/araihu/goshtoso/components/dropdown"
	"github.com/araihu/goshtoso/components/form"
	"github.com/araihu/goshtoso/components/navbar"
	"github.com/araihu/goshtoso/components/radio"
	"github.com/araihu/goshtoso/components/schemaform"
	"github.com/araihu/goshtoso/components/toast"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

type statusTextContrastCase struct {
	family    string
	tone      string
	state     string
	selector  string
	lightRole string
	darkRole  string
	fontSize  string
	focus     bool
}

type statusTextContrastMeasurement struct {
	Color          string  `json:"color"`
	Background     string  `json:"background"`
	CascadeClasses string  `json:"cascadeClasses"`
	FontSize       string  `json:"fontSize"`
	Ratio          float64 `json:"ratio"`
}

func TestStatusTextContrastAcrossFoundationThemes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	server := newStatusTextFixtureServer(t)
	page := newPage(t, sharedBrowser)
	require.NoError(t, page.SetViewportSize(800, 1200))
	response, err := page.Goto(server.URL, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateLoad})
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, http.StatusOK, response.Status())
	_, err = page.Evaluate(`() => {
		document.querySelectorAll('[x-cloak]').forEach(element => element.removeAttribute('x-cloak'));
		const style = document.createElement('style');
		style.textContent = '* { transition: none !important; animation: none !important; }';
		document.head.append(style);
	}`, nil)
	require.NoError(t, err)

	cases := []statusTextContrastCase{
		{family: "toast", tone: "info", state: "default", selector: "#status-toast-info h3", lightRole: "text-info-text", darkRole: "dark:text-info-text-dark", fontSize: "14px"},
		{family: "toast", tone: "success", state: "default", selector: "#status-toast-success h3", lightRole: "text-success-text", darkRole: "dark:text-success-text-dark", fontSize: "14px"},
		{family: "toast", tone: "warning", state: "default", selector: "#status-toast-warning h3", lightRole: "text-warning-text", darkRole: "dark:text-warning-text-dark", fontSize: "14px"},
		{family: "toast", tone: "danger", state: "default", selector: "#status-toast-danger h3", lightRole: "text-danger-text", darkRole: "dark:text-danger-text-dark", fontSize: "14px"},
		{family: "banner", tone: "info", state: "tinted", selector: "#status-banner-info p", lightRole: "text-on-surface", darkRole: "dark:text-on-surface-dark", fontSize: "14px"},
		{family: "banner", tone: "success", state: "tinted", selector: "#status-banner-success p", lightRole: "text-on-surface", darkRole: "dark:text-on-surface-dark", fontSize: "14px"},
		{family: "banner", tone: "warning", state: "tinted", selector: "#status-banner-warning p", lightRole: "text-on-surface", darkRole: "dark:text-on-surface-dark", fontSize: "14px"},
		{family: "banner", tone: "danger", state: "tinted", selector: "#status-banner-danger p", lightRole: "text-on-surface", darkRole: "dark:text-on-surface-dark", fontSize: "14px"},
		{family: "radio", tone: "info", state: "tinted", selector: `#status-radio label[for="status-radio-info"] .text-info-text`, lightRole: "text-info-text", darkRole: "dark:text-info-text-dark", fontSize: "12px"},
		{family: "radio", tone: "success", state: "tinted", selector: `#status-radio label[for="status-radio-success"] .text-success-text`, lightRole: "text-success-text", darkRole: "dark:text-success-text-dark", fontSize: "12px"},
		{family: "radio", tone: "warning", state: "tinted", selector: `#status-radio label[for="status-radio-warning"] .text-warning-text`, lightRole: "text-warning-text", darkRole: "dark:text-warning-text-dark", fontSize: "12px"},
		{family: "radio", tone: "danger", state: "tinted", selector: `#status-radio label[for="status-radio-danger"] .text-danger-text`, lightRole: "text-danger-text", darkRole: "dark:text-danger-text-dark", fontSize: "12px"},
		{family: "schema-form", tone: "danger", state: "required", selector: `#status-schema [aria-label="required"]`, lightRole: "text-danger-text", darkRole: "dark:text-danger-text-dark", fontSize: "14px"},
		{family: "schema-form", tone: "danger", state: "invalid", selector: "#status-schema .text-xs.text-danger-text", lightRole: "text-danger-text", darkRole: "dark:text-danger-text-dark", fontSize: "12px"},
		{family: "dropdown", tone: "danger", state: "destructive", selector: "#status-dropdown-delete", lightRole: "text-danger-text", darkRole: "dark:text-danger-text-dark", fontSize: "14px"},
		{family: "dropdown", tone: "danger", state: "focus", selector: "#status-dropdown-delete", lightRole: "text-danger-text", darkRole: "dark:text-danger-text-dark", fontSize: "14px", focus: true},
		{family: "navbar", tone: "danger", state: "destructive", selector: `[data-status-navbar-danger].text-sm`, lightRole: "text-danger-text", darkRole: "dark:text-danger-text-dark", fontSize: "14px"},
		{family: "navbar", tone: "danger", state: "focus", selector: `[data-status-navbar-danger].text-sm`, lightRole: "text-danger-text", darkRole: "dark:text-danger-text-dark", fontSize: "14px", focus: true},
		{family: "form-errors", tone: "danger", state: "invalid-link-focus", selector: `#status-form-errors a[href="#status-field"]`, lightRole: "text-danger-text", darkRole: "dark:text-danger-text-dark", fontSize: "14px", focus: true},
		{family: "form-errors", tone: "danger", state: "invalid-path", selector: "#status-form-errors code", lightRole: "text-danger-text", darkRole: "dark:text-danger-text-dark", fontSize: "12px"},
		{family: "form", tone: "success", state: "default", selector: `#status-form-flip button[x-show="isEditing"]`, lightRole: "text-success-text", darkRole: "dark:text-success-text-dark", fontSize: "14px"},
	}

	for _, theme := range []string{"araihu", "modern", "goshtoso"} {
		for _, dark := range []bool{false, true} {
			setStatusTextTheme(t, page, theme, dark)
			for _, tc := range cases {
				t.Run(fmt.Sprintf("%s/dark=%t/%s/%s/%s", theme, dark, tc.family, tc.tone, tc.state), func(t *testing.T) {
					locator := page.Locator(tc.selector).First()
					require.NoError(t, locator.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateAttached}))
					if tc.focus {
						require.NoError(t, locator.Focus())
						require.True(t, statusTextMatchesFocusVisible(t, locator))
					} else {
						_, err := page.Evaluate(`() => document.activeElement instanceof HTMLElement && document.activeElement.blur()`, nil)
						require.NoError(t, err)
					}

					measurement := measureStatusTextContrast(t, locator)
					require.Contains(t, measurement.CascadeClasses, tc.lightRole)
					require.Contains(t, measurement.CascadeClasses, tc.darkRole)
					require.Equal(t, tc.fontSize, measurement.FontSize)
					require.GreaterOrEqualf(t, measurement.Ratio, 4.5,
						"%s %s %s text must meet WCAG AA on actual background: %+v", tc.family, tc.tone, tc.state, measurement)
					t.Logf("family=%s tone=%s state=%s theme=%s dark=%t role=%s/%s color=%s background=%s font=%s ratio=%.3f",
						tc.family, tc.tone, tc.state, theme, dark, tc.lightRole, tc.darkRole,
						measurement.Color, measurement.Background, measurement.FontSize, measurement.Ratio)
				})
			}
		}
	}
}

func newStatusTextFixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	body := statusTextFixtureBody(t)
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", assets.Handler())
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(w, `<!doctype html><html data-theme="goshtoso"><head><meta charset="utf-8"><link rel="stylesheet" href="%s"></head><body class="bg-surface p-4 text-on-surface dark:bg-surface-dark dark:text-on-surface-dark">%s</body></html>`, assets.StylesURL, body)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func statusTextFixtureBody(t *testing.T) string {
	t.Helper()
	var body strings.Builder
	appendComponent := func(id string, component templ.Component) {
		body.WriteString(`<section id="` + id + `" class="mb-4">`)
		require.NoError(t, component.Render(context.Background(), &body))
		body.WriteString(`</section>`)
	}

	for _, fixture := range []struct {
		name string
		tone toast.Tone
	}{
		{name: "info", tone: toast.ToneInfo},
		{name: "success", tone: toast.ToneSuccess},
		{name: "warning", tone: toast.ToneWarning},
		{name: "danger", tone: toast.ToneDanger},
	} {
		appendComponent("status-toast-"+fixture.name, toast.Toast(toast.Config{
			Tone: fixture.tone, Title: strings.ToUpper(fixture.name), Message: "Status detail", DisplayDuration: -1,
		}))
	}

	for _, fixture := range []struct {
		name string
		tone banner.Tone
	}{
		{name: "info", tone: banner.ToneInfo},
		{name: "success", tone: banner.ToneSuccess},
		{name: "warning", tone: banner.ToneWarning},
		{name: "danger", tone: banner.ToneDanger},
	} {
		appendComponent("status-banner-"+fixture.name, banner.Banner(banner.Config{
			Tone: fixture.tone, Description: strings.ToUpper(fixture.name) + " banner", Persistent: true,
		}))
	}

	appendComponent("status-radio", radio.RadioGroup(radio.GroupConfig{Title: "Status", Items: []radio.Config{
		{ID: "status-radio-info", Name: "status", Value: "info", Label: "Info", HelperText: "Info detail", BadgeColor: "info"},
		{ID: "status-radio-success", Name: "status", Value: "success", Label: "Success", HelperText: "Success detail", BadgeColor: "success"},
		{ID: "status-radio-warning", Name: "status", Value: "warning", Label: "Warning", HelperText: "Warning detail", BadgeColor: "warning"},
		{ID: "status-radio-danger", Name: "status", Value: "danger", Label: "Danger", HelperText: "Danger detail", BadgeColor: "danger"},
	}}))

	appendComponent("status-schema", schemaform.Fields(schemaform.FieldsConfig{Fields: []schemaform.Field{{
		Path: "name", Label: "Name", Kind: schemaform.KindString, Required: true, Errors: []string{"Name is required"},
	}}}))
	appendComponent("status-dropdown", dropdown.Dropdown(dropdown.Config{
		ID: "status-dropdown-menu", Label: "Actions", Sections: []dropdown.Section{{Items: []dropdown.Item{{
			ID: "status-dropdown-delete", Label: "Delete", Href: "#delete", Danger: true,
		}}}},
	}))
	appendComponent("status-navbar", navbar.Navbar(navbar.Config{
		Brand: templ.Raw("<span>Status fixture</span>"),
		User:  &navbar.UserProfile{Name: "Ada", Email: "ada@example.test"},
		UserMenu: []navbar.UserMenuItem{{
			Label: "Sign out", Href: "#sign-out", Danger: true,
			LinkAttrs: templ.Attributes{"data-status-navbar-danger": "true"},
		}},
	}))
	appendComponent("status-form-errors", form.FormErrors(form.FormErrorsConfig{Items: []form.FormErrorItem{
		{Path: "Name", Message: "Name is required", TargetID: "status-field"},
		{Path: "config.token", Message: "Token is invalid"},
	}}))
	appendComponent("status-form-flip", form.FlipSection(form.FlipSectionConfig{
		SectionConfig: form.SectionConfig{ID: "status-flip", Title: "Profile"},
		Flipped:       true,
	}, templ.Raw("<span>Profile summary</span>")))
	return body.String()
}

func setStatusTextTheme(t *testing.T, page playwright.Page, theme string, dark bool) {
	t.Helper()
	_, err := page.Evaluate(`([theme, dark]) => {
		const root = document.documentElement;
		root.dataset.theme = theme;
		root.classList.toggle('dark', dark);
	}`, []any{theme, dark})
	require.NoError(t, err)
}

func statusTextMatchesFocusVisible(t *testing.T, locator playwright.Locator) bool {
	t.Helper()
	value, err := locator.Evaluate(`element => element.matches(':focus-visible')`, nil)
	require.NoError(t, err)
	result, ok := value.(bool)
	require.True(t, ok, "unexpected focus-visible result %T: %v", value, value)
	return result
}

func measureStatusTextContrast(t *testing.T, locator playwright.Locator) statusTextContrastMeasurement {
	t.Helper()
	result, err := locator.Evaluate(`element => {
		const rgba = value => {
			const canvas = document.createElement('canvas');
			canvas.width = 1;
			canvas.height = 1;
			const context = canvas.getContext('2d', { willReadFrequently: true });
			context.clearRect(0, 0, 1, 1);
			context.fillStyle = '#000000';
			context.fillStyle = value;
			context.fillRect(0, 0, 1, 1);
			const [r, g, b, a] = context.getImageData(0, 0, 1, 1).data;
			return [r, g, b, a / 255];
		};
		const composite = (foreground, background) => {
			const alpha = foreground[3] + background[3] * (1 - foreground[3]);
			if (alpha === 0) return [0, 0, 0, 0];
			return [
				(foreground[0] * foreground[3] + background[0] * background[3] * (1 - foreground[3])) / alpha,
				(foreground[1] * foreground[3] + background[1] * background[3] * (1 - foreground[3])) / alpha,
				(foreground[2] * foreground[3] + background[2] * background[3] * (1 - foreground[3])) / alpha,
				alpha,
			];
		};
		const linear = value => {
			value /= 255;
			return value <= 0.04045 ? value / 12.92 : Math.pow((value + 0.055) / 1.055, 2.4);
		};
		const luminance = value => {
			const [r, g, b] = value.slice(0, 3).map(linear);
			return 0.2126 * r + 0.7152 * g + 0.0722 * b;
		};
		const layers = [];
		for (let current = element; current; current = current.parentElement) {
			layers.push(rgba(getComputedStyle(current).backgroundColor));
		}
		const background = layers.reverse().reduce(
			(computed, layer) => composite(layer, computed),
			[255, 255, 255, 1],
		);
		const style = getComputedStyle(element);
		const foreground = composite(rgba(style.color), background);
		const foregroundLuminance = luminance(foreground);
		const backgroundLuminance = luminance(background);
		return {
			color: style.color,
			background: 'rgb(' + background.slice(0, 3).map(Math.round).join(', ') + ')',
			cascadeClasses: Array.from(function * () {
				for (let current = element; current; current = current.parentElement) yield current.className || '';
			}()).join(' '),
			fontSize: style.fontSize,
			ratio: (Math.max(foregroundLuminance, backgroundLuminance) + 0.05) /
				(Math.min(foregroundLuminance, backgroundLuminance) + 0.05),
		};
	}`, nil)
	require.NoError(t, err)
	values, ok := result.(map[string]any)
	require.True(t, ok, "unexpected contrast result %T: %v", result, result)
	ratio, ok := values["ratio"].(float64)
	require.True(t, ok, "unexpected ratio %T: %v", values["ratio"], values["ratio"])
	return statusTextContrastMeasurement{
		Color: values["color"].(string), Background: values["background"].(string),
		CascadeClasses: values["cascadeClasses"].(string), FontSize: values["fontSize"].(string), Ratio: ratio,
	}
}
