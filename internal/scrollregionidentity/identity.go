// Package scrollregionidentity independently derives T-GS-011 dirty-candidate
// identity. Capture and validation use this package instead of trusting a
// sidecar's mutually consistent claims.
package scrollregionidentity

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

const Schema = "goshtoso.t-gs-011.candidate-identity.v2"

var (
	gitObjectPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
	sha256Pattern    = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// DependencyPins are source-derived inputs that affect capture and validation.
type DependencyPins struct {
	RootGoDirective  string `json:"root_go_directive"`
	SiteGoDirective  string `json:"site_go_directive"`
	Templ            string `json:"templ"`
	PlaywrightGo     string `json:"playwright_go"`
	AxeCore          string `json:"axe_core"`
	AxeArchiveSHA256 string `json:"axe_archive_sha256"`
	AxeScriptSHA256  string `json:"axe_script_sha256"`
}

// CandidatePath binds a changed worktree path to exact bytes.
type CandidatePath struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// CandidateIdentity is the frozen sidecar contract for a dirty worktree.
type CandidateIdentity struct {
	Schema         string          `json:"schema"`
	RepositoryURL  string          `json:"repository_url"`
	Head           string          `json:"head"`
	Tree           string          `json:"tree"`
	CandidateTree  string          `json:"candidate_tree"`
	ManifestSHA256 string          `json:"manifest_sha256"`
	StatusSHA256   string          `json:"status_sha256"`
	Paths          []CandidatePath `json:"paths"`
	DependencyPins DependencyPins  `json:"dependency_pins"`
}

// Read decodes one strict candidate sidecar.
func Read(path string) (CandidateIdentity, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return CandidateIdentity{}, fmt.Errorf("read candidate identity: %w", err)
	}
	var identity CandidateIdentity
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&identity); err != nil {
		return CandidateIdentity{}, fmt.Errorf("decode candidate identity: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return CandidateIdentity{}, fmt.Errorf("decode candidate identity: trailing JSON values")
		}
		return CandidateIdentity{}, fmt.Errorf("decode candidate identity: trailing data: %w", err)
	}
	return identity, nil
}

// Verify reads and independently re-derives every mutable identity fact.
func Verify(repositoryRoot, sidecarPath string) (CandidateIdentity, error) {
	identity, err := Read(sidecarPath)
	if err != nil {
		return CandidateIdentity{}, err
	}
	if err := VerifyIdentity(repositoryRoot, identity); err != nil {
		return CandidateIdentity{}, err
	}
	return identity, nil
}

// VerifyIdentity independently checks Git, every declared byte, the candidate
// tree/path set, and source dependency pins. Sidecar data is never authority.
func VerifyIdentity(repositoryRoot string, identity CandidateIdentity) error {
	if err := ValidateShape(identity); err != nil {
		return err
	}
	head, err := gitOutput(repositoryRoot, nil, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if identity.Head != strings.TrimSpace(head) {
		return fmt.Errorf("candidate identity HEAD mismatch: got %s, want %s", identity.Head, strings.TrimSpace(head))
	}
	tree, err := gitOutput(repositoryRoot, nil, "rev-parse", "HEAD^{tree}")
	if err != nil {
		return err
	}
	if identity.Tree != strings.TrimSpace(tree) {
		return fmt.Errorf("candidate identity tree mismatch: got %s, want %s", identity.Tree, strings.TrimSpace(tree))
	}
	remote, err := gitOutput(repositoryRoot, nil, "config", "--get", "remote.origin.url")
	if err != nil {
		return err
	}
	if identity.RepositoryURL != strings.TrimSpace(remote) {
		return fmt.Errorf("candidate identity repository URL mismatch: got %q, want %q", identity.RepositoryURL, strings.TrimSpace(remote))
	}
	status, err := gitOutput(repositoryRoot, nil, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		return err
	}
	if identity.StatusSHA256 != SHA256([]byte(status)) {
		return fmt.Errorf("candidate identity status SHA-256 mismatch")
	}
	candidateTree, err := CandidateTree(repositoryRoot)
	if err != nil {
		return err
	}
	if identity.CandidateTree != candidateTree {
		return fmt.Errorf("candidate identity candidate tree mismatch: got %s, want %s", identity.CandidateTree, candidateTree)
	}
	if err := verifyPaths(repositoryRoot, identity); err != nil {
		return err
	}
	pins, err := SourceDependencyPins(repositoryRoot)
	if err != nil {
		return err
	}
	if identity.DependencyPins != pins {
		return fmt.Errorf("candidate identity dependency pins mismatch: got %#v, want %#v", identity.DependencyPins, pins)
	}
	return nil
}

// ValidateShape rejects malformed, duplicated, unsorted, or incomplete sidecars.
func ValidateShape(identity CandidateIdentity) error {
	if identity.Schema != Schema {
		return fmt.Errorf("candidate identity schema mismatch: got %q", identity.Schema)
	}
	if strings.TrimSpace(identity.RepositoryURL) == "" || strings.TrimSpace(identity.RepositoryURL) != identity.RepositoryURL {
		return fmt.Errorf("candidate identity repository URL is invalid")
	}
	for name, value := range map[string]string{
		"head": identity.Head, "tree": identity.Tree, "candidate_tree": identity.CandidateTree,
	} {
		if !gitObjectPattern.MatchString(value) {
			return fmt.Errorf("candidate identity %s must be a lowercase 40-hex Git object", name)
		}
	}
	for name, value := range map[string]string{
		"manifest_sha256": identity.ManifestSHA256, "status_sha256": identity.StatusSHA256,
	} {
		if !sha256Pattern.MatchString(value) {
			return fmt.Errorf("candidate identity %s must be a lowercase 64-hex SHA-256", name)
		}
	}
	if len(identity.Paths) == 0 {
		return fmt.Errorf("candidate identity paths are empty")
	}
	previous := ""
	for _, path := range identity.Paths {
		if !CandidatePathIsSafe(path.Path) {
			return fmt.Errorf("candidate identity path is invalid: %q", path.Path)
		}
		if path.Path <= previous {
			return fmt.Errorf("candidate identity paths must be unique and sorted")
		}
		if !sha256Pattern.MatchString(path.SHA256) {
			return fmt.Errorf("candidate identity path %q has invalid SHA-256", path.Path)
		}
		previous = path.Path
	}
	if identity.ManifestSHA256 != ManifestSHA256(identity.Paths) {
		return fmt.Errorf("candidate identity manifest SHA-256 does not bind the canonical path list")
	}
	return validateDependencyPins(identity.DependencyPins)
}

// CandidatePathIsSafe permits only canonical repository-relative slash paths.
func CandidatePathIsSafe(path string) bool {
	if path == "" || filepath.IsAbs(path) || filepath.Clean(path) != path || filepath.ToSlash(path) != path {
		return false
	}
	return path != "." && !strings.HasPrefix(path, "../") && !strings.Contains(path, "/../")
}

// ManifestSHA256 hashes canonical path-tab-digest lines.
func ManifestSHA256(paths []CandidatePath) string {
	var manifest strings.Builder
	for _, path := range paths {
		manifest.WriteString(path.Path)
		manifest.WriteByte('\t')
		manifest.WriteString(path.SHA256)
		manifest.WriteByte('\n')
	}
	return SHA256([]byte(manifest.String()))
}

// SHA256 returns the lower-case SHA-256 digest of bytes.
func SHA256(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}

// CandidateTree writes the full current worktree into a throwaway alternate
// index, preserving the caller's real index and staged state.
func CandidateTree(repositoryRoot string) (string, error) {
	index, err := os.CreateTemp("", "goshtoso-scrollregion-candidate-index-")
	if err != nil {
		return "", fmt.Errorf("create alternate candidate index: %w", err)
	}
	indexPath := index.Name()
	if err := index.Close(); err != nil {
		return "", fmt.Errorf("close alternate candidate index: %w", err)
	}
	if err := os.Remove(indexPath); err != nil {
		return "", fmt.Errorf("prepare alternate candidate index: %w", err)
	}
	defer func() { _ = os.Remove(indexPath) }()
	environment := []string{"GIT_INDEX_FILE=" + indexPath}
	if _, err := gitOutput(repositoryRoot, environment, "read-tree", "HEAD"); err != nil {
		return "", err
	}
	if _, err := gitOutput(repositoryRoot, environment, "add", "-A"); err != nil {
		return "", err
	}
	tree, err := gitOutput(repositoryRoot, environment, "write-tree")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(tree), nil
}

// SourceDependencyPins derives root/site Go and locked axe inputs from source.
func SourceDependencyPins(repositoryRoot string) (DependencyPins, error) {
	rootGo, rootRequire, err := readGoMod(filepath.Join(repositoryRoot, "go.mod"))
	if err != nil {
		return DependencyPins{}, err
	}
	siteGo, siteRequire, err := readGoMod(filepath.Join(repositoryRoot, "site", "go.mod"))
	if err != nil {
		return DependencyPins{}, err
	}
	rootTempl := rootRequire["github.com/a-h/templ"]
	siteTempl := siteRequire["github.com/a-h/templ"]
	if rootTempl == "" || siteTempl == "" {
		return DependencyPins{}, fmt.Errorf("templ pin is missing from root or site go.mod")
	}
	if rootTempl != siteTempl {
		return DependencyPins{}, fmt.Errorf("root and site templ pins differ: %s versus %s", rootTempl, siteTempl)
	}
	playwright := siteRequire["github.com/mxschmitt/playwright-go"]
	if playwright == "" {
		return DependencyPins{}, fmt.Errorf("playwright-go pin is missing from site go.mod")
	}
	axe, err := readAxeLock(filepath.Join(repositoryRoot, "scripts", "axe-core.lock"))
	if err != nil {
		return DependencyPins{}, err
	}
	return DependencyPins{
		RootGoDirective: rootGo, SiteGoDirective: siteGo, Templ: rootTempl, PlaywrightGo: playwright,
		AxeCore: axe["version"], AxeArchiveSHA256: axe["archive_sha256"], AxeScriptSHA256: axe["script_sha256"],
	}, nil
}

func verifyPaths(repositoryRoot string, identity CandidateIdentity) error {
	changed, err := gitOutput(repositoryRoot, nil, "diff", "--name-only", identity.Tree, identity.CandidateTree)
	if err != nil {
		return err
	}
	actualPaths := make([]string, 0)
	for line := range strings.SplitSeq(strings.TrimSpace(changed), "\n") {
		if line != "" {
			actualPaths = append(actualPaths, line)
		}
	}
	declared := make([]string, 0, len(identity.Paths))
	for _, path := range identity.Paths {
		declared = append(declared, path.Path)
		content, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(path.Path)))
		if err != nil {
			return fmt.Errorf("read candidate path %q: %w", path.Path, err)
		}
		if path.SHA256 != SHA256(content) {
			return fmt.Errorf("candidate path %q SHA-256 mismatch", path.Path)
		}
	}
	slices.Sort(actualPaths)
	if !slices.Equal(declared, actualPaths) {
		return fmt.Errorf("candidate identity paths mismatch: declared %v, Git candidate tree has %v", declared, actualPaths)
	}
	return nil
}

func validateDependencyPins(pins DependencyPins) error {
	for name, value := range map[string]string{
		"root_go_directive": pins.RootGoDirective, "site_go_directive": pins.SiteGoDirective,
		"templ": pins.Templ, "playwright_go": pins.PlaywrightGo, "axe_core": pins.AxeCore,
	} {
		if strings.TrimSpace(value) == "" || strings.TrimSpace(value) != value {
			return fmt.Errorf("candidate identity dependency pin %s is invalid", name)
		}
	}
	for name, value := range map[string]string{
		"axe_archive_sha256": pins.AxeArchiveSHA256, "axe_script_sha256": pins.AxeScriptSHA256,
	} {
		if !sha256Pattern.MatchString(value) {
			return fmt.Errorf("candidate identity dependency pin %s must be lowercase 64-hex", name)
		}
	}
	return nil
}

func readGoMod(path string) (string, map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", nil, fmt.Errorf("read %s: %w", filepath.Base(path), err)
	}
	requires := make(map[string]string)
	goVersion := ""
	inRequire := false
	for rawLine := range strings.SplitSeq(string(content), "\n") {
		line := strings.TrimSpace(strings.SplitN(rawLine, "//", 2)[0])
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "go ") {
			fields := strings.Fields(line)
			if len(fields) == 2 {
				goVersion = fields[1]
			}
			continue
		}
		if line == "require (" {
			inRequire = true
			continue
		}
		if inRequire && line == ")" {
			inRequire = false
			continue
		}
		if after, ok := strings.CutPrefix(line, "require "); ok {
			line = strings.TrimSpace(after)
		} else if !inRequire {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			requires[fields[0]] = fields[1]
		}
	}
	if goVersion == "" {
		return "", nil, fmt.Errorf("%s has no Go directive", filepath.Base(path))
	}
	return goVersion, requires, nil
}

func readAxeLock(path string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read axe-core lock: %w", err)
	}
	values := make(map[string]string)
	for rawLine := range strings.SplitSeq(string(content), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !found || key == "" || value == "" {
			return nil, fmt.Errorf("invalid axe-core lock line %q", line)
		}
		if _, duplicate := values[key]; duplicate {
			return nil, fmt.Errorf("duplicate axe-core lock key %q", key)
		}
		values[key] = value
	}
	for _, key := range []string{"version", "archive_sha256", "script_sha256"} {
		if values[key] == "" {
			return nil, fmt.Errorf("axe-core lock missing %q", key)
		}
	}
	return values, nil
}

func gitOutput(repositoryRoot string, environment []string, arguments ...string) (string, error) {
	command := exec.Command("git", arguments...)
	command.Dir = repositoryRoot
	command.Env = append(os.Environ(), environment...)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(arguments, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}
