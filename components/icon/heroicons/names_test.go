package heroicons

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/internal/iconcatalog"
)

func TestReleasedCatalogPreservesGeneratedNames(t *testing.T) {
	catalogPath := filepath.Join(heroiconsRepoRoot(t), "internal", "iconcatalog", "testdata", "catalog.json")
	file, err := os.Open(catalogPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = file.Close() })
	catalog, err := iconcatalog.Load(file)
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Release != "v0.1.1" {
		t.Fatalf("catalog release = %q, want v0.1.1", catalog.Release)
	}

	releasedSymbols := make(map[string]string, len(Glyphs))
	for _, asset := range catalog.Assets {
		if asset.Namespace == "ui" && asset.Product == "heroicons" && asset.SpriteSymbol != "" {
			releasedSymbols[asset.CanonicalName] = asset.SpriteSymbol
		}
	}
	if len(releasedSymbols) != len(Glyphs) {
		t.Fatalf("released Heroicons count = %d, generated glyph count = %d", len(releasedSymbols), len(Glyphs))
	}

	seenNames := make(map[string]struct{}, len(Glyphs))
	seenSymbols := make(map[string]struct{}, len(Glyphs))
	for _, glyph := range Glyphs {
		if glyph.GoName == "" || glyph.CanonicalName == "" || glyph.Symbol == "" {
			t.Fatalf("invalid glyph: %#v", glyph)
		}
		if _, exists := seenNames[glyph.CanonicalName]; exists {
			t.Fatalf("duplicate generated canonical name %q", glyph.CanonicalName)
		}
		seenNames[glyph.CanonicalName] = struct{}{}
		if _, exists := seenSymbols[string(glyph.Symbol)]; exists {
			t.Fatalf("duplicate generated symbol %q", glyph.Symbol)
		}
		seenSymbols[string(glyph.Symbol)] = struct{}{}
		if got := releasedSymbols[glyph.CanonicalName]; got != string(glyph.Symbol) {
			t.Fatalf("generated glyph %q symbol = %q, released catalog symbol = %q", glyph.CanonicalName, glyph.Symbol, got)
		}
	}
}

func heroiconsRepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", ".."))
}

func TestEveryGeneratedSymbolExistsExactlyOnceInSprite(t *testing.T) {
	if SpriteURL != "/assets/icons/heroicons.svg" {
		t.Fatalf("SpriteURL = %q, want bundled Heroicons path", SpriteURL)
	}
	if len(Glyphs) != 67 {
		t.Fatalf("generated glyph count = %d, want 67", len(Glyphs))
	}

	seen := make(map[string]struct{}, len(Glyphs))
	for _, glyph := range Glyphs {
		if !strings.HasPrefix(string(glyph.Symbol), "hi-") {
			t.Errorf("%s symbol = %q, want hi-*", glyph.GoName, glyph.Symbol)
		}
		if _, duplicate := seen[string(glyph.Symbol)]; duplicate {
			t.Errorf("duplicate generated symbol %q", glyph.Symbol)
		}
		seen[string(glyph.Symbol)] = struct{}{}
	}

	recorder := httptest.NewRecorder()
	assets.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, SpriteURL, nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d, want 200", SpriteURL, recorder.Code)
	}

	counts := make(map[string]int, len(Glyphs))
	decoder := xml.NewDecoder(strings.NewReader(recorder.Body.String()))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("decode bundled sprite: %v", err)
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "symbol" {
			continue
		}
		for _, attr := range start.Attr {
			if attr.Name.Local == "id" {
				counts[attr.Value]++
			}
		}
	}
	if len(counts) != 1288 {
		t.Fatalf("sprite symbol count = %d, want Assets v0.2.1 count 1288", len(counts))
	}
	for _, glyph := range Glyphs {
		if got := counts[string(glyph.Symbol)]; got != 1 {
			t.Errorf("sprite symbol %q occurrences = %d, want 1", glyph.Symbol, got)
		}
	}
}
