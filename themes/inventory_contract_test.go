package themes

import (
	"os"
	"regexp"
	"testing"

	"github.com/araihu/goshtoso/internal/themeinventory"
	"github.com/stretchr/testify/require"
)

func TestThemeInventoryCatalogMatchesCSSAuthority(t *testing.T) {
	source, err := os.ReadFile("../all-themes.css")
	require.NoError(t, err)
	sourceThemes, err := themeinventory.Parse(source)
	require.NoError(t, err)

	catalogByKey := make(map[string]Theme, len(builtIn))
	for _, theme := range BuiltIn() {
		key := string(theme.Key)
		require.NotContains(t, catalogByKey, key, "duplicate catalog key")
		catalogByKey[key] = theme
	}
	require.Len(t, catalogByKey, len(sourceThemes))

	for _, sourceTheme := range sourceThemes {
		catalogTheme, ok := catalogByKey[sourceTheme.Key]
		require.True(t, ok, "source theme %q missing from public catalog", sourceTheme.Key)
		require.Equal(t, sourceTheme.Label, catalogTheme.Label, "canonical label drift for %q", sourceTheme.Key)
		require.Equal(t, string(sourceTheme.Ownership), string(catalogTheme.Ownership), "ownership drift for %q", sourceTheme.Key)
	}
}

func TestThemeInventoryCompiledSelectorsMatchCSSAuthority(t *testing.T) {
	source, err := os.ReadFile("../all-themes.css")
	require.NoError(t, err)
	sourceThemes, err := themeinventory.Parse(source)
	require.NoError(t, err)

	want := make(map[string]struct{}, len(sourceThemes))
	for _, theme := range sourceThemes {
		want[theme.Key] = struct{}{}
	}

	for _, path := range []string{"../assets/goshtoso-theme.css", "../assets/styles.css"} {
		t.Run(path, func(t *testing.T) {
			compiled, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, want, selectorKeySet(string(compiled)), "%s selector inventory drift", path)
		})
	}
}

var dataThemeSelector = regexp.MustCompile(`\[data-theme=([a-z0-9][a-z0-9-]*)\]`)

func selectorKeySet(source string) map[string]struct{} {
	keys := make(map[string]struct{})
	for _, match := range dataThemeSelector.FindAllStringSubmatch(source, -1) {
		keys[match[1]] = struct{}{}
	}
	return keys
}
