//go:build e2e && scrollregion && bfull && axe

package e2e

import (
	"archive/zip"
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestScrollRegionCandidateIdentityRejectsMutuallyConsistentFabrication(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identity.Head = strings.Repeat("a", 40)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)

	_, err := verifyScrollRegionCandidateIdentity(repository, identityPath)
	require.ErrorContains(t, err, "HEAD mismatch")
}

func TestScrollRegionCandidateIdentityRejectsMissingDeclaredPath(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	require.NoError(t, os.Remove(filepath.Join(repository, "evidence.txt")))
	scrollRegionRefreshFixtureCandidateIdentity(t, repository, &identity)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)

	_, err := verifyScrollRegionCandidateIdentity(repository, identityPath)
	require.ErrorContains(t, err, `read candidate path "evidence.txt"`)
}

func TestScrollRegionCandidateIdentityRejectsUndeclaredExtraPath(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	require.NoError(t, os.WriteFile(filepath.Join(repository, "extra.txt"), []byte("unexpected\n"), 0o600))
	scrollRegionRefreshFixtureCandidateIdentity(t, repository, &identity)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)

	_, err := verifyScrollRegionCandidateIdentity(repository, identityPath)
	require.ErrorContains(t, err, "candidate identity paths mismatch")
}

func TestScrollRegionCandidateIdentityRejectsTamperedDeclaredBytes(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	require.NoError(t, os.WriteFile(filepath.Join(repository, "evidence.txt"), []byte("tampered\n"), 0o600))
	scrollRegionRefreshFixtureCandidateIdentity(t, repository, &identity)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)

	_, err := verifyScrollRegionCandidateIdentity(repository, identityPath)
	require.ErrorContains(t, err, `candidate path "evidence.txt" SHA-256 mismatch`)
}

func TestScrollRegionCandidateIdentityRejectsDuplicatePath(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identity.Paths = append(identity.Paths, identity.Paths[0])
	identity.ManifestSHA256 = scrollRegionCandidateManifestSHA256(identity.Paths)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)

	_, err := verifyScrollRegionCandidateIdentity(repository, identityPath)
	require.ErrorContains(t, err, "paths must be unique and sorted")
}

func TestScrollRegionATReceiptRejectsReusedArtifacts(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	t.Setenv(scrollRegionATChallengeEnvironment, scrollRegionATChallengePathForFixture(repository))
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	receipt := scrollRegionATReceiptForFixture(t, repository, identity)
	receipt.Captures[1].Screenshot = receipt.Captures[0].Screenshot
	scrollRegionRefreshATFixtureCapture(t, repository, &receipt, 1, false)
	receiptPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "receipt.json", receipt)

	err := validateScrollRegionATReceipt(repository, identityPath, receiptPath)
	require.ErrorContains(t, err, "reused")
}

// A self-consistent receipt and artifacts are insufficient: final AT evidence
// must carry a trusted, pair-bound signed capture envelope.
func TestScrollRegionATReceiptRejectsUnsignedCapture(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	t.Setenv(scrollRegionATChallengeEnvironment, scrollRegionATChallengePathForFixture(repository))
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	receipt := scrollRegionATReceiptForFixture(t, repository, identity)
	for index := range receipt.Captures {
		receipt.Captures[index].Attestation = scrollRegionATEvidenceArtifact{}
	}
	receiptPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "receipt.json", receipt)

	err := validateScrollRegionATReceipt(repository, identityPath, receiptPath)
	require.ErrorContains(t, err, "signed attestation")
}

func TestScrollRegionATReceiptRejectsAttestationPayloadMutation(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	t.Setenv(scrollRegionATChallengeEnvironment, scrollRegionATChallengePathForFixture(repository))
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	receipt := scrollRegionATReceiptForFixture(t, repository, identity)
	scrollRegionMutateATFixtureAttestation(t, &receipt, 0, func(envelope *scrollRegionATDSSEEnvelope) {
		raw, err := base64.StdEncoding.DecodeString(envelope.Payload)
		require.NoError(t, err)
		var payload scrollRegionATAttestationPayload
		require.NoError(t, scrollRegionDecodeStrictJSON(raw, &payload))
		payload.Challenge = strings.Repeat("a", 64)
		raw, err = json.Marshal(payload)
		require.NoError(t, err)
		envelope.Payload = base64.StdEncoding.EncodeToString(raw)
	})
	receiptPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "receipt.json", receipt)

	err := validateScrollRegionATReceipt(repository, identityPath, receiptPath)
	require.ErrorContains(t, err, "signature verification failed")
}

func TestScrollRegionATReceiptRejectsWrongAttestationKey(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	t.Setenv(scrollRegionATChallengeEnvironment, scrollRegionATChallengePathForFixture(repository))
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	receipt := scrollRegionATReceiptForFixture(t, repository, identity)
	scrollRegionMutateATFixtureAttestation(t, &receipt, 0, func(envelope *scrollRegionATDSSEEnvelope) {
		envelope.Signatures[0].KeyID = "unknown-capture-authority"
	})
	receiptPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "receipt.json", receipt)

	err := validateScrollRegionATReceipt(repository, identityPath, receiptPath)
	require.ErrorContains(t, err, "uses untrusted key")
}

func TestScrollRegionATReceiptRejectsPairKeyMismatch(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	t.Setenv(scrollRegionATChallengeEnvironment, scrollRegionATChallengePathForFixture(repository))
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	receipt := scrollRegionATReceiptForFixture(t, repository, identity)
	capture := &receipt.Captures[0]
	payload := scrollRegionATAttestationPayload{
		Schema:    scrollRegionATAttestationPayloadSchema,
		Challenge: receipt.Challenge,
		Identity:  receipt.Identity,
		Capture:   scrollRegionATAttestedCaptureFromEvidence(*capture),
	}
	path := writeScrollRegionATFixtureAttestationWithKey(t, repository, filepath.Dir(capture.Attestation.Path), filepath.Base(capture.Attestation.Path), "fixture-windows-nvda", payload)
	capture.Attestation = scrollRegionEvidenceArtifactForFile(t, path)
	receiptPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "receipt.json", receipt)

	err := validateScrollRegionATReceipt(repository, identityPath, receiptPath)
	require.ErrorContains(t, err, "not trusted for pair")
}

func TestScrollRegionATReceiptRejectsChallengeReplay(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	receiptPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "receipt.json", scrollRegionATReceiptForFixture(t, repository, identity))
	freshChallenge := scrollRegionATChallenge{
		Schema:    scrollRegionATChallengeSchema,
		Challenge: "ace1f3dcbe3fb9a7a93a1c90e95a24f0c4a4e6d2be882719bef52544d86b0f3c",
		IssuedAt:  time.Date(2026, time.August, 12, 15, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}
	challengePath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "fresh-challenge.json", freshChallenge)
	t.Setenv(scrollRegionATChallengeEnvironment, challengePath)

	err := validateScrollRegionATReceipt(repository, identityPath, receiptPath)
	require.ErrorContains(t, err, "challenge does not match")
}

func TestScrollRegionATReceiptRejectsCandidateByteMutationBeforeReplayClaim(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	t.Setenv(scrollRegionATChallengeEnvironment, scrollRegionATChallengePathForFixture(repository))
	registry := filepath.Join(t.TempDir(), "independent-challenge-registry")
	require.NoError(t, os.Mkdir(registry, 0o700))
	t.Setenv(scrollRegionATReplayRegistryEnvironment, registry)
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	receiptPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "receipt.json", scrollRegionATReceiptForFixture(t, repository, identity))

	err := validateScrollRegionATReceiptWithBeforeClaim(repository, identityPath, receiptPath, func() error {
		return os.WriteFile(filepath.Join(repository, "evidence.txt"), []byte("mutated after first verification\n"), 0o600)
	})
	require.ErrorContains(t, err, "final frozen candidate identity")
	entries, readErr := os.ReadDir(registry)
	require.NoError(t, readErr)
	require.Empty(t, entries, "identity drift must reject before consuming the external challenge")
}

// A challenge may validate one immutable final receipt repeatedly for audit,
// but cannot be claimed by a second otherwise valid receipt. That makes the
// capture request single-use without making ordinary byte revalidation flaky.
func TestScrollRegionATReceiptRejectsSecondReceiptForClaimedChallenge(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	t.Setenv(scrollRegionATChallengeEnvironment, scrollRegionATChallengePathForFixture(repository))
	registry := filepath.Join(t.TempDir(), "independent-challenge-registry")
	require.NoError(t, os.Mkdir(registry, 0o700))
	t.Setenv(scrollRegionATReplayRegistryEnvironment, registry)
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	firstPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "first.json", scrollRegionATReceiptForFixture(t, repository, identity))
	require.NoError(t, validateScrollRegionATReceipt(repository, identityPath, firstPath))
	secondPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "second.json", scrollRegionATReceiptForFixture(t, repository, identity))

	err := validateScrollRegionATReceipt(repository, identityPath, secondPath)
	require.ErrorContains(t, err, "challenge already claimed")
}

// An authorized signer can fabricate patterned PNG/transcript/trace bytes.
// Raw candidate/browser/VoiceOver provenance must therefore be mandatory even
// after the fixture's valid ephemeral DSSE signature is refreshed.
func TestScrollRegionATReceiptRejectsSignedSyntheticCaptureWithoutRawProvenance(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	t.Setenv(scrollRegionATChallengeEnvironment, scrollRegionATChallengePathForFixture(repository))
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	receipt := scrollRegionATReceiptForFixture(t, repository, identity)
	receipt.Captures[0].ServedPage = scrollRegionATEvidenceArtifact{}
	receipt.Captures[0].ServedResponse = scrollRegionATEvidenceArtifact{}
	receipt.Captures[0].BrowserStates = nil
	receipt.Captures[0].VoiceOverCaptions = nil
	receipt.Captures[0].VoiceOverLog = scrollRegionATEvidenceArtifact{}
	receipt.Captures[0].BrowserWindow = scrollRegionATEvidenceArtifact{}
	scrollRegionRefreshATFixtureCapture(t, repository, &receipt, 0, false)
	receiptPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "receipt.json", receipt)

	err := validateScrollRegionATReceipt(repository, identityPath, receiptPath)
	require.ErrorContains(t, err, "served candidate HTML artifact")
}

// Before V8 the receipt accepted a claimant-authored focus-navigation field
// after it had been signed. This RED proves that an apparent traversal must be
// re-derived from the browser's own before/after/exit snapshots instead.
func TestScrollRegionATReceiptRejectsInferredFocusNavigation(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	t.Setenv(scrollRegionATChallengeEnvironment, scrollRegionATChallengePathForFixture(repository))
	registry := filepath.Join(t.TempDir(), "independent-challenge-registry")
	require.NoError(t, os.Mkdir(registry, 0o700))
	t.Setenv(scrollRegionATReplayRegistryEnvironment, registry)
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	receipt := scrollRegionATReceiptForFixture(t, repository, identity)
	capture := &receipt.Captures[0]
	capture.Observations[1].FocusNavigation.Before = "Invented focus predecessor Current release"
	transcript := scrollRegionATTranscript{
		Schema: scrollRegionATTranscriptSchema, Challenge: receipt.Challenge, Pair: capture.Pair, CapturedAt: capture.CapturedAt,
		Identity: receipt.Identity, PlatformVersion: capture.PlatformVersion, BrowserVersion: capture.BrowserVersion,
		ScreenReaderVersion: capture.ScreenReaderVersion, Route: capture.Route, Observations: capture.Observations,
	}
	transcriptPath := writeScrollRegionEvidenceJSON(t, filepath.Dir(capture.Transcript.Path), filepath.Base(capture.Transcript.Path), transcript)
	capture.Transcript = scrollRegionEvidenceArtifactForFile(t, transcriptPath)
	scrollRegionRefreshATFixtureCapture(t, repository, &receipt, 0, false)
	receiptPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "receipt.json", receipt)

	err := validateScrollRegionATReceipt(repository, identityPath, receiptPath)
	require.Error(t, err, "validator must reject a signed inferred focus-navigation claim")
}

// A phrase from a prior state must not be re-labeled as current just because a
// signer refreshed the envelope. The raw caption needs this action's unique
// challenge-derived token, PID, subsystem, and non-overlapping action record.
func TestScrollRegionATReceiptRejectsStaleActionBoundVoiceOverLog(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	t.Setenv(scrollRegionATChallengeEnvironment, scrollRegionATChallengePathForFixture(repository))
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	receipt := scrollRegionATReceiptForFixture(t, repository, identity)
	capture := &receipt.Captures[0]
	var action scrollRegionATActionRecord
	actionBytes, err := os.ReadFile(capture.ActionRecords[1].Path)
	require.NoError(t, err)
	require.NoError(t, scrollRegionDecodeStrictJSON(actionBytes, &action))
	// Exact state token but timestamped before the issued key event: a valid
	// JSON shape cannot make stale page-load speech look causal.
	stale := []scrollRegionATVoiceOverLogEvent{{Timestamp: action.BeforeAt, ProcessID: action.VoiceOverPID, Subsystem: action.VoiceOverSubsystem, EventMessage: "VoiceOver pair " + capture.Pair + " named region Current release action_token " + action.ActionToken + " stale before action"}}
	path := writeScrollRegionEvidenceJSON(t, t.TempDir(), "stale-voiceover-no-overflow.json", stale)
	capture.VoiceOverCaptions[1] = scrollRegionEvidenceArtifactForFile(t, path)
	scrollRegionRefreshATFixtureCapture(t, repository, &receipt, 0, false)
	receiptPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "receipt.json", receipt)

	err = validateScrollRegionATReceipt(repository, identityPath, receiptPath)
	require.ErrorContains(t, err, "exact action interval")
}

func TestScrollRegionATReceiptRejectsClaimantActionTimestampMismatch(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	t.Setenv(scrollRegionATChallengeEnvironment, scrollRegionATChallengePathForFixture(repository))
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	receipt := scrollRegionATReceiptForFixture(t, repository, identity)
	capture := &receipt.Captures[0]
	var action scrollRegionATActionRecord
	content, err := os.ReadFile(capture.ActionRecords[0].Path)
	require.NoError(t, err)
	require.NoError(t, scrollRegionDecodeStrictJSON(content, &action))
	action.AfterAt = action.ActionIssuedAt
	path := writeScrollRegionEvidenceJSON(t, filepath.Dir(capture.ActionRecords[0].Path), filepath.Base(capture.ActionRecords[0].Path), action)
	capture.ActionRecords[0] = scrollRegionEvidenceArtifactForFile(t, path)
	scrollRegionRefreshATFixtureCapture(t, repository, &receipt, 0, false)
	receiptPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "receipt.json", receipt)

	err = validateScrollRegionATReceipt(repository, identityPath, receiptPath)
	require.ErrorContains(t, err, "causal action interval")
}

// A signer can recompute the whole wrapper. The verifier must still derive
// each exact adapter-owned key event and the mandatory Tab exit contract from
// the fixed state map rather than accepting merely nonempty claimant strings.
func TestScrollRegionATReceiptRejectsActionContractMutation(t *testing.T) {
	for field, mutate := range map[string]func(*scrollRegionATActionRecord){
		"command":      func(action *scrollRegionATActionRecord) { action.Command = "ArrowDown" },
		"action-event": func(action *scrollRegionATActionRecord) { action.ActionEvent = "macOS System Events key code 126" },
		"exit-command": func(action *scrollRegionATActionRecord) { action.ExitCommand = "Shift+Tab" },
		"exit-event": func(action *scrollRegionATActionRecord) {
			action.ExitEvent = "macOS System Events key code 48 with shift"
		},
	} {
		t.Run(field, func(t *testing.T) {
			repository := newScrollRegionEvidenceFixtureRepository(t)
			t.Setenv(scrollRegionATChallengeEnvironment, scrollRegionATChallengePathForFixture(repository))
			identity := scrollRegionCandidateIdentityForFixture(t, repository)
			identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
			receipt := scrollRegionATReceiptForFixture(t, repository, identity)
			capture := &receipt.Captures[0]
			var action scrollRegionATActionRecord
			content, err := os.ReadFile(capture.ActionRecords[0].Path)
			require.NoError(t, err)
			require.NoError(t, scrollRegionDecodeStrictJSON(content, &action))
			mutate(&action)
			path := writeScrollRegionEvidenceJSON(t, filepath.Dir(capture.ActionRecords[0].Path), filepath.Base(capture.ActionRecords[0].Path), action)
			capture.ActionRecords[0] = scrollRegionEvidenceArtifactForFile(t, path)
			scrollRegionRefreshATFixtureCapture(t, repository, &receipt, 0, false)
			receiptPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "receipt.json", receipt)

			err = validateScrollRegionATReceipt(repository, identityPath, receiptPath)
			require.ErrorContains(t, err, "exact adapter action contract")
		})
	}
}

func TestScrollRegionATReceiptRejectsNoOpHomeTransition(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	t.Setenv(scrollRegionATChallengeEnvironment, scrollRegionATChallengePathForFixture(repository))
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	receipt := scrollRegionATReceiptForFixture(t, repository, identity)
	capture := &receipt.Captures[0]
	const startStateIndex = 2
	const beforePhaseOffset = 0
	artifactIndex := startStateIndex*3 + beforePhaseOffset
	var browser scrollRegionATBrowserState
	content, err := os.ReadFile(capture.BrowserStates[artifactIndex].Path)
	require.NoError(t, err)
	require.NoError(t, scrollRegionDecodeStrictJSON(content, &browser))
	browser.Snapshot.Boundary, browser.Snapshot.ScrollTop, browser.Snapshot.StartCueVisible, browser.Snapshot.EndCueVisible = "start", 0, false, true
	path := writeScrollRegionEvidenceJSON(t, filepath.Dir(capture.BrowserStates[artifactIndex].Path), filepath.Base(capture.BrowserStates[artifactIndex].Path), browser)
	capture.BrowserStates[artifactIndex] = scrollRegionEvidenceArtifactForFile(t, path)
	scrollRegionRefreshATFixtureCapture(t, repository, &receipt, 0, false)
	receiptPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "receipt.json", receipt)

	err = validateScrollRegionATReceipt(repository, identityPath, receiptPath)
	require.ErrorContains(t, err, "Home must move non-start boundary")
}

func TestScrollRegionATReceiptRejectsOverlappingActionWindow(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	t.Setenv(scrollRegionATChallengeEnvironment, scrollRegionATChallengePathForFixture(repository))
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	receipt := scrollRegionATReceiptForFixture(t, repository, identity)
	capture := &receipt.Captures[0]
	var first, second scrollRegionATActionRecord
	firstBytes, err := os.ReadFile(capture.ActionRecords[0].Path)
	require.NoError(t, err)
	require.NoError(t, scrollRegionDecodeStrictJSON(firstBytes, &first))
	secondBytes, err := os.ReadFile(capture.ActionRecords[1].Path)
	require.NoError(t, err)
	require.NoError(t, scrollRegionDecodeStrictJSON(secondBytes, &second))
	second.LogStartedAt = first.LogStartedAt
	secondPath := writeScrollRegionEvidenceJSON(t, filepath.Dir(capture.ActionRecords[1].Path), filepath.Base(capture.ActionRecords[1].Path), second)
	capture.ActionRecords[1] = scrollRegionEvidenceArtifactForFile(t, secondPath)
	scrollRegionRefreshATFixtureCapture(t, repository, &receipt, 0, false)
	receiptPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "receipt.json", receipt)

	err = validateScrollRegionATReceipt(repository, identityPath, receiptPath)
	require.ErrorContains(t, err, "stale, overlapping, or incomplete causal action interval")
}

func TestScrollRegionATReceiptRejectsSignedSyntheticSolidPNGBeforeSemantics(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	t.Setenv(scrollRegionATChallengeEnvironment, scrollRegionATChallengePathForFixture(repository))
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	receipt := scrollRegionATReceiptForFixture(t, repository, identity)
	solid := writeScrollRegionEvidenceSolidPNG(t, t.TempDir(), "claimant-solid.png", color.RGBA{R: 21, G: 88, B: 147, A: 255})
	receipt.Captures[0].Screenshot = scrollRegionEvidenceArtifactForFile(t, solid)
	scrollRegionRefreshATFixtureCapture(t, repository, &receipt, 0, false)
	receiptPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "receipt.json", receipt)

	err := validateScrollRegionATReceipt(repository, identityPath, receiptPath)
	require.ErrorContains(t, err, "lacks visual structure")
}

func TestScrollRegionATReceiptRejectsGenericObservedSpeech(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	t.Setenv(scrollRegionATChallengeEnvironment, scrollRegionATChallengePathForFixture(repository))
	registry := filepath.Join(t.TempDir(), "independent-challenge-registry")
	require.NoError(t, os.Mkdir(registry, 0o700))
	t.Setenv(scrollRegionATReplayRegistryEnvironment, registry)
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	receipt := scrollRegionATReceiptForFixture(t, repository, identity)
	receipt.Captures[0].Observations[0].ObservedSpeech = []string{"generic region Activity history evidence"}
	scrollRegionRefreshATFixtureCapture(t, repository, &receipt, 0, true)
	receiptPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "receipt.json", receipt)

	err := validateScrollRegionATReceipt(repository, identityPath, receiptPath)
	require.ErrorContains(t, err, "generic observed speech")
}

func TestScrollRegionATReceiptRejectsPlaceholderOrZeroVersion(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	t.Setenv(scrollRegionATChallengeEnvironment, scrollRegionATChallengePathForFixture(repository))
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	receipt := scrollRegionATReceiptForFixture(t, repository, identity)
	receipt.Captures[0].BrowserVersion = "Safari 0.0"
	scrollRegionRefreshATFixtureCapture(t, repository, &receipt, 0, true)
	receiptPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "receipt.json", receipt)

	err := validateScrollRegionATReceipt(repository, identityPath, receiptPath)
	require.ErrorContains(t, err, `browser version "Safari 0.0" violates pair contract`)
}

func TestScrollRegionATReceiptRejectsPlainTextScreenshot(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	t.Setenv(scrollRegionATChallengeEnvironment, scrollRegionATChallengePathForFixture(repository))
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	receipt := scrollRegionATReceiptForFixture(t, repository, identity)
	screenshotPath := receipt.Captures[0].Screenshot.Path
	plainText := []byte("not a PNG screenshot\n")
	require.NoError(t, os.WriteFile(screenshotPath, plainText, 0o600))
	receipt.Captures[0].Screenshot.SHA256 = scrollRegionBFullSHA256(plainText)
	scrollRegionRefreshATFixtureCapture(t, repository, &receipt, 0, false)
	receiptPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "receipt.json", receipt)

	err := validateScrollRegionATReceipt(repository, identityPath, receiptPath)
	require.ErrorContains(t, err, "screenshot is not decodable image data")
}

func TestScrollRegionBFullIdentityBindingRequiresSidecarForDirtyWorktree(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	t.Setenv(scrollRegionBFullIdentityEnvironment, "")
	_, err := resolveScrollRegionBFullIdentity(repository, true)
	require.ErrorContains(t, err, scrollRegionBFullIdentityEnvironment)
}

func TestScrollRegionBFullIdentityBindingSealsVerifiedSidecar(t *testing.T) {
	repository := newScrollRegionEvidenceFixtureRepository(t)
	identity := scrollRegionCandidateIdentityForFixture(t, repository)
	identityPath := writeScrollRegionEvidenceJSON(t, t.TempDir(), "identity.json", identity)
	t.Setenv(scrollRegionBFullIdentityEnvironment, identityPath)
	binding, err := resolveScrollRegionBFullIdentity(repository, true)
	require.NoError(t, err)
	require.Equal(t, "sealed-dirty-candidate", binding.Binding)
	require.Equal(t, scrollRegionReceiptIdentityFromCandidate(identity), *binding.Identity)
	require.NoError(t, binding.revalidate(repository))
}

func TestScrollRegionReadAxeLockAllowsHumanReviewComments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "axe-core.lock")
	require.NoError(t, os.WriteFile(path, []byte("# authenticated archive and payload pins\n\nversion=4.10.3\narchive_sha256=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef\nscript_sha256=abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789\n"), 0o600))

	values, err := scrollRegionReadAxeLock(path)
	require.NoError(t, err)
	require.Equal(t, "4.10.3", values["version"])
}

func TestScrollRegionBFullReceiptWrapperRejectsTampering(t *testing.T) {
	receipt := scrollRegionBFullReceiptForFixture(t)
	digest, err := scrollRegionBFullReceiptDigest(receipt)
	require.NoError(t, err)
	receipt.WrapperSHA256 = digest
	encoded, err := json.Marshal(receipt)
	require.NoError(t, err)

	_, err = validateScrollRegionBFullReceiptWrapper(encoded)
	require.NoError(t, err)
	receipt.ExpectedCells = 3
	encoded, err = json.Marshal(receipt)
	require.NoError(t, err)
	_, err = validateScrollRegionBFullReceiptWrapper(encoded)
	require.ErrorContains(t, err, "wrapper SHA-256 mismatch")
}

// The selector is deliberately absent from locked component docs. A signed
// receipt must therefore reject a self-consistent claim that a site theme was
// persisted through a client control; only server-routed initial HTML may own
// that visual axis.
func TestScrollRegionBFullReceiptRejectsLockedThemePersistenceClaim(t *testing.T) {
	receipt := scrollRegionBFullReceipt{
		Schema:        scrollRegionBFullReceiptSchema,
		Closure:       "diagnostic-non-closure",
		ExpectedCells: 1,
		Binding:       scrollRegionBFullIdentityBinding{Binding: "diagnostic-unbound-dirty-worktree"},
		ToolVersions:  scrollRegionBFullToolVersions{GoRuntime: "go1.26.5", PlaywrightGo: "v0.6100.0", BrowserByZoom: map[string]string{"default": "Chromium 140.0.0.0"}, AxeCore: scrollRegionAxeCoreVersion},
		Widths:        []int{390},
		Cells: []scrollRegionBFullCellReceipt{{
			CellID: "locked-theme-forgery",
			Persistence: scrollRegionBFullPersistenceProof{
				ThemeInitialSource:            "server-routed-html",
				ThemePersistenceNotApplicable: "client-theme-selector",
				DarkPersistence:               "product-ui",
			},
			NAs: map[string]string{"theme-persistence": "client-theme-selector"},
		}},
	}
	digest, err := scrollRegionBFullReceiptDigest(receipt)
	require.NoError(t, err)
	receipt.WrapperSHA256 = digest
	encoded, err := json.Marshal(receipt)
	require.NoError(t, err)

	_, err = validateScrollRegionBFullReceiptWrapper(encoded)
	require.ErrorContains(t, err, "theme persistence")
}

// The wrapper digest is only integrity metadata. An attacker can recompute it,
// so B-FULL semantic validation must independently decode every raw Page A/B
// artifact and reject a self-consistent forged wrapper.
func TestScrollRegionBFullReceiptRejectsRecomputedPersistenceForgeries(t *testing.T) {
	for name, mutate := range map[string]func(t *testing.T, receipt *scrollRegionBFullReceipt){
		"page-a-action": func(t *testing.T, receipt *scrollRegionBFullReceipt) {
			var action scrollRegionBFullActionEvidence
			content, err := os.ReadFile(receipt.Cells[0].Persistence.PageAActions[0].Path)
			require.NoError(t, err)
			require.NoError(t, scrollRegionDecodeStrictJSON(content, &action))
			action.After.Cookie = ""
			path := writeScrollRegionEvidenceJSON(t, filepath.Dir(receipt.Cells[0].Persistence.PageAActions[0].Path), filepath.Base(receipt.Cells[0].Persistence.PageAActions[0].Path), action)
			receipt.Cells[0].Persistence.PageAActions[0] = scrollRegionBFullEvidenceArtifactForFile(t, path)
		},
		"page-b-storage": func(t *testing.T, receipt *scrollRegionBFullReceipt) {
			var storage scrollRegionBFullStorageEvidence
			content, err := os.ReadFile(receipt.Cells[0].Persistence.FreshLoadStorage.Path)
			require.NoError(t, err)
			require.NoError(t, scrollRegionDecodeStrictJSON(content, &storage))
			storage.State.DarkMode = "true"
			path := writeScrollRegionEvidenceJSON(t, filepath.Dir(receipt.Cells[0].Persistence.FreshLoadStorage.Path), filepath.Base(receipt.Cells[0].Persistence.FreshLoadStorage.Path), storage)
			receipt.Cells[0].Persistence.FreshLoadStorage = scrollRegionBFullEvidenceArtifactForFile(t, path)
		},
		"first-paint": func(t *testing.T, receipt *scrollRegionBFullReceipt) {
			var paint scrollRegionBFullPaintEvidence
			content, err := os.ReadFile(receipt.Cells[0].Persistence.FreshLoadPaintArtifact.Path)
			require.NoError(t, err)
			require.NoError(t, scrollRegionDecodeStrictJSON(content, &paint))
			paint.Events[1].Dark = true
			path := writeScrollRegionEvidenceJSON(t, filepath.Dir(receipt.Cells[0].Persistence.FreshLoadPaintArtifact.Path), filepath.Base(receipt.Cells[0].Persistence.FreshLoadPaintArtifact.Path), paint)
			receipt.Cells[0].Persistence.FreshLoadPaintArtifact = scrollRegionBFullEvidenceArtifactForFile(t, path)
		},
		"initial-paint-claims-rendered-region": func(t *testing.T, receipt *scrollRegionBFullReceipt) {
			var paint scrollRegionBFullPaintEvidence
			content, err := os.ReadFile(receipt.Cells[0].Persistence.FreshLoadPaintArtifact.Path)
			require.NoError(t, err)
			require.NoError(t, scrollRegionDecodeStrictJSON(content, &paint))
			paint.Events[0].Visible = true
			paint.Events[0].Role = "region"
			paint.Events[0].Name = "Activity history"
			path := writeScrollRegionEvidenceJSON(t, filepath.Dir(receipt.Cells[0].Persistence.FreshLoadPaintArtifact.Path), filepath.Base(receipt.Cells[0].Persistence.FreshLoadPaintArtifact.Path), paint)
			receipt.Cells[0].Persistence.FreshLoadPaintArtifact = scrollRegionBFullEvidenceArtifactForFile(t, path)
			receipt.Cells[0].Persistence.FreshLoadInitialHTML = paint.Events[0]
		},
		"count-and-axes": func(t *testing.T, receipt *scrollRegionBFullReceipt) {
			receipt.ExpectedCells = 2
		},
		"artifact-hash": func(t *testing.T, receipt *scrollRegionBFullReceipt) {
			receipt.Cells[0].Persistence.PageAInitialHTML.SHA256 = strings.Repeat("a", 64)
		},
		"synthetic-trace": func(t *testing.T, receipt *scrollRegionBFullReceipt) {
			path := filepath.Join(filepath.Dir(receipt.Cells[0].Trace.Path), "synthetic.trace.zip")
			require.NoError(t, os.WriteFile(path, []byte("PK\x03\x04fixture-trace"), 0o600))
			trace := scrollRegionBFullEvidenceArtifactForFile(t, path)
			receipt.Cells[0].Trace = &trace
			receipt.TraceByCell[receipt.Cells[0].CellID] = trace
		},
		"trace-wrong-literal-cell": func(t *testing.T, receipt *scrollRegionBFullReceipt) {
			wrongTheme, ok := scrollRegionBFullThemeByID("goshtoso")
			require.True(t, ok)
			wrongCellID := scrollRegionBFullCellID(wrongTheme, false, 390, scrollRegionBFullZoom{ID: "default"})
			trace := scrollRegionBFullFixturePlaywrightTrace(t, filepath.Dir(receipt.Cells[0].Trace.Path), "wrong-cell.trace.zip", wrongCellID)
			receipt.Cells[0].Trace = &trace
			receipt.TraceByCell[receipt.Cells[0].CellID] = trace
		},
		"screenshot-unbound-named-region": func(t *testing.T, receipt *scrollRegionBFullReceipt) {
			receipt.Cells[0].Screenshot.Capture.Anchors = nil
		},
	} {
		t.Run(name, func(t *testing.T) {
			receipt := scrollRegionBFullReceiptForFixture(t)
			mutate(t, &receipt)
			digest, err := scrollRegionBFullReceiptDigest(receipt)
			require.NoError(t, err)
			receipt.WrapperSHA256 = digest
			encoded, err := json.Marshal(receipt)
			require.NoError(t, err)
			_, err = validateScrollRegionBFullReceiptWrapper(encoded)
			require.Error(t, err, "recomputed wrapper must not authenticate raw persistence semantics")
			if name == "trace-wrong-literal-cell" {
				require.ErrorContains(t, err, "literal trace cell", "trace rejection must derive from the raw cell axes rather than an incidental artifact check")
			}
		})
	}
}

func scrollRegionBFullReceiptForFixture(t *testing.T) scrollRegionBFullReceipt {
	t.Helper()
	directory := t.TempDir()
	cellID := scrollRegionBFullRoute + "|araihu|light|390|default"
	pageAHTML := []byte(`<html data-theme="araihu" data-goshtoso-theme-initial-source="server-routed-html"><div id="scroll-region-fragment"><div data-goshtoso-scroll-region><div data-goshtoso-scroll-viewport tabindex="0" role="region" aria-label="Activity history"></div></div></div></html>`)
	pageBHTML := append([]byte(nil), pageAHTML...)
	pageAInitial := scrollRegionBFullFixtureBytesArtifact(t, directory, "page-a.html", pageAHTML)
	pageBInitial := scrollRegionBFullFixtureBytesArtifact(t, directory, "page-b.html", pageBHTML)
	before := scrollRegionBFullPageState{DialogVisible: true}
	consented := scrollRegionBFullPageState{Cookie: "gt_storage=allowed", DialogVisible: false}
	dark := scrollRegionBFullPageState{Cookie: "gt_storage=allowed", DarkMode: "true", Dark: true, DialogVisible: false}
	settled := scrollRegionBFullPageState{Cookie: "gt_storage=allowed", DarkMode: "false", Dark: false, DialogVisible: false}
	storageBefore := scrollRegionBFullFixtureJSONArtifact(t, directory, "page-a-storage.json", scrollRegionBFullStorageEvidence{Schema: scrollRegionBFullStorageEvidenceSchema, CellID: cellID, Phase: "page-a-storage-before", State: before})
	consentAction := scrollRegionBFullActionEvidence{Schema: scrollRegionBFullActionEvidenceSchema, CellID: cellID, Phase: "page-a-consent", Before: before, Action: scrollRegionBFullAction{Before: "browser-storage-dialog-visible; gt_storage is absent", Action: "Playwright mouse click Allow browser storage", Return: "browser-storage-dialog-hidden; gt_storage=allowed"}, After: consented}
	resetAction := scrollRegionBFullActionEvidence{Schema: scrollRegionBFullActionEvidenceSchema, CellID: cellID, Phase: "page-a-dark-reset", Before: consented, Action: scrollRegionBFullAction{Before: "dark=false", Action: "Playwright mouse click dark mode toggle", Return: "dark=true; localStorage darkMode=true"}, After: dark}
	setAction := scrollRegionBFullActionEvidence{Schema: scrollRegionBFullActionEvidenceSchema, CellID: cellID, Phase: "page-a-dark-set", Before: dark, Action: scrollRegionBFullAction{Before: "dark=true", Action: "Playwright mouse click dark mode toggle", Return: "dark=false; localStorage darkMode=false"}, After: settled}
	actionArtifacts := []scrollRegionBFullArtifact{
		scrollRegionBFullFixtureJSONArtifact(t, directory, "page-a-action-00.json", consentAction),
		scrollRegionBFullFixtureJSONArtifact(t, directory, "page-a-action-01.json", resetAction),
		scrollRegionBFullFixtureJSONArtifact(t, directory, "page-a-action-02.json", setAction),
	}
	pageBStorage := scrollRegionBFullFixtureJSONArtifact(t, directory, "page-b-storage.json", scrollRegionBFullStorageEvidence{Schema: scrollRegionBFullStorageEvidenceSchema, CellID: cellID, Phase: "page-b-storage", State: settled})
	paint := scrollRegionBFullPaint{Phase: "init", ReadyState: "loading", Theme: "araihu", ThemeSource: scrollRegionBFullThemeInitialSource, Dark: false}
	first := paint
	first.Phase, first.ReadyState = "first-animation-frame", "interactive"
	first.Visible, first.Role, first.Name = true, "region", "Activity history"
	final := first
	final.Phase, final.ReadyState = "settled", "complete"
	paintArtifact := scrollRegionBFullFixtureJSONArtifact(t, directory, "page-b-paint.json", scrollRegionBFullPaintEvidence{Schema: scrollRegionBFullPaintEvidenceSchema, CellID: cellID, Events: []scrollRegionBFullPaint{paint, first}, Settled: final})
	swatch := color.RGBA{R: 31, G: 95, B: 176, A: 255}
	screenshotPath := writeScrollRegionEvidencePNG(t, directory, "cell.png", swatch)
	screenshot := scrollRegionBFullEvidenceArtifactForFile(t, screenshotPath)
	screenshot.Width, screenshot.Height, screenshot.CapturedRegion = 160, 100, "named-scroll-region"
	screenshot.Capture = &scrollRegionBFullCaptureProof{
		Method: "cdp-page-capture-screenshot", VisualViewportWidth: 160, VisualViewportHeight: 100, DevicePixelRatio: 1,
		SourceWidth: 160, SourceHeight: 100, ScaleX: 1, ScaleY: 1, CropCSSWidth: 160, CropCSSHeight: 100, CropPixelWidth: 160, CropPixelHeight: 100,
		Anchors:     []scrollRegionBFullPixelAnchor{{Name: "activity-card", DOMText: "Activity fixture", PixelX: 80, PixelY: 20, Expected: [4]uint8{71, 135, 216, 255}, Tolerance: 0}},
		BoundaryCue: scrollRegionBFullBoundaryCue{State: "end", Visible: true, PixelX: 150, PixelY: 90, PixelWidth: 4, PixelHeight: 4},
	}
	trace := scrollRegionBFullFixturePlaywrightTrace(t, directory, "cell.trace.zip", cellID)
	cell := scrollRegionBFullCellReceipt{
		CellID: cellID, Route: scrollRegionBFullRoute, Theme: "araihu", Mode: "light", ViewportWidth: 390, Zoom: "default",
		States: []string{"default", "no-overflow", "start", "middle", "end", "focused"}, Inputs: []string{"mouse", "keyboard", "cdp-touch"}, FirstHTMLSHA256: scrollRegionBFullSHA256(pageBHTML),
		SetupActions: []scrollRegionBFullAction{consentAction.Action, resetAction.Action, setAction.Action},
		Persistence: scrollRegionBFullPersistenceProof{
			ThemeInitialSource: scrollRegionBFullThemeInitialSource, ThemePersistenceNotApplicable: scrollRegionBFullThemePersistenceNA, DarkPersistence: scrollRegionBFullDarkPersistence,
			StorageBefore: map[string]string{"darkMode": "", "cookie": ""}, Actions: []scrollRegionBFullAction{consentAction.Action, resetAction.Action, setAction.Action},
			FreshLoadInitialHTML: paint, FreshLoadFirstPaint: first, FreshLoadSettled: final,
			PageAInitialHTML: pageAInitial, PageAStorageBefore: storageBefore, PageAActions: actionArtifacts, FreshLoadStorage: pageBStorage, FreshLoadInitialHTMLArtifact: pageBInitial, FreshLoadPaintArtifact: paintArtifact,
		},
		Screenshot: screenshot,
		Trace:      &trace,
		NAs:        map[string]string{"theme-persistence": scrollRegionBFullThemePersistenceNA},
	}
	return scrollRegionBFullReceipt{
		Schema: scrollRegionBFullReceiptSchema, Closure: "diagnostic-non-closure", ExpectedCells: 1,
		Binding: scrollRegionBFullIdentityBinding{Binding: "diagnostic-unbound-dirty-worktree"}, ToolVersions: scrollRegionBFullToolVersions{GoRuntime: "go1.26.5", PlaywrightGo: "v0.6100.0", BrowserByZoom: map[string]string{"default": "Chromium 140.0.0.0"}, AxeCore: scrollRegionAxeCoreVersion}, Widths: []int{390},
		Cells: []scrollRegionBFullCellReceipt{cell}, TraceByCell: map[string]scrollRegionBFullArtifact{cellID: trace},
	}
}

func scrollRegionBFullFixtureJSONArtifact(t *testing.T, directory, name string, value any) scrollRegionBFullArtifact {
	t.Helper()
	path := writeScrollRegionEvidenceJSON(t, directory, name, value)
	return scrollRegionBFullEvidenceArtifactForFile(t, path)
}

func scrollRegionBFullFixtureBytesArtifact(t *testing.T, directory, name string, content []byte) scrollRegionBFullArtifact {
	t.Helper()
	path := filepath.Join(directory, name)
	require.NoError(t, os.WriteFile(path, content, 0o600))
	return scrollRegionBFullEvidenceArtifactForFile(t, path)
}

// scrollRegionBFullFixturePlaywrightTrace creates the minimum structured
// Playwright archive used by validator tests. It deliberately writes the same
// trace.trace, trace.network, and screencast resource layout emitted by
// Playwright rather than a ZIP-magic placeholder.
func scrollRegionBFullFixturePlaywrightTrace(t *testing.T, directory, name, cellID string) scrollRegionBFullArtifact {
	t.Helper()
	path := filepath.Join(directory, name)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	archive := zip.NewWriter(file)
	write := func(entry string, content []byte) {
		writer, err := archive.Create(entry)
		require.NoError(t, err)
		_, err = writer.Write(content)
		require.NoError(t, err)
	}
	parts := strings.Split(cellID, "|")
	require.Len(t, parts, 5)
	theme, ok := scrollRegionBFullThemeByID(parts[1])
	require.True(t, ok)
	width, err := strconv.Atoi(parts[3])
	require.NoError(t, err)
	dark := parts[2] == "dark"
	route := scrollRegionBFullCellRoutedURL(theme, dark, width, scrollRegionBFullZoom{ID: parts[4]})
	pageA := "page@fixture-a"
	pageB := "page@fixture-b"
	traceLines := []string{
		`{"version":8,"type":"context-options","browserName":"chromium","playwrightVersion":"1.61.1","sdkLanguage":"javascript","contextId":"browser-context@fixture"}`,
		`{"type":"before","callId":"call@viewport-a","class":"Page","method":"setViewportSize","pageId":"` + pageA + `","params":{"viewportSize":{"width":` + strconv.Itoa(width) + `,"height":900}}}`,
		`{"type":"after","callId":"call@viewport-a"}`,
		`{"type":"before","callId":"call@goto-a","class":"Frame","method":"goto","pageId":"` + pageA + `","params":{"url":"` + route + `"}}`,
		`{"type":"after","callId":"call@goto-a","result":{"response":"<Response>"}}`,
		`{"type":"before","callId":"call@consent","class":"Frame","method":"click","pageId":"` + pageA + `","params":{"selector":"internal:role=button[name=\"Allow browser storage\"i]"}}`,
		`{"type":"after","callId":"call@consent"}`,
	}
	darkCalls := 2
	if dark {
		darkCalls = 1
	}
	for index := 0; index < darkCalls; index++ {
		callID := "call@dark-" + strconv.Itoa(index)
		traceLines = append(traceLines,
			`{"type":"before","callId":"`+callID+`","class":"Frame","method":"click","pageId":"`+pageA+`","params":{"selector":"#darkModeToggleBtn"}}`,
			`{"type":"after","callId":"`+callID+`"}`,
		)
	}
	traceLines = append(traceLines,
		`{"type":"before","callId":"call@viewport-b","class":"Page","method":"setViewportSize","pageId":"`+pageB+`","params":{"viewportSize":{"width":`+strconv.Itoa(width)+`,"height":900}}}`,
		`{"type":"after","callId":"call@viewport-b"}`,
		`{"type":"before","callId":"call@goto-b","class":"Frame","method":"goto","pageId":"`+pageB+`","params":{"url":"`+route+`"}}`,
		`{"type":"after","callId":"call@goto-b","result":{"response":"<Response>"}}`,
	)
	write("trace.trace", []byte(strings.Join(traceLines, "\n")+"\n"))
	write("trace.network", []byte(`{"type":"resource-snapshot","snapshot":{"_resourceType":"document","pageref":"`+pageA+`","request":{"url":"`+route+`"}}}`+"\n"+`{"type":"resource-snapshot","snapshot":{"_resourceType":"document","pageref":"`+pageB+`","request":{"url":"`+route+`"}}}`+"\n"))
	write("resources/page@fixture-1.jpeg", []byte{0xff, 0xd8, 0xff, 0xd9})
	write("resources/fixture-cell.txt", []byte(cellID))
	require.NoError(t, archive.Close())
	require.NoError(t, file.Close())
	return scrollRegionBFullEvidenceArtifactForFile(t, path)
}

func scrollRegionBFullEvidenceArtifactForFile(t *testing.T, path string) scrollRegionBFullArtifact {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return scrollRegionBFullArtifact{Path: path, SHA256: scrollRegionBFullSHA256(content)}
}

func scrollRegionATReceiptForFixture(t *testing.T, repository string, candidate scrollRegionCandidateIdentity) scrollRegionATEvidenceReceipt {
	t.Helper()
	if os.Getenv(scrollRegionATReplayRegistryEnvironment) == "" {
		// Fixture-only external custody. Production harness requires a caller
		// supplied owner-only registry and never creates one in the candidate.
		t.Setenv(scrollRegionATReplayRegistryEnvironment, t.TempDir())
	}
	identity := scrollRegionReceiptIdentityFromCandidate(candidate)
	directory := t.TempDir()
	capturedAt := time.Date(2026, time.August, 12, 14, 30, 0, 0, time.UTC).Format(time.RFC3339)
	challenge := scrollRegionATChallengeForFixture(t, repository)
	observations := scrollRegionATObservationsForFixture()
	return scrollRegionATEvidenceReceipt{
		Schema:    scrollRegionATReceiptSchema,
		Status:    "captured",
		Challenge: challenge.Challenge,
		Identity:  identity,
		Captures: []scrollRegionATEvidenceCapture{
			scrollRegionATEvidenceCaptureForFixture(t, repository, directory, identity, challenge.Challenge, "macos-safari-voiceover", "macOS 15.6", "Safari 18.6", "VoiceOver 15.6", capturedAt, observations, color.RGBA{R: 29, G: 78, B: 216, A: 255}),
			scrollRegionATEvidenceCaptureForFixture(t, repository, directory, identity, challenge.Challenge, "macos-chromium-voiceover", "macOS 15.6", "Chromium 140.0.7339.0", "VoiceOver 15.6", capturedAt, observations, color.RGBA{R: 5, G: 150, B: 105, A: 255}),
		},
	}
}

func scrollRegionATEvidenceCaptureForFixture(t *testing.T, repository, directory string, identity scrollRegionReceiptIdentity, challenge, pair, platformVersion, browserVersion, screenReaderVersion, capturedAt string, observations []scrollRegionATObservation, swatch color.RGBA) scrollRegionATEvidenceCapture {
	t.Helper()
	prefix := strings.ReplaceAll(pair, "-", "_")
	screenshot := writeScrollRegionEvidencePNG(t, directory, prefix+".png", swatch)
	capture := scrollRegionATEvidenceCapture{
		Pair:                pair,
		PlatformVersion:     platformVersion,
		BrowserVersion:      browserVersion,
		ScreenReaderVersion: screenReaderVersion,
		Route:               scrollRegionBFullRoute,
		CapturedAt:          capturedAt,
		Observations:        append([]scrollRegionATObservation(nil), observations...),
		Screenshot:          scrollRegionEvidenceArtifactForFile(t, screenshot),
	}
	scrollRegionWriteATFixtureRawProvenance(t, directory, prefix, &capture, identity, challenge)
	transcript := scrollRegionATTranscript{
		Schema:              scrollRegionATTranscriptSchema,
		Challenge:           challenge,
		Pair:                pair,
		CapturedAt:          capturedAt,
		Identity:            identity,
		PlatformVersion:     platformVersion,
		BrowserVersion:      browserVersion,
		ScreenReaderVersion: screenReaderVersion,
		Route:               scrollRegionBFullRoute,
		Observations:        capture.Observations,
	}
	transcriptPath := writeScrollRegionEvidenceJSON(t, directory, prefix+".transcript.json", transcript)
	capture.Transcript = scrollRegionEvidenceArtifactForFile(t, transcriptPath)
	trace := scrollRegionATTraceLog{
		Schema:     scrollRegionATTraceLogSchema,
		Challenge:  challenge,
		Pair:       pair,
		CapturedAt: capturedAt,
		Identity:   identity,
		Route:      scrollRegionBFullRoute,
		Events:     scrollRegionATTraceEventsForFixture(capture.Observations),
	}
	tracePath := writeScrollRegionEvidenceJSON(t, directory, prefix+".trace.json", trace)
	capture.TraceLog = scrollRegionEvidenceArtifactForFile(t, tracePath)
	payload := scrollRegionATAttestationPayload{
		Schema:    scrollRegionATAttestationPayloadSchema,
		Challenge: challenge,
		Identity:  identity,
		Capture:   scrollRegionATAttestedCaptureFromEvidence(capture),
	}
	attestationPath := writeScrollRegionATFixtureAttestation(t, repository, directory, prefix+".attestation.json", pair, payload)
	capture.Attestation = scrollRegionEvidenceArtifactForFile(t, attestationPath)
	return capture
}

func scrollRegionWriteATFixtureRawProvenance(t *testing.T, directory, prefix string, capture *scrollRegionATEvidenceCapture, identity scrollRegionReceiptIdentity, challenge string) {
	t.Helper()
	defaultToken := scrollRegionATActionToken(challenge, "default")
	servedHTML := []byte(`<!doctype html><html><head><meta name="goshtoso-t-gs-011-at-challenge" content="` + challenge + `"><meta name="goshtoso-t-gs-011-candidate-tree" content="` + identity.CandidateTree + `"><meta name="goshtoso-t-gs-011-manifest-sha256" content="` + identity.ManifestSHA256 + `"><meta name="goshtoso-t-gs-011-at-pair" content="` + capture.Pair + `"><meta name="goshtoso-t-gs-011-at-action-state" content="default"><meta name="goshtoso-t-gs-011-at-action-token" content="` + defaultToken + `"></head><body><div role="region" data-goshtoso-scroll-viewport aria-label="Activity history"></div></body></html>`)
	servedPath := filepath.Join(directory, prefix+".served-page.html")
	require.NoError(t, os.WriteFile(servedPath, servedHTML, 0o600))
	capture.ServedPage = scrollRegionEvidenceArtifactForFile(t, servedPath)
	response := scrollRegionATServedPage{
		Schema: scrollRegionATServedPageSchema, URL: "http://127.0.0.1:44111" + scrollRegionBFullRoute + "?t-gs-011-at-capture=" + challenge + "&t-gs-011-at-state=default&t-gs-011-at-action-token=" + defaultToken,
		Status: 200, Challenge: challenge, CandidateTree: identity.CandidateTree, ManifestSHA256: identity.ManifestSHA256,
		BodySHA256: scrollRegionBFullSHA256(servedHTML), Pair: capture.Pair, ActionState: "default", ActionToken: defaultToken,
	}
	responsePath := writeScrollRegionEvidenceJSON(t, directory, prefix+".served-response.json", response)
	capture.ServedResponse = scrollRegionEvidenceArtifactForFile(t, responsePath)
	window := scrollRegionATBrowserWindow{
		Schema: scrollRegionATBrowserStateSchema, Pair: capture.Pair, Route: capture.Route, Challenge: challenge,
		CandidateTree: identity.CandidateTree, ManifestSHA256: identity.ManifestSHA256,
		Window: scrollRegionATRectangle{X: 40, Y: 50, Width: 160, Height: 100}, CandidateRegion: scrollRegionATRectangle{X: 10, Y: 10, Width: 140, Height: 80},
	}
	windowPath := writeScrollRegionEvidenceJSON(t, directory, prefix+".browser-window.json", window)
	capture.BrowserWindow = scrollRegionEvidenceArtifactForFile(t, windowPath)
	capture.BrowserStates = make([]scrollRegionATEvidenceArtifact, 0, len(capture.Observations)*3)
	capture.ActionRecords = make([]scrollRegionATEvidenceArtifact, 0, len(capture.Observations))
	capture.VoiceOverCaptions = make([]scrollRegionATEvidenceArtifact, 0, len(capture.Observations))
	for index := range capture.Observations {
		observation := capture.Observations[index]
		token := scrollRegionATActionToken(challenge, observation.State)
		started := time.Date(2026, time.August, 12, 16, 0, index*3, 0, time.UTC)
		beforeAt := started.Add(time.Millisecond)
		logStartedAt := started.Add(2 * time.Millisecond)
		actionIssuedAt := started.Add(3 * time.Millisecond)
		afterAt := started.Add(4 * time.Millisecond)
		exitIssuedAt := started.Add(5 * time.Millisecond)
		exitAt := started.Add(6 * time.Millisecond)
		logEndedAt := started.Add(7 * time.Millisecond)
		expected := scrollRegionATExpectedStates[index]
		action := scrollRegionATActionRecord{
			Schema: scrollRegionATActionRecordSchema, Pair: capture.Pair, State: observation.State, Command: observation.Commands[0], Route: capture.Route,
			Challenge: challenge, CandidateTree: identity.CandidateTree, ManifestSHA256: identity.ManifestSHA256, ActionToken: token,
			ActionEvent: expected.ActionEvent, ExitCommand: expected.ExitCommand, ExitEvent: expected.ExitEvent, VoiceOverPID: 4242, VoiceOverSubsystem: scrollRegionATVoiceOverSubsystem,
			BeforeAt: beforeAt.Format(time.RFC3339Nano), LogStartedAt: logStartedAt.Format(time.RFC3339Nano), ActionIssuedAt: actionIssuedAt.Format(time.RFC3339Nano),
			AfterAt: afterAt.Format(time.RFC3339Nano), ExitIssuedAt: exitIssuedAt.Format(time.RFC3339Nano), ExitAt: exitAt.Format(time.RFC3339Nano), LogEndedAt: logEndedAt.Format(time.RFC3339Nano),
		}
		actionPath := writeScrollRegionEvidenceJSON(t, directory, prefix+".action-"+observation.State+".json", action)
		capture.ActionRecords = append(capture.ActionRecords, scrollRegionEvidenceArtifactForFile(t, actionPath))
		before, after, exit := scrollRegionATFixtureSnapshots(observation)
		phases := []struct {
			name     string
			snapshot scrollRegionATBrowserSnapshot
		}{{"before", before}, {"after", after}, {"exit", exit}}
		for phaseIndex, phase := range phases {
			observedAt := []time.Time{beforeAt, afterAt, exitAt}[phaseIndex]
			browser := scrollRegionATBrowserState{
				Schema: scrollRegionATBrowserStateSchema, Pair: capture.Pair, State: observation.State, Command: observation.Commands[0], Route: capture.Route,
				Challenge: challenge, CandidateTree: identity.CandidateTree, ManifestSHA256: identity.ManifestSHA256,
				ActionToken: token, Phase: phase.name, ObservedAt: observedAt.Format(time.RFC3339Nano),
				Snapshot: phase.snapshot, CandidateRegion: scrollRegionATRectangle{X: 80, Y: 140, Width: 520, Height: 280},
			}
			browserPath := writeScrollRegionEvidenceJSON(t, directory, prefix+".browser-"+observation.State+"-"+phase.name+".json", browser)
			capture.BrowserStates = append(capture.BrowserStates, scrollRegionEvidenceArtifactForFile(t, browserPath))
		}
		capture.Observations[index].FocusNavigation = scrollRegionATFocusNavigation{
			Before: scrollRegionATSnapshotDescription(before), Entry: scrollRegionATSnapshotDescription(after), Exit: scrollRegionATSnapshotDescription(exit),
		}
		observation = capture.Observations[index]
		reported := strings.Join(observation.ObservedSpeech, " ")
		speech := "VoiceOver pair " + capture.Pair + " named region " + observation.Name + " announced after " + observation.Commands[0] + " state " + observation.State
		if strings.TrimSpace(reported) != "" {
			speech += " reported " + reported
		}
		capture.Observations[index].ObservedSpeech = []string{speech}
		voicePath := filepath.Join(directory, prefix+".voiceover-"+observation.State+".json")
		voiceEvents := []scrollRegionATVoiceOverLogEvent{{Timestamp: afterAt.Add(time.Nanosecond).Format(time.RFC3339Nano), ProcessID: 4242, Subsystem: scrollRegionATVoiceOverSubsystem, EventMessage: speech}}
		_ = writeScrollRegionEvidenceJSON(t, directory, filepath.Base(voicePath), voiceEvents)
		capture.VoiceOverCaptions = append(capture.VoiceOverCaptions, scrollRegionEvidenceArtifactForFile(t, voicePath))
	}
	logPath := filepath.Join(directory, prefix+".voiceover-system-log.json")
	combined := make([]scrollRegionATVoiceOverLogEvent, 0, len(capture.Observations))
	for index, observation := range capture.Observations {
		at := time.Date(2026, time.August, 12, 16, 0, index*3, 0, time.UTC).Add(4*time.Millisecond + time.Nanosecond)
		combined = append(combined, scrollRegionATVoiceOverLogEvent{Timestamp: at.Format(time.RFC3339Nano), ProcessID: 4242, Subsystem: scrollRegionATVoiceOverSubsystem, EventMessage: "VoiceOver pair " + capture.Pair + " named region " + observation.Name + " announced state " + observation.State})
	}
	_ = writeScrollRegionEvidenceJSON(t, directory, filepath.Base(logPath), combined)
	capture.VoiceOverLog = scrollRegionEvidenceArtifactForFile(t, logPath)
}

func scrollRegionATActionToken(challenge, state string) string {
	return "goshtoso-t-gs-011-" + challenge + "-" + state
}

func scrollRegionATFixtureSnapshots(observation scrollRegionATObservation) (scrollRegionATBrowserSnapshot, scrollRegionATBrowserSnapshot, scrollRegionATBrowserSnapshot) {
	clientHeight, scrollHeight := 140.0, 280.0
	if observation.State == "no-overflow" {
		scrollHeight = clientHeight
	}
	before := scrollRegionATBrowserSnapshot{ActiveRole: "button", ActiveName: "Previous document control", RegionRole: "region", RegionName: observation.Name, Boundary: "start", ClientHeight: clientHeight, ScrollHeight: scrollHeight}
	switch observation.State {
	case "start":
		before.Boundary, before.ScrollTop, before.StartCueVisible, before.EndCueVisible = "middle", 70, true, true
	case "end":
		before.Boundary, before.ScrollTop, before.StartCueVisible, before.EndCueVisible = "middle", 70, true, true
	case "no-overflow":
		before.Boundary = "no-overflow"
	}
	after := scrollRegionATBrowserSnapshot{ActiveRole: "body", ActiveName: "Document body", RegionRole: "region", RegionName: observation.Name, RegionFocused: observation.Focused, Boundary: observation.Boundary, ClientHeight: clientHeight, ScrollHeight: scrollHeight}
	if observation.Focused {
		after.ActiveRole, after.ActiveName = "region", observation.Name
	}
	if observation.Boundary == "middle" {
		after.ScrollTop, after.StartCueVisible, after.EndCueVisible = 70, true, true
	}
	if observation.Boundary == "end" {
		after.ScrollTop, after.StartCueVisible = 140, true
	}
	exit := after
	exit.ActiveRole, exit.ActiveName, exit.RegionFocused = "button", "Next document control", false
	return before, after, exit
}

func scrollRegionATSnapshotDescription(snapshot scrollRegionATBrowserSnapshot) string {
	return strings.TrimSpace(snapshot.ActiveRole + " " + snapshot.ActiveName)
}

func scrollRegionATObservationsForFixture() []scrollRegionATObservation {
	result := make([]scrollRegionATObservation, 0, len(scrollRegionATExpectedStates))
	for _, expected := range scrollRegionATExpectedStates {
		name := "Activity history"
		if expected.Name == "no-overflow" {
			name = "Current release"
		}
		result = append(result, scrollRegionATObservation{
			State:    expected.Name,
			Role:     "region",
			Name:     name,
			Focused:  expected.Focused,
			Boundary: expected.Boundary,
			Commands: []string{expected.RequiredCommand},
			FocusNavigation: scrollRegionATFocusNavigation{
				Before: "Main content heading before " + name,
				Entry:  "Named region " + name + " receives focus",
				Exit:   "Next document control follows " + name + " region",
			},
			ObservedSpeech:   []string{"VoiceOver named region " + name + " announced at " + expected.Boundary + " boundary state " + expected.Name},
			UnexpectedSpeech: []string{},
		})
	}
	return result
}

func scrollRegionATTraceEventsForFixture(observations []scrollRegionATObservation) []scrollRegionATTraceEvent {
	events := make([]scrollRegionATTraceEvent, 0, len(observations))
	for _, observation := range observations {
		events = append(events, scrollRegionATTraceEvent{
			State:          observation.State,
			Command:        observation.Commands[0],
			BeforeBoundary: "before-" + observation.Boundary,
			AfterBoundary:  observation.Boundary,
			Focused:        observation.Focused,
		})
	}
	return events
}

func scrollRegionATChallengePathForFixture(repository string) string {
	return filepath.Join(repository, "at-capture-challenge.json")
}

func scrollRegionATChallengeForFixture(t *testing.T, repository string) scrollRegionATChallenge {
	t.Helper()
	content, err := os.ReadFile(scrollRegionATChallengePathForFixture(repository))
	require.NoError(t, err)
	var challenge scrollRegionATChallenge
	require.NoError(t, scrollRegionDecodeStrictJSON(content, &challenge))
	return challenge
}

func writeScrollRegionATFixtureAttestation(t *testing.T, repository, directory, name, pair string, payload scrollRegionATAttestationPayload) string {
	t.Helper()
	keyID := "fixture-macos-voiceover"
	if pair == "windows-chromium-nvda" {
		keyID = "fixture-windows-nvda"
	}
	return writeScrollRegionATFixtureAttestationWithKey(t, repository, directory, name, keyID, payload)
}

func writeScrollRegionATFixtureAttestationWithKey(t *testing.T, repository, directory, name, keyID string, payload scrollRegionATAttestationPayload) string {
	t.Helper()
	private := scrollRegionATFixturePrivateKey(t, repository, keyID)
	rawPayload, err := json.Marshal(payload)
	require.NoError(t, err)
	signature := ed25519.Sign(private, scrollRegionATDSSEPAE(scrollRegionATAttestationPayloadType, rawPayload))
	envelope := scrollRegionATDSSEEnvelope{
		Schema:      scrollRegionATAttestationSchema,
		PayloadType: scrollRegionATAttestationPayloadType,
		Payload:     base64.StdEncoding.EncodeToString(rawPayload),
		Signatures: []scrollRegionATDSSESignature{{
			KeyID:     keyID,
			Signature: base64.StdEncoding.EncodeToString(signature),
		}},
	}
	return writeScrollRegionEvidenceJSON(t, directory, name, envelope)
}

func scrollRegionATFixturePrivateKey(t *testing.T, repository, keyID string) ed25519.PrivateKey {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(repository, "fixture-private-keys", keyID+".pem"))
	require.NoError(t, err)
	block, remainder := pem.Decode(content)
	require.NotNil(t, block)
	require.Empty(t, bytes.TrimSpace(remainder))
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	require.NoError(t, err)
	private, ok := parsed.(ed25519.PrivateKey)
	require.True(t, ok)
	return private
}

func writeScrollRegionEvidencePNG(t *testing.T, directory, name string, swatch color.RGBA) string {
	t.Helper()
	canvas := image.NewRGBA(image.Rect(0, 0, 160, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 160; x++ {
			pixel := swatch
			switch {
			case x < 4 || x >= 156 || y < 4 || y >= 96:
				pixel = color.RGBA{R: swatch.R / 3, G: swatch.G / 3, B: swatch.B / 3, A: 255}
			case (y >= 14 && y < 40) || (y >= 54 && y < 80):
				pixel = color.RGBA{R: uint8(min(255, int(swatch.R)+40)), G: uint8(min(255, int(swatch.G)+40)), B: uint8(min(255, int(swatch.B)+40)), A: 255}
			case x >= 18 && x < 24:
				pixel = color.RGBA{R: 24, G: 24, B: 24, A: 255}
			case x >= 32 && x < 38:
				pixel = color.RGBA{R: 58, G: 58, B: 58, A: 255}
			case x >= 48 && x < 54:
				pixel = color.RGBA{R: 92, G: 92, B: 92, A: 255}
			}
			canvas.SetRGBA(x, y, pixel)
		}
	}
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, canvas))
	path := filepath.Join(directory, name)
	require.NoError(t, os.WriteFile(path, encoded.Bytes(), 0o600))
	return path
}

func writeScrollRegionEvidenceSolidPNG(t *testing.T, directory, name string, swatch color.RGBA) string {
	t.Helper()
	canvas := image.NewRGBA(image.Rect(0, 0, 160, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 160; x++ {
			canvas.SetRGBA(x, y, swatch)
		}
	}
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, canvas))
	path := filepath.Join(directory, name)
	require.NoError(t, os.WriteFile(path, encoded.Bytes(), 0o600))
	return path
}

func scrollRegionMutateATFixtureAttestation(t *testing.T, receipt *scrollRegionATEvidenceReceipt, index int, mutate func(*scrollRegionATDSSEEnvelope)) {
	t.Helper()
	capture := &receipt.Captures[index]
	content, err := os.ReadFile(capture.Attestation.Path)
	require.NoError(t, err)
	var envelope scrollRegionATDSSEEnvelope
	require.NoError(t, scrollRegionDecodeStrictJSON(content, &envelope))
	mutate(&envelope)
	path := writeScrollRegionEvidenceJSON(t, filepath.Dir(capture.Attestation.Path), filepath.Base(capture.Attestation.Path), envelope)
	capture.Attestation = scrollRegionEvidenceArtifactForFile(t, path)
}

func scrollRegionRefreshATFixtureCapture(t *testing.T, repository string, receipt *scrollRegionATEvidenceReceipt, index int, synchronizeTranscriptAndTrace bool) {
	t.Helper()
	capture := &receipt.Captures[index]
	if synchronizeTranscriptAndTrace {
		scrollRegionWriteATFixtureRawProvenance(t, filepath.Dir(capture.TraceLog.Path), strings.ReplaceAll(capture.Pair, "-", "_"), capture, receipt.Identity, receipt.Challenge)
		transcript := scrollRegionATTranscript{
			Schema:              scrollRegionATTranscriptSchema,
			Challenge:           receipt.Challenge,
			Pair:                capture.Pair,
			CapturedAt:          capture.CapturedAt,
			Identity:            receipt.Identity,
			PlatformVersion:     capture.PlatformVersion,
			BrowserVersion:      capture.BrowserVersion,
			ScreenReaderVersion: capture.ScreenReaderVersion,
			Route:               capture.Route,
			Observations:        capture.Observations,
		}
		transcriptPath := writeScrollRegionEvidenceJSON(t, filepath.Dir(capture.Transcript.Path), filepath.Base(capture.Transcript.Path), transcript)
		capture.Transcript = scrollRegionEvidenceArtifactForFile(t, transcriptPath)
		trace := scrollRegionATTraceLog{
			Schema:     scrollRegionATTraceLogSchema,
			Challenge:  receipt.Challenge,
			Pair:       capture.Pair,
			CapturedAt: capture.CapturedAt,
			Identity:   receipt.Identity,
			Route:      capture.Route,
			Events:     scrollRegionATTraceEventsForFixture(capture.Observations),
		}
		tracePath := writeScrollRegionEvidenceJSON(t, filepath.Dir(capture.TraceLog.Path), filepath.Base(capture.TraceLog.Path), trace)
		capture.TraceLog = scrollRegionEvidenceArtifactForFile(t, tracePath)
	}
	payload := scrollRegionATAttestationPayload{
		Schema:    scrollRegionATAttestationPayloadSchema,
		Challenge: receipt.Challenge,
		Identity:  receipt.Identity,
		Capture:   scrollRegionATAttestedCaptureFromEvidence(*capture),
	}
	attestationPath := writeScrollRegionATFixtureAttestation(t, repository, filepath.Dir(capture.Attestation.Path), filepath.Base(capture.Attestation.Path), capture.Pair, payload)
	capture.Attestation = scrollRegionEvidenceArtifactForFile(t, attestationPath)
}

func scrollRegionEvidenceArtifactForFile(t *testing.T, path string) scrollRegionATEvidenceArtifact {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return scrollRegionATEvidenceArtifact{Path: path, SHA256: scrollRegionBFullSHA256(content)}
}

func newScrollRegionEvidenceFixtureRepository(t *testing.T) string {
	t.Helper()
	repository := t.TempDir()
	for path, content := range map[string]string{
		"go.mod":                "module example.test/scrollregion\n\ngo 1.26.5\n\nrequire github.com/a-h/templ v0.3.1020\n",
		"site/go.mod":           "module example.test/scrollregion/site\n\ngo 1.26.5\n\nrequire (\n\tgithub.com/a-h/templ v0.3.1020\n\tgithub.com/mxschmitt/playwright-go v0.6100.0\n)\n",
		"scripts/axe-core.lock": "version=4.10.3\nurl=https://registry.npmjs.org/axe-core/-/axe-core-4.10.3.tgz\narchive_sha256=0f2b4d7dcdf7d1219df8d1959ad68e565f51d14c3f0d88bb71cd59abeb956292\nscript_path=package/axe.min.js\nscript_sha256=880970c081707360e64f34cea25ff91892f5bc95675b0776925b9709dd8a68bb\n",
		"tracked.txt":           "base\n",
	} {
		path = filepath.Join(repository, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	}
	macOSKey := writeScrollRegionEvidenceFixtureATKey(t, repository, "fixture-macos-voiceover", []string{"macos-safari-voiceover", "macos-chromium-voiceover"})
	windowsKey := writeScrollRegionEvidenceFixtureATKey(t, repository, "fixture-windows-nvda", []string{"windows-chromium-nvda"})
	manifest := scrollRegionATTrustedKeyManifest{
		Schema: scrollRegionATTrustedKeysSchema,
		Keys:   []scrollRegionATTrustedKeyRef{macOSKey, windowsKey},
	}
	writeScrollRegionEvidenceJSON(t, filepath.Join(repository, "tests", "external", "scrollregion-a11y"), "attestation-keys.json", manifest)
	challenge := scrollRegionATChallenge{
		Schema:    scrollRegionATChallengeSchema,
		Challenge: "95e2b8221dd9a85c1e3b7ee4b2c82568c2b29d4f181efd33d631820eac381c21",
		IssuedAt:  time.Date(2026, time.August, 12, 14, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}
	writeScrollRegionEvidenceJSON(t, repository, "at-capture-challenge.json", challenge)
	runScrollRegionEvidenceGit(t, repository, "init", "-q")
	runScrollRegionEvidenceGit(t, repository, "config", "user.email", "evidence@example.test")
	runScrollRegionEvidenceGit(t, repository, "config", "user.name", "Evidence Fixture")
	runScrollRegionEvidenceGit(t, repository, "remote", "add", "origin", "https://github.com/araihu/goshtoso.git")
	runScrollRegionEvidenceGit(t, repository, "add", ".")
	runScrollRegionEvidenceGit(t, repository, "commit", "-qm", "base")
	require.NoError(t, os.WriteFile(filepath.Join(repository, "evidence.txt"), []byte("candidate\n"), 0o600))
	return repository
}

// Fixture keys exist only below t.TempDir. They prove the validator's
// cryptographic behavior without copying a capture authority private key into
// the candidate source, fixtures, receipts, or test output.
func writeScrollRegionEvidenceFixtureATKey(t *testing.T, repository, keyID string, pairs []string) scrollRegionATTrustedKeyRef {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	publicDER, err := x509.MarshalPKIXPublicKey(public)
	require.NoError(t, err)
	privateDER, err := x509.MarshalPKCS8PrivateKey(private)
	require.NoError(t, err)
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER})
	publicPath := filepath.ToSlash(filepath.Join("tests", "external", "scrollregion-a11y", "attestation-keys", keyID+".pem"))
	fullPublicPath := filepath.Join(repository, filepath.FromSlash(publicPath))
	fullPrivatePath := filepath.Join(repository, "fixture-private-keys", keyID+".pem")
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPublicPath), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPrivatePath), 0o755))
	require.NoError(t, os.WriteFile(fullPublicPath, publicPEM, 0o600))
	require.NoError(t, os.WriteFile(fullPrivatePath, privatePEM, 0o600))
	return scrollRegionATTrustedKeyRef{
		KeyID:           keyID,
		PublicKeyPath:   publicPath,
		PublicKeySHA256: scrollRegionBFullSHA256(publicPEM),
		Pairs:           pairs,
	}
}

func scrollRegionCandidateIdentityForFixture(t *testing.T, repository string) scrollRegionCandidateIdentity {
	t.Helper()
	head := strings.TrimSpace(runScrollRegionEvidenceGit(t, repository, "rev-parse", "HEAD"))
	tree := strings.TrimSpace(runScrollRegionEvidenceGit(t, repository, "rev-parse", "HEAD^{tree}"))
	status := runScrollRegionEvidenceGit(t, repository, "status", "--porcelain=v1", "--untracked-files=all")
	candidateTree := scrollRegionEvidenceCandidateTree(t, repository)
	content, err := os.ReadFile(filepath.Join(repository, "evidence.txt"))
	require.NoError(t, err)
	digest := sha256.Sum256(content)
	paths := []scrollRegionCandidatePath{{Path: "evidence.txt", SHA256: hex.EncodeToString(digest[:])}}
	return scrollRegionCandidateIdentity{
		Schema:         scrollRegionCandidateIdentitySchema,
		RepositoryURL:  "https://github.com/araihu/goshtoso.git",
		Head:           head,
		Tree:           tree,
		CandidateTree:  candidateTree,
		ManifestSHA256: scrollRegionCandidateManifestSHA256(paths),
		StatusSHA256:   scrollRegionBFullSHA256([]byte(status)),
		Paths:          paths,
		DependencyPins: scrollRegionDependencyPins{
			RootGoDirective:  "1.26.5",
			SiteGoDirective:  "1.26.5",
			Templ:            "v0.3.1020",
			PlaywrightGo:     "v0.6100.0",
			AxeCore:          scrollRegionAxeCoreVersion,
			AxeArchiveSHA256: scrollRegionAxeArchiveSHA256,
			AxeScriptSHA256:  scrollRegionAxeScriptSHA256,
		},
	}
}

func scrollRegionRefreshFixtureCandidateIdentity(t *testing.T, repository string, identity *scrollRegionCandidateIdentity) {
	t.Helper()
	status := runScrollRegionEvidenceGit(t, repository, "status", "--porcelain=v1", "--untracked-files=all")
	identity.StatusSHA256 = scrollRegionBFullSHA256([]byte(status))
	identity.CandidateTree = scrollRegionEvidenceCandidateTree(t, repository)
}

func scrollRegionEvidenceCandidateTree(t *testing.T, repository string) string {
	t.Helper()
	index := filepath.Join(t.TempDir(), "candidate.index")
	environment := append(os.Environ(), "GIT_INDEX_FILE="+index)
	runScrollRegionEvidenceGitEnv(t, repository, environment, "read-tree", "HEAD")
	runScrollRegionEvidenceGitEnv(t, repository, environment, "add", "-A")
	return strings.TrimSpace(runScrollRegionEvidenceGitEnv(t, repository, environment, "write-tree"))
}

func writeScrollRegionEvidenceJSON(t *testing.T, directory, name string, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	path := filepath.Join(directory, name)
	require.NoError(t, os.WriteFile(path, encoded, 0o600))
	return path
}

func runScrollRegionEvidenceGit(t *testing.T, repository string, arguments ...string) string {
	t.Helper()
	return runScrollRegionEvidenceGitEnv(t, repository, os.Environ(), arguments...)
}

func runScrollRegionEvidenceGitEnv(t *testing.T, repository string, environment []string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", arguments...)
	command.Dir = repository
	command.Env = environment
	output, err := command.CombinedOutput()
	require.NoErrorf(t, err, "git %s: %s", strings.Join(arguments, " "), output)
	return string(output)
}
