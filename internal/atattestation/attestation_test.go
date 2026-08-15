package atattestation

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestVerifyRejectsUnsignedAndMutatedClaims(t *testing.T) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	claims := Claims{SourceCommit: "commit", SourceTree: "tree", Route: "/public", State: "initial", Browser: "Safari 1", ScreenReader: "VoiceOver 1", Pair: "safari-voiceover", CapturedAt: "2026-08-15T00:00:00Z", Challenge: "challenge", ActionToken: "token", ActionCommand: "Tab", ActionSequence: []string{"Tab"}, Recorder: "test", ServedSHA256: "served", ScreenshotSHA256: "screen", TraceSHA256: "trace"}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	envelope := Envelope{PayloadType: PayloadType, Payload: base64.StdEncoding.EncodeToString(payload), Signatures: []Signature{{KeyID: "test", Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, pae(PayloadType, payload)))}}}
	raw, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	if err := Verify(raw, map[string]ed25519.PublicKey{"test": public}, claims); err != nil {
		t.Fatal(err)
	}
	wrongPair := claims
	wrongPair.Pair = "chromium-screen-reader"
	if err := Verify(raw, map[string]ed25519.PublicKey{"test": public}, wrongPair); err == nil {
		t.Fatal("wrong signed AT pair accepted")
	}
	wrongAction := claims
	wrongAction.Challenge = "other-challenge"
	if err := Verify(raw, map[string]ed25519.PublicKey{"test": public}, wrongAction); err == nil {
		t.Fatal("wrong expected AT challenge accepted")
	}
	wrongAction = claims
	wrongAction.ActionCommand = "End"
	if err := Verify(raw, map[string]ed25519.PublicKey{"test": public}, wrongAction); err == nil {
		t.Fatal("wrong expected AT action command accepted")
	}
	wrongSigner := claims
	wrongSigner.Recorder = "other-recorder"
	wrongPayload, err := json.Marshal(wrongSigner)
	if err != nil {
		t.Fatal(err)
	}
	wrongRaw, err := json.Marshal(Envelope{PayloadType: PayloadType, Payload: base64.StdEncoding.EncodeToString(wrongPayload), Signatures: []Signature{{KeyID: "test", Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, pae(PayloadType, wrongPayload)))}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := Verify(wrongRaw, map[string]ed25519.PublicKey{"test": public}, wrongSigner); err == nil {
		t.Fatal("signed wrong recorder key accepted")
	}
	envelope.Payload = base64.StdEncoding.EncodeToString([]byte(`{"route":"/forged"}`))
	raw, _ = json.Marshal(envelope)
	if err := Verify(raw, map[string]ed25519.PublicKey{"test": public}, claims); err == nil {
		t.Fatal("mutated signed claims accepted")
	}
}

func TestVerifyRejectsMissingPairChallengeAndCaptureTime(t *testing.T) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	claims := Claims{SourceCommit: "commit", SourceTree: "tree", Route: "/public", State: "initial", Browser: "Safari 1", ScreenReader: "VoiceOver 1", CapturedAt: "2026-08-15T00:00:00Z", ActionToken: "token", ActionSequence: []string{"Tab"}, Recorder: "recorder", ServedSHA256: "served", ScreenshotSHA256: "screen", TraceSHA256: "trace"}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(Envelope{PayloadType: PayloadType, Payload: base64.StdEncoding.EncodeToString(payload), Signatures: []Signature{{KeyID: "recorder", Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, pae(PayloadType, payload)))}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := Verify(raw, map[string]ed25519.PublicKey{"recorder": public}, claims); err == nil || !strings.Contains(err.Error(), "pair") {
		t.Fatalf("missing AT provenance error = %v", err)
	}
}
