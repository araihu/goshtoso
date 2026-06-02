package assets

import (
	"strings"
	"testing"
)

func TestTailwindVersion(t *testing.T) {
	got := TailwindVersion()
	if got != "4.3.0" {
		t.Fatalf("TailwindVersion() = %q, want %q", got, "4.3.0")
	}
	if strings.HasPrefix(got, "v") {
		t.Fatalf("TailwindVersion() must not include a leading v: %q", got)
	}
}
