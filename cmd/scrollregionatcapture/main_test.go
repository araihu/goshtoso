package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCapturePlanRejectsUnsupportedPlatform(t *testing.T) {
	_, err := capturePlanFor("linux", "macos-safari-voiceover")
	if err == nil || !strings.Contains(err.Error(), "only supported on macOS") {
		t.Fatalf("capture plan must fail closed outside macOS, got %v", err)
	}
}

func TestCapturePlanRejectsUnknownPair(t *testing.T) {
	_, err := capturePlanFor("darwin", "chromium-supported-screen-reader")
	if err == nil || !strings.Contains(err.Error(), "unsupported AT pair") {
		t.Fatalf("capture plan must reject an unpinned AT pair, got %v", err)
	}
}

func TestCapturePlanUsesOnlySupportedDirectSystemCommands(t *testing.T) {
	plan, err := capturePlanFor("darwin", "macos-chromium-voiceover")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"go", "sw_vers", "mdls", "pgrep", "open", "osascript", "screencapture", "log"}
	if strings.Join(plan.RequiredCommands, ",") != strings.Join(want, ",") {
		t.Fatalf("direct capture command contract mismatch: got %v, want %v", plan.RequiredCommands, want)
	}
}

func TestCapturePlannerOwnsExactNonDefaultKeyboardSequences(t *testing.T) {
	plan, err := capturePlanFor("darwin", "macos-safari-voiceover")
	require.NoError(t, err)
	identity := candidateIdentity{CandidateTree: strings.Repeat("a", 40), ManifestSHA256: strings.Repeat("b", 64)}
	challenge := captureChallenge{Challenge: strings.Repeat("c", 64)}

	for _, want := range []struct {
		name    string
		setup   string
		keyCode int
	}{
		{name: "start", setup: "middle", keyCode: 115},
		{name: "middle", setup: "start", keyCode: 121},
		{name: "end", setup: "middle", keyCode: 119},
	} {
		var state captureState
		for _, candidate := range captureStates {
			if candidate.Name == want.name {
				state = candidate
				break
			}
		}
		require.NotEmpty(t, state.Name)
		require.Equal(t, want.keyCode, state.KeyCode)
		prepare := browserPrepareScript(plan, state, identity, challenge)
		require.Contains(t, prepare, `\"setup\":\"`+want.setup+`\"`)
		require.Contains(t, browserKeyboardScript(plan, state.KeyCode), fmt.Sprintf("key code %d", want.keyCode))
		require.Equal(t, "macOS System Events key code "+fmt.Sprint(want.keyCode), state.ActionEvent)
	}
}

// TestDirectCaptureRequiresActionBoundVoiceOverAndObservedSnapshots is a
// source-level TDD guard for the production adapter boundary. A signed
// cumulative log window or a DOM-derived focus-neighbour claim is not an AT
// observation and must never return after refactoring.
func TestDirectCaptureRequiresActionBoundVoiceOverAndObservedSnapshots(t *testing.T) {
	repository, err := filepath.Abs("../..")
	require.NoError(t, err)
	directCapture, err := os.ReadFile(filepath.Join(repository, "cmd", "scrollregionatcapture", "direct_capture.go"))
	require.NoError(t, err)
	browserReader, err := os.ReadFile(filepath.Join(repository, "cmd", "scrollregionatcapture", "browser_read_state.js"))
	require.NoError(t, err)

	require.NotContains(t, string(directCapture), `"--last", "20s"`, "cumulative VoiceOver windows permit stale speech reuse")
	require.NotContains(t, string(directCapture), `"--last", "2m"`, "capture must not add a cumulative system-log fallback")
	for _, want := range []string{`"--start"`, `"--end"`, `processID ==`, `com.apple.VoiceOver`, "action token"} {
		require.Contains(t, string(directCapture), want)
	}
	for _, want := range []string{"BeforeAt", "ActionIssuedAt", "AfterAt", "ExitAt", "LogStartedAt", "LogEndedAt", "validateCaptureStateTransition", "parseVoiceOverLogEvents"} {
		require.Contains(t, string(directCapture), want, "direct adapter must retain action-bound timestamp and raw-log validation")
	}
	require.NotContains(t, string(directCapture), "\n\tStartedAt ", "v2 action schema must not retain broad claimant interval fields")
	require.NotContains(t, string(directCapture), "\n\tEndedAt ", "v2 action schema must not retain broad claimant interval fields")
	for _, want := range []string{"before", "after", "exit", "document.activeElement", "scrollTop"} {
		require.Contains(t, string(browserReader), want)
	}
	require.Contains(t, string(browserReader), `meta[name="goshtoso-t-gs-011-at-action-token"]`, "browser state must bind the action token emitted by the candidate server")
	require.NotContains(t, string(browserReader), `#goshtoso-t-gs-011-at-action-token`, "candidate HTML emits token metadata, not a synthetic token DOM node")
	require.NotContains(t, string(browserReader), "focusable = Array.from", "focus-navigation evidence must record actual traversal, not DOM adjacency")
}

func TestCaptureCLIRejectsClaimantArtifactPathFlags(t *testing.T) {
	var stderr bytes.Buffer
	exitCode := run(
		[]string{
			"--pair", "macos-safari-voiceover",
			"--identity", "/tmp/identity.json",
			"--challenge", "/tmp/challenge.json",
			"--signing-key", "/tmp/capture-authority.pem",
			"--output-dir", "/tmp/capture",
			"--screenshot", "/tmp/claimant.png",
		},
		io.Discard,
		&stderr,
		captureRuntime{GOOS: "darwin"},
	)
	if exitCode != 2 || !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Fatalf("capture CLI must reject raw artifact input flags: exit=%d stderr=%q", exitCode, stderr.String())
	}
}

func TestCaptureCLIRejectsExternalURLOverride(t *testing.T) {
	var stderr bytes.Buffer
	exitCode := run([]string{"--url", "https://evidence.example/components/scroll-region"}, io.Discard, &stderr, captureRuntime{GOOS: "darwin"})
	if exitCode != 2 || !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Fatalf("capture adapter accepted a caller-selected external URL: exit=%d stderr=%q", exitCode, stderr.String())
	}
}

func TestCapturedScreenshotBindsNamedRegionToWindowPixels(t *testing.T) {
	path := filepath.Join(t.TempDir(), "capture.png")
	canvas := image.NewRGBA(image.Rect(0, 0, 200, 120))
	for y := range 120 {
		for x := range 200 {
			canvas.SetRGBA(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 128, A: 255})
		}
	}
	file, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, png.Encode(file, canvas))
	require.NoError(t, file.Close())

	window := browserWindowCapture{Window: browserRectangle{Width: 100, Height: 60}, CandidateRegion: browserRectangle{X: 20, Y: 10, Width: 40, Height: 30}}
	require.NoError(t, validateCapturedScreenshot(path, window))
	window.CandidateRegion.X = 90
	require.Error(t, validateCapturedScreenshot(path, window), "a claimed named region outside the OS window screenshot must be rejected")
}

func TestVerifyCandidateIdentityRejectsDirtyByteMutationWithSameStatus(t *testing.T) {
	base := t.TempDir()
	repository, identityPath, _, _ := newAdapterBridgeRepository(t, base)
	identity := readAdapterBridgeIdentity(t, identityPath)
	if err := os.WriteFile(filepath.Join(repository, "evidence.txt"), []byte("tampered but still untracked\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := verifyCandidateIdentity(repository, identity); err == nil {
		t.Fatal("candidate verifier accepted changed dirty bytes with unchanged porcelain status")
	}
}

func TestCaptureRejectsInjectedRuntimeAndCannedResults(t *testing.T) {
	err := executeCapture(captureConfig{
		Pair: "macos-safari-voiceover",
	}, io.Discard, captureRuntime{GOOS: "darwin"})
	if err == nil {
		t.Fatal("capture adapter accepted fake direct-command runtime and canned results")
	}
}

func TestCaptureAdapterBridgeCannotCreateSignedEvidence(t *testing.T) {
	err := executeCapture(captureConfig{
		Pair: "macos-safari-voiceover",
	}, io.Discard, captureRuntime{GOOS: "darwin"})
	require.ErrorContains(t, err, "injected command runtimes cannot attest")
}

func newAdapterBridgeRepository(t *testing.T, base string) (repository, identityPath, challengePath, privatePath string) {
	t.Helper()
	repository = filepath.Join(base, "repository")
	for path, content := range map[string]string{
		"go.mod":                "module example.test/scrollregion\n\ngo 1.26.5\n\nrequire github.com/a-h/templ v0.3.1020\n",
		"site/go.mod":           "module example.test/scrollregion/site\n\ngo 1.26.5\n\nrequire (\n\tgithub.com/a-h/templ v0.3.1020\n\tgithub.com/mxschmitt/playwright-go v0.6100.0\n)\n",
		"scripts/axe-core.lock": "version=4.10.3\narchive_sha256=0f2b4d7dcdf7d1219df8d1959ad68e565f51d14c3f0d88bb71cd59abeb956292\nscript_sha256=880970c081707360e64f34cea25ff91892f5bc95675b0776925b9709dd8a68bb\n",
		"tracked.txt":           "base\n",
	} {
		full := filepath.Join(repository, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(content), 0o600))
	}
	macPublic, macPrivate, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	windowsPublic, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	writeAdapterBridgePublicKey(t, repository, "bridge-macos-voiceover", macPublic)
	writeAdapterBridgePublicKey(t, repository, "bridge-windows-nvda", windowsPublic)
	manifest := trustedKeyManifest{Schema: captureTrustedKeysSchema, Keys: []trustedKeyRef{
		adapterBridgeKeyReference(t, repository, "bridge-macos-voiceover", []string{"macos-safari-voiceover", "macos-chromium-voiceover"}),
		adapterBridgeKeyReference(t, repository, "bridge-windows-nvda", []string{"windows-chromium-nvda"}),
	}}
	require.NoError(t, writeJSONExclusive(filepath.Join(repository, "tests", "external", "scrollregion-a11y", "attestation-keys.json"), manifest))
	privateDER, err := x509.MarshalPKCS8PrivateKey(macPrivate)
	require.NoError(t, err)
	privatePath = filepath.Join(t.TempDir(), "bridge-capture-authority.pem")
	require.NoError(t, os.WriteFile(privatePath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER}), 0o600))

	adapterBridgeGit(t, repository, "init", "-q")
	adapterBridgeGit(t, repository, "config", "user.email", "bridge@example.test")
	adapterBridgeGit(t, repository, "config", "user.name", "Adapter Bridge")
	adapterBridgeGit(t, repository, "remote", "add", "origin", "https://github.com/araihu/goshtoso.git")
	adapterBridgeGit(t, repository, "add", ".")
	adapterBridgeGit(t, repository, "commit", "-qm", "base")
	require.NoError(t, os.WriteFile(filepath.Join(repository, "evidence.txt"), []byte("candidate\n"), 0o600))

	identity := adapterBridgeCandidateIdentity(t, repository)
	identityPath = filepath.Join(base, "identity.json")
	require.NoError(t, writeJSONExclusive(identityPath, identity))
	randomChallenge := make([]byte, 32)
	_, err = rand.Read(randomChallenge)
	require.NoError(t, err)
	challenge := captureChallenge{Schema: captureChallengeSchema, Challenge: fmt.Sprintf("%x", randomChallenge), IssuedAt: time.Date(2026, time.August, 12, 17, 0, 0, 0, time.UTC).Format(time.RFC3339)}
	challengePath = filepath.Join(base, "challenge.json")
	require.NoError(t, writeJSONExclusive(challengePath, challenge))
	return repository, identityPath, challengePath, privatePath
}

func writeAdapterBridgePublicKey(t *testing.T, repository, keyID string, public ed25519.PublicKey) {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(public)
	require.NoError(t, err)
	path := filepath.Join(repository, "tests", "external", "scrollregion-a11y", "attestation-keys", keyID+".pem")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}), 0o600))
}

func adapterBridgeKeyReference(t *testing.T, repository, keyID string, pairs []string) trustedKeyRef {
	t.Helper()
	relative := filepath.ToSlash(filepath.Join("tests", "external", "scrollregion-a11y", "attestation-keys", keyID+".pem"))
	content, err := os.ReadFile(filepath.Join(repository, filepath.FromSlash(relative)))
	require.NoError(t, err)
	return trustedKeyRef{KeyID: keyID, PublicKeyPath: relative, PublicKeySHA256: sha256Hex(content), Pairs: pairs}
}

func adapterBridgeCandidateIdentity(t *testing.T, repository string) candidateIdentity {
	t.Helper()
	head := strings.TrimSpace(adapterBridgeGit(t, repository, "rev-parse", "HEAD"))
	tree := strings.TrimSpace(adapterBridgeGit(t, repository, "rev-parse", "HEAD^{tree}"))
	status := adapterBridgeGit(t, repository, "status", "--porcelain=v1", "--untracked-files=all")
	evidence, err := os.ReadFile(filepath.Join(repository, "evidence.txt"))
	require.NoError(t, err)
	paths := []candidatePath{{Path: "evidence.txt", SHA256: sha256Hex(evidence)}}
	manifest := "evidence.txt\t" + paths[0].SHA256 + "\n"
	return candidateIdentity{
		Schema:         "goshtoso.t-gs-011.candidate-identity.v2",
		RepositoryURL:  "https://github.com/araihu/goshtoso.git",
		Head:           head,
		Tree:           tree,
		CandidateTree:  adapterBridgeCandidateTree(t, repository),
		ManifestSHA256: sha256Hex([]byte(manifest)),
		StatusSHA256:   sha256Hex([]byte(status)),
		Paths:          paths,
		DependencyPins: dependencyPins{
			RootGoDirective: "1.26.5", SiteGoDirective: "1.26.5", Templ: "v0.3.1020", PlaywrightGo: "v0.6100.0",
			AxeCore: "4.10.3", AxeArchiveSHA256: "0f2b4d7dcdf7d1219df8d1959ad68e565f51d14c3f0d88bb71cd59abeb956292", AxeScriptSHA256: "880970c081707360e64f34cea25ff91892f5bc95675b0776925b9709dd8a68bb",
		},
	}
}

func adapterBridgeCandidateTree(t *testing.T, repository string) string {
	t.Helper()
	index := filepath.Join(t.TempDir(), "candidate.index")
	command := func(arguments ...string) string {
		cmd := exec.Command("git", arguments...)
		cmd.Dir = repository
		cmd.Env = append(os.Environ(), "GIT_INDEX_FILE="+index)
		output, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "git %s: %s", strings.Join(arguments, " "), output)
		return string(output)
	}
	command("read-tree", "HEAD")
	command("add", "-A")
	return strings.TrimSpace(command("write-tree"))
}

func adapterBridgeGit(t *testing.T, directory string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", arguments...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	require.NoErrorf(t, err, "git %s: %s", strings.Join(arguments, " "), output)
	return string(output)
}

func readAdapterBridgeIdentity(t *testing.T, path string) candidateIdentity {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	var identity candidateIdentity
	require.NoError(t, json.Unmarshal(content, &identity))
	return identity
}
