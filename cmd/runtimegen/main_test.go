package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/araihu/assets/assetmeta"
)

func TestInventoryAdaptsGeneratedMuambaRecords(t *testing.T) {
	inventory, err := inventory()
	if err != nil {
		t.Fatal(err)
	}
	if got := len(inventory.Resources()); got != 5 {
		t.Fatalf("resources = %d, want 5", got)
	}
	resolved, ok := inventory.Resolve(assetmeta.Ref{Resource: "alpinejs", Download: "core-js"})
	if !ok || resolved.Download.Hash == "" || resolved.Download.Integrity == "" {
		t.Fatalf("alpinejs/core-js = %#v, %t", resolved, ok)
	}
}

func TestRunRejectsPositionalArguments(t *testing.T) {
	err := run([]string{"unexpected"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "unexpected arguments") {
		t.Fatalf("error = %v", err)
	}
}
