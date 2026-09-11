package iconpack

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiskFileRejectsChangesBeforePublication(t *testing.T) {
	for _, mode := range []string{"rewrite", "replace", "grow", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "source")
			data := []byte("verified")
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			file := fileData{path: path, size: int64(len(data)), hash: hashBytes(data)}
			if _, err := file.read(t.Context()); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if mode == "cancel" {
				cancel()
			} else {
				if mode == "replace" {
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
				}
				replacement := []byte("modified")
				if mode == "grow" {
					replacement = append(data, '!')
				}
				if err := os.WriteFile(path, replacement, 0600); err != nil {
					t.Fatal(err)
				}
			}
			output := filepath.Join(root, "output")
			if _, _, err := publishFileOutput(ctx, output, map[string]fileData{"images/a.png": file}, false); err == nil {
				t.Fatal("published unverified bytes")
			}
			if _, err := os.Stat(output); !os.IsNotExist(err) {
				t.Fatalf("output exists: %v", err)
			}
			matches, err := filepath.Glob(filepath.Join(root, ".output.tmp-*"))
			if err != nil || len(matches) != 0 {
				t.Fatalf("staging leaked: %v %v", matches, err)
			}
		})
	}
}

func TestDiskFileReadDoesNotTrustReplacementPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(path, []byte("bad"), 0600); err != nil {
		t.Fatal(err)
	}
	file := fileData{path: path, size: 3, hash: hashBytes([]byte("old"))}
	if err := file.copy(t.Context(), io.Discard); err == nil || !strings.Contains(err.Error(), "integrity mismatch") {
		t.Fatalf("err=%v", err)
	}
}
