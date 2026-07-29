package assets

import (
	"crypto/sha512"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
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
		RuntimeRoleAlpineJS,
		RuntimeRoleHTMX,
		RuntimeRoleFirstParty,
		RuntimeRoleCombobox,
		RuntimeRoleActionGroup,
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
	wantMinimal := []bool{false, false, false, true, true, true, true, true}
	wantEnabled := []bool{true, true, true, true, true, true, false, false}
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
	assertRuntimeAsset(t, manifest.Dependencies[3], AlpineJSCDNURL, AlpineJSURL, true, false)
	assertRuntimeAsset(t, manifest.Dependencies[4], HTMXCDNURL, HTMXURL, false, true)
	assertRuntimeAsset(t, manifest.Dependencies[5], FirstPartyBundleURL, FirstPartyBundleURL, true, false)
	assertRuntimeAsset(t, manifest.Dependencies[6], ComboboxURL, ComboboxURL, true, false)
	assertRuntimeAsset(t, manifest.Dependencies[7], ActionGroupURL, ActionGroupURL, true, false)
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
	if len(fresh.Dependencies) != 8 {
		t.Fatalf("fresh dependency count = %d, want 8", len(fresh.Dependencies))
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
	if integrityCount != 5 {
		t.Fatalf("manifest SRI count = %d, want 5 version-matched third-party dependencies", integrityCount)
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
