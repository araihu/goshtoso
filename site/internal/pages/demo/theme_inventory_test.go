package demo

import (
	"os"
	"regexp"
	"testing"

	demothemes "github.com/araihu/goshtoso/site/internal/themes"
	"github.com/stretchr/testify/require"
)

func TestThemeOptionsMatchCompiledThemeSelectors(t *testing.T) {
	content, err := os.ReadFile("../../../../all-themes.css")
	require.NoError(t, err)

	re := regexp.MustCompile(`(?m)^[\t ]*\[data-theme=([a-z0-9][a-z0-9-]*)\][\t ]*\{`)
	matches := re.FindAllStringSubmatch(string(content), -1)
	cssThemes := make(map[string]int, len(matches))
	for _, match := range matches {
		cssThemes[match[1]]++
	}

	options := getThemeOptions()
	catalog := demothemes.All()
	require.Len(t, options, len(catalog))
	require.Len(t, cssThemes, len(catalog))
	for index, option := range options {
		require.Equal(t, catalog[index].Key, option.Value)
		require.Equal(t, catalog[index].Label, option.Label)
		require.Equalf(t, 1, cssThemes[option.Value], "theme option %q must have exactly one canonical CSS selector", option.Value)
	}
	require.Zero(t, cssThemes["totvs"])
}
