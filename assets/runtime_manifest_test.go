package assets

import (
	"crypto/sha512"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"reflect"
	"testing"
)

func TestDefaultRuntimeManifestHasCompleteOrderedContract(t *testing.T) {
	manifest := DefaultRuntimeManifest()

	if manifest.Stylesheet != (RuntimeAsset{
		Role:             RuntimeRoleStylesheet,
		Kind:             RuntimeAssetStylesheet,
		PrimaryURL:       StylesURL,
		LocalURL:         StylesURL,
		Enabled:          true,
		IncludeInMinimal: true,
	}) {
		t.Fatalf("stylesheet = %#v", manifest.Stylesheet)
	}
	if manifest.Loader != (RuntimeAsset{
		Role:             RuntimeRoleDependencyLoader,
		Kind:             RuntimeAssetScript,
		PrimaryURL:       DependencyLoaderURL,
		LocalURL:         DependencyLoaderURL,
		Enabled:          true,
		IncludeInMinimal: true,
		Defer:            true,
	}) {
		t.Fatalf("loader = %#v", manifest.Loader)
	}
	wantRoles := []RuntimeAssetRole{
		RuntimeRoleAlpineCollapse,
		RuntimeRoleAlpineFocus,
		RuntimeRoleAlpineMask,
		RuntimeRoleFirstParty,
		RuntimeRoleDarkMode,
		RuntimeRoleAlpineJS,
		RuntimeRoleHTMX,
		RuntimeRoleHTMXExtSSE,
		RuntimeRoleHTMXExtWS,
		RuntimeRoleCombobox,
		RuntimeRoleActionGroup,
		RuntimeRoleCodeBlock,
	}
	gotRoles := make([]RuntimeAssetRole, 0, len(manifest.Dependencies))
	for _, dependency := range manifest.Dependencies {
		gotRoles = append(gotRoles, dependency.Role)
		if dependency.Kind != RuntimeAssetScript {
			t.Errorf("%s kind = %q, want %q", dependency.Role, dependency.Kind, RuntimeAssetScript)
		}
		if dependency.LocalURL == "" {
			t.Errorf("%s local URL is empty", dependency.Role)
		}
	}
	if !reflect.DeepEqual(gotRoles, wantRoles) {
		t.Fatalf("dependency roles = %v, want %v", gotRoles, wantRoles)
	}
	wantMinimal := []bool{false, false, false, true, true, true, true, true, true, true, true, true}
	wantEnabled := []bool{true, true, true, true, false, true, true, false, false, false, false, false}
	for index, dependency := range manifest.Dependencies {
		if dependency.IncludeInMinimal != wantMinimal[index] {
			t.Errorf("%s IncludeInMinimal = %t, want %t", dependency.Role, dependency.IncludeInMinimal, wantMinimal[index])
		}
		if dependency.Enabled != wantEnabled[index] {
			t.Errorf("%s Enabled = %t, want %t", dependency.Role, dependency.Enabled, wantEnabled[index])
		}
	}

	assertRuntimeAsset(t, manifest.Dependencies[0], AlpineCollapseCDNURL, AlpineCollapseURL, true, false)
	assertRuntimeAsset(t, manifest.Dependencies[1], AlpineFocusCDNURL, AlpineFocusURL, true, false)
	assertRuntimeAsset(t, manifest.Dependencies[2], AlpineMaskCDNURL, AlpineMaskURL, true, false)
	assertRuntimeAsset(t, manifest.Dependencies[3], FirstPartyBundleURL, FirstPartyBundleURL, true, false)
	assertRuntimeAsset(t, manifest.Dependencies[4], DarkModeURL, DarkModeURL, true, false)
	assertRuntimeAsset(t, manifest.Dependencies[5], AlpineJSCDNURL, AlpineJSURL, true, false)
	assertRuntimeAsset(t, manifest.Dependencies[6], HTMXCDNURL, HTMXURL, false, true)
	assertRuntimeAsset(t, manifest.Dependencies[7], HTMXExtSSECDNURL, HTMXExtSSEURL, true, false)
	assertRuntimeAsset(t, manifest.Dependencies[8], HTMXExtWSCDNURL, HTMXExtWSURL, true, false)
	assertRuntimeAsset(t, manifest.Dependencies[9], ComboboxURL, ComboboxURL, true, false)
	assertRuntimeAsset(t, manifest.Dependencies[10], ActionGroupURL, ActionGroupURL, true, false)
	assertRuntimeAsset(t, manifest.Dependencies[11], CodeBlockURL, CodeBlockURL, true, false)
}

func TestDefaultRuntimeManifestIsCallerOwned(t *testing.T) {
	mutated := DefaultRuntimeManifest()
	mutated.Stylesheet.LocalURL = "/mutated.css"
	mutated.Loader.PrimaryURL = "/mutated-loader.js"
	mutated.Dependencies[0].LocalURL = "/mutated-collapse.js"
	mutated.Dependencies = append(mutated.Dependencies, RuntimeAsset{Role: "mutated"})

	fresh := DefaultRuntimeManifest()
	if fresh.Stylesheet.LocalURL != StylesURL {
		t.Fatalf("fresh stylesheet URL = %q", fresh.Stylesheet.LocalURL)
	}
	if fresh.Loader.PrimaryURL != DependencyLoaderURL {
		t.Fatalf("fresh loader URL = %q", fresh.Loader.PrimaryURL)
	}
	if fresh.Dependencies[0].LocalURL != AlpineCollapseURL {
		t.Fatalf("fresh collapse URL = %q", fresh.Dependencies[0].LocalURL)
	}
	if len(fresh.Dependencies) != 12 {
		t.Fatalf("fresh dependency count = %d, want 12", len(fresh.Dependencies))
	}
}

func TestDefaultRuntimeMetadataExposesCanonicalDependencyMetadata(t *testing.T) {
	metadata := DefaultRuntimeMetadata()
	byRole := make(map[RuntimeAssetRole]RuntimeAssetMetadata, len(metadata))
	for _, dependency := range metadata {
		byRole[dependency.Role] = dependency
	}

	alpine := byRole[RuntimeRoleAlpineJS]
	if alpine.Name != "Alpine.js" || alpine.Version != AlpineVersion() {
		t.Fatalf("Alpine metadata = (%q, %q)", alpine.Name, alpine.Version)
	}
	if alpine.Homepage != "https://alpinejs.dev" || alpine.License != "MIT" || alpine.Purpose == "" {
		t.Fatalf("Alpine attribution metadata = %#v", alpine)
	}
	if alpine.PackageName != "alpinejs" || alpine.ProvenanceURL != "https://unpkg.com/alpinejs@3.14.9/package.json" {
		t.Fatalf("Alpine provenance metadata = %#v", alpine)
	}
	if alpine.LicenseURL != path.Join(path.Dir(AlpineJSURL), "LICENSE.txt") {
		t.Fatalf("Alpine license URL = %q", alpine.LicenseURL)
	}
	htmx := byRole[RuntimeRoleHTMX]
	if htmx.Name != "htmx" || htmx.Version != HTMXVersion() || htmx.License != "Zero-Clause BSD" {
		t.Fatalf("HTMX metadata = %#v", htmx)
	}

	firstParty := byRole[RuntimeRoleFirstParty]
	if firstParty.Name == "" || firstParty.Version != "" || firstParty.Purpose == "" {
		t.Fatalf("first-party metadata = %#v", firstParty)
	}
}

func TestDefaultRuntimeMetadataIsCallerOwned(t *testing.T) {
	mutated := DefaultRuntimeMetadata()
	mutated[0].Name = "mutated"
	if fresh := DefaultRuntimeMetadata(); fresh[0].Name == "mutated" {
		t.Fatal("DefaultRuntimeMetadata returned shared mutable state")
	}
}

func TestDefaultRuntimeManifestLocalURLsMatchHandlerBytesAndSRI(t *testing.T) {
	server := httptest.NewServer(Handler())
	t.Cleanup(server.Close)

	manifest := DefaultRuntimeManifest()
	declared := append([]RuntimeAsset{manifest.Stylesheet, manifest.Loader}, manifest.Dependencies...)
	integrityCount := 0
	for _, asset := range declared {
		response, err := http.Get(server.URL + asset.LocalURL)
		if err != nil {
			t.Fatalf("GET %s: %v", asset.LocalURL, err)
		}
		body, readErr := io.ReadAll(response.Body)
		closeErr := response.Body.Close()
		if readErr != nil || closeErr != nil {
			t.Fatalf("read %s: %v; close: %v", asset.LocalURL, readErr, closeErr)
		}
		if response.StatusCode != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", asset.LocalURL, response.StatusCode)
		}
		if asset.Integrity == "" {
			continue
		}
		integrityCount++
		sum := sha512.Sum384(body)
		got := "sha384-" + base64.StdEncoding.EncodeToString(sum[:])
		if got != asset.Integrity {
			t.Errorf("%s served SRI = %q, manifest = %q", asset.Role, got, asset.Integrity)
		}
	}
	if integrityCount != 7 {
		t.Fatalf("manifest SRI count = %d, want 7 version-matched third-party dependencies", integrityCount)
	}
}

func TestHandlerDoesNotPublishAuthoredJavaScriptSources(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/assets/js/src/combobox.js", nil)
	recorder := httptest.NewRecorder()
	Handler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("authored source status = %d, want 404", recorder.Code)
	}
}

func assertRuntimeAsset(
	t *testing.T,
	asset RuntimeAsset,
	primaryURL string,
	localURL string,
	deferScript bool,
	waitForWindowLoaded bool,
) {
	t.Helper()
	if asset.PrimaryURL != primaryURL || asset.LocalURL != localURL {
		t.Errorf("%s URLs = (%q, %q), want (%q, %q)", asset.Role, asset.PrimaryURL, asset.LocalURL, primaryURL, localURL)
	}
	if asset.Defer != deferScript {
		t.Errorf("%s defer = %t, want %t", asset.Role, asset.Defer, deferScript)
	}
	if asset.WaitForWindowLoaded != waitForWindowLoaded {
		t.Errorf("%s wait for window = %t, want %t", asset.Role, asset.WaitForWindowLoaded, waitForWindowLoaded)
	}
}
