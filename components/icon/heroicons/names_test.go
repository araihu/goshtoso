package heroicons

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/assets"
)

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
	for _, glyph := range Glyphs {
		if got := counts[string(glyph.Symbol)]; got != 1 {
			t.Errorf("sprite symbol %q occurrences = %d, want 1", glyph.Symbol, got)
		}
	}
}
