// Package atattestation verifies source-pinned DSSE envelopes for raw AT
// capture bundles. It deliberately does not accept claimant-generated hashes
// or artifact metadata as an attestation authority.
package atattestation

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"
)

const PayloadType = "application/vnd.araihu.goshtoso.at-capture.v1+json"

type Envelope struct {
	PayloadType string      `json:"payloadType"`
	Payload     string      `json:"payload"`
	Signatures  []Signature `json:"signatures"`
}

type Signature struct {
	KeyID     string `json:"keyid"`
	Signature string `json:"sig"`
}

// Claims bind independently recorded raw capture bytes to one served target
// and action. Artifact hashes are part of signed payload, never outer claims.
type Claims struct {
	SourceCommit     string   `json:"source_commit"`
	SourceTree       string   `json:"source_tree"`
	Route            string   `json:"route"`
	State            string   `json:"state"`
	Browser          string   `json:"browser"`
	ScreenReader     string   `json:"screen_reader"`
	Pair             string   `json:"pair"`
	CapturedAt       string   `json:"captured_at"`
	Challenge        string   `json:"challenge"`
	ActionToken      string   `json:"action_token"`
	ActionCommand    string   `json:"action_command"`
	ActionSequence   []string `json:"action_sequence"`
	Recorder         string   `json:"recorder"`
	ServedSHA256     string   `json:"served_sha256"`
	ScreenshotSHA256 string   `json:"screenshot_sha256"`
	TraceSHA256      string   `json:"trace_sha256"`
}

func Verify(raw []byte, trusted map[string]ed25519.PublicKey, want Claims) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var envelope Envelope
	if err := decoder.Decode(&envelope); err != nil {
		return fmt.Errorf("decode signed AT envelope: %w", err)
	}
	if envelope.PayloadType != PayloadType || len(envelope.Signatures) != 1 {
		return fmt.Errorf("signed AT envelope shape is invalid")
	}
	payload, err := base64.StdEncoding.DecodeString(envelope.Payload)
	if err != nil {
		return fmt.Errorf("signed AT payload is not base64")
	}
	signature, err := base64.StdEncoding.DecodeString(envelope.Signatures[0].Signature)
	if err != nil {
		return fmt.Errorf("signed AT signature is not base64")
	}
	key := trusted[envelope.Signatures[0].KeyID]
	if len(key) != ed25519.PublicKeySize || !ed25519.Verify(key, pae(envelope.PayloadType, payload), signature) {
		return fmt.Errorf("signed AT envelope is not verified by a trusted recorder key")
	}
	var claims Claims
	decoder = json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&claims); err != nil {
		return fmt.Errorf("decode signed AT claims: %w", err)
	}
	if claims.SourceCommit != want.SourceCommit || claims.SourceTree != want.SourceTree || claims.Route != want.Route || claims.State != want.State || claims.Browser != want.Browser || claims.ScreenReader != want.ScreenReader || claims.Pair != want.Pair || claims.Recorder != envelope.Signatures[0].KeyID || claims.Recorder != want.Recorder || claims.ServedSHA256 != want.ServedSHA256 || claims.ScreenshotSHA256 != want.ScreenshotSHA256 || claims.TraceSHA256 != want.TraceSHA256 || claims.Challenge != want.Challenge || claims.ActionToken != want.ActionToken || claims.ActionCommand != want.ActionCommand || !slices.Equal(claims.ActionSequence, want.ActionSequence) || strings.TrimSpace(claims.Challenge) == "" || strings.TrimSpace(claims.ActionToken) == "" || strings.TrimSpace(claims.ActionCommand) == "" || len(claims.ActionSequence) == 0 {
		return fmt.Errorf("signed AT claims do not bind exact pair/key/recorder/challenge/action/tool/raw artifacts")
	}
	if _, err := time.Parse(time.RFC3339, claims.CapturedAt); err != nil {
		return fmt.Errorf("signed AT claims capture time is invalid")
	}
	for _, action := range claims.ActionSequence {
		if strings.TrimSpace(action) == "" {
			return fmt.Errorf("signed AT claims action sequence contains an empty action")
		}
	}
	return nil
}

func pae(kind string, payload []byte) []byte {
	return fmt.Appendf(nil, "DSSEv1 %d %s %d %s", len(kind), kind, len(payload), payload)
}
