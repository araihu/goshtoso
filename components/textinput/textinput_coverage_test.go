package textinput

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errCoverageWriter = errors.New("coverage writer failure")

type failingCoverageWriter struct{}

func (failingCoverageWriter) Write([]byte) (int, error) {
	return 0, errCoverageWriter
}

// TestCoverageRenderDefault exercises the default input branch with the full
// set of optional attributes so each conditional in defaultInput is rendered.
func TestCoverageRenderDefault(t *testing.T) {
	html := renderTextInput(t, Config{
		ID:           "email",
		Name:         "email",
		Label:        "Email",
		Placeholder:  "you@example.com",
		Value:        "seed",
		Type:         TypeEmail,
		Disabled:     true,
		Required:     true,
		Readonly:     true,
		Autocomplete: "email",
		Pattern:      "[a-z]+",
		MaxLength:    32,
		HelperText:   "Helper",
		RootClass:    "root-extra",
		InputAttrs:   templ.Attributes{"data-test": "input"},
	})

	for _, want := range []string{
		`type="email"`,
		`autocomplete="email"`,
		`pattern="[a-z]+"`,
		`maxlength="32"`,
		"disabled",
		"required",
		"readonly",
		`data-test="input"`,
		"root-extra",
		"<label",
		"<small",
		"Helper",
	} {
		assert.Contains(t, html, want, "default input missing %q", want)
	}
}

// TestCoverageMaskInput hits the masked <input> branch of defaultInput.
func TestCoverageMaskInput(t *testing.T) {
	html := renderTextInput(t, Config{
		ID:    "phone",
		Name:  "phone",
		Type:  TypeTel,
		Mask:  "(999) 999-9999",
		Value: "",
	})

	assert.Contains(t, html, "x-data")
	assert.Contains(t, html, `x-mask="(999) 999-9999"`)
	assert.Contains(t, html, `type="tel"`)
}

// TestCoveragePasswordInput exercises the passwordInput template including the
// Alpine toggle and the conditional attributes.
func TestCoveragePasswordInput(t *testing.T) {
	html := renderTextInput(t, Config{
		ID:           "password",
		Name:         "password",
		Label:        "Password",
		Type:         TypePassword,
		Placeholder:  "secret",
		Disabled:     true,
		Required:     true,
		Readonly:     true,
		Autocomplete: "current-password",
		Pattern:      ".{8,}",
		MaxLength:    64,
		HelperText:   "Min 8 chars",
		InputAttrs:   templ.Attributes{"data-test": "pw"},
	})

	for _, want := range []string{
		`x-data="{ showPassword: false }"`,
		`x-bind:type="showPassword ? 'text' : 'password'"`,
		`x-on:click="showPassword = !showPassword"`,
		`aria-label="Show password"`,
		`autocomplete="current-password"`,
		`pattern=".{8,}"`,
		`maxlength="64"`,
		"disabled",
		"required",
		"readonly",
		`data-test="pw"`,
		"Min 8 chars",
	} {
		assert.Contains(t, html, want, "password input missing %q", want)
	}
}

// TestCoverageSearchInput exercises searchInput including its optional id and
// the search-specific class helper.
func TestCoverageSearchInput(t *testing.T) {
	html := renderTextInput(t, Config{
		ID:         "q",
		Name:       "q",
		Type:       TypeSearch,
		Disabled:   true,
		Readonly:   true,
		Pattern:    "[a-z]+",
		MaxLength:  10,
		InputAttrs: templ.Attributes{"hx-get": "/search"},
	})

	for _, want := range []string{
		`type="search"`,
		`aria-label="search"`,
		`id="q"`,
		"pl-10",
		`pattern="[a-z]+"`,
		`maxlength="10"`,
		"disabled",
		"readonly",
		`hx-get="/search"`,
	} {
		assert.Contains(t, html, want, "search input missing %q", want)
	}
}

// TestCoverageSearchInputNoID confirms the id attribute is omitted when ID is
// empty (the false branch of the conditional).
func TestCoverageSearchInputNoID(t *testing.T) {
	html := renderTextInput(t, Config{
		Name: "q",
		Type: TypeSearch,
	})

	assert.NotContains(t, html, "id=")
}

// TestCoverageLabelStateIcons covers the error/success icon branches in
// inputLabel and the state-specific class helpers.
func TestCoverageLabelStateIcons(t *testing.T) {
	errorHTML := renderTextInput(t, Config{
		ID:         "e",
		Label:      "Field",
		State:      StateError,
		HelperText: "bad",
	})
	assert.Contains(t, errorHTML, "text-danger")
	assert.Contains(t, errorHTML, "border-danger")
	assert.Contains(t, errorHTML, "M5.28 4.22") // error icon path

	successHTML := renderTextInput(t, Config{
		ID:         "s",
		Label:      "Field",
		State:      StateSuccess,
		HelperText: "good",
	})
	assert.Contains(t, successHTML, "text-success")
	assert.Contains(t, successHTML, "border-success")
	assert.Contains(t, successHTML, "M12.416 3.376") // success icon path
}

// TestCoverageConfigHelpers exercises the pure config helpers directly so each
// branch is covered without rendering.
func TestCoverageConfigHelpers(t *testing.T) {
	assert.Equal(t, TypeText, Config{}.GetType())
	assert.Equal(t, TypeEmail, Config{Type: TypeEmail}.GetType())

	assert.True(t, Config{Type: TypePassword}.IsPassword())
	assert.False(t, Config{}.IsPassword())
	assert.True(t, Config{Type: TypeSearch}.IsSearch())
	assert.False(t, Config{}.IsSearch())

	assert.True(t, Config{Mask: "x"}.HasMask())
	assert.False(t, Config{}.HasMask())
	assert.True(t, Config{Pattern: "x"}.HasPattern())
	assert.False(t, Config{}.HasPattern())
	assert.True(t, Config{MaxLength: 1}.HasMaxLength())
	assert.False(t, Config{}.HasMaxLength())
	assert.Equal(t, "0", Config{}.MaxLengthStr())
	assert.Equal(t, "42", Config{MaxLength: 42}.MaxLengthStr())

	// Class helpers: default, error, success branches.
	for _, st := range []State{StateDefault, StateError, StateSuccess} {
		cfg := Config{State: st}
		assert.NotEmpty(t, cfg.InputClasses())
		assert.NotEmpty(t, cfg.LabelClasses())
		assert.NotEmpty(t, cfg.HelperTextClasses())
		assert.NotEmpty(t, searchInputClasses(cfg))
	}

	// Container with and without RootClass.
	assert.NotContains(t, Config{}.ContainerClasses(), "extra")
	assert.Contains(t, Config{RootClass: "extra"}.ContainerClasses(), "extra")
}

// TestCoverageDefaultNoLabelNoHelper covers the false branches for label and
// helper text in defaultInput.
func TestCoverageDefaultNoLabelNoHelper(t *testing.T) {
	html := renderTextInput(t, Config{Name: "bare"})
	assert.NotContains(t, html, "<label")
	assert.NotContains(t, html, "<small")
	assert.True(t, strings.Contains(html, "<input"))
}

// TestCoverageDirectTemplatesUseNonBufferWriters covers generated rendering
// branches that are skipped when tests render only through bytes.Buffer.
func TestCoverageDirectTemplatesUseNonBufferWriters(t *testing.T) {
	ctx := templ.WithChildren(context.Background(), templ.Raw("ignored"))

	cases := []struct {
		name      string
		component templ.Component
	}{
		{"textInputDefault", TextInput(Config{Name: "default"})},
		{"textInputPassword", TextInput(Config{Type: TypePassword, Name: "password"})},
		{"textInputSearch", TextInput(Config{Type: TypeSearch, Name: "search"})},
		{"defaultInputBare", defaultInput(Config{Name: "bare"})},
		{"defaultInputMaskedFull", defaultInput(Config{
			ID:           "masked",
			Name:         "masked",
			Type:         TypeTel,
			Mask:         "999-999",
			Disabled:     true,
			Required:     true,
			Readonly:     true,
			Autocomplete: "tel",
			Pattern:      "[0-9-]+",
			MaxLength:    7,
			InputAttrs:   templ.Attributes{"data-mask": "phone"},
		})},
		{"passwordInputBare", passwordInput(Config{Name: "password"})},
		{"searchInputBare", searchInput(Config{Name: "search"})},
		{"inputLabelDefault", inputLabel(Config{ID: "field", Label: "Field"})},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, tc.component.Render(ctx, io.Discard))
		})
	}
}

// TestCoverageTemplatesPropagateCanceledContext covers each generated
// component's early context error return.
func TestCoverageTemplatesPropagateCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cases := []struct {
		name      string
		component templ.Component
	}{
		{"textInput", TextInput(Config{})},
		{"defaultInput", defaultInput(Config{})},
		{"passwordInput", passwordInput(Config{})},
		{"searchInput", searchInput(Config{})},
		{"inputLabel", inputLabel(Config{})},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.component.Render(ctx, io.Discard)
			require.ErrorIs(t, err, context.Canceled)
		})
	}
}

// TestCoverageTemplatesPropagateWriterErrors covers the generated buffer
// release path when rendering to a non-buffer writer fails.
func TestCoverageTemplatesPropagateWriterErrors(t *testing.T) {
	cases := []struct {
		name      string
		component templ.Component
	}{
		{"textInput", TextInput(Config{Name: "default"})},
		{"defaultInput", defaultInput(Config{Name: "default"})},
		{"passwordInput", passwordInput(Config{Name: "password"})},
		{"searchInput", searchInput(Config{Name: "search"})},
		{"inputLabel", inputLabel(Config{ID: "field", Label: "Field"})},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.component.Render(context.Background(), failingCoverageWriter{})
			require.ErrorIs(t, err, errCoverageWriter)
		})
	}
}
