package runtimegen

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/araihu/assets/assetmeta"
)

func TestLoadPreservesApprovedRuntimeOrder(t *testing.T) {
	model, err := loadRepositoryOverlay(t)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"alpine-collapse", "alpine-focus", "alpine-mask", "first-party",
		"dark-mode", "alpine", "htmx", "htmx-ext-sse", "htmx-ext-ws",
		"combobox", "action-group",
	}
	got := make([]string, len(model.Dependencies))
	for i, dependency := range model.Dependencies {
		got[i] = dependency.Role
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("roles = %v, want %v", got, want)
	}
}

func TestLoadRejectsInvalidOverlaySemantics(t *testing.T) {
	tests := []struct{ name, raw, want string }{
		{"unknown download", overlayWithOrder("missing/runtime-js"), "unknown references: missing/runtime-js"},
		{"download and local", overlayWithEntry("download: alpinejs/core-js\n      local: first-party"), "exactly one of download or local"},
		{"duplicate role", overlayWithDuplicateRole("alpine"), `duplicate role "alpine"`},
		{"duplicate go name", overlayWithDuplicateGoName("AlpineJS"), `duplicate go_name "AlpineJS"`},
		{"missing license", overlayWithout("license_ref"), "license_ref is required"},
		{"missing provenance", overlayWithout("provenance_ref"), "provenance_ref is required"},
		{"unsafe local URL", overlayWithLocalURL("../escape.js"), "local_url must be a safe /assets/js/*.js URL"},
		{"invalid wait and defer", overlayWithFlags(true, true), "wait_for_window_loaded and defer cannot both be true"},
		{"unsupported schema", strings.Replace(validOverlay, "schema: 1", "schema: 2", 1), "unsupported schema 2"},
		{"unknown field", strings.Replace(validOverlay, "metadata:\n", "metadata:\n  surprise: true\n", 1), "field surprise not found"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Load(strings.NewReader(test.raw), fixtureInventory(t))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func loadRepositoryOverlay(t *testing.T) (Model, error) {
	t.Helper()
	file, err := os.Open("../../assets/runtime.overlay.yaml")
	if err != nil {
		return Model{}, err
	}
	t.Cleanup(func() { _ = file.Close() })
	return Load(file, fixtureInventory(t))
}

func overlayWithOrder(ref string) string {
	return strings.Replace(validOverlay, "alpinejs/core-js", ref, 1)
}

func overlayWithEntry(entry string) string {
	return strings.Replace(validOverlay, "download: alpinejs/core-js", entry, 1)
}

func overlayWithDuplicateRole(role string) string {
	return strings.Replace(validOverlay, "role: htmx", "role: "+role, 1)
}

func overlayWithDuplicateGoName(name string) string {
	return strings.Replace(validOverlay, "go_name: HTMX", "go_name: "+name, 1)
}

func overlayWithout(field string) string {
	lines := strings.Split(validOverlay, "\n")
	for index, line := range lines {
		if strings.Contains(line, field+":") {
			return strings.Join(append(lines[:index], lines[index+1:]...), "\n")
		}
	}
	return validOverlay
}

func overlayWithLocalURL(value string) string {
	return strings.Replace(validOverlay, "/assets/js/goshtoso.min.js", value, 1)
}

func overlayWithFlags(deferAsset, wait bool) string {
	replacement := "defer: " + boolText(deferAsset) + "\n          wait_for_window_loaded: " + boolText(wait)
	return strings.Replace(validOverlay, "defer: false\n          wait_for_window_loaded: true", replacement, 1)
}

func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func fixtureInventory(t *testing.T) *assetmeta.Inventory {
	t.Helper()
	resource := func(name, version string, downloads ...string) assetmeta.Resource {
		result := assetmeta.Resource{Name: name, Version: version}
		for _, download := range downloads {
			result.Downloads = append(result.Downloads, assetmeta.Download{
				Name: download, URL: "https://example.test/" + name + "/" + download,
				Path:      "assets/js/runtime/" + name + "/" + version + "/" + download,
				Integrity: "sha384-fixture", Hash: "sha384:fixture",
			})
		}
		return result
	}
	inventory, err := assetmeta.NewInventory([]assetmeta.Resource{
		resource("alpinejs", "3.14.9", "collapse-js", "collapse-license", "collapse-package", "focus-js", "focus-license", "focus-package", "mask-js", "mask-license", "mask-package", "core-js", "core-license", "core-package"),
		resource("htmx", "2.0.8", "core-js", "license", "package"),
		resource("htmx-ext-sse", "2.2.3", "runtime-js", "license", "package"),
		resource("htmx-ext-ws", "2.0.3", "runtime-js", "license", "package"),
	})
	if err != nil {
		t.Fatal(err)
	}
	return inventory
}

const validOverlay = `schema: 1
metadata:
  loader:
    role: dependency-loader
    go_name: DependencyLoader
    name: Goshtoso dependency loader
    local_url: /assets/js/dependency-loader.js
    purpose: Loads dependencies
    enabled: true
    include_in_minimal: true
    defer: true
  order:
    - download: alpinejs/core-js
    - local: first-party
    - download: htmx/core-js
  first_party:
    first-party:
      role: first-party
      go_name: FirstPartyBundle
      role_go_name: FirstParty
      name: Goshtoso component runtime
      local_url: /assets/js/goshtoso.min.js
      purpose: First-party behavior
      enabled: true
      include_in_minimal: true
      defer: true
resources:
  alpinejs:
    downloads:
      core-js:
        metadata:
          role: alpine
          go_name: AlpineJS
          name: Alpine.js
          package_name: alpinejs
          homepage: https://alpinejs.dev
          license: MIT
          license_ref: alpinejs/core-license
          provenance_ref: alpinejs/core-package
          purpose: Reactive UI state
          attribution: true
          enabled: true
          include_in_minimal: true
          defer: true
  htmx:
    downloads:
      core-js:
        metadata:
          role: htmx
          go_name: HTMX
          name: htmx
          package_name: htmx.org
          homepage: https://htmx.org
          license: Zero-Clause BSD
          license_ref: htmx/license
          provenance_ref: htmx/package
          purpose: Server-driven interactions
          attribution: true
          enabled: true
          include_in_minimal: true
          defer: false
          wait_for_window_loaded: true
`
