package textinput

import (
	"bytes"
	"context"
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
