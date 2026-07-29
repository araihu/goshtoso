package iconcatalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T) Catalog {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := Load(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestLoadRejectsUnsupportedSchemaAndDuplicateSymbol(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schemaVersion":2,"assets":[]}`))
	if err == nil || !strings.Contains(err.Error(), "unsupported schemaVersion") {
		t.Fatalf("Load() error = %v, want unsupported schemaVersion", err)
	}

	duplicate := `{"schemaVersion":1,"assets":[` +
		`{"canonicalName":"ui-a","spriteSymbol":"same"},` +
		`{"canonicalName":"ui-b","spriteSymbol":"same"}` +
		`]}`
	_, err = Load(strings.NewReader(duplicate))
	if err == nil || !strings.Contains(err.Error(), "duplicate spriteSymbol") {
		t.Fatalf("Load() error = %v, want duplicate spriteSymbol", err)
	}
}

func TestLoadRejectsEmptyAndDuplicateCanonicalName(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schemaVersion":1,"assets":[{"canonicalName":"","spriteSymbol":"a"}]}`))
	if err == nil || !strings.Contains(err.Error(), "empty canonicalName") {
		t.Fatalf("Load() error = %v, want empty canonicalName", err)
	}

	duplicate := `{"schemaVersion":1,"assets":[` +
		`{"canonicalName":"ui-a","spriteSymbol":"a"},` +
		`{"canonicalName":"ui-a","spriteSymbol":"b"}` +
		`]}`
	_, err = Load(strings.NewReader(duplicate))
	if err == nil || !strings.Contains(err.Error(), "duplicate canonicalName") {
		t.Fatalf("Load() error = %v, want duplicate canonicalName", err)
	}
}
