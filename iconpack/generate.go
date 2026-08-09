package iconpack

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/format"
	"go/token"
	"sort"
	"strings"
	"unicode"

	"github.com/araihu/goshtoso/internal/iconurl"
)

type outputFile struct {
	Path   string `json:"path"`
	Mode   string `json:"mode"`
	Bytes  int    `json:"bytes"`
	SHA256 string `json:"sha256"`
}

type outputManifest struct {
	SchemaVersion        int             `json:"schemaVersion"`
	Tool                 string          `json:"tool"`
	Release              string          `json:"release"`
	IdentityRevision     int             `json:"identityRevision"`
	CatalogSchemaVersion int             `json:"catalogSchemaVersion"`
	CatalogSHA256        string          `json:"catalogSha256"`
	ReleaseJSONSHA256    string          `json:"releaseJsonSha256"`
	ChecksumsSHA256      string          `json:"checksumsSha256"`
	SourceKind           string          `json:"sourceKind"`
	ArchiveSHA256        string          `json:"archiveSha256"`
	SourceConfigSHA256   string          `json:"sourceConfigSha256,omitempty"`
	SourceLockSHA256     string          `json:"sourceLockSha256,omitempty"`
	Package              string          `json:"package"`
	ConstPrefix          string          `json:"constPrefix"`
	SpriteURL            string          `json:"spriteUrl"`
	Assets               []manifestAsset `json:"assets"`
	Files                []outputFile    `json:"files"`
}

type manifestAsset struct {
	CanonicalName string `json:"canonicalName"`
	Namespace     string `json:"namespace"`
	Product       string `json:"product"`
	Path          string `json:"path"`
	SpriteSymbol  string `json:"spriteSymbol"`
	ColorBehavior string `json:"colorBehavior"`
	SHA256        string `json:"sha256"`
	GoIdentifier  string `json:"goIdentifier"`
}

type provenanceDocument struct {
	SchemaVersion      int                `json:"schemaVersion"`
	Tool               string             `json:"tool"`
	Release            string             `json:"release"`
	CatalogSHA256      string             `json:"catalogSha256"`
	SourceKind         string             `json:"sourceKind"`
	ArchiveSHA256      string             `json:"archiveSha256"`
	SourceConfigSHA256 string             `json:"sourceConfigSha256,omitempty"`
	SourceLockSHA256   string             `json:"sourceLockSha256,omitempty"`
	Sources            []provenanceSource `json:"sources"`
	Assets             []provenanceAsset  `json:"assets"`
}

type provenanceSource struct {
	Namespace        string `json:"namespace"`
	Product          string `json:"product"`
	SpritePath       string `json:"spritePath"`
	SpriteSHA256     string `json:"spriteSha256"`
	ManifestPath     string `json:"manifestPath,omitempty"`
	ManifestSHA256   string `json:"manifestSha256,omitempty"`
	ProvenancePath   string `json:"provenancePath"`
	ProvenanceSHA256 string `json:"provenanceSha256"`
	LicensePath      string `json:"licensePath"`
	LicenseSHA256    string `json:"licenseSha256"`
}

type provenanceAsset struct {
	CanonicalName string `json:"canonicalName"`
	Source        string `json:"source"`
	License       string `json:"license"`
	Path          string `json:"path"`
	SHA256        string `json:"sha256"`
}

func generate(ctx context.Context, opts Options) (Result, error) {
	if err := validateGenerationOptions(opts); err != nil {
		return Result{}, err
	}
	names, err := selectionNames(opts)
	if err != nil {
		return Result{}, err
	}
	boundary, err := openSourceBoundary(ctx, opts)
	if err != nil {
		return Result{}, err
	}
	defer boundary.cleanup()
	if len(names) == 0 {
		for _, asset := range boundary.catalog.Assets {
			names = append(names, asset.CanonicalName)
		}
	}
	selected, families, err := selectAssets(boundary, names, opts.ConstPrefix)
	if err != nil {
		return Result{}, err
	}
	files, err := buildOutputs(boundary, opts, selected, families)
	if err != nil {
		return Result{}, err
	}
	published, outputDir, err := publishOutput(ctx, opts.OutputDir, files, opts.Check)
	if err != nil {
		return Result{}, err
	}
	return Result{
		OutputDir:     outputDir,
		Release:       boundary.catalog.Release,
		CatalogSHA256: boundary.catalog.Hash,
		SelectedCount: len(selected),
		Published:     published,
	}, nil
}

func validateGenerationOptions(opts Options) error {
	if err := validateOutputOptions(opts); err != nil {
		return err
	}
	return validateSourceOptions(opts)
}

func validateOutputOptions(opts Options) error {
	for _, field := range []struct{ name, value string }{
		{"out", opts.OutputDir}, {"package", opts.Package}, {"const-prefix", opts.ConstPrefix}, {"sprite-url", opts.SpriteURL},
	} {
		if field.value == "" {
			return fmt.Errorf("%s is required", field.name)
		}
	}
	if !validGoIdentifier(opts.Package) || opts.Package == "_" {
		return fmt.Errorf("package %q is not a valid Go package name", opts.Package)
	}
	if !validGoIdentifier(opts.ConstPrefix) || opts.ConstPrefix == "_" {
		return fmt.Errorf("const-prefix %q is not a valid Go identifier", opts.ConstPrefix)
	}
	if err := iconurl.ValidateSpriteURL(opts.SpriteURL); err != nil {
		return fmt.Errorf("sprite-url: %w", err)
	}
	return nil
}

func validateSourceOptions(opts Options) error {
	configSource := opts.ConfigPath != ""
	releaseSource := opts.ReleaseRoot != "" || opts.ReleaseArchive != ""
	manifestSource := opts.SourceRoot != "" || opts.SourceArchive != "" || opts.SourceManifest != ""
	if (releaseSource || manifestSource) && configSource {
		return fmt.Errorf("iconpack config and legacy source inputs are mutually exclusive")
	}
	if releaseSource && manifestSource {
		return fmt.Errorf("release source and generic source inputs are mutually exclusive")
	}
	if !configSource && !releaseSource && !manifestSource {
		return fmt.Errorf("exactly one of Arai Hu Assets release input or generic source input is required")
	}
	if configSource && !releaseSource && !manifestSource {
		return nil
	}
	return validateGenericSourceOptions(opts, manifestSource)
}

func validateGenericSourceOptions(opts Options, manifestSource bool) error {
	if manifestSource && opts.SourceManifest == "" {
		return fmt.Errorf("source-manifest is required with generic source input")
	}
	if manifestSource && (opts.SourceRoot == "") == (opts.SourceArchive == "") {
		return fmt.Errorf("exactly one of source-root or source-archive is required with source-manifest")
	}
	return nil
}

func buildOutputs(boundary releaseBoundary, opts Options, selected []selectedAsset, families []sourceFamily) (map[string][]byte, error) {
	files := map[string][]byte{}
	sprite, err := buildSprite(boundary, selected, families)
	if err != nil {
		return nil, err
	}
	files["sprite.svg"] = sprite
	bindings, err := generateBindings(boundary, opts, selected)
	if err != nil {
		return nil, err
	}
	files["icons_gen.go"] = bindings

	provenance := provenanceDocument{
		SchemaVersion: OutputSchemaVersion,
		Tool:          toolName,
		Release:       boundary.catalog.Release,
		CatalogSHA256: boundary.catalog.Hash,
		SourceKind:    boundary.sourceKind,
		ArchiveSHA256: boundary.archiveSHA256,
	}
	if err := addBoundaryOutputs(boundary, families, files, &provenance); err != nil {
		return nil, err
	}
	for _, asset := range selected {
		provenance.Assets = append(provenance.Assets, provenanceAsset{
			CanonicalName: asset.CanonicalName,
			Source:        asset.Source,
			License:       asset.License,
			Path:          asset.Path,
			SHA256:        asset.SHA256,
		})
	}
	provenanceBytes, err := marshalDocument(provenance)
	if err != nil {
		return nil, fmt.Errorf("encode provenance: %w", err)
	}
	files["provenance.json"] = provenanceBytes
	if boundary.generic == nil && boundary.muamba == nil {
		if notice, err := readVerifiedReleaseFile(boundary, "NOTICE"); err == nil {
			files["NOTICE"] = notice
		} else {
			return nil, err
		}
	}

	manifest := outputManifest{
		SchemaVersion:        OutputSchemaVersion,
		Tool:                 toolName,
		Release:              boundary.catalog.Release,
		IdentityRevision:     boundary.catalog.IdentityRevision,
		CatalogSchemaVersion: boundary.catalog.SchemaVersion,
		CatalogSHA256:        boundary.catalog.Hash,
		ReleaseJSONSHA256:    opts.ReleaseJSONSHA256,
		ChecksumsSHA256:      opts.ChecksumsSHA256,
		SourceKind:           boundary.sourceKind,
		ArchiveSHA256:        boundary.archiveSHA256,
		Package:              opts.Package,
		ConstPrefix:          opts.ConstPrefix,
		SpriteURL:            opts.SpriteURL,
	}
	if boundary.muamba != nil {
		manifest.SourceConfigSHA256 = boundary.muamba.configHash
		manifest.SourceLockSHA256 = boundary.muamba.lockHash
	}
	for _, asset := range selected {
		manifest.Assets = append(manifest.Assets, manifestAsset{
			CanonicalName: asset.CanonicalName,
			Namespace:     asset.Namespace,
			Product:       asset.Product,
			Path:          asset.Path,
			SpriteSymbol:  asset.SpriteSymbol,
			ColorBehavior: asset.ColorBehavior,
			SHA256:        asset.SHA256,
			GoIdentifier:  asset.goName,
		})
	}
	paths := make([]string, 0, len(files))
	for relative := range files {
		paths = append(paths, relative)
	}
	sort.Strings(paths)
	for _, relative := range paths {
		contents := files[relative]
		manifest.Files = append(manifest.Files, outputFile{
			Path: relative, Mode: "0644", Bytes: len(contents), SHA256: hashBytes(contents),
		})
	}
	manifestBytes, err := marshalDocument(manifest)
	if err != nil {
		return nil, fmt.Errorf("encode output manifest: %w", err)
	}
	files["manifest.json"] = manifestBytes
	return files, nil
}

func addBoundaryOutputs(boundary releaseBoundary, families []sourceFamily, files map[string][]byte, provenance *provenanceDocument) error {
	if boundary.muamba != nil {
		return addMuambaOutputs(boundary, families, files, provenance)
	}
	if boundary.generic != nil {
		return addGenericOutputs(boundary, files, provenance)
	}
	return addReleaseOutputs(boundary, families, files, provenance)
}

func addMuambaOutputs(boundary releaseBoundary, families []sourceFamily, files map[string][]byte, provenance *provenanceDocument) error {
	files["PROVENANCE/"+boundary.muamba.manifestPath] = append([]byte(nil), boundary.muamba.configBytes...)
	provenance.SourceConfigSHA256 = boundary.muamba.configHash
	provenance.SourceLockSHA256 = boundary.muamba.lockHash
	var notice bytes.Buffer
	for _, family := range families {
		licenseBytes, err := readVerifiedReleaseFile(boundary, family.LicensePath)
		if err != nil {
			return err
		}
		files[family.LicenseOutput] = licenseBytes
		sourceID, _, ok := strings.Cut(family.LicensePath, "/")
		if !ok {
			return fmt.Errorf("iconpack family %q has invalid license path", family.Product)
		}
		declared, ok := findConfigSource(boundary.muamba.config, sourceID)
		if !ok {
			return fmt.Errorf("iconpack family %q has no source declaration", family.Product)
		}
		provenance.Sources = append(provenance.Sources, provenanceSource{
			Namespace: family.Namespace, Product: family.Product,
			ManifestPath: boundary.muamba.manifestPath, ManifestSHA256: boundary.muamba.configHash,
			LicensePath: family.LicensePath, LicenseSHA256: boundary.checksums[family.LicensePath],
		})
		fmt.Fprintf(&notice, "%s\nSource: %s\nLicense: %s\n\n", family.Product, declared.URL, declared.License)
	}
	files["NOTICE"] = notice.Bytes()
	return nil
}

func findConfigSource(config Config, id string) (ConfigSource, bool) {
	for _, source := range config.Sources {
		if source.ID == id {
			return source, true
		}
	}
	return ConfigSource{}, false
}

func addGenericOutputs(boundary releaseBoundary, files map[string][]byte, provenance *provenanceDocument) error {
	family := boundary.generic.family
	files["PROVENANCE/"+boundary.generic.fileName+"-manifest.json"] = append([]byte(nil), boundary.generic.manifestBytes...)
	licenseBytes, err := readVerifiedReleaseFile(boundary, family.LicensePath)
	if err != nil {
		return err
	}
	files[family.LicenseOutput] = licenseBytes
	provenance.Sources = append(provenance.Sources, provenanceSource{
		Namespace:      family.Namespace,
		Product:        family.Product,
		SpritePath:     family.SpritePath,
		SpriteSHA256:   boundary.checksums[family.SpritePath],
		ManifestPath:   "source-manifest.json",
		ManifestSHA256: boundary.generic.manifestHash,
		LicensePath:    family.LicensePath,
		LicenseSHA256:  boundary.checksums[family.LicensePath],
	})
	if boundary.generic.manifest.NoticePath != "" {
		notice, err := readVerifiedReleaseFile(boundary, boundary.generic.manifest.NoticePath)
		if err != nil {
			return err
		}
		files["NOTICE"] = notice
		return nil
	}
	files["NOTICE"] = fmt.Appendf(nil, "%s %s\nSource: %s\nLicense: %s\n", boundary.generic.manifest.Name, boundary.generic.manifest.Release, boundary.generic.manifest.Source, boundary.generic.manifest.License)
	return nil
}

func addReleaseOutputs(boundary releaseBoundary, families []sourceFamily, files map[string][]byte, provenance *provenanceDocument) error {
	for _, family := range families {
		sourceBytes, err := readVerifiedReleaseFile(boundary, family.ProvenancePath)
		if err != nil {
			return err
		}
		files["PROVENANCE/"+family.Product+".json"] = sourceBytes
		licenseBytes, err := readVerifiedReleaseFile(boundary, family.LicensePath)
		if err != nil {
			return err
		}
		files[family.LicenseOutput] = licenseBytes
		provenance.Sources = append(provenance.Sources, provenanceSource{
			Namespace:        family.Namespace,
			Product:          family.Product,
			SpritePath:       family.SpritePath,
			SpriteSHA256:     boundary.checksums[family.SpritePath],
			ProvenancePath:   family.ProvenancePath,
			ProvenanceSHA256: boundary.checksums[family.ProvenancePath],
			LicensePath:      family.LicensePath,
			LicenseSHA256:    boundary.checksums[family.LicensePath],
		})
	}
	return nil
}

func readVerifiedReleaseFile(boundary releaseBoundary, relative string) ([]byte, error) {
	expected, ok := boundary.checksums[relative]
	if !ok {
		return nil, fmt.Errorf("required release file %q is absent from checksums.txt", relative)
	}
	b, ok := boundary.files[relative]
	if !ok {
		return nil, fmt.Errorf("required release file %q is absent from captured release boundary", relative)
	}
	if hashBytes(b) != expected {
		return nil, fmt.Errorf("captured release file %q failed its verified identity", relative)
	}
	return append([]byte(nil), b...), nil
}

func generateBindings(boundary releaseBoundary, opts Options, selected []selectedAsset) ([]byte, error) {
	if err := validateGeneratedNamespace(selected); err != nil {
		return nil, err
	}
	var output bytes.Buffer
	output.WriteString("// Code generated by goshtoso iconpack. DO NOT EDIT.\n")
	fmt.Fprintf(&output, "// Release: %s\n", boundary.catalog.Release)
	fmt.Fprintf(&output, "// Catalog SHA-256: %s\n\n", boundary.catalog.Hash)
	fmt.Fprintf(&output, "package %s\n\n", opts.Package)
	output.WriteString("import \"github.com/araihu/goshtoso/components/icon\"\n\n")
	output.WriteString("type Name string\n\n")
	output.WriteString("type Glyph struct {\n\tName Name\n\tCanonicalName string\n\tNamespace string\n\tProduct string\n\tSymbol icon.Symbol\n\tColorBehavior string\n}\n\n")
	output.WriteString("type Config struct {\n\tSymbol icon.Symbol\n\tSize icon.Size\n\tLabel string\n\tDecorative bool\n\tRootClass string\n}\n\n")
	fmt.Fprintf(&output, "const SpriteURL = %q\n\n", opts.SpriteURL)
	output.WriteString("const (\n")
	for _, asset := range selected {
		nameIdentifier, err := goIdentifier("Name", asset.CanonicalName)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(&output, "\t%s Name = %q\n", nameIdentifier, asset.CanonicalName)
		fmt.Fprintf(&output, "\t%s icon.Symbol = %q\n", asset.goName, asset.SpriteSymbol)
	}
	output.WriteString(")\n\n")
	output.WriteString("var Glyphs = []Glyph{\n")
	for _, asset := range selected {
		nameIdentifier, _ := goIdentifier("Name", asset.CanonicalName)
		fmt.Fprintf(&output, "\t{Name: %s, CanonicalName: %q, Namespace: %q, Product: %q, Symbol: %s, ColorBehavior: %q},\n", nameIdentifier, asset.CanonicalName, asset.Namespace, asset.Product, asset.goName, asset.ColorBehavior)
	}
	output.WriteString("}\n\n")
	output.WriteString("func Lookup(name Name) (Glyph, bool) {\n\tfor _, glyph := range Glyphs {\n\t\tif glyph.Name == name {\n\t\t\treturn glyph, true\n\t\t}\n\t}\n\treturn Glyph{}, false\n}\n\n")
	output.WriteString("func Icon(cfg Config) icon.Instance {\n\treturn icon.Icon(icon.Config{SpriteURL: SpriteURL, Symbol: cfg.Symbol, Size: cfg.Size, Label: cfg.Label, Decorative: cfg.Decorative, RootClass: cfg.RootClass})\n}\n")
	formatted, err := format.Source(output.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated Go bindings: %w", err)
	}
	return formatted, nil
}

func validateGeneratedNamespace(selected []selectedAsset) error {
	declarations := make(map[string]string, 7+2*len(selected))
	register := func(name, source string) error {
		if first, exists := declarations[name]; exists {
			return fmt.Errorf("go declaration collision %q between %s and %s", name, first, source)
		}
		declarations[name] = source
		return nil
	}
	for _, fixed := range []struct {
		name string
		kind string
	}{
		{name: "Name", kind: "fixed type"},
		{name: "Glyph", kind: "fixed type"},
		{name: "Config", kind: "fixed type"},
		{name: "SpriteURL", kind: "fixed constant"},
		{name: "Glyphs", kind: "fixed variable"},
		{name: "Lookup", kind: "fixed function"},
		{name: "Icon", kind: "fixed function"},
	} {
		if err := register(fixed.name, fixed.kind); err != nil {
			return err
		}
	}
	for _, asset := range selected {
		nameIdentifier, err := goIdentifier("Name", asset.CanonicalName)
		if err != nil {
			return err
		}
		if err := register(nameIdentifier, fmt.Sprintf("name constant for %q", asset.CanonicalName)); err != nil {
			return err
		}
		if err := register(asset.goName, fmt.Sprintf("symbol constant for %q", asset.CanonicalName)); err != nil {
			return err
		}
	}
	return nil
}

func goIdentifier(prefix, canonicalName string) (string, error) {
	var output strings.Builder
	output.WriteString(prefix)
	wordStart := true
	for _, character := range canonicalName {
		if character == '-' {
			wordStart = true
			continue
		}
		if character < '0' || character > '9' && character < 'A' || character > 'Z' && character < 'a' || character > 'z' {
			return "", fmt.Errorf("canonicalName %q cannot form a Go identifier", canonicalName)
		}
		if wordStart {
			output.WriteRune(unicode.ToUpper(character))
			wordStart = false
		} else {
			output.WriteRune(character)
		}
	}
	if !validGoIdentifier(output.String()) {
		return "", fmt.Errorf("canonicalName %q cannot form a Go identifier", canonicalName)
	}
	return output.String(), nil
}

func validGoIdentifier(value string) bool {
	if value == "" || token.Lookup(value).IsKeyword() {
		return false
	}
	for index, character := range value {
		if character == '_' || character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || index > 0 && character >= '0' && character <= '9' {
			continue
		}
		return false
	}
	return true
}

func marshalDocument(value any) ([]byte, error) {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

func hashBytes(contents []byte) string {
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:])
}
