package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/araihu/goshtoso/site/internal/server"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "8090", "Server port")
	flag.Parse()

	projectRoot := resolveProjectRoot()

	srv := server.New(projectRoot)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting server on http://localhost%s", addr)
	log.Printf("Button Component Demo: http://localhost%s/components/button", addr)

	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func resolveProjectRoot() string {
	if envRoot := os.Getenv("GOSHTOSO_PROJECT_ROOT"); envRoot != "" {
		return envRoot
	}

	// assets/ belongs to the library at the repo root, while this server lives
	// in the site module (site/cmd/server) — a fixed "../.." no longer reaches
	// it. Probe likely roots in order, climbing until we find assets/:
	//   1. cwd          — covers the deployed container (WORKDIR holds assets/).
	//   2. the binary   — covers an installed binary next to/above assets/.
	//   3. this source  — covers `go run` during local dev.
	if cwd, err := os.Getwd(); err == nil {
		if root, ok := findAssetsRoot(cwd); ok {
			return root
		}
	}
	if exe, err := os.Executable(); err == nil {
		if root, ok := findAssetsRoot(filepath.Dir(exe)); ok {
			return root
		}
	}
	_, filename, _, _ := runtime.Caller(0)
	if root, ok := findAssetsRoot(filepath.Dir(filename)); ok {
		return root
	}
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

// findAssetsRoot walks up from startDir until it finds a directory containing
// an assets/ subdirectory, returning (dir, true) on success.
func findAssetsRoot(startDir string) (string, bool) {
	dir := startDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "assets")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false // reached filesystem root without finding assets/
		}
		dir = parent
	}
}
