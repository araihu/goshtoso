// Command scrollregionatreceipt assembles two directly generated capture
// bundles. It never accepts caller-selected screenshot, transcript, trace, or
// attestation artifact paths; final verification remains fail-closed in the
// T-GS-011 browser harness.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	receiptSchema       = "goshtoso.t-gs-011.at-receipt.v3"
	bundleSchema        = "goshtoso.t-gs-011.at-capture-bundle.v1"
	challengeSchema     = "goshtoso.t-gs-011.at-challenge.v1"
	challengeHexPattern = `^[0-9a-f]{64}$`
)

type captureChallenge struct {
	Schema    string `json:"schema"`
	Challenge string `json:"challenge"`
	IssuedAt  string `json:"issued_at"`
}

type captureBundle struct {
	Schema    string           `json:"schema"`
	Challenge captureChallenge `json:"challenge"`
	Identity  json.RawMessage  `json:"identity"`
	Capture   json.RawMessage  `json:"capture"`
}

type assembledReceipt struct {
	Schema    string            `json:"schema"`
	Status    string            `json:"status"`
	Challenge string            `json:"challenge"`
	Identity  json.RawMessage   `json:"identity"`
	Captures  []json.RawMessage `json:"captures"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("scrollregionatreceipt", flag.ContinueOnError)
	flags.SetOutput(stderr)
	safariDirectory := flags.String("safari-capture-dir", "", "direct macOS Safari+VoiceOver capture directory")
	chromiumDirectory := flags.String("chromium-capture-dir", "", "direct macOS Chromium+VoiceOver capture directory")
	identityPath := flags.String("identity", "", "frozen candidate identity sidecar")
	challengePath := flags.String("challenge", "", "independent challenge JSON")
	outputPath := flags.String("output", "", "new final AT receipt path")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "scrollregionatreceipt: positional arguments are forbidden")
		return 2
	}
	for _, required := range []struct{ name, value string }{
		{"safari-capture-dir", *safariDirectory}, {"chromium-capture-dir", *chromiumDirectory},
		{"identity", *identityPath}, {"challenge", *challengePath}, {"output", *outputPath},
	} {
		if strings.TrimSpace(required.value) == "" {
			fmt.Fprintf(stderr, "scrollregionatreceipt: --%s is required\n", required.name)
			return 2
		}
	}
	if err := assemble(*safariDirectory, *chromiumDirectory, *identityPath, *challengePath, *outputPath); err != nil {
		fmt.Fprintf(stderr, "scrollregionatreceipt: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "T-GS-011 final AT receipt assembled at %s; validate it with scripts/validate-scrollregion-at-receipt.sh.\n", *outputPath)
	return 0
}

func assemble(safariDirectory, chromiumDirectory, identityPath, challengePath, outputPath string) error {
	identity, err := readRawJSON(identityPath)
	if err != nil {
		return fmt.Errorf("read frozen identity: %w", err)
	}
	challenge, err := readChallenge(challengePath)
	if err != nil {
		return err
	}
	safari, err := readBundle(safariDirectory, "macos-safari-voiceover", identity, challenge)
	if err != nil {
		return err
	}
	chromium, err := readBundle(chromiumDirectory, "macos-chromium-voiceover", identity, challenge)
	if err != nil {
		return err
	}
	receipt := assembledReceipt{Schema: receiptSchema, Status: "captured", Challenge: challenge.Challenge, Identity: identity, Captures: []json.RawMessage{safari.Capture, chromium.Capture}}
	encoded, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(outputPath)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(abs, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create new receipt: %w", err)
	}
	defer file.Close()
	if _, err := file.Write(append(encoded, '\n')); err != nil {
		return err
	}
	return file.Sync()
}

func captureBundlePath(directory string) (string, error) {
	abs, err := filepath.Abs(directory)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("stat direct capture directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("direct capture path is not a directory")
	}
	return filepath.Join(abs, "capture.json"), nil
}

func readBundle(directory, expectedPair string, identity json.RawMessage, challenge captureChallenge) (captureBundle, error) {
	path, err := captureBundlePath(directory)
	if err != nil {
		return captureBundle{}, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return captureBundle{}, fmt.Errorf("read direct capture bundle: %w", err)
	}
	var bundle captureBundle
	if err := decodeStrictJSON(content, &bundle); err != nil {
		return captureBundle{}, fmt.Errorf("decode direct capture bundle: %w", err)
	}
	if bundle.Schema != bundleSchema || bundle.Challenge != challenge || !sameJSON(bundle.Identity, identity) {
		return captureBundle{}, fmt.Errorf("direct capture bundle does not bind exact challenge and identity")
	}
	var capture struct {
		Pair string `json:"pair"`
	}
	if err := decodeStrictJSON(bundle.Capture, &capture); err != nil {
		return captureBundle{}, fmt.Errorf("decode direct capture pair: %w", err)
	}
	if capture.Pair != expectedPair {
		return captureBundle{}, fmt.Errorf("direct capture bundle pair mismatch: got %q, want %q", capture.Pair, expectedPair)
	}
	return bundle, nil
}

func readChallenge(path string) (captureChallenge, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return captureChallenge{}, fmt.Errorf("read independent challenge: %w", err)
	}
	var challenge captureChallenge
	if err := decodeStrictJSON(content, &challenge); err != nil {
		return captureChallenge{}, fmt.Errorf("decode independent challenge: %w", err)
	}
	if challenge.Schema != challengeSchema || !regexp.MustCompile(challengeHexPattern).MatchString(challenge.Challenge) || challenge.IssuedAt == "" {
		return captureChallenge{}, fmt.Errorf("independent challenge is invalid")
	}
	return challenge, nil
}

func readRawJSON(path string) (json.RawMessage, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var value any
	if err := decodeStrictJSON(content, &value); err != nil {
		return nil, err
	}
	return json.RawMessage(content), nil
}

func sameJSON(left, right json.RawMessage) bool {
	var leftValue any
	var rightValue any
	return decodeStrictJSON(left, &leftValue) == nil && decodeStrictJSON(right, &rightValue) == nil && bytes.Equal(canonicalJSON(leftValue), canonicalJSON(rightValue))
}

func canonicalJSON(value any) []byte {
	encoded, _ := json.Marshal(value)
	return encoded
}

func decodeStrictJSON(content []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("trailing JSON values")
		}
		return err
	}
	return nil
}
