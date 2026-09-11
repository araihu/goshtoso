package server

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExamplesOverview(t *testing.T) {
	s := &Server{}
	for _, partial := range []bool{false, true} {
		req := httptest.NewRequest(http.MethodGet, "/examples", nil)
		if partial {
			req.Header.Set("HX-Request", "true")
			req.Header.Set("HX-Request-Type", "partial")
		}
		rec := httptest.NewRecorder()
		s.handleExample(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), `id="examples-overview"`)
		require.Contains(t, rec.Body.String(), "HTMX")
		require.Contains(t, rec.Body.String(), "WebSockets")
		require.NotContains(t, rec.Body.String(), `href="/docs/application-patterns"`)
	}
}
