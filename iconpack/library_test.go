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
