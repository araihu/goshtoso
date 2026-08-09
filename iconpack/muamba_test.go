package iconpack

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateFromIconpackConfigUsesOwnTOFUAndMultipleSources(t *testing.T) {
	archive := iconpackTestArchive(t, map[string][]byte{
		"icons/16/solid/academic-cap.svg":   []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M1 1"/></svg>`),
		"icons/24/outline/academic-cap.svg": []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M2 2"/></svg>`),
		"LICENSE":                           []byte("MIT license\n"),
	})
	fileIcon := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><path d="M3 3"/></svg>`)
	fileLicense := []byte("Apache license\n")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/tree.tar.gz":
			w.Header().Set("Content-Type", "application/gzip")
			_, _ = w.Write(archive)
		case "/icon.svg":
			_, _ = w.Write(fileIcon)
		case "/file-license.txt":
			_, _ = w.Write(fileLicense)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	root := t.TempDir()
	// A malformed parent Muamba file must not be discovered by the iconpack
	// engine, which receives explicit declaration and lock paths.
	if err := os.WriteFile(filepath.Join(root, "muamba.yaml"), []byte("not: a muamba manifest\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, ".iconpack.yaml")
	config := fmt.Sprintf(`schemaVersion: 1
sources:
  - id: heroicons
    kind: archive
    url: %s/tree.tar.gz
    packName: heroicons
    paths:
      - icons/16/solid/academic-cap.svg
    license: MIT
    licensePath: LICENSE
  - id: custom-file
    kind: file
    url: %s/icon.svg
    path: custom.svg
    license: Apache-2.0
    licensePath: file-license.txt
    licenseUrl: %s/file-license.txt
`, server.URL, server.URL, server.URL)
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	output := filepath.Join(root, "generated", "appicons")
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		t.Fatal(err)
	}
	result, err := Generate(t.Context(), Options{
		ConfigPath: configPath, Trust: true, AllowHTTP: true,
		OutputDir: output, Package: "appicons", ConstPrefix: "Icon", SpriteURL: "/assets/icons/appicons/sprite.svg",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Published || result.SelectedCount != 2 {
		t.Fatalf("Generate() result = %+v", result)
	}
	for _, filename := range []string{".iconpack.lock.yaml", ".iconpack.engine.yaml", "generated/appicons/sprite.svg", "generated/appicons/PROVENANCE/.iconpack.yaml"} {
		if _, err := os.Stat(filepath.Join(root, filename)); err != nil {
			t.Fatalf("expected %s: %v", filename, err)
		}
	}
	assertFileContains(t, filepath.Join(output, "sprite.svg"), `id="heroicons-icons-16-solid-academic-cap"`)
	assertFileContains(t, filepath.Join(output, "sprite.svg"), `id="custom-file-custom"`)
	assertFileContains(t, filepath.Join(output, "icons_gen.go"), `NameHeroiconsIcons16SolidAcademicCap`)
	assertFileContains(t, filepath.Join(output, "icons_gen.go"), `NameCustomFileCustom`)
	assertFileContains(t, filepath.Join(output, "NOTICE"), "Apache-2.0")
	if strings.Contains(string(mustReadFile(t, filepath.Join(output, "sprite.svg"))), "outline/academic") {
		t.Fatal("unselected tree icon was generated")
	}

	second, err := Generate(t.Context(), Options{
		ConfigPath: configPath, AllowHTTP: true,
		OutputDir: output, Package: "appicons", ConstPrefix: "Icon", SpriteURL: "/assets/icons/appicons/sprite.svg",
	})
	if err != nil {
		t.Fatal(err)
	}
	if second.Published {
		t.Fatal("identical locked generation republished output")
	}

	if err := os.WriteFile(configPath, []byte(strings.Replace(config, "custom.svg", "changed.svg", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(t.Context(), Options{
		ConfigPath: configPath, AllowHTTP: true,
		OutputDir: output, Package: "appicons", ConstPrefix: "Icon", SpriteURL: "/assets/icons/appicons/sprite.svg",
	}); err == nil || !strings.Contains(err.Error(), "lock") {
		t.Fatalf("changed config was accepted: %v", err)
	}
}

func TestParseGitHubTreeURL(t *testing.T) {
	archiveURL, ref, root, repo, err := parseGitHubTreeURL("https://github.com/tailwindlabs/heroicons/tree/master/src")
	if err != nil {
		t.Fatal(err)
	}
	if archiveURL != "https://codeload.github.com/tailwindlabs/heroicons/tar.gz/master" || ref != "master" || repo != "heroicons" || strings.Join(root, "/") != "src" {
		t.Fatalf("parseGitHubTreeURL() = %q, %q, %v, %q", archiveURL, ref, root, repo)
	}
}

func TestDecodeJSONIconpackConfig(t *testing.T) {
	config, err := decodeIconpackConfig(".iconpack.json", []byte(`{
  "schemaVersion": 1,
  "sources": [{
    "id": "bootstrap",
    "kind": "file",
    "url": "https://example.test/icons/alarm.svg",
    "path": "icons/alarm.svg",
    "license": "MIT",
    "licensePath": "LICENSE",
    "licenseUrl": "https://example.test/LICENSE"
  }]
}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(config.Sources) != 1 || config.Sources[0].Path != "icons/alarm.svg" {
		t.Fatalf("decoded JSON config = %#v", config)
	}
}

func TestNonGitSourceWithoutPackNameUsesNormalizedPath(t *testing.T) {
	resolved, err := resolveConfigSource(ConfigSource{
		ID: "bootstrap-source", Kind: "file", URL: "https://example.test/icons/alarm.svg",
		Path: "icons/alarm.svg", License: "MIT", LicensePath: "LICENSE",
		LicenseURL: "https://example.test/LICENSE",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.PackName != "" {
		t.Fatalf("non-Git fallback pack name = %q, want empty path-prefix marker", resolved.PackName)
	}
	families, assets, err := buildMuambaAssets([]resolvedConfigSource{resolved}, map[string][]byte{
		"bootstrap-source/LICENSE":         []byte("MIT"),
		"bootstrap-source/icons/alarm.svg": []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"/>`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := families["custom/bootstrap-source"]; !ok || len(assets) != 1 || assets[0].CanonicalName != "bootstrap-source-icons-alarm" {
		t.Fatalf("families = %#v, assets = %#v", families, assets)
	}
}

func TestIconpackRejectsMuambaManifestNames(t *testing.T) {
	for _, path := range []string{"muamba.yaml", ".muamba.yaml", "muamba.lock.yaml", ".muamba.lock.yaml"} {
		t.Run(path, func(t *testing.T) {
			if err := rejectMuambaPath(path); err == nil {
				t.Fatal("Muamba manifest name was accepted")
			}
		})
	}
}

func iconpackTestArchive(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var output strings.Builder
	gzipWriter := gzip.NewWriter(writerFunc{write: func(value []byte) (int, error) {
		return output.Write(value)
	}})
	tarWriter := tar.NewWriter(gzipWriter)
	for name, contents := range files {
		if err := tarWriter.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(contents))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tarWriter.Write(contents); err != nil {
			t.Fatal(err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return []byte(output.String())
}

type writerFunc struct {
	write func([]byte) (int, error)
}

func (writer writerFunc) Write(value []byte) (int, error) { return writer.write(value) }

func mustReadFile(t *testing.T, filename string) []byte {
	t.Helper()
	contents, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	return contents
}

var _ io.Writer = writerFunc{}
