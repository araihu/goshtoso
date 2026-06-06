package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

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
