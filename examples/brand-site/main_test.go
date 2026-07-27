package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildProducesStaticDocumentAndStyles(t *testing.T) {
	out := t.TempDir()
	if err := build(out); err != nil {
		t.Fatal(err)
	}
	html, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"<main", "What we make", "data-theme=\"brand\""} {
		if !strings.Contains(string(html), want) {
			t.Fatalf("index.html missing %q", want)
		}
	}
	for _, name := range []string{"goshtoso.css", "brand.css"} {
		info, err := os.Stat(filepath.Join(out, "assets", name))
		if err != nil || info.Size() == 0 {
			t.Fatalf("asset %s = %v, %v", name, info, err)
		}
	}
}
