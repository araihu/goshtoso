package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

func TestChallengeCLIProducesIndependentRandomChallenge(t *testing.T) {
	path := filepath.Join(t.TempDir(), "capture-challenge.json")
	var stdout, stderr bytes.Buffer
	if exitCode := run([]string{"--output", path}, &stdout, &stderr, func() time.Time {
		return time.Date(2026, time.August, 12, 16, 0, 0, 0, time.UTC)
	}); exitCode != 0 {
		t.Fatalf("challenge command failed: exit=%d stderr=%q", exitCode, stderr.String())
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var challenge captureChallenge
	if err := json.Unmarshal(content, &challenge); err != nil {
		t.Fatal(err)
	}
	if challenge.Schema != challengeSchema || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(challenge.Challenge) || challenge.IssuedAt != "2026-08-12T16:00:00Z" {
		t.Fatalf("invalid independently generated challenge: %#v", challenge)
	}
}

func TestChallengeCLIRefusesOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "capture-challenge.json")
	if err := os.WriteFile(path, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	if exitCode := run([]string{"--output", path}, &bytes.Buffer{}, &stderr, time.Now); exitCode != 1 || !bytes.Contains(stderr.Bytes(), []byte("exists")) {
		t.Fatalf("challenge command must refuse an existing file: exit=%d stderr=%q", exitCode, stderr.String())
	}
}
