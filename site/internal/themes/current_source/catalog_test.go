package current_source_test

import (
	"fmt"
	"sort"
	"testing"

	demothemes "github.com/araihu/goshtoso/site/internal/themes"
	corethemes "github.com/araihu/goshtoso/themes"
	"github.com/stretchr/testify/require"
)

func TestCatalogMatchesPublicBuiltInCatalog(t *testing.T) {
	require.NoError(t, compareCatalogs(corethemes.BuiltIn(), demothemes.All()))
}

func TestCatalogAgreementRejectsKeyAndCanonicalLabelDrift(t *testing.T) {
	t.Run("key", func(t *testing.T) {
		drifted := demothemes.All()
		drifted[0].Key = "consumer-only"

		err := compareCatalogs(corethemes.BuiltIn(), drifted)
		require.Error(t, err)
		require.Contains(t, err.Error(), "araihu")
		require.Contains(t, err.Error(), "consumer-only")
	})

	t.Run("canonical label", func(t *testing.T) {
		drifted := demothemes.All()
		drifted[1].Label = "Demo-only Goshtoso"

		err := compareCatalogs(corethemes.BuiltIn(), drifted)
		require.EqualError(t, err, `theme label drift for "goshtoso": demo="Demo-only Goshtoso" canonical="Goshtoso"`)
	})

	t.Run("ownership", func(t *testing.T) {
		drifted := demothemes.All()
		drifted[0].Ownership = demothemes.OwnershipGeneric

		err := compareCatalogs(corethemes.BuiltIn(), drifted)
		require.EqualError(t, err, `theme ownership drift for "araihu": demo="generic" canonical="organization"`)
	})
}

func TestCatalogAgreementNamesSolePresentationOverride(t *testing.T) {
	canonical := catalogByKey(t, corethemes.BuiltIn())
	label, ok := demothemes.PresentationLabelOverride("zombie")
	require.True(t, ok)
	require.Equal(t, demothemes.ZombiePresentationLabelOverride, label)
	require.Equal(t, "Zombie", canonical["zombie"])
	require.NotEqual(t, canonical["zombie"], label)

	for key := range canonical {
		_, overridden := demothemes.PresentationLabelOverride(key)
		require.Equal(t, key == "zombie", overridden, "unexpected presentation override for %q", key)
	}
}

func compareCatalogs(public []corethemes.Theme, demo []demothemes.Theme) error {
	publicByKey := make(map[string]string, len(public))
	publicOwnership := make(map[string]string, len(public))
	for _, theme := range public {
		key := string(theme.Key)
		if _, exists := publicByKey[key]; exists {
			return fmt.Errorf("duplicate public theme key %q", key)
		}
		publicByKey[key] = theme.Label
		publicOwnership[key] = string(theme.Ownership)
	}

	demoByKey := make(map[string]string, len(demo))
	demoOwnership := make(map[string]string, len(demo))
	for _, theme := range demo {
		if _, exists := demoByKey[theme.Key]; exists {
			return fmt.Errorf("duplicate demo theme key %q", theme.Key)
		}
		demoByKey[theme.Key] = theme.Label
		demoOwnership[theme.Key] = string(theme.Ownership)
	}

	missing, extra := make([]string, 0), make([]string, 0)
	for key := range publicByKey {
		if _, exists := demoByKey[key]; !exists {
			missing = append(missing, key)
		}
	}
	for key := range demoByKey {
		if _, exists := publicByKey[key]; !exists {
			extra = append(extra, key)
		}
	}
	if len(missing) != 0 || len(extra) != 0 {
		sort.Strings(missing)
		sort.Strings(extra)
		return fmt.Errorf("theme key drift: missing from demo=%v extra in demo=%v", missing, extra)
	}

	for key, canonical := range publicByKey {
		want := canonical
		if override, ok := demothemes.PresentationLabelOverride(key); ok {
			want = override
		}
		if demoByKey[key] != want {
			return fmt.Errorf("theme label drift for %q: demo=%q canonical=%q", key, demoByKey[key], canonical)
		}
		if demoOwnership[key] != publicOwnership[key] {
			return fmt.Errorf("theme ownership drift for %q: demo=%q canonical=%q", key, demoOwnership[key], publicOwnership[key])
		}
	}
	return nil
}

func catalogByKey(t *testing.T, catalog []corethemes.Theme) map[string]string {
	t.Helper()
	byKey := make(map[string]string, len(catalog))
	for _, theme := range catalog {
		key := string(theme.Key)
		require.NotContains(t, byKey, key)
		byKey[key] = theme.Label
	}
	return byKey
}
