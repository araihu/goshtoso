package palette

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPalette_CustomHuesShades exercises the c.Hues / c.Shades override
// branches in hues()/shades() and the swatchGrid loops over custom sets.
func TestPalette_CustomHuesShades(t *testing.T) {
	html := render(t, Config{
		ID:     "p",
		Hues:   []string{"red", "blue"},
		Shades: []string{"100", "500"},
	})
	// 2 hues × 2 shades = 4 grid swatches, plus white/black neutrals.
	assert.Equal(t, 4, strings.Count(html, `style="background-color: var(--color-`))
	assert.Contains(t, html, `data-cls="red-100"`)
	assert.Contains(t, html, `data-cls="red-500"`)
	assert.Contains(t, html, `data-cls="blue-100"`)
	assert.Contains(t, html, `data-cls="blue-500"`)
	// Default sets must NOT leak when overridden.
	assert.NotContains(t, html, `data-cls="green-700"`)
	assert.NotContains(t, html, `data-cls="red-950"`)
}

// TestConfig_HuesShadesDefaults covers the default-return branch directly.
func TestConfig_HuesShadesDefaults(t *testing.T) {
	c := Config{}
	assert.Equal(t, DefaultHues, c.hues())
	assert.Equal(t, DefaultShades, c.shades())

	c2 := Config{Hues: []string{"teal"}, Shades: []string{"300"}}
	assert.Equal(t, []string{"teal"}, c2.hues())
	assert.Equal(t, []string{"300"}, c2.shades())
}

// TestConfig_ContainerClasses covers both the plain and RootClass-appended
// branches of ContainerClasses().
func TestConfig_ContainerClasses(t *testing.T) {
	assert.Equal(t, "p-2 space-y-2", Config{}.ContainerClasses())
	assert.Equal(t, "p-2 space-y-2 my-extra", Config{RootClass: "my-extra"}.ContainerClasses())
}

// TestPalette_RootClassRendered confirms RootClass reaches the wrapper class.
func TestPalette_RootClassRendered(t *testing.T) {
	html := render(t, Config{ID: "p", RootClass: "w-64 shadow"})
	assert.Contains(t, html, `class="p-2 space-y-2 w-64 shadow"`)
}

// TestPalette_LazyWhen keeps the full swatch grid out of the initial HTML so
// pages with many palettes do not ship thousands of inert template buttons.
func TestPalette_LazyWhen(t *testing.T) {
	html := render(t, Config{ID: "p", LazyWhen: "open"})
	assert.Contains(t, html, `<template x-if="open">`)
	assert.Contains(t, html, `x-html="swatchGridHTML()"`)
	assert.Contains(t, html, `@click="handleSwatchEvent($event, 'pick')"`)
	assert.NotContains(t, html, `data-cls="blue-700"`)
	assert.NotContains(t, html, `background-color: var(--color-blue-700)`)

	// Without LazyWhen, no x-if template wraps the grid.
	plain := render(t, Config{ID: "p"})
	assert.NotContains(t, plain, "<template x-if")
}

// TestConfig_ModelAssignExpr covers both branches of modelAssignExpr().
func TestConfig_ModelAssignExpr(t *testing.T) {
	assert.Equal(t, "", Config{}.modelAssignExpr())
	assert.Equal(t, "", Config{Alpine: &AlpineConfig{}}.modelAssignExpr())
	assert.Equal(t, "c = $event.detail", Config{Alpine: &AlpineConfig{Model: "c"}}.modelAssignExpr())
}

// TestPalette_NoModelNoListener confirms the x-on:select-close attribute is
// absent when no Alpine model is configured.
func TestPalette_NoModelNoListener(t *testing.T) {
	html := render(t, Config{ID: "p"})
	assert.NotContains(t, html, "x-on:select-close")
}

// TestPalette_HideNeutralKeepsGrid covers HideNeutral while keeping the hue
// grid, ensuring only the white/black quick swatches drop.
func TestPalette_HideNeutralKeepsGrid(t *testing.T) {
	html := render(t, Config{ID: "p", HideNeutral: true})
	assert.NotContains(t, html, `data-cls="white"`)
	assert.NotContains(t, html, `data-cls="black"`)
	assert.Contains(t, html, `data-cls="blue-700"`)
	assert.Contains(t, html, "Reset") // reset still present
}
