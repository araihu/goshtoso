package iconpack

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/araihu/goshtoso/internal/iconcatalog"
)

const (
	maxArchiveFiles         = 10_000
	maxArchiveBytes         = int64(1 << 30)
	maxMemberBytes          = int64(64 << 20)
	maxArchiveSnapshotBytes = int64(64 << 20)

	assetsV020Release           = "v0.2.0"
	assetsV020TarGzipSHA256     = "5d7d691e22d4071507b0bf2248713d7008adf57c18840cfd46e20901db0b78e5"
	assetsV020ZipSHA256         = "881094d3d161b79904fcfad320c26d947c9a1e526ee0b69ce8a2d04c3ff4b1b0"
	assetsV020CatalogSHA256     = "a0e8e5c8928e37de979ce9a60f3d66fad1aa1b4c7d2904f9275f0be9932a33d6"
	assetsV020ReleaseJSONSHA256 = "77c696ae5eceb5e7bc11d19affb7c2c7b7e8afc6414882b9b059239e315f2260"
	assetsV020ChecksumsSHA256   = "334005c77622250a1e827b9472161cd6e56c82d487fc0d44023d49261f8dbee5"
)

var (
	digestRE  = regexp.MustCompile(`^[0-9a-f]{64}$`)
	releaseRE = regexp.MustCompile(`^v(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)(?:[-+][0-9A-Za-z.-]+)?$`)
)

type releaseBoundary struct {
	root          string
	cleanup       func()
	sourceKind    string
	archiveSHA256 string
	catalog       iconcatalog.Catalog
	release       releaseDocument
	checksums     map[string]string
	files         map[string][]byte
	generic       *genericPack
	muamba        *muambaPack
}

func openSourceBoundary(ctx context.Context, opts Options) (releaseBoundary, error) {
	if opts.ConfigPath != "" {
		return openMuambaPack(ctx, opts)
	}
	if opts.SourceManifest != "" {
		return openGenericPack(opts)
	}
	return openRelease(opts)
}

type releaseDocument struct {
	SchemaVersion    int           `json:"schemaVersion"`
	Release          string        `json:"release"`
	IdentityRevision int           `json:"identityRevision"`
	RuntimeVersion   int           `json:"runtimeVersion"`
	CatalogSHA256    string        `json:"catalogSha256"`
	ThemesSHA256     string        `json:"themesSha256"`
	CampaignsSHA256  string        `json:"campaignsSha256"`
	Files            []releaseFile `json:"files"`
}

type releaseFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

func openRelease(opts Options) (releaseBoundary, error) {
	return openReleaseWithArchiveVerifiedHook(opts, nil)
}

func openReleaseWithArchiveVerifiedHook(opts Options, afterArchiveVerified func() error) (releaseBoundary, error) {
	if (opts.ReleaseRoot == "") == (opts.ReleaseArchive == "") {
		return releaseBoundary{}, fmt.Errorf("exactly one of release-root or release-archive is required")
	}
	if !releaseRE.MatchString(opts.Release) {
		return releaseBoundary{}, fmt.Errorf("invalid expected release %q", opts.Release)
	}
	for _, field := range []struct{ name, value string }{
		{"catalog-sha256", opts.CatalogSHA256},
		{"release-json-sha256", opts.ReleaseJSONSHA256},
		{"checksums-sha256", opts.ChecksumsSHA256},
	} {
		if !digestRE.MatchString(field.value) {
			return releaseBoundary{}, fmt.Errorf("%s must be a lowercase SHA-256", field.name)
		}
	}
	if err := validateTrustedReleaseIdentity(opts); err != nil {
		return releaseBoundary{}, err
	}

	root := opts.ReleaseRoot
	cleanup := func() {}
	sourceKind := "release-root"
	archiveSHA256 := ""
	if opts.ReleaseArchive != "" {
		if !digestRE.MatchString(opts.ArchiveSHA256) {
			return releaseBoundary{}, fmt.Errorf("archive-sha256 must be a lowercase SHA-256")
		}
		sourceKind = "release-archive"
		archiveSHA256 = opts.ArchiveSHA256
		var temporary string
		err := withVerifiedReleaseArchive(opts.ReleaseArchive, opts.ArchiveSHA256, afterArchiveVerified, func(archive *os.File, size int64) error {
			var err error
			temporary, err = os.MkdirTemp("", "goshtoso-iconpack-release-*")
			if err != nil {
				return fmt.Errorf("create archive extraction root: %w", err)
			}
			cleanup = func() { _ = os.RemoveAll(temporary) }
			return extractOpenedArchive(archive, opts.ReleaseArchive, size, temporary)
		})
		if err != nil {
			cleanup()
			return releaseBoundary{}, fmt.Errorf("verify release archive: %w", err)
		}
		root = temporary
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		cleanup()
		return releaseBoundary{}, fmt.Errorf("resolve release root: %w", err)
	}
	boundary := releaseBoundary{root: absRoot, cleanup: cleanup, sourceKind: sourceKind, archiveSHA256: archiveSHA256}
	if err := boundary.verify(opts); err != nil {
		cleanup()
		return releaseBoundary{}, err
	}
	return boundary, nil
}

func validateTrustedReleaseIdentity(opts Options) error {
	if opts.Release != assetsV020Release {
		return nil
	}
	for _, expected := range []struct {
		name, got, want string
	}{
		{"catalog-sha256", opts.CatalogSHA256, assetsV020CatalogSHA256},
		{"release-json-sha256", opts.ReleaseJSONSHA256, assetsV020ReleaseJSONSHA256},
		{"checksums-sha256", opts.ChecksumsSHA256, assetsV020ChecksumsSHA256},
	} {
		if expected.got != expected.want {
			return fmt.Errorf("v0.2.0 %s does not match the frozen Assets candidate boundary", expected.name)
		}
	}
	if opts.ReleaseArchive != "" {
		expected, ok := assetsV020ArchiveDigest(opts.ReleaseArchive)
		if !ok {
			return fmt.Errorf("v0.2.0 archive must use .tar.gz, .tgz, or .zip kind")
		}
		if opts.ArchiveSHA256 != expected {
			return fmt.Errorf("v0.2.0 archive-sha256 does not match the frozen Assets candidate %s boundary", archiveKind(opts.ReleaseArchive))
		}
	}
	return nil
}

func assetsV020ArchiveDigest(filename string) (string, bool) {
	switch archiveKind(filename) {
	case "tar.gz":
		return assetsV020TarGzipSHA256, true
	case "zip":
		return assetsV020ZipSHA256, true
	default:
		return "", false
	}
}

func archiveKind(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return "tar.gz"
	case strings.HasSuffix(lower, ".zip"):
		return "zip"
	default:
		return "unsupported"
	}
}

func (boundary *releaseBoundary) verify(opts Options) error {
	if err := validateReleaseRoot(boundary.root); err != nil {
		return err
	}
	root, err := os.OpenRoot(boundary.root)
	if err != nil {
		return fmt.Errorf("open contained release root: %w", err)
	}
	defer func() { _ = root.Close() }()
	if err := validateOpenedReleaseRoot(boundary.root, root); err != nil {
		return err
	}
	pinned, err := verifyPinnedReleaseFiles(root, opts)
	if err != nil {
		return err
	}
	checksums, files, err := verifyReleaseChecksums(root, pinned["checksums.txt"])
	if err != nil {
		return err
	}
	boundary.checksums = checksums
	boundary.files = files
	release, err := loadReleaseDocument(files["release.json"])
	if err != nil {
		return err
	}
	if err := validateReleaseDocument(release, checksums, files, opts); err != nil {
		return err
	}
	boundary.release = release
	catalog, err := loadReleaseCatalog(files["catalog.json"], opts)
	if err != nil {
		return err
	}
	boundary.catalog = catalog
	return nil
}

func validateReleaseRoot(root string) error {
	info, err := os.Lstat(root)
	if err != nil {
		return fmt.Errorf("inspect release root: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("release root must be a real directory")
	}
	for segment := range strings.SplitSeq(filepath.ToSlash(root), "/") {
		switch strings.ToLower(segment) {
		case ".git", "internal", "acquisition", "vendor":
			return fmt.Errorf("release root crosses forbidden source-tree segment %q", segment)
		}
	}
	for _, marker := range []string{"go.mod", ".git"} {
		if _, err := os.Lstat(filepath.Join(root, marker)); err == nil {
			return fmt.Errorf("release root is a source checkout, not an extracted release boundary")
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("inspect release root marker %q: %w", marker, err)
		}
	}
	return nil
}

func validateOpenedReleaseRoot(name string, root *os.Root) error {
	pathInfo, err := os.Lstat(name)
	if err != nil {
		return fmt.Errorf("reinspect release root: %w", err)
	}
	if !pathInfo.IsDir() || pathInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("release root must remain a real directory")
	}
	openedInfo, err := root.Stat(".")
	if err != nil {
		return fmt.Errorf("inspect opened release root: %w", err)
	}
	if !os.SameFile(pathInfo, openedInfo) {
		return fmt.Errorf("release root changed while opening")
	}
	return nil
}

func verifyPinnedReleaseFiles(root *os.Root, opts Options) (map[string][]byte, error) {
	pinned := make(map[string][]byte, 3)
	for _, expected := range []struct{ name, digest string }{
		{"catalog.json", opts.CatalogSHA256},
		{"release.json", opts.ReleaseJSONSHA256},
		{"checksums.txt", opts.ChecksumsSHA256},
	} {
		contents, err := readContainedRegularFile(root, expected.name)
		if err != nil {
			return nil, fmt.Errorf("verify %s: %w", expected.name, err)
		}
		got := hashBytes(contents)
		if got != expected.digest {
			return nil, fmt.Errorf("%s SHA-256 mismatch: got %s, want %s", expected.name, got, expected.digest)
		}
		pinned[expected.name] = contents
	}
	return pinned, nil
}

func verifyReleaseChecksums(root *os.Root, checksumsBytes []byte) (map[string]string, map[string][]byte, error) {
	checksums, err := loadChecksums(checksumsBytes)
	if err != nil {
		return nil, nil, err
	}
	if len(checksums) > maxArchiveFiles {
		return nil, nil, fmt.Errorf("release inventory exceeds %d-file limit", maxArchiveFiles)
	}
	files := make(map[string][]byte, len(checksums))
	var total int64
	for relative, expected := range checksums {
		contents, err := readContainedRegularFile(root, relative)
		if err != nil {
			return nil, nil, fmt.Errorf("verify checksums entry %q: %w", relative, err)
		}
		got := hashBytes(contents)
		if got != expected {
			return nil, nil, fmt.Errorf("checksums entry %q mismatch: got %s, want %s", relative, got, expected)
		}
		total += int64(len(contents))
		if total > maxArchiveBytes {
			return nil, nil, fmt.Errorf("release inventory exceeds %d-byte limit", maxArchiveBytes)
		}
		files[relative] = contents
	}
	return checksums, files, nil
}

func loadReleaseDocument(releaseBytes []byte) (releaseDocument, error) {
	var release releaseDocument
	decoder := json.NewDecoder(strings.NewReader(string(releaseBytes)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&release); err != nil {
		return releaseDocument{}, fmt.Errorf("decode release.json: %w", err)
	}
	if err := requireJSONEOF(decoder, "release.json"); err != nil {
		return releaseDocument{}, err
	}
	return release, nil
}

func validateReleaseDocument(release releaseDocument, checksums map[string]string, files map[string][]byte, opts Options) error {
	if release.SchemaVersion != 1 || release.IdentityRevision != 11 || release.RuntimeVersion != 1 {
		return fmt.Errorf("unsupported release metadata schema=%d identityRevision=%d runtimeVersion=%d", release.SchemaVersion, release.IdentityRevision, release.RuntimeVersion)
	}
	if release.Release != opts.Release {
		return fmt.Errorf("release.json release %q does not match expected %q", release.Release, opts.Release)
	}
	if release.CatalogSHA256 != opts.CatalogSHA256 {
		return fmt.Errorf("release.json catalogSha256 %s does not match expected %s", release.CatalogSHA256, opts.CatalogSHA256)
	}
	return verifyReleaseInventory(release, checksums, files)
}

func loadReleaseCatalog(catalogBytes []byte, opts Options) (iconcatalog.Catalog, error) {
	catalog, err := iconcatalog.Load(bytes.NewReader(catalogBytes))
	if err != nil {
		return iconcatalog.Catalog{}, fmt.Errorf("validate catalog.json: %w", err)
	}
	if catalog.SchemaVersion != 2 {
		return iconcatalog.Catalog{}, fmt.Errorf("unsupported iconpack catalog schemaVersion %d: want 2", catalog.SchemaVersion)
	}
	if catalog.Release != opts.Release || catalog.Hash != opts.CatalogSHA256 {
		return iconcatalog.Catalog{}, fmt.Errorf("catalog identity does not match expected release and SHA-256")
	}
	return catalog, nil
}

func verifyReleaseInventory(release releaseDocument, checksums map[string]string, files map[string][]byte) error {
	if len(release.Files) == 0 {
		return fmt.Errorf("release.json files must not be empty")
	}
	if err := verifyReleaseDocumentIdentities(release, checksums); err != nil {
		return err
	}
	if err := verifyReleaseFileOrder(release.Files); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(release.Files))
	for index, file := range release.Files {
		if err := safeRelativePath(file.Path); err != nil {
			return fmt.Errorf("release.json file %d: %w", index, err)
		}
		if !digestRE.MatchString(file.SHA256) || file.Size < 0 {
			return fmt.Errorf("release.json file %q has invalid hash or size", file.Path)
		}
		if _, exists := seen[file.Path]; exists {
			return fmt.Errorf("release.json has duplicate file %q", file.Path)
		}
		seen[file.Path] = struct{}{}
		checksum, ok := checksums[file.Path]
		if !ok || checksum != file.SHA256 {
			return fmt.Errorf("release.json file %q disagrees with checksums.txt", file.Path)
		}
		contents, ok := files[file.Path]
		if !ok || hashBytes(contents) != file.SHA256 || int64(len(contents)) != file.Size {
			return fmt.Errorf("release.json file %q content does not match inventory", file.Path)
		}
	}
	if len(checksums) != len(release.Files)+1 || checksums["release.json"] == "" {
		return fmt.Errorf("checksums.txt must contain release inventory plus release.json exactly")
	}
	return nil
}

func verifyReleaseDocumentIdentities(release releaseDocument, checksums map[string]string) error {
	for _, document := range []struct {
		path, digest string
	}{
		{"catalog.json", release.CatalogSHA256},
		{"themes.json", release.ThemesSHA256},
		{"campaigns.json", release.CampaignsSHA256},
	} {
		if !digestRE.MatchString(document.digest) || checksums[document.path] != document.digest {
			return fmt.Errorf("release.json %s identity disagrees with checksums.txt", document.path)
		}
	}
	return nil
}

func verifyReleaseFileOrder(files []releaseFile) error {
	documents := []string{"catalog.json", "themes.json", "campaigns.json"}
	if len(files) < len(documents) {
		return fmt.Errorf("release.json files must begin with catalog, themes, and campaigns documents")
	}
	for index, expected := range documents {
		if files[index].Path != expected {
			return fmt.Errorf("release.json file %d is %q: want %q", index, files[index].Path, expected)
		}
	}
	for index := len(documents) + 1; index < len(files); index++ {
		if files[index].Path <= files[index-1].Path {
			return fmt.Errorf("release.json remaining files are not strictly sorted")
		}
	}
	return nil
}

func loadChecksums(contents []byte) (map[string]string, error) {
	checksums := map[string]string{}
	caseFolded := map[string]string{}
	previous := ""
	scanner := bufio.NewScanner(bytes.NewReader(contents))
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 67 || line[64:66] != "  " {
			return nil, fmt.Errorf("invalid checksums.txt line %q", line)
		}
		digest, relative := line[:64], line[66:]
		if !digestRE.MatchString(digest) {
			return nil, fmt.Errorf("invalid checksum for %q", relative)
		}
		if err := safeRelativePath(relative); err != nil {
			return nil, fmt.Errorf("checksums.txt: %w", err)
		}
		if previous != "" && relative <= previous {
			return nil, fmt.Errorf("checksums.txt paths are not strictly sorted")
		}
		previous = relative
		if _, exists := checksums[relative]; exists {
			return nil, fmt.Errorf("duplicate checksums path %q", relative)
		}
		folded := strings.ToLower(relative)
		if first, exists := caseFolded[folded]; exists {
			return nil, fmt.Errorf("case-folded checksums collision %q and %q", first, relative)
		}
		caseFolded[folded] = relative
		checksums[relative] = digest
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read checksums.txt: %w", err)
	}
	if len(checksums) == 0 {
		return nil, fmt.Errorf("checksums.txt must not be empty")
	}
	return checksums, nil
}

func readContainedRegularFile(root *os.Root, relative string) ([]byte, error) {
	if err := safeRelativePath(relative); err != nil {
		return nil, err
	}
	if err := validateContainedPath(root, relative); err != nil {
		return nil, err
	}
	file, err := root.Open(relative)
	if err != nil {
		return nil, err
	}
	openedInfo, statErr := file.Stat()
	if statErr != nil {
		_ = file.Close()
		return nil, statErr
	}
	if !openedInfo.Mode().IsRegular() {
		_ = file.Close()
		return nil, fmt.Errorf("release path %q is not a regular file", relative)
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
		return nil, fmt.Errorf("release file %q exceeds %d-byte limit", relative, maxMemberBytes)
	}
	if err := validateContainedPath(root, relative); err != nil {
		return nil, fmt.Errorf("release path %q changed while reading: %w", relative, err)
	}
	currentInfo, err := root.Lstat(relative)
	if err != nil {
		return nil, fmt.Errorf("inspect release path %q after reading: %w", relative, err)
	}
	if !os.SameFile(openedInfo, currentInfo) {
		return nil, fmt.Errorf("release path %q changed while reading", relative)
	}
	return contents, nil
}

func validateContainedPath(root *os.Root, relative string) error {
	segments := strings.Split(relative, "/")
	for index := range segments {
		current := strings.Join(segments[:index+1], "/")
		info, err := root.Lstat(current)
		if err != nil {
			return fmt.Errorf("inspect release path %q: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("release path %q contains symbolic link %q", relative, current)
		}
		if index < len(segments)-1 {
			if !info.IsDir() {
				return fmt.Errorf("release path %q has non-directory parent %q", relative, current)
			}
			continue
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("release path %q is not a regular file", relative)
		}
	}
	return nil
}

func openRegularFile(filename string) (*os.File, error) {
	pathInfo, err := os.Lstat(filename)
	if err != nil {
		return nil, err
	}
	if !pathInfo.Mode().IsRegular() || pathInfo.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("not a regular file")
	}
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	openedInfo, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	currentInfo, err := os.Lstat(filename)
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	if !openedInfo.Mode().IsRegular() || currentInfo.Mode()&os.ModeSymlink != 0 || !currentInfo.Mode().IsRegular() || !os.SameFile(pathInfo, openedInfo) || !os.SameFile(openedInfo, currentInfo) {
		_ = file.Close()
		return nil, fmt.Errorf("archive path changed while opening")
	}
	return file, nil
}

func withVerifiedReleaseArchive(filename, expectedDigest string, afterVerified func() error, consume func(*os.File, int64) error) (err error) {
	source, err := openRegularFile(filename)
	if err != nil {
		return err
	}
	sourceClosed := false
	defer func() {
		if !sourceClosed {
			_ = source.Close()
		}
	}()
	sourceInfo, err := source.Stat()
	if err != nil {
		return fmt.Errorf("inspect release archive before snapshot: %w", err)
	}
	snapshot, err := os.CreateTemp("", "goshtoso-iconpack-archive-snapshot-*")
	if err != nil {
		return fmt.Errorf("create private archive snapshot: %w", err)
	}
	snapshotName := snapshot.Name()
	defer func() {
		_ = snapshot.Close()
		_ = os.Remove(snapshotName)
	}()
	digest, size, err := copyArchiveSnapshot(source, snapshot, sourceInfo.Size())
	if err != nil {
		return err
	}
	closeErr := source.Close()
	sourceClosed = true
	if closeErr != nil {
		return fmt.Errorf("close release archive: %w", closeErr)
	}
	if digest != expectedDigest {
		return fmt.Errorf("release archive SHA-256 mismatch: got %s, want %s", digest, expectedDigest)
	}
	if _, err := snapshot.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("rewind verified archive snapshot: %w", err)
	}
	if afterVerified != nil {
		if err := afterVerified(); err != nil {
			return fmt.Errorf("after archive verification: %w", err)
		}
	}
	return consume(snapshot, size)
}

func copyArchiveSnapshot(source io.Reader, snapshot io.Writer, declaredSize int64) (string, int64, error) {
	if declaredSize < 0 || declaredSize > maxArchiveSnapshotBytes {
		return "", 0, fmt.Errorf("release archive exceeds %d-byte snapshot limit", maxArchiveSnapshotBytes)
	}
	hash := sha256.New()
	limited := &io.LimitedReader{R: source, N: maxArchiveSnapshotBytes + 1}
	size, err := io.Copy(io.MultiWriter(hash, snapshot), limited)
	if err != nil {
		return "", 0, fmt.Errorf("copy release archive into private snapshot: %w", err)
	}
	if size > maxArchiveSnapshotBytes {
		return "", 0, fmt.Errorf("release archive exceeds %d-byte snapshot limit", maxArchiveSnapshotBytes)
	}
	if size != declaredSize {
		return "", 0, fmt.Errorf("release archive changed size while snapshotting: got %d, expected %d", size, declaredSize)
	}
	return hex.EncodeToString(hash.Sum(nil)), size, nil
}

func safeRelativePath(relative string) error {
	if relative == "" || strings.Contains(relative, "\\") || strings.ContainsRune(relative, '\x00') || strings.HasPrefix(relative, "/") || path.Clean(relative) != relative || hasWindowsDrivePrefix(relative) {
		return fmt.Errorf("unsafe relative path %q", relative)
	}
	for segment := range strings.SplitSeq(relative, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("unsafe relative path %q", relative)
		}
	}
	return nil
}

func hasWindowsDrivePrefix(name string) bool {
	if len(name) < 2 || name[1] != ':' {
		return false
	}
	first := name[0]
	return first >= 'A' && first <= 'Z' || first >= 'a' && first <= 'z'
}

func requireJSONEOF(decoder *json.Decoder, name string) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("%s contains more than one JSON value", name)
		}
		return fmt.Errorf("decode %s: %w", name, err)
	}
	return nil
}

func extractArchive(archivePath, destination string) error {
	archive, err := openRegularFile(archivePath)
	if err != nil {
		return fmt.Errorf("open release archive: %w", err)
	}
	defer func() { _ = archive.Close() }()
	info, err := archive.Stat()
	if err != nil {
		return fmt.Errorf("inspect release archive: %w", err)
	}
	return extractOpenedArchive(archive, archivePath, info.Size(), destination)
}

func extractOpenedArchive(archive *os.File, archivePath string, size int64, destination string) error {
	switch archiveKind(archivePath) {
	case "tar.gz":
		return extractTarGzip(archive, destination)
	case "zip":
		return extractZip(archive, size, destination)
	default:
		return fmt.Errorf("unsupported release archive type: want .tar.gz, .tgz, or .zip")
	}
}

func extractTarGzip(archive io.ReadSeeker, destination string) error {
	if err := countTarEntries(archive); err != nil {
		return err
	}
	if _, err := archive.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("rewind tar archive: %w", err)
	}
	gzipReader, err := gzip.NewReader(archive)
	if err != nil {
		return fmt.Errorf("open gzip stream: %w", err)
	}
	defer func() { _ = gzipReader.Close() }()
	reader := tar.NewReader(gzipReader)
	seen := map[string]string{}
	var files int
	var total int64
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar archive: %w", err)
		}
		relative, err := recordArchivePath(seen, header.Name)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("create archive directory: %w", err)
			}
		case tar.TypeReg:
			files++
			total += header.Size
			if err := archiveBudget(files, header.Size, total); err != nil {
				return err
			}
			if err := writeArchiveMember(target, io.LimitReader(reader, header.Size), header.Size); err != nil {
				return err
			}
		default:
			return fmt.Errorf("archive member %q has forbidden type %d", header.Name, header.Typeflag)
		}
	}
	return nil
}

func extractZip(archive io.ReaderAt, size int64, destination string) error {
	reader, err := zip.NewReader(archive, size)
	if err != nil {
		return fmt.Errorf("open zip archive: %w", err)
	}
	if err := archiveEntryBudget(len(reader.File)); err != nil {
		return err
	}
	seen := map[string]string{}
	var files int
	var total int64
	for _, member := range reader.File {
		relative, err := recordArchivePath(seen, member.Name)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if member.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("create archive directory: %w", err)
			}
			continue
		}
		if !member.Mode().IsRegular() {
			return fmt.Errorf("archive member %q is not a regular file", member.Name)
		}
		files++
		size := int64(member.UncompressedSize64)
		total += size
		if err := archiveBudget(files, size, total); err != nil {
			return err
		}
		input, err := member.Open()
		if err != nil {
			return fmt.Errorf("open archive member %q: %w", member.Name, err)
		}
		writeErr := writeArchiveMember(target, input, size)
		closeErr := input.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return fmt.Errorf("close archive member %q: %w", member.Name, closeErr)
		}
	}
	return nil
}

func recordArchivePath(seen map[string]string, name string) (string, error) {
	portable := strings.TrimSuffix(name, "/")
	if err := safeRelativePath(portable); err != nil {
		return "", fmt.Errorf("unsafe archive member %q", name)
	}
	relative, err := filepath.Localize(portable)
	if err != nil || !filepath.IsLocal(relative) || filepath.IsAbs(relative) || filepath.VolumeName(relative) != "" {
		return "", fmt.Errorf("unsafe archive member %q", name)
	}
	folded := strings.ToLower(portable)
	if first, exists := seen[folded]; exists {
		return "", fmt.Errorf("duplicate or case-folded archive members %q and %q", first, name)
	}
	seen[folded] = name
	return relative, nil
}

func countTarEntries(archive io.ReadSeeker) error {
	gzipReader, err := gzip.NewReader(archive)
	if err != nil {
		return fmt.Errorf("open gzip stream: %w", err)
	}
	reader := tar.NewReader(gzipReader)
	entries := 0
	for {
		_, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = gzipReader.Close()
			return fmt.Errorf("read tar archive: %w", err)
		}
		entries++
		if err := archiveEntryBudget(entries); err != nil {
			_ = gzipReader.Close()
			return err
		}
	}
	if err := gzipReader.Close(); err != nil {
		return fmt.Errorf("close gzip stream: %w", err)
	}
	return nil
}

func archiveEntryBudget(entries int) error {
	if entries > maxArchiveFiles {
		return fmt.Errorf("release archive exceeds safe extraction budget")
	}
	return nil
}

func archiveBudget(files int, size, total int64) error {
	if files > maxArchiveFiles || size < 0 || size > maxMemberBytes || total > maxArchiveBytes {
		return fmt.Errorf("release archive exceeds safe extraction budget")
	}
	return nil
}

func writeArchiveMember(target string, input io.Reader, size int64) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create archive member parent: %w", err)
	}
	output, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create archive member %q: %w", target, err)
	}
	written, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return fmt.Errorf("write archive member %q: %w", target, copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close archive member %q: %w", target, closeErr)
	}
	if written != size {
		return fmt.Errorf("archive member %q size mismatch: got %d, want %d", target, written, size)
	}
	return nil
}
