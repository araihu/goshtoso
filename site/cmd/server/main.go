package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

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
	log.Printf("Accordion Component Demo: http://localhost%s/components/accordion", addr)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: srv,
	}
	httpServer.Handler = e2eShutdownWrapper(httpServer.Handler, os.Getenv("GOSHTOSO_E2E_SHUTDOWN_TOKEN"), func() {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := httpServer.Shutdown(ctx); err != nil {
				log.Printf("Server shutdown error: %v", err)
			}
		}()
	})

	errs := make(chan error, 1)
	go func() {
		errs <- httpServer.ListenAndServe()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case sig := <-signals:
		log.Printf("Received %s, shutting down", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown error: %v", err)
		}
	case err := <-errs:
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		log.Fatalf("Server error: %v", err)
	}
}

func e2eShutdownWrapper(next http.Handler, token string, shutdown func()) http.Handler {
	if token == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/__e2e/shutdown" {
			next.ServeHTTP(w, r)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Query().Get("token") != token {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		shutdown()
	})
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

// findAssetsRoot walks up from startDir until it finds the library asset root.
// Requiring styles.css avoids mistaking the site module's separate site/assets
// package for the repository root.
func findAssetsRoot(startDir string) (string, bool) {
	dir := startDir
	for {
		if info, err := os.Stat(filepath.Join(dir, "assets", "styles.css")); err == nil && !info.IsDir() {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false // reached filesystem root without finding assets/
		}
		dir = parent
	}
}
