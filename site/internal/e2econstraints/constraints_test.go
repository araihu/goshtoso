package e2econstraints

import (
	"go/build/constraint"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateSuiteAcceptsSingleMultiAndSupportConstraints(t *testing.T) {
	suite := Suite{Files: []SourceFile{
		{Name: "support.go", Expr: mustConstraint(t, "//go:build e2e")},
		{Name: "button_test.go", Expr: mustConstraint(t, "//go:build e2e && (full || button)"), Tests: []string{"TestButton"}},
		{Name: "composition_test.go", Expr: mustConstraint(t, "//go:build e2e && (full || button || toolbar)"), Tests: []string{"TestComposition"}},
	}}
	manifest := Manifest{
		Identities: []Identity{
			{Name: "button", Files: []string{"button_test.go", "composition_test.go"}, Tests: []string{"TestButton", "TestComposition"}},
			{Name: "toolbar", Files: []string{"composition_test.go"}, Tests: []string{"TestComposition"}},
		},
		FullOnly: Identity{Name: "full_only"},
	}

	require.NoError(t, ValidateSuite(suite, manifest))
	require.Equal(t, []string{"TestButton", "TestComposition"}, suite.Inventory("button").Tests)
	require.Equal(t, []string{"TestComposition"}, suite.Inventory("toolbar").Tests)
}

func TestValidateSuiteAcceptsCurrentSourceOnlyRunnableConstraint(t *testing.T) {
	suite := Suite{Files: []SourceFile{
		{Name: "button_test.go", Expr: mustConstraint(t, "//go:build e2e && (full || button)"), Tests: []string{"TestButton"}},
		{Name: "theme_catalog_current_source_test.go", Expr: mustConstraint(t, "//go:build e2e && full && goshtoso_current_source"), Tests: []string{"TestThemeCatalog"}},
	}}
	manifest := Manifest{
		Identities: []Identity{
			{Name: "button", Files: []string{"button_test.go"}, Tests: []string{"TestButton"}},
		},
		FullOnly: Identity{Name: "full_only"},
	}

	require.NoError(t, ValidateSuite(suite, manifest))
	require.NotContains(t, suite.FullInventory().Tests, "TestThemeCatalog")
	require.Contains(t, suite.CurrentSourceInventory().Tests, "TestThemeCatalog")
}

func TestValidateSuiteRejectsUnsafeCurrentSourceConstraints(t *testing.T) {
	manifest := Manifest{Identities: []Identity{{Name: "button"}}}

	t.Run("leaks into standard full", func(t *testing.T) {
		suite := Suite{Files: []SourceFile{{
			Name:  "test.go",
			Expr:  mustConstraint(t, "//go:build e2e && (full || goshtoso_current_source)"),
			Tests: []string{"TestOne"},
		}}}
		err := ValidateSuite(suite, manifest)
		require.ErrorContains(t, err, "must be excluded from standard full")
	})

	t.Run("dedicated tags do not select test", func(t *testing.T) {
		suite := Suite{Files: []SourceFile{{
			Name:  "test.go",
			Expr:  mustConstraint(t, "//go:build e2e && !full && goshtoso_current_source"),
			Tests: []string{"TestOne"},
		}}}
		err := ValidateSuite(suite, manifest)
		require.ErrorContains(t, err, "missing dedicated tag inclusion")
	})

	t.Run("unknown tag remains rejected", func(t *testing.T) {
		suite := Suite{Files: []SourceFile{{
			Name:  "test.go",
			Expr:  mustConstraint(t, "//go:build e2e && full && goshtoso_current_source && typo"),
			Tests: []string{"TestOne"},
		}}}
		err := ValidateSuite(suite, manifest)
		require.ErrorContains(t, err, `unknown build tag "typo"`)
	})
}

func TestValidateSuiteRejectsUnsafeConstraints(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		want       string
	}{
		{name: "missing e2e", expression: "//go:build full || button", want: "does not require e2e"},
		{name: "missing full", expression: "//go:build e2e && button", want: "missing full fallback"},
		{name: "unknown", expression: "//go:build e2e && (full || typo)", want: "unknown build tag"},
		{name: "runnable support", expression: "//go:build e2e", want: "suite-only support"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			suite := Suite{Files: []SourceFile{{Name: "test.go", Expr: mustConstraint(t, test.expression), Tests: []string{"TestOne"}}}}
			err := ValidateSuite(suite, Manifest{Identities: []Identity{{Name: "button"}}})
			require.ErrorContains(t, err, test.want)
		})
	}
}

func TestLoadManifestRejectsDuplicateIdentities(t *testing.T) {
	path := t.TempDir() + "/identities.json"
	require.NoError(t, os.WriteFile(path, []byte(`{"identities":[{"name":"button"},{"name":"button"}]}`), 0o600))

	_, err := LoadManifest(path)
	require.ErrorContains(t, err, "duplicate identity")
}

func TestConstraintParserRejectsMalformedExpression(t *testing.T) {
	_, err := constraint.Parse("//go:build e2e && (")
	require.Error(t, err)
}

func mustConstraint(t *testing.T, expression string) constraint.Expr {
	t.Helper()
	parsed, err := constraint.Parse(expression)
	require.NoError(t, err)
	return parsed
}
