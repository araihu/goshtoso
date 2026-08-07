package iconpack

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestGenerateFromVerifiedRootPreservesLiteralIdentityAndOwnership(t *testing.T) {
	opts := syntheticRelease(t)
	opts.OutputDir = filepath.Join(t.TempDir(), "pack")
	opts.Names = []string{"brand-developer-icons-tRPC"}
	opts.Package = "appicons"
	opts.ConstPrefix = "Icon"
	opts.SpriteURL = "/assets/icons/app.svg"

	result, err := Generate(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Published || result.SelectedCount != 1 {
		t.Fatalf("result = %+v, want one published icon", result)
	}
	assertFileContains(t, filepath.Join(opts.OutputDir, "sprite.svg"), `id="devicon-trpc"`)
	assertFileContains(t, filepath.Join(opts.OutputDir, "icons_gen.go"), `NameBrandDeveloperIconsTRPC Name`)
	assertFileContains(t, filepath.Join(opts.OutputDir, "icons_gen.go"), `IconBrandDeveloperIconsTRPC icon.Symbol = "devicon-trpc"`)
	assertFileContains(t, filepath.Join(opts.OutputDir, "manifest.json"), `"canonicalName": "brand-developer-icons-tRPC"`)

	opts.Check = true
	result, err = Generate(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if result.Published {
		t.Fatal("check unexpectedly published output")
	}

	if err := os.WriteFile(filepath.Join(opts.OutputDir, "unrelated.txt"), []byte("owner data\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	opts.Check = false
	if _, err := Generate(context.Background(), opts); err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("Generate() error = %v, want unrelated-file refusal", err)
	}
	assertFileContains(t, filepath.Join(opts.OutputDir, "unrelated.txt"), "owner data")
}

func TestGenerateFromVerifiedArchive(t *testing.T) {
	opts := syntheticRelease(t)
	archive := filepath.Join(t.TempDir(), "araihu-assets-v0.2.0.tar.gz")
	writeTarGzip(t, opts.ReleaseRoot, archive)
	archiveBytes, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	opts.ReleaseRoot = ""
	opts.ReleaseArchive = archive
	opts.ArchiveSHA256 = hashBytes(archiveBytes)
	opts.OutputDir = filepath.Join(t.TempDir(), "pack")
	opts.Names = []string{"brand-developer-icons-tRPC"}
	opts.Package = "appicons"
	opts.ConstPrefix = "Icon"
	opts.SpriteURL = "/assets/icons/app.svg"

	if _, err := Generate(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(opts.OutputDir, "sprite.svg"), `id="devicon-trpc"`)
}

func TestGenerateFromVerifiedRootRejectsNestedOutsideSymlink(t *testing.T) {
	opts := syntheticRelease(t)
	outside := t.TempDir()
	originalIcons := filepath.Join(opts.ReleaseRoot, "icons")
	movedIcons := filepath.Join(outside, "icons")
	if err := os.Rename(originalIcons, movedIcons); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(movedIcons, originalIcons); err != nil {
		t.Fatal(err)
	}
	opts.OutputDir = filepath.Join(t.TempDir(), "pack")
	opts.Names = []string{"brand-developer-icons-tRPC"}
	opts.Package = "appicons"
	opts.ConstPrefix = "Icon"
	opts.SpriteURL = "/assets/icons/app.svg"

	_, err := Generate(context.Background(), opts)
	if err == nil || !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("Generate() error = %v, want nested symbolic-link rejection", err)
	}
	if _, statErr := os.Lstat(opts.OutputDir); !os.IsNotExist(statErr) {
		t.Fatalf("output directory exists after rejected release root: %v", statErr)
	}
}

func TestVerifiedRootUsesCapturedBytesAfterSourceMutation(t *testing.T) {
	opts := syntheticRelease(t)
	boundary, err := openRelease(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer boundary.cleanup()

	if err := os.WriteFile(
		filepath.Join(opts.ReleaseRoot, "icons/brand/developer-icons/tRPC.svg"),
		[]byte(`<svg xmlns="http://www.w3.org/2000/svg"/>`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(opts.ReleaseRoot, "icons/brand/developer-icons/sprite.svg"),
		[]byte(`<svg xmlns="http://www.w3.org/2000/svg"><symbol id="mutated" viewBox="0 0 1 1"/></svg>`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	selected, families, err := selectAssets(boundary, []string{"brand-developer-icons-tRPC"}, "Icon")
	if err != nil {
		t.Fatalf("select captured asset: %v", err)
	}
	sprite, err := buildSprite(boundary, selected, families)
	if err != nil {
		t.Fatalf("build from captured sprite: %v", err)
	}
	if !strings.Contains(string(sprite), `id="devicon-trpc"`) || strings.Contains(string(sprite), `id="mutated"`) {
		t.Fatalf("sprite uses mutable release-root bytes:\n%s", sprite)
	}
}

func TestExtractArchiveRejectsTraversal(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "bad.tar.gz")
	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)
	contents := []byte("escape")
	if err := tw.WriteHeader(&tar.Header{Name: "../escape", Mode: 0o644, Size: int64(len(contents)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(contents); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	err = extractArchive(archive, t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "unsafe archive member") {
		t.Fatalf("extractArchive() error = %v, want traversal rejection", err)
	}
}

func TestSelectionManifestJSONAndYAML(t *testing.T) {
	for _, test := range []struct {
		name, extension, contents string
	}{
		{"json", ".json", `{"schemaVersion":1,"names":["brand-developer-icons-tRPC"]}`},
		{"yaml", ".yaml", "schemaVersion: 1\nnames:\n  - brand-developer-icons-tRPC\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "selection"+test.extension)
			if err := os.WriteFile(path, []byte(test.contents), 0o644); err != nil {
				t.Fatal(err)
			}
			names, err := selectionNames(Options{SelectionManifest: path})
			if err != nil {
				t.Fatal(err)
			}
			if len(names) != 1 || names[0] != "brand-developer-icons-tRPC" {
				t.Fatalf("names = %v", names)
			}
		})
	}
}

func TestSelectionManifestRejectsDuplicateJSONKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "selection.json")
	contents := `{"schemaVersion":1,"schemaVersion":1,"names":["brand-developer-icons-tRPC"]}`
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := selectionNames(Options{SelectionManifest: path}); err == nil || !strings.Contains(err.Error(), "duplicate key") {
		t.Fatalf("selectionNames() error = %v, want duplicate key", err)
	}
}

func syntheticRelease(t *testing.T) Options {
	t.Helper()
	root := t.TempDir()
	asset := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><path fill="#398CCB" d="M0 0h100v100H0z"/></svg>`)
	sprite := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><symbol id="devicon-trpc" viewBox="0 0 100 100"><path fill="#398CCB" d="M0 0h100v100H0z"/></symbol></svg>`)
	catalog := fmt.Sprintf(`{
  "schemaVersion": 2,
  "release": "v0.2.0",
  "identityRevision": 11,
  "assets": [
    {
      "canonicalName": "brand-developer-icons-tRPC",
      "namespace": "brand",
      "path": "icons/brand/developer-icons/tRPC.svg",
      "product": "developer-icons",
      "artwork": "icon",
      "appearance": "default",
      "surface": "transparent",
      "framing": "optical",
      "format": "svg",
      "dimensions": {"viewBox": "0 0 100 100"},
      "spriteSymbol": "devicon-trpc",
      "colorBehavior": "protected",
      "license": "MIT",
      "source": "developer-icons@example:icons/tRPC.svg",
      "sha256": %q
    }
  ]
}
`, hashBytes(asset))
	files := map[string][]byte{
		"NOTICE":                                      []byte("Synthetic notice\n"),
		"campaigns.json":                              []byte("{}\n"),
		"catalog.json":                                []byte(catalog),
		"icons/brand/developer-icons/tRPC.svg":        asset,
		"icons/brand/developer-icons/sprite.svg":      sprite,
		"icons/brand/developer-icons/provenance.json": []byte("{}\n"),
		"licenses/developer-icons-MIT.txt":            []byte("Synthetic MIT license\n"),
		"themes.json":                                 []byte("{}\n"),
	}
	for relative, contents := range files {
		writeFixtureFile(t, root, relative, contents)
	}
	paths := make([]string, 0, len(files))
	for relative := range files {
		paths = append(paths, relative)
	}
	sort.Strings(paths)
	release := releaseDocument{
		SchemaVersion: 1, Release: "v0.2.0", IdentityRevision: 11, RuntimeVersion: 1,
		CatalogSHA256: hashBytes(files["catalog.json"]), ThemesSHA256: hashBytes(files["themes.json"]), CampaignsSHA256: hashBytes(files["campaigns.json"]),
	}
	releaseOrder := []string{"catalog.json", "themes.json", "campaigns.json"}
	for _, relative := range paths {
		if relative != "catalog.json" && relative != "themes.json" && relative != "campaigns.json" {
			releaseOrder = append(releaseOrder, relative)
		}
	}
	for _, relative := range releaseOrder {
		release.Files = append(release.Files, releaseFile{Path: relative, SHA256: hashBytes(files[relative]), Size: int64(len(files[relative]))})
	}
	releaseBytes, err := marshalDocument(release)
	if err != nil {
		t.Fatal(err)
	}
	files["release.json"] = releaseBytes
	writeFixtureFile(t, root, "release.json", releaseBytes)
	paths = append(paths, "release.json")
	sort.Strings(paths)
	var checksums strings.Builder
	for _, relative := range paths {
		fmt.Fprintf(&checksums, "%s  %s\n", hashBytes(files[relative]), relative)
	}
	checksumsBytes := []byte(checksums.String())
	writeFixtureFile(t, root, "checksums.txt", checksumsBytes)
	return Options{
		ReleaseRoot:       root,
		Release:           "v0.2.0",
		CatalogSHA256:     hashBytes(files["catalog.json"]),
		ReleaseJSONSHA256: hashBytes(releaseBytes),
		ChecksumsSHA256:   hashBytes(checksumsBytes),
	}
}

func writeFixtureFile(t *testing.T, root, relative string, contents []byte) {
	t.Helper()
	filename := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, contents, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeTarGzip(t *testing.T, root, archive string) {
	t.Helper()
	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		header := &tar.Header{Name: filepath.ToSlash(relative), Mode: 0o644, Size: int64(len(contents)), Typeflag: tar.TypeReg}
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		_, err = tw.Write(contents)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func assertFileContains(t *testing.T, filename, want string) {
	t.Helper()
	b, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), want) {
		t.Fatalf("%s does not contain %q", filename, want)
	}
}
