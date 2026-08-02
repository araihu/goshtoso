package e2eimpact

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDerivedTemplAndStylesAreIgnoredWhenAuthoredTemplChanged(t *testing.T) {
	changes := []Change{
		{Status: "M", OldPath: "components/button/button.templ", NewPath: "components/button/button.templ"},
		{Status: "M", OldPath: "components/button/button_templ.go", NewPath: "components/button/button_templ.go"},
		{Status: "M", OldPath: "assets/styles.css", NewPath: "assets/styles.css"},
	}
	graph := packageGraph{packages: map[string]goPackage{
		"github.com/araihu/goshtoso/components/button": {ImportPath: "github.com/araihu/goshtoso/components/button", Dir: "/repo/components/button"},
	}, repoRoot: "/repo", identities: map[string]string{}}
	classified := classifyChanges(changes, graph, identityManifest{})
	require.Empty(t, classified.full)
	require.Equal(t, []string{"github.com/araihu/goshtoso/components/button"}, classified.roots)
}

func TestGeneratedOnlyAndGlobalChangesSelectFull(t *testing.T) {
	for _, path := range []string{
		"assets/styles.css",
		"assets/css/theme.css",
		"muamba.yaml",
		"assets/runtime.overlay.yaml",
		"assets/muamba_gen.go",
		"internal/runtimegen/load.go",
		"assets/runtime_manifest.go",
		"assets/js/src/dependency-loader.js",
	} {
		t.Run(path, func(t *testing.T) {
			classified := classifyChanges([]Change{{Status: "M", OldPath: path, NewPath: path}}, packageGraph{}, identityManifest{})
			require.NotEmpty(t, classified.full)
		})
	}
}

func TestMuambaRuntimeInputsSelectFullAsGlobalRuntime(t *testing.T) {
	for _, path := range []string{
		"muamba.yaml",
		"assets/runtime.overlay.yaml",
		"assets/muamba_gen.go",
		"internal/runtimegen/load.go",
	} {
		t.Run(path, func(t *testing.T) {
			classified := classifyChanges([]Change{{Status: "M", OldPath: path, NewPath: path}}, packageGraph{}, identityManifest{})
			require.Equal(t, []string{"global runtime or theme change " + path}, classified.full)
		})
	}
}

func TestPackageIdentityRecognizesOnlyLeafPagePackages(t *testing.T) {
	identity, ok := packageIdentity("github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/button")
	require.True(t, ok)
	require.Equal(t, "button", identity)
	identity, ok = packageIdentity("github.com/araihu/goshtoso/site/internal/pages/demo/examplepages/chat")
	require.True(t, ok)
	require.Equal(t, "example_chat", identity)
	_, ok = packageIdentity("github.com/araihu/goshtoso/site/internal/pages/demo/componentpages")
	require.False(t, ok)
}
