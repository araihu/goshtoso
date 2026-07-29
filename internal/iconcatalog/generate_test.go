package iconcatalog

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var fixtureOptions = Options{
	Package:     "heroicons",
	Namespace:   "ui",
	Product:     "heroicons",
	SpriteURL:   "/assets/icons/heroicons.svg",
	ConstPrefix: "Icon",
}

func golden(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestGenerateTypedSymbolsDeterministically(t *testing.T) {
	got, err := Generate(fixture(t), fixtureOptions)
	if err != nil {
		t.Fatal(err)
	}
	if diff := firstDifference(golden(t, "names.golden"), got); diff != "" {
		t.Fatal(diff)
	}
}

func TestGenerateRejectsInvalidSelectedAssets(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Asset)
		want string
	}{
		{"non-SVG", func(a *Asset) { a.Format = "png" }, "format"},
		{"missing sprite", func(a *Asset) { a.SpriteSymbol = "" }, "spriteSymbol"},
		{"incompatible color", func(a *Asset) { a.ColorBehavior = "protected" }, "colorBehavior"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			catalog := fixture(t)
			tt.edit(&catalog.Assets[0])
			_, err := Generate(catalog, fixtureOptions)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Generate() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestGenerateRejectsIdentifierCollision(t *testing.T) {
	catalog := fixture(t)
	catalog.Assets[0].CanonicalName = "ui-hi-16-solid-a-b"
	catalog.Assets[1].CanonicalName = "ui-hi-16-solid-a--b"
	_, err := Generate(catalog, fixtureOptions)
	if err == nil || !strings.Contains(err.Error(), "identifier collision") {
		t.Fatalf("Generate() error = %v, want identifier collision", err)
	}
}

func TestGenerateRejectsDuplicateSelectedSymbol(t *testing.T) {
	catalog := fixture(t)
	catalog.Assets[1].SpriteSymbol = catalog.Assets[0].SpriteSymbol
	_, err := Generate(catalog, fixtureOptions)
	if err == nil || !strings.Contains(err.Error(), "duplicate spriteSymbol") {
		t.Fatalf("Generate() error = %v, want duplicate spriteSymbol", err)
	}
}

func TestRunWritesAndChecksGeneratedFile(t *testing.T) {
	out := filepath.Join(t.TempDir(), "names_gen.go")
	args := []string{
		"-catalog", filepath.Join("testdata", "catalog.json"),
		"-out", out,
		"-package", fixtureOptions.Package,
		"-namespace", fixtureOptions.Namespace,
		"-product", fixtureOptions.Product,
		"-sprite-url", fixtureOptions.SpriteURL,
		"-const-prefix", fixtureOptions.ConstPrefix,
	}
	var stdout, stderr bytes.Buffer
	if err := Run(args, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if err := Run(append(args, "-check"), &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(out, []byte("stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Run(append(args, "-check"), &stdout, &stderr); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("Run(-check) error = %v, want stale output", err)
	}
}

func firstDifference(want, got []byte) string {
	if bytes.Equal(want, got) {
		return ""
	}
	return "generated source differs from names.golden"
}
