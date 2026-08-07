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
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
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
	SchemaVersion int                `json:"schemaVersion"`
	Tool          string             `json:"tool"`
	Release       string             `json:"release"`
	CatalogSHA256 string             `json:"catalogSha256"`
	Sources       []provenanceSource `json:"sources"`
	Assets        []provenanceAsset  `json:"assets"`
}

type provenanceSource struct {
	Namespace        string `json:"namespace"`
	Product          string `json:"product"`
	SpritePath       string `json:"spritePath"`
	SpriteSHA256     string `json:"spriteSha256"`
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
	boundary, err := openRelease(opts)
	if err != nil {
		return Result{}, err
	}
	defer boundary.cleanup()
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
	if opts.SpriteURL != strings.TrimSpace(opts.SpriteURL) || strings.ContainsAny(opts.SpriteURL, "\\\r\n") {
		return fmt.Errorf("invalid sprite-url %q", opts.SpriteURL)
	}
	parsed, err := url.Parse(opts.SpriteURL)
	if err != nil || parsed.Path == "" || parsed.Fragment != "" || parsed.Host != "" && parsed.Scheme == "" || parsed.User != nil || parsed.Scheme != "" && parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("sprite-url must be a relative, http, or https sprite document URL")
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
	}
	for _, family := range families {
		sourceBytes, err := readVerifiedReleaseFile(boundary, family.ProvenancePath)
		if err != nil {
			return nil, err
		}
		provenanceOutput := "PROVENANCE/" + family.Product + ".json"
		files[provenanceOutput] = sourceBytes
		licenseBytes, err := readVerifiedReleaseFile(boundary, family.LicensePath)
		if err != nil {
			return nil, err
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
	if notice, err := readVerifiedReleaseFile(boundary, "NOTICE"); err == nil {
		files["NOTICE"] = notice
	} else {
		return nil, err
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
		Package:              opts.Package,
		ConstPrefix:          opts.ConstPrefix,
		SpriteURL:            opts.SpriteURL,
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

func readVerifiedReleaseFile(boundary releaseBoundary, relative string) ([]byte, error) {
	expected, ok := boundary.checksums[relative]
	if !ok {
		return nil, fmt.Errorf("required release file %q is absent from checksums.txt", relative)
	}
	b, err := os.ReadFile(filepath.Join(boundary.root, filepath.FromSlash(relative)))
	if err != nil {
		return nil, fmt.Errorf("read release file %q: %w", relative, err)
	}
	if hashBytes(b) != expected {
		return nil, fmt.Errorf("release file %q changed after verification", relative)
	}
	return b, nil
}

func generateBindings(boundary releaseBoundary, opts Options, selected []selectedAsset) ([]byte, error) {
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
