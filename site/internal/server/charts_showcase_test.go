package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	chartassets "github.com/araihu/goshtoso-charts/assets"
	"github.com/stretchr/testify/require"
)

func TestChartsShowcaseRouteAndAssetsUseLocalModule(t *testing.T) {
	t.Parallel()
	server := &Server{}

	page := httptest.NewRecorder()
	server.handleChartsShowcase(page, httptest.NewRequest(http.MethodGet, "/playground/extensions/charts", nil))
	require.Equal(t, http.StatusOK, page.Code)
	require.Contains(t, page.Header().Get("Content-Type"), "text/html")
	require.Equal(t, "public, max-age=3600, stale-while-revalidate=86400", page.Header().Get("Cache-Control"))
	require.Contains(t, page.Body.String(), `data-charts-showcase`)

	frame := httptest.NewRecorder()
	server.handleChartsShowcaseFrame(frame, httptest.NewRequest(http.MethodGet, "/playground/extensions/charts/frame", nil))
	require.Equal(t, http.StatusOK, frame.Code)
	require.Contains(t, frame.Header().Get("Content-Type"), "text/html")
	require.Equal(t, "public, max-age=3600, stale-while-revalidate=86400", frame.Header().Get("Cache-Control"))
	require.Contains(t, frame.Body.String(), `id="charts-showcase-frame-line-3d"`)

	runtime := httptest.NewRecorder()
	withImmutableCache(chartassets.Handler()).ServeHTTP(runtime, httptest.NewRequest(http.MethodGet, chartassets.RuntimeURL, nil))
	require.Equal(t, http.StatusOK, runtime.Code)
	require.Equal(t, "public, max-age=31536000, immutable", runtime.Header().Get("Cache-Control"))
	require.True(t, strings.Contains(runtime.Header().Get("Content-Type"), "javascript"))
	require.NotEmpty(t, runtime.Body.Bytes())
}
