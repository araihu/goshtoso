package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRequiresIdentityReceiptsBlockerAndOutput(t *testing.T) {
	var stdout bytes.Buffer
	err := run(nil, &stdout)
	if err == nil || !strings.Contains(err.Error(), "-repo, -commit, -tree, -receipts, -at-blocker, and -output are required") {
		t.Fatalf("error = %v", err)
	}
}

func TestRunGeneratesSiteLocalLedger(t *testing.T) {
	output := filepath.Join(t.TempDir(), "site-ledger")
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	if err := run([]string{"-repo", repo, "-site-ledger-output", output}, &stdout); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "site_ledger=") {
		t.Fatalf("output = %q", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(output, "bfull.go")); err != nil {
		t.Fatalf("generated site ledger missing bfull.go: %v", err)
	}
}
