package palette

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func render(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Palette(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestPalette_GridAndStaticDispatch(t *testing.T) {
	html := render(t, Config{ID: "palette-surface"})
	want := len(DefaultHues) * len(DefaultShades)
	assert.Equal(t, want, strings.Count(html, "background-color: var(--color-"), "one grid button per hue×shade")
	assert.Contains(t, html, `data-cls="blue-700"`)
	assert.Contains(t, html, `style="background-color: var(--color-blue-700)"`)
	assert.Contains(t, html, `@click="pick($el.dataset.cls, $el)"`)
	assert.Contains(t, html, `data-cls="white"`)
	assert.Contains(t, html, `data-cls="black"`)
	assert.Contains(t, html, `@click="pick('', null)"`)
	assert.Contains(t, html, "Reset")
	assert.NotContains(t, html, `type="color"`)
}

func TestPalette_AlpineModelAndHex(t *testing.T) {
	html := render(t, Config{ID: "p", AlpineModel: "myColor", ShowHex: true})
	assert.Contains(t, html, `x-on:select-close="myColor = $event.detail"`)
	assert.Contains(t, html, `type="color"`)
	assert.Contains(t, html, `data-selected-preview`)
	assert.Contains(t, html, `:value="selectedHex"`)
	assert.Contains(t, html, `@change="commitHex($event.target.value)"`)
	assert.Contains(t, html, `x-model="hexInput"`)
	assert.Contains(t, html, `pattern="^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$"`)
	assert.Contains(t, html, `@input="previewHex(hexInput)"`)
	assert.Contains(t, html, `@keydown.enter.prevent="commitHex(hexInput)"`)
	assert.Contains(t, html, `@blur="commitHex(hexInput)"`)
	assert.Contains(t, html, `:aria-invalid="hexInvalid ? 'true' : 'false'"`)
}

func TestPalette_HideFlags(t *testing.T) {
	html := render(t, Config{ID: "p", HideNeutral: true, HideReset: true})
	assert.NotContains(t, html, "Reset")
	assert.NotContains(t, html, `data-cls="white"`)
}
