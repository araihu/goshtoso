package main

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"testing"
)

func TestReceiptCLIRejectsClaimantArtifactPathFlags(t *testing.T) {
	var stderr bytes.Buffer
	exitCode := run(
		[]string{
			"--safari-capture-dir", "/tmp/safari",
			"--chromium-capture-dir", "/tmp/chromium",
			"--identity", "/tmp/identity.json",
			"--challenge", "/tmp/challenge.json",
			"--output", "/tmp/final.json",
			"--screenshot", "/tmp/claimant.png",
		},
		io.Discard,
		&stderr,
	)
	if exitCode != 2 || !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Fatalf("receipt assembler must reject raw artifact input flags: exit=%d stderr=%q", exitCode, stderr.String())
	}
}

func TestCaptureBundlePathIsFixedWithinCaptureDirectory(t *testing.T) {
	directory := t.TempDir()
	path, err := captureBundlePath(directory)
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(directory, "capture.json") {
		t.Fatalf("unexpected bundle path %q", path)
	}
}
