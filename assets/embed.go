// Package assets provides embedded static files and their public runtime
// contract for Goshtoso components. Use Handler to serve them at /assets/,
// FS or ReadFile for read-only static export, DefaultRuntimeManifest to inspect
// the ordered dependency/fallback set, and GoshtosoVersion to identify the
// exact library module linked into a consumer.
//
// Usage:
//
//	mux := http.NewServeMux()
//	mux.Handle("/assets/", assets.Handler())
//
// This serves compiled CSS, the Muamba-backed JavaScript runtime, first-party
// JavaScript, brand images, and icons. Use DefaultRuntimeManifest to inspect
// exact versions and local URLs.
package assets

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"path"
	"sort"
	"strings"
	"time"
)

//go:embed styles.css goshtoso-theme.css js/*.js images icons
var files embed.FS

type embeddedAssetSource struct {
	fileSystem fs.FS
	name       string
}

type embeddedAssetFileSystem struct {
	files       map[string]embeddedAssetSource
	directories map[string][]fs.DirEntry
}

var publicFiles = newEmbeddedAssetFileSystem()

// FS returns a read-only filesystem containing every regular file published by
// Handler, including first-party static files and Muamba-backed runtime files.
// It is rooted at Handler's /assets/ mount, so names use slash-separated
// fs.ValidPath form such as "styles.css" and "js/goshtoso.min.js". Callers
// choose destinations and perform any filesystem writes themselves.
func FS() fs.FS {
	return publicFiles
}

// ReadFile returns caller-owned bytes for one regular file in FS. It rejects
// absolute paths, traversal, backslash-separated paths, directories, and
// unknown files. Names are relative to Handler's /assets/ mount.
func ReadFile(name string) ([]byte, error) {
	return publicFiles.ReadFile(name)
}

func newEmbeddedAssetFileSystem() *embeddedAssetFileSystem {
	fileSources := make(map[string]embeddedAssetSource)
	err := fs.WalkDir(files, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type().IsRegular() {
			fileSources[name] = embeddedAssetSource{fileSystem: files, name: name}
		}
		return nil
	})
	if err != nil {
		panic("build embedded asset inventory: " + err.Error())
	}

	for name, ref := range muambaHTTPFiles {
		embeddedName, ok := muambaEmbeddedPaths[ref.resource+"\x00"+ref.download]
		if !ok {
			panic("build embedded asset inventory: missing Muamba path for " + ref.resource + "/" + ref.download)
		}
		fileSources[name] = embeddedAssetSource{fileSystem: muambaFiles, name: embeddedName}
	}

	directoryEntries := map[string]map[string]fs.DirEntry{".": {}}
	for name, source := range fileSources {
		parts := strings.Split(name, "/")
		parent := "."
		for index, part := range parts {
			isDirectory := index < len(parts)-1
			entry := embeddedAssetDirEntry{name: part, directory: isDirectory}
			if !isDirectory {
				entry.source = source
			}
			directoryEntries[parent][part] = entry
			if !isDirectory {
				continue
			}
			if parent == "." {
				parent = part
			} else {
				parent += "/" + part
			}
			if _, ok := directoryEntries[parent]; !ok {
				directoryEntries[parent] = make(map[string]fs.DirEntry)
			}
		}
	}

	directories := make(map[string][]fs.DirEntry, len(directoryEntries))
	for name, entriesByName := range directoryEntries {
		entries := make([]fs.DirEntry, 0, len(entriesByName))
		for _, entry := range entriesByName {
			entries = append(entries, entry)
		}
		sort.Slice(entries, func(left, right int) bool {
			return entries[left].Name() < entries[right].Name()
		})
		directories[name] = entries
	}

	return &embeddedAssetFileSystem{files: fileSources, directories: directories}
}

func (fileSystem *embeddedAssetFileSystem) Open(name string) (fs.File, error) {
	if !validEmbeddedAssetPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	if source, ok := fileSystem.files[name]; ok {
		return source.fileSystem.Open(source.name)
	}
	if entries, ok := fileSystem.directories[name]; ok {
		return &embeddedAssetDirectory{name: name, entries: entries}, nil
	}
	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

func (fileSystem *embeddedAssetFileSystem) ReadFile(name string) ([]byte, error) {
	file, err := fileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, &fs.PathError{Op: "read", Path: name, Err: fs.ErrInvalid}
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, &fs.PathError{Op: "read", Path: name, Err: err}
	}
	return data, nil
}

func validEmbeddedAssetPath(name string) bool {
	return name != "" && fs.ValidPath(name) && !strings.ContainsRune(name, '\\')
}

type embeddedAssetDirectory struct {
	name    string
	entries []fs.DirEntry
	offset  int
	closed  bool
}

func (directory *embeddedAssetDirectory) Close() error {
	directory.closed = true
	return nil
}

func (directory *embeddedAssetDirectory) Read([]byte) (int, error) {
	if directory.closed {
		return 0, fs.ErrClosed
	}
	return 0, &fs.PathError{Op: "read", Path: directory.name, Err: fs.ErrInvalid}
}

func (directory *embeddedAssetDirectory) ReadDir(count int) ([]fs.DirEntry, error) {
	if directory.closed {
		return nil, fs.ErrClosed
	}
	if directory.offset >= len(directory.entries) {
		if count > 0 {
			return nil, io.EOF
		}
		return nil, nil
	}
	end := len(directory.entries)
	if count > 0 && directory.offset+count < end {
		end = directory.offset + count
	}
	entries := append([]fs.DirEntry(nil), directory.entries[directory.offset:end]...)
	directory.offset = end
	return entries, nil
}

func (directory *embeddedAssetDirectory) Stat() (fs.FileInfo, error) {
	return embeddedAssetFileInfo{name: path.Base(directory.name), mode: fs.ModeDir | 0o555}, nil
}

type embeddedAssetDirEntry struct {
	name      string
	directory bool
	source    embeddedAssetSource
}

func (entry embeddedAssetDirEntry) Name() string {
	return entry.name
}

func (entry embeddedAssetDirEntry) IsDir() bool {
	return entry.directory
}

func (entry embeddedAssetDirEntry) Type() fs.FileMode {
	if entry.directory {
		return fs.ModeDir
	}
	return 0
}

func (entry embeddedAssetDirEntry) Info() (fs.FileInfo, error) {
	if entry.directory {
		return embeddedAssetFileInfo{name: entry.name, mode: fs.ModeDir | 0o555}, nil
	}
	info, err := fs.Stat(entry.source.fileSystem, entry.source.name)
	if err != nil {
		return nil, err
	}
	return embeddedAssetFileInfo{
		name:    entry.name,
		size:    info.Size(),
		mode:    info.Mode(),
		modTime: info.ModTime(),
		system:  info.Sys(),
	}, nil
}

type embeddedAssetFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	system  any
}

func (info embeddedAssetFileInfo) Name() string       { return info.name }
func (info embeddedAssetFileInfo) Size() int64        { return info.size }
func (info embeddedAssetFileInfo) Mode() fs.FileMode  { return info.mode }
func (info embeddedAssetFileInfo) ModTime() time.Time { return info.modTime }
func (info embeddedAssetFileInfo) IsDir() bool        { return info.mode.IsDir() }
func (info embeddedAssetFileInfo) Sys() any           { return info.system }

// Handler returns an http.Handler that serves the embedded Goshtoso assets
// (styles.css, js/, images/) — the generated library files head.Dependencies() links.
//
// Mount it at /assets/ WITHOUT wrapping it in your own StripPrefix — the
// handler already strips the /assets/ prefix internally:
//
//	http.Handle("/assets/", assets.Handler())             // correct
//	http.Handle("/assets/", http.StripPrefix("/assets/", assets.Handler())) // WRONG: double-strip → 404
func Handler() http.Handler {
	fileServer := http.FileServer(http.FS(files))
	assetHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if ref, ok := muambaHTTPFiles[request.URL.Path]; ok {
			serveMuambaFile(writer, request, ref)
			return
		}
		fileServer.ServeHTTP(writer, request)
	})
	return http.StripPrefix("/assets/", WithCacheControl(assetHandler))
}

// StylesCSS returns the compiled Goshtoso Tailwind CSS.
// Use this to extract the CSS to disk for Tailwind's @import directive.
func StylesCSS() ([]byte, error) {
	return files.ReadFile("styles.css")
}

// TailwindVersion returns the Tailwind CSS version locked in muamba.yaml.
// Match your own Tailwind build to this when compiling Goshtoso's theme source.
func TailwindVersion() string {
	resource, ok := MuambaResourceByName("tailwindcss")
	if !ok {
		return ""
	}
	return resource.Version
}

// ThemeCSS returns the Goshtoso theme SOURCE (tokens, @custom-variant, the 13
// [data-theme] blocks, base + utility layers) for importing into your OWN
// Tailwind v4 build. Unlike StylesCSS (compiled output you serve directly),
// this is source your Tailwind compiles. Pair it with a @source pointing at
// Goshtoso's components dir (see `goshtoso -source-path`).
func ThemeCSS() ([]byte, error) {
	return files.ReadFile("goshtoso-theme.css")
}

// AlpineVersion returns the Alpine.js version Goshtoso vendors (core, collapse,
// focus, and mask share this version). Generated from Muamba and the runtime overlay.
func AlpineVersion() string { return runtimeVersionAlpineJS }

// HTMXVersion returns the vendored HTMX version from the canonical manifest.
func HTMXVersion() string { return runtimeVersionHTMX }

// HTMXExtSSEVersion returns the vendored htmx-ext-sse version.
func HTMXExtSSEVersion() string { return runtimeVersionHTMXExtSSE }

// HTMXExtWSVersion returns the vendored htmx-ext-ws version.
func HTMXExtWSVersion() string { return runtimeVersionHTMXExtWS }
