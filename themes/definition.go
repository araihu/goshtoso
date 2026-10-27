package themes

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"regexp"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
)

const (
	// SchemaVersion identifies the complete, literal snapshot wire contract.
	SchemaVersion = 1
	// MaxDefinitionBytes bounds an encoded theme definition.
	MaxDefinitionBytes = 64 << 10
)

// ErrInvalid identifies an unsupported or malformed theme definition.
var ErrInvalid = errors.New("invalid theme definition")

// Color is an explicit sRGB or OKLCH color, never a mutable palette reference.
type Color struct {
	Space    string     `json:"space"`
	Channels [3]float64 `json:"channels"`
	Alpha    float64    `json:"alpha"`
}

// FontRef identifies a consumer-resolved font. It never contains a URL or CSS.
type FontRef struct {
	Source string `json:"source"`
	ID     string `json:"id"`
}

// Typography holds independent title and body font references.
type Typography struct {
	Title FontRef `json:"title"`
	Body  FontRef `json:"body"`
}

// Provenance describes a definition's origin without establishing inheritance.
type Provenance struct {
	PresetKey       Key    `json:"presetKey"`
	GoshtosoVersion string `json:"goshtosoVersion"`
	CatalogRevision string `json:"catalogRevision"`
}

// Definition is a complete versioned theme snapshot. Its maps are caller-owned.
type Definition struct {
	SchemaVersion int              `json:"schemaVersion"`
	Tokens        map[string]Color `json:"tokens"`
	Typography    Typography       `json:"typography"`
	RadiusRem     float64          `json:"radiusRem"`
	Provenance    Provenance       `json:"provenance"`
}

// TokenSpec identifies a supported theme role and its editing behavior.
type TokenSpec struct {
	Key        string `json:"key"`
	Label      string `json:"label"`
	Editable   bool   `json:"editable"`
	Expression string `json:"expression,omitempty"`
}

// Overrides is an explicit edit to a materialized definition.
type Overrides struct {
	Tokens     map[string]Color
	Typography *Typography
	RadiusRem  *float64
}

type sourceCatalog struct {
	MaterializerDigest string                    `json:"materializerDigest"`
	Defaults           map[string]string         `json:"defaults"`
	Overrides          map[Key]map[string]string `json:"overrides"`
	Palette            map[string]string         `json:"palette"`
}

//go:embed source_catalog.json
var catalogData []byte
var source = loadSource()
var identifier = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9:._/-]{0,199}$`)

func loadSource() sourceCatalog {
	var v sourceCatalog
	if err := json.Unmarshal(catalogData, &v); err != nil {
		panic(err)
	}
	return v
}

// CatalogRevision identifies the source and palette used to materialize presets.
func CatalogRevision() string { sum := sha256.Sum256(catalogData); return hex.EncodeToString(sum[:]) }

// Clone returns a definition with caller-owned token storage.
func (d Definition) Clone() Definition { d.Tokens = maps.Clone(d.Tokens); return d }

// Decode accepts only one bounded definition with known fields.
func Decode(data []byte) (Definition, error) {
	var d Definition
	if len(data) > MaxDefinitionBytes {
		return d, fmt.Errorf("%w: definition size", ErrInvalid)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&d); err != nil {
		return d, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if dec.Decode(new(any)) != io.EOF {
		return d, fmt.Errorf("%w: trailing data", ErrInvalid)
	}
	return d, Validate(d)
}

// Tokens returns every supported role in deterministic order.
func Tokens() []TokenSpec {
	keys := []string{}
	for key := range source.Defaults {
		if after, ok := strings.CutPrefix(key, "--color-"); ok {
			keys = append(keys, after)
		}
	}
	for _, values := range source.Overrides {
		for key := range values {
			if after, ok := strings.CutPrefix(key, "--color-"); ok {
				keys = append(keys, after)
			}
		}
	}
	slices.Sort(keys)
	keys = slices.Compact(keys)
	result := make([]TokenSpec, 0, len(keys))
	for _, key := range keys {
		expr := source.Defaults["--color-"+key]
		derived := strings.Contains(expr, "color-mix(") || strings.HasPrefix(expr, "var(--color-outline)") || strings.HasPrefix(expr, "var(--color-outline-dark)")
		result = append(result, TokenSpec{Key: key, Label: strings.ReplaceAll(key, "-", " "), Editable: !derived, Expression: expr})
	}
	return result
}

// ValidFontRef checks the closed source/identifier grammar, not catalog existence.
func ValidFontRef(ref FontRef) bool {
	return slices.Contains([]string{"system", "google", "bundled", "upload"}, ref.Source) && identifier.MatchString(ref.ID)
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// Valid reports whether a color is a bounded literal in the v1 grammar.
func (c Color) Valid() bool {
	if !finite(c.Alpha) || c.Alpha < 0 || c.Alpha > 1 {
		return false
	}
	for _, v := range c.Channels {
		if !finite(v) {
			return false
		}
	}
	switch c.Space {
	case "srgb":
		for _, v := range c.Channels {
			if v < 0 || v > 1 {
				return false
			}
		}
		return true
	case "oklch":
		return c.Channels[0] >= 0 && c.Channels[0] <= 1 && c.Channels[1] >= 0 && c.Channels[1] <= .5 && c.Channels[2] >= 0 && c.Channels[2] < 360
	}
	return false
}

// Validate checks completeness and literal values without consulting a base preset.
func Validate(d Definition) error {
	if d.SchemaVersion != SchemaVersion || !finite(d.RadiusRem) || d.RadiusRem < 0 || d.RadiusRem > 4 || !ValidFontRef(d.Typography.Title) || !ValidFontRef(d.Typography.Body) {
		return ErrInvalid
	}
	specs := Tokens()
	if len(d.Tokens) != len(specs) || len(d.Tokens) > 256 {
		return fmt.Errorf("%w: incomplete tokens", ErrInvalid)
	}
	for _, spec := range specs {
		c, ok := d.Tokens[spec.Key]
		if !ok || !c.Valid() {
			return fmt.Errorf("%w: color %s", ErrInvalid, spec.Key)
		}
	}
	if len(d.Provenance.GoshtosoVersion) > 200 || len(d.Provenance.CatalogRevision) > 128 || len(d.Provenance.PresetKey) > 100 {
		return ErrInvalid
	}
	data, err := json.Marshal(d)
	if err != nil || len(data) > MaxDefinitionBytes {
		return ErrInvalid
	}
	return nil
}

// Preset materializes the installed release's preset into a complete snapshot.
func Preset(key Key) (Definition, error) {
	values, ok := source.Overrides[key]
	if !ok {
		return Definition{}, fmt.Errorf("%w: preset %s", ErrInvalid, key)
	}
	merged := maps.Clone(source.Palette)
	maps.Copy(merged, source.Defaults)
	maps.Copy(merged, values)
	d := Definition{SchemaVersion: SchemaVersion, Tokens: map[string]Color{}, Provenance: Provenance{PresetKey: key, GoshtosoVersion: Version(), CatalogRevision: CatalogRevision()}}
	for _, spec := range Tokens() {
		c, err := resolveColor(merged["--color-"+spec.Key], merged, map[string]bool{})
		if err != nil {
			return Definition{}, fmt.Errorf("preset %s %s: %w", key, spec.Key, err)
		}
		d.Tokens[spec.Key] = c
	}
	d.Typography = Typography{Title: stackRef(merged["--font-title"]), Body: stackRef(merged["--font-body"])}
	radius := merged["--radius-radius"]
	if after, ok0 := strings.CutPrefix(radius, "var("); ok0 {
		radius = merged[strings.TrimSuffix(after, ")")]
	}
	v, err := strconv.ParseFloat(strings.TrimSuffix(radius, "rem"), 64)
	if err != nil {
		return Definition{}, fmt.Errorf("radius %q: %w", radius, err)
	}
	d.RadiusRem = v
	return d, Validate(d)
}

// Materialize applies edits and recomputes default-derived roles into literal values.
// Preset overrides and imported literals that differ from the default relation stay frozen.
func Materialize(base Definition, edits Overrides) (Definition, error) {
	if err := Validate(base); err != nil {
		return Definition{}, err
	}
	d := base.Clone()
	allowed := map[string]TokenSpec{}
	for _, spec := range Tokens() {
		allowed[spec.Key] = spec
	}
	for key, c := range edits.Tokens {
		spec, ok := allowed[key]
		if !ok || !spec.Editable || !c.Valid() {
			return Definition{}, ErrInvalid
		}
		d.Tokens[key] = c
	}
	if edits.Typography != nil {
		d.Typography = *edits.Typography
	}
	if edits.RadiusRem != nil {
		d.RadiusRem = *edits.RadiusRem
	}
	if len(edits.Tokens) > 0 {
		derived := derivedEdits(base, edits.Tokens)
		vars := map[string]string{}
		for key, c := range d.Tokens {
			vars["--color-"+key] = c.CSS()
		}
		for _, spec := range Tokens() {
			if !derived[spec.Key] {
				continue
			}
			vars["--color-"+spec.Key] = spec.Expression
		}
		for _, spec := range Tokens() {
			if !derived[spec.Key] {
				continue
			}
			c, err := resolveColor(vars["--color-"+spec.Key], vars, map[string]bool{})
			if err != nil {
				return Definition{}, err
			}
			d.Tokens[spec.Key] = c
		}
	}
	return d, Validate(d)
}

func derivedEdits(base Definition, edits map[string]Color) map[string]bool {
	original := map[string]string{}
	for key, c := range base.Tokens {
		original["--color-"+key] = c.CSS()
	}
	derived := map[string]bool{}
	for _, spec := range Tokens() {
		if spec.Editable {
			continue
		}
		value, err := resolveColor(spec.Expression, original, map[string]bool{})
		derived[spec.Key] = err == nil && sameColor(value, base.Tokens[spec.Key])
	}
	affected := map[string]bool{}
	for key := range edits {
		affected[key] = true
	}
	for range derived {
		for _, spec := range Tokens() {
			if !derived[spec.Key] {
				continue
			}
			for key := range affected {
				if strings.Contains(spec.Expression, "var(--color-"+key+")") {
					affected[spec.Key] = true
				}
			}
		}
	}
	for key := range derived {
		derived[key] = derived[key] && affected[key]
	}
	return derived
}

func sameColor(a, b Color) bool {
	if a.Space != b.Space || math.Abs(a.Alpha-b.Alpha) > 1e-9 {
		return false
	}
	for i := range a.Channels {
		if math.Abs(a.Channels[i]-b.Channels[i]) > 1e-9 {
			return false
		}
	}
	return true
}

// UnmarshalJSON rejects extra or missing color channels and unknown color fields.
func (c *Color) UnmarshalJSON(data []byte) error {
	var wire struct {
		Space    string    `json:"space"`
		Channels []float64 `json:"channels"`
		Alpha    *float64  `json:"alpha"`
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&wire); err != nil {
		return err
	}
	if len(wire.Channels) != 3 || wire.Alpha == nil {
		return ErrInvalid
	}
	candidate := Color{Space: wire.Space, Channels: [3]float64(wire.Channels), Alpha: *wire.Alpha}
	if !candidate.Valid() {
		return ErrInvalid
	}
	*c = candidate
	return nil
}

// Version returns the compiled module version or (devel) for an unpublished checkout.
func Version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(devel)"
	}
	if info.Main.Path == "github.com/araihu/goshtoso" && info.Main.Version != "" {
		return info.Main.Version
	}
	for _, module := range info.Deps {
		if module.Path == "github.com/araihu/goshtoso" {
			if module.Replace != nil && module.Replace.Version == "" {
				return "(devel)"
			}
			if module.Version != "" {
				return module.Version
			}
		}
	}
	return "(devel)"
}

// UnmarshalJSON enforces the complete v1 wire shape, including zero-valued fields.
func (d *Definition) UnmarshalJSON(data []byte) error {
	if len(data) > MaxDefinitionBytes {
		return ErrInvalid
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, key := range []string{"schemaVersion", "tokens", "typography", "radiusRem", "provenance"} {
		if value, ok := fields[key]; !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return ErrInvalid
		}
	}
	var provenance map[string]json.RawMessage
	if err := json.Unmarshal(fields["provenance"], &provenance); err != nil {
		return err
	}
	for _, key := range []string{"presetKey", "goshtosoVersion", "catalogRevision"} {
		if value, ok := provenance[key]; !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return ErrInvalid
		}
	}
	type wire Definition
	var decoded wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	candidate := Definition(decoded)
	if err := Validate(candidate); err != nil {
		return err
	}
	*d = candidate
	return nil
}
