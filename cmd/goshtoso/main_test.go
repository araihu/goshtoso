package main

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"
)

func TestInitBrandSiteCopiesTemplateAndRejectsNonEmptyDestination(t *testing.T) {
	module := t.TempDir()
	source := filepath.Join(module, "examples", "brand-site")
	if err := os.MkdirAll(source, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "README.md"), []byte("starter"), 0644); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(t.TempDir(), "brand")
	if err := initBrandSite(module, destination); err != nil {
		t.Fatalf("initBrandSite: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(destination, "README.md"))
	if err != nil || string(got) != "starter" {
		t.Fatalf("copied README = %q, %v", got, err)
	}
	if err := initBrandSite(module, destination); err == nil {
		t.Fatal("initBrandSite accepted non-empty destination")
	}
}

func TestVersionString(t *testing.T) {
	info := &debug.BuildInfo{Main: debug.Module{Version: "v0.0.1"}}
	if got := versionString(info, "4.3.0"); got != "goshtoso v0.0.1 (tailwindcss 4.3.0)" {
		t.Fatalf("got %q", got)
	}
	if got := versionString(nil, "4.3.0"); got != "goshtoso (devel) (tailwindcss 4.3.0)" {
		t.Fatalf("nil buildinfo: got %q", got)
	}
	empty := &debug.BuildInfo{Main: debug.Module{Version: ""}}
	if got := versionString(empty, "4.3.0"); got != "goshtoso (devel) (tailwindcss 4.3.0)" {
		t.Fatalf("empty version: got %q", got)
	}
}

func TestSourcePath(t *testing.T) {
	if got := sourcePath("/mod/cache/goshtoso@v1"); got != "/mod/cache/goshtoso@v1/components" {
		t.Fatalf("got %q", got)
	}
}
