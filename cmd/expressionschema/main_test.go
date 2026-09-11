package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSchemaIsCurrent(t *testing.T) {
	output := filepath.Join(t.TempDir(), "schema.json")
	if err := run("../../expressions/set.go", output); err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("../../expressions/schema.json")
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("expression schema is stale; run go generate ./expressions")
	}
}
