package atattestation

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestVerifyRejectsUnsignedAndMutatedClaims(t *testing.T) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	claims := Claims{SourceCommit: "commit", SourceTree: "tree", Route: "/public", State: "initial", Browser: "Safari 1", ScreenReader: "VoiceOver 1", CapturedAt: "2026-08-15T00:00:00Z", ActionToken: "token", ActionSequence: []string{"Tab"}, Recorder: "recorder", ServedSHA256: "served", ScreenshotSHA256: "screen", TraceSHA256: "trace"}
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
	envelope.Payload = base64.StdEncoding.EncodeToString([]byte(`{"route":"/forged"}`))
	raw, _ = json.Marshal(envelope)
	if err := Verify(raw, map[string]ed25519.PublicKey{"test": public}, claims); err == nil {
		t.Fatal("mutated signed claims accepted")
	}
}
