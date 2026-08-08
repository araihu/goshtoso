package iconpack

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/internal/iconcatalog"
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
	assertFileContains(t, filepath.Join(opts.OutputDir, "manifest.json"), `"sourceKind": "release-root"`)
	assertFileContains(t, filepath.Join(opts.OutputDir, "provenance.json"), `"archiveSha256": ""`)
	manifestBefore, err := os.Stat(filepath.Join(opts.OutputDir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	idempotent, err := Generate(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if idempotent.Published {
		t.Fatal("identical output was replaced instead of reused")
	}
	manifestAfter, err := os.Stat(filepath.Join(opts.OutputDir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(manifestBefore, manifestAfter) {
		t.Fatal("identical publication replaced existing output")
	}

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
	assertFileContains(t, filepath.Join(opts.OutputDir, "manifest.json"), `"sourceKind": "release-archive"`)
	assertFileContains(t, filepath.Join(opts.OutputDir, "manifest.json"), `"archiveSha256": "`+opts.ArchiveSHA256+`"`)
}

func TestGenerateRejectsUnsafeSpriteURL(t *testing.T) {
	for _, spriteURL := range []string{
		" /assets/icons/app.svg",
		"/assets/icons/app.svg\n",
		`..\icons\app.svg`,
		"javascript:app.svg",
		"http:/assets/icons/app.svg",
		"//example.com/app.svg",
		"/assets/icons/app.svg#symbol",
	} {
		t.Run(fmt.Sprintf("%q", spriteURL), func(t *testing.T) {
			opts := syntheticRelease(t)
			opts.OutputDir = filepath.Join(t.TempDir(), "pack")
			opts.Names = []string{"brand-developer-icons-tRPC"}
			opts.Package = "appicons"
			opts.ConstPrefix = "Icon"
			opts.SpriteURL = spriteURL

			if _, err := Generate(context.Background(), opts); err == nil || !strings.Contains(err.Error(), "sprite-url") {
				t.Fatalf("Generate() error = %v, want sprite-url rejection", err)
			}
		})
	}
}

func TestTrustedV020ReleaseBoundaryAcceptsOfficialArchiveKinds(t *testing.T) {
	opts := Options{
		Release:           assetsV020Release,
		CatalogSHA256:     assetsV020CatalogSHA256,
		ReleaseJSONSHA256: assetsV020ReleaseJSONSHA256,
		ChecksumsSHA256:   assetsV020ChecksumsSHA256,
	}
	if err := validateTrustedReleaseIdentity(opts); err != nil {
		t.Fatalf("validateTrustedReleaseIdentity(root) = %v", err)
	}
	for _, archive := range []struct {
		name, digest string
	}{
		{"assets-v0.2.0.tar.gz", assetsV020TarGzipSHA256},
		{"assets-v0.2.0.tgz", assetsV020TarGzipSHA256},
		{"assets-v0.2.0.zip", assetsV020ZipSHA256},
	} {
		t.Run(archive.name, func(t *testing.T) {
			candidate := opts
			candidate.ReleaseArchive = archive.name
			candidate.ArchiveSHA256 = archive.digest
			if err := validateTrustedReleaseIdentity(candidate); err != nil {
				t.Fatalf("validateTrustedReleaseIdentity(archive) = %v", err)
			}
		})
	}
}

func TestTrustedV020ReleaseBoundaryRejectsWrongKindDigestAndRepack(t *testing.T) {
	base := Options{
		Release: assetsV020Release, CatalogSHA256: assetsV020CatalogSHA256,
		ReleaseJSONSHA256: assetsV020ReleaseJSONSHA256, ChecksumsSHA256: assetsV020ChecksumsSHA256,
	}
	for _, test := range []struct {
		name, archive, digest string
	}{
		{"tar with zip digest", "assets-v0.2.0.tar.gz", assetsV020ZipSHA256},
		{"zip with tar digest", "assets-v0.2.0.zip", assetsV020TarGzipSHA256},
		{"repacked", "assets-v0.2.0.tar.gz", strings.Repeat("0", 64)},
		{"unsupported", "assets-v0.2.0.7z", strings.Repeat("0", 64)},
	} {
		t.Run(test.name, func(t *testing.T) {
			opts := base
			opts.ReleaseArchive, opts.ArchiveSHA256 = test.archive, test.digest
			if err := validateTrustedReleaseIdentity(opts); err == nil {
				t.Fatal("validateTrustedReleaseIdentity() accepted wrong archive boundary")
			}
		})
	}
}

func TestGenerateRejectsConstPrefixNameCollisionBeforePublishing(t *testing.T) {
	opts := syntheticRelease(t)
	opts.OutputDir = filepath.Join(t.TempDir(), "pack")
	opts.Names = []string{"brand-developer-icons-tRPC"}
	opts.Package = "appicons"
	opts.ConstPrefix = "Name"
	opts.SpriteURL = "/assets/icons/app.svg"

	_, err := Generate(context.Background(), opts)
	if err == nil || !strings.Contains(err.Error(), `go declaration collision "NameBrandDeveloperIconsTRPC"`) {
		t.Fatalf("Generate() error = %v, want complete declaration namespace collision", err)
	}
	if _, statErr := os.Lstat(opts.OutputDir); !os.IsNotExist(statErr) {
		t.Fatalf("output directory exists after declaration collision: %v", statErr)
	}
}

func TestGeneratedDeclarationNamespaceReservesFixedNames(t *testing.T) {
	for _, name := range []string{"Name", "Glyph", "Config", "SpriteURL", "Glyphs", "Lookup", "Icon"} {
		t.Run(name, func(t *testing.T) {
			selected := []selectedAsset{{
				Asset:  iconcatalog.Asset{CanonicalName: "brand-example"},
				goName: name,
			}}
			err := validateGeneratedNamespace(selected)
			if err == nil || !strings.Contains(err.Error(), `go declaration collision "`+name+`"`) {
				t.Fatalf("validateGeneratedNamespace() error = %v, want reserved-name collision", err)
			}
		})
	}
}

func TestStageAndPublishRejectsConcurrentDestination(t *testing.T) {
	for _, kind := range []string{"file", "directory", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			parent := t.TempDir()
			output := filepath.Join(parent, "pack")
			location := outputLocation{output: output, parent: parent, base: "pack"}
			target := filepath.Join(parent, "attacker-target")
			files := map[string][]byte{"manifest.json": []byte("generated\n")}

			err := stageAndPublishWithHook(location, files, func() error {
				return createConcurrentDestination(kind, output, target)
			})
			if err == nil {
				t.Fatal("stageAndPublishWithHook() succeeded after concurrent destination creation")
			}
			assertConcurrentDestinationPreserved(t, kind, output, target)
			assertNoStagedDirectory(t, parent)
		})
	}
}

func createConcurrentDestination(kind, output, target string) error {
	switch kind {
	case "file":
		return os.WriteFile(output, []byte("attacker\n"), 0o644)
	case "directory":
		return os.Mkdir(output, 0o755)
	case "symlink":
		if err := os.WriteFile(target, []byte("attacker target\n"), 0o644); err != nil {
			return err
		}
		return os.Symlink(target, output)
	default:
		return fmt.Errorf("unknown destination kind %q", kind)
	}
}

func assertConcurrentDestinationPreserved(t *testing.T, kind, output, target string) {
	t.Helper()
	info, err := os.Lstat(output)
	if err != nil {
		t.Fatalf("concurrent destination missing after rejected publish: %v", err)
	}
	switch kind {
	case "file":
		contents, readErr := os.ReadFile(output)
		if readErr != nil || string(contents) != "attacker\n" {
			t.Fatalf("concurrent file changed: contents=%q error=%v", contents, readErr)
		}
	case "directory":
		if !info.IsDir() {
			t.Fatalf("concurrent directory replaced by mode %v", info.Mode())
		}
	case "symlink":
		link, readErr := os.Readlink(output)
		if info.Mode()&os.ModeSymlink == 0 || readErr != nil || link != target {
			t.Fatalf("concurrent symbolic link changed: target=%q mode=%v error=%v", link, info.Mode(), readErr)
		}
	}
}

func assertNoStagedDirectory(t *testing.T, parent string) {
	t.Helper()
	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".pack.tmp-") {
			t.Fatalf("staged directory leaked after rejected publish: %s", entry.Name())
		}
	}
}

func TestArchiveSnapshotRejectsOversizedAndGrowingSources(t *testing.T) {
	for _, test := range []struct {
		name         string
		contents     []byte
		declaredSize int64
		want         string
	}{
		{"oversized", nil, maxArchiveSnapshotBytes + 1, "snapshot limit"},
		{"growing", []byte("grew"), 3, "changed size while snapshotting"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var snapshot bytes.Buffer
			_, _, err := copyArchiveSnapshot(bytes.NewReader(test.contents), &snapshot, test.declaredSize)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("copyArchiveSnapshot() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestVerifiedArchiveSnapshotSurvivesPathReplacement(t *testing.T) {
	for _, format := range []string{"tar.gz", "zip"} {
		t.Run(format, func(t *testing.T) {
			opts := syntheticRelease(t)
			directory := t.TempDir()
			archive := filepath.Join(directory, "araihu-assets-v0.2.0."+format)
			replacement := filepath.Join(directory, "replacement."+format)
			writeReleaseArchive(t, opts.ReleaseRoot, archive, nil)
			writeReleaseArchive(t, opts.ReleaseRoot, replacement, map[string][]byte{
				"replacement-only.txt": []byte("unverified replacement\n"),
			})
			verifiedBytes, err := os.ReadFile(archive)
			if err != nil {
				t.Fatal(err)
			}
			opts.ReleaseRoot = ""
			opts.ReleaseArchive = archive
			opts.ArchiveSHA256 = hashBytes(verifiedBytes)

			boundary, err := openReleaseWithArchiveVerifiedHook(opts, func() error {
				return os.Rename(replacement, archive)
			})
			if err != nil {
				t.Fatal(err)
			}
			defer boundary.cleanup()
			if _, err := os.Lstat(filepath.Join(boundary.root, "replacement-only.txt")); !os.IsNotExist(err) {
				t.Fatalf("replacement-only archive member materialized: %v", err)
			}
		})
	}
}

func TestVerifiedArchiveSnapshotSurvivesSameInodeRewrite(t *testing.T) {
	for _, format := range []string{"tar.gz", "zip"} {
		t.Run(format, func(t *testing.T) {
			opts := syntheticRelease(t)
			directory := t.TempDir()
			archive := filepath.Join(directory, "araihu-assets-v0.2.0."+format)
			replacement := filepath.Join(directory, "replacement."+format)
			writeReleaseArchive(t, opts.ReleaseRoot, archive, nil)
			writeReleaseArchive(t, opts.ReleaseRoot, replacement, map[string][]byte{
				"replacement-only.txt": []byte("unverified replacement\n"),
			})
			replacementBytes, err := os.ReadFile(replacement)
			if err != nil {
				t.Fatal(err)
			}
			padFileToSize(t, archive, int64(len(replacementBytes)))
			verifiedBytes, err := os.ReadFile(archive)
			if err != nil {
				t.Fatal(err)
			}
			if len(verifiedBytes) != len(replacementBytes) {
				t.Fatalf("archive sizes differ: verified=%d replacement=%d", len(verifiedBytes), len(replacementBytes))
			}
			originalInfo, err := os.Stat(archive)
			if err != nil {
				t.Fatal(err)
			}
			opts.ReleaseRoot = ""
			opts.ReleaseArchive = archive
			opts.ArchiveSHA256 = hashBytes(verifiedBytes)

			boundary, err := openReleaseWithArchiveVerifiedHook(opts, func() error {
				file, err := os.OpenFile(archive, os.O_WRONLY|os.O_TRUNC, 0)
				if err != nil {
					return err
				}
				if _, err := file.Write(replacementBytes); err != nil {
					_ = file.Close()
					return err
				}
				if err := file.Close(); err != nil {
					return err
				}
				currentInfo, err := os.Stat(archive)
				if err != nil {
					return err
				}
				if !os.SameFile(originalInfo, currentInfo) {
					return fmt.Errorf("test rewrite replaced inode")
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			defer boundary.cleanup()
			if _, err := os.Lstat(filepath.Join(boundary.root, "replacement-only.txt")); !os.IsNotExist(err) {
				t.Fatalf("same-inode replacement archive member materialized: %v", err)
			}
		})
	}
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

func TestExtractArchiveRejectsTooManyDirectoryEntries(t *testing.T) {
	for _, format := range []string{"tar.gz", "zip"} {
		t.Run(format, func(t *testing.T) {
			archive := filepath.Join(t.TempDir(), "too-many."+format)
			writeDirectoryArchive(t, archive, maxArchiveFiles+1)
			destination := t.TempDir()

			err := extractArchive(archive, destination)
			if err == nil || !strings.Contains(err.Error(), "safe extraction budget") {
				t.Fatalf("extractArchive() error = %v, want entry-budget rejection", err)
			}
			entries, readErr := os.ReadDir(destination)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if len(entries) != 0 {
				t.Fatalf("rejected directory-only archive wrote destination entries: %v", entries)
			}
		})
	}
}

func TestExtractArchiveRejectsNonPortableMemberPaths(t *testing.T) {
	members := []struct {
		name string
		path string
	}{
		{name: "parent traversal", path: "../escape"},
		{name: "absolute", path: "/escape"},
		{name: "drive absolute", path: "C:/escape"},
		{name: "drive relative", path: "C:escape"},
		{name: "backslash traversal", path: `..\escape`},
		{name: "backslash separator", path: `nested\escape`},
		{name: "UNC", path: `\\server\share\escape`},
	}
	for _, format := range []string{"tar.gz", "zip"} {
		for _, member := range members {
			t.Run(format+"/"+member.name, func(t *testing.T) {
				archive := filepath.Join(t.TempDir(), "bad."+format)
				writeSingleMemberArchive(t, archive, member.path, []byte("escape"))
				destination := t.TempDir()

				err := extractArchive(archive, destination)
				if err == nil || !strings.Contains(err.Error(), "unsafe archive member") {
					t.Fatalf("extractArchive() error = %v, want unsafe-member rejection", err)
				}
				entries, readErr := os.ReadDir(destination)
				if readErr != nil {
					t.Fatal(readErr)
				}
				if len(entries) != 0 {
					t.Fatalf("rejected archive wrote destination entries: %v", entries)
				}
			})
		}
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
  "release": "v0.2.0-test",
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
		SchemaVersion: 1, Release: "v0.2.0-test", IdentityRevision: 11, RuntimeVersion: 1,
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
		Release:           "v0.2.0-test",
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

func writeReleaseArchive(t *testing.T, root, archive string, extra map[string][]byte) {
	t.Helper()
	files := readFixtureTree(t, root)
	maps.Copy(files, extra)
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	switch {
	case strings.HasSuffix(archive, ".tar.gz"):
		writeTarMembers(t, file, names, files)
	case strings.HasSuffix(archive, ".zip"):
		writeZipMembers(t, file, names, files)
	default:
		t.Fatalf("unsupported test archive %q", archive)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func readFixtureTree(t *testing.T, root string) map[string][]byte {
	t.Helper()
	files := map[string][]byte{}
	err := filepath.WalkDir(root, func(filename string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		relative, err := filepath.Rel(root, filename)
		if err != nil {
			return err
		}
		contents, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(relative)] = contents
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func writeTarMembers(t *testing.T, file *os.File, names []string, files map[string][]byte) {
	t.Helper()
	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)
	for _, name := range names {
		contents := files[name]
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(contents)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(contents); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
}

func writeDirectoryArchive(t *testing.T, archive string, count int) {
	t.Helper()
	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	switch {
	case strings.HasSuffix(archive, ".tar.gz"):
		gz := gzip.NewWriter(file)
		tw := tar.NewWriter(gz)
		for index := range count {
			name := fmt.Sprintf("directory-%05d/", index)
			if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Typeflag: tar.TypeDir}); err != nil {
				t.Fatal(err)
			}
		}
		if err := tw.Close(); err != nil {
			t.Fatal(err)
		}
		if err := gz.Close(); err != nil {
			t.Fatal(err)
		}
	case strings.HasSuffix(archive, ".zip"):
		zw := zip.NewWriter(file)
		for index := range count {
			name := fmt.Sprintf("directory-%05d/", index)
			if _, err := zw.Create(name); err != nil {
				t.Fatal(err)
			}
		}
		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatalf("unsupported test archive %q", archive)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func writeSingleMemberArchive(t *testing.T, archive, name string, contents []byte) {
	t.Helper()
	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	switch {
	case strings.HasSuffix(archive, ".tar.gz"):
		gz := gzip.NewWriter(file)
		tw := tar.NewWriter(gz)
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(contents)), Typeflag: tar.TypeReg}); err != nil {
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
	case strings.HasSuffix(archive, ".zip"):
		zw := zip.NewWriter(file)
		member, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := member.Write(contents); err != nil {
			t.Fatal(err)
		}
		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatalf("unsupported test archive %q", archive)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func writeZipMembers(t *testing.T, file *os.File, names []string, files map[string][]byte) {
	t.Helper()
	zw := zip.NewWriter(file)
	for _, name := range names {
		member, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := member.Write(files[name]); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}

func padFileToSize(t *testing.T, filename string, size int64) {
	t.Helper()
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() > size {
		t.Fatalf("cannot pad %q from %d down to %d bytes", filename, info.Size(), size)
	}
	if err := os.Truncate(filename, size); err != nil {
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
