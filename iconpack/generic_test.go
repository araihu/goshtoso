package iconpack

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateFromGenericBootstrapIconsManifest(t *testing.T) {
	root := t.TempDir()
	alarm := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-alarm" viewBox="0 0 16 16"><path d="M8 1a7 7 0 1 0 0 14A7 7 0 0 0 8 1z"/></svg>`)
	bootstrap := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-bootstrap" viewBox="0 0 16 16"><path d="M5.062 12h2.39c1.77 0 2.91-.893 2.91-2.35 0-1.11-.66-1.87-1.71-2.15v-.04c.8-.27 1.3-.97 1.3-1.87C9.952 4.32 8.87 3 7.07 3H5.062z"/></svg>`)
	license := []byte("The MIT License (MIT)\nCopyright (c) 2019-2024 The Bootstrap Authors\n")
	writeFixtureFile(t, root, "icons/alarm.svg", alarm)
	writeFixtureFile(t, root, "icons/bootstrap.svg", bootstrap)
	writeFixtureFile(t, root, "LICENSE", license)
	manifest := sourceManifest{
		SchemaVersion: 1,
		Name:          "Bootstrap Icons",
		Release:       "v1.11.3",
		Source:        "https://github.com/twbs/icons/tree/v1.11.3",
		License:       "MIT",
		LicensePath:   "LICENSE",
		LicenseSHA256: hashBytes(license),
		Icons: []sourceManifestIcon{
			{CanonicalName: "bootstrap-icons-alarm", Path: "icons/alarm.svg", SpriteSymbol: "bi-alarm", ViewBox: "0 0 16 16", ColorBehavior: "currentColor", SHA256: hashBytes(alarm)},
			{CanonicalName: "bootstrap-icons-bootstrap", Path: "icons/bootstrap.svg", SpriteSymbol: "bi-bootstrap", ViewBox: "0 0 16 16", ColorBehavior: "currentColor", SHA256: hashBytes(bootstrap)},
		},
	}
	manifestBytes, err := marshalDocument(manifest)
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(root, "bootstrap-icons.goshtoso.json")
	if err := os.WriteFile(manifestPath, manifestBytes, 0o644); err != nil {
		t.Fatal(err)
	}

	output := filepath.Join(t.TempDir(), "bootstrapicons")
	result, err := Generate(t.Context(), Options{
		SourceRoot:     root,
		SourceManifest: manifestPath,
		Names:          []string{"bootstrap-icons-alarm", "bootstrap-icons-bootstrap"},
		OutputDir:      output,
		Package:        "bootstrapicons",
		ConstPrefix:    "Icon",
		SpriteURL:      "/assets/icons/bootstrapicons/sprite.svg",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Published || result.SelectedCount != 2 {
		t.Fatalf("Generate() result = %+v", result)
	}
	assertFileContains(t, filepath.Join(output, "sprite.svg"), `id="bi-alarm"`)
	assertFileContains(t, filepath.Join(output, "sprite.svg"), `id="bi-bootstrap"`)
	assertFileContains(t, filepath.Join(output, "sprite.svg"), `viewBox="0 0 16 16"`)
	assertFileContains(t, filepath.Join(output, "icons_gen.go"), `IconBootstrapIconsAlarm`)
	assertFileContains(t, filepath.Join(output, "icons_gen.go"), `"bi-alarm"`)
	assertFileContains(t, filepath.Join(output, "manifest.json"), `"sourceKind": "source-root"`)
	assertFileContains(t, filepath.Join(output, "manifest.json"), `"canonicalName": "bootstrap-icons-alarm"`)
	assertFileContains(t, filepath.Join(output, "provenance.json"), `"source": "https://github.com/twbs/icons/tree/v1.11.3:icons/alarm.svg"`)
	assertFileContains(t, filepath.Join(output, "NOTICE"), "Bootstrap Icons v1.11.3")

	generated, err := os.ReadFile(filepath.Join(output, "sprite.svg"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(generated), "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"16\"") {
		t.Fatal("standalone source root leaked into generated sprite")
	}

	archive := filepath.Join(t.TempDir(), "bootstrap-icons.tar.gz")
	writeTarGzip(t, root, archive)
	archiveBytes, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	archiveOutput := filepath.Join(t.TempDir(), "bootstrapicons")
	archiveResult, err := Generate(t.Context(), Options{
		SourceArchive:       archive,
		SourceArchiveSHA256: hashBytes(archiveBytes),
		SourceManifest:      manifestPath,
		Names:               []string{"bootstrap-icons-alarm"},
		OutputDir:           archiveOutput,
		Package:             "bootstrapicons",
		ConstPrefix:         "Icon",
		SpriteURL:           "/assets/icons/bootstrapicons/sprite.svg",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !archiveResult.Published || archiveResult.SelectedCount != 1 {
		t.Fatalf("archive Generate() result = %+v", archiveResult)
	}
	assertFileContains(t, filepath.Join(archiveOutput, "manifest.json"), `"sourceKind": "source-archive"`)
}

func TestGenericSourceManifestRejectsUnhashedLicense(t *testing.T) {
	manifest := sourceManifest{
		SchemaVersion: 1,
		Name:          "Bootstrap Icons",
		Release:       "v1.11.3",
		Source:        "bootstrap-icons",
		License:       "MIT",
		LicensePath:   "LICENSE",
		Icons: []sourceManifestIcon{{
			CanonicalName: "bootstrap-icons-alarm",
			Path:          "icons/alarm.svg",
			SpriteSymbol:  "bi-alarm",
			ViewBox:       "0 0 16 16",
			SHA256:        strings.Repeat("a", 64),
		}},
	}
	if err := validateSourceManifest(manifest); err == nil || !strings.Contains(err.Error(), "licenseSha256") {
		t.Fatalf("validateSourceManifest() error = %v, want missing license hash", err)
	}
}

func TestStandaloneSVGRejectsExternalReferences(t *testing.T) {
	raw := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><image href="https://example.test/icon.svg"/></svg>`)
	if _, err := standaloneSVGSymbol(raw, "icon.svg", "bi-icon", "0 0 16 16"); err == nil || !strings.Contains(err.Error(), "external SVG reference") {
		t.Fatalf("standaloneSVGSymbol() error = %v, want external-reference rejection", err)
	}
}

func TestStandaloneSVGPreservesRootPaintAttributes(t *testing.T) {
	raw := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M0 0h24v24H0z"/></svg>`)
	symbol, err := standaloneSVGSymbol(raw, "icon.svg", "hi-icon", "0 0 24 24")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`fill="none"`, `stroke="currentColor"`, `stroke-width="1.5"`} {
		if !strings.Contains(string(symbol), want) {
			t.Errorf("standalone symbol is missing root paint attribute %q: %s", want, symbol)
		}
	}
}
