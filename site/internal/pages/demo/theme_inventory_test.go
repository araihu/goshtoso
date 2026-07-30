package demo

import (
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestThemeOptionsMatchCompiledThemeSelectors(t *testing.T) {
	content, err := os.ReadFile("../../../../all-themes.css")
	require.NoError(t, err)

	re := regexp.MustCompile(`\[data-theme=([^\]]+)\]`)
	matches := re.FindAllStringSubmatch(string(content), -1)
	cssThemes := make(map[string]bool, len(matches))
	for _, match := range matches {
		cssThemes[match[1]] = true
	}

	options := getThemeOptions()
	require.Len(t, options, 16)
	require.Equal(t, "araihu", options[0].Value)
	require.Len(t, cssThemes, len(options))
	for _, option := range options {
		require.Truef(t, cssThemes[option.Value], "theme option %q has no CSS selector", option.Value)
	}
	require.False(t, cssThemes["totvs"])
}
