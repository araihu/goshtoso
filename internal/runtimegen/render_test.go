package runtimegen

import (
	"regexp"
	"strings"
	"testing"
)

func TestRenderVendorConstantsIncludesRuntimeHashLookup(t *testing.T) {
	got := renderVendorConstants(fixtureModel(t))
	for _, want := range []string{
		`func RuntimeHash(role RuntimeAssetRole) (string, bool)`,
		`return MuambaHash("alpinejs", "core-js")`,
		`return MuambaHash("htmx", "core-js")`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q\n%s", want, got)
		}
	}
}

func TestRenderPreservesRuntimeContracts(t *testing.T) {
	model := fixtureModel(t)
	vendor := renderVendorConstants(model)
	manifest := renderRuntimeManifest(model)
	attributions := renderRuntimeAttributions(model)
	documentation := renderRuntimeDocumentation(model)

	for _, want := range []struct{ name, value string }{
		{"DependencyLoaderURL", `"/assets/js/dependency-loader.js"`},
		{"AlpineJSCDNURL", `"https://example.test/alpinejs/core-js"`},
		{"AlpineJSIntegrity", `"sha384-fixture"`},
		{"runtimeVersionHTMX", `"2.0.8"`},
	} {
		pattern := regexp.MustCompile(regexp.QuoteMeta(want.name) + `\s*=\s*` + regexp.QuoteMeta(want.value))
		if !pattern.MatchString(vendor) {
			t.Errorf("vendor constants missing %s = %s", want.name, want.value)
		}
	}
	for _, want := range []string{
		`RuntimeRoleDependencyLoader RuntimeAssetRole = "dependency-loader"`,
		`RuntimeRoleFirstParty RuntimeAssetRole = "first-party"`,
		`RuntimeRoleHTMX RuntimeAssetRole = "htmx"`,
		`WaitForWindowLoaded: true`,
		`ProvenanceURL: "https://example.test/htmx/package"`,
	} {
		if !strings.Contains(manifest, want) {
			t.Errorf("runtime manifest missing %q", want)
		}
	}
	if !strings.Contains(attributions, `Version: assets.HTMXVersion`) {
		t.Errorf("attributions missing HTMX version function\n%s", attributions)
	}
	for _, want := range []string{"`assets/runtime.overlay.yaml`", "`muamba.yaml`", "`assets.RuntimeHash(role)`"} {
		if !strings.Contains(documentation, want) {
			t.Errorf("documentation missing %q", want)
		}
	}
	for _, forbidden := range []string{"manifest.json", "versions.json", "cmd/vendorgen"} {
		if strings.Contains(documentation, forbidden) {
			t.Errorf("documentation retains %q", forbidden)
		}
	}
}

func TestRenderLocalAssetIgnoresIntegrityWithoutCDNURL(t *testing.T) {
	model := Model{Loader: Asset{
		Role: "loader", RoleGoName: "Loader", GoName: "Loader",
		Name: "Loader", LocalURL: "/assets/js/loader.js", Integrity: "sha384-unused",
	}}
	manifest := renderRuntimeManifest(model)
	if strings.Contains(manifest, "Integrity:") {
		t.Fatalf("local-only asset references an integrity constant:\n%s", manifest)
	}
}

func fixtureModel(t *testing.T) Model {
	t.Helper()
	model, err := Load(strings.NewReader(validOverlay), fixtureInventory(t))
	if err != nil {
		t.Fatal(err)
	}
	return model
}
