package combobox

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderHTML(t *testing.T, c templ.Component) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, c.Render(context.Background(), &buf))
	return buf.String()
}

func TestInitialState_CopiesStaticAndSelected(t *testing.T) {
	cfg := Config{
		ID:       "i",
		Name:     "i",
		Selected: []string{"tech"},
		Source: Source{Static: []Option{
			{Value: "tech", Label: "Technology"},
			{Value: "fin", Label: "Finance"},
		}},
	}
	st := cfg.InitialState()
	assert.Equal(t, cfg.Source.Static, st.Options, "options seeded from Static")
	assert.Equal(t, []string{"tech"}, st.Selected, "selection copied from cfg")
	assert.Empty(t, st.Search)
	assert.Nil(t, st.Deps)
}

func TestInitialState_EmptyWhenNoStatic(t *testing.T) {
	st := Config{ID: "x", Name: "x"}.InitialState()
	assert.Empty(t, st.Options)
	assert.Empty(t, st.Selected)
}

func TestClientOptionLI_SingleMode_DisabledOption(t *testing.T) {
	// Client single-select with a disabled option: no checkbox, but the
	// disabled styling/ARIA branch of clientOptionLI is exercised.
	cfg := Config{
		ID: "industry", Name: "industry", Mode: ModeSingle,
		Source: Source{Static: []Option{
			{Value: "tech", Label: "Technology"},
			{Value: "legacy", Label: "Legacy", Disabled: true},
		}},
	}
	state := State{Options: cfg.Source.Static, Selected: []string{"tech"}}

	html := renderHTML(t, optionsList(cfg, state))

	// Single mode never renders the checkbox.
	assert.NotContains(t, html, `type="checkbox"`, "single-select has no per-row checkbox")

	// Disabled branch attributes on the legacy row.
	legacyIdx := strings.Index(html, `data-value="legacy"`)
	require.Greater(t, legacyIdx, -1)
	legacyRow := html[legacyIdx:]
	assert.Contains(t, legacyRow, `aria-disabled="true"`)
	assert.Contains(t, legacyRow, `tabindex="-1"`)
	assert.Contains(t, legacyRow, `cursor-not-allowed`)
	assert.Contains(t, legacyRow, `opacity-50`)

	// Client mode: no hx-post on any row.
	assert.NotContains(t, html, `hx-post`)

	// Selected tech row carries the selected label styling.
	techIdx := strings.Index(html, `data-value="tech"`)
	require.Greater(t, techIdx, -1)
	assert.Contains(t, html[techIdx:legacyIdx], `aria-selected="true"`)
	assert.Contains(t, html[techIdx:legacyIdx], `font-semibold`)
}

func TestOptionCheckbox_CheckedReflectsSelection(t *testing.T) {
	// Multi-select client mode renders a checkbox per row; the selected row's
	// checkbox carries the `checked` attribute, the unselected one does not.
	cfg := Config{
		ID: "skills", Name: "skills", Mode: ModeMultiple,
		Source: Source{Static: []Option{
			{Value: "go", Label: "Go"},
			{Value: "rust", Label: "Rust"},
		}},
	}
	state := State{Options: cfg.Source.Static, Selected: []string{"go"}}

	html := renderHTML(t, optionsList(cfg, state))

	goIdx := strings.Index(html, `data-value="go"`)
	rustIdx := strings.Index(html, `data-value="rust"`)
	require.Greater(t, goIdx, -1)
	require.Greater(t, rustIdx, -1)

	goRow := html[goIdx:rustIdx]
	rustRow := html[rustIdx:]

	// The boolean `checked` attribute renders right before tabindex on the input.
	// (The class string also contains "checked:" utilities, so match the attribute.)
	assert.Contains(t, goRow, `type="checkbox"`)
	assert.Contains(t, goRow, `checked tabindex="-1"`, "selected option's checkbox is checked")
	assert.Contains(t, rustRow, `type="checkbox"`)
	assert.NotContains(t, rustRow, `checked tabindex="-1"`, "unselected option's checkbox is not checked")
}

func TestChevronClass_ReflectsSelection(t *testing.T) {
	assert.Equal(t, "", chevronClass(State{}), "no selection: no accent class")
	assert.Equal(t, "", chevronClass(State{Selected: []string{}}))
	got := chevronClass(State{Selected: []string{"a"}})
	assert.Contains(t, got, "text-secondary")
	assert.Contains(t, got, "dark:text-secondary-dark")
}

func TestCombobox_LabelDisabledRootClass(t *testing.T) {
	cfg := Config{
		ID: "industry", Name: "industry", Mode: ModeSingle,
		Label:     "Industry",
		RootClass: "max-w-xs custom-root",
		Disabled:  true,
		Source:    Source{Static: []Option{{Value: "tech", Label: "Technology"}}},
	}
	html := renderHTML(t, Combobox(cfg, State{Options: cfg.Source.Static}))

	// Label branch renders a <label for=...>.
	assert.Contains(t, html, `<label for="industry-trigger"`)
	assert.Contains(t, html, `>Industry</label>`)

	// RootClass appended to the container.
	assert.Contains(t, html, `class="relative max-w-xs custom-root"`)

	// Disabled trigger button.
	triggerIdx := strings.Index(html, `id="industry-trigger"`)
	require.Greater(t, triggerIdx, -1)
	end := strings.Index(html[triggerIdx:], ">") + triggerIdx
	assert.Contains(t, html[triggerIdx:end], `disabled`, "disabled config disables the trigger button")
}

func TestCombobox_NoLabelWhenEmpty(t *testing.T) {
	cfg := Config{
		ID: "industry", Name: "industry", Mode: ModeSingle,
		Source: Source{Static: []Option{{Value: "tech", Label: "Technology"}}},
	}
	html := renderHTML(t, Combobox(cfg, State{Options: cfg.Source.Static}))
	assert.NotContains(t, html, `<label for="industry-trigger"`)
}

func TestTriggerLabelText_DefaultPlaceholder(t *testing.T) {
	// No placeholder, empty selection → "Select…".
	cfg := Config{ID: "x", Name: "x", Mode: ModeSingle,
		Source: Source{Static: []Option{{Value: "a", Label: "A"}}}}
	assert.Equal(t, "Select…", triggerLabelText(cfg, State{Options: cfg.Source.Static}))
}

func TestTriggerLabelText_SingleSelection_FallsBackToValue(t *testing.T) {
	// Single mode, selected value not present in Options → falls back to value.
	cfg := Config{ID: "x", Name: "x", Mode: ModeSingle}
	got := triggerLabelText(cfg, State{Options: []Option{{Value: "a", Label: "A"}}, Selected: []string{"ghost"}})
	assert.Equal(t, "ghost", got)
}

func TestTriggerLabelText_MultiSingleSelection_FallsBackToValue(t *testing.T) {
	// Multi mode with exactly one selection not present in Options → value.
	cfg := Config{ID: "x", Name: "x", Mode: ModeMultiple}
	got := triggerLabelText(cfg, State{Options: []Option{{Value: "a", Label: "A"}}, Selected: []string{"ghost"}})
	assert.Equal(t, "ghost", got)
}

func TestTriggerLabelText_MultiSingleSelection_UsesLabel(t *testing.T) {
	cfg := Config{ID: "x", Name: "x", Mode: ModeMultiple}
	opts := []Option{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}
	got := triggerLabelText(cfg, State{Options: opts, Selected: []string{"b"}})
	assert.Equal(t, "Beta", got)
}

func TestProviderError_RendersRetryAndEscapesID(t *testing.T) {
	cfg := Config{
		ID: "x<script>", Name: "x",
		OptionsEndpoint: "/api/x/options",
	}
	html := renderHTML(t, providerError(cfg))

	assert.Contains(t, html, `Failed to load`)
	assert.Contains(t, html, `hx-get="/api/x/options"`)
	assert.Contains(t, html, `>Retry</button>`)
	assert.Contains(t, html, `role="listbox"`)
	// ID is escaped in HTML output.
	assert.NotContains(t, html, `id="x<script>-options"`)
	assert.Contains(t, html, `&lt;script&gt;`)
}

func TestClientEvent_Constant(t *testing.T) {
	assert.Equal(t, "combobox:change", ClientEvent)
}

func TestHandler_UnsupportedRoute_Returns404(t *testing.T) {
	resetRegistry()
	cfg := Config{
		ID: "u", Name: "u", Mode: ModeMultiple,
		ToggleEndpoint: "/u/toggle", OptionsEndpoint: "/u/options", ClearEndpoint: "/u/clear",
		Source: Source{LazyEndpoint: "/u/options"},
	}
	h := Handler(cfg, staticProvider(nil))

	// POST to /options (only GET is supported there) → falls through to default.
	req := httptest.NewRequest(http.MethodPost, "/u/options", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "unsupported route")
}

func TestHandler_GetToggle_WrongMethod_Returns404(t *testing.T) {
	resetRegistry()
	cfg := Config{
		ID: "u", Name: "u", Mode: ModeMultiple,
		ToggleEndpoint: "/u/toggle", OptionsEndpoint: "/u/options", ClearEndpoint: "/u/clear",
		Source: Source{LazyEndpoint: "/u/options"},
	}
	h := Handler(cfg, staticProvider(nil))

	// GET on the toggle path is not supported (POST only).
	req := httptest.NewRequest(http.MethodGet, "/u/toggle", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_ParseFormError_Returns400(t *testing.T) {
	resetRegistry()
	cfg := Config{
		ID: "u", Name: "u", Mode: ModeMultiple,
		ToggleEndpoint: "/u/toggle", OptionsEndpoint: "/u/options", ClearEndpoint: "/u/clear",
		Source: Source{LazyEndpoint: "/u/options"},
	}
	h := Handler(cfg, staticProvider(nil))

	// Malformed percent-encoding in the query makes ParseForm fail.
	req := httptest.NewRequest(http.MethodGet, "/u/options?%zz=1", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestOptionsList_EmptyState_RendersNoMatches(t *testing.T) {
	cfg := Config{ID: "x", Name: "x", Mode: ModeMultiple,
		Source: Source{Static: []Option{{Value: "a"}}}}
	html := renderHTML(t, optionsList(cfg, State{Options: nil}))
	assert.Contains(t, html, "No matches found")
}
