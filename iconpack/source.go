package iconpack

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
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
	maxArchiveFiles = 10_000
	maxArchiveBytes = int64(1 << 30)
	maxMemberBytes  = int64(64 << 20)
)

var (
	digestRE  = regexp.MustCompile(`^[0-9a-f]{64}$`)
	releaseRE = regexp.MustCompile(`^v(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)(?:[-+][0-9A-Za-z.-]+)?$`)
)

type releaseBoundary struct {
	root      string
	cleanup   func()
	catalog   iconcatalog.Catalog
	release   releaseDocument
	checksums map[string]string
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

	root := opts.ReleaseRoot
	cleanup := func() {}
	if opts.ReleaseArchive != "" {
		if !digestRE.MatchString(opts.ArchiveSHA256) {
			return releaseBoundary{}, fmt.Errorf("archive-sha256 must be a lowercase SHA-256")
		}
		got, _, err := hashRegularFile(opts.ReleaseArchive)
		if err != nil {
			return releaseBoundary{}, fmt.Errorf("verify release archive: %w", err)
		}
		if got != opts.ArchiveSHA256 {
			return releaseBoundary{}, fmt.Errorf("release archive SHA-256 mismatch: got %s, want %s", got, opts.ArchiveSHA256)
		}
		temporary, err := os.MkdirTemp("", "goshtoso-iconpack-release-*")
		if err != nil {
			return releaseBoundary{}, fmt.Errorf("create archive extraction root: %w", err)
		}
		cleanup = func() { _ = os.RemoveAll(temporary) }
		if err := extractArchive(opts.ReleaseArchive, temporary); err != nil {
			cleanup()
			return releaseBoundary{}, err
		}
		root = temporary
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		cleanup()
		return releaseBoundary{}, fmt.Errorf("resolve release root: %w", err)
	}
	boundary := releaseBoundary{root: absRoot, cleanup: cleanup}
	if err := boundary.verify(opts); err != nil {
		cleanup()
		return releaseBoundary{}, err
	}
	return boundary, nil
}

func (boundary *releaseBoundary) verify(opts Options) error {
	if err := validateReleaseRoot(boundary.root); err != nil {
		return err
	}
	if err := verifyPinnedReleaseFiles(boundary.root, opts); err != nil {
		return err
	}
	checksums, err := verifyReleaseChecksums(boundary.root)
	if err != nil {
		return err
	}
	boundary.checksums = checksums
	release, err := loadReleaseDocument(boundary.root)
	if err != nil {
		return err
	}
	if err := validateReleaseDocument(boundary.root, release, checksums, opts); err != nil {
		return err
	}
	boundary.release = release
	catalog, err := loadReleaseCatalog(boundary.root, opts)
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

func verifyPinnedReleaseFiles(root string, opts Options) error {
	for _, expected := range []struct{ name, digest string }{
		{"catalog.json", opts.CatalogSHA256},
		{"release.json", opts.ReleaseJSONSHA256},
		{"checksums.txt", opts.ChecksumsSHA256},
	} {
		got, _, err := hashRegularFile(filepath.Join(root, expected.name))
		if err != nil {
			return fmt.Errorf("verify %s: %w", expected.name, err)
		}
		if got != expected.digest {
			return fmt.Errorf("%s SHA-256 mismatch: got %s, want %s", expected.name, got, expected.digest)
		}
	}
	return nil
}

func verifyReleaseChecksums(root string) (map[string]string, error) {
	checksums, err := loadChecksums(filepath.Join(root, "checksums.txt"))
	if err != nil {
		return nil, err
	}
	for relative, expected := range checksums {
		got, _, err := hashReleaseFile(root, relative)
		if err != nil {
			return nil, fmt.Errorf("verify checksums entry %q: %w", relative, err)
		}
		if got != expected {
			return nil, fmt.Errorf("checksums entry %q mismatch: got %s, want %s", relative, got, expected)
		}
	}
	return checksums, nil
}

func loadReleaseDocument(root string) (releaseDocument, error) {
	releaseBytes, err := os.ReadFile(filepath.Join(root, "release.json"))
	if err != nil {
		return releaseDocument{}, fmt.Errorf("read release.json: %w", err)
	}
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

func validateReleaseDocument(root string, release releaseDocument, checksums map[string]string, opts Options) error {
	if release.SchemaVersion != 1 || release.IdentityRevision != 11 || release.RuntimeVersion != 1 {
		return fmt.Errorf("unsupported release metadata schema=%d identityRevision=%d runtimeVersion=%d", release.SchemaVersion, release.IdentityRevision, release.RuntimeVersion)
	}
	if release.Release != opts.Release {
		return fmt.Errorf("release.json release %q does not match expected %q", release.Release, opts.Release)
	}
	if release.CatalogSHA256 != opts.CatalogSHA256 {
		return fmt.Errorf("release.json catalogSha256 %s does not match expected %s", release.CatalogSHA256, opts.CatalogSHA256)
	}
	return verifyReleaseInventory(root, release, checksums)
}

func loadReleaseCatalog(root string, opts Options) (iconcatalog.Catalog, error) {
	catalogFile, err := os.Open(filepath.Join(root, "catalog.json"))
	if err != nil {
		return iconcatalog.Catalog{}, fmt.Errorf("open catalog.json: %w", err)
	}
	catalog, loadErr := iconcatalog.Load(catalogFile)
	closeErr := catalogFile.Close()
	if loadErr != nil {
		return iconcatalog.Catalog{}, fmt.Errorf("validate catalog.json: %w", loadErr)
	}
	if closeErr != nil {
		return iconcatalog.Catalog{}, fmt.Errorf("close catalog.json: %w", closeErr)
	}
	if catalog.SchemaVersion != 2 {
		return iconcatalog.Catalog{}, fmt.Errorf("unsupported iconpack catalog schemaVersion %d: want 2", catalog.SchemaVersion)
	}
	if catalog.Release != opts.Release || catalog.Hash != opts.CatalogSHA256 {
		return iconcatalog.Catalog{}, fmt.Errorf("catalog identity does not match expected release and SHA-256")
	}
	return catalog, nil
}

func verifyReleaseInventory(root string, release releaseDocument, checksums map[string]string) error {
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
		got, size, err := hashReleaseFile(root, file.Path)
		if err != nil {
			return err
		}
		if got != file.SHA256 || size != file.Size {
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

func loadChecksums(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open checksums.txt: %w", err)
	}
	defer func() { _ = file.Close() }()
	checksums := map[string]string{}
	caseFolded := map[string]string{}
	previous := ""
	scanner := bufio.NewScanner(file)
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

func hashReleaseFile(root, relative string) (string, int64, error) {
	if err := safeRelativePath(relative); err != nil {
		return "", 0, err
	}
	return hashRegularFile(filepath.Join(root, filepath.FromSlash(relative)))
}

func hashRegularFile(filename string) (string, int64, error) {
	info, err := os.Lstat(filename)
	if err != nil {
		return "", 0, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return "", 0, fmt.Errorf("not a regular file")
	}
	file, err := os.Open(filename)
	if err != nil {
		return "", 0, err
	}
	hash := sha256.New()
	size, copyErr := io.Copy(hash, file)
	closeErr := file.Close()
	if copyErr != nil {
		return "", 0, copyErr
	}
	if closeErr != nil {
		return "", 0, closeErr
	}
	return hex.EncodeToString(hash.Sum(nil)), size, nil
}

func safeRelativePath(relative string) error {
	if relative == "" || strings.Contains(relative, "\\") || strings.ContainsRune(relative, '\x00') || strings.HasPrefix(relative, "/") || path.Clean(relative) != relative {
		return fmt.Errorf("unsafe relative path %q", relative)
	}
	for segment := range strings.SplitSeq(relative, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("unsafe relative path %q", relative)
		}
	}
	return nil
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
	switch {
	case strings.HasSuffix(archivePath, ".tar.gz"), strings.HasSuffix(archivePath, ".tgz"):
		return extractTarGzip(archivePath, destination)
	case strings.HasSuffix(archivePath, ".zip"):
		return extractZip(archivePath, destination)
	default:
		return fmt.Errorf("unsupported release archive type: want .tar.gz, .tgz, or .zip")
	}
}

func extractTarGzip(archivePath, destination string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open tar archive: %w", err)
	}
	defer func() { _ = file.Close() }()
	gzipReader, err := gzip.NewReader(file)
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
		if err := recordArchivePath(seen, header.Name); err != nil {
			return err
		}
		target := filepath.Join(destination, filepath.FromSlash(header.Name))
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

func extractZip(archivePath, destination string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open zip archive: %w", err)
	}
	defer func() { _ = reader.Close() }()
	seen := map[string]string{}
	var files int
	var total int64
	for _, member := range reader.File {
		if err := recordArchivePath(seen, member.Name); err != nil {
			return err
		}
		target := filepath.Join(destination, filepath.FromSlash(member.Name))
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

func recordArchivePath(seen map[string]string, name string) error {
	if err := safeRelativePath(strings.TrimSuffix(name, "/")); err != nil {
		return fmt.Errorf("unsafe archive member %q", name)
	}
	folded := strings.ToLower(strings.TrimSuffix(name, "/"))
	if first, exists := seen[folded]; exists {
		return fmt.Errorf("duplicate or case-folded archive members %q and %q", first, name)
	}
	seen[folded] = name
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
