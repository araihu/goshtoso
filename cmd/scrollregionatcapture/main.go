// Command scrollregionatcapture creates a signed capture bundle for the final
// T-GS-011 AT receipt. It never accepts pre-existing evidence artifact paths.
package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/araihu/goshtoso/internal/scrollregionidentity"
)

const (
	captureRoute                    = "/components/scroll-region"
	captureChallengeSchema          = "goshtoso.t-gs-011.at-challenge.v1"
	captureAttestationSchema        = "goshtoso.t-gs-011.at-dsse-envelope.v1"
	captureAttestationPayloadType   = "application/vnd.goshtoso.t-gs-011.at-capture.v1+json"
	captureAttestationPayloadSchema = "goshtoso.t-gs-011.at-capture.v1"
	captureTranscriptSchema         = "goshtoso.t-gs-011.at-transcript.v2"
	captureTraceSchema              = "goshtoso.t-gs-011.at-trace-log.v2"
	captureTrustedKeysSchema        = "goshtoso.t-gs-011.at-attestation-keys.v1"
	captureBundleSchema             = "goshtoso.t-gs-011.at-capture-bundle.v1"
)

var (
	captureSHA256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
	captureGitPattern    = regexp.MustCompile(`^[0-9a-f]{40}$`)
	captureVersion       = regexp.MustCompile(`^[1-9][0-9]*(?:\.[0-9]+){1,3}$`)
	captureKeyID         = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{2,80}$`)
)

type capturePlan struct {
	Pair             string
	BrowserApp       string
	BrowserName      string
	RequiredCommands []string
}

func capturePlanFor(goos, pair string) (capturePlan, error) {
	if goos != "darwin" {
		return capturePlan{}, fmt.Errorf("T-GS-011 AT capture is only supported on macOS; got %q", goos)
	}
	plan := capturePlan{Pair: pair, RequiredCommands: []string{"go", "sw_vers", "mdls", "pgrep", "open", "osascript", "screencapture", "log"}}
	switch pair {
	case "macos-safari-voiceover":
		plan.BrowserApp, plan.BrowserName = "Safari", "Safari"
	case "macos-chromium-voiceover":
		plan.BrowserApp, plan.BrowserName = "Chromium", "Chromium"
	default:
		return capturePlan{}, fmt.Errorf("unsupported AT pair %q", pair)
	}
	return plan, nil
}

type captureRuntime struct {
	GOOS             string
	Now              func() time.Time
	WorkingDirectory string
	RunCommand       func(context.Context, string, ...string) ([]byte, error)
	DirectCommands   bool
}

func (environment captureRuntime) defaults() captureRuntime {
	if environment.GOOS == "" {
		environment.GOOS = runtime.GOOS
	}
	if environment.Now == nil {
		environment.Now = time.Now
	}
	if environment.WorkingDirectory == "" {
		directory, err := os.Getwd()
		if err == nil {
			environment.WorkingDirectory = directory
		}
	}
	if environment.RunCommand == nil {
		environment.RunCommand = func(ctx context.Context, name string, args ...string) ([]byte, error) {
			command := exec.CommandContext(ctx, name, args...)
			command.Dir = environment.WorkingDirectory
			return command.CombinedOutput()
		}
		environment.DirectCommands = true
	}
	return environment
}

type captureConfig struct {
	Pair           string
	IdentityPath   string
	ChallengePath  string
	SigningKeyPath string
	OutputDir      string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, captureRuntime{}))
}

func run(args []string, stdout, stderr io.Writer, environment captureRuntime) int {
	flags := flag.NewFlagSet("scrollregionatcapture", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var config captureConfig
	flags.StringVar(&config.Pair, "pair", "", "macos-safari-voiceover or macos-chromium-voiceover")
	flags.StringVar(&config.IdentityPath, "identity", "", "independently frozen candidate identity sidecar")
	flags.StringVar(&config.ChallengePath, "challenge", "", "independently random AT capture challenge JSON")
	flags.StringVar(&config.SigningKeyPath, "signing-key", "", "pair-authorized Ed25519 PKCS#8 private key")
	flags.StringVar(&config.OutputDir, "output-dir", "", "new empty directory for directly captured artifacts")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "scrollregionatcapture: positional arguments are forbidden")
		return 2
	}
	for _, required := range []struct{ name, value string }{
		{"pair", config.Pair}, {"identity", config.IdentityPath},
		{"challenge", config.ChallengePath}, {"signing-key", config.SigningKeyPath}, {"output-dir", config.OutputDir},
	} {
		if strings.TrimSpace(required.value) == "" {
			fmt.Fprintf(stderr, "scrollregionatcapture: --%s is required\n", required.name)
			return 2
		}
	}
	if err := executeCapture(config, stdout, environment.defaults()); err != nil {
		fmt.Fprintf(stderr, "scrollregionatcapture: %v\n", err)
		return 1
	}
	return 0
}

type dependencyPins struct {
	RootGoDirective  string `json:"root_go_directive"`
	SiteGoDirective  string `json:"site_go_directive"`
	Templ            string `json:"templ"`
	PlaywrightGo     string `json:"playwright_go"`
	AxeCore          string `json:"axe_core"`
	AxeArchiveSHA256 string `json:"axe_archive_sha256"`
	AxeScriptSHA256  string `json:"axe_script_sha256"`
}

type candidatePath struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type candidateIdentity struct {
	Schema         string          `json:"schema"`
	RepositoryURL  string          `json:"repository_url"`
	Head           string          `json:"head"`
	Tree           string          `json:"tree"`
	CandidateTree  string          `json:"candidate_tree"`
	ManifestSHA256 string          `json:"manifest_sha256"`
	StatusSHA256   string          `json:"status_sha256"`
	Paths          []candidatePath `json:"paths"`
	DependencyPins dependencyPins  `json:"dependency_pins"`
}

type receiptIdentity struct {
	RepositoryURL  string         `json:"repository_url"`
	Head           string         `json:"head"`
	Tree           string         `json:"tree"`
	CandidateTree  string         `json:"candidate_tree"`
	ManifestSHA256 string         `json:"manifest_sha256"`
	StatusSHA256   string         `json:"status_sha256"`
	DependencyPins dependencyPins `json:"dependency_pins"`
}

func receiptIdentityFor(candidate candidateIdentity) receiptIdentity {
	return receiptIdentity{RepositoryURL: candidate.RepositoryURL, Head: candidate.Head, Tree: candidate.Tree, CandidateTree: candidate.CandidateTree, ManifestSHA256: candidate.ManifestSHA256, StatusSHA256: candidate.StatusSHA256, DependencyPins: candidate.DependencyPins}
}

type captureChallenge struct {
	Schema    string `json:"schema"`
	Challenge string `json:"challenge"`
	IssuedAt  string `json:"issued_at"`
}

type evidenceArtifact struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type focusNavigation struct {
	Before string `json:"before"`
	Entry  string `json:"entry"`
	Exit   string `json:"exit"`
}

type observation struct {
	State            string          `json:"state"`
	Role             string          `json:"role"`
	Name             string          `json:"name"`
	Focused          bool            `json:"focused"`
	Boundary         string          `json:"boundary"`
	Commands         []string        `json:"commands"`
	FocusNavigation  focusNavigation `json:"focus_navigation"`
	ObservedSpeech   []string        `json:"observed_speech"`
	UnexpectedSpeech []string        `json:"unexpected_speech"`
}

type traceEvent struct {
	State          string `json:"state"`
	Command        string `json:"command"`
	BeforeBoundary string `json:"before_boundary"`
	AfterBoundary  string `json:"after_boundary"`
	Focused        bool   `json:"focused"`
}

type transcript struct {
	Schema              string          `json:"schema"`
	Challenge           string          `json:"challenge"`
	Pair                string          `json:"pair"`
	CapturedAt          string          `json:"captured_at"`
	Identity            receiptIdentity `json:"identity"`
	PlatformVersion     string          `json:"platform_version"`
	BrowserVersion      string          `json:"browser_version"`
	ScreenReaderVersion string          `json:"screen_reader_version"`
	Route               string          `json:"route"`
	Observations        []observation   `json:"observations"`
}

type traceLog struct {
	Schema     string          `json:"schema"`
	Challenge  string          `json:"challenge"`
	Pair       string          `json:"pair"`
	CapturedAt string          `json:"captured_at"`
	Identity   receiptIdentity `json:"identity"`
	Route      string          `json:"route"`
	Events     []traceEvent    `json:"events"`
}

type evidenceCapture struct {
	Pair                string             `json:"pair"`
	PlatformVersion     string             `json:"platform_version"`
	BrowserVersion      string             `json:"browser_version"`
	ScreenReaderVersion string             `json:"screen_reader_version"`
	Route               string             `json:"route"`
	CapturedAt          string             `json:"captured_at"`
	Observations        []observation      `json:"observations"`
	ServedPage          evidenceArtifact   `json:"served_page"`
	ServedResponse      evidenceArtifact   `json:"served_response"`
	BrowserStates       []evidenceArtifact `json:"browser_states"`
	ActionRecords       []evidenceArtifact `json:"action_records"`
	VoiceOverCaptions   []evidenceArtifact `json:"voiceover_captions"`
	VoiceOverLog        evidenceArtifact   `json:"voiceover_log"`
	BrowserWindow       evidenceArtifact   `json:"browser_window"`
	Screenshot          evidenceArtifact   `json:"screenshot"`
	Transcript          evidenceArtifact   `json:"transcript"`
	TraceLog            evidenceArtifact   `json:"trace_log"`
	Attestation         evidenceArtifact   `json:"attestation"`
}

type attestedCapture struct {
	Pair                string             `json:"pair"`
	PlatformVersion     string             `json:"platform_version"`
	BrowserVersion      string             `json:"browser_version"`
	ScreenReaderVersion string             `json:"screen_reader_version"`
	Route               string             `json:"route"`
	CapturedAt          string             `json:"captured_at"`
	Observations        []observation      `json:"observations"`
	ServedPage          evidenceArtifact   `json:"served_page"`
	ServedResponse      evidenceArtifact   `json:"served_response"`
	BrowserStates       []evidenceArtifact `json:"browser_states"`
	ActionRecords       []evidenceArtifact `json:"action_records"`
	VoiceOverCaptions   []evidenceArtifact `json:"voiceover_captions"`
	VoiceOverLog        evidenceArtifact   `json:"voiceover_log"`
	BrowserWindow       evidenceArtifact   `json:"browser_window"`
	Screenshot          evidenceArtifact   `json:"screenshot"`
	Transcript          evidenceArtifact   `json:"transcript"`
	TraceLog            evidenceArtifact   `json:"trace_log"`
}

func attestCapture(capture evidenceCapture) attestedCapture {
	return attestedCapture{Pair: capture.Pair, PlatformVersion: capture.PlatformVersion, BrowserVersion: capture.BrowserVersion, ScreenReaderVersion: capture.ScreenReaderVersion, Route: capture.Route, CapturedAt: capture.CapturedAt, Observations: capture.Observations, ServedPage: capture.ServedPage, ServedResponse: capture.ServedResponse, BrowserStates: capture.BrowserStates, ActionRecords: capture.ActionRecords, VoiceOverCaptions: capture.VoiceOverCaptions, VoiceOverLog: capture.VoiceOverLog, BrowserWindow: capture.BrowserWindow, Screenshot: capture.Screenshot, Transcript: capture.Transcript, TraceLog: capture.TraceLog}
}

type attestationPayload struct {
	Schema    string          `json:"schema"`
	Challenge string          `json:"challenge"`
	Identity  receiptIdentity `json:"identity"`
	Capture   attestedCapture `json:"capture"`
}

type dsseSignature struct {
	KeyID     string `json:"keyid"`
	Signature string `json:"sig"`
}

type dsseEnvelope struct {
	Schema      string          `json:"schema"`
	PayloadType string          `json:"payloadType"`
	Payload     string          `json:"payload"`
	Signatures  []dsseSignature `json:"signatures"`
}

type trustedKeyManifest struct {
	Schema string          `json:"schema"`
	Keys   []trustedKeyRef `json:"keys"`
}

type trustedKeyRef struct {
	KeyID           string   `json:"key_id"`
	PublicKeyPath   string   `json:"public_key_path"`
	PublicKeySHA256 string   `json:"public_key_sha256"`
	Pairs           []string `json:"pairs"`
}

type trustedKey struct {
	ID     string
	Pairs  map[string]struct{}
	Public ed25519.PublicKey
}

type captureState struct {
	Name        string
	Boundary    string
	Focused     bool
	Command     string
	ActionEvent string
	KeyCode     int
}

var captureStates = []captureState{
	{Name: "default", Boundary: "start", Focused: false, Command: "Navigate", ActionEvent: "adapter-owned candidate navigation"},
	{Name: "no-overflow", Boundary: "no-overflow", Focused: true, Command: "Tab", ActionEvent: "macOS System Events key code 48", KeyCode: 48},
	{Name: "start", Boundary: "start", Focused: true, Command: "Home", ActionEvent: "macOS System Events key code 115", KeyCode: 115},
	{Name: "middle", Boundary: "middle", Focused: true, Command: "PageDown", ActionEvent: "macOS System Events key code 121", KeyCode: 121},
	{Name: "end", Boundary: "end", Focused: true, Command: "End", ActionEvent: "macOS System Events key code 119", KeyCode: 119},
	{Name: "focused", Boundary: "start", Focused: true, Command: "Tab", ActionEvent: "macOS System Events key code 48", KeyCode: 48},
}

type authorizedSigningKey struct {
	ID      string
	Private ed25519.PrivateKey
}

func executeCapture(config captureConfig, stdout io.Writer, environment captureRuntime) error {
	if !environment.DirectCommands {
		return fmt.Errorf("production capture requires direct command runtime; injected command runtimes cannot attest AT evidence")
	}
	plan, err := capturePlanFor(environment.GOOS, config.Pair)
	if err != nil {
		return err
	}
	repositoryRoot, err := repositoryRoot(environment)
	if err != nil {
		return err
	}
	identity, err := readCandidateIdentity(config.IdentityPath)
	if err != nil {
		return err
	}
	if err := verifyCandidateIdentity(repositoryRoot, identity); err != nil {
		return err
	}
	challenge, err := readChallenge(config.ChallengePath)
	if err != nil {
		return err
	}
	signingKey, err := readAuthorizedSigningKey(repositoryRoot, plan.Pair, config.SigningKeyPath)
	if err != nil {
		return err
	}
	directory, err := prepareOutputDirectory(config.OutputDir)
	if err != nil {
		return err
	}
	server, err := startVerifiedCandidateServer(repositoryRoot, identity, challenge, plan.Pair)
	if err != nil {
		return err
	}
	defer func() { _ = server.close() }()
	served, servedHTML, err := server.fetch(identity, challenge)
	if err != nil {
		return err
	}
	servedPagePath := filepath.Join(directory, "served-page.html")
	if err := writeBytesExclusive(servedPagePath, servedHTML); err != nil {
		return fmt.Errorf("write directly served candidate HTML: %w", err)
	}
	servedPageArtifact, err := artifactForFile(servedPagePath)
	if err != nil {
		return err
	}
	servedResponsePath := filepath.Join(directory, "served-response.json")
	if err := writeJSONExclusive(servedResponsePath, served); err != nil {
		return fmt.Errorf("write directly served candidate response binding: %w", err)
	}
	servedResponseArtifact, err := artifactForFile(servedResponsePath)
	if err != nil {
		return err
	}
	if err := verifyCandidateIdentity(repositoryRoot, identity); err != nil {
		return fmt.Errorf("verify frozen candidate immediately before direct browser capture: %w", err)
	}
	if finalChallenge, err := readChallenge(config.ChallengePath); err != nil || !reflect.DeepEqual(finalChallenge, challenge) {
		if err != nil {
			return fmt.Errorf("re-read independent challenge immediately before direct browser capture: %w", err)
		}
		return fmt.Errorf("independent challenge changed before direct browser capture")
	}

	runCommand := func(name string, arguments ...string) ([]byte, error) {
		output, commandErr := environment.RunCommand(context.Background(), name, arguments...)
		if commandErr != nil {
			return nil, fmt.Errorf("direct capture command %s failed: %w: %s", name, commandErr, strings.TrimSpace(string(output)))
		}
		return output, nil
	}
	productVersion, err := runCommand("sw_vers", "-productVersion")
	if err != nil {
		return err
	}
	productVersion = []byte(strings.Trim(strings.TrimSpace(string(productVersion)), "\""))
	if !captureVersion.Match(productVersion) {
		return fmt.Errorf("derived macOS product version is invalid: %q", productVersion)
	}
	platformVersion := "macOS " + string(productVersion)
	browserVersionBytes, err := runCommand("mdls", "-name", "kMDItemVersion", "-raw", filepath.Join("/Applications", plan.BrowserApp+".app"))
	if err != nil {
		return err
	}
	browserVersionBytes = []byte(strings.Trim(strings.TrimSpace(string(browserVersionBytes)), "\""))
	if !captureVersion.Match(browserVersionBytes) {
		return fmt.Errorf("derived %s version is invalid: %q", plan.BrowserName, browserVersionBytes)
	}
	browserVersion := plan.BrowserName + " " + string(browserVersionBytes)
	voiceOverVersionBytes, err := runCommand("mdls", "-name", "kMDItemVersion", "-raw", "/System/Library/CoreServices/VoiceOver.app")
	if err != nil {
		return err
	}
	voiceOverVersionBytes = []byte(strings.Trim(strings.TrimSpace(string(voiceOverVersionBytes)), "\""))
	if !captureVersion.Match(voiceOverVersionBytes) {
		return fmt.Errorf("derived VoiceOver version is invalid: %q", voiceOverVersionBytes)
	}
	screenReaderVersion := "VoiceOver " + string(voiceOverVersionBytes)
	if _, err := runCommand("open", "-a", plan.BrowserApp, server.URL.String()); err != nil {
		return err
	}
	observations, traceEvents, browserStates, voiceOverCaptions, actionRecords, voiceOverLog, err := captureFixedOutcomes(plan, runCommand, directory, server, identity, challenge)
	if err != nil {
		return err
	}
	browserWindowPath := filepath.Join(directory, "browser-window.json")
	rawBrowserWindow, err := runCommand("osascript", "-e", browserWindowScript(plan, identity, challenge))
	if err != nil {
		return fmt.Errorf("capture direct browser window scope: %w", err)
	}
	window, err := parseBrowserWindowCapture(rawBrowserWindow, plan, identity, challenge)
	if err != nil {
		return err
	}
	if err := writeBytesExclusive(browserWindowPath, rawBrowserWindow); err != nil {
		return fmt.Errorf("write direct browser window scope: %w", err)
	}
	browserWindowArtifact, err := artifactForFile(browserWindowPath)
	if err != nil {
		return err
	}
	screenshotPath := filepath.Join(directory, "screenshot.png")
	if _, err := runCommand("screencapture", "-x", "-R", window.captureRect(), screenshotPath); err != nil {
		return err
	}
	screenshot, err := artifactForFile(screenshotPath)
	if err != nil {
		return err
	}
	if err := validateCapturedScreenshot(screenshotPath, window); err != nil {
		return err
	}
	if err := writeBytesExclusive(filepath.Join(directory, "voiceover-system-log.json"), voiceOverLog); err != nil {
		return fmt.Errorf("write directly captured VoiceOver system log: %w", err)
	}
	voiceOverLogArtifact, err := artifactForFile(filepath.Join(directory, "voiceover-system-log.json"))
	if err != nil {
		return err
	}

	capturedAt := environment.Now().UTC().Format(time.RFC3339)
	receiptID := receiptIdentityFor(identity)
	transcriptPath := filepath.Join(directory, "transcript.json")
	if err := writeJSONExclusive(transcriptPath, transcript{
		Schema: captureTranscriptSchema, Challenge: challenge.Challenge, Pair: plan.Pair, CapturedAt: capturedAt,
		Identity: receiptID, PlatformVersion: platformVersion, BrowserVersion: browserVersion,
		ScreenReaderVersion: screenReaderVersion, Route: captureRoute, Observations: observations,
	}); err != nil {
		return fmt.Errorf("write directly generated transcript: %w", err)
	}
	transcriptArtifact, err := artifactForFile(transcriptPath)
	if err != nil {
		return err
	}
	tracePath := filepath.Join(directory, "trace-log.json")
	if err := writeJSONExclusive(tracePath, traceLog{
		Schema: captureTraceSchema, Challenge: challenge.Challenge, Pair: plan.Pair, CapturedAt: capturedAt,
		Identity: receiptID, Route: captureRoute, Events: traceEvents,
	}); err != nil {
		return fmt.Errorf("write directly generated trace log: %w", err)
	}
	traceArtifact, err := artifactForFile(tracePath)
	if err != nil {
		return err
	}
	capture := evidenceCapture{
		Pair: plan.Pair, PlatformVersion: platformVersion, BrowserVersion: browserVersion,
		ScreenReaderVersion: screenReaderVersion, Route: captureRoute,
		CapturedAt: capturedAt, Observations: observations, ServedPage: servedPageArtifact,
		ServedResponse: servedResponseArtifact, BrowserStates: browserStates, ActionRecords: actionRecords, VoiceOverCaptions: voiceOverCaptions,
		VoiceOverLog: voiceOverLogArtifact, BrowserWindow: browserWindowArtifact, Screenshot: screenshot,
		Transcript: transcriptArtifact, TraceLog: traceArtifact,
	}
	if err := verifyCandidateIdentity(repositoryRoot, identity); err != nil {
		return fmt.Errorf("verify frozen candidate immediately before signing: %w", err)
	}
	if finalChallenge, err := readChallenge(config.ChallengePath); err != nil || !reflect.DeepEqual(finalChallenge, challenge) {
		if err != nil {
			return fmt.Errorf("re-read independent challenge immediately before signing: %w", err)
		}
		return fmt.Errorf("independent challenge changed before signing")
	}
	payload, err := json.Marshal(attestationPayload{Schema: captureAttestationPayloadSchema, Challenge: challenge.Challenge, Identity: receiptID, Capture: attestCapture(capture)})
	if err != nil {
		return fmt.Errorf("encode AT attestation payload: %w", err)
	}
	signature := ed25519.Sign(signingKey.Private, dssePAE(captureAttestationPayloadType, payload))
	envelope := dsseEnvelope{
		Schema:      captureAttestationSchema,
		PayloadType: captureAttestationPayloadType,
		Payload:     base64.StdEncoding.EncodeToString(payload),
		Signatures: []dsseSignature{{
			KeyID:     signingKey.ID,
			Signature: base64.StdEncoding.EncodeToString(signature),
		}},
	}
	attestationPath := filepath.Join(directory, "attestation.json")
	if err := writeJSONExclusive(attestationPath, envelope); err != nil {
		return fmt.Errorf("write signed AT attestation: %w", err)
	}
	attestationArtifact, err := artifactForFile(attestationPath)
	if err != nil {
		return err
	}
	capture.Attestation = attestationArtifact
	bundle := struct {
		Schema    string           `json:"schema"`
		Challenge captureChallenge `json:"challenge"`
		Identity  receiptIdentity  `json:"identity"`
		Capture   evidenceCapture  `json:"capture"`
	}{Schema: captureBundleSchema, Challenge: challenge, Identity: receiptID, Capture: capture}
	if err := writeJSONExclusive(filepath.Join(directory, "capture.json"), bundle); err != nil {
		return fmt.Errorf("write signed capture bundle: %w", err)
	}
	fmt.Fprintf(stdout, "T-GS-011 directly captured %s bundle at %s. Capture remains PENDING until second pair and final receipt verification.\n", plan.Pair, directory)
	return nil
}

func repositoryRoot(environment captureRuntime) (string, error) {
	output, err := environment.RunCommand(context.Background(), "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("resolve Git repository root: %w", err)
	}
	root := strings.TrimSpace(string(output))
	if root == "" {
		return "", fmt.Errorf("resolve Git repository root: empty output")
	}
	return root, nil
}

func readCandidateIdentity(path string) (candidateIdentity, error) {
	var identity candidateIdentity
	content, err := os.ReadFile(path)
	if err != nil {
		return identity, fmt.Errorf("read frozen candidate identity: %w", err)
	}
	if err := decodeStrictJSON(content, &identity); err != nil {
		return identity, fmt.Errorf("decode frozen candidate identity: %w", err)
	}
	if identity.Schema != "goshtoso.t-gs-011.candidate-identity.v2" || identity.RepositoryURL == "" || !captureGitPattern.MatchString(identity.Head) || !captureGitPattern.MatchString(identity.Tree) || !captureGitPattern.MatchString(identity.CandidateTree) || !captureSHA256Pattern.MatchString(identity.ManifestSHA256) || !captureSHA256Pattern.MatchString(identity.StatusSHA256) || len(identity.Paths) == 0 {
		return identity, fmt.Errorf("frozen candidate identity shape is invalid")
	}
	return identity, nil
}

func verifyCandidateIdentity(repositoryRoot string, identity candidateIdentity) error {
	return scrollregionidentity.VerifyIdentity(repositoryRoot, scrollRegionIdentityFromCapture(identity))
}

func scrollRegionIdentityFromCapture(identity candidateIdentity) scrollregionidentity.CandidateIdentity {
	paths := make([]scrollregionidentity.CandidatePath, 0, len(identity.Paths))
	for _, path := range identity.Paths {
		paths = append(paths, scrollregionidentity.CandidatePath{Path: path.Path, SHA256: path.SHA256})
	}
	return scrollregionidentity.CandidateIdentity{
		Schema: identity.Schema, RepositoryURL: identity.RepositoryURL, Head: identity.Head, Tree: identity.Tree,
		CandidateTree: identity.CandidateTree, ManifestSHA256: identity.ManifestSHA256, StatusSHA256: identity.StatusSHA256,
		Paths: paths,
		DependencyPins: scrollregionidentity.DependencyPins{
			RootGoDirective: identity.DependencyPins.RootGoDirective, SiteGoDirective: identity.DependencyPins.SiteGoDirective,
			Templ: identity.DependencyPins.Templ, PlaywrightGo: identity.DependencyPins.PlaywrightGo,
			AxeCore: identity.DependencyPins.AxeCore, AxeArchiveSHA256: identity.DependencyPins.AxeArchiveSHA256,
			AxeScriptSHA256: identity.DependencyPins.AxeScriptSHA256,
		},
	}
}

func readChallenge(path string) (captureChallenge, error) {
	var challenge captureChallenge
	content, err := os.ReadFile(path)
	if err != nil {
		return challenge, fmt.Errorf("read independently generated capture challenge: %w", err)
	}
	if err := decodeStrictJSON(content, &challenge); err != nil {
		return challenge, fmt.Errorf("decode independently generated capture challenge: %w", err)
	}
	if challenge.Schema != captureChallengeSchema || !captureSHA256Pattern.MatchString(challenge.Challenge) {
		return challenge, fmt.Errorf("capture challenge schema or random challenge is invalid")
	}
	if _, err := time.Parse(time.RFC3339, challenge.IssuedAt); err != nil {
		return challenge, fmt.Errorf("capture challenge issued_at must be RFC3339: %w", err)
	}
	return challenge, nil
}

func readAuthorizedSigningKey(repositoryRoot, pair, privatePath string) (authorizedSigningKey, error) {
	keys, err := loadTrustedKeys(repositoryRoot)
	if err != nil {
		return authorizedSigningKey{}, err
	}
	var trusted trustedKey
	found := false
	for _, candidate := range keys {
		if _, allowed := candidate.Pairs[pair]; allowed {
			trusted, found = candidate, true
			break
		}
	}
	if !found {
		return authorizedSigningKey{}, fmt.Errorf("no source-pinned capture authority permits pair %q", pair)
	}
	info, err := os.Stat(privatePath)
	if err != nil {
		return authorizedSigningKey{}, fmt.Errorf("stat capture authority private key: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		return authorizedSigningKey{}, fmt.Errorf("capture authority private key must be a regular owner-only file")
	}
	content, err := os.ReadFile(privatePath)
	if err != nil {
		return authorizedSigningKey{}, fmt.Errorf("read capture authority private key: %w", err)
	}
	block, remainder := pem.Decode(content)
	if block == nil || block.Type != "PRIVATE KEY" || len(bytes.TrimSpace(remainder)) != 0 {
		return authorizedSigningKey{}, fmt.Errorf("capture authority private key PEM is invalid")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return authorizedSigningKey{}, fmt.Errorf("parse capture authority private key: %w", err)
	}
	private, ok := parsed.(ed25519.PrivateKey)
	if !ok || len(private) != ed25519.PrivateKeySize {
		return authorizedSigningKey{}, fmt.Errorf("capture authority private key is not Ed25519")
	}
	if !bytes.Equal(private.Public().(ed25519.PublicKey), trusted.Public) {
		return authorizedSigningKey{}, fmt.Errorf("capture authority private key does not match source-pinned key %q for pair %q", trusted.ID, pair)
	}
	return authorizedSigningKey{ID: trusted.ID, Private: private}, nil
}

func loadTrustedKeys(repositoryRoot string) (map[string]trustedKey, error) {
	manifestPath := filepath.Join(repositoryRoot, "tests", "external", "scrollregion-a11y", "attestation-keys.json")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read source-pinned AT key manifest: %w", err)
	}
	var manifest trustedKeyManifest
	if err := decodeStrictJSON(content, &manifest); err != nil {
		return nil, fmt.Errorf("decode source-pinned AT key manifest: %w", err)
	}
	if manifest.Schema != captureTrustedKeysSchema || len(manifest.Keys) < 2 {
		return nil, fmt.Errorf("source-pinned AT key manifest schema or key set is invalid")
	}
	keys := make(map[string]trustedKey, len(manifest.Keys))
	for _, reference := range manifest.Keys {
		if !captureKeyID.MatchString(reference.KeyID) || !captureSHA256Pattern.MatchString(reference.PublicKeySHA256) || len(reference.Pairs) == 0 || filepath.IsAbs(reference.PublicKeyPath) || strings.Contains(reference.PublicKeyPath, "..") {
			return nil, fmt.Errorf("source-pinned AT key reference is invalid")
		}
		if _, duplicate := keys[reference.KeyID]; duplicate {
			return nil, fmt.Errorf("source-pinned AT key manifest repeats key ID %q", reference.KeyID)
		}
		publicBytes, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(reference.PublicKeyPath)))
		if err != nil {
			return nil, fmt.Errorf("read source-pinned AT public key %q: %w", reference.KeyID, err)
		}
		if sha256Hex(publicBytes) != reference.PublicKeySHA256 {
			return nil, fmt.Errorf("source-pinned AT public key %q fingerprint mismatch", reference.KeyID)
		}
		block, remainder := pem.Decode(publicBytes)
		if block == nil || block.Type != "PUBLIC KEY" || len(bytes.TrimSpace(remainder)) != 0 {
			return nil, fmt.Errorf("source-pinned AT public key %q PEM is invalid", reference.KeyID)
		}
		parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse source-pinned AT public key %q: %w", reference.KeyID, err)
		}
		public, ok := parsed.(ed25519.PublicKey)
		if !ok || len(public) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("source-pinned AT public key %q is not Ed25519", reference.KeyID)
		}
		pairs := make(map[string]struct{}, len(reference.Pairs))
		for _, pair := range reference.Pairs {
			if pair != "macos-safari-voiceover" && pair != "macos-chromium-voiceover" && pair != "windows-chromium-nvda" {
				return nil, fmt.Errorf("source-pinned AT key %q has unsupported pair %q", reference.KeyID, pair)
			}
			if _, duplicate := pairs[pair]; duplicate {
				return nil, fmt.Errorf("source-pinned AT key %q repeats pair %q", reference.KeyID, pair)
			}
			pairs[pair] = struct{}{}
		}
		keys[reference.KeyID] = trustedKey{ID: reference.KeyID, Pairs: pairs, Public: public}
	}
	return keys, nil
}

func prepareOutputDirectory(path string) (string, error) {
	directory, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve capture output directory: %w", err)
	}
	info, statErr := os.Stat(directory)
	switch {
	case statErr == nil:
		if !info.IsDir() {
			return "", fmt.Errorf("capture output path is not a directory")
		}
		entries, err := os.ReadDir(directory)
		if err != nil {
			return "", fmt.Errorf("read capture output directory: %w", err)
		}
		if len(entries) != 0 {
			return "", fmt.Errorf("capture output directory must be new or empty")
		}
	case errors.Is(statErr, os.ErrNotExist):
		if err := os.MkdirAll(directory, 0o700); err != nil {
			return "", fmt.Errorf("create capture output directory: %w", err)
		}
	default:
		return "", fmt.Errorf("stat capture output directory: %w", statErr)
	}
	return directory, nil
}

func artifactForFile(path string) (evidenceArtifact, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return evidenceArtifact{}, err
	}
	info, err := os.Lstat(absolute)
	if err != nil {
		return evidenceArtifact{}, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return evidenceArtifact{}, fmt.Errorf("directly captured artifact must be a regular non-symlink file")
	}
	content, err := os.ReadFile(absolute)
	if err != nil {
		return evidenceArtifact{}, err
	}
	return evidenceArtifact{Path: absolute, SHA256: sha256Hex(content)}, nil
}

func validateCapturedScreenshot(path string, window browserWindowCapture) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(content) < 100 {
		return fmt.Errorf("directly captured screenshot is unexpectedly small")
	}
	// The final validator also decodes and rejects a solid PNG before semantics.
	// This capture command has no image-processing dependency or alternate path.
	if !bytes.HasPrefix(content, []byte("\x89PNG\r\n\x1a\n")) {
		return fmt.Errorf("directly captured screenshot is not PNG")
	}
	decoded, format, err := image.Decode(bytes.NewReader(content))
	if err != nil || format != "png" {
		return fmt.Errorf("directly captured screenshot is not decodable PNG")
	}
	bounds := decoded.Bounds()
	if window.Window.Width <= 0 || window.Window.Height <= 0 || window.CandidateRegion.Width <= 0 || window.CandidateRegion.Height <= 0 {
		return fmt.Errorf("directly captured screenshot has invalid browser/window region geometry")
	}
	scaleX := float64(bounds.Dx()) / window.Window.Width
	scaleY := float64(bounds.Dy()) / window.Window.Height
	if scaleX <= 0 || scaleY <= 0 {
		return fmt.Errorf("directly captured screenshot has invalid pixel scale")
	}
	left := int(window.CandidateRegion.X * scaleX)
	top := int(window.CandidateRegion.Y * scaleY)
	right := int((window.CandidateRegion.X + window.CandidateRegion.Width) * scaleX)
	bottom := int((window.CandidateRegion.Y + window.CandidateRegion.Height) * scaleY)
	if left < bounds.Min.X || top < bounds.Min.Y || right <= left || bottom <= top || right > bounds.Max.X || bottom > bounds.Max.Y {
		return fmt.Errorf("directly captured screenshot does not contain the claimed named ScrollRegion window crop")
	}
	colors := make(map[[4]uint32]struct{})
	for y := top; y < bottom; y += max(1, (bottom-top)/8) {
		for x := left; x < right; x += max(1, (right-left)/8) {
			red, green, blue, alpha := decoded.At(x, y).RGBA()
			colors[[4]uint32{red, green, blue, alpha}] = struct{}{}
		}
	}
	if len(colors) < 4 {
		return fmt.Errorf("directly captured named ScrollRegion crop lacks visual structure")
	}
	return nil
}

func writeJSONExclusive(path string, value any) error {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeBytesExclusive(path, append(encoded, '\n'))
}

func writeBytesExclusive(path string, content []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(content); err != nil {
		return err
	}
	return file.Sync()
}

func decodeStrictJSON(content []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON values")
		}
		return err
	}
	return nil
}

func sha256Hex(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}

func dssePAE(payloadType string, payload []byte) []byte {
	return fmt.Appendf(nil, "DSSEv1 %d %s %d %s", len([]byte(payloadType)), payloadType, len(payload), payload)
}
