package themes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestThemeInventoryClassifiesOnlyOrganizationThemes(t *testing.T) {
	organizationKeys := make([]string, 0, 2)
	genericKeys := make([]string, 0, 14)
	for _, theme := range All() {
		switch theme.Ownership {
		case "organization":
			organizationKeys = append(organizationKeys, theme.Key)
		case "generic":
			genericKeys = append(genericKeys, theme.Key)
		default:
			t.Fatalf("theme %q Ownership = %q; want organization or generic", theme.Key, theme.Ownership)
		}
	}

	require.ElementsMatch(t, []string{"araihu", "goshtoso"}, organizationKeys)
	require.ElementsMatch(t, []string{
		"90s",
		"arctic",
		"christmas",
		"dracula",
		"halloween",
		"high-contrast",
		"industrial",
		"minimal",
		"modern",
		"neo-brutalism",
		"news",
		"pastel",
		"prototype",
		"zombie",
	}, genericKeys)
}
