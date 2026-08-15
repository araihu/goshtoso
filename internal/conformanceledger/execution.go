package conformanceledger

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/araihu/goshtoso/internal/atattestation"
)

type ExecutionReceipt struct {
	ID             string                `json:"id"`
	Path           string                `json:"path"`
	ExpectedSHA256 string                `json:"expected_sha256"`
	Status         ReceiptStatus         `json:"status"`
	RowIDs         []string              `json:"row_ids"`
	Rationale      string                `json:"rationale,omitempty"`
	ToolVersions   map[string]string     `json:"tool_versions,omitempty"`
	Artifacts      []EvidenceArtifact    `json:"artifacts,omitempty"`
	Reproduction   ExecutionReproduction `json:"reproduction"`
	Context        ExecutionContext      `json:"context"`
}

// ExecutionContext binds the evidence wrapper to its repository, resolved
// dependency graph, and browser/AT context. A receipt cannot silently borrow
// a result produced against a different checkout or route.
type ExecutionContext struct {
	RepoRoot       string            `json:"repo_root"`
	SourceCommit   string            `json:"source_commit"`
	SourceTree     string            `json:"source_tree"`
	Identity       BFullIdentity     `json:"identity"`
	DependencyPins map[string]string `json:"dependency_pins"`
	Route          string            `json:"route,omitempty"`
	State          string            `json:"state,omitempty"`
	Viewport       int               `json:"viewport,omitempty"`
	Breakpoint     string            `json:"breakpoint,omitempty"`
	AT             string            `json:"at,omitempty"`
}

// ExecutionReceiptEnvelope is the authenticated wrapper consumed by the CLI.
// Its exact serialized bytes are hash-pinned outside the file using
// -executions-sha256, so changing status, row IDs, tool versions, or artifact
// metadata invalidates the invocation.
type ExecutionReceiptEnvelope struct {
	SchemaVersion int                `json:"schema_version"`
	SourceCommit  string             `json:"source_commit"`
	SourceTree    string             `json:"source_tree"`
	Receipts      []ExecutionReceipt `json:"receipts"`
}

type EvidenceArtifact struct {
	Kind           string `json:"kind"`
	Path           string `json:"path"`
	ExpectedSHA256 string `json:"expected_sha256"`
	Route          string `json:"route,omitempty"`
	State          string `json:"state,omitempty"`
	Browser        string `json:"browser,omitempty"`
	ATVersion      string `json:"at_version,omitempty"`
}

func ApplyExecutionReceipts(ledger *Ledger, receipts []ExecutionReceipt) error {
	if ledger == nil {
		return fmt.Errorf("nil ledger")
	}
	rowIndex := make(map[string]int, len(ledger.Rows))
	for index, row := range ledger.Rows {
		if _, duplicate := rowIndex[row.ID]; duplicate {
			return fmt.Errorf("duplicate ledger row id %s", row.ID)
		}
		rowIndex[row.ID] = index
	}
	receiptIDs := map[string]struct{}{}
	for _, receipt := range receipts {
		if receipt.ID == "" || receipt.Path == "" || receipt.ExpectedSHA256 == "" || len(receipt.RowIDs) == 0 {
			return fmt.Errorf("execution receipt id, path, expected SHA-256, and row ids are required")
		}
		if receipt.Status != StatusExecuted && receipt.Status != StatusFailed && receipt.Status != StatusBlocked && receipt.Status != StatusNotApplicable {
			return fmt.Errorf("execution receipt %s has invalid status %q", receipt.ID, receipt.Status)
		}
		if _, duplicate := receiptIDs[receipt.ID]; duplicate {
			return fmt.Errorf("duplicate execution receipt id %s", receipt.ID)
		}
		receiptIDs[receipt.ID] = struct{}{}
		if receipt.Status == StatusNotApplicable && strings.TrimSpace(receipt.Rationale) == "" {
			return fmt.Errorf("execution receipt %s N/A requires rationale", receipt.ID)
		}
		if strings.TrimSpace(receipt.Reproduction.Producer) == "" || strings.TrimSpace(receipt.Reproduction.RunID) == "" || strings.TrimSpace(receipt.Reproduction.Recorder) == "" {
			return fmt.Errorf("execution receipt %s reproduction producer, run id, and recorder are required", receipt.ID)
		}
		if err := validateExecutionContext(*ledger, receipt); err != nil {
			return fmt.Errorf("execution receipt %s: %w", receipt.ID, err)
		}
		content, err := os.ReadFile(receipt.Path)
		if err != nil {
			return fmt.Errorf("read execution receipt %s: %w", receipt.ID, err)
		}
		digest := sha256.Sum256(content)
		got := hex.EncodeToString(digest[:])
		if got != strings.ToLower(receipt.ExpectedSHA256) {
			return fmt.Errorf("execution receipt %s SHA-256 = %s, want %s", receipt.ID, got, receipt.ExpectedSHA256)
		}
		artifactHashes := map[string]string{}
		for _, artifact := range receipt.Artifacts {
			if artifact.Kind == "" || artifact.Path == "" || artifact.ExpectedSHA256 == "" {
				return fmt.Errorf("execution receipt %s artifact kind, path, and expected SHA-256 are required", receipt.ID)
			}
			if strings.HasPrefix(artifact.Kind, "at-") && (artifact.Route == "" || artifact.State == "" || artifact.Browser == "" || artifact.ATVersion == "") {
				return fmt.Errorf("execution receipt %s AT artifact %s requires route, state, browser, and AT version", receipt.ID, artifact.Kind)
			}
			artifactContent, err := os.ReadFile(artifact.Path)
			if err != nil {
				return fmt.Errorf("read execution receipt %s artifact %s: %w", receipt.ID, artifact.Kind, err)
			}
			artifactDigest := sha256.Sum256(artifactContent)
			artifactGot := hex.EncodeToString(artifactDigest[:])
			if artifactGot != strings.ToLower(artifact.ExpectedSHA256) {
				return fmt.Errorf("execution receipt %s artifact %s SHA-256 = %s, want %s", receipt.ID, artifact.Kind, artifactGot, artifact.ExpectedSHA256)
			}
			key := artifact.Kind + ":" + artifact.Path
			if _, duplicate := artifactHashes[key]; duplicate {
				return fmt.Errorf("execution receipt %s duplicate artifact %s", receipt.ID, key)
			}
			artifactHashes[key] = artifactGot
			if artifact.Kind == "b-full-state-matrix" {
				if err := ReadAndValidateBFullManifest(artifact.Path, *ledger); err != nil {
					return fmt.Errorf("execution receipt %s: %w", receipt.ID, err)
				}
			}
			if err := validateEvidenceArtifactFormat(artifact, artifactContent); err != nil {
				return fmt.Errorf("execution receipt %s artifact %s: %w", receipt.ID, artifact.Kind, err)
			}
		}
		if err := validateBFullStateReceipt(*ledger, rowIndex, receipt); err != nil {
			return err
		}
		if err := validateATExecutionReceipt(*ledger, rowIndex, receipt); err != nil {
			return err
		}
		artifactDigests := make([]string, 0, len(artifactHashes))
		for _, artifactDigest := range artifactHashes {
			artifactDigests = append(artifactDigests, artifactDigest)
		}
		sort.Strings(artifactDigests)
		for _, rowID := range receipt.RowIDs {
			index, ok := rowIndex[rowID]
			if !ok {
				return fmt.Errorf("execution receipt %s references unknown row %s", receipt.ID, rowID)
			}
			if ledger.Rows[index].Class != ClassExecution {
				return fmt.Errorf("execution receipt %s references non-execution row %s", receipt.ID, rowID)
			}
			if receipt.Status == StatusNotApplicable && mandatoryExecutionRow(ledger.Rows[index]) {
				return fmt.Errorf("execution receipt %s cannot mark mandatory execution row %s not applicable", receipt.ID, rowID)
			}
			row := &ledger.Rows[index]
			for _, prior := range row.Reproductions {
				if prior.Producer == receipt.Reproduction.Producer {
					return fmt.Errorf("execution row %s duplicate reproduction producer %s", rowID, prior.Producer)
				}
				if prior.RunID == receipt.Reproduction.RunID {
					return fmt.Errorf("execution row %s duplicate reproduction run id %s", rowID, prior.RunID)
				}
				if prior.Recorder == receipt.Reproduction.Recorder {
					return fmt.Errorf("execution row %s duplicate reproduction recorder %s", rowID, prior.Recorder)
				}
				for _, priorArtifact := range prior.ArtifactSHA256s {
					if slices.Contains(artifactDigests, priorArtifact) {
						return fmt.Errorf("execution row %s aliases reproduction artifact SHA-256", rowID)
					}
				}
			}
			row.ReceiptStatus = receipt.Status
			if receipt.Status == StatusNotApplicable {
				row.Applicability = NotApplicable
			} else {
				row.Applicability = Applicable
			}
			row.Receipt = receipt.Path
			row.Rationale = receipt.Rationale
			if row.EvidenceHashes == nil {
				row.EvidenceHashes = map[string]string{}
			}
			row.EvidenceHashes["receipt_sha256:"+receipt.Reproduction.RunID] = got
			for key, digest := range artifactHashes {
				row.EvidenceHashes["artifact:"+receipt.Reproduction.RunID+":"+key] = digest
			}
			row.Reproductions = append(row.Reproductions, ExecutionReproduction{
				Producer:        receipt.Reproduction.Producer,
				RunID:           receipt.Reproduction.RunID,
				Recorder:        receipt.Reproduction.Recorder,
				ReceiptSHA256:   got,
				ArtifactSHA256s: artifactDigests,
			})
		}
		ledger.Metadata["execution."+receipt.ID+".path"] = receipt.Path
		ledger.Metadata["execution."+receipt.ID+".sha256"] = got
		keys := make([]string, 0, len(receipt.ToolVersions))
		for name := range receipt.ToolVersions {
			keys = append(keys, name)
		}
		sort.Strings(keys)
		for _, name := range keys {
			ledger.Metadata["execution."+receipt.ID+".tool."+name] = receipt.ToolVersions[name]
		}
	}
	return nil
}

func ReadAndApplyExecutionReceiptEnvelope(ledger *Ledger, path, expectedSHA256 string) error {
	if ledger == nil {
		return fmt.Errorf("nil ledger")
	}
	if !validSHA256(expectedSHA256) {
		return fmt.Errorf("execution receipt envelope expected SHA-256 is required")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read execution receipt envelope: %w", err)
	}
	digest := sha256.Sum256(content)
	if got := hex.EncodeToString(digest[:]); got != strings.ToLower(expectedSHA256) {
		return fmt.Errorf("execution receipt envelope SHA-256 = %s, want %s", got, expectedSHA256)
	}
	var envelope ExecutionReceiptEnvelope
	if err := json.Unmarshal(content, &envelope); err != nil {
		return fmt.Errorf("parse execution receipt envelope: %w", err)
	}
	if envelope.SchemaVersion != SchemaVersion || envelope.SourceCommit != ledger.SourceCommit || envelope.SourceTree != ledger.SourceTree {
		return fmt.Errorf("execution receipt envelope source identity mismatch")
	}
	return ApplyExecutionReceipts(ledger, envelope.Receipts)
}

func validateExecutionContext(ledger Ledger, receipt ExecutionReceipt) error {
	context := receipt.Context
	if strings.TrimSpace(context.RepoRoot) == "" || context.SourceCommit != ledger.SourceCommit || context.SourceTree != ledger.SourceTree || len(context.DependencyPins) == 0 {
		return fmt.Errorf("repo root, source commit/tree, and dependency pins are required")
	}
	repoRoot, err := filepath.Abs(context.RepoRoot)
	if err != nil {
		return fmt.Errorf("resolve execution repository root: %w", err)
	}
	identityRoot, err := filepath.Abs(context.Identity.RepoRoot)
	if err != nil {
		return fmt.Errorf("resolve execution identity repository root: %w", err)
	}
	if repoRoot != identityRoot {
		return fmt.Errorf("execution repo root does not match authenticated identity root")
	}
	if err := VerifyBFullIdentity(context.Identity, ledger.SourceCommit, ledger.SourceTree); err != nil {
		return fmt.Errorf("authenticate execution repository identity: %w", err)
	}
	expectedPins, err := deriveExecutionDependencyPins(repoRoot, ledger.SourceCommit)
	if err != nil {
		return err
	}
	if !equalExecutionDependencyPins(context.DependencyPins, expectedPins) {
		return fmt.Errorf("execution dependency pins mismatch")
	}
	if context.AT != "" && context.AT != "safari-voiceover" && context.AT != "chromium-screen-reader" {
		return fmt.Errorf("unknown execution AT context %q", context.AT)
	}
	return nil
}

func deriveExecutionDependencyPins(repoRoot, sourceCommit string) (map[string]string, error) {
	rootGo, rootRequire, err := readExecutionGoMod(filepath.Join(repoRoot, "go.mod"))
	if err != nil {
		return nil, err
	}
	siteGo, siteRequire, err := readExecutionGoMod(filepath.Join(repoRoot, "site", "go.mod"))
	if err != nil {
		return nil, err
	}
	templ := rootRequire["github.com/a-h/templ"]
	if templ == "" || siteRequire["github.com/a-h/templ"] != templ {
		return nil, fmt.Errorf("execution dependency pins require matching root/site templ version")
	}
	playwright := siteRequire["github.com/mxschmitt/playwright-go"]
	if playwright == "" {
		return nil, fmt.Errorf("execution dependency pins require site playwright-go version")
	}
	lock, err := readExecutionAxeLock(filepath.Join(repoRoot, "scripts", "axe-core.lock"))
	if err != nil {
		return nil, err
	}
	pins := map[string]string{
		"github.com/araihu/goshtoso":         sourceCommit,
		"go.root":                            rootGo,
		"go.site":                            siteGo,
		"github.com/a-h/templ":               templ,
		"github.com/mxschmitt/playwright-go": playwright,
		"axe-core":                           lock["version"],
		"axe-core.archive_sha256":            lock["archive_sha256"],
		"axe-core.script_sha256":             lock["script_sha256"],
	}
	for name, value := range pins {
		if strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("execution dependency pin %s is missing", name)
		}
	}
	return pins, nil
}

func readExecutionGoMod(path string) (string, map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", nil, fmt.Errorf("read execution dependency %s: %w", path, err)
	}
	goDirective := ""
	requires := map[string]string{}
	inRequire := false
	for raw := range strings.SplitSeq(string(content), "\n") {
		line := strings.TrimSpace(strings.SplitN(raw, "//", 2)[0])
		if line == "" {
			continue
		}
		if after, ok := strings.CutPrefix(line, "go "); ok {
			goDirective = strings.TrimSpace(after)
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
	if goDirective == "" {
		return "", nil, fmt.Errorf("execution dependency %s has no go directive", path)
	}
	return goDirective, requires, nil
}

func readExecutionAxeLock(path string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read execution axe lock: %w", err)
	}
	values := map[string]string{}
	for raw := range strings.SplitSeq(string(content), "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(raw), "=")
		if ok {
			values[key] = value
		}
	}
	return values, nil
}

func equalExecutionDependencyPins(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func validateBFullStateReceipt(ledger Ledger, rowIndex map[string]int, receipt ExecutionReceipt) error {
	stateRows := make(map[string]struct{})
	for _, row := range ledger.Rows {
		if row.Class == ClassExecution && strings.HasPrefix(row.ID, "execution/state/") {
			stateRows[row.ID] = struct{}{}
		}
	}
	mapped := make(map[string]struct{})
	for _, rowID := range receipt.RowIDs {
		index, ok := rowIndex[rowID]
		if !ok || !strings.HasPrefix(ledger.Rows[index].ID, "execution/state/") {
			continue
		}
		mapped[rowID] = struct{}{}
	}
	if len(mapped) == 0 || receipt.Status != StatusExecuted {
		return nil
	}
	manifestCount := 0
	for _, artifact := range receipt.Artifacts {
		if artifact.Kind == "b-full-state-matrix" {
			manifestCount++
		}
	}
	if manifestCount != 1 {
		return fmt.Errorf("execution receipt %s maps public states and requires exactly one b-full-state-matrix artifact", receipt.ID)
	}
	if len(mapped) != len(stateRows) {
		return fmt.Errorf("execution receipt %s B-FULL state mapping has %d rows, want all %d", receipt.ID, len(mapped), len(stateRows))
	}
	return nil
}

func validateATExecutionReceipt(ledger Ledger, rowIndex map[string]int, receipt ExecutionReceipt) error {
	var target *Row
	for _, rowID := range receipt.RowIDs {
		index, ok := rowIndex[rowID]
		if !ok {
			continue
		}
		row := &ledger.Rows[index]
		if row.AT == "" {
			continue
		}
		if target != nil {
			return fmt.Errorf("execution receipt %s must map exactly one AT row", receipt.ID)
		}
		target = row
	}
	if target == nil || receipt.Status != StatusExecuted {
		return nil
	}
	if len(receipt.RowIDs) != 1 {
		return fmt.Errorf("execution receipt %s must map exactly one AT row", receipt.ID)
	}
	if receipt.Context.AT != target.AT || receipt.Context.Route != target.Route || receipt.Context.State != target.State {
		return fmt.Errorf("execution receipt %s AT context %s/%s/%s does not match row %s/%s/%s", receipt.ID, receipt.Context.AT, receipt.Context.Route, receipt.Context.State, target.AT, target.Route, target.State)
	}
	browser := strings.TrimSpace(receipt.ToolVersions["browser"])
	screenReader := strings.TrimSpace(receipt.ToolVersions["screen_reader"])
	if browser == "" || screenReader == "" {
		return fmt.Errorf("execution receipt %s AT execution requires browser and screen_reader tool versions", receipt.ID)
	}
	if err := validateATToolIdentity(target.AT, browser, screenReader); err != nil {
		return fmt.Errorf("execution receipt %s: %w", receipt.ID, err)
	}
	for _, artifact := range receipt.Artifacts {
		if artifact.Route != target.Route || artifact.State != target.State {
			return fmt.Errorf("execution receipt %s artifact %s route/state %s/%s does not match row %s/%s", receipt.ID, artifact.Kind, artifact.Route, artifact.State, target.Route, target.State)
		}
		if artifact.Browser != browser {
			return fmt.Errorf("execution receipt %s artifact %s browser %q does not match tool browser %q", receipt.ID, artifact.Kind, artifact.Browser, browser)
		}
		if artifact.ATVersion != screenReader {
			return fmt.Errorf("execution receipt %s artifact %s AT version %q does not match tool screen reader %q", receipt.ID, artifact.Kind, artifact.ATVersion, screenReader)
		}
	}
	counts := map[string]int{}
	artifacts := map[string]EvidenceArtifact{}
	for _, artifact := range receipt.Artifacts {
		counts[artifact.Kind]++
		artifacts[artifact.Kind] = artifact
	}
	if len(receipt.Artifacts) != 4 || counts["at-caption-screenshot"] != 1 || counts["at-cursor-trace"] != 1 || counts["at-served-response"] != 1 || counts["at-signed-attestation"] != 1 {
		return fmt.Errorf("execution receipt %s requires exactly one at-caption-screenshot and one at-cursor-trace plus served response and signed attestation", receipt.ID)
	}
	trusted, err := loadConformanceATTrustedKeys(receipt.Context.RepoRoot)
	if err != nil {
		return err
	}
	attestation, err := os.ReadFile(artifacts["at-signed-attestation"].Path)
	if err != nil {
		return fmt.Errorf("read signed AT attestation: %w", err)
	}
	want := atattestation.Claims{SourceCommit: ledger.SourceCommit, SourceTree: ledger.SourceTree, Route: target.Route, State: target.State, Browser: browser, ScreenReader: screenReader, Pair: target.AT, Recorder: trustedATRecorderForPair(target.AT), ServedSHA256: artifacts["at-served-response"].ExpectedSHA256, ScreenshotSHA256: artifacts["at-caption-screenshot"].ExpectedSHA256, TraceSHA256: artifacts["at-cursor-trace"].ExpectedSHA256}
	if err := atattestation.Verify(attestation, trusted, want); err != nil {
		return fmt.Errorf("execution receipt %s signed AT attestation: %w", receipt.ID, err)
	}
	return nil
}

func trustedATRecorderForPair(pair string) string {
	switch pair {
	case "safari-voiceover":
		return "macos-voiceover-20260812"
	case "chromium-screen-reader":
		return "windows-nvda-20260812"
	default:
		return ""
	}
}

func loadConformanceATTrustedKeys(repoRoot string) (map[string]ed25519.PublicKey, error) {
	paths := map[string]string{"macos-voiceover-20260812": "tests/external/scrollregion-a11y/attestation-keys/macos-voiceover-public.pem", "windows-nvda-20260812": "tests/external/scrollregion-a11y/attestation-keys/windows-nvda-public.pem"}
	keys := make(map[string]ed25519.PublicKey, len(paths))
	for id, relative := range paths {
		raw, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(relative)))
		if err != nil {
			return nil, fmt.Errorf("read source-pinned AT recorder key %s: %w", id, err)
		}
		block, _ := pem.Decode(raw)
		if block == nil {
			return nil, fmt.Errorf("decode source-pinned AT recorder key %s", id)
		}
		parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse source-pinned AT recorder key %s: %w", id, err)
		}
		key, ok := parsed.(ed25519.PublicKey)
		if !ok {
			return nil, fmt.Errorf("source-pinned AT recorder key %s is not Ed25519", id)
		}
		keys[id] = key
	}
	return keys, nil
}

func validateATToolIdentity(at, browser, screenReader string) error {
	switch at {
	case "safari-voiceover":
		if !strings.HasPrefix(browser, "Safari ") || !strings.HasPrefix(screenReader, "VoiceOver ") {
			return fmt.Errorf("AT row %s requires Safari and VoiceOver identities", at)
		}
	case "chromium-screen-reader":
		if !(strings.HasPrefix(browser, "Google Chrome ") || strings.HasPrefix(browser, "Chromium ") || strings.HasPrefix(browser, "Chrome ")) {
			return fmt.Errorf("AT row %s requires a Chromium browser identity", at)
		}
	default:
		return fmt.Errorf("unknown AT row identity %q", at)
	}
	return nil
}

func validateEvidenceArtifactFormat(artifact EvidenceArtifact, content []byte) error {
	switch artifact.Kind {
	case "at-caption-screenshot":
		// A screenshot must be a real PNG. Plain text named .png is not AT
		// evidence and must not pass simply because its SHA-256 was recorded.
		if len(content) < 8 || !bytes.Equal(content[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}) {
			return fmt.Errorf("caption screenshot is not a PNG")
		}
		config, format, err := image.DecodeConfig(bytes.NewReader(content))
		if err != nil || format != "png" || config.Width < 2 || config.Height < 2 {
			return fmt.Errorf("caption screenshot is not a decodable non-empty PNG")
		}
	case "at-cursor-trace":
		var trace struct {
			Route        string `json:"route"`
			State        string `json:"state"`
			Browser      string `json:"browser"`
			ScreenReader string `json:"screen_reader"`
			Events       []struct {
				At   string `json:"at"`
				Type string `json:"type"`
				Role string `json:"role"`
				Name string `json:"name"`
			} `json:"events"`
		}
		if err := json.Unmarshal(content, &trace); err != nil {
			return fmt.Errorf("cursor trace is not JSON: %w", err)
		}
		if trace.Route != artifact.Route || trace.State != artifact.State || trace.Browser != artifact.Browser || trace.ScreenReader != artifact.ATVersion || len(trace.Events) == 0 {
			return fmt.Errorf("cursor trace context or events do not match artifact metadata")
		}
		for _, event := range trace.Events {
			if strings.TrimSpace(event.At) == "" || strings.TrimSpace(event.Type) == "" || (strings.TrimSpace(event.Role) == "" && strings.TrimSpace(event.Name) == "") {
				return fmt.Errorf("cursor trace requires non-empty typed events")
			}
		}
	}
	return nil
}
