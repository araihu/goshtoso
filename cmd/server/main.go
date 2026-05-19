package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/araihu/goshtoso/internal/server"
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

	if cwd, err := os.Getwd(); err == nil {
		if _, err := os.Stat(filepath.Join(cwd, "assets")); err == nil {
			return cwd
		}
	}

	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..")
}
