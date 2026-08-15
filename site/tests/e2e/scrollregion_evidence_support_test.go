//go:build e2e && scrollregion && bfull && axe

package e2e

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/araihu/goshtoso/internal/scrollregionidentity"
	"golang.org/x/mod/modfile"
)

const scrollRegionCandidateIdentitySchema = "goshtoso.t-gs-011.candidate-identity.v2"

const (
	scrollRegionBFullIdentityEnvironment = "GOSHTOSO_SCROLLREGION_BFULL_IDENTITY"
	scrollRegionBFullRequireSealedEnv    = "GOSHTOSO_SCROLLREGION_BFULL_REQUIRE_SEALED"
)

var (
	scrollRegionGitObjectPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
	scrollRegionSHA256Pattern    = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// scrollRegionCandidatePath binds one worktree path to its exact candidate
// bytes. The canonical path list itself is hashed into ManifestSHA256.
type scrollRegionCandidatePath = scrollregionidentity.CandidatePath

// scrollRegionDependencyPins are derived again from the candidate source, not
// trusted because they merely appear in a caller-provided sidecar.
type scrollRegionDependencyPins = scrollregionidentity.DependencyPins

// scrollRegionCandidateIdentity is the only acceptable sealing input for a
// dirty candidate. Each field is checked against Git and source bytes before a
// B-FULL or AT receipt can claim final identity.
type scrollRegionCandidateIdentity = scrollregionidentity.CandidateIdentity

func readScrollRegionCandidateIdentity(path string) (scrollRegionCandidateIdentity, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return scrollRegionCandidateIdentity{}, fmt.Errorf("read candidate identity: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	var identity scrollRegionCandidateIdentity
	if err := decoder.Decode(&identity); err != nil {
		return scrollRegionCandidateIdentity{}, fmt.Errorf("decode candidate identity: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return scrollRegionCandidateIdentity{}, fmt.Errorf("decode candidate identity: trailing JSON values")
		}
		return scrollRegionCandidateIdentity{}, fmt.Errorf("decode candidate identity: trailing data: %w", err)
	}
	return identity, nil
}

// verifyScrollRegionCandidateIdentity independently derives all mutable Git and
// source facts. A sidecar is an input contract, never evidence by assertion.
func verifyScrollRegionCandidateIdentity(repositoryRoot, sidecarPath string) (scrollRegionCandidateIdentity, error) {
	return scrollregionidentity.Verify(repositoryRoot, sidecarPath)
}

// scrollRegionBFullIdentityBinding distinguishes final, independently verified
// evidence from a deliberately unbound local diagnostic. Only a clean commit
// or a verified dirty-candidate sidecar can produce a sealed binding.
type scrollRegionBFullIdentityBinding struct {
	Binding        string                       `json:"binding"`
	Note           string                       `json:"note,omitempty"`
	Identity       *scrollRegionReceiptIdentity `json:"identity,omitempty"`
	CandidatePaths []scrollRegionCandidatePath  `json:"candidate_paths,omitempty"`

	sidecarPath string
}

func resolveScrollRegionBFullIdentity(repositoryRoot string, requireSealed bool) (scrollRegionBFullIdentityBinding, error) {
	sidecarPath := strings.TrimSpace(os.Getenv(scrollRegionBFullIdentityEnvironment))
	if sidecarPath != "" {
		candidate, err := verifyScrollRegionCandidateIdentity(repositoryRoot, sidecarPath)
		if err != nil {
			return scrollRegionBFullIdentityBinding{}, fmt.Errorf("verify B-FULL candidate sidecar: %w", err)
		}
		identity := scrollRegionReceiptIdentityFromCandidate(candidate)
		return scrollRegionBFullIdentityBinding{
			Binding:        "sealed-dirty-candidate",
			Identity:       &identity,
			CandidatePaths: slices.Clone(candidate.Paths),
			sidecarPath:    sidecarPath,
		}, nil
	}
	status, err := scrollRegionGitOutput(repositoryRoot, nil, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		return scrollRegionBFullIdentityBinding{}, err
	}
	if status != "" {
		if requireSealed || strings.TrimSpace(os.Getenv(scrollRegionBFullRequireSealedEnv)) == "1" {
			return scrollRegionBFullIdentityBinding{}, fmt.Errorf("sealed B-FULL evidence requires %s for a dirty worktree", scrollRegionBFullIdentityEnvironment)
		}
		return scrollRegionBFullIdentityBinding{
			Binding: "diagnostic-unbound-dirty-worktree",
			Note:    "No candidate identity sidecar was supplied; this dirty-worktree run is diagnostic only and carries no source identity claim.",
		}, nil
	}
	head, err := scrollRegionGitOutput(repositoryRoot, nil, "rev-parse", "HEAD")
	if err != nil {
		return scrollRegionBFullIdentityBinding{}, err
	}
	tree, err := scrollRegionGitOutput(repositoryRoot, nil, "rev-parse", "HEAD^{tree}")
	if err != nil {
		return scrollRegionBFullIdentityBinding{}, err
	}
	repositoryURL, err := scrollRegionGitOutput(repositoryRoot, nil, "config", "--get", "remote.origin.url")
	if err != nil {
		return scrollRegionBFullIdentityBinding{}, err
	}
	pins, err := scrollRegionSourceDependencyPins(repositoryRoot)
	if err != nil {
		return scrollRegionBFullIdentityBinding{}, err
	}
	candidateTree, err := scrollRegionCandidateTree(repositoryRoot)
	if err != nil {
		return scrollRegionBFullIdentityBinding{}, err
	}
	head = strings.TrimSpace(head)
	tree = strings.TrimSpace(tree)
	if candidateTree != tree {
		return scrollRegionBFullIdentityBinding{}, fmt.Errorf("clean worktree candidate tree mismatch: got %s, want %s", candidateTree, tree)
	}
	identity := scrollRegionReceiptIdentity{
		RepositoryURL:  strings.TrimSpace(repositoryURL),
		Head:           head,
		Tree:           tree,
		CandidateTree:  candidateTree,
		ManifestSHA256: scrollRegionCandidateManifestSHA256(nil),
		StatusSHA256:   scrollRegionBFullSHA256(nil),
		DependencyPins: pins,
	}
	return scrollRegionBFullIdentityBinding{Binding: "sealed-clean-commit", Identity: &identity}, nil
}

func (binding scrollRegionBFullIdentityBinding) revalidate(repositoryRoot string) error {
	if binding.Identity == nil {
		return nil
	}
	if binding.sidecarPath != "" {
		candidate, err := verifyScrollRegionCandidateIdentity(repositoryRoot, binding.sidecarPath)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(scrollRegionReceiptIdentityFromCandidate(candidate), *binding.Identity) || !slices.Equal(candidate.Paths, binding.CandidatePaths) {
			return fmt.Errorf("candidate sidecar identity changed while B-FULL evidence was running")
		}
		return nil
	}
	current, err := resolveScrollRegionBFullIdentity(repositoryRoot, true)
	if err != nil {
		return err
	}
	if current.Binding != binding.Binding || !reflect.DeepEqual(current.Identity, binding.Identity) {
		return fmt.Errorf("clean commit identity changed while B-FULL evidence was running")
	}
	return nil
}

func validateScrollRegionCandidateIdentityShape(identity scrollRegionCandidateIdentity) error {
	if identity.Schema != scrollRegionCandidateIdentitySchema {
		return fmt.Errorf("candidate identity schema mismatch: got %q", identity.Schema)
	}
	if strings.TrimSpace(identity.RepositoryURL) == "" || strings.TrimSpace(identity.RepositoryURL) != identity.RepositoryURL {
		return fmt.Errorf("candidate identity repository URL is invalid")
	}
	for name, value := range map[string]string{
		"head":           identity.Head,
		"tree":           identity.Tree,
		"candidate_tree": identity.CandidateTree,
	} {
		if !scrollRegionGitObjectPattern.MatchString(value) {
			return fmt.Errorf("candidate identity %s must be a lowercase 40-hex Git object", name)
		}
	}
	for name, value := range map[string]string{
		"manifest_sha256": identity.ManifestSHA256,
		"status_sha256":   identity.StatusSHA256,
	} {
		if !scrollRegionSHA256Pattern.MatchString(value) {
			return fmt.Errorf("candidate identity %s must be a lowercase 64-hex SHA-256", name)
		}
	}
	if len(identity.Paths) == 0 {
		return fmt.Errorf("candidate identity paths are empty")
	}
	previous := ""
	for _, path := range identity.Paths {
		if !scrollRegionCandidatePathIsSafe(path.Path) {
			return fmt.Errorf("candidate identity path is invalid: %q", path.Path)
		}
		if path.Path <= previous {
			return fmt.Errorf("candidate identity paths must be unique and sorted")
		}
		if !scrollRegionSHA256Pattern.MatchString(path.SHA256) {
			return fmt.Errorf("candidate identity path %q has invalid SHA-256", path.Path)
		}
		previous = path.Path
	}
	if identity.ManifestSHA256 != scrollRegionCandidateManifestSHA256(identity.Paths) {
		return fmt.Errorf("candidate identity manifest SHA-256 does not bind the canonical path list")
	}
	if err := validateScrollRegionDependencyPinsShape(identity.DependencyPins); err != nil {
		return err
	}
	return nil
}

func scrollRegionCandidatePathIsSafe(path string) bool {
	if path == "" || filepath.IsAbs(path) || filepath.Clean(path) != path || filepath.ToSlash(path) != path {
		return false
	}
	return path != "." && !strings.HasPrefix(path, "../") && !strings.Contains(path, "/../")
}

func validateScrollRegionDependencyPinsShape(pins scrollRegionDependencyPins) error {
	for name, value := range map[string]string{
		"root_go_directive": pins.RootGoDirective,
		"site_go_directive": pins.SiteGoDirective,
		"templ":             pins.Templ,
		"playwright_go":     pins.PlaywrightGo,
		"axe_core":          pins.AxeCore,
	} {
		if strings.TrimSpace(value) == "" || strings.TrimSpace(value) != value {
			return fmt.Errorf("candidate identity dependency pin %s is invalid", name)
		}
	}
	for name, value := range map[string]string{
		"axe_archive_sha256": pins.AxeArchiveSHA256,
		"axe_script_sha256":  pins.AxeScriptSHA256,
	} {
		if !scrollRegionSHA256Pattern.MatchString(value) {
			return fmt.Errorf("candidate identity dependency pin %s must be lowercase 64-hex", name)
		}
	}
	return nil
}

func verifyScrollRegionCandidatePaths(repositoryRoot string, identity scrollRegionCandidateIdentity) error {
	changed, err := scrollRegionGitOutput(repositoryRoot, nil, "diff", "--name-only", identity.Tree, identity.CandidateTree)
	if err != nil {
		return err
	}
	var actualPaths []string
	for line := range strings.SplitSeq(strings.TrimSpace(changed), "\n") {
		if line != "" {
			actualPaths = append(actualPaths, line)
		}
	}
	declaredPaths := make([]string, 0, len(identity.Paths))
	for _, path := range identity.Paths {
		declaredPaths = append(declaredPaths, path.Path)
		content, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(path.Path)))
		if err != nil {
			return fmt.Errorf("read candidate path %q: %w", path.Path, err)
		}
		if path.SHA256 != scrollRegionBFullSHA256(content) {
			return fmt.Errorf("candidate path %q SHA-256 mismatch", path.Path)
		}
	}
	slices.Sort(actualPaths)
	if !slices.Equal(declaredPaths, actualPaths) {
		return fmt.Errorf("candidate identity paths mismatch: declared %v, Git candidate tree has %v", declaredPaths, actualPaths)
	}
	return nil
}

func scrollRegionCandidateManifestSHA256(paths []scrollRegionCandidatePath) string {
	var manifest strings.Builder
	for _, path := range paths {
		manifest.WriteString(path.Path)
		manifest.WriteByte('\t')
		manifest.WriteString(path.SHA256)
		manifest.WriteByte('\n')
	}
	digest := sha256.Sum256([]byte(manifest.String()))
	return hex.EncodeToString(digest[:])
}

func scrollRegionCandidateTree(repositoryRoot string) (string, error) {
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
	if _, err := scrollRegionGitOutput(repositoryRoot, environment, "read-tree", "HEAD"); err != nil {
		return "", err
	}
	if _, err := scrollRegionGitOutput(repositoryRoot, environment, "add", "-A"); err != nil {
		return "", err
	}
	tree, err := scrollRegionGitOutput(repositoryRoot, environment, "write-tree")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(tree), nil
}

func scrollRegionGitOutput(repositoryRoot string, environment []string, arguments ...string) (string, error) {
	command := exec.Command("git", arguments...)
	command.Dir = repositoryRoot
	command.Env = append(os.Environ(), environment...)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(arguments, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func scrollRegionSourceDependencyPins(repositoryRoot string) (scrollRegionDependencyPins, error) {
	root, err := scrollRegionReadModFile(filepath.Join(repositoryRoot, "go.mod"))
	if err != nil {
		return scrollRegionDependencyPins{}, err
	}
	site, err := scrollRegionReadModFile(filepath.Join(repositoryRoot, "site", "go.mod"))
	if err != nil {
		return scrollRegionDependencyPins{}, err
	}
	rootTempl, err := scrollRegionModuleVersion(root, "github.com/a-h/templ")
	if err != nil {
		return scrollRegionDependencyPins{}, err
	}
	siteTempl, err := scrollRegionModuleVersion(site, "github.com/a-h/templ")
	if err != nil {
		return scrollRegionDependencyPins{}, err
	}
	if rootTempl != siteTempl {
		return scrollRegionDependencyPins{}, fmt.Errorf("root and site templ pins differ: %s versus %s", rootTempl, siteTempl)
	}
	playwrightVersion, err := scrollRegionModuleVersion(site, "github.com/mxschmitt/playwright-go")
	if err != nil {
		return scrollRegionDependencyPins{}, err
	}
	axeLock, err := scrollRegionReadAxeLock(filepath.Join(repositoryRoot, "scripts", "axe-core.lock"))
	if err != nil {
		return scrollRegionDependencyPins{}, err
	}
	return scrollRegionDependencyPins{
		RootGoDirective:  root.Go.Version,
		SiteGoDirective:  site.Go.Version,
		Templ:            rootTempl,
		PlaywrightGo:     playwrightVersion,
		AxeCore:          axeLock["version"],
		AxeArchiveSHA256: axeLock["archive_sha256"],
		AxeScriptSHA256:  axeLock["script_sha256"],
	}, nil
}

func scrollRegionReadModFile(path string) (*modfile.File, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", filepath.Base(path), err)
	}
	parsed, err := modfile.Parse(path, content, nil)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", filepath.Base(path), err)
	}
	if parsed.Go == nil || parsed.Go.Version == "" {
		return nil, fmt.Errorf("%s has no Go directive", filepath.Base(path))
	}
	return parsed, nil
}

func scrollRegionModuleVersion(file *modfile.File, modulePath string) (string, error) {
	for _, require := range file.Require {
		if require.Mod.Path == modulePath {
			return require.Mod.Version, nil
		}
	}
	return "", fmt.Errorf("%s is not required by %s", modulePath, file.Module.Mod.Path)
}

func scrollRegionReadAxeLock(path string) (map[string]string, error) {
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
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
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

const (
	scrollRegionATReceiptSchema             = "goshtoso.t-gs-011.at-receipt.v3"
	scrollRegionATTranscriptSchema          = "goshtoso.t-gs-011.at-transcript.v2"
	scrollRegionATTraceLogSchema            = "goshtoso.t-gs-011.at-trace-log.v2"
	scrollRegionATChallengeSchema           = "goshtoso.t-gs-011.at-challenge.v1"
	scrollRegionATAttestationSchema         = "goshtoso.t-gs-011.at-dsse-envelope.v1"
	scrollRegionATAttestationPayloadType    = "application/vnd.goshtoso.t-gs-011.at-capture.v1+json"
	scrollRegionATAttestationPayloadSchema  = "goshtoso.t-gs-011.at-capture.v1"
	scrollRegionATTrustedKeysSchema         = "goshtoso.t-gs-011.at-attestation-keys.v1"
	scrollRegionATChallengeEnvironment      = "GOSHTOSO_SCROLLREGION_AT_CHALLENGE"
	scrollRegionATChallengeClaimSchema      = "goshtoso.t-gs-011.at-challenge-claim.v1"
	scrollRegionATReplayRegistryEnvironment = "GOSHTOSO_SCROLLREGION_AT_REPLAY_REGISTRY"
	scrollRegionATServedPageSchema          = "goshtoso.t-gs-011.at-served-page.v1"
	scrollRegionATBrowserStateSchema        = "goshtoso.t-gs-011.at-browser-state.v1"
)

// scrollRegionReceiptIdentity is deliberately smaller than the sidecar but is
// derived only after the complete sidecar has passed Git and byte checks.
type scrollRegionReceiptIdentity struct {
	RepositoryURL  string                     `json:"repository_url"`
	Head           string                     `json:"head"`
	Tree           string                     `json:"tree"`
	CandidateTree  string                     `json:"candidate_tree"`
	ManifestSHA256 string                     `json:"manifest_sha256"`
	StatusSHA256   string                     `json:"status_sha256"`
	DependencyPins scrollRegionDependencyPins `json:"dependency_pins"`
}

func scrollRegionReceiptIdentityFromCandidate(candidate scrollRegionCandidateIdentity) scrollRegionReceiptIdentity {
	return scrollRegionReceiptIdentity{
		RepositoryURL:  candidate.RepositoryURL,
		Head:           candidate.Head,
		Tree:           candidate.Tree,
		CandidateTree:  candidate.CandidateTree,
		ManifestSHA256: candidate.ManifestSHA256,
		StatusSHA256:   candidate.StatusSHA256,
		DependencyPins: candidate.DependencyPins,
	}
}

type scrollRegionATEvidenceArtifact struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type scrollRegionATServedPage struct {
	Schema         string `json:"schema"`
	URL            string `json:"url"`
	Status         int    `json:"status"`
	Challenge      string `json:"challenge"`
	CandidateTree  string `json:"candidate_tree"`
	ManifestSHA256 string `json:"manifest_sha256"`
	BodySHA256     string `json:"body_sha256"`
	Pair           string `json:"pair"`
	ActionState    string `json:"action_state"`
	ActionToken    string `json:"action_token"`
}

type scrollRegionATRectangle struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type scrollRegionATBrowserState struct {
	Schema          string                        `json:"schema"`
	Pair            string                        `json:"pair"`
	State           string                        `json:"state"`
	Command         string                        `json:"command"`
	Route           string                        `json:"route"`
	Challenge       string                        `json:"challenge"`
	CandidateTree   string                        `json:"candidate_tree"`
	ManifestSHA256  string                        `json:"manifest_sha256"`
	Role            string                        `json:"role"`
	Name            string                        `json:"name"`
	Focused         bool                          `json:"focused"`
	Boundary        string                        `json:"boundary"`
	BeforeBoundary  string                        `json:"before_boundary"`
	FocusNavigation scrollRegionATFocusNavigation `json:"focus_navigation"`
	ActionToken     string                        `json:"action_token"`
	Phase           string                        `json:"phase"`
	ObservedAt      string                        `json:"observed_at"`
	Snapshot        scrollRegionATBrowserSnapshot `json:"snapshot"`
	CandidateRegion scrollRegionATRectangle       `json:"candidate_region"`
}

// scrollRegionATBrowserSnapshot is emitted from the browser for one actual
// before, after, or exit phase. Focus navigation is reconstructed from these
// values; it must not be supplied by a DOM adjacency heuristic.
type scrollRegionATBrowserSnapshot struct {
	ActiveRole      string  `json:"active_role"`
	ActiveName      string  `json:"active_name"`
	RegionRole      string  `json:"region_role"`
	RegionName      string  `json:"region_name"`
	RegionFocused   bool    `json:"region_focused"`
	Boundary        string  `json:"boundary"`
	ScrollTop       float64 `json:"scroll_top"`
	ClientHeight    float64 `json:"client_height"`
	ScrollHeight    float64 `json:"scroll_height"`
	StartCueVisible bool    `json:"start_cue_visible"`
	EndCueVisible   bool    `json:"end_cue_visible"`
}

type scrollRegionATActionRecord struct {
	Schema             string `json:"schema"`
	Pair               string `json:"pair"`
	State              string `json:"state"`
	Command            string `json:"command"`
	Route              string `json:"route"`
	Challenge          string `json:"challenge"`
	CandidateTree      string `json:"candidate_tree"`
	ManifestSHA256     string `json:"manifest_sha256"`
	ActionToken        string `json:"action_token"`
	ActionEvent        string `json:"action_event"`
	ExitCommand        string `json:"exit_command"`
	ExitEvent          string `json:"exit_event"`
	VoiceOverPID       int    `json:"voiceover_pid"`
	VoiceOverSubsystem string `json:"voiceover_subsystem"`
	BeforeAt           string `json:"before_at"`
	LogStartedAt       string `json:"log_started_at"`
	ActionIssuedAt     string `json:"action_issued_at"`
	AfterAt            string `json:"after_at"`
	ExitIssuedAt       string `json:"exit_issued_at"`
	ExitAt             string `json:"exit_at"`
	LogEndedAt         string `json:"log_ended_at"`
}

const scrollRegionATActionRecordSchema = "goshtoso.t-gs-011.at-action-record.v2"
const scrollRegionATVoiceOverSubsystem = "com.apple.VoiceOver"

type scrollRegionATVoiceOverLogEvent struct {
	Timestamp    string `json:"timestamp"`
	ProcessID    int    `json:"processID"`
	Subsystem    string `json:"subsystem"`
	EventMessage string `json:"eventMessage"`
}

type scrollRegionATBrowserWindow struct {
	Schema          string                  `json:"schema"`
	Pair            string                  `json:"pair"`
	Route           string                  `json:"route"`
	Challenge       string                  `json:"challenge"`
	CandidateTree   string                  `json:"candidate_tree"`
	ManifestSHA256  string                  `json:"manifest_sha256"`
	Window          scrollRegionATRectangle `json:"window"`
	CandidateRegion scrollRegionATRectangle `json:"candidate_region"`
}

// scrollRegionATChallenge is independently generated for every final capture
// request. A signed envelope must echo it exactly, so prior capture bytes
// cannot be replayed for a fresh attestation request.
type scrollRegionATChallenge struct {
	Schema    string `json:"schema"`
	Challenge string `json:"challenge"`
	IssuedAt  string `json:"issued_at"`
}

// scrollRegionATChallengeClaim is external custody state, not candidate
// evidence. It permits repeat validation of identical immutable bytes while
// rejecting a second receipt for the same independent capture challenge.
type scrollRegionATChallengeClaim struct {
	Schema        string                      `json:"schema"`
	Challenge     string                      `json:"challenge"`
	ReceiptSHA256 string                      `json:"receipt_sha256"`
	Identity      scrollRegionReceiptIdentity `json:"identity"`
}

// scrollRegionATAdapterBridge is test-only plumbing for a real adapter output
// generated under a fake direct-command runtime and then checked by this
// production receipt validator. No root capture-authority key participates.
type scrollRegionATAdapterBridge struct {
	Repository    string `json:"repository"`
	IdentityPath  string `json:"identity_path"`
	ChallengePath string `json:"challenge_path"`
	ReceiptPath   string `json:"receipt_path"`
}

type scrollRegionATDSSESignature struct {
	KeyID     string `json:"keyid"`
	Signature string `json:"sig"`
}

// scrollRegionATDSSEEnvelope is intentionally small and strict: the raw
// payload is signed with DSSE pre-authentication encoding, not a re-marshaled
// caller-selected JSON representation.
type scrollRegionATDSSEEnvelope struct {
	Schema      string                        `json:"schema"`
	PayloadType string                        `json:"payloadType"`
	Payload     string                        `json:"payload"`
	Signatures  []scrollRegionATDSSESignature `json:"signatures"`
}

type scrollRegionATFocusNavigation struct {
	Before string `json:"before"`
	Entry  string `json:"entry"`
	Exit   string `json:"exit"`
}

// scrollRegionATObservation carries a state-specific reported AT outcome. It
// rejects prose-only claims by requiring state, role, name, focus, boundary,
// command, focus traversal, speech, and explicit absence of surprise speech.
type scrollRegionATObservation struct {
	State            string                        `json:"state"`
	Role             string                        `json:"role"`
	Name             string                        `json:"name"`
	Focused          bool                          `json:"focused"`
	Boundary         string                        `json:"boundary"`
	Commands         []string                      `json:"commands"`
	FocusNavigation  scrollRegionATFocusNavigation `json:"focus_navigation"`
	ObservedSpeech   []string                      `json:"observed_speech"`
	UnexpectedSpeech []string                      `json:"unexpected_speech"`
}

type scrollRegionATTraceEvent struct {
	State          string `json:"state"`
	Command        string `json:"command"`
	BeforeBoundary string `json:"before_boundary"`
	AfterBoundary  string `json:"after_boundary"`
	Focused        bool   `json:"focused"`
}

type scrollRegionATTranscript struct {
	Schema              string                      `json:"schema"`
	Challenge           string                      `json:"challenge"`
	Pair                string                      `json:"pair"`
	CapturedAt          string                      `json:"captured_at"`
	Identity            scrollRegionReceiptIdentity `json:"identity"`
	PlatformVersion     string                      `json:"platform_version"`
	BrowserVersion      string                      `json:"browser_version"`
	ScreenReaderVersion string                      `json:"screen_reader_version"`
	Route               string                      `json:"route"`
	Observations        []scrollRegionATObservation `json:"observations"`
}

type scrollRegionATTraceLog struct {
	Schema     string                      `json:"schema"`
	Challenge  string                      `json:"challenge"`
	Pair       string                      `json:"pair"`
	CapturedAt string                      `json:"captured_at"`
	Identity   scrollRegionReceiptIdentity `json:"identity"`
	Route      string                      `json:"route"`
	Events     []scrollRegionATTraceEvent  `json:"events"`
}

type scrollRegionATEvidenceCapture struct {
	Pair                string                           `json:"pair"`
	PlatformVersion     string                           `json:"platform_version"`
	BrowserVersion      string                           `json:"browser_version"`
	ScreenReaderVersion string                           `json:"screen_reader_version"`
	Route               string                           `json:"route"`
	CapturedAt          string                           `json:"captured_at"`
	Observations        []scrollRegionATObservation      `json:"observations"`
	ServedPage          scrollRegionATEvidenceArtifact   `json:"served_page"`
	ServedResponse      scrollRegionATEvidenceArtifact   `json:"served_response"`
	BrowserStates       []scrollRegionATEvidenceArtifact `json:"browser_states"`
	ActionRecords       []scrollRegionATEvidenceArtifact `json:"action_records"`
	VoiceOverCaptions   []scrollRegionATEvidenceArtifact `json:"voiceover_captions"`
	VoiceOverLog        scrollRegionATEvidenceArtifact   `json:"voiceover_log"`
	BrowserWindow       scrollRegionATEvidenceArtifact   `json:"browser_window"`
	Screenshot          scrollRegionATEvidenceArtifact   `json:"screenshot"`
	Transcript          scrollRegionATEvidenceArtifact   `json:"transcript"`
	TraceLog            scrollRegionATEvidenceArtifact   `json:"trace_log"`
	Attestation         scrollRegionATEvidenceArtifact   `json:"attestation"`
}

type scrollRegionATEvidenceReceipt struct {
	Schema    string                          `json:"schema"`
	Status    string                          `json:"status"`
	Challenge string                          `json:"challenge"`
	Identity  scrollRegionReceiptIdentity     `json:"identity"`
	Captures  []scrollRegionATEvidenceCapture `json:"captures"`
}

// scrollRegionATAttestedCapture intentionally excludes the envelope itself;
// all other receipt-visible capture fields must match this signed payload.
type scrollRegionATAttestedCapture struct {
	Pair                string                           `json:"pair"`
	PlatformVersion     string                           `json:"platform_version"`
	BrowserVersion      string                           `json:"browser_version"`
	ScreenReaderVersion string                           `json:"screen_reader_version"`
	Route               string                           `json:"route"`
	CapturedAt          string                           `json:"captured_at"`
	Observations        []scrollRegionATObservation      `json:"observations"`
	ServedPage          scrollRegionATEvidenceArtifact   `json:"served_page"`
	ServedResponse      scrollRegionATEvidenceArtifact   `json:"served_response"`
	BrowserStates       []scrollRegionATEvidenceArtifact `json:"browser_states"`
	ActionRecords       []scrollRegionATEvidenceArtifact `json:"action_records"`
	VoiceOverCaptions   []scrollRegionATEvidenceArtifact `json:"voiceover_captions"`
	VoiceOverLog        scrollRegionATEvidenceArtifact   `json:"voiceover_log"`
	BrowserWindow       scrollRegionATEvidenceArtifact   `json:"browser_window"`
	Screenshot          scrollRegionATEvidenceArtifact   `json:"screenshot"`
	Transcript          scrollRegionATEvidenceArtifact   `json:"transcript"`
	TraceLog            scrollRegionATEvidenceArtifact   `json:"trace_log"`
}

type scrollRegionATAttestationPayload struct {
	Schema    string                        `json:"schema"`
	Challenge string                        `json:"challenge"`
	Identity  scrollRegionReceiptIdentity   `json:"identity"`
	Capture   scrollRegionATAttestedCapture `json:"capture"`
}

type scrollRegionATTrustedKeyManifest struct {
	Schema string                        `json:"schema"`
	Keys   []scrollRegionATTrustedKeyRef `json:"keys"`
}

type scrollRegionATTrustedKeyRef struct {
	KeyID           string   `json:"key_id"`
	PublicKeyPath   string   `json:"public_key_path"`
	PublicKeySHA256 string   `json:"public_key_sha256"`
	Pairs           []string `json:"pairs"`
}

type scrollRegionATTrustedKey struct {
	ID     string
	Pairs  map[string]struct{}
	Public ed25519.PublicKey
}

type scrollRegionATPairContract struct {
	Platform     *regexp.Regexp
	Browser      *regexp.Regexp
	ScreenReader *regexp.Regexp
}

var scrollRegionATPairContracts = map[string]scrollRegionATPairContract{
	"macos-safari-voiceover": {
		Platform:     regexp.MustCompile(`^macOS [1-9][0-9]*(?:\.[0-9]+){1,2}$`),
		Browser:      regexp.MustCompile(`^Safari [1-9][0-9]*(?:\.[0-9]+){1,3}$`),
		ScreenReader: regexp.MustCompile(`^VoiceOver [1-9][0-9]*(?:\.[0-9]+){1,3}$`),
	},
	"macos-chromium-voiceover": {
		Platform:     regexp.MustCompile(`^macOS [1-9][0-9]*(?:\.[0-9]+){1,2}$`),
		Browser:      regexp.MustCompile(`^Chromium [1-9][0-9]*(?:\.[0-9]+){1,3}$`),
		ScreenReader: regexp.MustCompile(`^VoiceOver [1-9][0-9]*(?:\.[0-9]+){1,3}$`),
	},
	"windows-chromium-nvda": {
		Platform:     regexp.MustCompile(`^Windows [1-9][0-9]*(?:\.[0-9]+){0,2}$`),
		Browser:      regexp.MustCompile(`^Chromium [1-9][0-9]*(?:\.[0-9]+){1,3}$`),
		ScreenReader: regexp.MustCompile(`^NVDA [1-9][0-9]*(?:\.[0-9]+){1,3}$`),
	},
}

type scrollRegionATExpectedState struct {
	Name            string
	Boundary        string
	Focused         bool
	RequiredCommand string
	ActionEvent     string
	ExitCommand     string
	ExitEvent       string
}

var scrollRegionATExpectedStates = []scrollRegionATExpectedState{
	{Name: "default", Boundary: "start", Focused: false, RequiredCommand: "Navigate", ActionEvent: "adapter-owned candidate navigation", ExitCommand: "Tab", ExitEvent: "macOS System Events key code 48"},
	{Name: "no-overflow", Boundary: "no-overflow", Focused: true, RequiredCommand: "Tab", ActionEvent: "macOS System Events key code 48", ExitCommand: "Tab", ExitEvent: "macOS System Events key code 48"},
	{Name: "start", Boundary: "start", Focused: true, RequiredCommand: "Home", ActionEvent: "macOS System Events key code 115", ExitCommand: "Tab", ExitEvent: "macOS System Events key code 48"},
	{Name: "middle", Boundary: "middle", Focused: true, RequiredCommand: "PageDown", ActionEvent: "macOS System Events key code 121", ExitCommand: "Tab", ExitEvent: "macOS System Events key code 48"},
	{Name: "end", Boundary: "end", Focused: true, RequiredCommand: "End", ActionEvent: "macOS System Events key code 119", ExitCommand: "Tab", ExitEvent: "macOS System Events key code 48"},
	{Name: "focused", Boundary: "start", Focused: true, RequiredCommand: "Tab", ActionEvent: "macOS System Events key code 48", ExitCommand: "Tab", ExitEvent: "macOS System Events key code 48"},
}

// validateScrollRegionATReceipt validates an externally captured receipt. It
// never creates or upgrades evidence: every identity and artifact claim is
// re-derived or parsed from its authenticated bytes.
func validateScrollRegionATReceipt(repositoryRoot, identityPath, receiptPath string) error {
	return validateScrollRegionATReceiptWithBeforeClaim(repositoryRoot, identityPath, receiptPath, nil)
}

// validateScrollRegionATReceiptWithBeforeClaim retains a narrow test seam for
// proving that identity is checked again immediately before an irreversible
// replay-registry claim. Production always supplies nil.
func validateScrollRegionATReceiptWithBeforeClaim(repositoryRoot, identityPath, receiptPath string, beforeClaim func() error) error {
	candidate, err := verifyScrollRegionCandidateIdentity(repositoryRoot, identityPath)
	if err != nil {
		return fmt.Errorf("verify frozen candidate identity: %w", err)
	}
	challenge, err := readScrollRegionATChallenge(os.Getenv(scrollRegionATChallengeEnvironment))
	if err != nil {
		return err
	}
	trustedKeys, err := loadScrollRegionATTrustedKeys(repositoryRoot)
	if err != nil {
		return err
	}
	receipt, receiptBytes, err := readScrollRegionATEvidenceReceipt(receiptPath)
	if err != nil {
		return err
	}
	if receipt.Schema != scrollRegionATReceiptSchema {
		return fmt.Errorf("AT receipt schema mismatch: got %q", receipt.Schema)
	}
	if receipt.Status != "captured" {
		return fmt.Errorf("AT receipt status must be captured, got %q", receipt.Status)
	}
	if receipt.Identity != scrollRegionReceiptIdentityFromCandidate(candidate) {
		return fmt.Errorf("AT receipt identity does not match independently verified frozen candidate")
	}
	if receipt.Challenge != challenge.Challenge {
		return fmt.Errorf("AT receipt challenge does not match the independently supplied capture challenge")
	}
	if len(receipt.Captures) != 2 {
		return fmt.Errorf("AT receipt must contain exactly Safari+VoiceOver and one supported Chromium capture")
	}
	seenPairs := make(map[string]struct{}, len(receipt.Captures))
	seenArtifactPaths := make(map[string]string, len(receipt.Captures)*12)
	seenArtifactHashes := make(map[string]string, len(receipt.Captures)*12)
	chromiumCaptures := 0
	for _, capture := range receipt.Captures {
		contract, known := scrollRegionATPairContracts[capture.Pair]
		if !known {
			return fmt.Errorf("AT receipt has unexpected pair %q", capture.Pair)
		}
		if _, duplicate := seenPairs[capture.Pair]; duplicate {
			return fmt.Errorf("AT receipt repeats pair %q", capture.Pair)
		}
		seenPairs[capture.Pair] = struct{}{}
		if capture.Pair == "macos-chromium-voiceover" || capture.Pair == "windows-chromium-nvda" {
			chromiumCaptures++
		}
		if err := validateScrollRegionATCapture(capture, contract, receipt.Identity, challenge, trustedKeys, seenArtifactPaths, seenArtifactHashes); err != nil {
			return fmt.Errorf("AT pair %q: %w", capture.Pair, err)
		}
	}
	if _, captured := seenPairs["macos-safari-voiceover"]; !captured {
		return fmt.Errorf("AT receipt omits required pair %q", "macos-safari-voiceover")
	}
	if chromiumCaptures != 1 {
		return fmt.Errorf("AT receipt must contain exactly one supported Chromium pair")
	}
	if beforeClaim != nil {
		if err := beforeClaim(); err != nil {
			return fmt.Errorf("prepare final candidate identity recheck: %w", err)
		}
	}
	finalCandidate, err := verifyScrollRegionCandidateIdentity(repositoryRoot, identityPath)
	if err != nil {
		return fmt.Errorf("verify final frozen candidate identity before replay claim: %w", err)
	}
	if !reflect.DeepEqual(finalCandidate, candidate) || receipt.Identity != scrollRegionReceiptIdentityFromCandidate(finalCandidate) {
		return fmt.Errorf("final frozen candidate identity changed before replay claim")
	}
	finalChallenge, err := readScrollRegionATChallenge(os.Getenv(scrollRegionATChallengeEnvironment))
	if err != nil {
		return fmt.Errorf("re-read final independently supplied capture challenge: %w", err)
	}
	if finalChallenge != challenge || receipt.Challenge != finalChallenge.Challenge {
		return fmt.Errorf("final capture challenge changed before replay claim")
	}
	if err := claimScrollRegionATChallenge(os.Getenv(scrollRegionATReplayRegistryEnvironment), challenge, receipt.Identity, receiptBytes); err != nil {
		return err
	}
	return nil
}

func readScrollRegionATChallenge(path string) (scrollRegionATChallenge, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return scrollRegionATChallenge{}, fmt.Errorf("%s must name an independently generated capture challenge", scrollRegionATChallengeEnvironment)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return scrollRegionATChallenge{}, fmt.Errorf("read AT capture challenge: %w", err)
	}
	var challenge scrollRegionATChallenge
	if err := scrollRegionDecodeStrictJSON(content, &challenge); err != nil {
		return scrollRegionATChallenge{}, fmt.Errorf("decode AT capture challenge: %w", err)
	}
	if challenge.Schema != scrollRegionATChallengeSchema || !scrollRegionSHA256Pattern.MatchString(challenge.Challenge) {
		return scrollRegionATChallenge{}, fmt.Errorf("AT capture challenge schema or random challenge is invalid")
	}
	if _, err := time.Parse(time.RFC3339, challenge.IssuedAt); err != nil {
		return scrollRegionATChallenge{}, fmt.Errorf("AT capture challenge issued_at must be RFC3339: %w", err)
	}
	return challenge, nil
}

func loadScrollRegionATTrustedKeys(repositoryRoot string) (map[string]scrollRegionATTrustedKey, error) {
	manifestPath := filepath.Join(repositoryRoot, "tests", "external", "scrollregion-a11y", "attestation-keys.json")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read trusted AT key manifest: %w", err)
	}
	var manifest scrollRegionATTrustedKeyManifest
	if err := scrollRegionDecodeStrictJSON(content, &manifest); err != nil {
		return nil, fmt.Errorf("decode trusted AT key manifest: %w", err)
	}
	if manifest.Schema != scrollRegionATTrustedKeysSchema || len(manifest.Keys) < 2 {
		return nil, fmt.Errorf("trusted AT key manifest schema or key set is invalid")
	}
	keys := make(map[string]scrollRegionATTrustedKey, len(manifest.Keys))
	for _, reference := range manifest.Keys {
		if !regexp.MustCompile(`^[a-z0-9][a-z0-9-]{2,80}$`).MatchString(reference.KeyID) || !scrollRegionCandidatePathIsSafe(reference.PublicKeyPath) || !scrollRegionSHA256Pattern.MatchString(reference.PublicKeySHA256) || len(reference.Pairs) == 0 {
			return nil, fmt.Errorf("trusted AT key reference is invalid")
		}
		if _, duplicate := keys[reference.KeyID]; duplicate {
			return nil, fmt.Errorf("trusted AT key manifest repeats key ID %q", reference.KeyID)
		}
		pairs := make(map[string]struct{}, len(reference.Pairs))
		for _, pair := range reference.Pairs {
			if _, known := scrollRegionATPairContracts[pair]; !known {
				return nil, fmt.Errorf("trusted AT key %q has unknown pair %q", reference.KeyID, pair)
			}
			if _, duplicate := pairs[pair]; duplicate {
				return nil, fmt.Errorf("trusted AT key %q repeats pair %q", reference.KeyID, pair)
			}
			pairs[pair] = struct{}{}
		}
		publicPath := filepath.Join(repositoryRoot, filepath.FromSlash(reference.PublicKeyPath))
		publicBytes, err := os.ReadFile(publicPath)
		if err != nil {
			return nil, fmt.Errorf("read trusted AT public key %q: %w", reference.KeyID, err)
		}
		if scrollRegionBFullSHA256(publicBytes) != reference.PublicKeySHA256 {
			return nil, fmt.Errorf("trusted AT public key %q fingerprint mismatch", reference.KeyID)
		}
		block, remainder := pem.Decode(publicBytes)
		if block == nil || block.Type != "PUBLIC KEY" || len(bytes.TrimSpace(remainder)) != 0 {
			return nil, fmt.Errorf("trusted AT public key %q PEM is invalid", reference.KeyID)
		}
		parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse trusted AT public key %q: %w", reference.KeyID, err)
		}
		public, ok := parsed.(ed25519.PublicKey)
		if !ok || len(public) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("trusted AT public key %q is not Ed25519", reference.KeyID)
		}
		keys[reference.KeyID] = scrollRegionATTrustedKey{ID: reference.KeyID, Pairs: pairs, Public: public}
	}
	return keys, nil
}

func readScrollRegionATEvidenceReceipt(path string) (scrollRegionATEvidenceReceipt, []byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return scrollRegionATEvidenceReceipt{}, nil, fmt.Errorf("read AT receipt: %w", err)
	}
	var receipt scrollRegionATEvidenceReceipt
	if err := scrollRegionDecodeStrictJSON(content, &receipt); err != nil {
		return scrollRegionATEvidenceReceipt{}, nil, fmt.Errorf("decode AT receipt: %w", err)
	}
	return receipt, content, nil
}

func claimScrollRegionATChallenge(registryPath string, challenge scrollRegionATChallenge, identity scrollRegionReceiptIdentity, receiptBytes []byte) error {
	registryPath = strings.TrimSpace(registryPath)
	if registryPath == "" {
		return fmt.Errorf("%s must name an external owner-only challenge registry", scrollRegionATReplayRegistryEnvironment)
	}
	registryPath, err := filepath.Abs(registryPath)
	if err != nil {
		return fmt.Errorf("resolve AT challenge registry: %w", err)
	}
	info, err := os.Lstat(registryPath)
	if err != nil {
		return fmt.Errorf("stat AT challenge registry: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("AT challenge registry must be an owner-only non-symlink directory")
	}
	claim := scrollRegionATChallengeClaim{
		Schema:        scrollRegionATChallengeClaimSchema,
		Challenge:     challenge.Challenge,
		ReceiptSHA256: scrollRegionBFullSHA256(receiptBytes),
		Identity:      identity,
	}
	encoded, err := json.Marshal(claim)
	if err != nil {
		return fmt.Errorf("encode AT challenge claim: %w", err)
	}
	claimPath := filepath.Join(registryPath, challenge.Challenge+".json")
	file, err := os.OpenFile(claimPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err == nil {
		if _, writeErr := file.Write(append(encoded, '\n')); writeErr != nil {
			_ = file.Close()
			return fmt.Errorf("write AT challenge claim: %w", writeErr)
		}
		if syncErr := file.Sync(); syncErr != nil {
			_ = file.Close()
			return fmt.Errorf("sync AT challenge claim: %w", syncErr)
		}
		if closeErr := file.Close(); closeErr != nil {
			return fmt.Errorf("close AT challenge claim: %w", closeErr)
		}
		return nil
	}
	if !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("create AT challenge claim: %w", err)
	}
	existingInfo, statErr := os.Lstat(claimPath)
	if statErr != nil || !existingInfo.Mode().IsRegular() || existingInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("existing AT challenge claim is not a regular non-symlink file")
	}
	existingBytes, readErr := os.ReadFile(claimPath)
	if readErr != nil {
		return fmt.Errorf("read existing AT challenge claim: %w", readErr)
	}
	var existing scrollRegionATChallengeClaim
	if decodeErr := scrollRegionDecodeStrictJSON(existingBytes, &existing); decodeErr != nil {
		return fmt.Errorf("decode existing AT challenge claim: %w", decodeErr)
	}
	if existing == claim {
		return nil
	}
	return fmt.Errorf("AT capture challenge already claimed by a different receipt")
}

func scrollRegionDecodeStrictJSON(content []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON values")
		}
		return fmt.Errorf("trailing data: %w", err)
	}
	return nil
}

func validateScrollRegionATCapture(capture scrollRegionATEvidenceCapture, contract scrollRegionATPairContract, identity scrollRegionReceiptIdentity, challenge scrollRegionATChallenge, trustedKeys map[string]scrollRegionATTrustedKey, seenPaths, seenHashes map[string]string) error {
	// Verify the pair-specific external signature before trusting any caller
	// supplied JSON, artifact hash, observation, or version field.
	attestationBytes, err := readScrollRegionATEvidenceArtifact(capture.Attestation, "signed attestation", seenPaths, seenHashes)
	if err != nil {
		return err
	}
	payload, err := validateScrollRegionATAttestation(attestationBytes, capture, identity, challenge, trustedKeys)
	if err != nil {
		return err
	}
	if payload.Challenge != challenge.Challenge {
		return fmt.Errorf("signed attestation challenge mismatch")
	}
	if payload.Identity != identity || !reflect.DeepEqual(payload.Capture, scrollRegionATAttestedCaptureFromEvidence(capture)) {
		return fmt.Errorf("signed attestation does not bind the exact capture identity and artifact claims")
	}
	derived, err := validateScrollRegionATRawProvenance(capture, identity, challenge, seenPaths, seenHashes)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(capture.Observations, derived) {
		return fmt.Errorf("structured observations are not derived from the signed direct browser and VoiceOver bytes")
	}

	// Artifact structural validation deliberately precedes semantics. A signed
	// solid-color PNG is still not visual AT provenance for the named region.
	screenshot, err := readScrollRegionATEvidenceArtifact(capture.Screenshot, "screenshot", seenPaths, seenHashes)
	if err != nil {
		return err
	}
	if err := validateScrollRegionATScreenshot(screenshot); err != nil {
		return err
	}

	if !contract.Platform.MatchString(capture.PlatformVersion) {
		return fmt.Errorf("platform version %q violates pair contract", capture.PlatformVersion)
	}
	if !contract.Browser.MatchString(capture.BrowserVersion) {
		return fmt.Errorf("browser version %q violates pair contract", capture.BrowserVersion)
	}
	if !contract.ScreenReader.MatchString(capture.ScreenReaderVersion) {
		return fmt.Errorf("screen-reader version %q violates pair contract", capture.ScreenReaderVersion)
	}
	if capture.Route != scrollRegionBFullRoute {
		return fmt.Errorf("route mismatch: got %q", capture.Route)
	}
	if _, err := time.Parse(time.RFC3339, capture.CapturedAt); err != nil {
		return fmt.Errorf("captured_at must be RFC3339: %w", err)
	}
	if err := validateScrollRegionATObservations(capture.Observations); err != nil {
		return err
	}

	transcriptBytes, err := readScrollRegionATEvidenceArtifact(capture.Transcript, "transcript", seenPaths, seenHashes)
	if err != nil {
		return err
	}
	var transcript scrollRegionATTranscript
	if err := scrollRegionDecodeStrictJSON(transcriptBytes, &transcript); err != nil {
		return fmt.Errorf("transcript schema: %w", err)
	}
	if transcript.Schema != scrollRegionATTranscriptSchema || transcript.Challenge != challenge.Challenge || transcript.Pair != capture.Pair || transcript.Identity != identity ||
		transcript.PlatformVersion != capture.PlatformVersion || transcript.BrowserVersion != capture.BrowserVersion ||
		transcript.ScreenReaderVersion != capture.ScreenReaderVersion || transcript.Route != capture.Route ||
		transcript.CapturedAt != capture.CapturedAt || !reflect.DeepEqual(transcript.Observations, capture.Observations) {
		return fmt.Errorf("transcript does not bind the exact capture identity, pair, versions, route, and observations")
	}
	if _, err := time.Parse(time.RFC3339, transcript.CapturedAt); err != nil {
		return fmt.Errorf("transcript captured_at must be RFC3339: %w", err)
	}
	traceBytes, err := readScrollRegionATEvidenceArtifact(capture.TraceLog, "trace log", seenPaths, seenHashes)
	if err != nil {
		return err
	}
	var trace scrollRegionATTraceLog
	if err := scrollRegionDecodeStrictJSON(traceBytes, &trace); err != nil {
		return fmt.Errorf("trace-log schema: %w", err)
	}
	if trace.Schema != scrollRegionATTraceLogSchema || trace.Challenge != challenge.Challenge || trace.Pair != capture.Pair || trace.Identity != identity ||
		trace.Route != capture.Route || trace.CapturedAt != capture.CapturedAt {
		return fmt.Errorf("trace log does not bind the exact capture identity, pair, route, and timestamp")
	}
	if _, err := time.Parse(time.RFC3339, trace.CapturedAt); err != nil {
		return fmt.Errorf("trace-log captured_at must be RFC3339: %w", err)
	}
	if err := validateScrollRegionATTraceEvents(trace.Events, capture.Observations); err != nil {
		return err
	}
	return nil
}

// validateScrollRegionATRawProvenance rejects signer-authored transcript-only
// claims. Every outcome is rebuilt from one exact candidate response, direct
// browser state bytes, and direct VoiceOver bytes before semantic validation.
func validateScrollRegionATRawProvenance(capture scrollRegionATEvidenceCapture, identity scrollRegionReceiptIdentity, challenge scrollRegionATChallenge, seenPaths, seenHashes map[string]string) ([]scrollRegionATObservation, error) {
	servedHTML, err := readScrollRegionATEvidenceArtifact(capture.ServedPage, "served candidate HTML", seenPaths, seenHashes)
	if err != nil {
		return nil, err
	}
	servedResponse, err := readScrollRegionATEvidenceArtifact(capture.ServedResponse, "served candidate response", seenPaths, seenHashes)
	if err != nil {
		return nil, err
	}
	var page scrollRegionATServedPage
	if err := scrollRegionDecodeStrictJSON(servedResponse, &page); err != nil {
		return nil, fmt.Errorf("served candidate response schema: %w", err)
	}
	defaultToken := scrollRegionATActionToken(challenge.Challenge, "default")
	if page.Schema != scrollRegionATServedPageSchema || page.Status != 200 || page.Challenge != challenge.Challenge || page.CandidateTree != identity.CandidateTree || page.ManifestSHA256 != identity.ManifestSHA256 || page.Pair != capture.Pair || page.ActionState != "default" || page.ActionToken != defaultToken || page.BodySHA256 != scrollRegionBFullSHA256(servedHTML) {
		return nil, fmt.Errorf("served candidate response does not bind exact challenge, pair, identity, and HTML bytes")
	}
	parsedURL, err := url.Parse(page.URL)
	if err != nil || parsedURL.Scheme != "http" || parsedURL.User != nil || parsedURL.Path != scrollRegionBFullRoute || parsedURL.Query().Get("t-gs-011-at-capture") != challenge.Challenge || parsedURL.Query().Get("t-gs-011-at-state") != "default" || parsedURL.Query().Get("t-gs-011-at-action-token") != defaultToken || !isScrollRegionATLoopbackHost(parsedURL.Hostname()) {
		return nil, fmt.Errorf("served candidate response URL is not adapter-owned loopback challenge route")
	}
	for _, required := range []string{
		`name="goshtoso-t-gs-011-at-challenge" content="` + challenge.Challenge + `"`,
		`name="goshtoso-t-gs-011-candidate-tree" content="` + identity.CandidateTree + `"`,
		`name="goshtoso-t-gs-011-manifest-sha256" content="` + identity.ManifestSHA256 + `"`,
		`name="goshtoso-t-gs-011-at-pair" content="` + capture.Pair + `"`,
		`name="goshtoso-t-gs-011-at-action-state" content="default"`,
		`name="goshtoso-t-gs-011-at-action-token" content="` + defaultToken + `"`,
		`role="region"`,
		`data-goshtoso-scroll-viewport`,
	} {
		if !bytes.Contains(servedHTML, []byte(required)) {
			return nil, fmt.Errorf("served candidate HTML lacks bound content %q", required)
		}
	}

	rawWindow, err := readScrollRegionATEvidenceArtifact(capture.BrowserWindow, "direct browser window", seenPaths, seenHashes)
	if err != nil {
		return nil, err
	}
	var window scrollRegionATBrowserWindow
	if err := scrollRegionDecodeStrictJSON(rawWindow, &window); err != nil {
		return nil, fmt.Errorf("direct browser window schema: %w", err)
	}
	if window.Schema != scrollRegionATBrowserStateSchema || window.Pair != capture.Pair || window.Route != capture.Route || window.Challenge != challenge.Challenge || window.CandidateTree != identity.CandidateTree || window.ManifestSHA256 != identity.ManifestSHA256 || window.Window.Width < 100 || window.Window.Height < 80 || window.CandidateRegion.Width <= 0 || window.CandidateRegion.Height <= 0 {
		return nil, fmt.Errorf("direct browser window does not prove bounded candidate-region capture")
	}

	if len(capture.BrowserStates) != len(scrollRegionATExpectedStates)*3 || len(capture.ActionRecords) != len(scrollRegionATExpectedStates) || len(capture.VoiceOverCaptions) != len(scrollRegionATExpectedStates) {
		return nil, fmt.Errorf("direct action, browser before/after/exit, and VoiceOver artifacts must cover every required state")
	}
	derived := make([]scrollRegionATObservation, 0, len(scrollRegionATExpectedStates))
	actionRecords := make([]scrollRegionATActionRecord, 0, len(scrollRegionATExpectedStates))
	seenTokens := make(map[string]struct{}, len(scrollRegionATExpectedStates))
	var previousLogEnd time.Time
	for index, expected := range scrollRegionATExpectedStates {
		rawAction, err := readScrollRegionATEvidenceArtifact(capture.ActionRecords[index], "direct action record", seenPaths, seenHashes)
		if err != nil {
			return nil, err
		}
		var action scrollRegionATActionRecord
		if err := scrollRegionDecodeStrictJSON(rawAction, &action); err != nil {
			return nil, fmt.Errorf("direct action record %q schema: %w", expected.Name, err)
		}
		token := scrollRegionATActionToken(challenge.Challenge, expected.Name)
		if action.Schema != scrollRegionATActionRecordSchema || action.Pair != capture.Pair || action.State != expected.Name || action.Command != expected.RequiredCommand || action.ActionEvent != expected.ActionEvent || action.ExitCommand != expected.ExitCommand || action.ExitEvent != expected.ExitEvent || action.Route != capture.Route || action.Challenge != challenge.Challenge || action.CandidateTree != identity.CandidateTree || action.ManifestSHA256 != identity.ManifestSHA256 || action.ActionToken != token || action.VoiceOverPID < 2 || action.VoiceOverSubsystem != scrollRegionATVoiceOverSubsystem {
			return nil, fmt.Errorf("direct action record %q does not bind exact adapter action contract, token, VoiceOver process, and candidate", expected.Name)
		}
		if _, duplicate := seenTokens[action.ActionToken]; duplicate {
			return nil, fmt.Errorf("direct action record %q reuses action token", expected.Name)
		}
		seenTokens[action.ActionToken] = struct{}{}
		beforeAt, logStartedAt, _, afterAt, _, exitAt, logEndedAt, timelineErr := scrollRegionATActionTimeline(action)
		if timelineErr != nil || (!previousLogEnd.IsZero() && !logStartedAt.After(previousLogEnd)) {
			return nil, fmt.Errorf("direct action record %q has stale, overlapping, or incomplete causal action interval", expected.Name)
		}
		previousLogEnd = logEndedAt
		actionRecords = append(actionRecords, action)

		phases := make(map[string]scrollRegionATBrowserState, 3)
		for phaseIndex, phase := range []string{"before", "after", "exit"} {
			rawBrowser, err := readScrollRegionATEvidenceArtifact(capture.BrowserStates[index*3+phaseIndex], "direct browser "+phase+" state", seenPaths, seenHashes)
			if err != nil {
				return nil, err
			}
			var browser scrollRegionATBrowserState
			if err := scrollRegionDecodeStrictJSON(rawBrowser, &browser); err != nil {
				return nil, fmt.Errorf("direct browser state %q phase %q schema: %w", expected.Name, phase, err)
			}
			if browser.Schema != scrollRegionATBrowserStateSchema || browser.Pair != capture.Pair || browser.State != expected.Name || browser.Command != expected.RequiredCommand || browser.Route != capture.Route || browser.Challenge != challenge.Challenge || browser.CandidateTree != identity.CandidateTree || browser.ManifestSHA256 != identity.ManifestSHA256 || browser.ActionToken != token || browser.Phase != phase || browser.CandidateRegion.Width <= 0 || browser.CandidateRegion.Height <= 0 {
				return nil, fmt.Errorf("direct browser state %q phase %q does not bind exact action and candidate region", expected.Name, phase)
			}
			observedAt, observedErr := scrollRegionATUTC(browser.ObservedAt)
			wantObservedAt := map[string]time.Time{"before": beforeAt, "after": afterAt, "exit": exitAt}[phase]
			if observedErr != nil || !observedAt.Equal(wantObservedAt) || !scrollRegionATBrowserSnapshotValid(browser.Snapshot, expected, phase) {
				return nil, fmt.Errorf("direct browser state %q phase %q lacks observed active-element, scroll, cue, or interval provenance", expected.Name, phase)
			}
			phases[phase] = browser
		}
		name := "Activity history"
		if expected.Name == "no-overflow" {
			name = "Current release"
		}
		before, after, exit := phases["before"].Snapshot, phases["after"].Snapshot, phases["exit"].Snapshot
		if after.RegionRole != "region" || after.RegionName != name || after.RegionFocused != expected.Focused || after.Boundary != expected.Boundary || !scrollRegionATMeaningful(before.ActiveRole) || !scrollRegionATMeaningful(before.ActiveName) || !scrollRegionATMeaningful(after.ActiveRole) || !scrollRegionATMeaningful(after.ActiveName) || !scrollRegionATMeaningful(exit.ActiveRole) || !scrollRegionATMeaningful(exit.ActiveName) {
			return nil, fmt.Errorf("direct browser state %q does not prove actual role, name, focus, boundary, and before/after/exit active elements", expected.Name)
		}
		if expected.Focused && (after.ActiveRole != "region" || after.ActiveName != name || exit.ActiveRole == "region") {
			return nil, fmt.Errorf("direct browser state %q does not prove actual focus entry and exit", expected.Name)
		}
		if err := scrollRegionATValidateCausalTransition(expected, before, after, exit); err != nil {
			return nil, fmt.Errorf("direct browser state %q: %w", expected.Name, err)
		}
		rawVoiceOver, err := readScrollRegionATEvidenceArtifact(capture.VoiceOverCaptions[index], "direct VoiceOver caption", seenPaths, seenHashes)
		if err != nil {
			return nil, err
		}
		phrase, err := scrollRegionATVoiceOverPhrase(rawVoiceOver, name, action)
		if err != nil {
			return nil, fmt.Errorf("direct VoiceOver state %q: %w", expected.Name, err)
		}
		derived = append(derived, scrollRegionATObservation{
			State: expected.Name, Role: after.RegionRole, Name: after.RegionName, Focused: after.RegionFocused, Boundary: after.Boundary,
			Commands: []string{expected.RequiredCommand}, FocusNavigation: scrollRegionATFocusNavigation{Before: scrollRegionATSnapshotDescription(before), Entry: scrollRegionATSnapshotDescription(after), Exit: scrollRegionATSnapshotDescription(exit)}, ObservedSpeech: []string{phrase}, UnexpectedSpeech: []string{},
		})
	}
	rawVoiceOverLog, err := readScrollRegionATEvidenceArtifact(capture.VoiceOverLog, "direct VoiceOver system log", seenPaths, seenHashes)
	if err != nil {
		return nil, err
	}
	var combinedEvents []scrollRegionATVoiceOverLogEvent
	if err := scrollRegionDecodeStrictJSON(rawVoiceOverLog, &combinedEvents); err != nil || len(combinedEvents) < len(scrollRegionATExpectedStates) {
		return nil, fmt.Errorf("direct VoiceOver system log must be a nonempty action-bound JSON event set")
	}
	for index, expected := range scrollRegionATExpectedStates {
		name := "Activity history"
		if expected.Name == "no-overflow" {
			name = "Current release"
		}
		if _, err := scrollRegionATVoiceOverPhrase(rawVoiceOverLog, name, actionRecords[index]); err != nil {
			return nil, fmt.Errorf("direct VoiceOver system log state %q lacks action-bound named-region event: %w", expected.Name, err)
		}
	}
	return derived, nil
}

func scrollRegionATUTC(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || parsed.Location() != time.UTC {
		return time.Time{}, fmt.Errorf("must be UTC RFC3339Nano")
	}
	return parsed, nil
}

func scrollRegionATActionTimeline(action scrollRegionATActionRecord) (time.Time, time.Time, time.Time, time.Time, time.Time, time.Time, time.Time, error) {
	values := []string{action.BeforeAt, action.LogStartedAt, action.ActionIssuedAt, action.AfterAt, action.ExitIssuedAt, action.ExitAt, action.LogEndedAt}
	parsed := make([]time.Time, len(values))
	for index, value := range values {
		var err error
		parsed[index], err = scrollRegionATUTC(value)
		if err != nil {
			return time.Time{}, time.Time{}, time.Time{}, time.Time{}, time.Time{}, time.Time{}, time.Time{}, err
		}
		if index > 0 && !parsed[index].After(parsed[index-1]) {
			return time.Time{}, time.Time{}, time.Time{}, time.Time{}, time.Time{}, time.Time{}, time.Time{}, fmt.Errorf("timestamps must be strictly monotonic")
		}
	}
	return parsed[0], parsed[1], parsed[2], parsed[3], parsed[4], parsed[5], parsed[6], nil
}

func scrollRegionATValidateCausalTransition(expected scrollRegionATExpectedState, before, after, exit scrollRegionATBrowserSnapshot) error {
	switch expected.Name {
	case "start":
		if before.Boundary == "start" || after.Boundary != "start" || !after.RegionFocused {
			return fmt.Errorf("Home must move non-start boundary to focused start")
		}
	case "middle":
		if before.Boundary != "start" || after.Boundary != "middle" || !after.RegionFocused {
			return fmt.Errorf("PageDown must move start boundary to focused middle")
		}
	case "end":
		if before.Boundary == "end" || after.Boundary != "end" || !after.RegionFocused {
			return fmt.Errorf("End must move non-end boundary to focused end")
		}
	case "focused":
		if before.RegionFocused || before.ActiveRole == "region" || !after.RegionFocused || after.ActiveRole != "region" || exit.RegionFocused || exit.ActiveRole == "region" {
			return fmt.Errorf("Tab must prove outside-to-region traversal and exit")
		}
	}
	return nil
}

func scrollRegionATBrowserSnapshotValid(snapshot scrollRegionATBrowserSnapshot, expected scrollRegionATExpectedState, phase string) bool {
	if !scrollRegionATMeaningful(snapshot.RegionRole) || !scrollRegionATMeaningful(snapshot.RegionName) || !scrollRegionATMeaningful(snapshot.ActiveRole) || !scrollRegionATMeaningful(snapshot.ActiveName) || snapshot.ScrollTop < 0 || snapshot.ClientHeight <= 0 || snapshot.ScrollHeight < snapshot.ClientHeight {
		return false
	}
	if phase == "after" && (snapshot.RegionFocused != expected.Focused || snapshot.Boundary != expected.Boundary) {
		return false
	}
	if expected.Name == "no-overflow" {
		return snapshot.ScrollHeight <= snapshot.ClientHeight+1 && !snapshot.StartCueVisible && !snapshot.EndCueVisible
	}
	return snapshot.ScrollHeight > snapshot.ClientHeight
}

func isScrollRegionATLoopbackHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "localhost" {
		return true
	}
	parsed := net.ParseIP(host)
	return parsed != nil && parsed.IsLoopback()
}

func scrollRegionATVoiceOverPhrase(raw []byte, name string, action scrollRegionATActionRecord) (string, error) {
	var events []scrollRegionATVoiceOverLogEvent
	if err := scrollRegionDecodeStrictJSON(raw, &events); err != nil || len(events) == 0 {
		return "", fmt.Errorf("raw output must be action-bound VoiceOver JSON events")
	}
	_, _, actionIssuedAt, _, _, _, logEndedAt, err := scrollRegionATActionTimeline(action)
	if err != nil {
		return "", err
	}
	for _, event := range events {
		timestamp, timestampErr := scrollRegionATUTC(event.Timestamp)
		if timestampErr != nil || !timestamp.After(actionIssuedAt) || timestamp.After(logEndedAt) || event.ProcessID != action.VoiceOverPID || event.Subsystem != action.VoiceOverSubsystem {
			continue
		}
		message := strings.TrimSpace(event.EventMessage)
		lower := strings.ToLower(message)
		if strings.Contains(lower, "voiceover") && strings.Contains(lower, "region") && strings.Contains(lower, strings.ToLower(name)) {
			return message, nil
		}
	}
	return "", fmt.Errorf("raw output lacks named-region VoiceOver event in exact action interval")
}

func scrollRegionATAttestedCaptureFromEvidence(capture scrollRegionATEvidenceCapture) scrollRegionATAttestedCapture {
	return scrollRegionATAttestedCapture{
		Pair:                capture.Pair,
		PlatformVersion:     capture.PlatformVersion,
		BrowserVersion:      capture.BrowserVersion,
		ScreenReaderVersion: capture.ScreenReaderVersion,
		Route:               capture.Route,
		CapturedAt:          capture.CapturedAt,
		Observations:        capture.Observations,
		ServedPage:          capture.ServedPage,
		ServedResponse:      capture.ServedResponse,
		BrowserStates:       capture.BrowserStates,
		ActionRecords:       capture.ActionRecords,
		VoiceOverCaptions:   capture.VoiceOverCaptions,
		VoiceOverLog:        capture.VoiceOverLog,
		BrowserWindow:       capture.BrowserWindow,
		Screenshot:          capture.Screenshot,
		Transcript:          capture.Transcript,
		TraceLog:            capture.TraceLog,
	}
}

func validateScrollRegionATAttestation(content []byte, capture scrollRegionATEvidenceCapture, identity scrollRegionReceiptIdentity, challenge scrollRegionATChallenge, trustedKeys map[string]scrollRegionATTrustedKey) (scrollRegionATAttestationPayload, error) {
	var envelope scrollRegionATDSSEEnvelope
	if err := scrollRegionDecodeStrictJSON(content, &envelope); err != nil {
		return scrollRegionATAttestationPayload{}, fmt.Errorf("signed attestation schema: %w", err)
	}
	if envelope.Schema != scrollRegionATAttestationSchema || envelope.PayloadType != scrollRegionATAttestationPayloadType || len(envelope.Signatures) != 1 {
		return scrollRegionATAttestationPayload{}, fmt.Errorf("signed attestation envelope is invalid")
	}
	rawPayload, err := base64.StdEncoding.DecodeString(envelope.Payload)
	if err != nil || len(rawPayload) == 0 {
		return scrollRegionATAttestationPayload{}, fmt.Errorf("signed attestation payload is not base64")
	}
	signature := envelope.Signatures[0]
	trustedKey, known := trustedKeys[signature.KeyID]
	if !known {
		return scrollRegionATAttestationPayload{}, fmt.Errorf("signed attestation uses untrusted key %q", signature.KeyID)
	}
	if _, allowed := trustedKey.Pairs[capture.Pair]; !allowed {
		return scrollRegionATAttestationPayload{}, fmt.Errorf("signed attestation key %q is not trusted for pair %q", signature.KeyID, capture.Pair)
	}
	rawSignature, err := base64.StdEncoding.DecodeString(signature.Signature)
	if err != nil || len(rawSignature) != ed25519.SignatureSize {
		return scrollRegionATAttestationPayload{}, fmt.Errorf("signed attestation signature is invalid")
	}
	if !ed25519.Verify(trustedKey.Public, scrollRegionATDSSEPAE(envelope.PayloadType, rawPayload), rawSignature) {
		return scrollRegionATAttestationPayload{}, fmt.Errorf("signed attestation signature verification failed")
	}
	var payload scrollRegionATAttestationPayload
	if err := scrollRegionDecodeStrictJSON(rawPayload, &payload); err != nil {
		return scrollRegionATAttestationPayload{}, fmt.Errorf("signed attestation payload schema: %w", err)
	}
	if payload.Schema != scrollRegionATAttestationPayloadSchema || payload.Challenge != challenge.Challenge || payload.Identity != identity || payload.Capture.Pair != capture.Pair {
		return scrollRegionATAttestationPayload{}, fmt.Errorf("signed attestation payload does not bind challenge, identity, and pair")
	}
	return payload, nil
}

func scrollRegionATDSSEPAE(payloadType string, payload []byte) []byte {
	return []byte(fmt.Sprintf("DSSEv1 %d %s %d %s", len([]byte(payloadType)), payloadType, len(payload), payload))
}

func validateScrollRegionATObservations(observations []scrollRegionATObservation) error {
	if len(observations) != len(scrollRegionATExpectedStates) {
		return fmt.Errorf("must contain exactly %d state observations", len(scrollRegionATExpectedStates))
	}
	seen := make(map[string]struct{}, len(observations))
	for _, expected := range scrollRegionATExpectedStates {
		var observation *scrollRegionATObservation
		for index := range observations {
			if observations[index].State == expected.Name {
				observation = &observations[index]
				break
			}
		}
		if observation == nil {
			return fmt.Errorf("missing state observation %q", expected.Name)
		}
		if _, duplicate := seen[observation.State]; duplicate {
			return fmt.Errorf("duplicate state observation %q", observation.State)
		}
		seen[observation.State] = struct{}{}
		if observation.Role != "region" {
			return fmt.Errorf("state %q observed role must be region", expected.Name)
		}
		wantName := "Activity history"
		if expected.Name == "no-overflow" {
			wantName = "Current release"
		}
		if observation.Name != wantName {
			return fmt.Errorf("state %q observed name mismatch: got %q, want %q", expected.Name, observation.Name, wantName)
		}
		if observation.Focused != expected.Focused {
			return fmt.Errorf("state %q focused mismatch", expected.Name)
		}
		if observation.Boundary != expected.Boundary {
			return fmt.Errorf("state %q boundary mismatch: got %q, want %q", expected.Name, observation.Boundary, expected.Boundary)
		}
		if !slices.Contains(observation.Commands, expected.RequiredCommand) {
			return fmt.Errorf("state %q commands must include %q", expected.Name, expected.RequiredCommand)
		}
		for _, command := range observation.Commands {
			if !scrollRegionATMeaningful(command) {
				return fmt.Errorf("state %q has generic command %q", expected.Name, command)
			}
		}
		if !scrollRegionATMeaningful(observation.FocusNavigation.Before) || !scrollRegionATMeaningful(observation.FocusNavigation.Entry) || !scrollRegionATMeaningful(observation.FocusNavigation.Exit) {
			return fmt.Errorf("state %q focus navigation is generic or incomplete", expected.Name)
		}
		entry := strings.ToLower(observation.FocusNavigation.Entry)
		if expected.Focused && (!strings.Contains(entry, "region") || !strings.Contains(entry, strings.ToLower(wantName))) {
			return fmt.Errorf("state %q focus entry must identify the observed named region", expected.Name)
		}
		if observation.ObservedSpeech == nil || len(observation.ObservedSpeech) == 0 {
			return fmt.Errorf("state %q has no observed speech", expected.Name)
		}
		speech := strings.ToLower(strings.Join(observation.ObservedSpeech, " "))
		if !strings.Contains(speech, "region") || !strings.Contains(speech, strings.ToLower(wantName)) {
			return fmt.Errorf("state %q observed speech does not report named region role and name", expected.Name)
		}
		for _, phrase := range observation.ObservedSpeech {
			if !scrollRegionATMeaningful(phrase) {
				return fmt.Errorf("state %q has generic observed speech", expected.Name)
			}
		}
		if observation.UnexpectedSpeech == nil || len(observation.UnexpectedSpeech) != 0 {
			return fmt.Errorf("state %q must explicitly report zero unexpected speech", expected.Name)
		}
	}
	return nil
}

func scrollRegionATMeaningful(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed != value || len(trimmed) < 3 {
		return false
	}
	lower := strings.ToLower(trimmed)
	for _, prohibited := range []string{"replace", "todo", "tbd", "unknown", "n/a", "generic", "example"} {
		if strings.Contains(lower, prohibited) {
			return false
		}
	}
	return true
}

func validateScrollRegionATTraceEvents(events []scrollRegionATTraceEvent, observations []scrollRegionATObservation) error {
	if len(events) < len(scrollRegionATExpectedStates) {
		return fmt.Errorf("trace log must include at least one event for every state")
	}
	byState := make(map[string][]scrollRegionATTraceEvent, len(events))
	for _, event := range events {
		byState[event.State] = append(byState[event.State], event)
	}
	for _, observation := range observations {
		eventsForState := byState[observation.State]
		if len(eventsForState) == 0 {
			return fmt.Errorf("trace log omits state %q", observation.State)
		}
		matched := false
		for _, event := range eventsForState {
			if event.AfterBoundary == observation.Boundary && event.Focused == observation.Focused && slices.Contains(observation.Commands, event.Command) &&
				scrollRegionATMeaningful(event.Command) && scrollRegionATMeaningful(event.BeforeBoundary) {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("trace log does not prove command, focus, and boundary outcome for state %q", observation.State)
		}
	}
	return nil
}

func readScrollRegionATEvidenceArtifact(artifact scrollRegionATEvidenceArtifact, role string, seenPaths, seenHashes map[string]string) ([]byte, error) {
	if artifact.Path == "" || !scrollRegionSHA256Pattern.MatchString(artifact.SHA256) {
		return nil, fmt.Errorf("%s artifact path or SHA-256 is invalid", role)
	}
	resolved, err := filepath.Abs(artifact.Path)
	if err != nil {
		return nil, fmt.Errorf("resolve %s artifact: %w", role, err)
	}
	info, err := os.Lstat(resolved)
	if err != nil {
		return nil, fmt.Errorf("stat %s artifact: %w", role, err)
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("%s artifact must be a regular non-symlink file", role)
	}
	if prior, reused := seenPaths[resolved]; reused {
		return nil, fmt.Errorf("reused artifact path %q for %s and %s", resolved, prior, role)
	}
	content, err := os.ReadFile(resolved)
	if err != nil {
		return nil, fmt.Errorf("read %s artifact: %w", role, err)
	}
	actual := scrollRegionBFullSHA256(content)
	if artifact.SHA256 != actual {
		return nil, fmt.Errorf("%s artifact SHA-256 mismatch", role)
	}
	if prior, reused := seenHashes[actual]; reused {
		return nil, fmt.Errorf("reused artifact bytes %s for %s and %s", actual, prior, role)
	}
	seenPaths[resolved] = role
	seenHashes[actual] = role
	return content, nil
}

func validateScrollRegionATScreenshot(content []byte) error {
	decoded, format, err := image.Decode(bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("screenshot is not decodable image data: %w", err)
	}
	bounds := decoded.Bounds()
	if format != "png" || bounds.Dx() < 100 || bounds.Dy() < 80 {
		return fmt.Errorf("screenshot must be PNG at least 100x80, got %s %dx%d", format, bounds.Dx(), bounds.Dy())
	}
	// A solid or nearly solid claimant-generated image has no visual structure
	// from which a reviewer could identify browser/AT state. Sample across the
	// complete decoded capture before evaluating self-authored observations.
	colors := make(map[[4]uint32]struct{})
	for y := bounds.Min.Y; y < bounds.Max.Y; y += max(1, bounds.Dy()/12) {
		for x := bounds.Min.X; x < bounds.Max.X; x += max(1, bounds.Dx()/12) {
			red, green, blue, alpha := decoded.At(x, y).RGBA()
			colors[[4]uint32{red, green, blue, alpha}] = struct{}{}
		}
	}
	if len(colors) < 4 {
		return fmt.Errorf("screenshot lacks visual structure required for AT provenance")
	}
	return nil
}
