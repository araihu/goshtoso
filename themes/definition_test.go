package themes

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEveryPresetMaterializesIndependently(t *testing.T) {
	for _, entry := range BuiltIn() {
		t.Run(string(entry.Key), func(t *testing.T) {
			d, err := Preset(entry.Key)
			if err != nil {
				t.Fatal(err)
			}
			if err = Validate(d); err != nil {
				t.Fatal(err)
			}
			css, err := RenderCSS(d, nil)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(css, "var(") || strings.Contains(css, "color-mix(") {
				t.Fatal("snapshot leaked mutable references")
			}
			encoded, _ := json.Marshal(d)
			again, err := Decode(encoded)
			if err != nil {
				t.Fatal(err)
			}
			again.Tokens["primary"] = Color{Space: "srgb", Alpha: 1}
			fresh, _ := Preset(entry.Key)
			if fresh.Tokens["primary"] != d.Tokens["primary"] {
				t.Fatal("caller mutated preset")
			}
		})
	}
}
func TestDefinitionRejectsMalformedAndIncompleteValues(t *testing.T) {
	base, err := Preset(KeyGoshtoso)
	if err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*Definition){func(d *Definition) { d.SchemaVersion = 2 }, func(d *Definition) { delete(d.Tokens, "primary") }, func(d *Definition) { d.Tokens["bad"] = Color{} }, func(d *Definition) { d.RadiusRem = math.NaN() }, func(d *Definition) { d.Typography.Body.ID = "</style>" }, func(d *Definition) { d.Tokens["primary"] = Color{Space: "var", Alpha: 1} }} {
		d := base.Clone()
		mutate(&d)
		if Validate(d) == nil {
			t.Fatal("invalid accepted")
		}
	}
	data, _ := json.Marshal(base)
	data = append(data[:len(data)-1], []byte(`,"extra":true}`)...)
	if _, err := Decode(data); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
}
func TestMaterializeRecomputesDerivedValuesAndFreezesCopy(t *testing.T) {
	d, _ := Preset(KeyGoshtoso)
	before := d.Tokens["info-text"]
	changed, err := Materialize(d, Overrides{Tokens: map[string]Color{"on-surface-strong": {Space: "srgb", Channels: [3]float64{1, 0, 0}, Alpha: 1}}})
	if err != nil {
		t.Fatal(err)
	}
	if changed.Tokens["info-text"] == before {
		t.Fatal("dependent color was not updated")
	}
	if d.Tokens["info-text"] != before {
		t.Fatal("base was mutated")
	}
}
func TestFontResolverDoesNotAcceptURLsFromDefinition(t *testing.T) {
	d, _ := Preset(KeyGoshtoso)
	d.Typography.Body = FontRef{Source: "upload", ID: "font:example"}
	if _, err := RenderCSS(d, nil); err == nil {
		t.Fatal("unknown font accepted")
	}
	resolver := func(FontRef) (ResolvedFont, error) {
		return ResolvedFont{Families: []string{"safe"}, URL: "javascript:alert(1)", Format: "woff2", Weight: "400", Style: "normal"}, nil
	}
	if _, err := RenderCSS(d, resolver); err == nil {
		t.Fatal("unsafe provider URL accepted")
	}
}
func TestContrastBlackWhite(t *testing.T) {
	a, _ := ParseColor("#000000")
	b, _ := ParseColor("#ffffff")
	if math.Abs(Contrast(a, b)-21) > .0001 {
		t.Fatal(Contrast(a, b))
	}
}

func TestWireRequiresZeroValuedFieldsAndProvenance(t *testing.T) {
	d, _ := Preset(KeyGoshtoso)
	data, _ := json.Marshal(d)
	for _, field := range []string{"radiusRem", "provenance", "schemaVersion", "tokens", "typography"} {
		var fields map[string]json.RawMessage
		_ = json.Unmarshal(data, &fields)
		delete(fields, field)
		encoded, _ := json.Marshal(fields)
		if _, err := Decode(encoded); err == nil {
			t.Fatal("missing required field accepted", field)
		}
	}
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(data, &fields)
	fields["radiusRem"] = json.RawMessage("null")
	encoded, _ := json.Marshal(fields)
	if _, err := Decode(encoded); err == nil {
		t.Fatal("null radius accepted")
	}
	_ = json.Unmarshal(data, &fields)
	fields["provenance"] = json.RawMessage(`{"presetKey":"goshtoso","catalogRevision":"v1"}`)
	encoded, _ = json.Marshal(fields)
	if _, err := Decode(encoded); err == nil {
		t.Fatal("incomplete provenance accepted")
	}
}

func TestAssetFontUsesOpaqueAlias(t *testing.T) {
	ref := FontRef{Source: "upload", ID: "font:fixture"}
	css, family, err := RenderFontCSS(ref, func(FontRef) (ResolvedFont, error) {
		return ResolvedFont{Families: []string{"<embedded metadata>"}, URL: "/fonts/fixture", Format: "woff2", Weight: "400", Style: "normal"}, nil
	})
	if err != nil || strings.Contains(css+family, "embedded metadata") || !strings.Contains(family, "goshtoso-font-font-fixture") {
		t.Fatalf("alias: %q %q %v", css, family, err)
	}
	if _, err := ResolveFont([]FontOption{{Source: "google", ID: "empty"}}, FontRef{Source: "google", ID: "empty"}); err == nil {
		t.Fatal("missing Google family accepted")
	}
}

func TestMaterializePreservesUnrelatedPresetRoles(t *testing.T) {
	color, _ := ParseColor("#165d9b")
	specs := Tokens()
	for _, preset := range BuiltIn() {
		t.Run(string(preset.Key), func(t *testing.T) {
			base, _ := Preset(preset.Key)
			for _, edit := range specs {
				if !edit.Editable {
					continue
				}
				changed, err := Materialize(base, Overrides{Tokens: map[string]Color{edit.Key: color}})
				if err != nil {
					t.Fatal(err)
				}
				affected := map[string]bool{edit.Key: true}
				for range specs {
					for _, spec := range specs {
						if spec.Editable {
							continue
						}
						for key := range affected {
							if strings.Contains(spec.Expression, "var(--color-"+key+")") {
								affected[spec.Key] = true
							}
						}
					}
				}
				for key, before := range base.Tokens {
					if !affected[key] && !sameColor(before, changed.Tokens[key]) {
						t.Errorf("editing %s changed unrelated %s", edit.Key, key)
					}
				}
			}
		})
	}
	minimal, _ := Preset(KeyMinimal)
	changed, err := Materialize(minimal, Overrides{Tokens: map[string]Color{"outline": color, "outline-dark": color}})
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"control-outline", "control-outline-dark"} {
		if changed.Tokens[key] != minimal.Tokens[key] {
			t.Errorf("lost Minimal override %s", key)
		}
	}
	// Imported/customized derived literals remain snapshots even when their default inputs change.
	minimal.Tokens["info-text"] = color
	changed, err = Materialize(minimal, Overrides{Tokens: map[string]Color{"info": color}})
	if err != nil || changed.Tokens["info-text"] != color {
		t.Fatalf("lost custom literal: %v", err)
	}
}

func TestMaterializerCatalogDigestMatchesSource(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.New()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		hash.Write([]byte(file + "\x00"))
		hash.Write(data)
	}
	if got := hex.EncodeToString(hash.Sum(nil)); got != source.MaterializerDigest {
		t.Fatalf("materializer catalog is stale: have %s want %s; run go generate ./themes", source.MaterializerDigest, got)
	}
}
