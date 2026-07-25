package demo

import (
	"context"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/stretchr/testify/require"
)

func TestComponentGoAPIReferenceUsesExactPackageAndBuildVersion(t *testing.T) {
	reference := componentGoAPIReferenceData("schema-form")

	require.Equal(t, "github.com/araihu/goshtoso/components/schemaform", reference.PackagePath)
	require.Equal(t, "development", reference.Version)
	require.False(t, reference.Versioned)
	require.Empty(t, reference.URL)
}

func TestNewGoAPIReferenceBuildsExactVersionedURL(t *testing.T) {
	entry, ok := catalog.LookupActive("schema-form")
	require.True(t, ok)

	reference := newGoAPIReference(entry, "v0.0.12")

	require.True(t, reference.Versioned)
	require.Equal(t, "v0.0.12", reference.Version)
	require.Equal(
		t,
		"https://pkg.go.dev/github.com/araihu/goshtoso@v0.0.12/components/schemaform",
		reference.URL,
	)
}

func TestGoModuleVersionPatternAcceptsReleaseAndPrereleaseVersions(t *testing.T) {
	for _, version := range []string{"v0.0.12", "v1.2.3-rc.1", "v0.0.0-20260725120000-deadbeefcafe"} {
		require.True(t, goModuleVersionPattern.MatchString(version), version)
	}
	for _, version := range []string{"", "development", "0.0.12", "../v0.0.12"} {
		require.False(t, goModuleVersionPattern.MatchString(version), version)
	}
}

func TestComponentGoAPIReferenceIgnoresNonComponentPages(t *testing.T) {
	require.Empty(t, componentGoAPIReferenceData("theme").PackagePath)
}

func TestComponentGoAPIReferenceDoesNotLinkDevelopmentBuilds(t *testing.T) {
	var rendered strings.Builder
	require.NoError(t, componentGoAPIReference("accordion").Render(context.Background(), &rendered))

	html := rendered.String()
	require.Contains(t, html, `data-go-api-version="development"`)
	require.Contains(t, html, `data-go-api-development`)
	require.NotContains(t, html, `data-go-api-link`)
	require.NotContains(t, html, `href="https://pkg.go.dev/`)
}
