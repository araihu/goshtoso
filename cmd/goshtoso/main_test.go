package main

import (
	"runtime/debug"
	"testing"
)

func TestVersionString(t *testing.T) {
	info := &debug.BuildInfo{Main: debug.Module{Version: "v0.0.1"}}
	if got := versionString(info, "4.3.0"); got != "goshtoso v0.0.1 (tailwindcss 4.3.0)" {
		t.Fatalf("got %q", got)
	}
	if got := versionString(nil, "4.3.0"); got != "goshtoso (devel) (tailwindcss 4.3.0)" {
		t.Fatalf("nil buildinfo: got %q", got)
	}
}

func TestSourcePath(t *testing.T) {
	if got := sourcePath("/mod/cache/goshtoso@v1"); got != "/mod/cache/goshtoso@v1/components" {
		t.Fatalf("got %q", got)
	}
}
