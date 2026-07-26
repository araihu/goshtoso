package textinput

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderTextInput(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, TextInput(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestTextInput_RenderTargets(t *testing.T) {
	html := renderTextInput(t, Config{
		ID:         "email",
		Name:       "email",
		RootClass:  "root-extra",
		InputAttrs: templ.Attributes{"data-test": "input"},
	})

	assert.Contains(t, html, "root-extra")
	assert.Contains(t, html, `data-test="input"`)
}

func TestTextInputAssociatesHelperAndInvalidStateWithControl(t *testing.T) {
	html := renderTextInput(t, Config{
		ID:         "email",
		Name:       "email",
		State:      StateError,
		HelperText: "Enter a valid email",
		InputAttrs: templ.Attributes{"aria-describedby": "email-policy"},
	})

	assert.Contains(t, html, `id="email-helper"`)
	assert.Contains(t, html, `aria-describedby="email-policy email-helper"`)
	assert.Contains(t, html, `aria-invalid="true"`)
}

func TestTextInputEscapesUserControlledText(t *testing.T) {
	payload := `<img src=x onerror=alert(1)>`
	html := renderTextInput(t, Config{
		ID:          "name",
		Name:        "name",
		Label:       payload,
		Placeholder: payload,
		Value:       payload,
		HelperText:  payload,
	})

	if strings.Contains(html, payload) {
		t.Fatalf("rendered raw payload:\n%s", html)
	}
	for _, want := range []string{
		`&lt;img src=x onerror=alert(1)&gt;`,
		`value="&lt;img src=x onerror=alert(1)&gt;"`,
		`placeholder="&lt;img src=x onerror=alert(1)&gt;"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered HTML missing escaped payload %q:\n%s", want, html)
		}
	}
}
