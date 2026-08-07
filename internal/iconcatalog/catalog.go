// Package iconcatalog loads supported asset catalogs and generates typed sprite
// bindings for a selected namespace and product.
package iconcatalog

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const (
	schemaVersionV1  = 1
	schemaVersionV2  = 2
	identityRevision = 11
)

var (
	lowerKebabRE        = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)
	lowerKebabVariantRE = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	mixedCaseKebabRE    = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*(?:-[A-Za-z0-9]+)*$`)
	semverRE            = regexp.MustCompile(`^v(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)(?:-(?:(?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*)?)?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)
	sha256RE            = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

var (
	topLevelKeys = keySet("schemaVersion", "release", "identityRevision", "assets")
	assetKeys    = keySet(
		"canonicalName", "namespace", "path", "product", "artwork", "appearance", "surface", "framing",
		"format", "dimensions", "spriteSymbol", "colorBehavior", "license", "source", "sha256",
	)
	dimensionsKeys = keySet("width", "height", "viewBox")
)

// Catalog is a supported assets catalog. Hash is the SHA-256 digest of the
// exact source bytes and is intentionally excluded from JSON.
type Catalog struct {
	SchemaVersion    int     `json:"schemaVersion"`
	Release          string  `json:"release"`
	IdentityRevision int     `json:"identityRevision"`
	Assets           []Asset `json:"assets"`
	Hash             string  `json:"-"`
}

// Asset is one entry from the assets schema-v1 catalog.
type Asset struct {
	CanonicalName string     `json:"canonicalName"`
	Namespace     string     `json:"namespace"`
	Path          string     `json:"path"`
	Product       string     `json:"product"`
	Artwork       string     `json:"artwork"`
	Appearance    string     `json:"appearance"`
	Surface       string     `json:"surface"`
	Framing       string     `json:"framing"`
	Format        string     `json:"format"`
	Dimensions    Dimensions `json:"dimensions"`
	SpriteSymbol  string     `json:"spriteSymbol"`
	ColorBehavior string     `json:"colorBehavior"`
	License       string     `json:"license"`
	Source        string     `json:"source"`
	SHA256        string     `json:"sha256"`
}

// Dimensions is the schema-v1 dimensions object. Width and Height are present
// together when an artifact has fixed raster dimensions.
type Dimensions struct {
	Width   int    `json:"width,omitempty"`
	Height  int    `json:"height,omitempty"`
	ViewBox string `json:"viewBox,omitempty"`
}

// Load decodes one exact schema-v1 or schema-v2 assets catalog and records its source hash.
// It rejects unknown, duplicate, and case-variant keys at every schema object.
func Load(r io.Reader) (Catalog, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return Catalog{}, fmt.Errorf("read catalog: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(b))
	fields, err := readObject(decoder, "top-level", topLevelKeys)
	if err != nil {
		return Catalog{}, err
	}
	if err := requireKeys(fields, "top-level", "schemaVersion"); err != nil {
		return Catalog{}, err
	}
	schema, err := intField(fields["schemaVersion"], "schemaVersion")
	if err != nil {
		return Catalog{}, err
	}
	if schema != schemaVersionV1 && schema != schemaVersionV2 {
		return Catalog{}, fmt.Errorf("unsupported schemaVersion %d: want %d or %d", schema, schemaVersionV1, schemaVersionV2)
	}
	if err := requireKeys(fields, "top-level", "release", "identityRevision", "assets"); err != nil {
		return Catalog{}, err
	}
	release, err := stringField(fields["release"], "release")
	if err != nil {
		return Catalog{}, err
	}
	revision, err := intField(fields["identityRevision"], "identityRevision")
	if err != nil {
		return Catalog{}, err
	}
	assets, err := decodeAssets(fields["assets"])
	if err != nil {
		return Catalog{}, err
	}
	if err := requireEOF(decoder); err != nil {
		return Catalog{}, err
	}
	catalog := Catalog{
		SchemaVersion: schema, Release: release, IdentityRevision: revision, Assets: assets,
	}
	if err := validateCatalog(catalog); err != nil {
		return Catalog{}, err
	}
	sum := sha256.Sum256(b)
	catalog.Hash = hex.EncodeToString(sum[:])
	return catalog, nil
}

func decodeAssets(raw json.RawMessage) ([]Asset, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("decode assets: %w", err)
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '[' {
		return nil, fmt.Errorf("assets must be an array")
	}
	assets := []Asset{}
	for decoder.More() {
		asset, err := decodeAsset(decoder)
		if err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}
	if token, err := decoder.Token(); err != nil {
		return nil, fmt.Errorf("decode assets: %w", err)
	} else if delimiter, ok := token.(json.Delim); !ok || delimiter != ']' {
		return nil, fmt.Errorf("assets must be an array")
	}
	if err := requireEOF(decoder); err != nil {
		return nil, err
	}
	return assets, nil
}

func decodeAsset(decoder *json.Decoder) (Asset, error) {
	fields, err := readObject(decoder, "asset", assetKeys)
	if err != nil {
		return Asset{}, err
	}
	if err := requireKeys(fields, "asset", "canonicalName", "namespace", "path", "product", "artwork", "appearance", "surface", "framing", "format", "dimensions", "spriteSymbol", "colorBehavior", "license", "source", "sha256"); err != nil {
		return Asset{}, err
	}
	asset := Asset{}
	for _, field := range []struct {
		name string
		into *string
	}{
		{"canonicalName", &asset.CanonicalName}, {"namespace", &asset.Namespace}, {"path", &asset.Path},
		{"product", &asset.Product}, {"artwork", &asset.Artwork}, {"appearance", &asset.Appearance},
		{"surface", &asset.Surface}, {"framing", &asset.Framing}, {"format", &asset.Format},
		{"spriteSymbol", &asset.SpriteSymbol}, {"colorBehavior", &asset.ColorBehavior}, {"license", &asset.License},
		{"source", &asset.Source}, {"sha256", &asset.SHA256},
	} {
		*field.into, err = stringField(fields[field.name], field.name)
		if err != nil {
			return Asset{}, err
		}
	}
	asset.Dimensions, err = decodeDimensions(fields["dimensions"])
	if err != nil {
		return Asset{}, err
	}
	return asset, nil
}

func decodeDimensions(raw json.RawMessage) (Dimensions, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	fields, err := readObject(decoder, "dimensions", dimensionsKeys)
	if err != nil {
		return Dimensions{}, err
	}
	if err := requireEOF(decoder); err != nil {
		return Dimensions{}, err
	}
	dimensions := Dimensions{}
	if raw, ok := fields["width"]; ok {
		dimensions.Width, err = intField(raw, "dimensions.width")
		if err != nil {
			return Dimensions{}, err
		}
	}
	if raw, ok := fields["height"]; ok {
		dimensions.Height, err = intField(raw, "dimensions.height")
		if err != nil {
			return Dimensions{}, err
		}
	}
	if raw, ok := fields["viewBox"]; ok {
		dimensions.ViewBox, err = stringField(raw, "dimensions.viewBox")
		if err != nil {
			return Dimensions{}, err
		}
	}
	return dimensions, nil
}

func validateCatalog(catalog Catalog) error {
	if !semverRE.MatchString(catalog.Release) {
		return fmt.Errorf("invalid release %q", catalog.Release)
	}
	if catalog.IdentityRevision != identityRevision {
		return fmt.Errorf("unsupported identityRevision %d: want %d", catalog.IdentityRevision, identityRevision)
	}
	if len(catalog.Assets) == 0 {
		return fmt.Errorf("assets must not be empty")
	}
	names := make(map[string]struct{}, len(catalog.Assets))
	symbols := make(map[string]struct{}, len(catalog.Assets))
	for index, asset := range catalog.Assets {
		if err := validateCatalogAsset(asset, catalog.SchemaVersion); err != nil {
			return fmt.Errorf("asset %d: %w", index, err)
		}
		if _, exists := names[asset.CanonicalName]; exists {
			return fmt.Errorf("duplicate canonicalName %q", asset.CanonicalName)
		}
		names[asset.CanonicalName] = struct{}{}
		if asset.SpriteSymbol != "" {
			if _, exists := symbols[asset.SpriteSymbol]; exists {
				return fmt.Errorf("duplicate spriteSymbol %q", asset.SpriteSymbol)
			}
			symbols[asset.SpriteSymbol] = struct{}{}
		}
		if index > 0 {
			previous := catalog.Assets[index-1]
			if asset.CanonicalName < previous.CanonicalName || asset.CanonicalName == previous.CanonicalName && asset.Path < previous.Path {
				return fmt.Errorf("assets are not sorted by canonicalName then path")
			}
		}
	}
	return nil
}

func validateCatalogAsset(asset Asset, schema int) error {
	if asset.CanonicalName == "" {
		return fmt.Errorf("empty canonicalName")
	}
	canonicalPattern := lowerKebabRE
	if schema == schemaVersionV2 {
		canonicalPattern = mixedCaseKebabRE
	}
	if !canonicalPattern.MatchString(asset.CanonicalName) {
		return fmt.Errorf("invalid canonicalName %q", asset.CanonicalName)
	}
	if asset.Namespace != "brand" && asset.Namespace != "ui" {
		return fmt.Errorf("invalid namespace %q", asset.Namespace)
	}
	for _, field := range []struct {
		name, value string
	}{
		{"product", asset.Product}, {"artwork", asset.Artwork},
		{"surface", asset.Surface}, {"framing", asset.Framing},
	} {
		if !lowerKebabRE.MatchString(field.value) {
			return fmt.Errorf("invalid %s %q", field.name, field.value)
		}
	}
	appearancePattern := lowerKebabRE
	if schema == schemaVersionV2 {
		appearancePattern = lowerKebabVariantRE
	}
	if !appearancePattern.MatchString(asset.Appearance) {
		return fmt.Errorf("invalid appearance %q", asset.Appearance)
	}
	if err := validatePath(asset.Path, asset.Format); err != nil {
		return err
	}
	if asset.ColorBehavior != "protected" && asset.ColorBehavior != "monochrome" && asset.ColorBehavior != "tintable" {
		return fmt.Errorf("invalid colorBehavior %q", asset.ColorBehavior)
	}
	if asset.SpriteSymbol != "" && !lowerKebabRE.MatchString(asset.SpriteSymbol) {
		return fmt.Errorf("invalid spriteSymbol %q", asset.SpriteSymbol)
	}
	if err := validateDimensions(asset); err != nil {
		return err
	}
	for _, field := range []struct {
		name, value string
	}{{"license", asset.License}, {"source", asset.Source}} {
		if !validText(field.value) {
			return fmt.Errorf("%s is empty or invalid", field.name)
		}
	}
	if !sha256RE.MatchString(asset.SHA256) {
		return fmt.Errorf("invalid sha256 %q", asset.SHA256)
	}
	return nil
}

func validText(value string) bool {
	if strings.TrimSpace(value) != value || value == "" {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func validatePath(path, format string) error {
	if format != "svg" && format != "png" {
		return fmt.Errorf("invalid format %q", format)
	}
	if !strings.HasPrefix(path, "brand/") && !strings.HasPrefix(path, "icons/") && !strings.HasPrefix(path, "platform/") {
		return fmt.Errorf("invalid path %q", path)
	}
	if strings.HasPrefix(path, "/") || strings.Contains(path, "\\") || strings.Contains(path, "//") || strings.HasPrefix(path, "dist/") || !strings.HasSuffix(path, "."+format) {
		return fmt.Errorf("invalid path %q", path)
	}
	for _, segment := range strings.Split(path, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("invalid path %q", path)
		}
	}
	return nil
}

func validateDimensions(asset Asset) error {
	width, height := asset.Dimensions.Width, asset.Dimensions.Height
	if width == 0 && height != 0 || width != 0 && height == 0 || width < 0 || height < 0 {
		return fmt.Errorf("width and height must occur together as positive values")
	}
	switch asset.Format {
	case "svg":
		if asset.Dimensions.ViewBox == "" {
			return fmt.Errorf("SVG requires viewBox")
		}
		if !validViewBox(asset.Dimensions.ViewBox) {
			return fmt.Errorf("invalid viewBox %q", asset.Dimensions.ViewBox)
		}
	case "png":
		if width <= 0 || height <= 0 {
			return fmt.Errorf("PNG requires width and height")
		}
		if asset.Dimensions.ViewBox != "" {
			return fmt.Errorf("PNG must omit viewBox")
		}
		if asset.SpriteSymbol != "" {
			return fmt.Errorf("PNG must have empty spriteSymbol")
		}
	}
	return nil
}

func validViewBox(value string) bool {
	parts := strings.Fields(value)
	if len(parts) != 4 {
		return false
	}
	for index, part := range parts {
		number, err := strconv.ParseFloat(part, 64)
		if err != nil || math.IsNaN(number) || math.IsInf(number, 0) || index >= 2 && number <= 0 {
			return false
		}
	}
	return true
}

func readObject(decoder *json.Decoder, scope string, allowed map[string]struct{}) (map[string]json.RawMessage, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("decode %s object: %w", scope, err)
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '{' {
		return nil, fmt.Errorf("%s must be an object", scope)
	}
	fields := make(map[string]json.RawMessage, len(allowed))
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("decode %s object: %w", scope, err)
		}
		key, ok := token.(string)
		if !ok {
			return nil, fmt.Errorf("decode %s object key", scope)
		}
		if _, ok := allowed[key]; !ok {
			return nil, fmt.Errorf("unknown %s key %q", scope, key)
		}
		if _, exists := fields[key]; exists {
			return nil, fmt.Errorf("duplicate %s key %q", scope, key)
		}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return nil, fmt.Errorf("decode %s key %q: %w", scope, key, err)
		}
		fields[key] = raw
	}
	token, err = decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("decode %s object: %w", scope, err)
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '}' {
		return nil, fmt.Errorf("%s must be an object", scope)
	}
	return fields, nil
}

func requireKeys(fields map[string]json.RawMessage, scope string, keys ...string) error {
	for _, key := range keys {
		if _, ok := fields[key]; !ok {
			return fmt.Errorf("missing %s key %q", scope, key)
		}
	}
	return nil
}

func stringField(raw json.RawMessage, name string) (string, error) {
	if len(raw) == 0 || raw[0] != '"' {
		return "", fmt.Errorf("%s must be a string", name)
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("decode %s: %w", name, err)
	}
	return value, nil
}

func intField(raw json.RawMessage, name string) (int, error) {
	if len(raw) == 0 || raw[0] < '0' || raw[0] > '9' {
		return 0, fmt.Errorf("%s must be an integer", name)
	}
	var value int
	if err := json.Unmarshal(raw, &value); err != nil {
		return 0, fmt.Errorf("decode %s: %w", name, err)
	}
	return value, nil
}

func requireEOF(decoder *json.Decoder) error {
	if token, err := decoder.Token(); err != io.EOF {
		if err != nil {
			return fmt.Errorf("decode catalog: %w", err)
		}
		return fmt.Errorf("catalog contains more than one JSON value (%v)", token)
	}
	return nil
}

func keySet(keys ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		set[key] = struct{}{}
	}
	return set
}
