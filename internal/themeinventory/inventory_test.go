package themeinventory

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestThemeInventoryParsesCanonicalMetadataAndBlocks(t *testing.T) {
	source := []byte(`@layer base {
    /* goshtoso-theme: arctic | Arctic | generic */
    [data-theme=arctic] {
        --color-primary: blue;
    }

    /* goshtoso-theme: araihu | Arai Hû | organization */
    [data-theme=araihu] {
        --color-primary: #173b72;
    }
}`)

	themes, err := Parse(source)
	require.NoError(t, err)
	require.Equal(t, []Theme{
		{
			Key:       "arctic",
			Label:     "Arctic",
			Ownership: OwnershipGeneric,
			Block:     "[data-theme=arctic] {\n    --color-primary: blue;\n}",
		},
		{
			Key:       "araihu",
			Label:     "Arai Hû",
			Ownership: OwnershipOrganization,
			Block:     "[data-theme=araihu] {\n    --color-primary: #173b72;\n}",
		},
	}, themes)
}

func TestThemeInventoryReadsAllCanonicalSourceThemes(t *testing.T) {
	source, err := os.ReadFile("../../all-themes.css")
	require.NoError(t, err)

	themes, err := Parse(source)
	require.NoError(t, err)
	require.Len(t, themes, 16)

	organizationKeys := make([]string, 0, 2)
	for _, theme := range themes {
		if theme.Ownership == OwnershipOrganization {
			organizationKeys = append(organizationKeys, theme.Key)
		}
		require.NotEmpty(t, theme.Block)
	}
	require.ElementsMatch(t, []string{"araihu", "goshtoso"}, organizationKeys)
}

func TestThemeInventoryRejectsDuplicateMissingAndInvalidMetadata(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		message string
	}{
		{
			name: "duplicate metadata",
			source: `@layer base {
    /* goshtoso-theme: arctic | Arctic | generic */
    [data-theme=arctic] {}
    /* goshtoso-theme: arctic | Arctic | generic */
    [data-theme=arctic] {}
}`,
			message: `duplicate theme metadata key "arctic"`,
		},
		{
			name: "selector missing metadata",
			source: `@layer base {
    [data-theme=arctic] {}
}`,
			message: `selector "arctic" has no theme metadata`,
		},
		{
			name: "metadata key differs from selector",
			source: `@layer base {
    /* goshtoso-theme: araihu | Arai Hû | organization */
    [data-theme=goshtoso] {}
}`,
			message: `metadata key "araihu" does not match selector "goshtoso"`,
		},
		{
			name: "invalid ownership",
			source: `@layer base {
    /* goshtoso-theme: arctic | Arctic | product */
    [data-theme=arctic] {}
}`,
			message: `theme "arctic" has invalid ownership "product"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse([]byte(test.source))
			require.EqualError(t, err, test.message)
		})
	}
}
