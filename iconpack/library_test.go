package iconpack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/araihu/goshtoso/iconlibrary"
)

func TestLibraryLockedCatalogWithRasterAndVariants(t *testing.T) {
	var p bytes.Buffer
	if err := png.Encode(&p, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	archive := iconpackTestArchive(t, map[string][]byte{"png/demo.png": p.Bytes(), "png/demo-light.png": p.Bytes(), "LICENSE": []byte("CC-BY-4.0"), "index.json": []byte(`[{"Name":"Demo App","Reference":"demo","PNG":"Yes","Light":"Yes","Category":"Self-Hosted","Tags":"Notes,Tools"}]`)})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(archive) }))
	defer server.Close()
	root := t.TempDir()
	config := filepath.Join(root, ".iconpack.yaml")
	data := fmt.Sprintf("schemaVersion: 1\nsources:\n  - id: selfhst\n    kind: archive\n    url: %s/archive.tar.gz\n    license: CC-BY-4.0\n    licensePath: LICENSE\n    metadataPath: index.json\n    metadataFormat: selfhst\n    formats: [png]\n", server.URL)
	if err := os.WriteFile(config, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	opts := Options{Library: true, ConfigPath: config, OutputDir: filepath.Join(root, "icons"), Trust: true, AllowHTTP: true}
	result, err := Generate(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if result.SelectedCount != 1 {
		t.Fatal(result)
	}
	var catalog iconlibrary.Catalog
	if err := json.Unmarshal(mustReadFile(t, filepath.Join(result.OutputDir, "catalog.json")), &catalog); err != nil {
		t.Fatal(err)
	}
	icon := catalog.Icons[0]
	if icon.ID != "selfhst:demo" || icon.Name != "Demo App" || len(icon.Variants) != 2 || len(icon.Tags) != 3 || icon.Variants[0].MIME != "image/png" {
		t.Fatalf("catalog: %+v", icon)
	}
	var manifest outputManifest
	if err := json.Unmarshal(mustReadFile(t, filepath.Join(result.OutputDir, "manifest.json")), &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Release != result.Release || manifest.CatalogSchemaVersion != catalog.SchemaVersion || manifest.CatalogSHA256 != result.CatalogSHA256 {
		t.Fatalf("manifest does not identify its catalog: %+v", manifest)
	}
	assertLibraryManifestAndDeterminism(t, opts, result, manifest, root)
	opts.Trust = false
	opts.Check = true
	if _, err := Generate(t.Context(), opts); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(result.OutputDir, icon.Variants[0].Path), []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(t.Context(), opts); err == nil {
		t.Fatal("accepted modified output")
	}
}

func TestLibraryGroupsFormatsDeterministically(t *testing.T) {
	source := resolvedConfigSource{ID: "local", Formats: []string{"svg", "png"}}
	files := map[string][]byte{"local/foo.png": {}, "local/foo.svg": {}, "local/bar.svg": {}}
	icons, err := sourceLibraryIcons(t.Context(), source, memoryFiles(files))
	if err != nil {
		t.Fatal(err)
	}
	if len(icons) != 2 || icons[0].ID != "local:bar" || icons[1].ID != "local:foo" || len(icons[1].Variants) != 2 {
		t.Fatalf("expected one icon per reference: %+v", icons)
	}
	if icons[1].Variants[0].Path != "foo.svg" || icons[1].Variants[1].Path != "foo.png" {
		t.Fatalf("format preference not preserved: %+v", icons[1].Variants)
	}
}

func TestSelfhstAvailabilityStillRequiresDeclaredFiles(t *testing.T) {
	for _, tc := range []struct {
		name, png, light string
		files            map[string][]byte
		wantIcons        int
		wantError        bool
	}{
		{name: "unavailable format", png: "No", light: "Yes"},
		{name: "declared default missing", png: "Yes", wantError: true},
		{name: "declared light missing", png: "Yes", light: "Yes", files: map[string][]byte{"selfhst/png/demo.png": {}}, wantError: true},
		{name: "available default", png: "Yes", files: map[string][]byte{"selfhst/png/demo.png": {}}, wantIcons: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := tc.files
			if files == nil {
				files = map[string][]byte{}
			}
			files["selfhst/index.json"] = []byte(fmt.Sprintf(`[{"Name":"Demo","Reference":"demo","SVG":"Yes","PNG":%q,"Light":%q}]`, tc.png, tc.light))
			icons, err := selfhstIcons(t.Context(), resolvedConfigSource{ID: "selfhst", MetadataPath: "index.json"}, memoryFiles(files))
			if (err != nil) != tc.wantError || len(icons) != tc.wantIcons {
				t.Fatalf("icons=%+v, err=%v", icons, err)
			}
		})
	}
}

func assertLibraryManifestAndDeterminism(t *testing.T, opts Options, result Result, manifest outputManifest, root string) {
	t.Helper()
	for _, file := range manifest.Files {
		data := mustReadFile(t, filepath.Join(result.OutputDir, file.Path))
		if len(data) != file.Bytes || hashBytes(data) != file.SHA256 || file.Mode != "0644" {
			t.Fatalf("manifest mismatch: %+v", file)
		}
	}
	if len(manifest.Files) != 6 {
		t.Fatalf("unexpected output count: %d", len(manifest.Files))
	}
	first := readFixtureTree(t, result.OutputDir)
	second := opts
	second.Trust = false
	second.OutputDir = filepath.Join(root, "second")
	if _, err := Generate(t.Context(), second); err != nil {
		t.Fatal(err)
	}
	secondFiles := readFixtureTree(t, second.OutputDir)
	if len(first) != len(secondFiles) {
		t.Fatalf("nondeterministic output count: %d != %d", len(first), len(secondFiles))
	}
	for name, data := range first {
		other, ok := secondFiles[name]
		if !ok || !bytes.Equal(data, other) {
			t.Fatalf("nondeterministic output: %s", name)
		}
	}
}
