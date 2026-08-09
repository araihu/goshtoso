package iconpack

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/araihu/goshtoso/internal/iconcatalog"
	muambasource "github.com/araihu/muamba/source"
)

type muambaPack struct {
	config       Config
	configBytes  []byte
	manifestPath string
	configHash   string
	lockHash     string
	families     map[string]sourceFamily
}

type muambaInput struct {
	config      Config
	configBytes []byte
	sources     []resolvedConfigSource
	configDir   string
	lockPath    string
	engineBytes []byte
}

func openMuambaPack(ctx context.Context, opts Options) (releaseBoundary, error) {
	input, err := prepareMuambaInput(opts)
	if err != nil {
		return releaseBoundary{}, err
	}
	engine, err := openMuambaEngine(ctx, input, opts)
	if err != nil {
		return releaseBoundary{}, err
	}
	snapshot, err := engine.Snapshot(ctx, nil)
	if err != nil {
		return releaseBoundary{}, fmt.Errorf("verify iconpack sources: %w", err)
	}
	lockBytes, err := os.ReadFile(input.lockPath)
	if err != nil {
		return releaseBoundary{}, fmt.Errorf("read iconpack lock: %w", err)
	}
	configHash := hashBytes(input.configBytes)
	files, checksums, err := captureMuambaFiles(snapshot, input.sources)
	if err != nil {
		return releaseBoundary{}, err
	}
	families, assets, err := buildMuambaAssets(input.sources, files)
	if err != nil {
		return releaseBoundary{}, err
	}
	if len(assets) == 0 {
		return releaseBoundary{}, fmt.Errorf("iconpack sources contain no selected SVG files")
	}
	return releaseBoundary{
		sourceKind: "muamba-snapshot",
		catalog: iconcatalog.Catalog{
			SchemaVersion: 1, Release: "iconpack-" + configHash[:12], IdentityRevision: 1,
			Assets: assets, Hash: configHash,
		},
		checksums: checksums,
		files:     files,
		muamba: &muambaPack{
			config: input.config, configBytes: append([]byte(nil), input.configBytes...), manifestPath: filepath.Base(opts.ConfigPath), configHash: hashBytes(input.configBytes),
			lockHash: hashBytes(lockBytes), families: families,
		},
		cleanup: func() {},
	}, nil
}

func prepareMuambaInput(opts Options) (muambaInput, error) {
	if err := rejectMuambaPath(opts.ConfigPath); err != nil {
		return muambaInput{}, err
	}
	if opts.IconpackLockPath != "" {
		if err := rejectMuambaPath(opts.IconpackLockPath); err != nil {
			return muambaInput{}, err
		}
	}
	configBytes, config, err := readIconpackConfig(opts.ConfigPath)
	if err != nil {
		return muambaInput{}, err
	}
	configPath, err := filepath.Abs(opts.ConfigPath)
	if err != nil {
		return muambaInput{}, fmt.Errorf("resolve iconpack config: %w", err)
	}
	configDir := filepath.Dir(configPath)
	lockPath := opts.IconpackLockPath
	if lockPath == "" {
		lockPath = filepath.Join(configDir, ".iconpack.lock.yaml")
	} else if !filepath.IsAbs(lockPath) {
		lockPath = filepath.Join(configDir, lockPath)
	}
	lockPath, err = filepath.Abs(lockPath)
	if err != nil {
		return muambaInput{}, fmt.Errorf("resolve iconpack lock: %w", err)
	}
	backend, sources, err := backendForConfig(config)
	if err != nil {
		return muambaInput{}, err
	}
	engineBytes, err := marshalBackendManifest(backend)
	if err != nil {
		return muambaInput{}, fmt.Errorf("encode Muamba source declaration: %w", err)
	}
	return muambaInput{
		config: config, configBytes: configBytes, sources: sources,
		configDir: configDir, lockPath: lockPath, engineBytes: engineBytes,
	}, nil
}

func openMuambaEngine(ctx context.Context, input muambaInput, opts Options) (*muambasource.Engine, error) {
	enginePath := filepath.Join(input.configDir, ".iconpack.engine.yaml")
	if err := os.WriteFile(enginePath, input.engineBytes, 0o644); err != nil {
		return nil, fmt.Errorf("write generated iconpack engine declaration: %w", err)
	}
	engine, err := muambasource.New(muambasource.Options{
		ManifestPath: enginePath,
		LockPath:     input.lockPath,
		CacheDir:     filepath.Join(input.configDir, ".iconpack-cache"),
		AllowHTTP:    opts.AllowHTTP,
	})
	if err != nil {
		return nil, fmt.Errorf("open iconpack Muamba engine: %w", err)
	}
	if opts.Trust {
		if _, err := engine.Lock(ctx, nil); err != nil {
			return nil, fmt.Errorf("trust iconpack sources: %w", err)
		}
		return engine, nil
	}
	if _, err := os.Stat(input.lockPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("iconpack lock %s is missing; rerun with -trust for explicit first trust", input.lockPath)
	} else if err != nil {
		return nil, fmt.Errorf("inspect iconpack lock: %w", err)
	}
	return engine, nil
}

func captureMuambaFiles(snapshot []muambasource.SnapshotFile, sources []resolvedConfigSource) (map[string][]byte, map[string]string, error) {
	files := make(map[string][]byte, len(snapshot))
	checksums := make(map[string]string, len(snapshot))
	for _, file := range snapshot {
		key, err := snapshotKey(file.Path, sources)
		if err != nil {
			return nil, nil, err
		}
		if _, duplicate := files[key]; duplicate {
			return nil, nil, fmt.Errorf("iconpack source files resolve to duplicate path %q", key)
		}
		contents := append([]byte(nil), file.Contents...)
		files[key] = contents
		checksums[key] = hashBytes(contents)
	}
	return files, checksums, nil
}

func rejectMuambaPath(filename string) error {
	base := strings.ToLower(filepath.Base(filename))
	switch base {
	case "muamba.yaml", ".muamba.yaml", "muamba.lock.yaml", ".muamba.lock.yaml":
		return fmt.Errorf("iconpack path %q must not use a Muamba manifest name", filename)
	default:
		return nil
	}
}

func snapshotKey(target string, sources []resolvedConfigSource) (string, error) {
	target = filepath.ToSlash(target)
	for _, source := range sources {
		base := filepath.ToSlash(filepath.Join(".iconpack", "sources", source.ID))
		if !strings.HasPrefix(target, base+"/") {
			continue
		}
		relative := strings.TrimPrefix(target, base+"/")
		if relative == "" || strings.HasPrefix(relative, "../") || relative == ".." {
			return "", fmt.Errorf("iconpack snapshot path %q escapes source %q", target, source.ID)
		}
		return filepath.ToSlash(filepath.Join(source.ID, relative)), nil
	}
	return "", fmt.Errorf("muamba snapshot path %q does not belong to a declared iconpack source", target)
}

func buildMuambaAssets(sources []resolvedConfigSource, files map[string][]byte) (map[string]sourceFamily, []iconcatalog.Asset, error) {
	families := make(map[string]sourceFamily, len(sources))
	assets := make([]iconcatalog.Asset, 0)
	seenNames := make(map[string]string)
	for _, source := range sortedConfigSources(sources) {
		family, sourceAssets, err := buildMuambaSourceAssets(source, files, seenNames)
		if err != nil {
			return nil, nil, err
		}
		familyKey := family.Namespace + "/" + family.Product
		if previous, exists := families[familyKey]; exists && previous.LicensePath != family.LicensePath {
			return nil, nil, fmt.Errorf("iconpack packName %q collides across sources", source.PackName)
		}
		families[familyKey] = family
		assets = append(assets, sourceAssets...)
	}
	sort.Slice(assets, func(i, j int) bool { return assets[i].CanonicalName < assets[j].CanonicalName })
	return families, assets, nil
}

func buildMuambaSourceAssets(source resolvedConfigSource, files map[string][]byte, seenNames map[string]string) (sourceFamily, []iconcatalog.Asset, error) {
	familyProduct := source.PackName
	if familyProduct == "" {
		familyProduct = source.ID
	}
	family := sourceFamily{
		Namespace: "custom", Product: familyProduct, Generic: true,
		LicensePath:   filepath.ToSlash(filepath.Join(source.ID, source.LicensePath)),
		LicenseOutput: "LICENSES/" + sourcePackFileName(familyProduct) + "-LICENSE.txt",
	}
	licenseBytes, ok := files[family.LicensePath]
	if !ok || len(licenseBytes) == 0 {
		return sourceFamily{}, nil, fmt.Errorf("iconpack source %q license %q was not found in the verified snapshot", source.ID, source.LicensePath)
	}
	wanted := make(map[string]struct{}, len(source.Paths))
	for _, item := range source.Paths {
		wanted[item] = struct{}{}
	}
	assets := make([]iconcatalog.Asset, 0)
	base := source.ID + "/"
	for key, raw := range files {
		if !strings.HasPrefix(key, base) {
			continue
		}
		relative := strings.TrimPrefix(key, base)
		if !isSelectedMuambaIcon(source, relative, wanted) {
			continue
		}
		viewBox, err := svgViewBox(raw, relative)
		if err != nil {
			return sourceFamily{}, nil, err
		}
		namePrefix := source.PackName
		if namePrefix == "" {
			namePrefix = source.ID
		}
		canonical := namePrefix + "-" + normalizeWebPath(relative)
		symbol := normalizeWebPath(canonical)
		if canonical == "-" || symbol == "" {
			return sourceFamily{}, nil, fmt.Errorf("iconpack source %q path %q cannot form a canonical name", source.ID, relative)
		}
		if previous, exists := seenNames[canonical]; exists {
			return sourceFamily{}, nil, fmt.Errorf("generated canonical name %q collides between %s and %s", canonical, previous, key)
		}
		seenNames[canonical] = key
		assets = append(assets, iconcatalog.Asset{
			CanonicalName: canonical, Namespace: "custom", Product: familyProduct,
			Path: key, Artwork: "icon", Appearance: "default", Surface: "transparent", Framing: "optical",
			Format: "svg", Dimensions: iconcatalog.Dimensions{ViewBox: viewBox}, SpriteSymbol: symbol,
			ColorBehavior: "currentColor", License: source.License,
			Source: source.URL + ":" + relative, SHA256: hashBytes(raw),
		})
	}
	return family, assets, nil
}

func isSelectedMuambaIcon(source resolvedConfigSource, relative string, wanted map[string]struct{}) bool {
	if relative == source.LicensePath || !strings.EqualFold(filepath.Ext(relative), ".svg") {
		return false
	}
	if source.Kind == "file" && relative != source.IconPath {
		return false
	}
	if len(wanted) == 0 {
		return true
	}
	_, ok := wanted[relative]
	return ok
}

func svgViewBox(raw []byte, relative string) (string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(raw))
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", fmt.Errorf("parse source SVG %q: %w", relative, err)
		}
		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		if start.Name.Local != "svg" || start.Name.Space != "http://www.w3.org/2000/svg" {
			return "", fmt.Errorf("source SVG %q has invalid root element", relative)
		}
		for _, attr := range start.Attr {
			if attr.Name.Space == "" && attr.Name.Local == "viewBox" {
				if err := validateViewBox(attr.Value); err != nil {
					return "", err
				}
				return attr.Value, nil
			}
		}
		return "", fmt.Errorf("source SVG %q has no viewBox", relative)
	}
}
