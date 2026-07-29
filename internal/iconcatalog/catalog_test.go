package iconcatalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T) Catalog {
	t.Helper()
	catalog, err := Load(strings.NewReader(fixtureJSON(t)))
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func fixtureJSON(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestLoadRejectsUnsupportedSchemaAndDuplicateSymbol(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schemaVersion":2,"assets":[]}`))
	if err == nil || !strings.Contains(err.Error(), "unsupported schemaVersion") {
		t.Fatalf("Load() error = %v, want unsupported schemaVersion", err)
	}

	duplicate := strings.Replace(fixtureJSON(t), "\"spriteSymbol\": \"hi-16-solid-arrow-down-tray\"", "\"spriteSymbol\": \"hi-16-solid-arrow-down\"", 1)
	_, err = Load(strings.NewReader(duplicate))
	if err == nil || !strings.Contains(err.Error(), "duplicate spriteSymbol") {
		t.Fatalf("Load() error = %v, want duplicate spriteSymbol", err)
	}
}

func TestLoadRejectsEmptyAndDuplicateCanonicalName(t *testing.T) {
	_, err := Load(strings.NewReader(strings.Replace(fixtureJSON(t), "ui-hi-16-solid-arrow-down", "", 1)))
	if err == nil || !strings.Contains(err.Error(), "empty canonicalName") {
		t.Fatalf("Load() error = %v, want empty canonicalName", err)
	}

	duplicate := strings.Replace(fixtureJSON(t), "ui-hi-16-solid-arrow-down-tray", "ui-hi-16-solid-arrow-down", 1)
	_, err = Load(strings.NewReader(duplicate))
	if err == nil || !strings.Contains(err.Error(), "duplicate canonicalName") {
		t.Fatalf("Load() error = %v, want duplicate canonicalName", err)
	}
}

func TestLoadRejectsUnknownCaseVariantDuplicateAndMissingKeys(t *testing.T) {
	valid := fixtureJSON(t)
	tests := []struct {
		name string
		json string
		want string
	}{
		{
			name: "unknown top-level key",
			json: strings.Replace(valid, "\n  \"assets\":", "\n  \"unexpected\": true,\n  \"assets\":", 1),
			want: "unknown top-level key",
		},
		{
			name: "case variant top-level key",
			json: strings.Replace(valid, "\"schemaVersion\"", "\"SchemaVersion\"", 1),
			want: "unknown top-level key",
		},
		{
			name: "duplicate top-level key",
			json: strings.Replace(valid, "\"schemaVersion\": 1,", "\"schemaVersion\": 1,\n  \"schemaVersion\": 1,", 1),
			want: "duplicate top-level key",
		},
		{
			name: "unknown asset key",
			json: strings.Replace(valid, "\"namespace\": \"ui\",", "\"unknown\": true,\n      \"namespace\": \"ui\",", 1),
			want: "unknown asset key",
		},
		{
			name: "duplicate asset key",
			json: strings.Replace(valid, "\"namespace\": \"ui\",", "\"namespace\": \"ui\",\n      \"namespace\": \"ui\",", 1),
			want: "duplicate asset key",
		},
		{
			name: "unknown dimensions key",
			json: strings.Replace(valid, "\"viewBox\": \"0 0 16 16\"", "\"viewBox\": \"0 0 16 16\", \"depth\": 1", 1),
			want: "unknown dimensions key",
		},
		{
			name: "duplicate dimensions key",
			json: strings.Replace(valid, "\"viewBox\": \"0 0 16 16\"", "\"viewBox\": \"0 0 16 16\", \"viewBox\": \"0 0 16 16\"", 1),
			want: "duplicate dimensions key",
		},
		{
			name: "missing top-level key",
			json: strings.Replace(valid, "  \"release\": \"v0.1.0\",\n", "", 1),
			want: "missing top-level key \"release\"",
		},
		{
			name: "missing asset key",
			json: strings.Replace(valid, "      \"license\": \"MIT\",\n", "", 1),
			want: "missing asset key \"license\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(strings.NewReader(tt.json))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Load() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestLoadRejectsMoreThanOneJSONValue(t *testing.T) {
	_, err := Load(strings.NewReader(fixtureJSON(t) + "\nnull\n"))
	if err == nil || !strings.Contains(err.Error(), "more than one JSON value") {
		t.Fatalf("Load() error = %v, want more than one JSON value", err)
	}
}

func TestLoadValidatesSchemaV1Semantics(t *testing.T) {
	valid := fixtureJSON(t)
	tests := []struct {
		name string
		json string
		want string
	}{
		{"release", strings.Replace(valid, "v0.1.0", "v01.0.0", 1), "invalid release"},
		{"revision", strings.Replace(valid, "\"identityRevision\": 11", "\"identityRevision\": 12", 1), "unsupported identityRevision"},
		{"namespace", strings.Replace(valid, "\"namespace\": \"ui\"", "\"namespace\": \"UI\"", 1), "invalid namespace"},
		{"product", strings.Replace(valid, "\"product\": \"heroicons\"", "\"product\": \"Heroicons\"", 1), "invalid product"},
		{"canonical name", strings.Replace(valid, "ui-hi-16-solid-arrow-down", "UI-hi-16-solid-arrow-down", 1), "invalid canonicalName"},
		{"path", strings.Replace(valid, "icons/ui/heroicons/16-solid-arrow-down.svg", "dist/icons/ui/heroicons/16-solid-arrow-down.svg", 1), "invalid path"},
		{"checksum", strings.Replace(valid, "d211c881861313937a6189cdb711f04e4c8c68518b6e26979811d0e863844a3e", "D211c881861313937a6189cdb711f04e4c8c68518b6e26979811d0e863844a3e", 1), "invalid sha256"},
		{"provenance", strings.Replace(valid, "heroicons@v2.2.0", "heroicons\\nsource", 1), "source is empty or invalid"},
		{"viewBox", strings.Replace(valid, "0 0 16 16", "0 0 0 16", 1), "invalid viewBox"},
		{"dimensions pair", strings.Replace(valid, "\"viewBox\": \"0 0 16 16\"", "\"width\": 16, \"viewBox\": \"0 0 16 16\"", 1), "width and height"},
		{"png dimensions", strings.NewReplacer("\"format\": \"svg\"", "\"format\": \"png\"", "16-solid-arrow-down.svg", "16-solid-arrow-down.png").Replace(valid), "PNG requires width and height"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(strings.NewReader(tt.json))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Load() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestLoadRejectsLeadingDigitInLowerKebabFields(t *testing.T) {
	valid := fixtureJSON(t)
	tests := []struct {
		name string
		old  string
		new  string
		want string
	}{
		{"canonical name", "ui-hi-16-solid-arrow-down", "1ui-hi-16-solid-arrow-down", "invalid canonicalName"},
		{"product", "\"product\": \"heroicons\"", "\"product\": \"1heroicons\"", "invalid product"},
		{"artwork", "\"artwork\": \"icon\"", "\"artwork\": \"1icon\"", "invalid artwork"},
		{"appearance", "\"appearance\": \"default\"", "\"appearance\": \"1default\"", "invalid appearance"},
		{"surface", "\"surface\": \"transparent\"", "\"surface\": \"1transparent\"", "invalid surface"},
		{"framing", "\"framing\": \"optical\"", "\"framing\": \"1optical\"", "invalid framing"},
		{"sprite symbol", "\"spriteSymbol\": \"hi-16-solid-arrow-down\"", "\"spriteSymbol\": \"1hi-16-solid-arrow-down\"", "invalid spriteSymbol"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.Replace(valid, tt.old, tt.new, 1)
			_, err := Load(strings.NewReader(input))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Load() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestLoadRejectsInvalidProvenanceLabels(t *testing.T) {
	valid := fixtureJSON(t)
	tests := []struct {
		name string
		old  string
		new  string
		want string
	}{
		{"license leading space", "\"license\": \"MIT\"", "\"license\": \" MIT\"", "license"},
		{"license trailing space", "\"license\": \"MIT\"", "\"license\": \"MIT \"", "license"},
		{"source tab", "heroicons@v2.2.0", "heroicons\\tsource", "source"},
		{"source newline", "heroicons@v2.2.0", "heroicons\\nsource", "source"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.Replace(valid, tt.old, tt.new, 1)
			_, err := Load(strings.NewReader(input))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Load() error = %v, want invalid %s", err, tt.want)
			}
		})
	}
}

func TestValidTextRejectsEveryControlClassAndAcceptsUnicode(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"NUL", "source\x00label"},
		{"vertical tab", "source\x0blabel"},
		{"form feed", "source\x0clabel"},
		{"escape", "source\x1blabel"},
		{"C1 next line", "source\u0085label"},
		{"C1 application program command", "source\u009flabel"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if validText(tt.value) {
				t.Fatalf("validText(%q) = true, want control rejected", tt.value)
			}
		})
	}
	if value := "Licença Heroicons — 日本語"; !validText(value) {
		t.Fatalf("validText(%q) = false, want valid Unicode label accepted", value)
	}
}
