package vendorgen

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseManifestPreservesOrderAndRequiresRuntimeMetadata(t *testing.T) {
	manifest, err := parseManifest([]byte(testManifestJSON))
	if err != nil {
		t.Fatalf("parseManifest: %v", err)
	}

	if len(manifest.Dependencies) != 2 {
		t.Fatalf("dependency count = %d, want 2", len(manifest.Dependencies))
	}
	if manifest.Dependencies[0].Role != "alpine" || manifest.Dependencies[1].Role != "first-party" {
		t.Fatalf("dependency order = [%s %s]", manifest.Dependencies[0].Role, manifest.Dependencies[1].Role)
	}
	if manifest.Dependencies[0].Homepage == "" || manifest.Dependencies[0].License == "" || manifest.Dependencies[0].Purpose == "" {
		t.Fatalf("third-party attribution metadata missing: %#v", manifest.Dependencies[0])
	}

	invalid := strings.Replace(testManifestJSON, `"license": "MIT"`, `"license": ""`, 1)
	if _, err := parseManifest([]byte(invalid)); err == nil {
		t.Fatal("parseManifest accepted an attributed dependency without a license")
	}
	missingIntegrity := strings.Replace(testManifestJSON, `"integrity": "sha384-ywB1P0WjXou1oD1pmsZQBycsMqsO3tFjGotgWkP/W+2AhgcroefMI1i67KE0yCWn"`, `"integrity": ""`, 1)
	if _, err := parseManifest([]byte(missingIntegrity)); err == nil {
		t.Fatal("parseManifest accepted a vendored dependency without canonical integrity")
	}
	invalidIntegrity := strings.Replace(testManifestJSON, `"integrity": "sha384-ywB1P0WjXou1oD1pmsZQBycsMqsO3tFjGotgWkP/W+2AhgcroefMI1i67KE0yCWn"`, `"integrity": "sha256-example"`, 1)
	if _, err := parseManifest([]byte(invalidIntegrity)); err == nil {
		t.Fatal("parseManifest accepted a non-SHA-384 canonical integrity")
	}
	missingAttribution := strings.Replace(testManifestJSON, `"attribution": true`, `"attribution": false`, 1)
	if _, err := parseManifest([]byte(missingAttribution)); err == nil {
		t.Fatal("parseManifest accepted a third-party dependency without attribution")
	}
	insecureCDN := strings.Replace(testManifestJSON, `https://unpkg.com/alpinejs@{v}/dist/cdn.min.js`, `http://cdn.example/alpinejs@{v}.js`, 1)
	if _, err := parseManifest([]byte(insecureCDN)); err == nil {
		t.Fatal("parseManifest accepted a non-HTTPS CDN URL")
	}
	unsafeLocalPath := strings.Replace(testManifestJSON, `/assets/js/goshtoso.min.js`, `/assets/js/../secret.js`, 1)
	if _, err := parseManifest([]byte(unsafeLocalPath)); err == nil {
		t.Fatal("parseManifest accepted an unsafe embedded local URL")
	}
}

func TestVerifyRemoteDownloadsManifestURLAndRequiresCanonicalIntegrity(t *testing.T) {
	contents := []byte(`version:"3.14.9"`)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write(contents)
	}))
	t.Cleanup(server.Close)

	dependency := dep{
		Version:   "3.14.9",
		URL:       server.URL + "/alpine-{v}.js",
		Integrity: integrityForBytes(contents),
	}
	if err := verifyRemote(map[string]dep{"alpinejs": dependency}, &strings.Builder{}); err != nil {
		t.Fatalf("verifyRemote matching bytes: %v", err)
	}

	dependency.Integrity = integrityForBytes([]byte("different"))
	if err := verifyRemote(map[string]dep{"alpinejs": dependency}, &strings.Builder{}); err == nil {
		t.Fatal("verifyRemote accepted CDN bytes that differ from canonical integrity")
	}
}

func TestVerifyBytesRejectsEmbeddedBytesThatDifferFromManifestIntegrity(t *testing.T) {
	contents := []byte(`version:"3.14.9"`)
	dependency := dep{Version: "3.14.9", Integrity: integrityForBytes([]byte("different"))}
	if err := verifyBytes("alpinejs", dependency, contents); err == nil {
		t.Fatal("verifyBytes accepted embedded bytes that differ from the manifest integrity")
	}
}

func TestVerifyInventoryPathsRejectsUndeclaredOrMissingEmbeddedJavaScript(t *testing.T) {
	manifest, err := parseManifest([]byte(testManifestJSON))
	if err != nil {
		t.Fatal(err)
	}
	declared := []string{
		"assets/js/dependency-loader.js",
		"assets/js/runtime/alpinejs/3.14.9/alpine.min.js",
		"assets/js/goshtoso.min.js",
	}
	if err := verifyInventoryPaths(manifest, declared); err != nil {
		t.Fatalf("matching inventory: %v", err)
	}
	if err := verifyInventoryPaths(manifest, append(declared, "assets/js/runtime/alpinejs/old/alpine.min.js")); err == nil {
		t.Fatal("inventory accepted an undeclared embedded JavaScript file")
	}
	if err := verifyInventoryPaths(manifest, nil); err == nil {
		t.Fatal("inventory accepted a declared JavaScript file that is missing")
	}
}

func TestRenderRuntimeManifestUsesDeclaredOrderAndSemantics(t *testing.T) {
	manifest, err := parseManifest([]byte(testManifestJSON))
	if err != nil {
		t.Fatal(err)
	}
	manifest.Dependencies[0].Integrity = "sha384-ywB1P0WjXou1oD1pmsZQBycsMqsO3tFjGotgWkP/W+2AhgcroefMI1i67KE0yCWn"

	got := renderRuntimeManifest(manifest)
	for _, want := range []string{
		`Role: RuntimeRoleAlpineJS`,
		`Name: "Alpine.js"`,
		`Version: "3.14.9"`,
		`PrimaryURL: AlpineJSCDNURL`,
		`LocalURL: AlpineJSURL`,
		`Integrity: AlpineJSIntegrity`,
		`Enabled: true`,
		`IncludeInMinimal: true`,
		`Role: RuntimeRoleFirstParty`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated runtime manifest missing %q:\n%s", want, got)
		}
	}
	if strings.Index(got, "RuntimeRoleAlpineJS") > strings.Index(got, "RuntimeRoleFirstParty") {
		t.Fatalf("generated runtime order does not match manifest:\n%s", got)
	}
	if strings.Count(got, "RuntimeRoleAlpineJS") < 2 {
		t.Fatalf("generated runtime manifest does not declare its public role constant:\n%s", got)
	}
}

func TestRenderVersionsCompatibilityIncludesOnlyVendoredDependencies(t *testing.T) {
	manifest, err := parseManifest([]byte(testManifestJSON))
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]dep
	if err := json.Unmarshal([]byte(renderVersionsCompatibility(manifest)), &got); err != nil {
		t.Fatalf("generated compatibility manifest is invalid JSON: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("compatibility dependency count = %d, want 1", len(got))
	}
	if got["alpinejs"].Version != "3.14.9" || got["alpinejs"].URL == "" {
		t.Fatalf("alpinejs compatibility entry = %#v", got["alpinejs"])
	}
}

func TestRenderRuntimeAttributionsUsesExactVersionAndEmbeddedPath(t *testing.T) {
	manifest, err := parseManifest([]byte(testManifestJSON))
	if err != nil {
		t.Fatal(err)
	}

	got := renderRuntimeAttributions(manifest)
	for _, want := range []string{
		`Name: "Alpine.js"`,
		`Version: "3.14.9"`,
		`LocalURL: "/assets/js/runtime/alpinejs/3.14.9/alpine.min.js"`,
		`License: "MIT"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated attributions missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Goshtoso runtime") {
		t.Fatalf("first-party non-attribution entry was emitted:\n%s", got)
	}
}

func TestRenderRuntimeDocumentationListsCanonicalManifestAndInspectionAPI(t *testing.T) {
	manifest, err := parseManifest([]byte(testManifestJSON))
	if err != nil {
		t.Fatal(err)
	}

	got := renderRuntimeDocumentation(manifest)
	for _, want := range []string{
		"`assets/js/runtime/manifest.json`",
		"`assets.DefaultRuntimeManifest()`",
		"| 1 | Alpine.js | `3.14.9` |",
		"`/assets/js/runtime/alpinejs/3.14.9/alpine.min.js`",
		"Pinned versions are the tested combination",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated runtime documentation missing %q:\n%s", want, got)
		}
	}
}

func TestURLPath(t *testing.T) {
	got := urlPath("alpinejs", dep{Version: "3.14.9", File: "alpine.min.js"})
	want := "/assets/js/runtime/alpinejs/3.14.9/alpine.min.js"
	if got != want {
		t.Fatalf("urlPath = %q, want %q", got, want)
	}
	if indexOf(got, "/vendor/") >= 0 {
		t.Fatalf("urlPath must avoid /vendor/ because Go module zips omit vendor dirs: %q", got)
	}
}

func TestRenderVendorConstantsDeterministic(t *testing.T) {
	manifest, err := parseManifest([]byte(testManifestJSON))
	if err != nil {
		t.Fatal(err)
	}
	a := renderVendorConstants(manifest)
	b := renderVendorConstants(manifest)
	if a != b {
		t.Fatal("renderVendorConstants not deterministic")
	}
}

func TestRenderVendorConstantsIncludesAllManifestURLsAndCanonicalSRI(t *testing.T) {
	manifest, err := parseManifest([]byte(testManifestJSON))
	if err != nil {
		t.Fatal(err)
	}
	got := renderVendorConstants(manifest)

	for _, want := range []string{
		`DependencyLoaderURL`, `/assets/js/dependency-loader.js`,
		`AlpineJSURL`, `/assets/js/runtime/alpinejs/3.14.9/alpine.min.js`,
		`AlpineJSCDNURL`, `https://unpkg.com/alpinejs@3.14.9/dist/cdn.min.js`,
		`AlpineJSIntegrity`, `sha384-ywB1P0WjXou1oD1pmsZQBycsMqsO3tFjGotgWkP/W+2AhgcroefMI1i67KE0yCWn`,
		`FirstPartyBundleURL`, `/assets/js/goshtoso.min.js`,
	} {
		if indexOf(got, want) < 0 {
			t.Fatalf("generated constants missing %q:\n%s", want, got)
		}
	}
}

func TestIntegrityForBytesUsesSHA384SRIFormat(t *testing.T) {
	got := integrityForBytes([]byte("abc"))
	want := "sha384-ywB1P0WjXou1oD1pmsZQBycsMqsO3tFjGotgWkP/W+2AhgcroefMI1i67KE0yCWn"
	if got != want {
		t.Fatalf("integrityForBytes = %q, want %q", got, want)
	}
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

const testManifestJSON = `{
  "schema": 1,
  "loader": {
    "role": "dependency-loader",
    "go_name": "DependencyLoader",
    "name": "Goshtoso dependency loader",
    "local_url": "/assets/js/dependency-loader.js",
    "purpose": "Loads scripts in order with local fallback",
    "enabled": true,
    "include_in_minimal": true,
    "defer": true
  },
  "dependencies": [
    {
      "module": "alpinejs",
      "role": "alpine",
      "go_name": "AlpineJS",
      "name": "Alpine.js",
      "version": "3.14.9",
      "file": "alpine.min.js",
      "cdn_url": "https://unpkg.com/alpinejs@{v}/dist/cdn.min.js",
	  "integrity": "sha384-ywB1P0WjXou1oD1pmsZQBycsMqsO3tFjGotgWkP/W+2AhgcroefMI1i67KE0yCWn",
      "homepage": "https://alpinejs.dev",
      "license": "MIT",
      "purpose": "Reactive UI",
      "attribution": true,
      "enabled": true,
      "include_in_minimal": true,
      "defer": true
    },
    {
      "role": "first-party",
      "go_name": "FirstPartyBundle",
	  "role_go_name": "FirstParty",
      "name": "Goshtoso runtime",
      "local_url": "/assets/js/goshtoso.min.js",
      "purpose": "Reusable component behavior",
      "enabled": true,
      "include_in_minimal": true,
      "defer": true
    }
  ]
}`
