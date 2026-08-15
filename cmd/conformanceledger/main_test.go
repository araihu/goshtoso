package main

import (
	"bytes"
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
