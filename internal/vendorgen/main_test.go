package vendorgen

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	missingVersionToken := strings.Replace(testManifestJSON, `alpinejs@{v}/dist`, `alpinejs@3.14.9/dist`, 1)
	if _, err := parseManifest([]byte(missingVersionToken)); err == nil {
		t.Fatal("parseManifest accepted a CDN URL without exactly one {v} token")
	}
	duplicateVersionToken := strings.Replace(testManifestJSON, `alpinejs@{v}/dist`, `alpinejs@{v}/{v}/dist`, 1)
	if _, err := parseManifest([]byte(duplicateVersionToken)); err == nil {
		t.Fatal("parseManifest accepted a CDN URL with multiple {v} tokens")
	}
	missingLicenseFile := strings.Replace(testManifestJSON, `"license_file": "LICENSE.txt"`, `"license_file": ""`, 1)
	if _, err := parseManifest([]byte(missingLicenseFile)); err == nil {
		t.Fatal("parseManifest accepted a dependency without a bundled license file")
	}
	licenseOverwritesRuntime := strings.Replace(testManifestJSON, `"license_file": "LICENSE.txt"`, `"license_file": "alpine.min.js"`, 1)
	if _, err := parseManifest([]byte(licenseOverwritesRuntime)); err == nil {
		t.Fatal("parseManifest accepted license_file equal to the JavaScript file")
	}
	loaderWithVendoredMetadata := strings.Replace(testManifestJSON, `"name": "Goshtoso dependency loader",`, `"name": "Goshtoso dependency loader", "package_name": "ignored",`, 1)
	if _, err := parseManifest([]byte(loaderWithVendoredMetadata)); err == nil {
		t.Fatal("parseManifest accepted vendored-only metadata on the loader")
	}
	firstPartyWithVendoredMetadata := strings.Replace(testManifestJSON, `"name": "Goshtoso runtime",`, `"name": "Goshtoso runtime", "license_file": "ignored.txt",`, 1)
	if _, err := parseManifest([]byte(firstPartyWithVendoredMetadata)); err == nil {
		t.Fatal("parseManifest accepted vendored-only metadata on a first-party dependency")
	}
	unsafeLocalPath := strings.Replace(testManifestJSON, `/assets/js/goshtoso.min.js`, `/assets/js/../secret.js`, 1)
	if _, err := parseManifest([]byte(unsafeLocalPath)); err == nil {
		t.Fatal("parseManifest accepted an unsafe embedded local URL")
	}
}

func TestParseManifestRejectsDuplicateVendoredModule(t *testing.T) {
	manifest, err := parseManifest([]byte(testManifestJSON))
	if err != nil {
		t.Fatal(err)
	}
	duplicate := manifest.Dependencies[0]
	duplicate.Role = "alpine-copy"
	duplicate.GoName = "AlpineCopy"
	duplicate.RoleGoName = "AlpineCopy"
	manifest.Dependencies = append(manifest.Dependencies, duplicate)
	contents, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := parseManifest(contents); err == nil || !strings.Contains(err.Error(), "duplicate module") {
		t.Fatalf("parseManifest duplicate module error = %v", err)
	}
}

func TestVerifyRemoteDownloadsManifestURLAndRequiresCanonicalIntegrity(t *testing.T) {
	contents := []byte(`version:"3.14.9"`)
	license := []byte("MIT License")
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/package.json":
			_, _ = writer.Write([]byte(`{"name":"alpinejs","version":"3.14.9"}`))
		case "/LICENSE":
			_, _ = writer.Write(license)
		default:
			_, _ = writer.Write(contents)
		}
	}))
	t.Cleanup(server.Close)

	dependency := dep{
		Version:          "3.14.9",
		URL:              server.URL + "/alpine-{v}.js",
		Integrity:        integrityForBytes(contents),
		PackageName:      "alpinejs",
		ProvenanceURL:    server.URL + "/package.json",
		LicenseFile:      "LICENSE.txt",
		LicenseURL:       server.URL + "/LICENSE",
		LicenseIntegrity: integrityForBytes(license),
	}
	dependencies := []vendoredDependency{{Module: "alpinejs", Dependency: dependency}}
	if err := verifyRemote(dependencies, &strings.Builder{}); err != nil {
		t.Fatalf("verifyRemote matching bytes: %v", err)
	}

	dependency.Integrity = integrityForBytes([]byte("different"))
	dependencies[0].Dependency = dependency
	if err := verifyRemote(dependencies, &strings.Builder{}); err == nil {
		t.Fatal("verifyRemote accepted CDN bytes that differ from canonical integrity")
	}
	dependency.Integrity = integrityForBytes(contents)
	dependency.PackageName = "wrong-package"
	dependencies[0].Dependency = dependency
	if err := verifyRemote(dependencies, &strings.Builder{}); err == nil {
		t.Fatal("verifyRemote accepted provenance for another package")
	}
}

func TestVerifyBytesRejectsEmbeddedBytesThatDifferFromManifestIntegrity(t *testing.T) {
	contents := []byte(`version:"3.14.9"`)
	dependency := dep{Version: "3.14.9", Integrity: integrityForBytes([]byte("different"))}
	if err := verifyBytes("alpinejs", dependency, contents); err == nil {
		t.Fatal("verifyBytes accepted embedded bytes that differ from the manifest integrity")
	}
}

func TestDownloadAllDoesNotWriteWhenALaterFetchFails(t *testing.T) {
	working := t.TempDir()
	oldWorking, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(working); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWorking) })
	old := []byte("old")
	target := filepath.Join("assets", "js", "runtime", "first", "1", "first.js")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, old, 0o644); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.Contains(request.URL.Path, "second") {
			http.Error(writer, "late failure", http.StatusBadGateway)
			return
		}
		_, _ = writer.Write([]byte("new"))
	}))
	t.Cleanup(server.Close)
	good := dep{Version: "1", File: "first.js", URL: server.URL + "/first-{v}.js", Integrity: integrityForBytes([]byte("new")), LicenseFile: "LICENSE.txt", LicenseURL: server.URL + "/first-{v}-license", LicenseIntegrity: integrityForBytes([]byte("new"))}
	bad := good
	bad.File, bad.URL, bad.LicenseURL = "second.js", server.URL+"/second-{v}.js", server.URL+"/second-{v}-license"
	if err := downloadAll([]vendoredDependency{{Module: "first", Dependency: good}, {Module: "second", Dependency: bad}}, &strings.Builder{}); err == nil {
		t.Fatal("downloadAll accepted a late fetch failure")
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(old) {
		t.Fatalf("first target changed after late failure: %q", got)
	}
}

func TestWriteArtifactsStagesEverythingBeforeReplacingTargets(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "first.txt")
	if err := os.WriteFile(first, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	blocker := filepath.Join(root, "not-a-directory")
	if err := os.WriteFile(blocker, []byte("block"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := writeArtifacts([]generatedArtifact{
		{path: first, contents: "new"},
		{path: filepath.Join(blocker, "second.txt"), contents: "new"},
	}, &strings.Builder{})
	if err == nil {
		t.Fatal("writeArtifacts accepted a late staging failure")
	}
	contents, readErr := os.ReadFile(first)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(contents) != "old" {
		t.Fatalf("first artifact changed after staging failure: %q", contents)
	}
}

func TestCommitFileUpdatesPreservesBackupWhenRestoreFails(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "first.txt")
	second := filepath.Join(root, "second.txt")
	for _, target := range []string{first, second} {
		if err := os.WriteFile(target, []byte("old:"+filepath.Base(target)), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	originalRename := renameFile
	renameFile = func(oldPath, newPath string) error {
		stageFailure := newPath == second && strings.Contains(filepath.Base(oldPath), ".vendorgen-stage-")
		rollbackFailure := newPath == first && strings.Contains(filepath.Base(oldPath), ".vendorgen-backup-")
		if stageFailure || rollbackFailure {
			return fmt.Errorf("injected rename failure")
		}
		return originalRename(oldPath, newPath)
	}
	t.Cleanup(func() { renameFile = originalRename })

	err := commitFileUpdates([]fileUpdate{
		{path: first, contents: []byte("new:first"), mode: 0o644},
		{path: second, contents: []byte("new:second"), mode: 0o644},
	})
	if err == nil {
		t.Fatal("commitFileUpdates accepted replacement and restoration failures")
	}
	backups, globErr := filepath.Glob(filepath.Join(root, ".vendorgen-backup-*"))
	if globErr != nil {
		t.Fatal(globErr)
	}
	if len(backups) != 1 {
		t.Fatalf("preserved backups = %v, want exactly one", backups)
	}
	if !strings.Contains(err.Error(), backups[0]) {
		t.Fatalf("error %q does not identify preserved backup %q", err, backups[0])
	}
	contents, readErr := os.ReadFile(backups[0])
	if readErr != nil || string(contents) != "old:first.txt" {
		t.Fatalf("preserved backup = %q, %v", contents, readErr)
	}
}

func TestCommitFileUpdatesPreservesCurrentBackupWhenImmediateRestoreFails(t *testing.T) {
	target := filepath.Join(t.TempDir(), "target.txt")
	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	originalRename := renameFile
	renameFile = func(oldPath, newPath string) error {
		if newPath == target && (strings.Contains(filepath.Base(oldPath), ".vendorgen-stage-") || strings.Contains(filepath.Base(oldPath), ".vendorgen-backup-")) {
			return fmt.Errorf("injected immediate restore failure")
		}
		return originalRename(oldPath, newPath)
	}
	t.Cleanup(func() { renameFile = originalRename })

	err := commitFileUpdates([]fileUpdate{{path: target, contents: []byte("new"), mode: 0o644}})
	backups, globErr := filepath.Glob(filepath.Join(filepath.Dir(target), ".vendorgen-backup-*"))
	if err == nil || globErr != nil || len(backups) != 1 {
		t.Fatalf("error = %v, backups = %v, glob error = %v", err, backups, globErr)
	}
	if !strings.Contains(err.Error(), backups[0]) {
		t.Fatalf("error %q does not identify preserved backup %q", err, backups[0])
	}
	contents, readErr := os.ReadFile(backups[0])
	if readErr != nil || string(contents) != "old" {
		t.Fatalf("preserved current backup = %q, %v", contents, readErr)
	}
}

func TestQuarantineRestorePreservesDirectoryWhenRenameFails(t *testing.T) {
	working := t.TempDir()
	oldWorking, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(working); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWorking) })
	stale := filepath.Join(vendorRoot, "module", "old", "runtime.js")
	if err := os.MkdirAll(filepath.Dir(stale), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stale, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, restore, _, err := quarantineStaleVersions([]vendoredDependency{{Module: "module", Dependency: dep{Version: "new"}}})
	if err != nil {
		t.Fatal(err)
	}
	quarantines, err := filepath.Glob(filepath.Join(filepath.Dir(vendorRoot), ".vendorgen-prune-*"))
	if err != nil || len(quarantines) != 1 {
		t.Fatalf("quarantine paths = %v, %v", quarantines, err)
	}

	originalRename := renameFile
	renameFile = func(oldPath, newPath string) error {
		if strings.HasPrefix(oldPath, quarantines[0]+string(filepath.Separator)) {
			return fmt.Errorf("injected quarantine restore failure")
		}
		return originalRename(oldPath, newPath)
	}
	t.Cleanup(func() { renameFile = originalRename })

	restoreErr := restore()
	if restoreErr == nil || !strings.Contains(restoreErr.Error(), quarantines[0]) {
		t.Fatalf("restore error = %v, want preserved quarantine path", restoreErr)
	}
	if _, statErr := os.Stat(quarantines[0]); statErr != nil {
		t.Fatalf("failed recovery quarantine was removed: %v", statErr)
	}
}

func TestCommitFileUpdatesRejectsDuplicateDestinations(t *testing.T) {
	target := filepath.Join(t.TempDir(), "same.txt")
	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := commitFileUpdates([]fileUpdate{
		{path: target, contents: []byte("first"), mode: 0o644},
		{path: filepath.Dir(target) + string(filepath.Separator) + "." + string(filepath.Separator) + filepath.Base(target), contents: []byte("second"), mode: 0o644},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate destination") {
		t.Fatalf("duplicate destination error = %v", err)
	}
	contents, readErr := os.ReadFile(target)
	if readErr != nil || string(contents) != "old" {
		t.Fatalf("target changed after duplicate rejection: %q, %v", contents, readErr)
	}
}

func TestRestoreStagedFileReportsNoOriginalWhenRemovingNewFileFails(t *testing.T) {
	target := filepath.Join(t.TempDir(), "new.txt")
	originalRemove := removeFile
	removeFile = func(path string) error {
		if path == target {
			return fmt.Errorf("injected remove failure")
		}
		return originalRemove(path)
	}
	t.Cleanup(func() { removeFile = originalRemove })

	err := restoreStagedFile(&stagedFile{update: fileUpdate{path: target}})
	if err == nil || !strings.Contains(err.Error(), "no original existed") || !strings.Contains(err.Error(), target) {
		t.Fatalf("restore error = %v", err)
	}
	if strings.Contains(err.Error(), "recovery artifact") {
		t.Fatalf("restore error incorrectly claims an empty recovery artifact: %v", err)
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
		`PrimaryURL: AlpineJSCDNURL`,
		`LocalURL: AlpineJSURL`,
		`Integrity: AlpineJSIntegrity`,
		`Enabled: true`,
		`IncludeInMinimal: true`,
		`Role: RuntimeRoleFirstParty`,
		`func defaultRuntimeMetadata() []RuntimeAssetMetadata`,
		`Name: "Alpine.js"`,
		`Version: "3.14.9"`,
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
		`Role: assets.RuntimeAssetRole("alpine")`,
		`Version: assets.AlpineVersion`,
		`FallbackLocalURL: assets.AlpineJSURL`,
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
		"[MIT](../assets/js/runtime/alpinejs/3.14.9/LICENSE.txt)",
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
	  "package_name": "alpinejs",
	  "provenance_url": "https://unpkg.com/alpinejs@{v}/package.json",
	  "integrity": "sha384-ywB1P0WjXou1oD1pmsZQBycsMqsO3tFjGotgWkP/W+2AhgcroefMI1i67KE0yCWn",
      "homepage": "https://alpinejs.dev",
      "license": "MIT",
	  "license_file": "LICENSE.txt",
	  "license_url": "https://raw.githubusercontent.com/alpinejs/alpine/v{v}/LICENSE.md",
	  "license_integrity": "sha384-ywB1P0WjXou1oD1pmsZQBycsMqsO3tFjGotgWkP/W+2AhgcroefMI1i67KE0yCWn",
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
