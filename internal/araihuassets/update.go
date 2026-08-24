// Package araihuassets updates Goshtoso's allowlisted fallback assets from an
// already extracted, immutable araihu/assets release. It performs no network
// access; callers own release download and archive verification.
package araihuassets

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/araihu/goshtoso/internal/iconcatalog"
)

const DefaultManifestPath = "araihu-assets.json"

var (
	stableSemverRE = regexp.MustCompile(`^v(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)$`)
	fullSHARE      = regexp.MustCompile(`^[0-9a-f]{40}$`)
	sha256RE       = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// Manifest pins one immutable assets release and the only files the updater
// may copy into this repository.
type Manifest struct {
	SchemaVersion     int       `json:"schemaVersion"`
	AssetsRepository  string    `json:"assetsRepository"`
	AssetsRevision    string    `json:"assetsRevision"`
	Release           string    `json:"release"`
	ReleaseURL        string    `json:"releaseUrl"`
	ReleaseSHA256     string    `json:"releaseSha256"`
	ReleaseJSONSHA256 string    `json:"releaseJsonSha256"`
	Mappings          []Mapping `json:"mappings"`
}

// Mapping names one release file, its repository destination, and optionally
// the exact catalog identity and roles that must select it.
type Mapping struct {
	Source        string `json:"source"`
	Destination   string `json:"destination"`
	CanonicalName string `json:"canonicalName,omitempty"`
	Namespace     string `json:"namespace,omitempty"`
	Product       string `json:"product,omitempty"`
	Artwork       string `json:"artwork,omitempty"`
	Appearance    string `json:"appearance,omitempty"`
	Surface       string `json:"surface,omitempty"`
	Framing       string `json:"framing,omitempty"`
	Format        string `json:"format,omitempty"`
}

// ReleaseIdentity replaces only the immutable release identity. Mappings stay
// repository-owned and cannot be changed by a dispatch payload.
type ReleaseIdentity struct {
	AssetsRepository  string
	AssetsRevision    string
	Release           string
	ReleaseURL        string
	ReleaseSHA256     string
	ReleaseJSONSHA256 string
}

// Options identifies the repository, extracted release, and optional new
// release identity. Paths are local directories; Update has no network client.
type Options struct {
	RepoRoot      string
	ReleaseRoot   string
	ManifestPath  string
	Identity      *ReleaseIdentity
	beforeReplace func(index int, path string) error
}

// Result reports repository-relative paths whose bytes changed.
type Result struct {
	Changed []string
}

// ApplyError reports a failed transaction and whether rollback restored every
// path. AppliedPaths lists replacements attempted before failure;
// RollbackErrors names any path whose original bytes could not be restored.
type ApplyError struct {
	FailedPath     string
	AppliedPaths   []string
	RollbackErrors []string
	Cause          error
}

func (err *ApplyError) Error() string {
	state := "rollback complete"
	if len(err.RollbackErrors) != 0 {
		state = fmt.Sprintf("rollback incomplete: %s", strings.Join(err.RollbackErrors, "; "))
	}
	return fmt.Sprintf("replace %q after %d applied paths: %v (%s)", err.FailedPath, len(err.AppliedPaths), err.Cause, state)
}

func (err *ApplyError) Unwrap() error { return err.Cause }

// RollbackComplete reports whether no rollback operation failed.
func (err *ApplyError) RollbackComplete() bool { return len(err.RollbackErrors) == 0 }

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

type plannedWrite struct {
	path     string
	data     []byte
	original []byte
	existed  bool
}

type stagedWrite struct {
	plannedWrite
	temporary string
}

// Update verifies the pinned release metadata, catalog selections, file
// checksums, and path confinement before atomically replacing changed files.
func Update(opts Options) (Result, error) {
	manifestPath, err := validateOptions(opts)
	if err != nil {
		return Result{}, err
	}
	repo, err := os.OpenRoot(opts.RepoRoot)
	if err != nil {
		return Result{}, fmt.Errorf("open repo root: %w", err)
	}
	defer func() { _ = repo.Close() }()
	release, err := os.OpenRoot(opts.ReleaseRoot)
	if err != nil {
		return Result{}, fmt.Errorf("open release root: %w", err)
	}
	defer func() { _ = release.Close() }()

	manifest, manifestBytes, err := loadManifest(repo, manifestPath, opts.Identity)
	if err != nil {
		return Result{}, err
	}
	_, inventory, catalog, err := verifyRelease(release, manifest)
	if err != nil {
		return Result{}, err
	}
	writes, err := planMappingWrites(repo, release, manifest.Mappings, inventory, catalog)
	if err != nil {
		return Result{}, err
	}
	manifestWrite, changed, err := planManifestWrite(manifestPath, manifest, manifestBytes)
	if err != nil {
		return Result{}, err
	}
	if changed {
		writes = append(writes, manifestWrite)
	}
	sort.Slice(writes, func(i, j int) bool {
		if writes[i].path == manifestPath {
			return false
		}
		if writes[j].path == manifestPath {
			return true
		}
		return writes[i].path < writes[j].path
	})
	return applyWrites(repo, writes, opts.beforeReplace)
}

func validateOptions(opts Options) (string, error) {
	if opts.RepoRoot == "" || opts.ReleaseRoot == "" {
		return "", errors.New("repo root and release root are required")
	}
	manifestPath := opts.ManifestPath
	if manifestPath == "" {
		manifestPath = DefaultManifestPath
	}
	if err := validateRelativePath(manifestPath, "manifest path"); err != nil {
		return "", err
	}
	if err := rejectRootSymlink(opts.RepoRoot, "repo root"); err != nil {
		return "", err
	}
	if err := rejectRootSymlink(opts.ReleaseRoot, "release root"); err != nil {
		return "", err
	}
	return manifestPath, nil
}

func loadManifest(repo *os.Root, manifestPath string, identity *ReleaseIdentity) (Manifest, []byte, error) {
	if err := requireRegularPath(repo, manifestPath, false); err != nil {
		return Manifest{}, nil, fmt.Errorf("manifest: %w", err)
	}
	manifestBytes, err := repo.ReadFile(manifestPath)
	if err != nil {
		return Manifest{}, nil, fmt.Errorf("read manifest: %w", err)
	}
	manifest, err := decodeManifest(manifestBytes)
	if err != nil {
		return Manifest{}, nil, err
	}
	if identity != nil {
		manifest.AssetsRepository = identity.AssetsRepository
		manifest.AssetsRevision = identity.AssetsRevision
		manifest.Release = identity.Release
		manifest.ReleaseURL = identity.ReleaseURL
		manifest.ReleaseSHA256 = identity.ReleaseSHA256
		manifest.ReleaseJSONSHA256 = identity.ReleaseJSONSHA256
	}
	if err := validateManifest(manifest); err != nil {
		return Manifest{}, nil, err
	}
	return manifest, manifestBytes, nil
}

func planMappingWrites(repo, release *os.Root, mappings []Mapping, inventory map[string]releaseFile, catalog iconcatalog.Catalog) ([]plannedWrite, error) {
	writes := make([]plannedWrite, 0, len(mappings))
	for _, mapping := range mappings {
		contents, err := verifiedMappingBytes(release, mapping, inventory[mapping.Source], catalog)
		if err != nil {
			return nil, err
		}
		if err := requireRegularPath(repo, mapping.Destination, true); err != nil {
			return nil, fmt.Errorf("destination %q: %w", mapping.Destination, err)
		}
		current, err := repo.ReadFile(mapping.Destination)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("read destination %q: %w", mapping.Destination, err)
		}
		if !bytes.Equal(current, contents) {
			writes = append(writes, plannedWrite{
				path: mapping.Destination, data: contents,
				original: append([]byte(nil), current...), existed: err == nil,
			})
		}
	}
	return writes, nil
}

func verifiedMappingBytes(release *os.Root, mapping Mapping, entry releaseFile, catalog iconcatalog.Catalog) ([]byte, error) {
	if mapping.CanonicalName != "" {
		if err := verifyCatalogSelection(catalog, mapping, entry); err != nil {
			return nil, err
		}
	}
	if err := requireRegularPath(release, mapping.Source, false); err != nil {
		return nil, fmt.Errorf("source %q: %w", mapping.Source, err)
	}
	contents, err := release.ReadFile(mapping.Source)
	if err != nil {
		return nil, fmt.Errorf("read source %q: %w", mapping.Source, err)
	}
	if int64(len(contents)) != entry.Size {
		return nil, fmt.Errorf("source %q size = %d, want %d from release.json", mapping.Source, len(contents), entry.Size)
	}
	if got := digest(contents); got != entry.SHA256 {
		return nil, fmt.Errorf("source %q SHA-256 = %s, want %s from release.json", mapping.Source, got, entry.SHA256)
	}
	return contents, nil
}

func planManifestWrite(manifestPath string, manifest Manifest, current []byte) (plannedWrite, bool, error) {
	next, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return plannedWrite{}, false, fmt.Errorf("encode manifest: %w", err)
	}
	next = append(next, '\n')
	return plannedWrite{
		path: manifestPath, data: next,
		original: append([]byte(nil), current...), existed: true,
	}, !bytes.Equal(current, next), nil
}

func applyWrites(repo *os.Root, writes []plannedWrite, beforeReplace func(int, string) error) (Result, error) {
	staged, err := stageWrites(repo, writes)
	if err != nil {
		return Result{}, err
	}
	applied := make([]string, 0, len(staged))
	for index := range staged {
		if beforeReplace != nil {
			if err := beforeReplace(index, staged[index].path); err != nil {
				return Result{}, failedTransaction(repo, staged, index, applied, err)
			}
		}
		if err := repo.Rename(staged[index].temporary, staged[index].path); err != nil {
			return Result{}, failedTransaction(repo, staged, index, applied, err)
		}
		applied = append(applied, staged[index].path)
	}
	return Result{Changed: applied}, nil
}

func stageWrites(repo *os.Root, writes []plannedWrite) ([]stagedWrite, error) {
	staged := make([]stagedWrite, 0, len(writes))
	for _, write := range writes {
		if err := repo.MkdirAll(path.Dir(write.path), 0o755); err != nil {
			cleanupErrors := cleanupStages(repo, staged)
			return nil, &ApplyError{FailedPath: write.path, RollbackErrors: cleanupErrors, Cause: fmt.Errorf("create destination directory: %w", err)}
		}
		temporary, err := stageFile(repo, write.path, write.data)
		if err != nil {
			cleanupErrors := cleanupStages(repo, staged)
			return nil, &ApplyError{FailedPath: write.path, RollbackErrors: cleanupErrors, Cause: err}
		}
		staged = append(staged, stagedWrite{plannedWrite: write, temporary: temporary})
	}
	return staged, nil
}

func failedTransaction(repo *os.Root, staged []stagedWrite, failedIndex int, applied []string, cause error) *ApplyError {
	rollbackErrors := rollbackApplied(repo, staged[:failedIndex])
	rollbackErrors = append(rollbackErrors, cleanupStages(repo, staged[failedIndex:])...)
	return &ApplyError{
		FailedPath: staged[failedIndex].path, AppliedPaths: append([]string(nil), applied...),
		RollbackErrors: rollbackErrors, Cause: cause,
	}
}

func rollbackApplied(repo *os.Root, applied []stagedWrite) []string {
	problems := []string{}
	for _, write := range slices.Backward(applied) {

		if !write.existed {
			if err := repo.Remove(write.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
				problems = append(problems, fmt.Sprintf("remove new %q: %v", write.path, err))
			}
			continue
		}
		temporary, err := stageFile(repo, write.path, write.original)
		if err != nil {
			problems = append(problems, fmt.Sprintf("stage original %q: %v", write.path, err))
			continue
		}
		if err := repo.Rename(temporary, write.path); err != nil {
			problems = append(problems, fmt.Sprintf("restore %q: %v", write.path, err))
			_ = repo.Remove(temporary)
		}
	}
	return problems
}

func cleanupStages(repo *os.Root, staged []stagedWrite) []string {
	problems := []string{}
	for _, write := range staged {
		if err := repo.Remove(write.temporary); err != nil && !errors.Is(err, fs.ErrNotExist) {
			problems = append(problems, fmt.Sprintf("remove staged %q: %v", write.temporary, err))
		}
	}
	return problems
}

func decodeManifest(contents []byte) (Manifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return Manifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	return manifest, nil
}

func validateManifest(manifest Manifest) error {
	if manifest.SchemaVersion != 1 {
		return fmt.Errorf("manifest schemaVersion = %d, want 1", manifest.SchemaVersion)
	}
	if manifest.AssetsRepository != "araihu/assets" {
		return fmt.Errorf("assetsRepository = %q, want %q", manifest.AssetsRepository, "araihu/assets")
	}
	if !fullSHARE.MatchString(manifest.AssetsRevision) {
		return fmt.Errorf("invalid assetsRevision %q", manifest.AssetsRevision)
	}
	if !stableSemverRE.MatchString(manifest.Release) {
		return fmt.Errorf("invalid stable release tag %q", manifest.Release)
	}
	wantURL := fmt.Sprintf("https://github.com/araihu/assets/releases/download/%s/araihu-assets-%s.tar.gz", manifest.Release, manifest.Release)
	if manifest.ReleaseURL != wantURL {
		return fmt.Errorf("releaseUrl = %q, want %q", manifest.ReleaseURL, wantURL)
	}
	if !sha256RE.MatchString(manifest.ReleaseSHA256) {
		return fmt.Errorf("invalid releaseSha256 %q", manifest.ReleaseSHA256)
	}
	if !sha256RE.MatchString(manifest.ReleaseJSONSHA256) {
		return fmt.Errorf("invalid releaseJsonSha256 %q", manifest.ReleaseJSONSHA256)
	}
	if len(manifest.Mappings) == 0 {
		return errors.New("manifest mappings must not be empty")
	}
	destinations := make(map[string]string, len(manifest.Mappings))
	for index, mapping := range manifest.Mappings {
		if err := validateRelativePath(mapping.Source, "source"); err != nil {
			return fmt.Errorf("mapping %d: %w", index, err)
		}
		if err := validateRelativePath(mapping.Destination, "destination"); err != nil {
			return fmt.Errorf("mapping %d: %w", index, err)
		}
		folded := strings.ToLower(mapping.Destination)
		if previous, exists := destinations[folded]; exists {
			return fmt.Errorf("destination collision between %q and %q", previous, mapping.Destination)
		}
		destinations[folded] = mapping.Destination
		roles := []string{mapping.Namespace, mapping.Product, mapping.Artwork, mapping.Appearance, mapping.Surface, mapping.Framing, mapping.Format}
		if mapping.CanonicalName == "" {
			for _, role := range roles {
				if role != "" {
					return fmt.Errorf("mapping %d has catalog roles without canonicalName", index)
				}
			}
		} else {
			if slices.Contains(roles, "") {
				return fmt.Errorf("mapping %d canonical selection has an empty role", index)
			}
		}
	}
	return nil
}

func verifyRelease(root *os.Root, manifest Manifest) (releaseDocument, map[string]releaseFile, iconcatalog.Catalog, error) {
	document, err := loadReleaseDocument(root, manifest)
	if err != nil {
		return releaseDocument{}, nil, iconcatalog.Catalog{}, err
	}
	inventory, err := buildInventory(document, manifest.Mappings)
	if err != nil {
		return releaseDocument{}, nil, iconcatalog.Catalog{}, err
	}
	catalog, err := loadReleaseCatalog(root, document, inventory, manifest.Release)
	if err != nil {
		return releaseDocument{}, nil, iconcatalog.Catalog{}, err
	}
	return document, inventory, catalog, nil
}

func loadReleaseDocument(root *os.Root, manifest Manifest) (releaseDocument, error) {
	if err := requireRegularPath(root, "release.json", false); err != nil {
		return releaseDocument{}, fmt.Errorf("release.json: %w", err)
	}
	releaseBytes, err := root.ReadFile("release.json")
	if err != nil {
		return releaseDocument{}, fmt.Errorf("read release.json: %w", err)
	}
	if got := digest(releaseBytes); got != manifest.ReleaseJSONSHA256 {
		return releaseDocument{}, fmt.Errorf("release.json SHA-256 = %s, want %s", got, manifest.ReleaseJSONSHA256)
	}
	var document releaseDocument
	if err := json.Unmarshal(releaseBytes, &document); err != nil {
		return releaseDocument{}, fmt.Errorf("decode release.json: %w", err)
	}
	if document.SchemaVersion != 1 {
		return releaseDocument{}, fmt.Errorf("release.json schemaVersion = %d, want 1", document.SchemaVersion)
	}
	if document.Release != manifest.Release {
		return releaseDocument{}, fmt.Errorf("release.json release = %q, want %q", document.Release, manifest.Release)
	}
	return document, nil
}

func buildInventory(document releaseDocument, mappings []Mapping) (map[string]releaseFile, error) {
	inventory := make(map[string]releaseFile, len(document.Files))
	for index, entry := range document.Files {
		if err := validateRelativePath(entry.Path, "release.json file path"); err != nil {
			return nil, fmt.Errorf("release.json file %d: %w", index, err)
		}
		if !sha256RE.MatchString(entry.SHA256) || entry.Size < 0 {
			return nil, fmt.Errorf("release.json file %q has invalid identity", entry.Path)
		}
		if _, exists := inventory[entry.Path]; exists {
			return nil, fmt.Errorf("release.json file collision %q", entry.Path)
		}
		inventory[entry.Path] = entry
	}
	for _, mapping := range mappings {
		if _, exists := inventory[mapping.Source]; !exists {
			return nil, fmt.Errorf("source %q missing from release.json", mapping.Source)
		}
	}
	return inventory, nil
}

func loadReleaseCatalog(root *os.Root, document releaseDocument, inventory map[string]releaseFile, release string) (iconcatalog.Catalog, error) {
	catalogEntry, exists := inventory["catalog.json"]
	if !exists {
		return iconcatalog.Catalog{}, errors.New("catalog.json missing from release.json")
	}
	if catalogEntry.SHA256 != document.CatalogSHA256 {
		return iconcatalog.Catalog{}, errors.New("catalog.json identity disagrees with release.json catalogSha256")
	}
	if err := requireRegularPath(root, "catalog.json", false); err != nil {
		return iconcatalog.Catalog{}, fmt.Errorf("catalog.json: %w", err)
	}
	catalogBytes, err := root.ReadFile("catalog.json")
	if err != nil {
		return iconcatalog.Catalog{}, fmt.Errorf("read catalog.json: %w", err)
	}
	if got := digest(catalogBytes); got != catalogEntry.SHA256 {
		return iconcatalog.Catalog{}, fmt.Errorf("catalog.json SHA-256 = %s, want %s", got, catalogEntry.SHA256)
	}
	catalog, err := iconcatalog.Load(bytes.NewReader(catalogBytes))
	if err != nil {
		return iconcatalog.Catalog{}, fmt.Errorf("load catalog.json: %w", err)
	}
	if catalog.Release != release {
		return iconcatalog.Catalog{}, fmt.Errorf("catalog release = %q, want %q", catalog.Release, release)
	}
	return catalog, nil
}

func verifyCatalogSelection(catalog iconcatalog.Catalog, mapping Mapping, inventory releaseFile) error {
	var selected *iconcatalog.Asset
	for index := range catalog.Assets {
		if catalog.Assets[index].CanonicalName == mapping.CanonicalName {
			selected = &catalog.Assets[index]
			break
		}
	}
	if selected == nil {
		return fmt.Errorf("catalog canonicalName %q not found", mapping.CanonicalName)
	}
	roles := []struct {
		name, got, want string
	}{
		{"path", selected.Path, mapping.Source},
		{"namespace", selected.Namespace, mapping.Namespace},
		{"product", selected.Product, mapping.Product},
		{"artwork", selected.Artwork, mapping.Artwork},
		{"appearance", selected.Appearance, mapping.Appearance},
		{"surface", selected.Surface, mapping.Surface},
		{"framing", selected.Framing, mapping.Framing},
		{"format", selected.Format, mapping.Format},
	}
	for _, role := range roles {
		if role.got != role.want {
			return fmt.Errorf("catalog role %s = %q, want %q for %q", role.name, role.got, role.want, mapping.CanonicalName)
		}
	}
	if selected.SHA256 != inventory.SHA256 {
		return fmt.Errorf("catalog SHA-256 for %q disagrees with release.json", mapping.CanonicalName)
	}
	return nil
}

func validateRelativePath(name, field string) error {
	if name == "" || name == "." || strings.Contains(name, "\\") || path.IsAbs(name) || path.Clean(name) != name || strings.HasPrefix(name, "../") {
		return fmt.Errorf("unsafe %s %q", field, name)
	}
	return nil
}

func rejectRootSymlink(name, field string) error {
	info, err := os.Lstat(name)
	if err != nil {
		return fmt.Errorf("inspect %s: %w", field, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s must not be a symbolic link", field)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s must be a directory", field)
	}
	return nil
}

func requireRegularPath(root *os.Root, name string, allowMissing bool) error {
	parts := strings.Split(name, "/")
	for index := range parts {
		current := strings.Join(parts[:index+1], "/")
		info, err := root.Lstat(current)
		if errors.Is(err, fs.ErrNotExist) && allowMissing {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("path component %q is a symbolic link", current)
		}
		if index < len(parts)-1 && !info.IsDir() {
			return fmt.Errorf("path component %q is not a directory", current)
		}
		if index == len(parts)-1 && !info.Mode().IsRegular() {
			return fmt.Errorf("%q is not a regular file", current)
		}
	}
	return nil
}

func stageFile(root *os.Root, name string, contents []byte) (string, error) {
	random := make([]byte, 8)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("create temporary name for %q: %w", name, err)
	}
	temporary := fmt.Sprintf("%s.araihu-assets-%x.tmp", name, random)
	file, err := root.OpenFile(temporary, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", fmt.Errorf("create temporary file for %q: %w", name, err)
	}
	ok := false
	defer func() {
		_ = file.Close()
		if !ok {
			_ = root.Remove(temporary)
		}
	}()
	if _, err := file.Write(contents); err != nil {
		return "", fmt.Errorf("write temporary file for %q: %w", name, err)
	}
	if err := file.Sync(); err != nil {
		return "", fmt.Errorf("sync temporary file for %q: %w", name, err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close temporary file for %q: %w", name, err)
	}
	ok = true
	return temporary, nil
}

func requireJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return errors.New("more than one JSON value")
	} else if !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func digest(contents []byte) string {
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:])
}
