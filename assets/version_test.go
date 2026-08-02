package assets

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestTailwindVersion(t *testing.T) {
	got := TailwindVersion()
	if got != "4.3.3" {
		t.Fatalf("TailwindVersion() = %q, want %q", got, "4.3.3")
	}
	if strings.HasPrefix(got, "v") {
		t.Fatalf("TailwindVersion() must not include a leading v: %q", got)
	}
}

func TestRuntimeUsesOverlayWithoutLegacyManifests(t *testing.T) {
	if _, err := os.Stat("runtime.overlay.yaml"); err != nil {
		t.Fatalf("runtime overlay missing: %v", err)
	}
	for _, retired := range []string{"js/runtime/manifest.json", "js/runtime/versions.json", "tailwind.version"} {
		if _, err := os.Stat(retired); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("retired runtime source %s still exists: %v", retired, err)
		}
	}
}

func TestThemeCSS(t *testing.T) {
	b, err := ThemeCSS()
	if err != nil {
		t.Fatalf("ThemeCSS() error: %v", err)
	}
	s := string(b)
	for _, want := range []string{"@custom-variant dark", "[data-theme=minimal]", "@theme"} {
		if !strings.Contains(s, want) {
			t.Errorf("ThemeCSS() missing %q", want)
		}
	}
	if strings.Contains(s, `@import "tailwindcss"`) {
		t.Error("ThemeCSS() must not contain the tailwind import")
	}
}

func TestVendorVersions(t *testing.T) {
	byRole := make(map[RuntimeAssetRole]RuntimeAssetMetadata)
	for _, dependency := range DefaultRuntimeMetadata() {
		byRole[dependency.Role] = dependency
	}
	cases := map[string]string{
		"Alpine":     AlpineVersion(),
		"HTMX":       HTMXVersion(),
		"HTMXExtSSE": HTMXExtSSEVersion(),
		"HTMXExtWS":  HTMXExtWSVersion(),
	}
	want := map[string]string{
		"Alpine":     byRole[RuntimeRoleAlpineJS].Version,
		"HTMX":       byRole[RuntimeRoleHTMX].Version,
		"HTMXExtSSE": byRole[RuntimeRoleHTMXExtSSE].Version,
		"HTMXExtWS":  byRole[RuntimeRoleHTMXExtWS].Version,
	}
	for k, got := range cases {
		if got != want[k] {
			t.Errorf("%sVersion() = %q, want %q", k, got, want[k])
		}
	}
}

func TestRuntimeFilesUseTheirDeclaredEmbed(t *testing.T) {
	manifest := DefaultRuntimeManifest()
	declared := append([]RuntimeAsset{manifest.Loader}, manifest.Dependencies...)
	for _, asset := range declared {
		p := strings.TrimPrefix(asset.LocalURL, "/assets/")
		_, vendored := RuntimeHash(asset.Role)
		_, err := files.ReadFile(p)
		if vendored && err == nil {
			t.Errorf("vendored file for %s is duplicated in first-party embed: %s", asset.Role, p)
		}
		if !vendored && err != nil {
			t.Errorf("first-party embedded file missing for %s: %s: %v", asset.Role, p, err)
		}
	}
}
