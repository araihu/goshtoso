package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/araihu/goshtoso/site/internal/server"
)

const (
	atCaptureChallengeEnvironment = "GOSHTOSO_SCROLLREGION_AT_CAPTURE_CHALLENGE"
	atCaptureTreeEnvironment      = "GOSHTOSO_SCROLLREGION_AT_CAPTURE_CANDIDATE_TREE"
	atCaptureManifestEnvironment  = "GOSHTOSO_SCROLLREGION_AT_CAPTURE_MANIFEST_SHA256"
	atCapturePairEnvironment      = "GOSHTOSO_SCROLLREGION_AT_CAPTURE_PAIR"
	atCaptureRoute                = "/components/scroll-region"
	atCaptureStateQuery           = "t-gs-011-at-state"
	atCaptureActionTokenQuery     = "t-gs-011-at-action-token"

	atCaptureChallengeHeader     = "X-Goshtoso-T-GS-011-AT-Challenge"
	atCaptureCandidateTreeHeader = "X-Goshtoso-T-GS-011-Candidate-Tree"
	atCaptureManifestHeader      = "X-Goshtoso-T-GS-011-Manifest-SHA256"
	atCaptureBodySHA256Header    = "X-Goshtoso-T-GS-011-Body-SHA256"
	atCapturePairHeader          = "X-Goshtoso-T-GS-011-AT-Pair"
)

var atCaptureHex64 = regexp.MustCompile(`^[0-9a-f]{64}$`)
var atCaptureGitObject = regexp.MustCompile(`^[0-9a-f]{40}$`)
var atCapturePair = regexp.MustCompile(`^macos-(?:safari|chromium)-voiceover$`)
var atCaptureState = regexp.MustCompile(`^(?:default|no-overflow|start|middle|end|focused)$`)

// atCaptureBinding is enabled only by the final T-GS-011 capture adapter.
// It makes exact page bytes carry the independent challenge and frozen source
// identifiers. Normal site responses remain byte-for-byte unchanged.
type atCaptureBinding struct {
	Challenge      string
	CandidateTree  string
	ManifestSHA256 string
	Pair           string
}

func (binding atCaptureBinding) enabled() bool {
	return binding.Challenge != "" || binding.CandidateTree != "" || binding.ManifestSHA256 != "" || binding.Pair != ""
}

func (binding atCaptureBinding) valid() bool {
	return atCaptureHex64.MatchString(binding.Challenge) && atCaptureGitObject.MatchString(binding.CandidateTree) && atCaptureHex64.MatchString(binding.ManifestSHA256) && atCapturePair.MatchString(binding.Pair)
}

func atCaptureBindingFromEnvironment() (atCaptureBinding, error) {
	binding := atCaptureBinding{
		Challenge:      os.Getenv(atCaptureChallengeEnvironment),
		CandidateTree:  os.Getenv(atCaptureTreeEnvironment),
		ManifestSHA256: os.Getenv(atCaptureManifestEnvironment),
		Pair:           os.Getenv(atCapturePairEnvironment),
	}
	if !binding.enabled() {
		return atCaptureBinding{}, nil
	}
	if !binding.valid() {
		return atCaptureBinding{}, fmt.Errorf("T-GS-011 AT capture binding environment is incomplete or malformed")
	}
	return binding, nil
}

type atCaptureResponseWriter struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (writer *atCaptureResponseWriter) Header() http.Header { return writer.header }

func (writer *atCaptureResponseWriter) WriteHeader(status int) {
	if writer.status == 0 {
		writer.status = status
	}
}

func (writer *atCaptureResponseWriter) Write(content []byte) (int, error) {
	if writer.status == 0 {
		writer.status = http.StatusOK
	}
	return writer.body.Write(content)
}

// atCaptureResponseWrapper annotates only the capture route. The raw served
// HTML itself contains the challenge. The adapter saves that direct body and
// rechecks candidate identity before it signs its digest.
func atCaptureResponseWrapper(next http.Handler, binding atCaptureBinding) http.Handler {
	if !binding.enabled() {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != atCaptureRoute {
			next.ServeHTTP(w, r)
			return
		}
		if !binding.valid() {
			http.Error(w, "invalid T-GS-011 AT capture binding", http.StatusInternalServerError)
			return
		}
		if r.URL.Query().Get("t-gs-011-at-capture") != binding.Challenge {
			http.Error(w, "missing independent T-GS-011 AT capture challenge", http.StatusForbidden)
			return
		}
		state := r.URL.Query().Get(atCaptureStateQuery)
		token := r.URL.Query().Get(atCaptureActionTokenQuery)
		if !atCaptureState.MatchString(state) || token != atCaptureActionToken(binding.Challenge, state) {
			http.Error(w, "missing exact T-GS-011 AT action token", http.StatusForbidden)
			return
		}
		recorded := &atCaptureResponseWriter{header: make(http.Header)}
		next.ServeHTTP(recorded, r)
		status := recorded.status
		if status == 0 {
			status = http.StatusOK
		}
		body := recorded.body.Bytes()
		if status < http.StatusOK || status >= http.StatusMultipleChoices || !strings.Contains(strings.ToLower(recorded.header.Get("Content-Type")), "text/html") {
			copyATCaptureHeaders(w.Header(), recorded.header)
			w.WriteHeader(status)
			_, _ = w.Write(body)
			return
		}
		metadata := fmt.Appendf(nil, `<meta name="goshtoso-t-gs-011-at-challenge" content="%s"><meta name="goshtoso-t-gs-011-candidate-tree" content="%s"><meta name="goshtoso-t-gs-011-manifest-sha256" content="%s"><meta name="goshtoso-t-gs-011-at-pair" content="%s"><meta name="goshtoso-t-gs-011-at-action-state" content="%s"><meta name="goshtoso-t-gs-011-at-action-token" content="%s">`, binding.Challenge, binding.CandidateTree, binding.ManifestSHA256, binding.Pair, state, token)
		index := bytes.LastIndex(bytes.ToLower(body), []byte("</head>"))
		if index < 0 {
			http.Error(w, "T-GS-011 AT capture route lacks HTML head", http.StatusInternalServerError)
			return
		}
		boundBody := make([]byte, 0, len(body)+len(metadata))
		boundBody = append(boundBody, body[:index]...)
		boundBody = append(boundBody, metadata...)
		boundBody = append(boundBody, body[index:]...)
		digest := sha256.Sum256(boundBody)
		copyATCaptureHeaders(w.Header(), recorded.header)
		w.Header().Del("Content-Length")
		w.Header().Set(atCaptureChallengeHeader, binding.Challenge)
		w.Header().Set(atCaptureCandidateTreeHeader, binding.CandidateTree)
		w.Header().Set(atCaptureManifestHeader, binding.ManifestSHA256)
		w.Header().Set(atCaptureBodySHA256Header, hex.EncodeToString(digest[:]))
		w.Header().Set(atCapturePairHeader, binding.Pair)
		w.WriteHeader(status)
		_, _ = w.Write(boundBody)
	})
}

func atCaptureActionToken(challenge, state string) string {
	return "goshtoso-t-gs-011-" + challenge + "-" + state
}

func copyATCaptureHeaders(destination, source http.Header) {
	for name, values := range source {
		for _, value := range values {
			destination.Add(name, value)
		}
	}
}

func main() {
	var port string
	flag.StringVar(&port, "port", "8090", "Server port")
	flag.Parse()

	projectRoot := resolveProjectRoot()

	srv := server.New(projectRoot)
	binding, err := atCaptureBindingFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting server on http://localhost%s", addr)
	log.Printf("Accordion Component Demo: http://localhost%s/components/accordion", addr)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: atCaptureResponseWrapper(srv, binding),
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
