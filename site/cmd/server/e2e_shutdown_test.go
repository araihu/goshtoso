package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestATCaptureResponseWrapperBindsDirectCandidateHTML(t *testing.T) {
	wrapped := atCaptureResponseWrapper(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, "<!doctype html><html><head><title>Scroll Region</title></head><body><div role=\"region\" aria-label=\"Activity history\"></div></body></html>")
	}), atCaptureBinding{
		Challenge:      strings.Repeat("a", 64),
		CandidateTree:  strings.Repeat("b", 40),
		ManifestSHA256: strings.Repeat("c", 64),
		Pair:           "macos-safari-voiceover",
	})

	recorder := httptest.NewRecorder()
	challenge := strings.Repeat("a", 64)
	token := atCaptureActionToken(challenge, "default")
	wrapped.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/scroll-region?t-gs-011-at-capture="+challenge+"&t-gs-011-at-state=default&t-gs-011-at-action-token="+token, nil))
	body := recorder.Body.String()
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, challenge, recorder.Header().Get(atCaptureChallengeHeader))
	require.Equal(t, strings.Repeat("b", 40), recorder.Header().Get(atCaptureCandidateTreeHeader))
	require.Equal(t, strings.Repeat("c", 64), recorder.Header().Get(atCaptureManifestHeader))
	require.NotEmpty(t, recorder.Header().Get(atCaptureBodySHA256Header))
	require.Equal(t, "macos-safari-voiceover", recorder.Header().Get(atCapturePairHeader))
	require.Contains(t, body, `name="goshtoso-t-gs-011-at-challenge" content="`+challenge+`"`)
	require.Contains(t, body, `name="goshtoso-t-gs-011-at-action-token" content="`+token+`"`)
	require.NotContains(t, body, `aria-live=`, "capture metadata must not inject synthetic AT speech into the tested DOM")
	require.NotContains(t, body, `T-GS-011 AT action token `+token)
	require.Contains(t, body, `role="region"`)
}

func TestFindAssetsRootSkipsNestedSiteAssetsPackage(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	start := filepath.Join(root, "site", "tests", "e2e")
	require.NoError(t, os.MkdirAll(filepath.Join(root, "assets"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "site", "assets"), 0o755))
	require.NoError(t, os.MkdirAll(start, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "assets", "styles.css"), []byte("fixture"), 0o644))

	got, ok := findAssetsRoot(start)
	require.True(t, ok)
	require.Equal(t, root, got)
}

func TestE2EShutdownWrapperDisabledWithoutToken(t *testing.T) {
	called := false
	handler := e2eShutdownWrapper(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	}), "", func() {})

	req := httptest.NewRequest(http.MethodPost, "/__e2e/shutdown", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.True(t, called)
	require.Equal(t, http.StatusTeapot, rec.Code)
}

func TestE2EShutdownWrapperRequiresPostAndToken(t *testing.T) {
	shutdowns := 0
	handler := e2eShutdownWrapper(http.NotFoundHandler(), "secret", func() {
		shutdowns++
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/__e2e/shutdown?token=secret", nil))
	require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	require.Equal(t, 0, shutdowns)

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/__e2e/shutdown?token=wrong", nil))
	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Equal(t, 0, shutdowns)
}

func TestE2EShutdownWrapperCallsShutdown(t *testing.T) {
	shutdowns := 0
	handler := e2eShutdownWrapper(http.NotFoundHandler(), "secret", func() {
		shutdowns++
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/__e2e/shutdown?token=secret", nil))

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, 1, shutdowns)
}
