package iconpack

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/araihu/goshtoso/iconlibrary"
)

func libraryExtension(path string) bool {
	return slices.Contains([]string{".svg", ".png", ".jpg", ".jpeg", ".webp"}, strings.ToLower(filepath.Ext(path)))
}

func generateLibrary(ctx context.Context, opts Options) (Result, error) {
	if opts.ConfigPath == "" || opts.OutputDir == "" {
		return Result{}, fmt.Errorf("library requires -config and -out")
	}
	if opts.ReleaseRoot != "" || opts.ReleaseArchive != "" || opts.SourceRoot != "" || opts.SourceArchive != "" || len(opts.Names) > 0 || opts.SelectionManifest != "" {
		return Result{}, fmt.Errorf("library uses only config sources and source paths")
	}
	input, err := prepareMuambaInput(opts)
	if err != nil {
		return Result{}, err
	}
	engine, err := openMuambaEngine(ctx, input, opts)
	if err != nil {
		return Result{}, err
	}
	snapshot, err := engine.Snapshot(ctx, nil)
	if err != nil {
		return Result{}, err
	}
	files, _, err := captureMuambaFiles(snapshot, input.sources)
	if err != nil {
		return Result{}, err
	}
	output, catalog, err := libraryOutputs(input.sources, files)
	if err != nil {
		return Result{}, err
	}
	lock, err := os.ReadFile(input.lockPath)
	if err != nil {
		return Result{}, err
	}
	output["PROVENANCE/config.yaml"] = input.configBytes
	output["PROVENANCE/lock.yaml"] = lock
	release := "iconpack-" + hashBytes(input.configBytes)[:12]
	catalogHash := hashBytes(output["catalog.json"])
	manifest := outputManifest{SchemaVersion: OutputSchemaVersion, Tool: toolName, Release: release, CatalogSchemaVersion: catalog.SchemaVersion, CatalogSHA256: catalogHash, SourceKind: "muamba-snapshot", SourceConfigSHA256: hashBytes(input.configBytes), SourceLockSHA256: hashBytes(lock)}
	for name, data := range output {
		manifest.Files = append(manifest.Files, outputFile{Path: name, Mode: "0644", Bytes: len(data), SHA256: hashBytes(data)})
	}
	sort.Slice(manifest.Files, func(i, j int) bool { return manifest.Files[i].Path < manifest.Files[j].Path })
	output["manifest.json"], err = marshalDocument(manifest)
	if err != nil {
		return Result{}, err
	}
	published, path, err := publishOutput(ctx, opts.OutputDir, output, opts.Check)
	return Result{Release: release, OutputDir: path, Published: published, SelectedCount: len(catalog.Icons), CatalogSHA256: catalogHash}, err
}

func libraryOutputs(sources []resolvedConfigSource, files map[string][]byte) (map[string][]byte, iconlibrary.Catalog, error) {
	out := map[string][]byte{}
	catalog := iconlibrary.Catalog{SchemaVersion: 1}
	seen := map[string]bool{}
	for _, source := range sources {
		icons, err := sourceLibraryIcons(source, files)
		if err != nil {
			return nil, catalog, err
		}
		license := files[source.ID+"/"+source.LicensePath]
		if len(license) == 0 {
			return nil, catalog, fmt.Errorf("missing license for %s", source.ID)
		}
		out["LICENSES/"+source.ID+".txt"] = license
		for _, entry := range icons {
			if seen[entry.ID] {
				return nil, catalog, fmt.Errorf("duplicate icon %s", entry.ID)
			}
			seen[entry.ID] = true
			for i := range entry.Variants {
				v := &entry.Variants[i]
				raw, ok := files[source.ID+"/"+v.Path]
				if !ok {
					return nil, catalog, fmt.Errorf("missing icon %s", v.Path)
				}
				mime, w, h, err := iconlibrary.ValidateImage(raw)
				if err != nil {
					return nil, catalog, fmt.Errorf("%s: %w", v.Path, err)
				}
				v.MIME, v.Width, v.Height, v.SHA256 = mime, w, h, hashBytes(raw)
				ext := map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "image/webp": ".webp", "image/svg+xml": ".svg"}[mime]
				v.Path = "images/" + v.SHA256 + ext
				out[v.Path] = raw
			}
			catalog.Icons = append(catalog.Icons, entry)
		}
	}
	if len(catalog.Icons) == 0 {
		return nil, catalog, fmt.Errorf("no icons selected")
	}
	sort.Slice(catalog.Icons, func(i, j int) bool { return catalog.Icons[i].ID < catalog.Icons[j].ID })
	data, err := marshalDocument(catalog)
	if err != nil {
		return nil, catalog, err
	}
	out["catalog.json"] = data
	var notice strings.Builder
	for _, s := range sources {
		fmt.Fprintf(&notice, "%s: %s\nSource: %s\nLicense: LICENSES/%s.txt\n\n", s.ID, s.License, s.URL, s.ID)
	}
	out["NOTICE"] = []byte(strings.TrimSpace(notice.String()) + "\n")
	return out, catalog, nil
}

func sourceLibraryIcons(s resolvedConfigSource, files map[string][]byte) ([]iconlibrary.Icon, error) {
	if s.MetadataFormat == "selfhst" {
		return selfhstIcons(s, files)
	}
	var result []iconlibrary.Icon
	byReference := map[string]int{}
	for _, key := range slices.Sorted(maps.Keys(files)) {
		relative, ok := strings.CutPrefix(key, s.ID+"/")
		if !ok || !libraryExtension(relative) || relative == s.LicensePath {
			continue
		}
		format := libraryFormat(relative)
		if len(s.Formats) > 0 && !slices.Contains(s.Formats, format) {
			continue
		}
		if len(s.Paths) > 0 && !slices.Contains(s.Paths, relative) {
			continue
		}
		ref := normalizeWebPath(relative)
		index, exists := byReference[ref]
		if !exists {
			index = len(result)
			byReference[ref] = index
			result = append(result, iconlibrary.Icon{ID: s.ID + ":" + ref, Name: ref, Source: s.ID, Reference: ref, License: s.License, SourceURL: s.URL})
		}
		result[index].Variants = append(result[index].Variants, iconlibrary.Variant{Appearance: "default", Path: relative})
	}
	for i := range result {
		slices.SortStableFunc(result[i].Variants, func(a, b iconlibrary.Variant) int {
			return cmp.Compare(slices.Index(s.Formats, libraryFormat(a.Path)), slices.Index(s.Formats, libraryFormat(b.Path)))
		})
	}
	return result, nil
}

func libraryFormat(path string) string {
	format := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	if format == "jpg" {
		return "jpeg"
	}
	return format
}

func selfhstIcons(s resolvedConfigSource, files map[string][]byte) ([]iconlibrary.Icon, error) {
	var rows []struct{ Name, Reference, SVG, PNG, Light, Dark, Category, Tags string }
	if err := json.Unmarshal(files[s.ID+"/"+s.MetadataPath], &rows); err != nil {
		return nil, fmt.Errorf("selfhst metadata: %w", err)
	}
	formats := s.Formats
	if len(formats) == 0 {
		formats = []string{"png"}
	}
	var result []iconlibrary.Icon
	for _, row := range rows {
		if row.Reference == "" || normalizeWebPath(row.Reference) != row.Reference || row.Name == "" {
			return nil, fmt.Errorf("invalid selfhst reference %q", row.Reference)
		}
		availableFormats := selfhstAvailableFormats(formats, row.SVG, row.PNG)
		if len(availableFormats) == 0 {
			continue
		}
		entry := iconlibrary.Icon{ID: s.ID + ":" + row.Reference, Reference: row.Reference, Name: row.Name, Source: s.ID, SourceURL: s.URL, License: s.License}
		for _, tag := range strings.FieldsFunc(row.Category+","+row.Tags, func(r rune) bool { return r == ',' || r == ';' }) {
			if tag = strings.TrimSpace(tag); tag != "" {
				entry.Tags = append(entry.Tags, tag)
			}
		}
		for _, appearance := range []string{"default", "light", "dark"} {
			if appearance == "light" && row.Light != "Yes" || appearance == "dark" && row.Dark != "Yes" {
				continue
			}
			ref := row.Reference
			if appearance != "default" {
				ref += "-" + appearance
			}
			found := false
			for _, format := range availableFormats {
				p := format + "/" + ref + "." + format
				if _, ok := files[s.ID+"/"+p]; ok {
					entry.Variants = append(entry.Variants, iconlibrary.Variant{Appearance: appearance, Path: p})
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("missing %s variant for %s", appearance, row.Reference)
			}
		}
		result = append(result, entry)
	}
	return result, nil
}

func selfhstAvailableFormats(formats []string, svg, png string) []string {
	return slices.DeleteFunc(slices.Clone(formats), func(format string) bool {
		return format == "svg" && svg != "Yes" || format == "png" && png != "Yes"
	})
}
