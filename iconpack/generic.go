package iconpack

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/araihu/goshtoso/internal/iconcatalog"
	"go.yaml.in/yaml/v3"
)

const genericManifestSchemaVersion = 1

// sourceManifest describes a consumer-owned icon pack. It deliberately keeps
// the source contract smaller than the Arai Hu Assets catalog: each selected
// SVG is named, hashed, attributed, and either read as a standalone SVG or
// extracted from one declared symbol sprite.
type sourceManifest struct {
	SchemaVersion int                  `json:"schemaVersion" yaml:"schemaVersion"`
	Name          string               `json:"name" yaml:"name"`
	Release       string               `json:"release" yaml:"release"`
	Source        string               `json:"source" yaml:"source"`
	License       string               `json:"license" yaml:"license"`
	LicensePath   string               `json:"licensePath" yaml:"licensePath"`
	LicenseSHA256 string               `json:"licenseSha256" yaml:"licenseSha256"`
	NoticePath    string               `json:"noticePath,omitempty" yaml:"noticePath,omitempty"`
	NoticeSHA256  string               `json:"noticeSha256,omitempty" yaml:"noticeSha256,omitempty"`
	SpritePath    string               `json:"spritePath,omitempty" yaml:"spritePath,omitempty"`
	SpriteSHA256  string               `json:"spriteSha256,omitempty" yaml:"spriteSha256,omitempty"`
	Icons         []sourceManifestIcon `json:"icons" yaml:"icons"`
}

type sourceManifestIcon struct {
	CanonicalName string `json:"canonicalName" yaml:"canonicalName"`
	Namespace     string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Product       string `json:"product,omitempty" yaml:"product,omitempty"`
	Path          string `json:"path" yaml:"path"`
	SpriteSymbol  string `json:"spriteSymbol" yaml:"spriteSymbol"`
	ViewBox       string `json:"viewBox" yaml:"viewBox"`
	ColorBehavior string `json:"colorBehavior,omitempty" yaml:"colorBehavior,omitempty"`
	License       string `json:"license,omitempty" yaml:"license,omitempty"`
	Source        string `json:"source,omitempty" yaml:"source,omitempty"`
	SHA256        string `json:"sha256" yaml:"sha256"`
}

type genericPack struct {
	manifest      sourceManifest
	manifestBytes []byte
	manifestHash  string
	fileName      string
	family        sourceFamily
}

func openGenericPack(opts Options) (releaseBoundary, error) {
	manifestBytes, err := readExternalManifest(opts.SourceManifest)
	if err != nil {
		return releaseBoundary{}, fmt.Errorf("read source manifest: %w", err)
	}
	manifest, err := decodeSourceManifest(opts.SourceManifest, manifestBytes)
	if err != nil {
		return releaseBoundary{}, err
	}

	root := opts.SourceRoot
	cleanup := func() {}
	sourceKind := "source-root"
	archiveSHA256 := ""
	if opts.SourceArchive != "" {
		if !digestRE.MatchString(opts.SourceArchiveSHA256) {
			return releaseBoundary{}, fmt.Errorf("source-archive-sha256 must be a lowercase SHA-256")
		}
		sourceKind = "source-archive"
		archiveSHA256 = opts.SourceArchiveSHA256
		var temporary string
		err := withVerifiedReleaseArchive(opts.SourceArchive, opts.SourceArchiveSHA256, nil, func(archive *os.File, size int64) error {
			var err error
			temporary, err = os.MkdirTemp("", "goshtoso-iconpack-source-*")
			if err != nil {
				return fmt.Errorf("create source extraction root: %w", err)
			}
			cleanup = func() { _ = os.RemoveAll(temporary) }
			return extractOpenedArchive(archive, opts.SourceArchive, size, temporary)
		})
		if err != nil {
			cleanup()
			return releaseBoundary{}, fmt.Errorf("verify source archive: %w", err)
		}
		root = temporary
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		cleanup()
		return releaseBoundary{}, fmt.Errorf("resolve source root: %w", err)
	}
	openedRoot, err := os.OpenRoot(absRoot)
	if err != nil {
		cleanup()
		return releaseBoundary{}, fmt.Errorf("open source root: %w", err)
	}
	defer func() { _ = openedRoot.Close() }()
	if err := validateGenericRoot(absRoot, openedRoot); err != nil {
		cleanup()
		return releaseBoundary{}, err
	}

	manifestHash := hashBytes(manifestBytes)
	files, checksums, err := verifyGenericFiles(openedRoot, manifest)
	if err != nil {
		cleanup()
		return releaseBoundary{}, err
	}
	assets := make([]iconcatalog.Asset, 0, len(manifest.Icons))
	for _, icon := range manifest.Icons {
		namespace := icon.Namespace
		if namespace == "" {
			namespace = "custom"
		}
		product := icon.Product
		if product == "" {
			product = manifest.Name
		}
		license := icon.License
		if license == "" {
			license = manifest.License
		}
		source := icon.Source
		if source == "" {
			source = manifest.Source + ":" + icon.Path
		}
		colorBehavior := icon.ColorBehavior
		if colorBehavior == "" {
			colorBehavior = "unspecified"
		}
		assets = append(assets, iconcatalog.Asset{
			CanonicalName: icon.CanonicalName,
			Namespace:     namespace,
			Path:          icon.Path,
			Product:       product,
			Artwork:       "icon",
			Appearance:    "default",
			Surface:       "transparent",
			Framing:       "optical",
			Format:        "svg",
			Dimensions:    iconcatalog.Dimensions{ViewBox: icon.ViewBox},
			SpriteSymbol:  icon.SpriteSymbol,
			ColorBehavior: colorBehavior,
			License:       license,
			Source:        source,
			SHA256:        icon.SHA256,
		})
	}
	family := sourceFamily{
		Namespace:     "custom",
		Product:       manifest.Name,
		SpritePath:    manifest.SpritePath,
		LicensePath:   manifest.LicensePath,
		LicenseOutput: "LICENSES/" + sourcePackFileName(manifest.Name) + ".txt",
		Generic:       true,
	}
	return releaseBoundary{
		root:          absRoot,
		cleanup:       cleanup,
		sourceKind:    sourceKind,
		archiveSHA256: archiveSHA256,
		catalog: iconcatalog.Catalog{
			SchemaVersion:    genericManifestSchemaVersion,
			Release:          manifest.Release,
			IdentityRevision: 1,
			Assets:           assets,
			Hash:             manifestHash,
		},
		checksums: checksums,
		files:     files,
		generic: &genericPack{
			manifest:      manifest,
			manifestBytes: append([]byte(nil), manifestBytes...),
			manifestHash:  manifestHash,
			fileName:      sourcePackFileName(manifest.Name),
			family:        family,
		},
	}, nil
}

func readExternalManifest(filename string) ([]byte, error) {
	file, err := openRegularFile(filename)
	if err != nil {
		return nil, err
	}
	contents, readErr := io.ReadAll(io.LimitReader(file, maxMemberBytes+1))
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if int64(len(contents)) > maxMemberBytes {
		return nil, fmt.Errorf("source manifest exceeds %d-byte limit", maxMemberBytes)
	}
	return contents, nil
}

func decodeSourceManifest(filename string, contents []byte) (sourceManifest, error) {
	var manifest sourceManifest
	trimmed := bytes.TrimSpace(contents)
	if filepath.Ext(filename) == ".json" || len(trimmed) > 0 && trimmed[0] == '{' {
		decoder := json.NewDecoder(bytes.NewReader(contents))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&manifest); err != nil {
			return sourceManifest{}, fmt.Errorf("decode JSON source manifest: %w", err)
		}
		if err := requireJSONEOF(decoder, "source manifest"); err != nil {
			return sourceManifest{}, err
		}
	} else {
		decoder := yaml.NewDecoder(bytes.NewReader(contents))
		decoder.KnownFields(true)
		if err := decoder.Decode(&manifest); err != nil {
			return sourceManifest{}, fmt.Errorf("decode YAML source manifest: %w", err)
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			if err == nil {
				return sourceManifest{}, fmt.Errorf("source manifest contains more than one YAML document")
			}
			return sourceManifest{}, fmt.Errorf("decode YAML source manifest: %w", err)
		}
	}
	if err := validateSourceManifest(manifest); err != nil {
		return sourceManifest{}, err
	}
	return manifest, nil
}

func validateSourceManifest(manifest sourceManifest) error {
	if manifest.SchemaVersion != genericManifestSchemaVersion {
		return fmt.Errorf("unsupported source manifest schemaVersion %d: want %d", manifest.SchemaVersion, genericManifestSchemaVersion)
	}
	if err := validateManifestMetadata(manifest); err != nil {
		return err
	}
	if err := validateManifestPaths(manifest); err != nil {
		return err
	}
	return validateManifestIcons(manifest.Icons)
}

func validateManifestMetadata(manifest sourceManifest) error {
	for _, field := range []struct{ name, value string }{
		{"name", manifest.Name}, {"release", manifest.Release}, {"source", manifest.Source},
		{"license", manifest.License}, {"licensePath", manifest.LicensePath}, {"licenseSha256", manifest.LicenseSHA256},
	} {
		if strings.TrimSpace(field.value) == "" || field.value != strings.TrimSpace(field.value) || strings.ContainsAny(field.value, "\r\n\x00") {
			return fmt.Errorf("source manifest %s is invalid", field.name)
		}
	}
	return nil
}

func validateManifestPaths(manifest sourceManifest) error {
	for _, relative := range []struct{ name, value string }{
		{"licensePath", manifest.LicensePath}, {"noticePath", manifest.NoticePath}, {"spritePath", manifest.SpritePath},
	} {
		if relative.value != "" {
			if err := safeRelativePath(relative.value); err != nil {
				return fmt.Errorf("source manifest %s: %w", relative.name, err)
			}
		}
	}
	if manifest.SpritePath != "" && !digestRE.MatchString(manifest.SpriteSHA256) {
		return fmt.Errorf("source manifest spriteSha256 must be a lowercase SHA-256 when spritePath is set")
	}
	if manifest.NoticePath != "" && !digestRE.MatchString(manifest.NoticeSHA256) {
		return fmt.Errorf("source manifest noticeSha256 must be a lowercase SHA-256 when noticePath is set")
	}
	return nil
}

func validateManifestIcons(icons []sourceManifestIcon) error {
	if len(icons) == 0 || len(icons) > maxArchiveFiles {
		return fmt.Errorf("source manifest icons must contain between 1 and %d entries", maxArchiveFiles)
	}
	seenNames := make(map[string]struct{}, len(icons))
	seenSymbols := make(map[string]string, len(icons))
	for index, icon := range icons {
		if err := validateManifestIcon(index, icon); err != nil {
			return err
		}
		if _, exists := seenNames[icon.CanonicalName]; exists {
			return fmt.Errorf("source manifest has duplicate canonicalName %q", icon.CanonicalName)
		}
		seenNames[icon.CanonicalName] = struct{}{}
		if first, exists := seenSymbols[icon.SpriteSymbol]; exists {
			return fmt.Errorf("source manifest has duplicate spriteSymbol %q for %q and %q", icon.SpriteSymbol, first, icon.CanonicalName)
		}
		seenSymbols[icon.SpriteSymbol] = icon.CanonicalName
	}
	return nil
}

func validateManifestIcon(index int, icon sourceManifestIcon) error {
	for _, field := range []struct{ name, value string }{
		{"canonicalName", icon.CanonicalName}, {"path", icon.Path}, {"spriteSymbol", icon.SpriteSymbol}, {"viewBox", icon.ViewBox}, {"sha256", icon.SHA256},
	} {
		if strings.TrimSpace(field.value) == "" || field.value != strings.TrimSpace(field.value) || strings.ContainsAny(field.value, "\r\n\x00") {
			return fmt.Errorf("source manifest icon %d %s is invalid", index, field.name)
		}
	}
	if err := safeRelativePath(icon.Path); err != nil {
		return fmt.Errorf("source manifest icon %d: %w", index, err)
	}
	if !digestRE.MatchString(icon.SHA256) {
		return fmt.Errorf("source manifest icon %d sha256 must be a lowercase SHA-256", index)
	}
	if !validGenericSymbol(icon.SpriteSymbol) {
		return fmt.Errorf("source manifest icon %d spriteSymbol %q is not a lower-kebab SVG symbol", index, icon.SpriteSymbol)
	}
	if err := validateViewBox(icon.ViewBox); err != nil {
		return fmt.Errorf("source manifest icon %d: %w", index, err)
	}
	return nil
}

func verifyGenericFiles(root *os.Root, manifest sourceManifest) (map[string][]byte, map[string]string, error) {
	files := make(map[string][]byte, len(manifest.Icons)+3)
	checksums := make(map[string]string, len(manifest.Icons)+3)
	readExpected := func(relative, expected, label string) error {
		contents, err := readContainedRegularFile(root, relative)
		if err != nil {
			return fmt.Errorf("verify %s %q: %w", label, relative, err)
		}
		if got := hashBytes(contents); got != expected {
			return fmt.Errorf("%s %q SHA-256 mismatch: got %s, want %s", label, relative, got, expected)
		}
		if prior, exists := checksums[relative]; exists && prior != expected {
			return fmt.Errorf("source manifest gives conflicting SHA-256 values for %q", relative)
		}
		files[relative] = contents
		checksums[relative] = expected
		return nil
	}
	if err := readExpected(manifest.LicensePath, manifest.LicenseSHA256, "license"); err != nil {
		return nil, nil, err
	}
	if manifest.NoticePath != "" {
		if err := readExpected(manifest.NoticePath, manifest.NoticeSHA256, "notice"); err != nil {
			return nil, nil, err
		}
	}
	if manifest.SpritePath != "" {
		if err := readExpected(manifest.SpritePath, manifest.SpriteSHA256, "sprite"); err != nil {
			return nil, nil, err
		}
	}
	for _, icon := range manifest.Icons {
		if err := readExpected(icon.Path, icon.SHA256, "icon"); err != nil {
			return nil, nil, err
		}
	}
	return files, checksums, nil
}

func validateGenericRoot(name string, root *os.Root) error {
	pathInfo, err := os.Lstat(name)
	if err != nil {
		return fmt.Errorf("inspect source root: %w", err)
	}
	if !pathInfo.IsDir() || pathInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("source root must be a real directory")
	}
	openedInfo, err := root.Stat(".")
	if err != nil {
		return fmt.Errorf("inspect opened source root: %w", err)
	}
	if !os.SameFile(pathInfo, openedInfo) {
		return fmt.Errorf("source root changed while opening")
	}
	return nil
}

func sourcePackFileName(name string) string {
	var output strings.Builder
	separator := false
	for _, character := range strings.ToLower(name) {
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' {
			if separator && output.Len() > 0 {
				output.WriteByte('-')
			}
			output.WriteRune(character)
			separator = false
			continue
		}
		separator = true
	}
	result := strings.Trim(output.String(), "-")
	if result == "" {
		return "iconpack"
	}
	return result
}

func validGenericSymbol(value string) bool {
	if value == "" {
		return false
	}
	for index, character := range value {
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' {
			continue
		}
		if character == '-' && index > 0 && index < len(value)-1 {
			continue
		}
		return false
	}
	return true
}

func validateViewBox(value string) error {
	parts := strings.Fields(value)
	if len(parts) != 4 {
		return fmt.Errorf("viewBox %q must contain four numbers", value)
	}
	for index, part := range parts {
		number, err := strconv.ParseFloat(part, 64)
		if err != nil || math.IsNaN(number) || math.IsInf(number, 0) || index >= 2 && number <= 0 {
			return fmt.Errorf("viewBox %q contains invalid geometry", value)
		}
	}
	return nil
}

func standaloneSVGSymbol(raw []byte, relative, symbol, expectedViewBox string) ([]byte, error) {
	decoder := xml.NewDecoder(bytes.NewReader(raw))
	var content bytes.Buffer
	encoder := xml.NewEncoder(&content)
	state := standaloneSVGState{
		encoder:         encoder,
		relative:        relative,
		expectedViewBox: expectedViewBox,
	}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parse source SVG %q: %w", relative, err)
		}
		switch value := token.(type) {
		case xml.Directive, xml.ProcInst:
			return nil, fmt.Errorf("source SVG %q contains forbidden XML directive", relative)
		case xml.StartElement:
			if err := state.start(value); err != nil {
				return nil, err
			}
		case xml.EndElement:
			if err := state.end(value); err != nil {
				return nil, err
			}
		default:
			if state.depth > 0 {
				if err := encoder.EncodeToken(token); err != nil {
					return nil, err
				}
			}
		}
	}
	if !state.rootSeen || state.depth != 0 {
		return nil, fmt.Errorf("source SVG %q has incomplete root", relative)
	}
	if err := encoder.Flush(); err != nil {
		return nil, fmt.Errorf("encode source SVG %q: %w", relative, err)
	}
	if !validGenericSymbol(symbol) {
		return nil, fmt.Errorf("source SVG %q has invalid output symbol %q", relative, symbol)
	}
	var output bytes.Buffer
	symbolEncoder := xml.NewEncoder(&output)
	start := xml.StartElement{
		Name: xml.Name{Local: "symbol"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "id"}, Value: symbol},
			{Name: xml.Name{Local: "viewBox"}, Value: expectedViewBox},
		},
	}
	start.Attr = append(start.Attr, state.rootPaintAttributes...)
	if err := symbolEncoder.EncodeToken(start); err != nil {
		return nil, fmt.Errorf("encode standalone symbol %q: %w", symbol, err)
	}
	if err := symbolEncoder.Flush(); err != nil {
		return nil, fmt.Errorf("flush standalone symbol %q: %w", symbol, err)
	}
	output.Write(content.Bytes())
	output.WriteString(`</symbol>`)
	return output.Bytes(), nil
}

type standaloneSVGState struct {
	encoder             *xml.Encoder
	relative            string
	expectedViewBox     string
	depth               int
	rootSeen            bool
	rootPaintAttributes []xml.Attr
}

func (state *standaloneSVGState) start(value xml.StartElement) error {
	if err := validateSVGAttributes(value); err != nil {
		return fmt.Errorf("source SVG %q: %w", state.relative, err)
	}
	if state.depth == 0 {
		return state.startRoot(value)
	}
	if forbiddenSVGElement(value.Name.Local) {
		return fmt.Errorf("source SVG %q contains forbidden element %q", state.relative, value.Name.Local)
	}
	state.depth++
	if err := state.encoder.EncodeToken(value); err != nil {
		return err
	}
	return nil
}

func (state *standaloneSVGState) startRoot(value xml.StartElement) error {
	if state.rootSeen || value.Name.Local != "svg" || value.Name.Space != "http://www.w3.org/2000/svg" {
		return fmt.Errorf("source SVG %q has invalid root element", state.relative)
	}
	state.rootSeen = true
	rootViewBox := ""
	for _, attr := range value.Attr {
		if attr.Name.Space == "" && attr.Name.Local == "viewBox" {
			rootViewBox = attr.Value
		}
	}
	if rootViewBox != state.expectedViewBox {
		return fmt.Errorf("source SVG %q viewBox %q does not match manifest %q", state.relative, rootViewBox, state.expectedViewBox)
	}
	for _, attr := range value.Attr {
		if attr.Name.Space == "" && isSVGPresentationAttribute(attr.Name.Local) {
			state.rootPaintAttributes = append(state.rootPaintAttributes, attr)
		}
	}
	state.depth = 1
	return nil
}

func isSVGPresentationAttribute(name string) bool {
	switch strings.ToLower(name) {
	case "color", "fill", "fill-opacity", "fill-rule", "paint-order", "stroke", "stroke-dasharray", "stroke-dashoffset", "stroke-linecap", "stroke-linejoin", "stroke-miterlimit", "stroke-opacity", "stroke-width":
		return true
	default:
		return false
	}
}

func (state *standaloneSVGState) end(value xml.EndElement) error {
	if state.depth == 0 {
		return fmt.Errorf("source SVG %q has unbalanced XML", state.relative)
	}
	state.depth--
	if state.depth > 0 {
		if err := state.encoder.EncodeToken(value); err != nil {
			return err
		}
	}
	return nil
}

func forbiddenSVGElement(name string) bool {
	switch strings.ToLower(name) {
	case "script", "foreignobject", "iframe", "object", "embed", "image":
		return true
	default:
		return false
	}
}

func validateSVGAttributes(start xml.StartElement) error {
	for _, attr := range start.Attr {
		name := strings.ToLower(attr.Name.Local)
		if strings.HasPrefix(name, "on") {
			return fmt.Errorf("SVG event attribute %q is forbidden", attr.Name.Local)
		}
		if name == "href" || name == "src" {
			if !strings.HasPrefix(attr.Value, "#") {
				return fmt.Errorf("external SVG reference %q is forbidden", attr.Value)
			}
		}
	}
	return nil
}
