package assets

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
)

func TestFSExposesExactHandlerFileInventory(t *testing.T) {
	want := embeddedHandlerFileNames(t)
	got := regularFileNames(t, FS())
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FS inventory = %v, want %v", got, want)
	}
	if err := fstest.TestFS(FS(), want...); err != nil {
		t.Fatalf("FS contract: %v", err)
	}

	server := httptest.NewServer(Handler())
	t.Cleanup(server.Close)
	for _, name := range got {
		gotBytes, err := fs.ReadFile(FS(), name)
		if err != nil {
			t.Fatalf("read %s from FS: %v", name, err)
		}
		response, err := http.Get(server.URL + "/assets/" + name)
		if err != nil {
			t.Fatalf("GET /assets/%s: %v", name, err)
		}
		wantBytes, readErr := io.ReadAll(response.Body)
		closeErr := response.Body.Close()
		if readErr != nil || closeErr != nil {
			t.Fatalf("read /assets/%s: %v; close: %v", name, readErr, closeErr)
		}
		if response.StatusCode != http.StatusOK {
			t.Fatalf("GET /assets/%s status = %d, want 200", name, response.StatusCode)
		}
		if !bytes.Equal(gotBytes, wantBytes) {
			t.Fatalf("FS bytes differ from Handler for %s", name)
		}
	}
}

func TestReadFileRejectsUnsafeUnknownAndNonRegularPaths(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr error
	}{
		{name: "empty", path: "", wantErr: fs.ErrInvalid},
		{name: "absolute", path: "/styles.css", wantErr: fs.ErrInvalid},
		{name: "traversal", path: "../styles.css", wantErr: fs.ErrInvalid},
		{name: "embedded traversal", path: "images/../styles.css", wantErr: fs.ErrInvalid},
		{name: "windows traversal", path: `images\..\styles.css`, wantErr: fs.ErrInvalid},
		{name: "directory", path: "images", wantErr: fs.ErrInvalid},
		{name: "root directory", path: ".", wantErr: fs.ErrInvalid},
		{name: "unknown", path: "unknown.js", wantErr: fs.ErrNotExist},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ReadFile(test.path)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ReadFile(%q) error = %v, want %v", test.path, err, test.wantErr)
			}
		})
	}

	if _, err := FS().Open("../styles.css"); !errors.Is(err, fs.ErrInvalid) {
		t.Fatalf("FS.Open traversal error = %v, want %v", err, fs.ErrInvalid)
	}
}

func TestReadFileReturnsCallerOwnedBytesFromReadOnlyFS(t *testing.T) {
	first, err := ReadFile("styles.css")
	if err != nil {
		t.Fatal(err)
	}
	wantFirstByte := first[0]
	first[0] ^= 0xff
	second, err := ReadFile("styles.css")
	if err != nil {
		t.Fatal(err)
	}
	if second[0] != wantFirstByte {
		t.Fatal("ReadFile returned shared mutable bytes")
	}

	file, err := FS().Open("styles.css")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	if _, writable := file.(io.Writer); writable {
		t.Fatal("FS returned a writable file")
	}
}

func TestFSSupportsByteIdenticalIdempotentConsumerExport(t *testing.T) {
	destination := t.TempDir()
	exportEmbeddedFS(t, destination)
	first := diskFileContents(t, destination)

	exportEmbeddedFS(t, destination)
	second := diskFileContents(t, destination)
	if !reflect.DeepEqual(second, first) {
		t.Fatal("repeated export changed destination bytes")
	}

	server := httptest.NewServer(Handler())
	t.Cleanup(server.Close)
	for name, got := range second {
		response, err := http.Get(server.URL + "/assets/" + name)
		if err != nil {
			t.Fatalf("GET /assets/%s: %v", name, err)
		}
		want, readErr := io.ReadAll(response.Body)
		_ = response.Body.Close()
		if readErr != nil || response.StatusCode != http.StatusOK {
			t.Fatalf("GET /assets/%s: status=%d read=%v", name, response.StatusCode, readErr)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("exported bytes differ from Handler for %s", name)
		}
	}
}

func embeddedHandlerFileNames(t *testing.T) []string {
	t.Helper()
	names := make(map[string]struct{})
	if err := fs.WalkDir(files, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type().IsRegular() {
			names[name] = struct{}{}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for _, resource := range MuambaResources() {
		for _, download := range resource.Downloads {
			name, ok := strings.CutPrefix(download.Path, "assets/")
			if ok {
				names[name] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(names))
	for name := range names {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func regularFileNames(t *testing.T, fileSystem fs.FS) []string {
	t.Helper()
	var names []string
	if err := fs.WalkDir(fileSystem, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type().IsRegular() {
			names = append(names, name)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	sort.Strings(names)
	return names
}

func exportEmbeddedFS(t *testing.T, destination string) {
	t.Helper()
	names := regularFileNames(t, FS())
	for _, name := range names {
		data, err := ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		target := filepath.Join(destination, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatalf("create parent for %s: %v", name, err)
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
}

func diskFileContents(t *testing.T, root string) map[string][]byte {
	t.Helper()
	result := make(map[string][]byte)
	if err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		result[filepath.ToSlash(relative)] = data
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return result
}
